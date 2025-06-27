package cache

import (
	"context"
	"time"
)

func StartClean[T any](c *Cache[T], ctx context.Context, interval time.Duration) {
	go func() {
	Loop:
		for {
			select {
			case <-ctx.Done():
				break Loop

			default:
				time.Sleep(interval)
				clean(c)
			}
		}
	}()
}

func clean[T any](c *Cache[T]) {
	type tmp struct {
		key   string
		value *Item[T]
	}

	var expiredKeys []*tmp

	for k, v := range c.items {
		c.mu.RLock()
		if time.Now().After(v.TTL) {
			expiredKeys = append(expiredKeys, &tmp{key: k, value: v})
		}
		c.mu.RUnlock()
	}

	if len(expiredKeys) > 0 {
		c.mu.Lock()
		for _, k := range expiredKeys {
			if c.onEvicted != nil {
				c.onEvicted(k.key, k.value.Value)
			}

			delete(c.items, k.key)

			c.stat.Evictions++
			c.stat.SizeBytes -= int64(c.sizeof)
		}
		c.mu.Unlock()

	}
}
