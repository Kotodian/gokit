package redis

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/letsfire/redigo/v2"
	"github.com/letsfire/redigo/v2/mode/alone"
	"github.com/letsfire/redigo/v2/mode/sentinel"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	EnvMaxIdleConns   = "REDIS_MAX_IDLE_CONNS"
	EnvMaxActiveConns = "REDIS_MAX_ACTIVE_CONNS"
	EnvRedisPool      = "REDIS_POOL"
	EnvReidsAuth      = "REDIS_AUTH"
)

var redisMode redigo.ModeInterface

func Init() {
	addrs := strings.Split(os.Getenv(EnvRedisPool), ",")
	auth := os.Getenv(EnvReidsAuth)
	if len(addrs) == 0 {
		return
	}
	if len(addrs) == 1 {
		redisMode = alone.New(
			alone.PoolOpts(
				redigo.MaxActive(0),       // 最大连接数，默认0无限制
				redigo.MaxIdle(0),         // 最多保持空闲连接数，默认2*runtime.GOMAXPROCS(0)
				redigo.Wait(false),        // 连接耗尽时是否等待，默认false
				redigo.IdleTimeout(0),     // 空闲连接超时时间，默认0不超时
				redigo.MaxConnLifetime(0), // 连接的生命周期，默认0不失效
				redigo.TestOnBorrow(nil),  // 空间连接取出后检测是否健康，默认nil
			),
			alone.DialOpts(
				redis.DialReadTimeout(10*time.Second),    // 读取超时，默认time.Second
				redis.DialWriteTimeout(10*time.Second),   // 写入超时，默认time.Second
				redis.DialConnectTimeout(10*time.Second), // 连接超时，默认500*time.Millisecond
				redis.DialPassword(auth),                 // 鉴权密码，默认空
				redis.DialDatabase(0),                    // 数据库号，默认0
				redis.DialKeepAlive(time.Minute*5),       // 默认5*time.Minute
				redis.DialNetDial(nil),                   // 自定义dial，默认nil
				redis.DialUseTLS(false),                  // 是否用TLS，默认false
				redis.DialTLSSkipVerify(false),           // 服务器证书校验，默认false
				redis.DialTLSConfig(nil),                 // 默认nil，详见tls.Config
			))
	} else {
		redisMode = sentinel.New(
			sentinel.MasterName("mymaster"),
			sentinel.Addrs(addrs),
			sentinel.PoolOpts(
				redigo.MaxActive(0),       // 最大连接数，默认0无限制
				redigo.MaxIdle(0),         // 最多保持空闲连接数，默认2*runtime.GOMAXPROCS(0)
				redigo.Wait(false),        // 连接耗尽时是否等待，默认false
				redigo.IdleTimeout(0),     // 空闲连接超时时间，默认0不超时
				redigo.MaxConnLifetime(0), // 连接的生命周期，默认0不失效
				redigo.TestOnBorrow(nil),  // 空间连接取出后检测是否健康，默认nil
			),
			sentinel.DialOpts(
				redis.DialReadTimeout(10*time.Second),    // 读取超时，默认time.Second
				redis.DialWriteTimeout(10*time.Second),   // 写入超时，默认time.Second
				redis.DialConnectTimeout(10*time.Second), // 连接超时，默认500*time.Millisecond
				redis.DialPassword(auth),                 // 鉴权密码，默认空
				redis.DialDatabase(0),                    // 数据库号，默认0
				redis.DialKeepAlive(time.Minute*5),       // 默认5*time.Minute
				redis.DialNetDial(nil),                   // 自定义dial，默认nil
				redis.DialUseTLS(false),                  // 是否用TLS，默认false
				redis.DialTLSSkipVerify(false),           // 服务器证书校验，默认false
				redis.DialTLSConfig(nil),                 // 默认nil，详见tls.Config
			))
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
	return redisMode.GetConn()
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
