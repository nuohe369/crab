package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// Recovery returns a panic recovery middleware
// Recovery 返回 panic 恢复中间件
func Recovery() fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true, // Enable stack trace logging | 启用堆栈跟踪日志
	})
}
