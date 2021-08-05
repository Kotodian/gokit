package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// RandInt64 int64随机数
func RandInt64(min, max int64) int64 {
	_rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	if min >= max || min == 0 || max == 0 {
		return max
	}
	return _rand.Int63n(max-min) + min
}

// VerifyCode 6位验证码
func VerifyCode() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	vcode := fmt.Sprintf("%06v", rnd.Int31n(1000000))
	return vcode
}
