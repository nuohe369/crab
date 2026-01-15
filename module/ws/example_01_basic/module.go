// Package example_01_basic basic example submodule
//
// Demonstrates the simplest WebSocket usage:
// - Create Hub to manage connections
// - Client connect/disconnect
// - Send/receive messages
//
// Test:
//
//	websocat ws://localhost:3000/ws/basic
//	> {"type": "ping"}
//	< {"type": "pong"}
package example_01_basic

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nuohe369/crab/module/ws/example_01_basic/internal/handler"
)

// Setup registers routes
func Setup(router fiber.Router) {
	handler.Setup(router)
}
