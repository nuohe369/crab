package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/nuohe369/crab/pkg/logger"
	"github.com/nuohe369/crab/pkg/ws"
)

var log = logger.NewWithName[struct{}]("ws.basic")
var hub *ws.Hub

func init() {
	hub = ws.NewHub()
	go hub.Run()

	hub.OnMessage = func(client *ws.Client, msg *ws.Message) {
		log.Info("received message: type=%s", msg.Type)

		switch msg.Type {
		case "ping":
			client.Send(&ws.Message{Type: "pong"})
		default:
			client.Send(msg)
		}
	}

	hub.OnConnect = func(client *ws.Client) {
		log.Info("client connected: userID=%d", client.UserID)
	}

	hub.OnDisconnect = func(client *ws.Client) {
		log.Info("client disconnected: userID=%d", client.UserID)
	}
}

// Setup registers routes
func Setup(router fiber.Router) {
	log.Info("registering route: /basic")

	router.Use("/basic", func(c *fiber.Ctx) error {
		log.Info("received request: %s, IsWebSocket=%v", c.Path(), websocket.IsWebSocketUpgrade(c))
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	router.Get("/basic", websocket.New(handleWS))
}

func handleWS(conn *websocket.Conn) {
	log.Info("WebSocket connection established")

	client := ws.NewClient(hub, 0, conn)
	hub.Register(client)
	defer hub.Unregister(client)

	go client.WritePump()
	client.ReadPump()

	log.Info("WebSocket connection closed")
}
