// Package example_03_callback callback example submodule
//
// Demonstrates Hub callback mechanism:
// - OnConnect: triggered when connection established
// - OnDisconnect: triggered when connection closed
// - OnMessage: triggered when message received
//
// Test:
//
//	websocat "ws://localhost:3000/ws/callback?user_id=1"
package example_03_callback

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nuohe369/crab/module/ws/example_03_callback/internal/handler"
)

func Setup(router fiber.Router) {
	handler.Setup(router)
}
