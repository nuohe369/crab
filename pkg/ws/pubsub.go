package ws

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

// RedisClient defines the Redis client interface.
//
// pkg/ws doesn't directly depend on pkg/redis, but defines its own interface.
// This allows pkg/ws to be independently packaged as long as the client implements this interface.
type RedisClient interface {
	Publish(ctx context.Context, channel string, message any) *redis.IntCmd
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub
}

// EnableCluster enables cluster mode.
//
// Pass in a Redis client, Hub will automatically subscribe to the specified channel,
// and call DeliverLocal to deliver messages to local connections when received.
//
// Parameters:
//   - ctx: Context (for canceling subscription)
//   - rdb: Redis client (must implement RedisClient interface)
//   - channel: Channel name to subscribe
//
// Usage:
//
//	hub := ws.NewHub()
//	go hub.Run()
//	hub.EnableCluster(ctx, redisClient, "ws:user")
//
// Note:
//   - Must be called after Hub.Run()
//   - Subscription will be closed when ctx is canceled
func (h *Hub) EnableCluster(ctx context.Context, rdb RedisClient, channel string) {
	if rdb == nil {
		log.Println("ws: redis client is nil, cluster mode disabled")
		return
	}

	h.redis = rdb
	h.channel = channel

	go h.subscribeLoop(ctx, channel)

	log.Printf("ws: cluster mode enabled, channel: %s", channel)
}

// subscribeLoop is the subscription loop
func (h *Hub) subscribeLoop(ctx context.Context, channel string) {
	sub := h.redis.Subscribe(ctx, channel)
	if sub == nil {
		log.Printf("ws: failed to subscribe channel: %s", channel)
		return
	}
	defer sub.Close()

	log.Printf("ws: subscribed to channel: %s", channel)

	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			log.Printf("ws: unsubscribed from channel: %s", channel)
			return
		case msg, ok := <-ch:
			if !ok {
				log.Printf("ws: channel closed: %s", channel)
				return
			}
			if msg == nil {
				continue
			}

			wsMsg, err := ParseMessage([]byte(msg.Payload))
			if err != nil {
				log.Printf("ws: invalid message on %s: %v", channel, err)
				continue
			}
			h.DeliverLocal(wsMsg)
		}
	}
}

// Publish publishes message to Redis channel.
//
// In cluster mode, messages are broadcast to all nodes via Redis Pub/Sub.
// In standalone mode (EnableCluster not called), messages are delivered locally.
//
// Parameters:
//   - ctx: Context
//   - msg: Message to publish
//
// Returns:
//   - error: Publish error
//
// Usage:
//
//	// Send to specific user
//	hub.Publish(ctx, ws.NewMessage(123, "notify", payload))
//
//	// Broadcast to all users
//	hub.Publish(ctx, ws.NewBroadcast("system", payload))
func (h *Hub) Publish(ctx context.Context, msg *Message) error {
	if h.redis == nil {
		h.DeliverLocal(msg)
		return nil
	}

	return h.redis.Publish(ctx, h.channel, msg.Bytes()).Err()
}

// PublishToUser publishes message to specific user.
//
// This is a convenience method of Publish that automatically sets UserID.
//
// Parameters:
//   - ctx: Context
//   - userID: Target user ID (0 means broadcast)
//   - msg: Message to publish
func (h *Hub) PublishToUser(ctx context.Context, userID int64, msg *Message) error {
	msg.UserID = userID
	return h.Publish(ctx, msg)
}
