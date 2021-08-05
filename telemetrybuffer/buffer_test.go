package telemetrybuffer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Kotodian/gokit/container"
)

func TestAddToBuffer(t *testing.T) {
	b := NewBuffer(2, 3*time.Second, func(key, value interface{}) bool {
		l := value.(*container.List).GetData()
		v := l.Front()
		for i := 0; i < l.Len(); i++ {
			fmt.Println("zzzzzzzzzzz", key, v)
			v = v.Next()
		}
		time.Sleep(1 * time.Second)
		return true
	})
	go func() {
		b.Daemon(context.TODO())
	}()
	for i := 0; i < 10; i++ {
		go func(i int) {
			b.Store("1", i)
		}(i)
	}
	time.Sleep(10 * time.Second)
}
