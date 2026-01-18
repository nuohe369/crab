// Package ws provides WebSocket connection management and message broadcasting
// Package ws 提供 WebSocket 连接管理和消息广播
package ws

import "github.com/bytedance/sonic"

// Message represents WebSocket message structure.
// Message 表示 WebSocket 消息结构
//
// This is the core data structure of pkg/ws package, used for:
// 这是 pkg/ws 包的核心数据结构，用于：
// 1. Client → Server messages | 客户端 → 服务器消息
// 2. Server → Client messages | 服务器 → 客户端消息
// 3. Redis Pub/Sub messages (cluster mode) | Redis Pub/Sub 消息（集群模式）
type Message struct {
	UserID  int64  `json:"user_id,omitempty"` // Target user ID (0 means broadcast) | 目标用户 ID（0 表示广播）
	Type    string `json:"type"`              // Message type | 消息类型
	Payload any    `json:"payload,omitempty"` // Message content | 消息内容
}

// NewMessage creates a message for specific user
// NewMessage 创建发送给特定用户的消息
func NewMessage(userID int64, msgType string, payload any) *Message {
	return &Message{
		UserID:  userID,
		Type:    msgType,
		Payload: payload,
	}
}

// NewBroadcast creates a broadcast message for all users
// NewBroadcast 创建广播给所有用户的消息
func NewBroadcast(msgType string, payload any) *Message {
	return &Message{
		UserID:  0,
		Type:    msgType,
		Payload: payload,
	}
}

// Bytes serializes message to JSON bytes
// Bytes 将消息序列化为 JSON 字节
func (m *Message) Bytes() []byte {
	data, _ := sonic.Marshal(m)
	return data
}

// ParseMessage parses message from JSON bytes
// ParseMessage 从 JSON 字节解析消息
func ParseMessage(data []byte) (*Message, error) {
	var msg Message
	if err := sonic.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
