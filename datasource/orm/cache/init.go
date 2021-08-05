package cache

import (
	"fmt"
	"time"
)

// Cache interface contains all behaviors for cache adapter.
// usage:
//	cache.Register("file",cache.NewFileCache) // this operation is run in init method of file.go.
//	c,err := cache.NewCache("file","{....}")
//	c.Put("key",value, 3600 * time.Second)
//	v := c.Get("key")
//
//	c.Incr("counter")  // now is 1
//	c.Incr("counter")  // now is 2
//	count := c.Get("counter").(int)
type Cache interface {
	// get cached value by key.
	//Get(key string) interface{}
	Get(key string, ret interface{}) bool

	// GetMulti is a batch version of Get.
	//GetMulti(keys []string) []interface{}
	// set cached value with key and expire time.
	Put(key string, val interface{}, timeout time.Duration) error

	// GetOrStore
	GetOrStore(name string, value interface{}, lifespan time.Duration) (interface{}, bool)

	// delete cached value by key.
	Delete(key string) error
	// increase cached int value by key, as a counter.
	Incr(key string) (uint64, error)
	// decrease cached int value by key, as a counter.
	Decr(key string) (uint64, error)
	// check if cached value exists or not.
	//IsExist(key string) bool
	// clear all cache.
	ClearAll() error
	// start gc routine based on config string settings.
	StartAndGC(config string) error
	// clone object
	Clone(key string, obj interface{}) bool
	// get item ver
	//GetItemVer(key string) uint64

	//IsEnabled 是否启用缓存
	//IsEnabled(string) bool

	//Enable 启用缓存
	//Enable(string)
}

type ICacheVer interface {
	SetCacheVer(v uint64)
	GetCacheVer() uint64
}

type ICacheExists interface {
	SetExists()
}

// Instance is a function create a new Cache Instance
type Instance func() Cache

var adapters = make(map[string]Instance)

// Register makes a cache adapter available by the adapter name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, adapter Instance) {
	if adapter == nil {
		panic("cache: Register adapter is nil")
	}
	if _, ok := adapters[name]; ok {
		panic("cache: Register called twice for adapter " + name)
	}
	adapters[name] = adapter
}

// NewCache Create a new cache driver by adapter name and config string.
// config need to be correct JSON as string: {"interval":360}.
// it will start gc automatically.
func NewCache(adapterName, config string) (adapter Cache, err error) {
	instanceFunc, ok := adapters[adapterName]
	if !ok {
		err = fmt.Errorf("cache: unknown adapter name %q (forgot to import?)", adapterName)
		return
	}
	adapter = instanceFunc()
	err = adapter.StartAndGC(config)
	if err != nil {
		adapter = nil
	}
	return
}
