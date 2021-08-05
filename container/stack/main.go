package stack


import (
	"container/list"
)

type Stack struct {
	list *list.List
	len  int
}

func NewStack(len int) *Stack {
	return &Stack{list.New(), len}
}

func (stack *Stack) Push(value interface{}) bool {
	if stack.Full() {
		return false
	}
	stack.list.PushBack(value)
	return true
}

func (stack *Stack) Pop() interface{} {
	if e := stack.list.Back(); e != nil {
		return stack.list.Remove(e)
	}
	return nil
}

func (stack *Stack) Peak() interface{} {
	e := stack.list.Back()
	if e != nil {
		return e.Value
	}

	return nil
}

func (stack Stack) Len() int {
	return stack.list.Len()
}

func (stack *Stack) Empty() bool {
	return stack.list.Len() == 0
}

func (stack Stack) Full() bool {
	if stack.len > 0 && stack.list.Len() >= stack.len {
		return true
	}
	return false
}