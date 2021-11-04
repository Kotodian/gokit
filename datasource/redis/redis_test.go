package redis

import (
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	os.Setenv("REDIS_POOL", "10.43.0.20:6379")
	os.Setenv("REDIS_POOL", "redis-sentinel-0.redis-sentinel-headless.default:26379,redis-sentinel-1.redis-sentinel-headless.default:26379,redis-sentinel-2.redis-sentinel-headless.default:26379")
	os.Setenv("REDIS_AUTH", "LhBIOQumQdgIm4ro")
	Init()
	_ = GetRedis()
	//conn.Do("set", "you", "me")
}
