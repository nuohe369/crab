package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/nuohe369/crab/pkg/logger"
	"github.com/nuohe369/crab/pkg/ws"
)

var log = logger.NewWithName[struct{}]("ws.multiuser")
var hub *ws.Hub

func init() {
	hub = ws.NewHub()
	go hub.Run()

	hub.OnMessage = func(client *ws.Client, msg *ws.Message) {
		log.Info("user=%d type=%s", client.UserID, msg.Type)

		switch msg.Type {
		case "send_to":
			payload, ok := msg.Payload.(map[string]any)
			if !ok {
				return
			}
			toID, _ := payload["to"].(float64)
			content, _ := payload["content"].(string)

			hub.SendToUser(int64(toID), &ws.Message{
				Type: "private_msg",
				Payload: map[string]any{
					"from":    client.UserID,
					"content": content,
				},
			})

		case "broadcast":
			payload, ok := msg.Payload.(map[string]any)
			if !ok {
				return
			}
			content, _ := payload["content"].(string)

			hub.Broadcast(&ws.Message{
				Type: "broadcast_msg",
				Payload: map[string]any{
					"from":    client.UserID,
					"content": content,
				},
			})

		case "online":
			payload, ok := msg.Payload.(map[string]any)
			if !ok {
				return
			}
			checkID, _ := payload["user_id"].(float64)
			online := hub.IsUserOnline(int64(checkID))

			client.Send(&ws.Message{
				Type: "online_result",
				Payload: map[string]any{
					"user_id": int64(checkID),
					"online":  online,
				},
			})
		}
	}
}

func Setup(router fiber.Router) {
	router.Get("/multiuser", websocket.New(handleWS))
}

func handleWS(conn *websocket.Conn) {
	userIDStr := conn.Query("user_id", "0")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	client := ws.NewClient(hub, userID, conn)
	hub.Register(client)
	defer hub.Unregister(client)

	client.Send(&ws.Message{
		Type: "connected",
		Payload: map[string]any{
			"user_id":      userID,
			"online_count": hub.UserCount(),
		},
	})

	go client.WritePump()
	client.ReadPump()
}
