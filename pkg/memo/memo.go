package memo

import "memo/internal/cache"

func New[T any]() *cache.Cache[T] {
	c := cache.New[T]()
	cache.StartClean(c)
	return c
}
