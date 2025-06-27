package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/crewcrew23/memo/internal/stat"
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
	stat      *stat.Stats
	sizeof    int64
}

func New[T any](ctx context.Context, cancel context.CancelFunc) *Cache[T] {
	return &Cache[T]{
		items:  make(map[string]*Item[T]),
		ctx:    ctx,
		cancel: cancel,
		stat:   &stat.Stats{},
	}
}

func (c *Cache[T]) OnEvicted(fn func(key string, value T)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.items == nil {
		return errors.New("cache is closed")
	}

	c.onEvicted = fn
	return nil
}

func (c *Cache[T]) Set(key string, value T, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.items == nil {
		return errors.New("cache is closed")
	}

	if c.sizeof == 0 {
		c.sizeof = getSize(value)
	}

	c.stat.SizeBytes += int64(c.sizeof)

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

		if c.sizeof == 0 {
			c.sizeof = getSize(value)
		}

		c.stat.SizeBytes += int64(c.sizeof)

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
		c.stat.Misses++
		return zero[T](), fmt.Errorf("key %s does not exists", key)
	}

	if time.Now().After(item.TTL) {
		c.mu.Lock()
		if c.onEvicted != nil {
			c.onEvicted(key, item.Value)
		}

		delete(c.items, key)

		c.stat.Evictions++
		c.stat.SizeBytes -= int64(c.sizeof)

		c.mu.Unlock()

		return zero[T](), fmt.Errorf("TTL of key %s has expire", key)
	}

	c.stat.Hits++
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
			c.stat.Misses++
			return zero[T](), fmt.Errorf("key %s does not exists", key)
		}

		if time.Now().After(item.TTL) {
			c.mu.Lock()
			if c.onEvicted != nil {
				c.onEvicted(key, item.Value)
			}

			delete(c.items, key)

			c.stat.Evictions++
			c.stat.SizeBytes -= int64(c.sizeof)

			c.mu.Unlock()

			return zero[T](), fmt.Errorf("TTL of key %s has expire", key)
		}

		c.stat.Hits++
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

func (c *Cache[T]) Stat() stat.Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.stat.Hits + c.stat.Misses
	rate := 0.0
	if total > 0 {
		rate = float64(c.stat.Hits) / float64(total) * 100
	}

	return stat.Stats{
		Hits:      c.stat.Hits,
		Misses:    c.stat.Misses,
		Evictions: c.stat.Evictions,
		HitRate:   rate,
		SizeBytes: c.stat.SizeBytes,
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

func getSize[T any](val T) int64 {
	if reflect.ValueOf(val).Kind() == reflect.Ptr {
		ptrSize := int64(reflect.TypeOf(val).Size())
		valSize := int64(reflect.TypeOf(val).Elem().Size())
		return ptrSize + valSize
	}

	return int64(reflect.TypeOf(val).Size())
}

func zero[T any]() T {
	var zero T
	return zero
}
