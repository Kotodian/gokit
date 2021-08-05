package cache

import (
	"encoding/json"
	"reflect"
	"strconv"
	"time"
	libredis "github.com/Kotodian/gokit/datasource/redis"
	"github.com/gomodule/redigo/redis"
)

type RedisCache struct {
	conn redis.Conn
}

func NewRedisCache() Cache {
	return &RedisCache{}
}

// get cached value by key.
//Get(key string) interface{}
func (c *RedisCache) Get(key string, ret interface{}) bool {
	reply, err := c.conn.Do("get", key)
	if err != nil {
		return false
	}
	if reply == nil {
		return false
	}

	err = json.Unmarshal([]byte(reply.(string)), ret)
	return err == nil
}

// GetMulti is a batch version of Get.
//GetMulti(keys []string) []interface{}
// set cached value with key and expire time.
func (c *RedisCache) Put(key string, val interface{}, timeout time.Duration) error {
	marshal, err := json.Marshal(val)
	if err != nil {
		return err
	}
	_, err = c.conn.Do("set", key, marshal, timeout)
	return err
}

// GetOrStore
func (c *RedisCache) GetOrStore(name string, value interface{}, lifespan time.Duration) (ret interface{}, exists bool) {
	var err error
	//Original struct
	t := reflect.TypeOf(value)
	if t.Kind() != reflect.Ptr {
		ret = reflect.New(t).Interface()
	} else {
		// reflected pointer
		ret = reflect.New(t.Elem()).Interface()
	}

	if exists = c.Get(name, ret); exists {
		return
	} else if err = c.Put(name, value, lifespan); err != nil {
		return
	}
	ret = value
	return
}

// delete cached value by key.
func (c *RedisCache) Delete(key string) error {
	_, err := c.conn.Do("del", key)
	return err
}

// increase cached int value by key, as a counter.
func (c *RedisCache) Incr(key string) (uint64, error) {
	numStr, err := c.conn.Do("incr", key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(numStr.(string), 10, 64)
}

// decrease cached int value by key, as a counter.
func (c *RedisCache) Decr(key string) (uint64, error) {
	numStr, err := c.conn.Do("decr", key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseUint(numStr.(string), 10, 64)
}

// check if cached value exists or not.
//IsExist(key string) bool
// clear all cache.
func (c *RedisCache) ClearAll() error {
	return nil
}

// start gc routine based on config string settings.
func (c *RedisCache) StartAndGC(config string) error {
	c.conn = libredis.GetRedis()
	return nil // TODO: Implement
}

// clone object
// get item ver
//GetItemVer(key string) uint64
func (c *RedisCache) Clone(key string, obj interface{}) bool {
	return c.Get(key, obj)
}

func init() {
	Register("redis", NewRedisCache)
}
