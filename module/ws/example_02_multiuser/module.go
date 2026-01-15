// Package example_02_multiuser multi-user example submodule
//
// Demonstrates multi-user scenarios:
// - Pass userID via query parameter
// - Send to specific user
// - Broadcast to all users
//
// Test:
//
//	websocat "ws://localhost:3000/ws/multiuser?user_id=1"
//	websocat "ws://localhost:3000/ws/multiuser?user_id=2"
//	> {"type": "send_to", "payload": {"to": 2, "content": "hello"}}
package example_02_multiuser

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nuohe369/crab/module/ws/example_02_multiuser/internal/handler"
)

func Setup(router fiber.Router) {
	handler.Setup(router)
}
