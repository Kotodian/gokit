package conditions

type ifElse[T any] struct {
	result T
	done   bool
}

func IF[T any](cond bool, res T) *ifElse[T] {
	if cond {
		return &ifElse[T]{result: res, done: true}
	}
	var t T
	return &ifElse[T]{t, false}
}

func (i *ifElse[T]) ElseIf(cond bool, res T) *ifElse[T] {
	if !i.done && cond {
		i.result = res
		i.done = true
	}

	return i
}

func (i *ifElse[T]) Else(res T) T {
	if i.done {
		return i.result
	}
	return res
}

type switchCase[T comparable, R any] struct {
	predicate T
	result    R
	done      bool
}

func Switch[T comparable, R any](predicate T) *switchCase[T, R] {
	var result R

	return &switchCase[T, R]{
		predicate: predicate,
		result:    result,
		done:      false,
	}
}

func (s *switchCase[T, R]) Case(val T, result R) *switchCase[T, R] {
	if !s.done && s.predicate == val {
		s.result = result
		s.done = true
	}

	return s
}

func (s *switchCase[T, R]) Default(result R) R {
	if !s.done {
		s.result = result
	}
	return s.result
}
