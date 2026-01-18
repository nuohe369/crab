// Package service provides business service layer functionality
// service 包提供业务服务层功能
package service

import (
	"context"

	"github.com/nuohe369/crab/pkg/logger"
	"github.com/nuohe369/crab/pkg/redis"
	"github.com/nuohe369/crab/pkg/ws"
)

var wsLog = logger.NewSystem("ws")

// ============================================================
// WebSocket Service Wrapper | WebSocket 服务包装器
//
// This is a business layer wrapper for pkg/ws, responsible for:
// 这是 pkg/ws 的业务层包装器，负责：
// 1. Managing multiple Hubs (user-side, admin-side, etc.)
//    管理多个 Hub（用户端、管理端等）
// 2. Providing simple send interfaces for other modules
//    为其他模块提供简单的发送接口
// 3. Initializing Redis cluster mode
//    初始化 Redis 集群模式
//
// Usage | 用法:
//
//	// Initialize (called during boot phase) | 初始化（在启动阶段调用）
//	service.InitWS()
//
//	// Send messages from other modules | 从其他模块发送消息
//	service.PublishToUser(ctx, 123, &ws.Message{Type: "notify", Payload: xxx})
//	service.PublishToAdmin(ctx, 0, &ws.Message{Type: "broadcast", Payload: xxx})
//
// ============================================================

// Redis channel names | Redis 频道名称
const (
	channelUser  = "ws:user"  // User-side channel | 用户端频道
	channelAdmin = "ws:admin" // Admin-side channel | 管理端频道
)

var (
	// userHub is the user-side Hub | userHub 用户端 Hub
	userHub *ws.Hub

	// adminHub is the admin-side Hub | adminHub 管理端 Hub
	adminHub *ws.Hub

	// ctx is the global context for canceling subscriptions | ctx 用于取消订阅的全局上下文
	ctx    context.Context
	cancel context.CancelFunc
)

// InitWS initializes the WebSocket service
// Called during the boot phase to start all Hubs and Redis subscriptions
// If Redis is not initialized, only local mode (single instance) is started
// InitWS 初始化 WebSocket 服务
// 在启动阶段调用，启动所有 Hub 和 Redis 订阅
// 如果 Redis 未初始化，则仅启动本地模式（单实例）
func InitWS() {
	ctx, cancel = context.WithCancel(context.Background())

	// Create Hubs | 创建 Hub
	userHub = ws.NewHub()
	adminHub = ws.NewHub()

	// Start Hub event loops | 启动 Hub 事件循环
	go userHub.Run()
	go adminHub.Run()

	// Enable cluster mode if Redis is available | 如果 Redis 可用，启用集群模式
	if rdb := redis.Get(); rdb != nil {
		rawClient := rdb.GetRaw()
		// Both standalone and cluster clients implement ws.RedisClient interface
		// 独立和集群客户端都实现了 ws.RedisClient 接口
		if client, ok := rawClient.(ws.RedisClient); ok {
			userHub.EnableCluster(ctx, client, channelUser)
			adminHub.EnableCluster(ctx, client, channelAdmin)
			wsLog.Info("Cluster mode enabled (redis pub/sub)")
		} else {
			wsLog.Warn("Redis client does not support Pub/Sub")
		}
	} else {
		wsLog.Info("Standalone mode (no redis)")
	}
}

// CloseWS closes the WebSocket service
// Called during service shutdown to cancel all Redis subscriptions
// CloseWS 关闭 WebSocket 服务
// 在服务关闭时调用，取消所有 Redis 订阅
func CloseWS() {
	if cancel != nil {
		cancel()
	}
}

// ============================================================
// User-side Hub | 用户端 Hub
// ============================================================

// GetUserHub returns the user-side Hub
// Used by module/ws_user to register connections
// GetUserHub 返回用户端 Hub
// 由 module/ws_user 使用以注册连接
func GetUserHub() *ws.Hub {
	return userHub
}

// PublishToUser sends a message to a user
// This is the main interface called by other modules
// In cluster mode, messages are broadcast to all nodes via Redis
// PublishToUser 向用户发送消息
// 这是其他模块调用的主要接口
// 在集群模式下，消息通过 Redis 广播到所有节点
//
// Parameters | 参数:
//   - ctx: context | 上下文
//   - userID: target user ID (0 means broadcast to all users) | 目标用户 ID（0 表示广播给所有用户）
//   - msg: message to send | 要发送的消息
//
// Example | 示例:
//
//	// Send to a single user | 发送给单个用户
//	service.PublishToUser(ctx, 123, &ws.Message{
//	    Type:    "order_paid",
//	    Payload: map[string]any{"order_id": "xxx"},
//	})
//
//	// Broadcast to all users | 广播给所有用户
//	service.PublishToUser(ctx, 0, &ws.Message{
//	    Type:    "system_notice",
//	    Payload: "Server maintenance notification",
//	})
func PublishToUser(ctx context.Context, userID int64, msg *ws.Message) error {
	if userHub == nil {
		return nil
	}
	return userHub.PublishToUser(ctx, userID, msg)
}

// IsUserOnline checks if a user is online (local node only)
// Note: In cluster mode, this only checks the local node, not other nodes
// IsUserOnline 检查用户是否在线（仅本地节点）
// 注意：在集群模式下，这仅检查本地节点，不检查其他节点
func IsUserOnline(userID int64) bool {
	if userHub == nil {
		return false
	}
	return userHub.IsUserOnline(userID)
}

// GetUserOnlineCount returns the number of online users (local node only)
// GetUserOnlineCount 返回在线用户数（仅本地节点）
func GetUserOnlineCount() int {
	if userHub == nil {
		return 0
	}
	return userHub.UserCount()
}

// ============================================================
// Admin-side Hub | 管理端 Hub
// ============================================================

// GetAdminHub returns the admin-side Hub
// Used by module/ws_admin to register connections
// GetAdminHub 返回管理端 Hub
// 由 module/ws_admin 使用以注册连接
func GetAdminHub() *ws.Hub {
	return adminHub
}

// PublishToAdmin sends a message to an admin
// PublishToAdmin 向管理员发送消息
//
// Parameters | 参数:
//   - ctx: context | 上下文
//   - adminID: target admin ID (0 means broadcast to all admins) | 目标管理员 ID（0 表示广播给所有管理员）
//   - msg: message to send | 要发送的消息
func PublishToAdmin(ctx context.Context, adminID int64, msg *ws.Message) error {
	if adminHub == nil {
		return nil
	}
	return adminHub.PublishToUser(ctx, adminID, msg)
}

// IsAdminOnline checks if an admin is online (local node only)
// IsAdminOnline 检查管理员是否在线（仅本地节点）
func IsAdminOnline(adminID int64) bool {
	if adminHub == nil {
		return false
	}
	return adminHub.IsUserOnline(adminID)
}

// GetAdminOnlineCount returns the number of online admins (local node only)
// GetAdminOnlineCount 返回在线管理员数（仅本地节点）
func GetAdminOnlineCount() int {
	if adminHub == nil {
		return 0
	}
	return adminHub.UserCount()
}
