package maps

import (
	"golang.org/x/exp/constraints"
	"github.com/Kotodian/gokit/lodash/slice"
)

type Entry[K comparable, V any] struct {
	Key   K
	Value V
}

func Keys[K comparable, V any](in map[K]V) []K {
	result := make([]K, 0, len(in))

	for k := range in {
		result = append(result, k)
	}
	return result
}

func Values[K comparable, V any](in map[K]V) []V {
	result := make([]V, 0, len(in))

	for _, v := range in {
		result = append(result, v)
	}

	return result
}

func Entries[K comparable, V any](in map[K]V) []Entry[K, V] {
	entries := make([]Entry[K, V], 0, len(in))

	for k, v := range in {
		entries = append(entries, Entry[K, V]{
			Key:   k,
			Value: v,
		})
	}
	return entries
}

func Assign[K comparable, V any](maps ...map[K]V) map[K]V {
	out := make(map[K]V)

	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}

	return out
}

func MapValues[K comparable, V any, R any](in map[K]V, iteratee func(V, K) R) map[K]R {
	result := make(map[K]R)

	for k, v := range in {
		result[k] = iteratee(v, k)
	}

	return result
}

func MaxKey[K constraints.Ordered, V any](in map[K]V) K {
	keys := Keys(in)
	maxKey := slice.Max(keys)
	return maxKey
}

func MaxValue[K comparable, V constraints.Ordered](in map[K]V) V {
	values := Values(in)
	maxValue := slice.Max(values)
	return maxValue
}

func MinKey[K constraints.Ordered, V any](in map[K]V) K {
	return slice.Min(Keys(in))
}

func MinValue[K comparable, V constraints.Ordered](in map[K]V) V {
	return slice.Min(Values(in))
}
