package slice

import (
	"github.com/Kotodian/gokit/lodash/types"
)

// 1.18 泛型支持 添加一些可复用的操作

func Filter[V any](collection []V, predicate func(V, int) bool) []V {
	result := make([]V, 0)

	for i, item := range collection {
		if predicate(item, i) {
			result = append(result, item)
		}
	}
	return result
}

func Map[T any, R any](collection []T, iteratee func(T, int) R) []R {
	result := make([]R, len(collection))

	for i, item := range collection {
		result[i] = iteratee(item, i)
	}
	return result
}

func ToMap[T any](collection []T) map[int]T {
	result := make(map[int]T)

	for i, v := range collection {
		result[i] = v
	}

	return result
}

func FlatMap[T any, R any](collection []T, iteratee func(T, int) []R) []R {
	result := make([]R, 0)

	for i, item := range collection {
		result = append(result, iteratee(item, i)...)
	}
	return result
}

func Reduce[T any, R any](collection []T, accumulator func(R, T, int) R, initial R) R {
	for i, item := range collection {
		initial = accumulator(initial, item, i)
	}
	return initial
}

func ForEach[T any](collection []T, iteratee func(T, int)) {
	for i, item := range collection {
		iteratee(item, i)
	}
}

func Times[T any](count int, iteratee func(int) T) []T {
	result := make([]T, count)

	for i := 0; i < count; i++ {
		result[i] = iteratee(i)
	}
	return result
}

func Uniq[T comparable](collection []T) []T {
	result := make([]T, 0, len(collection))
	seen := make(map[T]struct{}, len(collection))

	for _, item := range collection {
		if _, ok := seen[item]; !ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func Max[T types.Ordered](collection []T) T {
	var max T
	max = collection[0]

	for i := 1; i < len(collection); i++ {
		item := collection[i]
		if item > max {
			max = item
		}
	}
	return max
}

func Min[T types.Ordered](collection []T) T {
	min := collection[0]

	for i := 1; i < len(collection); i++ {
		if collection[i] < min {
			min = collection[i]
		}
	}
	return min
}

func IndexOf[T comparable](collection []T, element T) int {
	for i, item := range collection {
		if item == element {
			return i
		}
	}
	return -1
}

func LastIndexOf[T comparable](collection []T, element T) int {
	length := len(collection)
	for i := length - 1; i >= 0; i-- {
		if collection[i] == element {
			return i
		}
	}
	return -1
}

func Find[T any](collection []T, predicate func(T) bool) (T, bool) {
	for _, item := range collection {
		if predicate(item) {
			return item, true
		}
	}
	var result T
	return result, false
}

func Equal[E comparable](s1, s2 []E) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func EqualFunc[E1, E2 any](s1 []E1, s2 []E2, eq func(E1, E2) bool) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i, v1 := range s1 {
		v2 := s2[i]
		if !eq(v1, v2) {
			return false
		}
	}
	return true
}

func Compare[E types.Ordered](s1, s2 []E) int {
	s2len := len(s2)
	for i, v1 := range s1 {
		if i >= s2len {
			return +1
		}
		v2 := s2[i]
		switch {
		case v1 < v2:
			return -1
		case v1 > v2:
			return +1
		}
	}
	if len(s1) < s2len {
		return -1
	}
	return 0
}

func Clone[S ~[]E, E any](s S) S {
	if s == nil {
		return nil
	}
	return append(S([]E{}), s...)
}

func Compact[S ~[]E, E comparable](s S) S {
	if len(s) == 0 {
		return s
	}
	i := 1
	last := s[0]
	for _, v := range s[1:] {
		if v != last {
			s[i] = v
			i++
			last = v
		}
	}
	return s[:i]
}
