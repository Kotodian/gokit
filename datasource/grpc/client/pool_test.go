package client

import (
	"log"
	"sync"
	"testing"
	"time"
)

func Test_Pool(t *testing.T) {
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cc, _ := Conn()
			time.Sleep(3 * time.Second)
			Close(cc)
			log.Printf(" conns len:[%d] \r\n", p.Len())
			// time.Sleep(1 * time.Second)
		}()
	}

	wg.Wait()
	log.Printf(" conns len:[%d] \r\n", p.Len())
}
