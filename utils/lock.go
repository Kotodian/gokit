package utils

import (
	"sync"
	"sync/atomic"
)

var LocksMap sync.Map

type Lock struct {
	key   string
	l     *sync.RWMutex
	count int32
}

func (l *Lock) RUnlock() {
	l.l.RUnlock()
	atomic.AddInt32(&l.count, -1)
	// fmt.Println("fffffffffff", RandInt64(0, 2))
	if l.count <= 0 {
		LocksMap.Delete(l.key)
	}
}

func (l *Lock) RLock() {
	l.l.RLock()
	atomic.AddInt32(&l.count, 1)
}

func (l *Lock) Lock() {
	atomic.AddInt32(&l.count, 1)
	l.l.Lock()
	// return atomic.LoadInt32(&l.count)
}

func (l *Lock) Unlock() {
	l.l.Unlock()
	atomic.AddInt32(&l.count, -1)
	// fmt.Println("fffffffffff", RandInt64(0, 2))
	if l.count <= 0 {
		LocksMap.Delete(l.key)
	}
}

func (l *Lock) GetRefCount() int32 {
	return atomic.LoadInt32(&l.count)
}

func NewLock(key string) *Lock {
	data, _ := LocksMap.LoadOrStore(key, &Lock{
		key: key,
		l:   new(sync.RWMutex),
	})
	l := data.(*Lock)
	// atomic.AddInt32(&l.count, 1)
	return l
}

func HasLock(key string) bool {
	data, b := LocksMap.Load(key)
	if !b {
		return false
	}
	return data.(*Lock).count >= 1
}
