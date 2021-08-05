package telemetrybuffer

import (
	"context"
	"sync"
	"time"

	"github.com/Kotodian/gokit/container"
	"github.com/yangwenmai/ratelimit/simpleratelimit"
)

type SyncFn func(key, value interface{}) bool

type Buffer struct {
	l            sync.RWMutex
	map1         sync.Map
	map2         sync.Map
	current      *sync.Map
	syncInterval time.Duration
	rateLimit    *simpleratelimit.RateLimiter
	syncSize     uint
	flush        chan struct{}
	syncFn       SyncFn
}

func NewBuffer(n uint, t time.Duration, fn SyncFn) *Buffer {
	m := &Buffer{
		syncInterval: t,
		syncSize:     n,
		flush:        make(chan struct{}, 100),
		syncFn:       fn,
	}
	m.current = &m.map1
	m.rateLimit = simpleratelimit.New(int(n), t)
	return m
}

func (m *Buffer) Flush() {
	m.flush <- struct{}{}
}

func (m *Buffer) Store(key interface{}, val interface{}) {
	if m.rateLimit.Limit() {
		m.rateLimit.Undo()
		m.Flush()
	}

	m.l.RLock()
	defer m.l.RUnlock()

	_list, _ := m.current.LoadOrStore(key, container.NewList())
	_list.(*container.List).PushBack(val)

}

func (m *Buffer) LoadOrStore(key interface{}, val interface{}) {
	m.l.RLock()
	defer m.l.RUnlock()
	m.current.LoadOrStore(key, val)
}

func (m *Buffer) Remove(key interface{}) {
	m.l.Lock()
	defer m.l.Unlock()
	m.current.Delete(key)
}

func (m *Buffer) Swap() (old *sync.Map) {
	m.l.Lock()
	defer m.l.Unlock()
	old = m.current
	if m.current == &m.map1 {
		m.current = &m.map2
	} else {
		m.current = &m.map1
	}
	return old
}

func (m *Buffer) Current() *sync.Map {
	m.l.RLock()
	defer m.l.RUnlock()
	return m.current
}

func (m *Buffer) RangeAndSwap(f func(key, value interface{}) bool) bool {
	//m.l.Lock()
	//defer m.l.Unlock()
	old := m.Swap()
	var ret bool
	old.Range(func(key, value interface{}) bool {
		ret = f(key, value)
		return ret
	})
	if !ret {
		return ret
	}
	if old == &m.map1 {
		m.map1 = sync.Map{}
	} else {
		m.map2 = sync.Map{}
	}
	old = m.Swap()
	old.Range(func(key, value interface{}) bool {
		ret = f(key, value)
		return ret
	})
	if !ret {
		return ret
	}
	if old == &m.map1 {
		m.map1 = sync.Map{}
	} else {
		m.map2 = sync.Map{}
	}
	return true
}

func (m *Buffer) Daemon(ctx context.Context) {
	t := time.NewTicker(m.syncInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			m.RangeAndSwap(m.syncFn)
		case <-m.flush:
			m.RangeAndSwap(m.syncFn)
		}
	}
}
