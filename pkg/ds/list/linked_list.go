package list

// node represents a node in the doubly linked list
type node[T any] struct {
	value T
	next  *node[T]
	prev  *node[T]
}

// LinkedList represents the doubly linked list
type LinkedList[T any] struct {
	head *node[T]
	tail *node[T]
	size int
}

// NewLinkedList creates a new empty linked list
func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{
		head: nil,
		tail: nil,
		size: 0,
	}
}

// Head returns the value of the first element in the list
func (l *LinkedList[T]) Head() (T, bool) {
	var zero T
	if l.head == nil {
		return zero, false
	}
	return l.head.value, true
}

// Tail returns the value of the last element in the list
func (l *LinkedList[T]) Tail() (T, bool) {
	var zero T
	if l.tail == nil {
		return zero, false
	}
	return l.tail.value, true
}

// Size returns the number of elements in the list
func (l *LinkedList[T]) Size() int {
	return l.size
}

// Append adds a new value to the end of the list
func (l *LinkedList[T]) Append(value T) {
	newNode := &node[T]{value: value, next: nil, prev: l.tail}
	l.size++

	if l.head == nil {
		l.head = newNode
		l.tail = newNode
		return
	}

	l.tail.next = newNode
	l.tail = newNode
}

// Prepend adds a new value to the beginning of the list
func (l *LinkedList[T]) Prepend(value T) {
	newNode := &node[T]{value: value, next: l.head, prev: nil}
	l.size++

	if l.head == nil {
		l.head = newNode
		l.tail = newNode
		return
	}

	l.head.prev = newNode
	l.head = newNode
}

// InsertAt inserts a value at the specified index
func (l *LinkedList[T]) InsertAt(value T, index int) bool {
	if index < 0 || index > l.size {
		return false
	}

	if index == 0 {
		l.Prepend(value)
		return true
	}

	if index == l.size {
		l.Append(value)
		return true
	}

	current := l.head
	for i := 0; i < index; i++ {
		current = current.next
	}

	newNode := &node[T]{value: value, next: current, prev: current.prev}
	current.prev.next = newNode
	current.prev = newNode
	l.size++
	return true
}

// RemoveAt removes the node at the specified index
func (l *LinkedList[T]) RemoveAt(index int) bool {
	if index < 0 || index >= l.size {
		return false
	}

	if index == 0 {
		return l.DeleteHead()
	}

	if index == l.size-1 {
		return l.DeleteTail()
	}

	current := l.head
	for i := 0; i < index; i++ {
		current = current.next
	}

	current.prev.next = current.next
	current.next.prev = current.prev
	l.size--
	return true
}

// DeleteHead removes the first element of the list
func (l *LinkedList[T]) DeleteHead() bool {
	if l.head == nil {
		return false
	}

	l.head = l.head.next
	l.size--

	if l.head == nil {
		l.tail = nil
	} else {
		l.head.prev = nil
	}

	return true
}

// DeleteTail removes the last element of the list
func (l *LinkedList[T]) DeleteTail() bool {
	if l.tail == nil {
		return false
	}

	l.tail = l.tail.prev
	l.size--

	if l.tail == nil {
		l.head = nil
	} else {
		l.tail.next = nil
	}

	return true
}

// GetAt returns the value at the specified index
func (l *LinkedList[T]) GetAt(index int) (T, bool) {
	var zero T
	if index < 0 || index >= l.size {
		return zero, false
	}

	var current *node[T]
	if index < l.size/2 {
		current = l.head
		for i := 0; i < index; i++ {
			current = current.next
		}
	} else {
		current = l.tail
		for i := l.size - 1; i > index; i-- {
			current = current.prev
		}
	}

	return current.value, true
}

// ToSlice converts the linked list to a slice
func (l *LinkedList[T]) ToSlice() []T {
	result := make([]T, l.size)
	current := l.head
	for i := 0; current != nil; i++ {
		result[i] = current.value
		current = current.next
	}
	return result
}

// IsEmpty returns true if the list is empty
func (l *LinkedList[T]) IsEmpty() bool {
	return l.size == 0
}
