// Package middleware provides HTTP middleware handlers
// middleware 包提供 HTTP 中间件处理器
package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// Cors returns a CORS middleware handler
// Cors 返回 CORS 中间件处理器
func Cors() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins: "*",                                        // Allowed origins | 允许的源
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",              // Allowed methods | 允许的方法
		AllowHeaders: "Origin,Content-Type,Accept,Authorization", // Allowed headers | 允许的请求头
	})
}
