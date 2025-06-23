package cache

import (
	"sync"
	"time"
)

type Item[T any] struct {
	value T
	ttl   time.Time
}

type Cache[T any] struct {
	items map[string]*Item[T]
	mu    sync.RWMutex
}

func New[T any]() *Cache[T] {
	return &Cache[T]{
		items: make(map[string]*Item[T]),
	}
}

func (c *Cache[T]) Set(key string, value T, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &Item[T]{
		value: value,
		ttl:   time.Now().Add(ttl),
	}
}

func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists || time.Now().After(item.ttl) {
		var zero T
		return zero, false
	}

	return item.value, true
}
