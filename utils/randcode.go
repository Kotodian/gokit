package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// Return
// return six bit num code
func RandCodeInt() int {
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)
	code := r.Intn(999999)
	if code < 100000 {
		code += 100000
	}
	return code
}

func RandCodeStr() string {
	return fmt.Sprintf("%06d", RandCodeInt())
}
