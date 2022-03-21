package list

type linkedList[T any] struct {
	count       uint
	front, back *node[T]
}

type node[T any] struct {
	value      T
	prev, next *node[T]
}

func NewNode[T any](v T) *node[T] {
	return &node[T]{
		value: v,
		next:  nil,
		prev:  nil,
	}
}

func (l *linkedList[T]) PushFront(n *node[T]) {
	n.next = l.front
	n.prev = nil
	if l.count == 0 {
		l.back = n
	} else {
		l.front.prev = n
	}
	l.front = n
	l.count++
}

func (l *linkedList[T]) PopFront() *node[T] {
	n := l.front
	l.count--
	if l.count == 0 {
		l.front, l.back = nil, nil
	} else {
		n.next.prev = nil
		l.front = n.next
	}
	n.next, n.prev = nil, nil
	return n
}

func (l *linkedList[T]) PopBack() *node[T] {
	n := l.back
	l.count--
	if l.count == 0 {
		l.front, l.back = nil, nil
	} else {
		n.prev.next = nil
		l.back = n.prev
	}
	n.next, n.prev = nil, nil
	return n
}