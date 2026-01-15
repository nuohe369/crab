package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/nuohe369/crab/pkg/logger"
	"github.com/nuohe369/crab/pkg/ws"
)

var log = logger.NewWithName[struct{}]("ws.callback")
var hub *ws.Hub

func init() {
	hub = ws.NewHub()
	go hub.Run()

	hub.OnConnect = func(client *ws.Client) {
		log.Info("user %d connected, online: %d", client.UserID, hub.ClientCount())

		hub.Broadcast(&ws.Message{
			Type: "user_joined",
			Payload: map[string]any{
				"user_id":      client.UserID,
				"online_count": hub.UserCount(),
			},
		})
	}

	hub.OnDisconnect = func(client *ws.Client) {
		log.Info("user %d disconnected, online: %d", client.UserID, hub.ClientCount())

		hub.Broadcast(&ws.Message{
			Type: "user_left",
			Payload: map[string]any{
				"user_id":      client.UserID,
				"online_count": hub.UserCount(),
			},
		})
	}

	hub.OnMessage = func(client *ws.Client, msg *ws.Message) {
		log.Info("user %d sent message: %s", client.UserID, msg.Type)

		client.Send(&ws.Message{
			Type:    "echo",
			Payload: msg.Payload,
		})
	}
}

func Setup(router fiber.Router) {
	router.Get("/callback", websocket.New(handleWS))
}

func handleWS(conn *websocket.Conn) {
	userIDStr := conn.Query("user_id", "0")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	client := ws.NewClient(hub, userID, conn)
	hub.Register(client)
	defer hub.Unregister(client)

	go client.WritePump()
	client.ReadPump()
}
