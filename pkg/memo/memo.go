package memo

import (
	"context"

	"github.com/crewcrew23/memo/internal/cache"
)

func New[T any]() *cache.Cache[T] {
	ctx, cancel := context.WithCancel(context.Background())
	c := cache.New[T](ctx, cancel)
	cache.StartClean(c, ctx)
	return c
}
