package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// Setup registers global middleware (excluding Logger, which is registered by each module).
func Setup(app *fiber.App) {
	app.Use(Recovery())
	app.Use(Cors())
	app.Use(Trace())
}
