package memcached

import (
	//"github.com/bradfitz/gomemcache/memcache"
	"encoding/json"

	"github.com/bradfitz/gomemcache/memcache"
)

var MC *memcache.Client

func init() {
	MC = memcache.New("memcached:11211")
	if MC == nil {
		panic("memcached connect fail")
	}
	//mc := memcache.New("memcached:11211")
}

func Get(key string, data interface{}) (ok bool, err error) {
	var item *memcache.Item
	if item, err = MC.Get(key); err != nil {
		if err != memcache.ErrCacheMiss {
			return
		} else {
			err = nil
			return
		}
	} else if item != nil {
		if err = json.Unmarshal(item.Value, &data); err != nil {
			return
		}
		ok = true
		return
	}
	return
}

func Save(key string, data interface{}, Expiration int32) (err error) {
	item := &memcache.Item{Key: key}
	item.Expiration = Expiration
	if item.Value, err = json.Marshal(data); err != nil {
		return
	} else if err = MC.Set(item); err != nil {
		return
	}
	return
}
