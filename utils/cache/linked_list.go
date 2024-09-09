package cache

// List is a double linked list with the limited number of methods.
// NOT thread safe!
type List[T any] struct {
	first *Node[T]
	last  *Node[T]
}

func NewList[T any]() *List[T] {
	return &List[T]{
		first: nil,
		last:  nil,
	}
}

// Node is an element of doubly linked list.
// NOT thread safe!
type Node[T any] struct {
	elem T
	prev *Node[T]
	next *Node[T]
	list *List[T]
}

// Insert inserts new node to the end of the list.
// NOT thread safe!
func (l *List[T]) Insert(elem T) *Node[T] {
	n := &Node[T]{
		elem: elem,
		prev: l.last,
		next: nil,
		list: l,
	}
	if l.last == nil {
		l.first = n
	} else {
		l.last.next = n
	}
	l.last = n
	return n
}

// Delete removes node from the list.
// NOT thread safe!
func (n *Node[T]) Delete() {
	if n.list.first == n {
		n.list.first = n.next
	}
	if n.list.last == n {
		n.list.last = n.prev
	}
	if n.next != nil {
		n.next.prev = n.prev
	}
	if n.prev != nil {
		n.prev.next = n.next
	}
	n.next, n.prev = nil, nil
}

// Range loops over all elements and calls f with each of them.
// NOT thread safe!
func (l *List[T]) Range(f func(T) bool) {
	n := l.first
	for n != nil {
		if !f(n.elem) {
			break
		}
		n = n.next
	}
}
