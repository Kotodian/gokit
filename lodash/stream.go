package lodash

import "github.com/Kotodian/gokit/lodash/slice"

type sliceStream[T comparable] struct {
	in []T
}

func NewSliceStream[T comparable](in []T) *sliceStream[T] {
	return &sliceStream[T]{in: in}
}

func (stream *sliceStream[T]) Filter(predicate func(T, int) bool) *sliceStream[T] {
	stream.in = slice.Filter(stream.in, predicate)
	return stream
}

func (stream *sliceStream[T]) ForEach(iteratee func(T, int)) *sliceStream[T] {
	slice.ForEach(stream.in, iteratee)
	return stream
}

func (stream *sliceStream[T]) Times(count int, iteratee func(int) T) *sliceStream[T] {
	stream.in = slice.Times(count, iteratee)
	return stream
}

func (stream *sliceStream[T]) Uniq() *sliceStream[T] {
	stream.in = slice.Uniq(stream.in)
	return stream
}

func (stream *sliceStream[T]) Result() []T  {
	return stream.in	
}

