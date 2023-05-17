package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/FZambia/sentinel"
	"github.com/Kotodian/gokit/retry"
	"github.com/Kotodian/gokit/retry/strategy"
	"github.com/gomodule/redigo/redis"
)

const (
	EnvMaxIdleConns   = "REDIS_MAX_IDLE_CONNS"
	EnvMaxActiveConns = "REDIS_MAX_ACTIVE_CONNS"
	EnvRedisPool      = "REDIS_POOL"
	EnvReidsAuth      = "REDIS_AUTH"
	EnvRedisMaster    = "REDIS_MASTER"
	EnvRedisDB        = "REDIS_DB"
)

var pool *redis.Pool

func Pool() *redis.Pool {
	return pool
}

func Init() {
	addrs := strings.Split(os.Getenv(EnvRedisPool), ",")
	auth := os.Getenv(EnvReidsAuth)
	db := int64(0)

	if redisDB := os.Getenv(EnvRedisDB); len(redisDB) > 0 {
		var err error
		db, err = strconv.ParseInt(redisDB, 10, 64)
		if err != nil {
			panic(err)
		}
	}
	if len(addrs) == 0 {
		return
	}
	if len(addrs) == 1 {
		pool = &redis.Pool{
			MaxActive:   getIntEnv(EnvMaxActiveConns, 10000),
			MaxIdle:     getIntEnv(EnvMaxIdleConns, 10),
			IdleTimeout: 300 * time.Second,
			Dial: func() (redis.Conn, error) {
				conn, err := redis.Dial("tcp", addrs[0],
					redis.DialConnectTimeout(time.Second),
					redis.DialWriteTimeout(3*time.Second),
					redis.DialDatabase(int(db)),
					redis.DialPassword(auth),
				)
				if err != nil {
					return nil, err
				}
				// conn = redis.NewLoggingConn(conn, log.Default(), "redis")
				return conn, nil
			},
		}
	} else {
		master := os.Getenv(EnvRedisMaster)
		if len(master) == 0 {
			master = "mymaster"
		}
		sntnl := &sentinel.Sentinel{
			Addrs:      addrs,
			MasterName: master,
			Dial: func(addr string) (redis.Conn, error) {
				timeout := 500 * time.Millisecond
				c, err := redis.DialTimeout("tcp", addr, timeout, timeout, timeout)
				if err != nil {
					return nil, err
				}
				return c, nil
			},
		}
		pool = &redis.Pool{
			Dial: func() (redis.Conn, error) {
				masterAddr, err := sntnl.MasterAddr()
				if err != nil {
					return nil, err
				}
				c, err := redis.Dial("tcp", masterAddr)
				if err != nil {
					return nil, err
				}
				if auth != "" {
					if _, err := c.Do("AUTH", auth); err != nil {
						c.Close()
						return nil, err
					}
				}
				if db != 0 {
					if _, err := c.Do("SELECT", db); err != nil {
						c.Close()
						return nil, err
					}
				}
				return c, nil

			},
		}
	}
}

func getIntEnv(key string, def int) int {
	envstr := os.Getenv(key)
	if envstr != "" {
		if tmp, _ := strconv.ParseInt(envstr, 10, 64); tmp > 0 {
			return int(tmp)
		}
	}
	return def
}

// GetRedis 从redis连接池中获取一个连接
func GetRedis() redis.Conn {
	return pool.Get()
}

// Do 执行一个redis命令
func Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	if len(args) < 1 {
		return nil, errors.New("missing required arguments")
	}
	c := GetRedis()
	reply, err = c.Do(commandName, args...)
	c.Close()
	return
}

func sliceHelper(reply interface{}, err error, name string, makeSlice func(int), assign func(int, interface{}) error) error {
	if err != nil {
		return err
	}
	switch reply := reply.(type) {
	case []interface{}:
		makeSlice(len(reply))
		for i := range reply {
			if reply[i] == nil {
				continue
			}
			if err := assign(i, reply[i]); err != nil {
				return err
			}
		}
		return nil
	case nil:
		return redis.ErrNil
	case redis.Error:
		return reply
	}
	return fmt.Errorf("redigo: unexpected type for %s, got type %T", name, reply)
}

func Uint64s(reply interface{}, err error) ([]uint64, error) {
	var result []uint64
	err = sliceHelper(reply, err, "Uint64s", func(n int) { result = make([]uint64, n) }, func(i int, v interface{}) error {
		switch v := v.(type) {
		case uint64:
			result[i] = v
			return nil
		case []byte:
			n, err := strconv.ParseUint(string(v), 10, 64)
			result[i] = n
			return err
		default:
			return fmt.Errorf("redigo: unexpected element type for Int64s, got type %T", v)
		}
	})
	return result, err
}

func Float64Map(result interface{}, err error) (map[string]float64, error) {
	values, err := redis.Values(result, err)
	if err != nil {
		return nil, err
	}
	if len(values)%2 != 0 {
		return nil, errors.New("redigo: Float64Map expects even number of values result")
	}
	m := make(map[string]float64, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].([]byte)
		if !ok {
			return nil, errors.New("redigo: Int64Map key not a bulk string value")
		}
		value, err := redis.Float64(values[i+1], nil)
		if err != nil {
			return nil, err
		}
		m[string(key)] = value
	}
	return m, nil
}

type SetFunc func(key string, val interface{}, keepalive int64)

func wrapSetFunc(redisConn redis.Conn, key string, val interface{}, keepalive int64) retry.Action {
	return func(attempt uint) error {
		var err error
		if keepalive == 0 {
			_, err = redisConn.Do("set", key, val)
		} else {
			_, err = redisConn.Do("set", key, val, "ex", keepalive)
		}
		return err
	}
}

func RetrySet(redisConn redis.Conn, key string, val interface{}, keepalive int64) error {
	return retry.Retry(wrapSetFunc(redisConn, key, val, keepalive), strategy.Limit(3))
}

func wrapHSetFunc(redisConn redis.Conn, key string, field string, val interface{}) retry.Action {
	return func(attempt uint) error {
		_, err := redisConn.Do("hset", key, field, val)
		return err
	}
}

func RetryHSet(redisConn redis.Conn, key string, field string, val interface{}) error {
	return retry.Retry(wrapHSetFunc(redisConn, key, field, val))
}

func HGet(redisConn redis.Conn, key string, field string, val interface{}) error {
	body, err := redis.Bytes(redisConn.Do("hget", key, field))
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, val)
	return err
}
