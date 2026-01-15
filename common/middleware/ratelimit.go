package middleware

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/common/response"
	"github.com/nuohe369/crab/pkg/ratelimit"
)

// 默认限流器(内存)
var defaultLimiter ratelimit.Limiter

func init() {
	defaultLimiter = ratelimit.NewMemory()
}

// SetLimiter sets the global rate limiter.
func SetLimiter(l ratelimit.Limiter) {
	defaultLimiter = l
}

// RateLimitConfig defines rate limiting configuration.
type RateLimitConfig struct {
	// Maximum number of requests within the time window
	Max int
	// Time window duration
	Window time.Duration
	// Function to generate rate limit key, defaults to IP-based
	KeyGenerator func(c *fiber.Ctx) string
	// Function to skip rate limiting
	Skip func(c *fiber.Ctx) bool
	// Custom rate limiter
	Limiter ratelimit.Limiter
}

// RateLimit returns a rate limiting middleware.
func RateLimit(max int, window time.Duration) fiber.Handler {
	return RateLimitWithConfig(RateLimitConfig{
		Max:    max,
		Window: window,
	})
}

// RateLimitByIP returns a rate limiting middleware based on IP address.
func RateLimitByIP(max int, window time.Duration) fiber.Handler {
	return RateLimitWithConfig(RateLimitConfig{
		Max:    max,
		Window: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "ip:" + c.IP()
		},
	})
}

// RateLimitByUser returns a rate limiting middleware based on user ID.
// Requires authentication middleware to be applied first.
func RateLimitByUser(max int, window time.Duration) fiber.Handler {
	return RateLimitWithConfig(RateLimitConfig{
		Max:    max,
		Window: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			userID := c.Locals("user_id")
			if userID == nil {
				return "ip:" + c.IP() // fallback to IP for unauthenticated users
			}
			return fmt.Sprintf("user:%d", userID.(int64))
		},
	})
}

// RateLimitByPath returns a rate limiting middleware based on API path.
func RateLimitByPath(max int, window time.Duration) fiber.Handler {
	return RateLimitWithConfig(RateLimitConfig{
		Max:    max,
		Window: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "path:" + c.IP() + ":" + c.Path()
		},
	})
}

// RateLimitWithConfig returns a rate limiting middleware with custom configuration.
func RateLimitWithConfig(cfg RateLimitConfig) fiber.Handler {
	// 默认值
	if cfg.Max <= 0 {
		cfg.Max = 100
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Minute
	}
	if cfg.KeyGenerator == nil {
		cfg.KeyGenerator = func(c *fiber.Ctx) string {
			return "ip:" + c.IP()
		}
	}
	if cfg.Limiter == nil {
		cfg.Limiter = defaultLimiter
	}

	return func(c *fiber.Ctx) error {
		// Skip check if configured
		if cfg.Skip != nil && cfg.Skip(c) {
			return c.Next()
		}

		key := cfg.KeyGenerator(c)
		allowed, remaining, resetAt := cfg.Limiter.Allow(c.Context(), key, cfg.Max, cfg.Window)

		// Set response headers
		c.Set("X-RateLimit-Limit", strconv.Itoa(cfg.Max))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

		if !allowed {
			c.Set("Retry-After", strconv.FormatInt(int64(time.Until(resetAt).Seconds()), 10))
			return response.FailMsg(c, response.CodeTooManyRequests, "Too many requests, please try again later")
		}

		return c.Next()
	}
}
