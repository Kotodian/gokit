package simple

import (
	"sort"
	"sync"
	"time"

	"github.com/Kotodian/gokit/cache"
)

type simpleCache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]*entry[V]
}

func NewCache[K comparable, V any]() cache.Interface[K, V] {
	return &simpleCache[K, V]{
		items: make(map[K]*entry[V], 0),
	}
}

func (c *simpleCache[K, V]) Set(key K, val V) {
	c.mu.Lock()
	c.items[key] = &entry[V]{val: val, createdAt: cache.NowFunc()}
	c.mu.Unlock()
}

func (c *simpleCache[K, V]) Get(key K) (val V, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	got, found := c.items[key]
	if !found {
		return
	}
	return got.val, true
}

func (c *simpleCache[K, V]) Keys() []K {
	c.mu.Lock()
	ret := make([]K, 0, len(c.items))
	for key := range c.items {
		ret = append(ret, key)
	}
	sort.Slice(ret, func(i, j int) bool {
		return c.items[ret[i]].createdAt.Before(c.items[ret[j]].createdAt)
	})
	c.mu.Unlock()
	return ret
}

func (c *simpleCache[K, V]) Delete(key K) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

type entry[V any] struct {
	val       V
	createdAt time.Time
}
