package cache

import (
	"time"
)

type Interface[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, val V)
	Keys() []K
	Delete(key K)
}

type Item[K comparable, V any] struct {
	Key        K
	Value      V
	Expiration time.Duration
}

var NowFunc = time.Now

type itemOptions struct {
	expiration time.Duration
}
type ItemOption func(*itemOptions)

func WithExipiration(exp time.Duration) ItemOption {
	return func(io *itemOptions) {
		io.expiration = exp
	}
}

func newItem[K comparable, V any](key K, val V, opts ...ItemOption) *Item[K, V] {
	o := new(itemOptions)
	for _, optFunc := range opts {
		optFunc(o)
	}
	return &Item[K, V]{
		Key:        key,
		Value:      val,
		Expiration: o.expiration,
	}
}

