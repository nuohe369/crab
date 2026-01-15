package internal

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQConfig RabbitMQ configuration
type RabbitMQConfig struct {
	URL string
}

// RabbitMQ RabbitMQ implementation
type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	mu      sync.Mutex
}

// NewRabbitMQ creates a RabbitMQ client
func NewRabbitMQ(cfg RabbitMQConfig) (*RabbitMQ, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("mq: rabbitmq url cannot be empty")
	}

	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("mq: rabbitmq connection failed: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("mq: rabbitmq create channel failed: %w", err)
	}

	log.Printf("mq: rabbitmq connected: %s", cfg.URL)

	return &RabbitMQ{
		conn:    conn,
		channel: ch,
	}, nil
}

// ensureQueue ensures the queue exists
func (r *RabbitMQ) ensureQueue(topic string) error {
	_, err := r.channel.QueueDeclare(
		topic, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	return err
}

// Publish publishes a message
func (r *RabbitMQ) Publish(ctx context.Context, topic string, payload []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.ensureQueue(topic); err != nil {
		return fmt.Errorf("mq: failed to create queue: %w", err)
	}

	return r.channel.PublishWithContext(ctx,
		"",    // exchange
		topic, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/octet-stream",
			Body:         payload,
		},
	)
}

// delayQueue returns the delay queue name (one queue per delay duration)
func delayQueue(topic string, delay time.Duration) string {
	return fmt.Sprintf("%s.delay.%ds", topic, int(delay.Seconds()))
}

// ensureDelayQueue ensures the delay queue exists (using TTL + DLX)
func (r *RabbitMQ) ensureDelayQueue(topic string, delay time.Duration) (string, error) {
	// Ensure target queue exists
	if err := r.ensureQueue(topic); err != nil {
		return "", err
	}

	// Declare delay queue (with TTL and dead letter exchange)
	// One independent queue per delay duration to avoid TTL conflicts
	delayQ := delayQueue(topic, delay)
	_, err := r.channel.QueueDeclare(
		delayQ,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-message-ttl":             int64(delay.Milliseconds()),
			"x-dead-letter-exchange":    "",    // Default exchange
			"x-dead-letter-routing-key": topic, // Route to target queue on expiration
		},
	)
	return delayQ, err
}

// PublishDelay publishes a delayed message
func (r *RabbitMQ) PublishDelay(ctx context.Context, topic string, payload []byte, delay time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delayQ, err := r.ensureDelayQueue(topic, delay)
	if err != nil {
		return fmt.Errorf("mq: failed to create delay queue: %w", err)
	}

	// Send to delay queue
	return r.channel.PublishWithContext(ctx,
		"",     // exchange
		delayQ, // routing key (延迟队列)
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/octet-stream",
			Body:         payload,
		},
	)
}

// Consume consumes messages
func (r *RabbitMQ) Consume(ctx context.Context, topic, group string, handler func(ctx context.Context, msg *Message) error) error {
	// Create independent channel for consumer
	ch, err := r.conn.Channel()
	if err != nil {
		return fmt.Errorf("mq: failed to create consumer channel: %w", err)
	}
	defer ch.Close()

	// Ensure queue exists
	_, err = ch.QueueDeclare(
		topic,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("mq: failed to create queue: %w", err)
	}

	// Set QoS
	if err := ch.Qos(10, 0, false); err != nil {
		return fmt.Errorf("mq: failed to set QoS: %w", err)
	}

	msgs, err := ch.Consume(
		topic, // queue
		group, // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("mq: consume failed: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("mq: channel closed")
			}

			m := &Message{
				ID:      d.MessageId,
				Topic:   topic,
				Payload: d.Body,
			}

			if err := handler(ctx, m); err != nil {
				log.Printf("mq: failed to process message: %v", err)
				// Nack, requeue
				d.Nack(false, true)
				continue
			}

			// Process success, Ack
			d.Ack(false)
		}
	}
}

// Ack acknowledges a message (RabbitMQ handles this in Consume)
func (r *RabbitMQ) Ack(ctx context.Context, topic, group, msgID string) error {
	// RabbitMQ Ack is handled via delivery.Ack() in Consume
	return nil
}

// Close closes the connection
func (r *RabbitMQ) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// GetRaw returns the underlying channel
func (r *RabbitMQ) GetRaw() any {
	return r.channel
}
