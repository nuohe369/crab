// Package ws provides WebSocket connection management and message broadcasting
// Package ws 提供 WebSocket 连接管理和消息广播
package ws

import (
	"log"
	"sync"
)

// Hub is the WebSocket connection pool.
// Hub 是 WebSocket 连接池
//
// Hub is the core of pkg/ws, responsible for managing a group of WebSocket connections.
// Hub 是 pkg/ws 的核心，负责管理一组 WebSocket 连接
//
// Main features | 主要功能:
// 1. Register/unregister connections | 注册/注销连接
// 2. Maintain userID → Client mapping | 维护 userID → Client 映射
// 3. Local message broadcasting | 本地消息广播
// 4. Redis Pub/Sub cluster support (optional) | Redis Pub/Sub 集群支持（可选）
//
// Usage | 使用方法:
//
//	hub := ws.NewHub()
//	go hub.Run()  // Must run in goroutine | 必须在 goroutine 中运行
//
// Cluster mode | 集群模式:
//
//	hub.EnableCluster(ctx, redisClient, "ws:user")
//	hub.Publish(ctx, msg)  // Auto broadcast via Redis | 通过 Redis 自动广播
type Hub struct {
	clients      map[*Client]bool                   // Stores all connected clients | 存储所有已连接的客户端
	userClients  map[int64]map[*Client]bool         // Maps user ID to clients | 用户 ID 到客户端的映射
	register     chan *Client                       // Register channel | 注册通道
	unregister   chan *Client                       // Unregister channel | 注销通道
	broadcast    chan []byte                        // Broadcast channel | 广播通道
	opts         *Options                           // Configuration options | 配置选项
	mu           sync.RWMutex                       // Protects clients and userClients | 保护 clients 和 userClients
	OnMessage    func(client *Client, msg *Message) // Message handler | 消息处理器
	OnConnect    func(client *Client)               // Connection established callback | 连接建立回调
	OnDisconnect func(client *Client)               // Connection closed callback | 连接关闭回调
	redis        RedisClient                        // Redis client (cluster mode) | Redis 客户端（集群模式）
	channel      string                             // Redis channel name (cluster mode) | Redis 频道名称（集群模式）
}

// NewHub creates a Hub.
// NewHub 创建 Hub
//
// Parameters | 参数:
//   - opts: optional configuration options | 可选配置选项
//
// Example | 示例:
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
// Run 启动 Hub 事件循环
//
// This is a blocking method, must run in goroutine.
// All Hub operations are serialized through channels for thread safety.
// 这是一个阻塞方法，必须在 goroutine 中运行
// 所有 Hub 操作通过通道序列化以保证线程安全
//
// Usage | 使用方法:
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
// addClient 添加客户端（内部方法）
func (h *Hub) addClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true

	// Add to user mapping | 添加到用户映射
	if client.UserID > 0 {
		if h.userClients[client.UserID] == nil {
			h.userClients[client.UserID] = make(map[*Client]bool)
		}
		h.userClients[client.UserID][client] = true
	}

	log.Printf("ws: client %d connected, total: %d", client.UserID, len(h.clients))

	// Trigger connect callback | 触发连接回调
	if h.OnConnect != nil {
		go h.OnConnect(client)
	}
}

// removeClient removes client (internal method)
// removeClient 移除客户端（内部方法）
func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; !ok {
		return
	}

	delete(h.clients, client)
	close(client.send)

	// Remove from user mapping | 从用户映射中移除
	if client.UserID > 0 {
		if clients, ok := h.userClients[client.UserID]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.userClients, client.UserID)
			}
		}
	}

	log.Printf("ws: client %d disconnected, total: %d", client.UserID, len(h.clients))

	// Trigger disconnect callback | 触发断开连接回调
	if h.OnDisconnect != nil {
		go h.OnDisconnect(client)
	}
}

// broadcastLocal broadcasts locally (internal method)
// broadcastLocal 本地广播（内部方法）
func (h *Hub) broadcastLocal(message []byte) {
	// Quickly copy client list, reduce lock holding time | 快速复制客户端列表，减少锁持有时间
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	// Release lock before sending messages | 释放锁后再发送消息
	dropped := 0
	for _, client := range clients {
		select {
		case client.send <- message:
		default:
			// Send queue full, skip | 发送队列已满，跳过
			dropped++
		}
	}

	// Log dropped message count | 记录丢弃的消息数量
	if dropped > 0 {
		log.Printf("ws: broadcast dropped %d messages (send buffer full)", dropped)
	}
}

// Register registers a client.
// Register 注册客户端
//
// Adds client to Hub management.
// Usually called after WebSocket connection is established.
// 将客户端添加到 Hub 管理
// 通常在 WebSocket 连接建立后调用
//
// Parameters | 参数:
//   - client: client to register | 要注册的客户端
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister unregisters a client.
// Unregister 注销客户端
//
// Removes client from Hub.
// Usually called when WebSocket connection is closed.
// 从 Hub 中移除客户端
// 通常在 WebSocket 连接关闭时调用
//
// Parameters | 参数:
//   - client: client to unregister | 要注销的客户端
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// Broadcast broadcasts message to all connections.
// Broadcast 向所有连接广播消息
//
// This is local broadcast, only sends to connections managed by current Hub.
// In cluster mode, use Publish instead.
// 这是本地广播，仅发送到当前 Hub 管理的连接
// 在集群模式下，请使用 Publish
//
// Parameters | 参数:
//   - msg: message to broadcast | 要广播的消息
func (h *Hub) Broadcast(msg *Message) {
	h.broadcast <- msg.Bytes()
}

// SendToUser sends message to specific user.
// SendToUser 向特定用户发送消息
//
// If user has multiple connections (multiple devices), sends to all connections.
// This is local send, only sends to connections managed by current Hub.
// In cluster mode, use Publish instead.
// 如果用户有多个连接（多设备），则发送到所有连接
// 这是本地发送，仅发送到当前 Hub 管理的连接
// 在集群模式下，请使用 Publish
//
// Parameters | 参数:
//   - userID: target user ID | 目标用户 ID
//   - msg: message to send | 要发送的消息
//
// Returns | 返回:
//   - bool: whether user was found (sent to at least one connection) | 是否找到用户（至少发送到一个连接）
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
// DeliverLocal 本地投递消息
//
// Decides to broadcast or send to specific user based on message UserID.
// This is called when receiving message from Redis Pub/Sub.
// 根据消息的 UserID 决定广播或发送给特定用户
// 这在从 Redis Pub/Sub 接收消息时调用
//
// Parameters | 参数:
//   - msg: message to deliver | 要投递的消息
func (h *Hub) DeliverLocal(msg *Message) {
	if msg.UserID == 0 {
		// Broadcast | 广播
		h.Broadcast(msg)
	} else {
		// Send to specific user | 发送给特定用户
		h.SendToUser(msg.UserID, msg)
	}
}

// ClientCount returns current connection count
// ClientCount 返回当前连接数
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// UserCount returns current online user count
// UserCount 返回当前在线用户数
func (h *Hub) UserCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.userClients)
}

// IsUserOnline checks if user is online
// IsUserOnline 检查用户是否在线
func (h *Hub) IsUserOnline(userID int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients, ok := h.userClients[userID]
	return ok && len(clients) > 0
}

// GetUserClients gets all connections of a user
// GetUserClients 获取用户的所有连接
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
