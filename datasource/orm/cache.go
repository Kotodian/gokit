package orm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Kotodian/gokit/datasource/orm/cache"
)

// MC 缓存
var MC cache.Cache
var cacheEnabled map[string]bool

//var syncList sync.Map

// InitCache 初始化缓存
func InitCache(c cache.Cache) {
	MC = c
	cacheEnabled = make(map[string]bool, 10)
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
func SaveToMC(obj IBase, ttl time.Duration) {
	if obj.GetCacheKey() == "" {
		return
	}
	SetSearchGroupCacheID(obj)
	if keyName, ok := getCacheEnabled(obj); ok {
		saveToMC(keyName, obj, ttl)
	}
	return
}

// GetByMC 在内存中获取对象
func GetByMC(obj IBase) (err error) {
	keyName, ok := getCacheEnabled(obj)
	if ok {
		if MC.Clone(keyName, obj) {
			return nil
		}
	}
	return errors.New("not found")
}

// DeleteFromMC ...
func DeleteFromMC(obj IBase) {
	if obj.GetCacheKey() == "" {
		return
	}
	SetSearchGroupCacheID(obj)
	if keyName, ok := getCacheEnabled(obj); ok {
		_ = MC.Delete(keyName)
	}
}

//GetCacheKeyName 第一个返回值为真实存储数据的key，第二个返回值为获取真实存储数据的key的key
func GetCacheKeyName(obj IBase, cond string, where ...interface{}) string {
	gid, sid, err := getGroupTableCacheID(obj).Get()
	if err != nil {
		return ""
	}
	return getGroupCacheKeyName(obj, gid, sid, fmt.Sprintf(strings.Replace(cond, "?", "%v", -1), where...))
	//return utils.MD5(fmt.Sprintf("%s:%d:%d:name:gcache:orm",
	//	obj.TableName()+fmt.Sprintf(strings.Replace(cond, "?", "%v", -1), where...),
	//	groupTableID))
}

// saveToMC ...
func saveToMC(keyName string, obj IBase, t time.Duration) {
	obj.SetIsExists(true)
	_ = MC.Put(keyName, obj, t)
}

func getCacheEnabled(obj IBase, ID ...interface{}) (name string, ok bool) {
	if cacheEnabled == nil {
		return "", false
	}

	//如果有id这个字段，就给id赋值，在处理不存在的记录时候，减少外部赋值ID的工作量
	if ID != nil {
		s := db.NewScope(obj)
		if s.HasColumn("id") {
			_ = s.SetColumn("id", ID[0])
		}
	}

	keyName := obj.GetCacheKey()
	if keyName == "" {
		return "", false
	}

	gid, _, err := getGroupTableCacheID(obj).Get()
	if err != nil {
		return "", false
	}
	return getGroupCacheKeyNameByID(obj, gid, keyName), true
}

func getCacheEnabledWithContext(ctx context.Context, obj IBase, ID ...interface{}) (name string, ok bool) {
	db := GetDBWithContext(ctx)
	if cacheEnabled == nil {
		return "", false
	}

	if ID != nil {
		s := db.NewScope(obj)
		if s.HasColumn("id") {
			_ = s.SetColumn("id", ID[0])
		}
	}

	keyName := obj.GetCacheKey()
	if keyName == "" {
		return "", false
	}
	gid, _, err := getGroupTableCacheID(obj).Get()
	if err != nil {
		return "", false
	}
	return getGroupCacheKeyNameByID(obj, gid, keyName), true
}

//
//// GetCacheList ...
//func GetCacheList(obj IBase) *container.List {
//	l, _ := syncList.LoadOrStore(obj.GetCacheKey(), container.NewList())
//	return l.(*container.List)
//}
//
//// SaveToIBaseCacheList ...
//func SaveToIBaseCacheList(obj IBase, ttl time.Duration, associateKey ...string) {
//	l := GetCacheList(obj)
//	for _, v := range associateKey {
//		_ = MC.Put(v, obj, ttl)
//		l.PushBack(v)
//	}
//}

//
//// SetGroupCacheID ...
//func SetGroupCacheID(obj IBase) {
//	deleteSyncList(obj)
//	IncrGroupIBaseCacheID(obj)
//}
//
//func deleteSyncList(obj IBase) {
//	DeleteGroupIBaseCacheID(obj)
//	GetCacheList(obj).Clear(func(i interface{}) {
//		_ = MC.Delete(i.(string))
//	})
//	syncList.Delete(obj.GetCacheKey())
//}

//---------------------------
