package utils

import (
	"fmt"
	"strconv"
)

func Decimal(value float64, n ...uint) float64 {
	_n := uint(2)
	if len(n) >= 1 {
		_n = n[0]
	}
	f := fmt.Sprintf("%%.%df", _n)
	value, _ = strconv.ParseFloat(fmt.Sprintf(f, value), 64)
	return value
}
