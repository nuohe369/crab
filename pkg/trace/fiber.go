package trace

import (
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
)

// FiberMiddleware returns Fiber tracing middleware
func FiberMiddleware() fiber.Handler {
	return otelfiber.Middleware()
}
