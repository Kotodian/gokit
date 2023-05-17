package goredis

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"

	"github.com/Kotodian/gokit/retry"
	"github.com/Kotodian/gokit/retry/strategy"
)

const (
	EnvMaxIdleConns   = "REDIS_MAX_IDLE_CONNS"
	EnvMaxActiveConns = "REDIS_MAX_ACTIVE_CONNS"
	EnvRedisPool      = "REDIS_POOL"
	EnvReidsAuth      = "REDIS_AUTH"
	EnvRedisMaster    = "REDIS_MASTER"
	EnvRedisDB        = "REDIS_DB"
)

const (
	Forever = 0 * time.Second
)

var rdb redis.UniversalClient
var db int64

func Redis() redis.UniversalClient {
	return rdb
}

func Init() {
	addrs := strings.Split(os.Getenv(EnvRedisPool), ",")
	if len(addrs) == 0 {
		return
	}

	auth := os.Getenv(EnvReidsAuth)

	if redisDB := os.Getenv(EnvRedisDB); len(redisDB) > 0 {
		var err error
		db, err = strconv.ParseInt(redisDB, 10, 64)
		if err != nil {
			panic(err)
		}
	}
	opts := redis.UniversalOptions{
		Addrs:    addrs,
		Password: auth,
		DB:       int(db),
	}
	master := os.Getenv(EnvRedisMaster)
	if len(addrs) > 1 {
		if len(master) == 0 {
			master = "mymaster"
		}
		opts.MasterName = master
	}

	rdb = redis.NewUniversalClient(&opts)
}

type SetFunc func(key string, val interface{}, keepalive int64)

func wrapSetFunc(key string, val interface{}, keepalive int64) retry.Action {
	return func(attempt uint) error {
		var err error
		ctx := context.Background()
		if keepalive == 0 {
			err = rdb.Set(ctx, key, val, Forever).Err()
		} else {
			err = rdb.Set(ctx, key, val, time.Duration(keepalive)*time.Second).Err()
		}
		return err
	}
}

func RetrySet(key string, val interface{}, keepalive int64) error {
	return retry.Retry(wrapSetFunc(key, val, keepalive), strategy.Limit(3))
}

func wrapHSetFunc(redisConn redis.Conn, key string, field string, val interface{}) retry.Action {
	return func(attempt uint) error {
		err := redisConn.HSet(context.Background(), key, field, val).Err()
		return err
	}
}

func RetryHSet(redisConn redis.Conn, key string, field string, val interface{}) error {
	return retry.Retry(wrapHSetFunc(redisConn, key, field, val))
}

func HGet(ctx context.Context, key string, field string, val interface{}) error {
	body, err := rdb.HGet(ctx, key, field).Bytes()
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, val)
	return err
}
