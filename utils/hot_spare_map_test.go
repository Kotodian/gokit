package utils

import (
	"fmt"
	"sync"
	"testing"
)

func TestHotSpareMap(t *testing.T) {
	m := NewHotSpareMap()

	w := sync.WaitGroup{}
	w.Add(100)
	for i := 0; i < 100; i++ {
		go func(i int) {
			defer w.Done()
			if i == 50 || i == 30 {
				m.Swap()
			}
			m.Store(i, i)
		}(i)
	}
	w.Wait()
	y := make([]int, 0)
	m.RangeAndSwap(func(key, val interface{}) bool {
		//y[key.(int)] = val.(int)
		y = append(y, key.(int))
		//if key.(int) == 99 {
		//	return false
		//}
		return true
	})
	total := 0
	for _, v := range y {
		total += v
	}
	m.Current().Range(func(key, value interface{}) bool {
		fmt.Println(key, value)
		return true
	})
	if total != 4950 {
		t.Error("并发写错误")
	}
}
