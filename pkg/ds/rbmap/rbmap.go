package rbmap

import "cmp"

// Color represents the color of a Red-Black tree node
type Color bool

const (
	Black Color = true
	Red   Color = false
)

type Sortable interface {
	comparable
	cmp.Ordered
}

// Node represents a node in the Red-Black tree
type Node[K Sortable, V any] struct {
	Key    K
	Value  V
	color  Color
	left   *Node[K, V]
	right  *Node[K, V]
	parent *Node[K, V]
}

// Map implements an ordered map using a Red-Black tree
type Map[K Sortable, V any] struct {
	root *Node[K, V]
	size int
	less func(a, b K) bool // Custom comparison function
}

// NewMap creates a new map with a custom comparison function
func NewMap[K cmp.Ordered, V any](less SortFunc[K]) *Map[K, V] {
	return &Map[K, V]{
		less: less,
	}
}

type SortFunc[K Sortable] func(a, b K) bool

func Ascending[K Sortable](a, b K) bool {
	return a < b
}

func Descending[K Sortable](a, b K) bool {
	return a > b
}

// rotateLeft performs a left rotation around the given node
func (m *Map[K, V]) rotateLeft(x *Node[K, V]) {
	y := x.right
	x.right = y.left
	if y.left != nil {
		y.left.parent = x
	}
	y.parent = x.parent
	if x.parent == nil {
		m.root = y
	} else if x == x.parent.left {
		x.parent.left = y
	} else {
		x.parent.right = y
	}
	y.left = x
	x.parent = y
}

// rotateRight performs a right rotation around the given node
func (m *Map[K, V]) rotateRight(x *Node[K, V]) {
	y := x.left
	x.left = y.right
	if y.right != nil {
		y.right.parent = x
	}
	y.parent = x.parent
	if x.parent == nil {
		m.root = y
	} else if x == x.parent.right {
		x.parent.right = y
	} else {
		x.parent.left = y
	}
	y.right = x
	x.parent = y
}

// Insert adds a new key-value pair to the map
func (m *Map[K, V]) Insert(key K, value V) {
	var parent *Node[K, V]
	current := m.root

	// Find the insertion point
	for current != nil {
		parent = current
		if m.less(key, current.Key) {
			current = current.left
		} else if m.less(current.Key, key) {
			current = current.right
		} else {
			// Key already exists, update value
			current.Value = value
			return
		}
	}

	// Create new node
	newNode := &Node[K, V]{
		Key:    key,
		Value:  value,
		color:  Red,
		parent: parent,
	}

	// Insert the node
	if parent == nil {
		m.root = newNode
	} else if m.less(key, parent.Key) {
		parent.left = newNode
	} else {
		parent.right = newNode
	}

	m.size++
	m.fixInsert(newNode)
}

// fixInsert maintains Red-Black tree properties after insertion
func (m *Map[K, V]) fixInsert(node *Node[K, V]) {
	for node != m.root && node.parent.color == Red {
		if node.parent == node.parent.parent.left {
			uncle := node.parent.parent.right
			if uncle != nil && uncle.color == Red {
				// Case 1: Uncle is red
				node.parent.color = Black
				uncle.color = Black
				node.parent.parent.color = Red
				node = node.parent.parent
			} else {
				if node == node.parent.right {
					// Case 2: Uncle is black, node is right child
					node = node.parent
					m.rotateLeft(node)
				}
				// Case 3: Uncle is black, node is left child
				node.parent.color = Black
				node.parent.parent.color = Red
				m.rotateRight(node.parent.parent)
			}
		} else {
			// Same as above with "left" and "right" exchanged
			uncle := node.parent.parent.left
			if uncle != nil && uncle.color == Red {
				node.parent.color = Black
				uncle.color = Black
				node.parent.parent.color = Red
				node = node.parent.parent
			} else {
				if node == node.parent.left {
					node = node.parent
					m.rotateRight(node)
				}
				node.parent.color = Black
				node.parent.parent.color = Red
				m.rotateLeft(node.parent.parent)
			}
		}
	}
	m.root.color = Black
}

// Get retrieves the value associated with the given key
func (m *Map[K, V]) Get(key K) (V, bool) {
	node := m.root
	for node != nil {
		if m.less(key, node.Key) {
			node = node.left
		} else if m.less(node.Key, key) {
			node = node.right
		} else {
			return node.Value, true
		}
	}
	var zero V
	return zero, false
}

// Delete removes a key-value pair from the map
func (m *Map[K, V]) Delete(key K) bool {
	node := m.root
	// Find the node to delete
	for node != nil {
		if m.less(key, node.Key) {
			node = node.left
		} else if m.less(node.Key, key) {
			node = node.right
		} else {
			break
		}
	}

	if node == nil {
		return false
	}

	m.size--
	m.deleteNode(node)
	return true
}

// deleteNode removes the given node from the tree
func (m *Map[K, V]) deleteNode(node *Node[K, V]) {
	var child, parent *Node[K, V]
	originalColor := node.color

	if node.left == nil {
		child = node.right
		m.transplant(node, node.right)
	} else if node.right == nil {
		child = node.left
		m.transplant(node, node.left)
	} else {
		// Node has two children
		successor := m.minimum(node.right)
		originalColor = successor.color
		child = successor.right

		if successor.parent == node {
			if child != nil {
				child.parent = successor
			}
		} else {
			m.transplant(successor, successor.right)
			successor.right = node.right
			successor.right.parent = successor
		}

		m.transplant(node, successor)
		successor.left = node.left
		successor.left.parent = successor
		successor.color = node.color
	}

	if originalColor == Black {
		m.fixDelete(child, parent)
	}
}

// transplant replaces one subtree with another
func (m *Map[K, V]) transplant(u, v *Node[K, V]) {
	if u.parent == nil {
		m.root = v
	} else if u == u.parent.left {
		u.parent.left = v
	} else {
		u.parent.right = v
	}
	if v != nil {
		v.parent = u.parent
	}
}

// minimum finds the minimum key in a subtree
func (m *Map[K, V]) minimum(node *Node[K, V]) *Node[K, V] {
	current := node
	for current.left != nil {
		current = current.left
	}
	return current
}

// fixDelete maintains Red-Black tree properties after deletion
func (m *Map[K, V]) fixDelete(node, parent *Node[K, V]) {
	for node != m.root && (node == nil || node.color == Black) {
		if node == parent.left {
			sibling := parent.right
			if sibling.color == Red {
				sibling.color = Black
				parent.color = Red
				m.rotateLeft(parent)
				sibling = parent.right
			}

			if (sibling.left == nil || sibling.left.color == Black) &&
				(sibling.right == nil || sibling.right.color == Black) {
				sibling.color = Red
				node = parent
				parent = node.parent
			} else {
				if sibling.right == nil || sibling.right.color == Black {
					if sibling.left != nil {
						sibling.left.color = Black
					}
					sibling.color = Red
					m.rotateRight(sibling)
					sibling = parent.right
				}
				sibling.color = parent.color
				parent.color = Black
				if sibling.right != nil {
					sibling.right.color = Black
				}
				m.rotateLeft(parent)
				node = m.root
			}
		} else {
			// Same as above with "left" and "right" exchanged
			sibling := parent.left
			if sibling.color == Red {
				sibling.color = Black
				parent.color = Red
				m.rotateRight(parent)
				sibling = parent.left
			}

			if (sibling.right == nil || sibling.right.color == Black) &&
				(sibling.left == nil || sibling.left.color == Black) {
				sibling.color = Red
				node = parent
				parent = node.parent
			} else {
				if sibling.left == nil || sibling.left.color == Black {
					if sibling.right != nil {
						sibling.right.color = Black
					}
					sibling.color = Red
					m.rotateLeft(sibling)
					sibling = parent.left
				}
				sibling.color = parent.color
				parent.color = Black
				if sibling.left != nil {
					sibling.left.color = Black
				}
				m.rotateRight(parent)
				node = m.root
			}
		}
	}
	if node != nil {
		node.color = Black
	}
}

// First returns the first (smallest) key-value pair in the map
func (m *Map[K, V]) First() (K, V, bool) {
	if m.root == nil {
		var zeroK K
		var zeroV V
		return zeroK, zeroV, false
	}

	// Find the leftmost node (smallest key)
	current := m.root
	for current.left != nil {
		current = current.left
	}
	return current.Key, current.Value, true
}

// First returns the last (largest) key-value pair in the map
func (m *Map[K, V]) Last() (K, V, bool) {
	if m.root == nil {
		var zeroK K
		var zeroV V
		return zeroK, zeroV, false
	}

	// Find the leftmost node (smallest key)
	current := m.root
	for current.right != nil {
		current = current.right
	}
	return current.Key, current.Value, true
}

// Size returns the number of elements in the map
func (m *Map[K, V]) Size() int {
	return m.size
}

// Empty return a boolean indicating if the map is empty
func (m *Map[K, V]) Empty() bool {
	return m.size == 0
}

// Clear removes all elements from the map
func (m *Map[K, V]) Clear() {
	m.root = nil
	m.size = 0
}

// Iterator provides in-order traversal of the map
type Iterator[K Sortable, V any] struct {
	current *Node[K, V]
}

// Next moves the iterator to the next element and returns true if successful
func (it *Iterator[K, V]) Next() bool {
	if it.current == nil {
		return false
	}

	if it.current.right != nil {
		// Find leftmost node in right subtree
		it.current = it.current.right
		for it.current.left != nil {
			it.current = it.current.left
		}
	} else {
		// Find first ancestor where current is in left subtree
		for it.current.parent != nil && it.current == it.current.parent.right {
			it.current = it.current.parent
		}
		it.current = it.current.parent
	}
	return it.current != nil
}

// Key returns the current key
func (it *Iterator[K, V]) Key() K {
	return it.current.Key
}

// Value returns the current value
func (it *Iterator[K, V]) Value() V {
	return it.current.Value
}

// Begin returns an iterator pointing to the first element
func (m *Map[K, V]) Begin() Iterator[K, V] {
	if m.root == nil {
		return Iterator[K, V]{nil}
	}
	// Find the leftmost node (smallest key)
	current := m.root
	for current.left != nil {
		current = current.left
	}
	return Iterator[K, V]{current}
}

// First returns true if the iterator is valid and points to the first element
func (it *Iterator[K, V]) First() bool {
	return it.current != nil
}
