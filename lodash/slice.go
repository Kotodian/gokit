package lodash

import "golang.org/x/exp/constraints"

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

func Max[T constraints.Ordered](collection []T) T {
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

func Min[T constraints.Ordered](collection []T) T {
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
