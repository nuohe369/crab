package cache

import (
	"sync"
	"time"
)

// localCache is the local memory cache
type localCache struct {
	data    map[string]*cacheItem
	mu      sync.RWMutex
	maxSize int
	ttl     time.Duration
	stopCh  chan struct{}
}

type cacheItem struct {
	value    []byte
	expireAt time.Time
}

func newLocalCache(maxSize int, ttl time.Duration) *localCache {
	c := &localCache{
		data:    make(map[string]*cacheItem),
		maxSize: maxSize,
		ttl:     ttl,
		stopCh:  make(chan struct{}),
	}
	// Start cleanup goroutine
	go c.cleanup()
	return c
}

func (c *localCache) get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.data[key]
	if !ok {
		return nil, false
	}

	// Check if expired
	if time.Now().After(item.expireAt) {
		return nil, false
	}

	return item.value, true
}

func (c *localCache) set(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simple eviction strategy: clean expired items when exceeding max capacity
	if len(c.data) >= c.maxSize {
		c.evict()
	}

	c.data[key] = &cacheItem{
		value:    value,
		expireAt: time.Now().Add(c.ttl),
	}
}

func (c *localCache) del(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// evict removes expired data
func (c *localCache) evict() {
	now := time.Now()
	for key, item := range c.data {
		if now.After(item.expireAt) {
			delete(c.data, key)
		}
	}
}

// cleanup periodically cleans expired data
func (c *localCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			c.evict()
			c.mu.Unlock()
		case <-c.stopCh:
			return
		}
	}
}

// Close stops the cleanup goroutine
func (c *localCache) Close() {
	close(c.stopCh)
}

// Len returns cache entry count
func (c *localCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}
