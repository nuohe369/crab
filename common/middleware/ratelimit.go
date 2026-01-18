package middleware

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/common/response"
	"github.com/nuohe369/crab/pkg/ratelimit"
)

// defaultLimiter is the default rate limiter (memory-based) | defaultLimiter 默认限流器（基于内存）
var defaultLimiter ratelimit.Limiter

func init() {
	defaultLimiter = ratelimit.NewMemory()
}

// SetLimiter sets the global rate limiter
// SetLimiter 设置全局限流器
func SetLimiter(l ratelimit.Limiter) {
	defaultLimiter = l
}

// RateLimitConfig defines rate limiting configuration
// RateLimitConfig 定义限流配置
type RateLimitConfig struct {
	Max          int                       // Maximum number of requests within the time window | 时间窗口内的最大请求数
	Window       time.Duration             // Time window duration | 时间窗口时长
	KeyGenerator func(c *fiber.Ctx) string // Function to generate rate limit key, defaults to IP-based | 生成限流键的函数，默认基于 IP
	Skip         func(c *fiber.Ctx) bool   // Function to skip rate limiting | 跳过限流的函数
	Limiter      ratelimit.Limiter         // Custom rate limiter | 自定义限流器
}

// RateLimit returns a rate limiting middleware
// RateLimit 返回限流中间件
func RateLimit(max int, window time.Duration) fiber.Handler {
	return RateLimitWithConfig(RateLimitConfig{
		Max:    max,
		Window: window,
	})
}

// RateLimitByIP returns a rate limiting middleware based on IP address
// RateLimitByIP 返回基于 IP 地址的限流中间件
func RateLimitByIP(max int, window time.Duration) fiber.Handler {
	return RateLimitWithConfig(RateLimitConfig{
		Max:    max,
		Window: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "ip:" + c.IP()
		},
	})
}

// RateLimitByUser returns a rate limiting middleware based on user ID
// Requires authentication middleware to be applied first
// RateLimitByUser 返回基于用户 ID 的限流中间件
// 需要先应用认证中间件
func RateLimitByUser(max int, window time.Duration) fiber.Handler {
	return RateLimitWithConfig(RateLimitConfig{
		Max:    max,
		Window: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			userID := c.Locals("user_id")
			if userID == nil {
				return "ip:" + c.IP() // Fallback to IP for unauthenticated users | 未认证用户回退到 IP
			}
			return fmt.Sprintf("user:%d", userID.(int64))
		},
	})
}

// RateLimitByPath returns a rate limiting middleware based on API path
// RateLimitByPath 返回基于 API 路径的限流中间件
func RateLimitByPath(max int, window time.Duration) fiber.Handler {
	return RateLimitWithConfig(RateLimitConfig{
		Max:    max,
		Window: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "path:" + c.IP() + ":" + c.Path()
		},
	})
}

// RateLimitWithConfig returns a rate limiting middleware with custom configuration
// RateLimitWithConfig 返回带自定义配置的限流中间件
func RateLimitWithConfig(cfg RateLimitConfig) fiber.Handler {
	// Set default values | 设置默认值
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
		// Skip check if configured | 如果配置了跳过，则跳过检查
		if cfg.Skip != nil && cfg.Skip(c) {
			return c.Next()
		}

		key := cfg.KeyGenerator(c)
		allowed, remaining, resetAt := cfg.Limiter.Allow(c.Context(), key, cfg.Max, cfg.Window)

		// Set response headers | 设置响应头
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
