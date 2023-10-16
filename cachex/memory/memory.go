package memory

import (
	"github.com/c1emon/gcommon/cachex"
	"github.com/fanjindong/go-cache"
)

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		cache: cache.NewMemCache(),
	}
}

var _ cachex.Cacher = &MemoryCache{}

type MemoryCache struct {
	cache cache.ICache
}

// Del implements cachex.Cacher.
func (c *MemoryCache) Del(keys ...string) int {
	return c.cache.Del(keys...)
}

// Get implements cachex.Cacher.
func (c *MemoryCache) Get(key string) (any, bool) {
	return c.cache.Get(key)
}

// Set implements cachex.Cacher.
func (c *MemoryCache) Set(key string, val any) {
	c.cache.Set(key, val)
}
