package health

import (
	"github.com/gofiber/fiber/v2"
)

// FiberHandler returns Fiber health check handler
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
func FiberLivenessHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		result := Liveness(c.Context())
		return c.Status(fiber.StatusOK).JSON(result)
	}
}

// FiberReadinessHandler returns Fiber readiness check handler
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
