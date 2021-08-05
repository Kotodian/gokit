package container

import (
	"container/list"
	"sync"
)

type List struct {
	data *list.List
	l    sync.RWMutex
	keys map[string]struct{}
}

func NewList() *List {
	l := new(List)
	l.data = list.New()
	return l
}

func (l *List) PushBack(v interface{}) {
	l.l.Lock()
	defer l.l.Unlock()
	if k, ok := v.(string); ok {
		if _, ok := l.keys[k]; !ok {
			l.data.PushBack(v)
		}
	} else {
		l.data.PushBack(v)
	}
}

func (l *List) Clear(fn ...func(interface{})) {
	l.l.Lock()
	defer l.l.Unlock()

	if len(fn) > 0 {
		for {
			e := l.data.Front()
			if e == nil {
				break
			} else if k, ok := e.Value.(string); ok {
				delete(l.keys, k)
			}
			fn[0](l.data.Remove(e))
		}
	} else {
		for {
			e := l.data.Front()
			if e == nil {
				break
			} else if k, ok := e.Value.(string); ok {
				delete(l.keys, k)
			}
			l.data.Remove(e)
		}
	}
}

func (l *List) GetData() *list.List {
	l.l.RLock()
	defer l.l.RUnlock()
	return l.data
}

func (l *List) Len() int {
	l.l.RLock()
	defer l.l.RUnlock()
	return l.data.Len()
}
