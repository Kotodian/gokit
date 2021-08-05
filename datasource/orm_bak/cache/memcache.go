package cache

import (
	"encoding/json"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

// MemcacheCache is Memory cache adapter.
// it contains a RW locker for safe map storage.
type MemcacheCache struct {
	client *memcache.Client
	prefix string
}

// NewMemcacheCache returns a new MemcacheCache.
func NewMemcacheCache() Cache {
	cache := MemcacheCache{}
	return &cache
}

func (bc *MemcacheCache) getKey(name string) string {
	return fmt.Sprintf("%s:%s", name, bc.prefix)
}

// Get cache from memory.
// if non-existed or expired, return nil.
func (bc *MemcacheCache) Get(name string, res interface{}) (ret bool) {
	var item *memcache.Item
	var err error
	defer func() {
		if err != nil {
			ret = false
			logrus.Errorf("get from memcache(key:%s) error, %s", name, err)
		}
	}()
	if item, err = bc.client.Get(bc.getKey(name)); err != nil {
		if err != memcache.ErrCacheMiss {
			//fmt.Println("get form cache, key:", name, " miss")
			return
		} else {
			err = nil
			return
		}
	}

	if item != nil {
		if err = json.Unmarshal(item.Value, res); err != nil {
			return
		}
		if _obj, ok := res.(ICacheExists); ok {
			_obj.SetExists()
		}
		ret = true
		return
	}
	return
}

func (bc *MemcacheCache) Clone(name string, obj interface{}) (exists bool) {
	return bc.Get(name, obj)
}

// GetOrStore store value to memory if not exists / get value from memory if exists
func (bc *MemcacheCache) GetOrStore(name string, value interface{}, lifespan time.Duration) (ret interface{}, exists bool) {
	var err error
	defer func() {
		if err != nil {
			logrus.Error("save to memcache(key:%s) error, %s", name, err)
			exists = false
			ret = value
		}
	}()

	//Original struct
	t := reflect.TypeOf(value)
	if t.Kind() != reflect.Ptr {
		ret = reflect.New(t).Interface()
	} else {
		// reflected pointer
		ret = reflect.New(t.Elem()).Interface()
	}

	if exists = bc.Get(name, ret); exists {
		return
	} else if err = bc.Put(name, value, lifespan); err != nil {
		return
	}
	ret = value
	return
}

// Put cache to memory.
// if lifespan is 0, it will be forever till restart.
// if force is false, will check the ver.
func (bc *MemcacheCache) Put(name string, value interface{}, lifespan time.Duration) (err error) {
	var val []byte

	if val, err = json.Marshal(value); err != nil {
		return
	}
	//fmt.Println("put to memcached", name, string(val))
	if err = bc.client.Set(&memcache.Item{
		Key:        bc.getKey(name),
		Value:      val,
		Expiration: int32(lifespan.Seconds()),
	}); err != nil {
		return
	}
	return
}

// Delete cache in memory.
func (bc *MemcacheCache) Delete(name string) error {
	return bc.client.Delete(bc.getKey(name))
}

// Incr increase cache counter in memory.
// it supports int,int32,int64,uint,uint32,uint64.
func (bc *MemcacheCache) Incr(key string) (id uint64, err error) {
	return bc.client.Increment(bc.getKey(key), 1)
}

// Decr decrease counter in memory.
func (bc *MemcacheCache) Decr(key string) (id uint64, err error) {
	return bc.client.Decrement(bc.getKey(key), 1)
}

//
//// IsExist check cache exist in memory.
//func (bc *MemcacheCache) IsExist(name string) bool {
//	return bc.client.Replace()
//	bc.RLock()
//	defer bc.RUnlock()
//	if v, ok := bc.items[name]; ok {
//		return !v.isExpire()
//	}
//	return false
//}

// ClearAll will delete all cache in memory.
func (bc *MemcacheCache) ClearAll() error {
	return bc.client.FlushAll()
	//bc.Lock()
	//defer bc.Unlock()
	//bc.items = make(map[string]*MemoryItem)
	//return nil
}

// StartAndGC start memory cache. it will check expiration in every clock time.
func (bc *MemcacheCache) StartAndGC(config string) error {
	var cf map[string]interface{}
	if err := json.Unmarshal([]byte(config), &cf); err != nil {
		return err
	}
	bc.client = memcache.New(cf["server"].(string))
	if maxIdleConns, ok := cf["maxIdleConns"]; ok {
		bc.client.MaxIdleConns = int(maxIdleConns.(float64))
	}
	if timeout, ok := cf["timeout"]; ok {
		bc.client.Timeout = time.Duration(timeout.(float64)) * time.Second
	}
	bc.prefix = cf["prefix"].(string)
	return nil
}

func init() {
	Register("memcache", NewMemcacheCache)
}
