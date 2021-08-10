package orm

import (
	"fmt"
	"sync"

	"github.com/Kotodian/gokit/id"
	"github.com/Kotodian/gokit/utils"
	"go.uber.org/atomic"
)

type CacheVer struct {
	//ver    *uint64       //主版本号
	//search *uint64       //搜索版本号
	ver    atomic.Uint64 //主版本号
	search atomic.Uint64 //搜索版本号
	//l      sync.RWMutex
}

func (c *CacheVer) Get() (uint64, uint64, error) {
	//c.l.Lock()
	//defer c.l.Unlock()
	return c.load()
}

func (c *CacheVer) load() (ver uint64, search uint64, err error) {
	if ver = c.ver.Load(); ver == 0 {
		ver = uint64(id.Next())
		if err != nil {
			return 0, 0, err
		}
		c.ver.Store(ver)
		//c.ver = &ver
	}
	if search = c.search.Load(); ver == 0 {
		search = id.Next().Uint64()
		if err != nil {
			return 0, 0, err
		}
		c.search.Store(search)
	}
	return ver, search, nil
}

func (c *CacheVer) Incr() error {
	if c.ver.Add(1) == 1 {
		c.ver.Add(id.Next().Uint64())
	}
	return nil
	//c.l.Lock()
	//defer c.l.Unlock()
	//
	//ver, _, err := c.load()
	//if err != nil {
	//	return nil
	//}
	//ver += 1
	//c.ver = &ver
}

func (c *CacheVer) IncrSearch() error {
	if c.search.Add(1) == 1 {
		c.search.Store(id.Next().Uint64())
	}
	return nil
}

var syncTableVer sync.Map

//
// getGroupTableCacheID
// 获取表的版本
func getGroupTableCacheID(obj IBase) *CacheVer {
	key := utils.MD5(fmt.Sprintf("%s:ver:table:gcache:orm", obj.TableName()))
	i, _ := syncTableVer.LoadOrStore(key, &CacheVer{})
	return i.(*CacheVer)
}

// delTableGroupCache
func delTableGroupCache(obj IBase) {
	//ver, _ := idGenr.Next()
	syncTableVer.Delete(utils.MD5(fmt.Sprintf("%s:ver:table:gcache:orm", obj.TableName())))
}

func SetSearchGroupCacheID(obj IBase) {
	if err := getGroupTableCacheID(obj).IncrSearch(); err != nil {
		delTableGroupCache(obj)
	}
}

func SetGroupCacheID(obj IBase) {
	if err := getGroupTableCacheID(obj).Incr(); err != nil {
		delTableGroupCache(obj)
	}
}

func getGroupCacheKeyNameByID(obj IBase, gid uint64, mark ...string) string {
	var key string
	if len(mark) == 0 {
		key = fmt.Sprintf("%d:%s:gcache:orm", gid, obj.TableName())
	} else {
		key = fmt.Sprintf("%s:%d:%s:gcache:orm", mark[0], gid, obj.TableName())
	}
	return utils.MD5(key)
}

func getGroupCacheKeyName(obj IBase, gid, sid uint64, mark ...string) string {
	var key string
	if len(mark) == 0 {
		key = fmt.Sprintf("%d:%d:%s:gcache:orm", sid, gid, obj.TableName())
	} else {
		key = fmt.Sprintf("%s:%d:%d:%s:gcache:orm", mark[0], sid, gid, obj.TableName())
	}
	return utils.MD5(key)
}

//
//// IncrGroupTableCacheID
//// 增加表的版本，主要处理数据一致性的问题
//func IncrGroupTableCacheID(obj IBase) uint64 {
//	i := GetGroupTableCacheID(obj)
//	n := atomic.AddUint64(&i, 1)
//	syncTableVer.Store(utils.MD5(fmt.Sprintf("%s:ver:table:gcache:orm", obj.TableName())), n)
//	return n
//}

//
//// GetGroupIBaseCacheID
//// 返回表的组ID以及对象的版本ID
//func GetGroupIBaseCacheID(obj IBase) (uint64, uint64) {
//	gid := GetGroupTableCacheID(obj)
//	key := utils.MD5(fmt.Sprintf("%d:%s:gcache:orm", gid, obj.GetCacheKey()))
//	i, _ := syncTableVer.LoadOrStore(key, (<-UUID).Uint64())
//	return gid, i.(uint64)
//}

//
//// IncrGroupIBaseCacheID
//// 增加对象的版本号
//func IncrGroupIBaseCacheID(obj IBase) uint64 {
//	if obj.IsNew() {
//		//如果是新的对象，需要把整个表的cache删除，这样才能解决列表数据一致性的问题
//		gid := IncrGroupTableCacheID(obj)
//		n := (<-UUID).Uint64()
//		syncTableVer.Store(utils.MD5(fmt.Sprintf("%d:%s:gcache:orm", gid, obj.GetCacheKey())), n)
//		return n
//	}
//	gid, ver := GetGroupIBaseCacheID(obj)
//	newVer := atomic.AddUint64(&ver, 1)
//	syncTableVer.Store(utils.MD5(fmt.Sprintf("%d:%s:gcache:orm", gid, obj.GetCacheKey())), newVer)
//	return newVer
//}
//
//// DeleteGroupIBaseCacheID
//func DeleteGroupIBaseCacheID(obj IBase) {
//	gid := GetGroupTableCacheID(obj)
//	key := utils.MD5(fmt.Sprintf("%d:%s:gcache:orm", gid, obj.GetCacheKey()))
//	syncTableVer.Delete(key)
//}
