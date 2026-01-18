package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"
)

// ErrNotFound indicates cache key not found
// ErrNotFound 表示缓存键未找到
var ErrNotFound = errors.New("cache: key not found")

// RedisClient defines the Redis client interface
// RedisClient 定义 Redis 客户端接口
type RedisClient interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

// Config represents cache configuration
// Config 表示缓存配置
type Config struct {
	LocalTTL    time.Duration // local cache TTL, default 1 minute | 本地缓存 TTL，默认 1 分钟
	LocalSize   int           // maximum local cache entries, default 10000 | 最大本地缓存条目数，默认 10000
	EnableLocal bool          // enable local cache, default true | 启用本地缓存，默认 true
}

// Cache implements a two-level cache system
// Cache 实现两级缓存系统
type Cache struct {
	redis  RedisClient
	local  *localCache
	config Config
}

var defaultCache *Cache

// Init initializes the default cache
// Init 初始化默认缓存
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

// Get returns the default cache instance
// Get 返回默认缓存实例
func Get() *Cache {
	return defaultCache
}

// ============ Convenience methods (using default cache) | 便捷方法（使用默认缓存） ============

// GetValue retrieves a cached value
// GetValue 获取缓存值
func GetValue(ctx context.Context, key string, dest any) error {
	return defaultCache.GetValue(ctx, key, dest)
}

// SetValue sets a cached value
// SetValue 设置缓存值
func SetValue(ctx context.Context, key string, val any, ttl time.Duration) error {
	return defaultCache.SetValue(ctx, key, val, ttl)
}

// GetOrSet retrieves a cached value, or loads and caches it if not found
// GetOrSet 获取缓存值，如果未找到则加载并缓存
func GetOrSet(ctx context.Context, key string, dest any, ttl time.Duration, loader func() (any, error)) error {
	return defaultCache.GetOrSet(ctx, key, dest, ttl, loader)
}

// Del deletes cached values
// Del 删除缓存值
func Del(ctx context.Context, keys ...string) error {
	return defaultCache.Del(ctx, keys...)
}

// New creates a new cache instance
// New 创建新的缓存实例
func New(redis RedisClient, cfg Config) *Cache {
	return &Cache{
		redis:  redis,
		local:  newLocalCache(cfg.LocalSize, cfg.LocalTTL),
		config: cfg,
	}
}

// GetValue retrieves a cached value
// GetValue 获取缓存值
func (c *Cache) GetValue(ctx context.Context, key string, dest any) error {
	// 1. Check local cache first | 首先检查本地缓存
	if c.config.EnableLocal {
		if data, ok := c.local.get(key); ok {
			return json.Unmarshal(data, dest)
		}
	}

	// 2. Check Redis | 检查 Redis
	if c.redis != nil {
		val, err := c.redis.Get(ctx, key)
		if err == nil {
			// Backfill local cache | 回填本地缓存
			if c.config.EnableLocal {
				c.local.set(key, []byte(val))
			}
			return json.Unmarshal([]byte(val), dest)
		}
		// Log Redis errors (except key not found) | 记录 Redis 错误（键未找到除外）
		if !isNotFound(err) {
			log.Printf("cache: redis get error: %v", err)
		}
	}

	return ErrNotFound
}

// SetValue sets a cached value
// SetValue 设置缓存值
func (c *Cache) SetValue(ctx context.Context, key string, val any, ttl time.Duration) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	// 1. Write to local cache | 写入本地缓存
	if c.config.EnableLocal {
		c.local.set(key, data)
	}

	// 2. Write to Redis | 写入 Redis
	if c.redis != nil {
		if err := c.redis.Set(ctx, key, string(data), ttl); err != nil {
			log.Printf("cache: redis set error: %v", err)
			// Don't return error if Redis write fails, local cache is already written
			// 如果 Redis 写入失败不返回错误，本地缓存已经写入
		}
	}

	return nil
}

// GetOrSet retrieves a cached value, or loads and caches it if not found
// GetOrSet 获取缓存值，如果未找到则加载并缓存
func (c *Cache) GetOrSet(ctx context.Context, key string, dest any, ttl time.Duration, loader func() (any, error)) error {
	// Try to get first | 先尝试获取
	if err := c.GetValue(ctx, key, dest); err == nil {
		return nil
	}

	// Load data | 加载数据
	val, err := loader()
	if err != nil {
		return err
	}

	// Cache and return | 缓存并返回
	if err := c.SetValue(ctx, key, val, ttl); err != nil {
		return err
	}

	// Copy value to dest | 复制值到目标
	data, _ := json.Marshal(val)
	return json.Unmarshal(data, dest)
}

// Del deletes cached values
// Del 删除缓存值
func (c *Cache) Del(ctx context.Context, keys ...string) error {
	// 1. Delete from local cache | 从本地缓存删除
	if c.config.EnableLocal {
		for _, key := range keys {
			c.local.del(key)
		}
	}

	// 2. Delete from Redis | 从 Redis 删除
	if c.redis != nil {
		if err := c.redis.Del(ctx, keys...); err != nil {
			log.Printf("cache: redis del error: %v", err)
		}
	}

	return nil
}

// Close closes the cache and stops cleanup goroutines
// Close 关闭缓存并停止清理协程
func (c *Cache) Close() {
	if c.local != nil {
		c.local.Close()
	}
}

// isNotFound checks if the error indicates a key not found
// isNotFound 检查错误是否表示键未找到
func isNotFound(err error) bool {
	return err != nil && err.Error() == "redis: nil"
}
