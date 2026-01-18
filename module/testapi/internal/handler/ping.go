package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/common/middleware"
	"github.com/nuohe369/crab/common/response"
)

// SetupPing 注册 Ping 和限流测试路由
func SetupPing(router fiber.Router) {
	// No rate limit
	router.Get("/ping", Ping)

	// Rate limit test: 5 times/10 seconds
	router.Get("/limit", middleware.RateLimit(5, 10*time.Second), LimitTest)

	// Rate limit by path: 3 times/10 seconds
	router.Get("/limit-path", middleware.RateLimitByPath(3, 10*time.Second), LimitPathTest)

	// Rate limit by IP: 10 times/minute
	router.Get("/limit-ip", middleware.RateLimitByIP(10, time.Minute), LimitIPTest)
}

// Ping health check
func Ping(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{
		"message": "pong",
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

// LimitTest rate limit test
func LimitTest(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{
		"message": "rate limit test passed",
		"limit":   "5 times/10 seconds",
	})
}

// LimitPathTest rate limit by path test
func LimitPathTest(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{
		"message": "rate limit by path test passed",
		"limit":   "3 times/10 seconds",
	})
}

// LimitIPTest rate limit by IP test
func LimitIPTest(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{
		"message": "rate limit by IP test passed",
		"limit":   "10 times/minute",
		"ip":      c.IP(),
	})
}
