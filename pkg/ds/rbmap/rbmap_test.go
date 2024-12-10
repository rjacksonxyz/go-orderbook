package rbmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMap(t *testing.T) {
	// Create a new map with integers as keys and strings as values
	intMap := NewMap[int, string](Ascending[int])

	// Insert some values
	intMap.Insert(5, "five")
	intMap.Insert(3, "three")
	intMap.Insert(7, "seven")

	expecetdOrder := []int{3, 5, 7}

	// Retrieve a value
	if val, exists := intMap.Get(3); exists {
		assert.Equal(t, val, "three") // Prints: three
	}

	// Iterate through the map in order
	i := 0
	for it := intMap.Begin(); it.First(); it.Next() {
		assert.Equal(t, it.Key(), int(expecetdOrder[i]))
		i++
	}

	// Assert that the first element key is	3
	it := intMap.Begin()
	assert.Equal(t, it.Key(), 3)
}
