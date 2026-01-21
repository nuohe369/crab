package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// Setup registers global middleware (excluding Logger, which is registered by each module)
// Setup 注册全局中间件（不包括 Logger，Logger 由各模块自行注册）
func Setup(app *fiber.App) {
	app.Use(Recovery())
	app.Use(Cors())
	app.Use(SmartLogger()) // Smart request logger with auto module detection | 智能请求日志，自动检测模块
}
