package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/common/response"
)

// SetupPing 注册 Ping 路由
func SetupPing(router fiber.Router) {
	// No rate limit
	router.Get("/ping", Ping)
}

// Ping health check
func Ping(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{
		"message": "pong",
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}
