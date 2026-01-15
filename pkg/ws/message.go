package ws

import "github.com/bytedance/sonic"

// Message represents WebSocket message structure.
//
// This is the core data structure of pkg/ws package, used for:
// 1. Client → Server messages
// 2. Server → Client messages
// 3. Redis Pub/Sub messages (cluster mode)
type Message struct {
	// UserID is the target user ID.
	// - 0 means broadcast to all connections
	// - >0 means send to specific user
	// Note: Client messages don't need to set this field
	UserID int64 `json:"user_id,omitempty"`

	// Type is the message type for distinguishing different business logic.
	// Recommended to use snake_case naming, such as:
	// - "ping" / "pong" for heartbeat
	// - "chat_message" for chat messages
	// - "order_paid" for order payment notifications
	// - "system_notice" for system announcements
	Type string `json:"type"`

	// Payload is the message content, can be any JSON serializable data.
	// Recommended to use map[string]any or custom struct
	Payload any `json:"payload,omitempty"`
}

// NewMessage creates a message for specific user
func NewMessage(userID int64, msgType string, payload any) *Message {
	return &Message{
		UserID:  userID,
		Type:    msgType,
		Payload: payload,
	}
}

// NewBroadcast creates a broadcast message for all users
func NewBroadcast(msgType string, payload any) *Message {
	return &Message{
		UserID:  0,
		Type:    msgType,
		Payload: payload,
	}
}

// Bytes serializes message to JSON bytes
func (m *Message) Bytes() []byte {
	data, _ := sonic.Marshal(m)
	return data
}

// ParseMessage parses message from JSON bytes
func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	if err := sonic.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
