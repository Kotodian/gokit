package channel

import (
	"fmt"
	"sync"
)

var mapchs sync.Map

func NewCH(name string, nums ...int) (chan interface{}, bool) {
	num := 1
	if len(nums) > 0 {
		num = nums[0]
	}
	ch, loaded := mapchs.LoadOrStore(name, make(chan interface{}, num))
	return ch.(chan interface{}), loaded

}

func IsExists(name string) (ok bool) {
	_, ok = mapchs.Load(name)
	return
}

func SetCH(name string, s interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("set ch error:[%v]", r)
		}
	}()

	ch, ok := mapchs.Load(name)
	if !ok || ch == nil {
		return fmt.Errorf("name:[%s] ch not found", name)
	}

	select {
	case ch.(chan interface{}) <- s:
		return nil
	}
}

func DeleteCH(names ...string) {
	for _, name := range names {
		mapchs.Delete(name)
	}
}
func CloseCH(chs ...chan interface{}) {
	for _, ch := range chs {
		if !IsColseCH(ch) {
			close(ch)
		}
	}
}

func IsColseCH(ch <-chan interface{}) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}
