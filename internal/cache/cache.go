package cache

import (
	"encoding/json"
	"sync"
	"time"
)

type Item[T any] struct {
	Value T         `json:"value"`
	TTL   time.Time `json:"ttl"`
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
		Value: value,
		TTL:   time.Now().Add(ttl),
	}
}

func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		var zero T
		return zero, false
	}

	if time.Now().After(item.TTL) {
		c.mu.Lock()
		defer c.mu.Unlock()

		delete(c.items, key)

		var zero T
		return zero, false
	}

	return item.Value, true
}

func (c *Cache[T]) MarshalJSON() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	serializable := make(map[string]struct {
		Value T         `json:"value"`
		TTL   time.Time `json:"ttl"`
	})

	for k, v := range c.items {
		serializable[k] = struct {
			Value T         `json:"value"`
			TTL   time.Time `json:"ttl"`
		}{
			Value: v.Value,
			TTL:   v.TTL,
		}
	}

	return json.Marshal(serializable)
}

func (c *Cache[T]) UnmarshalJSON(bytes []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var temp map[string]struct {
		Value T         `json:"value"`
		TTL   time.Time `json:"ttl"`
	}

	if err := json.Unmarshal(bytes, &temp); err != nil {
		return err
	}

	for k, v := range temp {
		c.items[k] = &Item[T]{
			Value: v.Value,
			TTL:   v.TTL,
		}
	}

	return nil

}
