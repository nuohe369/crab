package handler

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/nuohe369/crab/pkg/logger"
	"github.com/nuohe369/crab/pkg/redis"
	"github.com/nuohe369/crab/pkg/ws"
)

const channel = "ws:example:cluster"

var (
	log = logger.NewWithName("ws.cluster")
	hub *ws.Hub
	ctx context.Context
)

func Setup(router fiber.Router) {
	ctx = context.Background()
	hub = ws.NewHub()
	go hub.Run()

	// If Redis is available, enable cluster mode
	if rdb := redis.Get(); rdb != nil {
		if client, ok := rdb.GetRaw().(ws.RedisClient); ok {
			hub.EnableCluster(ctx, client, channel)
			log.Info("Redis cluster mode enabled")
		}
	} else {
		log.Info("Redis not available, using standalone mode")
	}

	hub.OnMessage = func(client *ws.Client, msg *ws.Message) {
		log.Info("user=%d type=%s", client.UserID, msg.Type)

		switch msg.Type {
		case "broadcast":
			payload, ok := msg.Payload.(map[string]any)
			if !ok {
				return
			}
			content, _ := payload["content"].(string)

			hub.Publish(ctx, &ws.Message{
				Type: "broadcast_msg",
				Payload: map[string]any{
					"from":    client.UserID,
					"content": content,
				},
			})

		case "send_to":
			payload, ok := msg.Payload.(map[string]any)
			if !ok {
				return
			}
			toID, _ := payload["to"].(float64)
			content, _ := payload["content"].(string)

			hub.PublishToUser(ctx, int64(toID), &ws.Message{
				Type: "private_msg",
				Payload: map[string]any{
					"from":    client.UserID,
					"content": content,
				},
			})
		}
	}

	router.Get("/cluster", websocket.New(handleWS))
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
			"user_id": userID,
			"mode":    getMode(),
		},
	})

	go client.WritePump()
	client.ReadPump()
}

func getMode() string {
	if redis.Get() != nil {
		return "cluster"
	}
	return "standalone"
}
