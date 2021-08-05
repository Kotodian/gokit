package utils

import (
	"fmt"
	"testing"
)

func TestRandCode(t *testing.T) {
	fmt.Println(RandCodeInt())
	fmt.Println(RandCodeStr())
}
