// Package example_04_cluster cluster example submodule
//
// Demonstrates Redis Pub/Sub cluster mode:
// - Messages auto-sync across multiple nodes
// - Send messages via hub.Publish()
//
// Test:
//
//	websocat "ws://localhost:3000/ws/cluster?user_id=1"
//	> {"type": "broadcast", "payload": {"content": "hello"}}
package example_04_cluster

import (
	"github.com/gofiber/fiber/v2"

	"github.com/nuohe369/crab/module/ws/example_04_cluster/internal/handler"
)

func Setup(router fiber.Router) {
	handler.Setup(router)
}
