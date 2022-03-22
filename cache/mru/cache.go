package mru

import (
	"container/list"
	"sync"

	"github.com/Kotodian/gokit/cache"
)

type mruCache[K comparable, V any] struct {
	mu    sync.RWMutex
	cap   int
	list  *list.List
	items map[K]*list.Element
}

type entry[K comparable, V any] struct {
	key K
	val V
}

func NewCache[K comparable, V any](capacity int) cache.Interface[K, V] {
	return &mruCache[K, V]{
		cap:   capacity,
		list:  list.New(),
		items: make(map[K]*list.Element, capacity),
	}
}

func (c *mruCache[K, V]) Get(key K) (val V, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok {
		return
	}
	// updates cache order
	c.list.MoveToBack(e)
	return e.Value.(*entry[K, V]).val, true
}

func (c *mruCache[K, V]) Set(key K, val V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e, ok := c.items[key]; ok {
		c.list.MoveToBack(e)
		entry := e.Value.(*entry[K, V])
		entry.val = val
		return
	}
	if c.list.Len() == c.cap {
		c.deleteNewest()
	}

	newEntry := &entry[K, V]{
		key: key,
		val: val,
	}
	e := c.list.PushBack(newEntry)
	c.items[key] = e
}

func (c *mruCache[K, V]) deleteNewest() {
	c.delete(c.list.Front())
}

func (c *mruCache[K, V]) Keys() []K {
	c.mu.Lock()
	keys := make([]K, 0, len(c.items))
	for ent := c.list.Back(); ent != nil; ent = ent.Prev() {
		entry := ent.Value.(*entry[K, V])
		keys = append(keys, entry.key)
	}
	c.mu.Unlock()
	return keys
}

func (c *mruCache[K, V]) Delete(key K) {
	if item, ok := c.items[key]; ok {
		c.delete(item)
	}
}

func (c *mruCache[K, V]) delete(e *list.Element) {
	c.list.Remove(e)
	entry := e.Value.(*entry[K, V])
	delete(c.items, entry.key)
}
