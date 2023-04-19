package xwcache

import (
	"sync"
	"xwace/xwcache/xwcache/lru"
)

type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) get(key string) (*ImmutableByte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if value, ok := c.lru.Get(key); ok {
		return value.(*ImmutableByte), ok
	}
	return nil, false
}

// todo 感觉这里如果value要用指针的话，是不是得加个nil判断
func (c *cache) add(key string, value *ImmutableByte) (*ImmutableByte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lru.Add(key, value)
	return value, true
}
