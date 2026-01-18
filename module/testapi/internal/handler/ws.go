package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/common/response"
	"github.com/nuohe369/crab/common/service"
	"github.com/nuohe369/crab/pkg/ws"
)

// SetupWS 注册 WebSocket 测试路由
func SetupWS(router fiber.Router) {
	g := router.Group("/ws")
	g.Get("/push", PushToUser)
	g.Get("/broadcast", Broadcast)
	g.Get("/online", GetOnlineCount)
}

// PushToUser pushes message to specified user
//
// GET /testapi/ws/push?user_id=123&content=hello
func PushToUser(c *fiber.Ctx) error {
	userIDStr := c.Query("user_id", "0")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)
	content := c.Query("content", "test message")

	if userID == 0 {
		return response.FailMsg(c, response.CodeParamError, "user_id cannot be empty")
	}

	err := service.PublishToUser(context.Background(), userID, &ws.Message{
		Type: "push_msg",
		Payload: map[string]any{
			"content": content,
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		},
	})
	if err != nil {
		return response.FailMsg(c, response.CodeServerError, err.Error())
	}

	return response.OK(c, fiber.Map{
		"message": "push success",
		"user_id": userID,
		"content": content,
	})
}

// Broadcast broadcasts message to all users
//
// GET /testapi/ws/broadcast?content=hello
func Broadcast(c *fiber.Ctx) error {
	content := c.Query("content", "broadcast message")

	err := service.PublishToUser(context.Background(), 0, &ws.Message{
		Type: "broadcast_msg",
		Payload: map[string]any{
			"content": content,
			"time":    time.Now().Format("2006-01-02 15:04:05"),
		},
	})
	if err != nil {
		return response.FailMsg(c, response.CodeServerError, err.Error())
	}

	return response.OK(c, fiber.Map{
		"message": "broadcast success",
		"content": content,
	})
}

// GetOnlineCount returns online user count
//
// GET /testapi/ws/online
func GetOnlineCount(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{
		"online_count": service.GetUserOnlineCount(),
	})
}
