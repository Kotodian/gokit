package utils

import (
	"fmt"
	"testing"
)

func Test_SetBit(t *testing.T) {
	fmt.Println("-----> ", SetBit(0, 0))
	fmt.Println("-----> ", SetBit(0, 1))
	fmt.Println("-----> ", GetBit(2, 1))
}
