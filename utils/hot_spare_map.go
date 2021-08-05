package utils

import "sync"

type HotSpareMap struct {
	l       sync.RWMutex
	map1    sync.Map
	map2    sync.Map
	current *sync.Map
}

func NewHotSpareMap() *HotSpareMap {
	m := &HotSpareMap{}
	m.current = &m.map1
	return m
}

func (m *HotSpareMap) Store(key interface{}, val interface{}) {
	m.l.Lock()
	defer m.l.Unlock()
	m.current.Store(key, val)
}

func (m *HotSpareMap) LoadOrStore(key interface{}, val interface{}) {
	m.l.Lock()
	defer m.l.Unlock()
	m.current.LoadOrStore(key, val)
}

func (m *HotSpareMap) Remove(key interface{}) {
	m.l.Lock()
	defer m.l.Unlock()
	m.current.Delete(key)
}

func (m *HotSpareMap) Swap() (old *sync.Map) {
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

func (m *HotSpareMap) Current() *sync.Map {
	m.l.RLock()
	defer m.l.RUnlock()
	return m.current
}

func (m *HotSpareMap) RangeAndSwap(f func(key, value interface{}) bool) bool {
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
