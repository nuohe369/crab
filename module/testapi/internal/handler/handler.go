package handler

import (
	"github.com/gofiber/fiber/v2"
)

// Setup 注册所有路由
func Setup(router fiber.Router) {
	// Ping examples
	SetupPing(router)

	// CRUD examples (multi-database demo)
	SetupUser(router)
	SetupCategory(router)
	SetupArticle(router)
}
