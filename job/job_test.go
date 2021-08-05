package job

import (
	"fmt"
	"sync"
	"testing"
)

func Test_Job(t *testing.T) {
	len := 10
	wg := sync.WaitGroup{}
	wg.Add(len)
	for i := 0; i < len; i++ {
		go func() {
			defer wg.Done()
			_job := GetJob("111111")
			_job.Release()
		}()
	}

	wg.Wait()
	fmt.Println("DDDDDDDDDDDDDDDDD")
	j := GetJob("111111")
	fmt.Println(j.GetRef())
	j.Release()
	fmt.Println(j.GetRef())
}
