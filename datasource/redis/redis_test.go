package redis

import (
	"fmt"
	"testing"
	"time"

	redislib "/redis"
)

func Test_Setnx(t *testing.T) {
	rd := GetRedis()
	_ = rd.Send("MULTI")
	_ = rd.Send("zadd", "test:breeze", time.Now().Local().Unix(), "11111")
	_ = rd.Send("zadd", "test:breeze", time.Now().Local().Unix(), "22222")
	_ = rd.Send("set", "22222", "3teat", "nx", "ex", 200)

	reply, err := redislib.Values(rd.Do("exec"))
	if reply[2] == nil {
		fmt.Println("*****")
	}
	fmt.Println("-------->exec err: ", reply[2], err)
	n, _ := redislib.Int(rd.Do("zcard", "test:breeze"))
	fmt.Println("------->n: ", n)
}
