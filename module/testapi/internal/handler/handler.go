package handler

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/common/middleware"
	"github.com/nuohe369/crab/common/response"
	"github.com/nuohe369/crab/common/service"
	"github.com/nuohe369/crab/pkg/ws"
)

func Setup(router fiber.Router) {
	// No rate limit
	router.Get("/ping", Ping)

	// Rate limit test: 5 times/10 seconds
	router.Get("/limit", middleware.RateLimit(5, 10*time.Second), LimitTest)

	// Rate limit by path: 3 times/10 seconds
	router.Get("/limit-path", middleware.RateLimitByPath(3, 10*time.Second), LimitPathTest)

	// Rate limit by IP: 10 times/minute
	router.Get("/limit-ip", middleware.RateLimitByIP(10, time.Minute), LimitIPTest)

	// WebSocket push test
	wsGroup := router.Group("/ws")
	wsGroup.Get("/push", PushToUser)
	wsGroup.Get("/broadcast", Broadcast)
	wsGroup.Get("/online", GetOnlineCount)

	// MQ test
	SetupMQ(router)
}

// Ping health check
func Ping(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{
		"message": "pong",
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})
}

// LimitTest rate limit test
func LimitTest(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{
		"message": "rate limit test passed",
		"limit":   "5 times/10 seconds",
	})
}

// LimitPathTest rate limit by path test
func LimitPathTest(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{
		"message": "rate limit by path test passed",
		"limit":   "3 times/10 seconds",
	})
}

// LimitIPTest rate limit by IP test
func LimitIPTest(c *fiber.Ctx) error {
	return response.OK(c, fiber.Map{
		"message": "rate limit by IP test passed",
		"limit":   "10 times/minute",
		"ip":      c.IP(),
	})
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
