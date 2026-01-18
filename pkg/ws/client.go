// Package ws provides WebSocket connection management and message broadcasting
// Package ws 提供 WebSocket 连接管理和消息广播
package ws

import (
	"log"
	"time"

	"github.com/gofiber/websocket/v2"
)

// Client represents a WebSocket client connection.
// Client 表示 WebSocket 客户端连接
//
// Each WebSocket connection corresponds to a Client instance, responsible for:
// 每个 WebSocket 连接对应一个 Client 实例，负责：
// 1. Maintaining connection state | 维护连接状态
// 2. Reading client messages (ReadPump) | 读取客户端消息（ReadPump）
// 3. Sending messages to client (WritePump) | 向客户端发送消息（WritePump）
// 4. Heartbeat detection | 心跳检测
//
// Lifecycle | 生命周期:
//
//	conn := websocket.Conn  // Get from Fiber | 从 Fiber 获取
//	client := ws.NewClient(hub, userID, conn)
//	hub.Register(client)
//	defer hub.Unregister(client)
//	go client.WritePump()
//	client.ReadPump()  // Blocks until connection closes | 阻塞直到连接关闭
type Client struct {
	hub    *Hub            // The connection pool this client belongs to | 此客户端所属的连接池
	UserID int64           // User ID (0 means unauthenticated) | 用户 ID（0 表示未认证）
	Conn   *websocket.Conn // WebSocket connection | WebSocket 连接
	send   chan []byte     // Channel for sending messages | 发送消息的通道
}

// NewClient creates a client.
// NewClient 创建客户端
//
// Parameters | 参数:
//   - hub: the Hub this client belongs to | 此客户端所属的 Hub
//   - userID: user ID (0 means unauthenticated) | 用户 ID（0 表示未认证）
//   - conn: WebSocket connection | WebSocket 连接
//
// Returns | 返回:
//   - *Client: client instance | 客户端实例
func NewClient(hub *Hub, userID int64, conn *websocket.Conn) *Client {
	return &Client{
		hub:    hub,
		UserID: userID,
		Conn:   conn,
		send:   make(chan []byte, hub.opts.SendBuffer),
	}
}

// Send sends message to client.
// Send 向客户端发送消息
//
// This is a non-blocking operation, message is put into send queue.
// If queue is full, message is dropped (to avoid slow clients blocking the system).
// 这是非阻塞操作，消息放入发送队列
// 如果队列已满，消息将被丢弃（避免慢客户端阻塞系统）
//
// Parameters | 参数:
//   - msg: message to send | 要发送的消息
func (c *Client) Send(msg *Message) {
	select {
	case c.send <- msg.Bytes():
	default:
		// Send queue full, drop message | 发送队列已满，丢弃消息
		log.Printf("ws: client %d send buffer full, message dropped", c.UserID)
	}
}

// SendBytes sends raw bytes
// SendBytes 发送原始字节
func (c *Client) SendBytes(data []byte) {
	select {
	case c.send <- data:
	default:
		log.Printf("ws: client %d send buffer full, message dropped", c.UserID)
	}
}

// ReadPump reads client messages.
// ReadPump 读取客户端消息
//
// This is a blocking method, runs until:
// 这是阻塞方法，运行直到：
// 1. Client disconnects | 客户端断开连接
// 2. Read timeout (no heartbeat) | 读取超时（无心跳）
// 3. Message format error | 消息格式错误
//
// Usage | 使用方法:
//
//	go client.WritePump()
//	client.ReadPump()  // Blocks here | 在此阻塞
//
// Note: Must start WritePump before calling ReadPump
// 注意：必须在调用 ReadPump 之前启动 WritePump
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.Conn.Close()
	}()

	// Set read limit | 设置读取限制
	c.Conn.SetReadLimit(c.hub.opts.MaxMessageSize)

	// Set read timeout | 设置读取超时
	c.Conn.SetReadDeadline(time.Now().Add(c.hub.opts.ReadTimeout))

	// Set pong handler (reset timeout when receiving pong) | 设置 pong 处理器（收到 pong 时重置超时）
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(c.hub.opts.ReadTimeout))
		return nil
	})

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			// Connection closed or read error | 连接关闭或读取错误
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
				websocket.CloseNormalClosure) {
				log.Printf("ws: client %d read error: %v", c.UserID, err)
			}
			break
		}

		// Reset read timeout | 重置读取超时
		c.Conn.SetReadDeadline(time.Now().Add(c.hub.opts.ReadTimeout))

		// Parse message | 解析消息
		msg, err := ParseMessage(data)
		if err != nil {
			log.Printf("ws: client %d invalid message: %v", c.UserID, err)
			continue
		}

		// Call message handler | 调用消息处理器
		if c.hub.OnMessage != nil {
			c.hub.OnMessage(c, msg)
		}
	}
}

// WritePump sends messages to client.
// WritePump 向客户端发送消息
//
// This is a blocking method, must run in goroutine.
// Responsible for:
// 这是阻塞方法，必须在 goroutine 中运行
// 负责：
// 1. Reading messages from send channel and sending | 从发送通道读取消息并发送
// 2. Periodically sending ping heartbeat | 定期发送 ping 心跳
//
// Usage | 使用方法:
//
//	go client.WritePump()
//	client.ReadPump()
func (c *Client) WritePump() {
	ticker := time.NewTicker(c.hub.opts.PingInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			// Set write timeout | 设置写入超时
			c.Conn.SetWriteDeadline(time.Now().Add(c.hub.opts.WriteTimeout))

			if !ok {
				// Hub closed send channel | Hub 关闭了发送通道
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send message | 发送消息
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("ws: client %d write error: %v", c.UserID, err)
				return
			}

		case <-ticker.C:
			// Send ping heartbeat | 发送 ping 心跳
			c.Conn.SetWriteDeadline(time.Now().Add(c.hub.opts.WriteTimeout))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Close closes the connection
// Close 关闭连接
func (c *Client) Close() {
	c.Conn.Close()
}
