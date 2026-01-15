package ws

import (
	"log"
	"time"

	"github.com/gofiber/websocket/v2"
)

// Client represents a WebSocket client connection.
//
// Each WebSocket connection corresponds to a Client instance, responsible for:
// 1. Maintaining connection state
// 2. Reading client messages (ReadPump)
// 3. Sending messages to client (WritePump)
// 4. Heartbeat detection
//
// Lifecycle:
//
//	conn := websocket.Conn  // Get from Fiber
//	client := ws.NewClient(hub, userID, conn)
//	hub.Register(client)
//	defer hub.Unregister(client)
//	go client.WritePump()
//	client.ReadPump()  // Blocks until connection closes
type Client struct {
	// hub is the connection pool this client belongs to
	hub *Hub

	// UserID is the user ID
	// For targeted message sending, 0 means unauthenticated user
	UserID int64

	// Conn is the WebSocket connection
	Conn *websocket.Conn

	// send is the channel for sending messages
	// WritePump reads from here and sends to client
	send chan []byte
}

// NewClient creates a client.
//
// Parameters:
//   - hub: the Hub this client belongs to
//   - userID: user ID (0 means unauthenticated)
//   - conn: WebSocket connection
//
// Returns:
//   - *Client: client instance
func NewClient(hub *Hub, userID int64, conn *websocket.Conn) *Client {
	return &Client{
		hub:    hub,
		UserID: userID,
		Conn:   conn,
		send:   make(chan []byte, hub.opts.SendBuffer),
	}
}

// Send sends message to client.
//
// This is a non-blocking operation, message is put into send queue.
// If queue is full, message is dropped (to avoid slow clients blocking the system).
//
// Parameters:
//   - msg: message to send
func (c *Client) Send(msg *Message) {
	select {
	case c.send <- msg.Bytes():
	default:
		// Send queue full, drop message
		log.Printf("ws: client %d send buffer full, message dropped", c.UserID)
	}
}

// SendBytes sends raw bytes
func (c *Client) SendBytes(data []byte) {
	select {
	case c.send <- data:
	default:
		log.Printf("ws: client %d send buffer full, message dropped", c.UserID)
	}
}

// ReadPump reads client messages.
//
// This is a blocking method, runs until:
// 1. Client disconnects
// 2. Read timeout (no heartbeat)
// 3. Message format error
//
// Usage:
//
//	go client.WritePump()
//	client.ReadPump()  // Blocks here
//
// Note: Must start WritePump before calling ReadPump
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.Conn.Close()
	}()

	// Set read limit
	c.Conn.SetReadLimit(c.hub.opts.MaxMessageSize)

	// Set read timeout
	c.Conn.SetReadDeadline(time.Now().Add(c.hub.opts.ReadTimeout))

	// Set pong handler (reset timeout when receiving pong)
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(c.hub.opts.ReadTimeout))
		return nil
	})

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			// Connection closed or read error
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
				websocket.CloseNormalClosure) {
				log.Printf("ws: client %d read error: %v", c.UserID, err)
			}
			break
		}

		// Reset read timeout
		c.Conn.SetReadDeadline(time.Now().Add(c.hub.opts.ReadTimeout))

		// Parse message
		msg, err := ParseMessage(data)
		if err != nil {
			log.Printf("ws: client %d invalid message: %v", c.UserID, err)
			continue
		}

		// Call message handler
		if c.hub.OnMessage != nil {
			c.hub.OnMessage(c, msg)
		}
	}
}

// WritePump sends messages to client.
//
// This is a blocking method, must run in goroutine.
// Responsible for:
// 1. Reading messages from send channel and sending
// 2. Periodically sending ping heartbeat
//
// Usage:
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
			// Set write timeout
			c.Conn.SetWriteDeadline(time.Now().Add(c.hub.opts.WriteTimeout))

			if !ok {
				// Hub closed send channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send message
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("ws: client %d write error: %v", c.UserID, err)
				return
			}

		case <-ticker.C:
			// Send ping heartbeat
			c.Conn.SetWriteDeadline(time.Now().Add(c.hub.opts.WriteTimeout))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Close closes the connection
func (c *Client) Close() {
	c.Conn.Close()
}
