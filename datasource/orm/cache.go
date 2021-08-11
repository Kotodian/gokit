package orm

import (
	"errors"
	"fmt"
	"time"

	"github.com/Kotodian/gokit/datasource/orm/cache"
)

// MC 缓存
var MC cache.Cache


//var syncList sync.Map

// InitCache 初始化缓存
func InitCache(c cache.Cache) {
	MC = c
}

// SetCacheEnabled ...
// args[0] engine 存储引擎
// args[1] prefix cache key的前缀，一般用于解决冲突
func SetCacheEnabled(args ...string) {
	if len(args) == 0 {
		args = append(args, "mc")
	}
	var mc cache.Cache
	switch args[0] {
	case "mc":
		mc, _ = cache.NewCache("mc", `{"interval":300}`)
	case "memcache":
		mc, _ = cache.NewCache("memcache", fmt.Sprintf(`{"server":"memcached:11211", "prefix":"%s", "maxIdleConns":100, "timeout":1}`, args[1]))
	}
	InitCache(mc)
}

// SaveToMC 保存对象到内存中
func SaveToMC(obj Object, ttl time.Duration) {
	if obj.Key() == "" {
		return
	}

	saveToMC(obj.Key(), obj, ttl)
	return
}

// GetByMC 在内存中获取对象
func GetByMC(obj Object) (err error) {
	if MC.Clone(obj.Key(), obj) {
		return nil
	}
	return errors.New("not found")
}

// DeleteFromMC ...
func DeleteFromMC(obj Object) {
	if obj.Key() == "" {
		return
	}

	_ = MC.Delete(obj.Key())
}

// saveToMC ...
func saveToMC(keyName string, obj Object, t time.Duration) {
	_ = MC.Put(keyName, obj, t)
}
