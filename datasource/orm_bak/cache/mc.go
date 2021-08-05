// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package cache

import (
	"encoding/json"
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/Kotodian/gokit/datasource/orm/utils"
)

var (
	// DefaultEvery means the clock time of recycling the expired cache items in memory.
	DefaultEvery = 60 // 1 minute
)

// MemoryItem store memory cache item.
type MemoryItem struct {
	val         interface{}
	createdTime time.Time
	lifespan    time.Duration
	ver         uint64
}

func (mi *MemoryItem) isExpire() bool {
	// 0 means forever
	if mi.lifespan == 0 {
		return false
	}
	return time.Now().Sub(mi.createdTime) > mi.lifespan
}

// MemoryCache is Memory cache adapter.
// it contains a RW locker for safe map storage.
type MemoryCache struct {
	sync.RWMutex
	dur   time.Duration
	items map[string]*MemoryItem
	Every int // run an expiration check Every clock time
}

// NewMemoryCache returns a new MemoryCache.
func NewMemoryCache() Cache {
	cache := MemoryCache{items: make(map[string]*MemoryItem)}
	return &cache
}

// Get cache from memory.
// if non-existed or expired, return nil.
func (bc *MemoryCache) Get(name string, ret interface{}) bool {
	bc.RLock()
	defer bc.RUnlock()
	if itm, ok := bc.items[name]; ok {
		if itm.isExpire() {
			return false
		}
		v := reflect.ValueOf(itm.val)
		n := reflect.New(v.Type())
		reflect.ValueOf(n.Interface()).Elem().Set(v)
		obj := n.Elem().Interface()
		if _obj, ok := obj.(ICacheVer); ok {
			_obj.SetCacheVer(itm.ver)
		}
		reflect.ValueOf(ret).Elem().Set(reflect.ValueOf(obj))
		return true

	}
	return false
}

//
//func (bc *MemoryCache) GetItemVer(name string) uint64 {
//	bc.RLock()
//	defer bc.RUnlock()
//	if itm, ok := bc.items[name]; ok {
//		return itm.ver
//	}
//	return uint64(0)
//}

func (bc *MemoryCache) Clone(name string, obj interface{}) (exists bool) {
	bc.RLock()
	defer bc.RUnlock()
	if itm, ok := bc.items[name]; ok {
		if itm.isExpire() {
			return
		}
		utils.CloneValue(itm.val, obj)
		exists = true
		if _obj, ok := obj.(ICacheVer); ok {
			_obj.SetCacheVer(itm.ver)
		}
		return
	}
	return
}

//
//// GetMulti gets caches from memory.
//// if non-existed or expired, return nil.
//func (bc *MemoryCache) GetMulti(names []string) []interface{} {
//	var rc []interface{}
//	for _, name := range names {
//		rc = append(rc, bc.Get(name))
//	}
//	return rc
//}

// GetOrStore store value to memory if not exists / get value from memory if exists
func (bc *MemoryCache) GetOrStore(name string, value interface{}, lifespan time.Duration) (interface{}, bool) {
	bc.Lock()
	defer bc.Unlock()

	if v, ok := bc.items[name]; ok && !v.isExpire() {
		return v.val, false
		//return !v.isExpire()
	}
	ref := reflect.ValueOf(value)
	var v interface{}
	if ref.Kind() == reflect.Ptr {
		v = ref.Elem().Interface()
	} else {
		v = value
	}
	ver := uint64(0)
	if _obj, ok := value.(ICacheVer); ok {
		if b, _ok := bc.items[name]; _ok {
			ver = b.ver
			if b.ver != _obj.GetCacheVer() { //版本号不一样就删除
				delete(bc.items, name)
				//return value, true
			}
		}
		ver = uint64(time.Now().UnixNano())
		defer _obj.SetCacheVer(ver)
	}
	bc.items[name] = &MemoryItem{
		val:         v,
		createdTime: time.Now(),
		lifespan:    lifespan,
		ver:         ver,
	}
	return value, true
}

// Put cache to memory.
// if lifespan is 0, it will be forever till restart.
// if force is false, will check the ver.
func (bc *MemoryCache) Put(name string, value interface{}, lifespan time.Duration) error {
	bc.Lock()
	defer bc.Unlock()
	ref := reflect.ValueOf(value)
	var v interface{}
	if ref.Kind() == reflect.Ptr {
		v = ref.Elem().Interface()
	} else {
		v = value
	}
	ver := uint64(0)
	if _obj, ok := value.(ICacheVer); ok {
		if b, _ok := bc.items[name]; _ok {
			ver = b.ver
			if b.ver != _obj.GetCacheVer() { //版本号不一样就删除
				delete(bc.items, name)
				return nil
			}
		}
		ver = uint64(time.Now().UnixNano())
		defer _obj.SetCacheVer(ver)
	}
	bc.items[name] = &MemoryItem{
		val:         v,
		createdTime: time.Now(),
		lifespan:    lifespan,
		ver:         ver,
	}
	return nil
}

// Delete cache in memory.
func (bc *MemoryCache) Delete(name string) error {
	bc.Lock()
	defer bc.Unlock()
	if _, ok := bc.items[name]; !ok {
		return errors.New("key not exist")
	}
	delete(bc.items, name)
	if _, ok := bc.items[name]; ok {
		return errors.New("delete key error")
	}
	return nil
}

// Incr increase cache counter in memory.
// it supports int,int32,int64,uint,uint32,uint64.
func (bc *MemoryCache) Incr(key string) (id uint64, err error) {
	bc.Lock()
	defer bc.Unlock()

	itm, ok := bc.items[key]
	if !ok {
		err = errors.New("key not exist")
		return
	}
	switch itm.val.(type) {
	case int:
		itm.val = itm.val.(int) + 1
		return uint64(itm.val.(int)), nil
	case int32:
		itm.val = itm.val.(int32) + 1
		return uint64(itm.val.(int32)), nil
	case int64:
		itm.val = itm.val.(int64) + 1
		return uint64(itm.val.(int64)), nil
	case uint:
		itm.val = itm.val.(uint) + 1
		return uint64(itm.val.(uint)), nil
	case uint32:
		itm.val = itm.val.(uint32) + 1
		return uint64(itm.val.(uint32)), nil
	case uint64:
		itm.val = itm.val.(uint64) + 1
		return itm.val.(uint64), nil
	default:
		err = errors.New("item val is not (u)int (u)int32 (u)int64")
	}
	return
	//return itm.val.(int), nil
}

// Decr decrease counter in memory.
func (bc *MemoryCache) Decr(key string) (id uint64, err error) {
	bc.Lock()
	defer bc.Unlock()
	itm, ok := bc.items[key]
	if !ok {
		err = errors.New("key not exist")
		return
	}
	switch itm.val.(type) {
	case int:
		itm.val = itm.val.(int) - 1
		return uint64(itm.val.(int)), nil
	case int32:
		itm.val = itm.val.(int32) - 1
		return uint64(itm.val.(int32)), nil
	case int64:
		itm.val = itm.val.(int64) - 1
		return uint64(itm.val.(int64)), nil
	case uint:
		itm.val = itm.val.(uint) - 1
		return uint64(itm.val.(uint)), nil
	case uint32:
		itm.val = itm.val.(uint32) - 1
		return uint64(itm.val.(uint32)), nil
	case uint64:
		itm.val = itm.val.(uint64) - 1
		return itm.val.(uint64), nil
	default:
		err = errors.New("item val is not (u)int (u)int32 (u)int64")
	}
	return
}

// IsExist check cache exist in memory.
func (bc *MemoryCache) IsExist(name string) bool {
	bc.RLock()
	defer bc.RUnlock()
	if v, ok := bc.items[name]; ok {
		return !v.isExpire()
	}
	return false
}

// ClearAll will delete all cache in memory.
func (bc *MemoryCache) ClearAll() error {
	bc.Lock()
	defer bc.Unlock()
	bc.items = make(map[string]*MemoryItem)
	return nil
}

// StartAndGC start memory cache. it will check expiration in every clock time.
func (bc *MemoryCache) StartAndGC(config string) error {
	var cf map[string]int
	if err := json.Unmarshal([]byte(config), &cf); err != nil {
		return err
	}
	if _, ok := cf["interval"]; !ok {
		cf = make(map[string]int)
		cf["interval"] = DefaultEvery
	}
	dur := time.Duration(cf["interval"]) * time.Second
	bc.Every = cf["interval"]
	bc.dur = dur
	go bc.vacuum()
	return nil
}

// check expiration.
func (bc *MemoryCache) vacuum() {
	bc.RLock()
	every := bc.Every
	bc.RUnlock()

	if every < 1 {
		return
	}
	for {
		<-time.After(bc.dur)
		if bc.items == nil {
			return
		}
		if keys := bc.expiredKeys(); len(keys) != 0 {
			bc.clearItems(keys)
		}
	}
}

// expiredKeys returns key list which are expired.
func (bc *MemoryCache) expiredKeys() (keys []string) {
	bc.RLock()
	defer bc.RUnlock()
	for key, itm := range bc.items {
		if itm.isExpire() {
			keys = append(keys, key)
		}
	}
	return
}

// clearItems removes all the items which key in keys.
func (bc *MemoryCache) clearItems(keys []string) {
	bc.Lock()
	defer bc.Unlock()
	for _, key := range keys {
		delete(bc.items, key)
	}
}

func init() {
	Register("mc", NewMemoryCache)
}
