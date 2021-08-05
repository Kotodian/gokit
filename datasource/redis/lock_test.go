package redis

import (
	"fmt"
	"sync"
	"testing"
)

func Test_Lock(t *testing.T) {
	wg := sync.WaitGroup{}
	n := 50
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			//_rd := GetRedis()
			//defer _rd.Close()

			fmt.Println(i, " try lock", GetPool().ActiveCount(), GetPool().IdleCount())
			//time.Sleep(1 * time.Second)
			l, val, err := TryLock("d:ggi", 3)
			if err != nil {
				fmt.Printf("%d %d %d get lock error,err:%s\n", i, GetPool().ActiveCount(), GetPool().IdleCount(), err.Error())
			} else {
				defer Unlock("d:ggi", val)
				fmt.Printf("%d %d %d get lock:%v ver:%d !!!!!!!!!!!!!!!!!!!!!!\n", i, GetPool().ActiveCount(), GetPool().IdleCount(), l, val)
			}

			//time.Sleep(1 * time.Second)
		}(i)
	}
	wg.Wait()
}
