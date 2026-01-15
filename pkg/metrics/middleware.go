package metrics

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Middleware 返回 Fiber 中间件,自动采集 HTTP 请求指标
func Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !enabled {
			return c.Next()
		}

		// 跳过 metrics 端点本身
		if c.Path() == path {
			return c.Next()
		}

		start := time.Now()
		IncHTTPInFlight()

		err := c.Next()

		DecHTTPInFlight()
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())

		// 记录指标
		RecordHTTPRequest(c.Method(), c.Path(), status, duration)

		return err
	}
}

// Handler 返回 Prometheus 指标暴露的 HTTP handler
func Handler() fiber.Handler {
	if !enabled || registry == nil {
		return func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusServiceUnavailable).SendString("metrics not enabled")
		}
	}
	return adaptor.HTTPHandler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
}
