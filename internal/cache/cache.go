package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Item[T any] struct {
	Value T         `json:"value"`
	TTL   time.Time `json:"ttl"`
}

type Cache[T any] struct {
	items     map[string]*Item[T]
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	onEvicted func(string, T)
}

func New[T any](ctx context.Context, cancel context.CancelFunc) *Cache[T] {
	return &Cache[T]{
		items:  make(map[string]*Item[T]),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *Cache[T]) OnEvicted(fn func(key string, value T)) {
	c.onEvicted = fn
}

func (c *Cache[T]) Set(key string, value T, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.items == nil {
		return errors.New("cache is closed")
	}

	c.items[key] = &Item[T]{
		Value: value,
		TTL:   time.Now().Add(ttl),
	}

	return nil
}

func (c *Cache[T]) SetWithContext(ctx context.Context, key string, value T, ttl time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		c.mu.Lock()
		defer c.mu.Unlock()

		if c.items == nil {
			return errors.New("cache is closed")
		}

		c.items[key] = &Item[T]{
			Value: value,
			TTL:   time.Now().Add(ttl),
		}

		return nil
	}
}

func (c *Cache[T]) Get(key string) (T, error) {
	c.mu.RLock()
	if c.items == nil {
		return zero[T](), errors.New("cache is closed")
	}

	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return zero[T](), fmt.Errorf("key %s does not exists", key)
	}

	if time.Now().After(item.TTL) {
		c.mu.Lock()
		if c.onEvicted != nil {
			c.onEvicted(key, item.Value)
		}
		delete(c.items, key)
		c.mu.Unlock()

		return zero[T](), fmt.Errorf("TTL of key %s has expire", key)
	}

	return item.Value, nil
}

func (c *Cache[T]) GetWithContext(ctx context.Context, key string) (T, error) {
	select {
	case <-ctx.Done():
		return zero[T](), ctx.Err()
	default:
		c.mu.RLock()
		if c.items == nil {
			return zero[T](), errors.New("cache is closed")
		}

		item, exists := c.items[key]
		c.mu.RUnlock()

		if !exists {
			return zero[T](), fmt.Errorf("key %s does not exists", key)
		}

		if time.Now().After(item.TTL) {
			c.mu.Lock()
			delete(c.items, key)
			c.mu.Unlock()

			return zero[T](), fmt.Errorf("TTL of key %s has expire", key)
		}

		return item.Value, nil
	}
}

func (c *Cache[T]) MarshalJSON() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.items == nil {
		return nil, errors.New("cache is closed")
	}

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

func (c *Cache[T]) MarshalJSONWithContext(ctx context.Context) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		c.mu.RLock()
		defer c.mu.RUnlock()

		if c.items == nil {
			return nil, errors.New("cache is closed")
		}

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
}

func (c *Cache[T]) UnmarshalJSON(bytes []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.items == nil {
		return errors.New("cache is closed")
	}

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

func (c *Cache[T]) UnmarshalJSONWithContext(ctx context.Context, bytes []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		c.mu.Lock()
		defer c.mu.Unlock()

		if c.items == nil {
			return errors.New("cache is closed")
		}

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

}

func (c *Cache[T]) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.items == nil {
		return
	}

	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}

	c.items = nil
}

func zero[T any]() T {
	var zero T
	return zero
}
