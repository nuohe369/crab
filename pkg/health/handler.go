// Package health provides health check functionality for monitoring service status
// Package health 提供健康检查功能，用于监控服务状态
package health

import (
	"github.com/gofiber/fiber/v2"
)

// FiberHandler returns Fiber health check handler
// FiberHandler 返回 Fiber 健康检查处理器
func FiberHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		result := Check(c.Context())

		status := fiber.StatusOK
		if result.Status != StatusUp {
			status = fiber.StatusServiceUnavailable
		}

		return c.Status(status).JSON(result)
	}
}

// FiberLivenessHandler returns Fiber liveness check handler
// FiberLivenessHandler 返回 Fiber 存活检查处理器
func FiberLivenessHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		result := Liveness(c.Context())
		return c.Status(fiber.StatusOK).JSON(result)
	}
}

// FiberReadinessHandler returns Fiber readiness check handler
// FiberReadinessHandler 返回 Fiber 就绪检查处理器
func FiberReadinessHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		result := Readiness(c.Context())

		status := fiber.StatusOK
		if result.Status != StatusUp {
			status = fiber.StatusServiceUnavailable
		}

		return c.Status(status).JSON(result)
	}
}

// RegisterFiberRoutes registers Fiber health check routes
// RegisterFiberRoutes 注册 Fiber 健康检查路由
func RegisterFiberRoutes(app *fiber.App, prefix string) {
	if prefix == "" {
		prefix = "/health"
	}

	group := app.Group(prefix)
	group.Get("", FiberHandler())
	group.Get("/live", FiberLivenessHandler())
	group.Get("/ready", FiberReadinessHandler())

	log.Info("Health check routes registered: %s, %s/live, %s/ready", prefix, prefix, prefix)
}
