package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// redisLimiter Redis rate limiter (sliding window)
type redisLimiter struct {
	rdb    *redis.Client
	prefix string
}

func newRedisLimiter(addr, password string, db int, prefix string) *redisLimiter {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &redisLimiter{rdb: rdb, prefix: prefix}
}

// Lua script: sliding window rate limiting
var luaScript = redis.NewScript(`
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

-- Remove records outside the window
redis.call('ZREMRANGEBYSCORE', key, 0, now - window)

-- Get current window request count
local count = redis.call('ZCARD', key)

if count < limit then
    -- Add current request
    redis.call('ZADD', key, now, now .. '-' .. math.random())
    redis.call('PEXPIRE', key, window)
    return {1, limit - count - 1}
else
    return {0, 0}
end
`)

func (r *redisLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Time) {
	fullKey := r.prefix + key
	now := time.Now()
	windowMs := window.Milliseconds()
	nowMs := now.UnixMilli()

	result, err := luaScript.Run(ctx, r.rdb, []string{fullKey}, limit, windowMs, nowMs).Slice()
	if err != nil {
		// Allow on Redis error to avoid affecting business
		return true, limit, now.Add(window)
	}

	allowed := result[0].(int64) == 1
	remaining := int(result[1].(int64))
	resetAt := now.Add(window)

	return allowed, remaining, resetAt
}
