package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"
)

var ErrNotFound = errors.New("cache: key not found")

// RedisClient defines the Redis client interface.
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

// Config represents cache configuration.
type Config struct {
	LocalTTL    time.Duration // local cache TTL, default 1 minute
	LocalSize   int           // maximum local cache entries, default 10000
	EnableLocal bool          // enable local cache, default true
}

// Cache implements a two-level cache system.
type Cache struct {
	redis  RedisClient
	local  *localCache
	config Config
}

var defaultCache *Cache

// Init initializes the default cache.
func Init(redis RedisClient, cfg ...Config) {
	c := Config{
		LocalTTL:    time.Minute,
		LocalSize:   10000,
		EnableLocal: true,
	}
	if len(cfg) > 0 {
		c = cfg[0]
	}
	defaultCache = New(redis, c)
	log.Println("cache: initialized")
}

// Get returns the default cache instance.
func Get() *Cache {
	return defaultCache
}

// ============ Convenience methods (using default cache) ============

// GetValue retrieves a cached value.
func GetValue(ctx context.Context, key string, dest any) error {
	return defaultCache.GetValue(ctx, key, dest)
}

// SetValue sets a cached value.
func SetValue(ctx context.Context, key string, val any, ttl time.Duration) error {
	return defaultCache.SetValue(ctx, key, val, ttl)
}

// GetOrSet retrieves a cached value, or loads and caches it if not found.
func GetOrSet(ctx context.Context, key string, dest any, ttl time.Duration, loader func() (any, error)) error {
	return defaultCache.GetOrSet(ctx, key, dest, ttl, loader)
}

// Del deletes cached values.
func Del(ctx context.Context, keys ...string) error {
	return defaultCache.Del(ctx, keys...)
}

// New creates a new cache instance.
func New(redis RedisClient, cfg Config) *Cache {
	return &Cache{
		redis:  redis,
		local:  newLocalCache(cfg.LocalSize, cfg.LocalTTL),
		config: cfg,
	}
}

// GetValue retrieves a cached value.
func (c *Cache) GetValue(ctx context.Context, key string, dest any) error {
	// 1. Check local cache first
	if c.config.EnableLocal {
		if data, ok := c.local.get(key); ok {
			return json.Unmarshal(data, dest)
		}
	}

	// 2. Check Redis
	if c.redis != nil {
		val, err := c.redis.Get(ctx, key)
		if err == nil {
			// Backfill local cache
			if c.config.EnableLocal {
				c.local.set(key, []byte(val))
			}
			return json.Unmarshal([]byte(val), dest)
		}
		// Log Redis errors (except key not found)
		if !isNotFound(err) {
			log.Printf("cache: redis get error: %v", err)
		}
	}

	return ErrNotFound
}

// SetValue sets a cached value.
func (c *Cache) SetValue(ctx context.Context, key string, val any, ttl time.Duration) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	// 1. Write to local cache
	if c.config.EnableLocal {
		c.local.set(key, data)
	}

	// 2. Write to Redis
	if c.redis != nil {
		if err := c.redis.Set(ctx, key, string(data), ttl); err != nil {
			log.Printf("cache: redis set error: %v", err)
			// Don't return error if Redis write fails, local cache is already written
		}
	}

	return nil
}

// GetOrSet retrieves a cached value, or loads and caches it if not found.
func (c *Cache) GetOrSet(ctx context.Context, key string, dest any, ttl time.Duration, loader func() (any, error)) error {
	// Try to get first
	if err := c.GetValue(ctx, key, dest); err == nil {
		return nil
	}

	// Load data
	val, err := loader()
	if err != nil {
		return err
	}

	// Cache and return
	if err := c.SetValue(ctx, key, val, ttl); err != nil {
		return err
	}

	// Copy value to dest
	data, _ := json.Marshal(val)
	return json.Unmarshal(data, dest)
}

// Del deletes cached values.
func (c *Cache) Del(ctx context.Context, keys ...string) error {
	// 1. Delete from local cache
	if c.config.EnableLocal {
		for _, key := range keys {
			c.local.del(key)
		}
	}

	// 2. Delete from Redis
	if c.redis != nil {
		if err := c.redis.Del(ctx, keys...); err != nil {
			log.Printf("cache: redis del error: %v", err)
		}
	}

	return nil
}

// isNotFound checks if the error indicates a key not found.
func isNotFound(err error) bool {
	return err != nil && err.Error() == "redis: nil"
}
