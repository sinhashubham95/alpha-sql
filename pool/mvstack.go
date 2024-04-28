package pool

import "github.com/sinhashubham95/go-utils/structures/stack"

// mvStack implements a multi-version stack.
//
// mvStack works as common stack except for the fact that all elements in the
// older version are guaranteed to be popped before any element in the newer
// version. New elements are always pushed to the current (latest)
// version.
//
// We could also say that mvStack behaves as a stack in case of a single
// version, but it behaves as a queue of individual version stacks.
type mvStack struct {
	old *stack.Stack[*Connection]
	new *stack.Stack[*Connection]
}

func newMVStack() *mvStack {
	s := stack.New[*Connection]()
	return &mvStack{
		old: s,
		new: s,
	}
}

func (s *mvStack) pop() (*Connection, bool) {
	if s.old.Length() == 0 && s.old != s.new {
		s.old = s.new
	}
	if s.old.Length() == 0 {
		return nil, false
	}
	return s.old.Pop()
}

func (s *mvStack) push(c *Connection) {
	s.new.Push(c)
}

func (s *mvStack) bump() {
	if s.old == s.new {
		s.new = stack.New[*Connection]()
		return
	}
	old := make([]*Connection, s.old.Length())
	for s.old.Length() > 0 {
		c, _ := s.old.Pop()
		old[s.old.Length()-1] = c
	}
	for _, c := range old {
		s.new.Push(c)
	}
	s.old, s.new = s.new, s.old
}

func (s *mvStack) length() int {
	l := s.old.Length()
	if s.old != s.new {
		l += s.new.Length()
	}
	return l
}
