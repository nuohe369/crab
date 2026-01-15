package handler

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/nuohe369/crab/common/response"
	"github.com/nuohe369/crab/pkg/mq"
)

// Store recently consumed messages (for test verification)
var (
	consumedMessages []ConsumedMessage
	consumedMu       sync.RWMutex
	maxMessages      = 100 // Keep at most 100 messages
)

// ConsumedMessage consumed message
type ConsumedMessage struct {
	ID        string    `json:"id"`
	Payload   string    `json:"payload"`
	ConsumeAt time.Time `json:"consume_at"`
}

// SetupMQ registers MQ test routes
func SetupMQ(router fiber.Router) {
	if !mq.Enabled() {
		log.Println("testapi: mq not enabled, skipping MQ test routes")
		return
	}

	mqGroup := router.Group("/mq")
	mqGroup.Get("/publish", MQPublish)
	mqGroup.Get("/publish-delay", MQPublishDelay)
	mqGroup.Get("/consumed", MQConsumed)
	mqGroup.Get("/status", MQStatus)

	// Start consumer
	go startConsumer()
}

// startConsumer starts test consumer
func startConsumer() {
	ctx := context.Background()
	topic := "testapi:demo"
	group := "testapi-consumer"

	log.Printf("testapi: starting MQ consumer, topic=%s, group=%s", topic, group)

	err := mq.Consume(ctx, topic, group, func(ctx context.Context, msg *mq.Message) error {
		log.Printf("testapi: received message id=%s, payload=%s", msg.ID, string(msg.Payload))

		// Save to memory
		consumedMu.Lock()
		consumedMessages = append(consumedMessages, ConsumedMessage{
			ID:        msg.ID,
			Payload:   string(msg.Payload),
			ConsumeAt: time.Now(),
		})
		// Remove oldest if exceeds limit
		if len(consumedMessages) > maxMessages {
			consumedMessages = consumedMessages[1:]
		}
		consumedMu.Unlock()

		return nil
	})

	if err != nil {
		log.Printf("testapi: MQ consumer exited: %v", err)
	}
}

// MQPublish publishes test message
//
// GET /testapi/mq/publish?content=hello
func MQPublish(c *fiber.Ctx) error {
	content := c.Query("content", "test message")

	payload, _ := json.Marshal(map[string]any{
		"content": content,
		"time":    time.Now().Format("2006-01-02 15:04:05"),
	})

	err := mq.Publish(context.Background(), "testapi:demo", payload)
	if err != nil {
		return response.FailMsg(c, response.CodeServerError, err.Error())
	}

	return response.OK(c, fiber.Map{
		"message": "publish success",
		"topic":   "testapi:demo",
		"content": content,
	})
}

// MQPublishDelay publishes delayed message
//
// GET /testapi/mq/publish-delay?content=hello&delay=30
// delay: delay in seconds, default 30 seconds
func MQPublishDelay(c *fiber.Ctx) error {
	content := c.Query("content", "delayed message")
	delaySec := c.QueryInt("delay", 30)

	payload, _ := json.Marshal(map[string]any{
		"content":    content,
		"publish_at": time.Now().Format("2006-01-02 15:04:05"),
		"delay_sec":  delaySec,
	})

	delay := time.Duration(delaySec) * time.Second
	err := mq.PublishDelay(context.Background(), "testapi:demo", payload, delay)
	if err != nil {
		return response.FailMsg(c, response.CodeServerError, err.Error())
	}

	executeAt := time.Now().Add(delay)
	return response.OK(c, fiber.Map{
		"message":    "delayed message publish success",
		"topic":      "testapi:demo",
		"content":    content,
		"delay_sec":  delaySec,
		"execute_at": executeAt.Format("2006-01-02 15:04:05"),
	})
}

// MQConsumed returns consumed messages
//
// GET /testapi/mq/consumed?limit=10
func MQConsumed(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	if limit > maxMessages {
		limit = maxMessages
	}

	consumedMu.RLock()
	defer consumedMu.RUnlock()

	// Return latest N messages (reverse order)
	total := len(consumedMessages)
	start := total - limit
	if start < 0 {
		start = 0
	}

	result := make([]ConsumedMessage, 0, limit)
	for i := total - 1; i >= start; i-- {
		result = append(result, consumedMessages[i])
	}

	return response.OK(c, fiber.Map{
		"total":    total,
		"messages": result,
	})
}

// MQStatus returns MQ status
//
// GET /testapi/mq/status
func MQStatus(c *fiber.Ctx) error {
	consumedMu.RLock()
	consumedCount := len(consumedMessages)
	consumedMu.RUnlock()

	return response.OK(c, fiber.Map{
		"enabled":        mq.Enabled(),
		"driver":         "redis",
		"consumed_count": consumedCount,
	})
}
