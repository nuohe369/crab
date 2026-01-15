package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/nuohe369/crab/common/service"
	"github.com/nuohe369/crab/pkg/logger"
	"github.com/nuohe369/crab/pkg/ws"
)

var log = logger.NewWithName[struct{}]("ws.service")

func Setup(router fiber.Router) {
	router.Get("/service", websocket.New(handleWS))
}

func handleWS(conn *websocket.Conn) {
	userIDStr := conn.Query("user_id", "0")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	hub := service.GetUserHub()
	if hub == nil {
		log.Error("UserHub not initialized")
		conn.Close()
		return
	}

	client := ws.NewClient(hub, userID, conn)
	hub.Register(client)
	defer hub.Unregister(client)

	client.Send(&ws.Message{
		Type: "connected",
		Payload: map[string]any{
			"user_id": userID,
			"hub":     "user",
		},
	})

	go client.WritePump()
	client.ReadPump()
}
