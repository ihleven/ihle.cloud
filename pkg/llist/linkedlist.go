package llist

import "fmt"

type elem struct {
	next    *elem
	content interface{}
}

func (e elem) String() string {
	if e.next != nil {
		return fmt.Sprintf("%v, %v", e.content, e.next)
	}
	return fmt.Sprintf("%v ]", e.content)

}

func (e *elem) Add(a interface{}) {
	elem := elem{content: a, next: e.next}
	e.next = &elem
}
func NewList(a interface{}) *elem {
	return &elem{content: a}
}
