package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/pkg/trace"
)

// Trace returns a distributed tracing middleware
// Trace 返回分布式追踪中间件
func Trace() fiber.Handler {
	return trace.FiberMiddleware()
}
