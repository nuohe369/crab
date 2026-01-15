// Package example_05_service service integration example
//
// Demonstrates integration with common/service/ws:
// - Use service.GetUserHub() to get global Hub
// - Other modules push messages via service.PublishToUser()
//
// Test:
//
//	# 1. Connect WebSocket
//	websocat "ws://localhost:3000/ws/service?user_id=123"
//
//	# 2. Push message via HTTP API
//	curl "http://localhost:3000/testapi/ws/push?user_id=123&content=hello"
//
//	# 3. WebSocket client will receive the message
package example_05_service

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nuohe369/crab/module/ws/example_05_service/internal/handler"
)

func Setup(router fiber.Router) {
	handler.Setup(router)
}
