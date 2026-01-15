package ratelimit

import (
	"context"
	"time"
)

// Limiter rate limiter interface
type Limiter interface {
	// Allow checks if request is allowed
	// key: rate limit key (e.g. IP, user ID)
	// limit: max requests in time window
	// window: time window
	// Returns: allowed, remaining count, reset time
	Allow(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, remaining int, resetAt time.Time)
}

// Config rate limiter configuration
type Config struct {
	// Storage type: memory, redis
	Store string
	// Redis address (required when Store=redis)
	RedisAddr string
	// Redis password
	RedisPassword string
	// Redis DB
	RedisDB int
	// Key prefix
	KeyPrefix string
}

// New creates a rate limiter
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

// NewMemory creates a memory rate limiter
func NewMemory() Limiter {
	return newMemoryLimiter("ratelimit:")
}

// NewRedis creates a Redis rate limiter
func NewRedis(addr, password string, db int) Limiter {
	return newRedisLimiter(addr, password, db, "ratelimit:")
}
