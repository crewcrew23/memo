package cache

import "time"

func StartClean[T any](c *Cache[T]) {
	go func() {
		for {
			time.Sleep(time.Minute * 5)
			clean(c)
		}
	}()
}

func clean[T any](c *Cache[T]) {
	var expiredKeys []string

	for k, v := range c.items {
		c.mu.RLock()
		if time.Now().After(v.TTL) {
			expiredKeys = append(expiredKeys, k)
		}
		c.mu.RUnlock()
	}

	if len(expiredKeys) > 0 {
		c.mu.Lock()
		for _, k := range expiredKeys {
			delete(c.items, k)
		}
		c.mu.Unlock()

	}
}
