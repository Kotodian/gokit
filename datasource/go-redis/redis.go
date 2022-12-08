package goredis

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/Kotodian/gokit/retry"
	"github.com/Kotodian/gokit/retry/strategy"
	redis "github.com/go-redis/redis/v8"
)

const (
	EnvMaxIdleConns   = "REDIS_MAX_IDLE_CONNS"
	EnvMaxActiveConns = "REDIS_MAX_ACTIVE_CONNS"
	EnvRedisPool      = "REDIS_POOL"
	EnvReidsAuth      = "REDIS_AUTH"
)

const (
	Forever = 0 * time.Second
)

var rdb redis.Cmdable

func Redis() redis.Cmdable {
	return rdb
}

func Init() {
	addrs := strings.Split(os.Getenv(EnvRedisPool), ",")
	auth := os.Getenv(EnvReidsAuth)
	if len(addrs) == 0 {
		return
	}
	if len(addrs) > 1 {
		opts := &redis.FailoverOptions{
			MasterName:     "mymaster",
			SentinelAddrs:  addrs,
			RouteByLatency: true,
		}
		if len(auth) > 0 {
			opts.Password = auth
		}
		rdb = redis.NewFailoverClusterClient(opts)
	} else {
		opts := &redis.Options{
			Addr: addrs[0],
		}
		if len(auth) > 0 {
			opts.Password = auth
		}
		rdb = redis.NewClient(opts)
	}
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
