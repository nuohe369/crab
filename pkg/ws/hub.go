package ws

import (
	"log"
	"sync"
)

// Hub is the WebSocket connection pool.
//
// Hub is the core of pkg/ws, responsible for managing a group of WebSocket connections.
//
// Main features:
// 1. Register/unregister connections
// 2. Maintain userID → Client mapping
// 3. Local message broadcasting
// 4. Redis Pub/Sub cluster support (optional)
//
// Usage:
//
//	hub := ws.NewHub()
//	go hub.Run()  // Must run in goroutine
//
// Cluster mode:
//
//	hub.EnableCluster(ctx, redisClient, "ws:user")
//	hub.Publish(ctx, msg)  // Auto broadcast via Redis
type Hub struct {
	// clients stores all connected clients
	// key: *Client, value: true
	clients map[*Client]bool

	// userClients maps user ID to clients
	// Supports same user with multiple device logins
	userClients map[int64]map[*Client]bool

	// register channel for registration
	register chan *Client

	// unregister channel for unregistration
	unregister chan *Client

	// broadcast channel for broadcasting
	broadcast chan []byte

	// opts configuration options
	opts *Options

	// mu protects clients and userClients read/write
	mu sync.RWMutex

	// OnMessage is the message handler
	// Called when receiving client message
	OnMessage func(client *Client, msg *Message)

	// OnConnect is the connection established callback
	OnConnect func(client *Client)

	// OnDisconnect is the connection closed callback
	OnDisconnect func(client *Client)

	// redis is the Redis client (cluster mode)
	redis RedisClient

	// channel is the Redis channel name (cluster mode)
	channel string
}

// NewHub creates a Hub.
//
// Parameters:
//   - opts: optional configuration options
//
// Example:
//
//	hub := ws.NewHub()
//	hub := ws.NewHub(ws.WithPingInterval(20 * time.Second))
func NewHub(opts ...Option) *Hub {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	return &Hub{
		clients:     make(map[*Client]bool),
		userClients: make(map[int64]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan []byte, 256),
		opts:        options,
	}
}

// Run starts the Hub event loop.
//
// This is a blocking method, must run in goroutine.
// All Hub operations are serialized through channels for thread safety.
//
// Usage:
//
//	hub := ws.NewHub()
//	go hub.Run()
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.addClient(client)

		case client := <-h.unregister:
			h.removeClient(client)

		case message := <-h.broadcast:
			h.broadcastLocal(message)
		}
	}
}

// addClient adds client (internal method)
func (h *Hub) addClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true

	// Add to user mapping
	if client.UserID > 0 {
		if h.userClients[client.UserID] == nil {
			h.userClients[client.UserID] = make(map[*Client]bool)
		}
		h.userClients[client.UserID][client] = true
	}

	log.Printf("ws: client %d connected, total: %d", client.UserID, len(h.clients))

	// Trigger connect callback
	if h.OnConnect != nil {
		go h.OnConnect(client)
	}
}

// removeClient removes client (internal method)
func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; !ok {
		return
	}

	delete(h.clients, client)
	close(client.send)

	// Remove from user mapping
	if client.UserID > 0 {
		if clients, ok := h.userClients[client.UserID]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.userClients, client.UserID)
			}
		}
	}

	log.Printf("ws: client %d disconnected, total: %d", client.UserID, len(h.clients))

	// Trigger disconnect callback
	if h.OnDisconnect != nil {
		go h.OnDisconnect(client)
	}
}

// broadcastLocal broadcasts locally (internal method)
func (h *Hub) broadcastLocal(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			// Send queue full, skip
		}
	}
}

// Register registers a client.
//
// Adds client to Hub management.
// Usually called after WebSocket connection is established.
//
// Parameters:
//   - client: client to register
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client.
//
// Removes client from Hub.
// Usually called when WebSocket connection is closed.
//
// Parameters:
//   - client: client to unregister
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast broadcasts message to all connections.
//
// This is local broadcast, only sends to connections managed by current Hub.
// In cluster mode, use Publish instead.
//
// Parameters:
//   - msg: message to broadcast
func (h *Hub) Broadcast(msg *Message) {
	h.broadcast <- msg.Bytes()
}

// SendToUser sends message to specific user.
//
// If user has multiple connections (multiple devices), sends to all connections.
// This is local send, only sends to connections managed by current Hub.
// In cluster mode, use Publish instead.
//
// Parameters:
//   - userID: target user ID
//   - msg: message to send
//
// Returns:
//   - bool: whether user was found (sent to at least one connection)
func (h *Hub) SendToUser(userID int64, msg *Message) bool {
	h.mu.RLock()
	clients, ok := h.userClients[userID]
	h.mu.RUnlock()

	if !ok || len(clients) == 0 {
		return false
	}

	data := msg.Bytes()
	for client := range clients {
		client.SendBytes(data)
	}
	return true
}

// DeliverLocal delivers message locally.
//
// Decides to broadcast or send to specific user based on message UserID.
// This is called when receiving message from Redis Pub/Sub.
//
// Parameters:
//   - msg: message to deliver
func (h *Hub) DeliverLocal(msg *Message) {
	if msg.UserID == 0 {
		// Broadcast
		h.Broadcast(msg)
	} else {
		// Send to specific user
		h.SendToUser(msg.UserID, msg)
	}
}

// ClientCount returns current connection count
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// UserCount returns current online user count
func (h *Hub) UserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.userClients)
}

// IsUserOnline checks if user is online
func (h *Hub) IsUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.userClients[userID]
	return ok && len(clients) > 0
}

// GetUserClients gets all connections of a user
func (h *Hub) GetUserClients(userID int64) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.userClients[userID]
	if !ok {
		return nil
	}

	result := make([]*Client, 0, len(clients))
	for client := range clients {
		result = append(result, client)
	}
	return result
}
