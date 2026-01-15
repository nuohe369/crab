package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/pkg/logger"
	"github.com/nuohe369/crab/pkg/trace"
)

// loggers 缓存各模块的日志器
var loggers = make(map[string]*logger.Logger[struct{}])

func getLogger(module string) *logger.Logger[struct{}] {
	if l, ok := loggers[module]; ok {
		return l
	}
	l := logger.NewWithName[struct{}](module)
	loggers[module] = l
	return l
}

// Logger returns a request logging middleware for the specified module.
func Logger(module string) fiber.Handler {
	log := getLogger(module)
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()
		traceID := trace.TraceID(c.UserContext())

		if traceID != "" {
			log.InfoCtx(c.UserContext(), "%s %s %d %v", method, path, status, latency)
		} else {
			log.Info("%s %s %d %v", method, path, status, latency)
		}

		return err
	}
}
