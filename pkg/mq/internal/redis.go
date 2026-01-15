package internal

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// RedisStreamsConfig Redis Streams configuration
type RedisStreamsConfig struct {
	Addr     string
	Password string
	DB       int
	Cluster  string
	MaxLen   int64
}

// RedisStreams Redis Streams implementation
type RedisStreams struct {
	client redis.UniversalClient
	maxLen int64
}

// NewRedisStreams creates a Redis Streams client
func NewRedisStreams(cfg RedisStreamsConfig) (*RedisStreams, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var client redis.UniversalClient

	if cfg.Cluster != "" {
		// Cluster mode
		addrs := strings.Split(cfg.Cluster, ",")
		for i := range addrs {
			addrs[i] = strings.TrimSpace(addrs[i])
		}

		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    addrs,
			Password: cfg.Password,
		})
		log.Printf("mq: redis streams cluster mode, nodes: %v", addrs)
	} else {
		// Standalone mode
		client = redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		})
		log.Printf("mq: redis streams standalone mode, address: %s", cfg.Addr)
	}

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("mq: redis connection failed: %w", err)
	}

	return &RedisStreams{
		client: client,
		maxLen: cfg.MaxLen,
	}, nil
}

// Publish publishes a message to Stream (immediately consumable)
func (r *RedisStreams) Publish(ctx context.Context, topic string, payload []byte) error {
	args := &redis.XAddArgs{
		Stream: topic,
		Values: map[string]any{"payload": payload},
	}

	// Set max length (approximate trimming)
	if r.maxLen > 0 {
		args.MaxLen = r.maxLen
		args.Approx = true
	}

	_, err := r.client.XAdd(ctx, args).Result()
	return err
}

// delayKey returns the Sorted Set key for delay queue
func delayKey(topic string) string {
	return topic + ":delay"
}

// PublishDelay publishes a delayed message
// Uses Sorted Set storage, score is the expiration timestamp
func (r *RedisStreams) PublishDelay(ctx context.Context, topic string, payload []byte, delay time.Duration) error {
	executeAt := time.Now().Add(delay).UnixMilli()
	member := fmt.Sprintf("%s:%s", uuid.New().String(), string(payload))

	return r.client.ZAdd(ctx, delayKey(topic), redis.Z{
		Score:  float64(executeAt),
		Member: member,
	}).Err()
}

// Message represents a message structure
type Message struct {
	ID      string
	Topic   string
	Payload []byte
}

// Consume consumes messages (both immediate and expired delayed messages)
func (r *RedisStreams) Consume(ctx context.Context, topic, group string, handler func(ctx context.Context, msg *Message) error) error {
	// Create consumer group if not exists
	err := r.client.XGroupCreateMkStream(ctx, topic, group, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("mq: failed to create consumer group: %w", err)
	}

	consumerName := fmt.Sprintf("%s-%d", group, time.Now().UnixNano())

	// Start delayed message transfer goroutine
	go r.transferDelayMessages(ctx, topic)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 读cancel息
		streams, err := r.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumerName,
			Streams:  []string{topic, ">"},
			Count:    10,
			Block:    time.Second * 5,
		}).Result()

		if err != nil {
			if err == redis.Nil {
				continue
			}
			log.Printf("mq: failed to read messages: %v", err)
			time.Sleep(time.Second)
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				payload, ok := msg.Values["payload"].(string)
				if !ok {
					log.Printf("mq: invalid message format: %v", msg.Values)
					continue
				}

				m := &Message{
					ID:      msg.ID,
					Topic:   topic,
					Payload: []byte(payload),
				}

				if err := handler(ctx, m); err != nil {
					log.Printf("mq: failed to process message: %v", err)
					// Don't Ack, message will be redelivered
					continue
				}

				// Process success, auto Ack
				r.client.XAck(ctx, topic, group, msg.ID)
			}
		}
	}
}

// transferDelayMessages transfers expired delayed messages to Stream
func (r *RedisStreams) transferDelayMessages(ctx context.Context, topic string) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.doTransfer(ctx, topic)
		}
	}
}

// doTransfer executes one transfer
func (r *RedisStreams) doTransfer(ctx context.Context, topic string) {
	now := float64(time.Now().UnixMilli())
	key := delayKey(topic)

	// Get all expired messages
	results, err := r.client.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: "-inf",
		Max: fmt.Sprintf("%f", now),
	}).Result()

	if err != nil {
		log.Printf("mq: failed to query delayed messages: %v", err)
		return
	}

	for _, z := range results {
		member := z.Member.(string)

		// Parse payload (format: uuid:payload)
		parts := strings.SplitN(member, ":", 2)
		if len(parts) != 2 {
			log.Printf("mq: invalid delayed message format: %s", member)
			r.client.ZRem(ctx, key, member)
			continue
		}
		payload := parts[1]

		// Transfer to Stream
		err := r.Publish(ctx, topic, []byte(payload))
		if err != nil {
			log.Printf("mq: failed to transfer delayed message: %v", err)
			continue
		}

		// Delete transferred message
		r.client.ZRem(ctx, key, member)
	}
}

// Ack acknowledges a message
func (r *RedisStreams) Ack(ctx context.Context, topic, group, msgID string) error {
	return r.client.XAck(ctx, topic, group, msgID).Err()
}

// Close closes the connection
func (r *RedisStreams) Close() error {
	return r.client.Close()
}

// GetRaw returns the underlying Redis client
func (r *RedisStreams) GetRaw() any {
	return r.client
}
