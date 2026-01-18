// Package ratelimit provides rate limiting functionality for API throttling
// Package ratelimit 提供用于 API 限流的速率限制功能
package ratelimit

import (
	"context"
	"time"
)

// Limiter is the rate limiter interface
// Limiter 是速率限制器接口
type Limiter interface {
	// Allow checks if request is allowed
	// key: rate limit key (e.g. IP, user ID)
	// limit: max requests in time window
	// window: time window
	// Returns: allowed, remaining count, reset time
	// Allow 检查请求是否允许
	// key: 限流键（例如 IP、用户 ID）
	// limit: 时间窗口内的最大请求数
	// window: 时间窗口
	// 返回：是否允许、剩余次数、重置时间
	Allow(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, remaining int, resetAt time.Time)
}

// Config represents rate limiter configuration
// Config 表示速率限制器配置
type Config struct {
	Store         string // Storage type: memory, redis | 存储类型：memory、redis
	RedisAddr     string // Redis address (required when Store=redis) | Redis 地址（Store=redis 时必需）
	RedisPassword string // Redis password | Redis 密码
	RedisDB       int    // Redis DB | Redis 数据库
	KeyPrefix     string // Key prefix | 键前缀
}

// New creates a rate limiter
// New 创建速率限制器
func New(cfg Config) Limiter {
	prefix := cfg.KeyPrefix
	if prefix == "" {
		prefix = "ratelimit:"
	}

	switch cfg.Store {
	case "redis":
		return newRedisLimiter(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, prefix)
	default:
		return newMemoryLimiter(prefix)
	}
}

// NewMemory creates a memory-based rate limiter
// NewMemory 创建基于内存的速率限制器
func NewMemory() Limiter {
	return newMemoryLimiter("ratelimit:")
}

// NewRedis creates a Redis-based rate limiter
// NewRedis 创建基于 Redis 的速率限制器
func NewRedis(addr, password string, db int) Limiter {
	return newRedisLimiter(addr, password, db, "ratelimit:")
}
