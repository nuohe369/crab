package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/pkg/trace"
)

// Trace returns a distributed tracing middleware.
func Trace() fiber.Handler {
	return trace.FiberMiddleware()
}
