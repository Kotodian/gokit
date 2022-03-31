package treemap

import (
	"time"

	"github.com/Kotodian/gokit/lodash/types"
)

type Comparator[K any] func(a, b K) int

func NumbersComparator[T types.Integer](a, b T) int {
	switch a > b {
	case a > b:
		return 1
	case a < b:
		return -1
	default:
		return 0
	}
}

func StringsComparator(s1, s2 string) int {
	min := len(s2)
	if len(s1) < len(s2) {
		min = len(s1)
	}
	diff := 0
	for i := 0; i < min && diff == 0; i++ {
		diff = int(s1[i]) - int(s2[i])
	}
	if diff == 0 {
		diff = len(s1) - len(s2)
	}
	if diff < 0 {
		return -1
	}
	if diff > 0 {
		return 1
	}
	return 0
}

func TimeComparator(a, b time.Time) int {
	switch {
	case a.After(b):
		return 1
	case a.Before(b):
		return -1
	default:
		return 0
	}
}