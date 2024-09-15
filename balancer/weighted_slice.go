package balancer

import (
	"math/rand"
	"sync"
	"time"
)

// WeightedItem is a generic type that holds an item with its weight.
type WeightedItem[T any] struct {
	Value  T
	Weight int
}

// WeightedSlice is a generic type that holds a slice of weighted items.
type WeightedSlice[T comparable] struct {
	elements     []WeightedItem[T]
	indexMap     map[T][]int // Map to track the indices of each value
	mu           sync.Mutex  // Mutex to synchronize access to the slice
}

func NewWeightedSlice[T comparable]() *WeightedSlice[T] {
	return &WeightedSlice[T]{
		elements: []WeightedItem[T]{},
		indexMap: make(map[T][]int),
	}
}

// AddItem adds an item to the weighted slice according to its weight in a thread-safe manner.
func (ws *WeightedSlice[T]) AddItem(item WeightedItem[T]) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	for i := 0; i < item.Weight; i++ {
		ws.elements = append(ws.elements, item)
		// Track the index in the map
		ws.indexMap[item.Value] = append(ws.indexMap[item.Value], len(ws.elements)-1)
	}
}

// GetRandomItem returns a random item from the weighted slice in a thread-safe manner.
func (ws *WeightedSlice[T]) GetRandomItem() T {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(ws.elements))
	return ws.elements[randomIndex].Value
}

// RemoveItemByValue removes all instances of items with the given value from the weighted slice.
func (ws *WeightedSlice[T]) RemoveItemByValue(value T) {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	indicesToRemove, exists := ws.indexMap[value]
	if !exists {
		return
	}

	removeMap := make(map[int]struct{})
	for _, idx := range indicesToRemove {
		removeMap[idx] = struct{}{}
	}

	// Create a new slice excluding the items to remove
	newElements := make([]WeightedItem[T], 0, len(ws.elements))
	for i, item := range ws.elements {
		if _, shouldRemove := removeMap[i]; !shouldRemove {
			newElements = append(newElements, item)
		}
	}

	ws.elements = newElements
	delete(ws.indexMap, value)
}

// Len returns the current number of elements in the weighted slice in a thread-safe manner.
func (ws *WeightedSlice[T]) Len() int {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	return len(ws.elements)
}
