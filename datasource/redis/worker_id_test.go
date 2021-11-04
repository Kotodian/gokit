package redis

import (
	"os"
	"sync"
	"testing"
)

func TestRegisterOne(t *testing.T) {
	os.Setenv("REDIS_POOL", "10.43.0.20:6379")
	os.Setenv("REDIS_AUTH", "LhBIOQumQdgIm4ro")
	Init()

	wg := sync.WaitGroup{}

	wg.Add(2)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		v, err := RegisterOne(15)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log("service A", v)
	}(&wg)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		v, err := RegisterOne(15)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log("service A", v)
	}(&wg)
	wg.Wait()

}
