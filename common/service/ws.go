package service

import (
	"context"

	"github.com/nuohe369/crab/pkg/logger"
	"github.com/nuohe369/crab/pkg/redis"
	"github.com/nuohe369/crab/pkg/ws"
)

var wsLog = logger.NewSystem("ws")

// ============================================================
// WebSocket Service Wrapper
//
// This is a business layer wrapper for pkg/ws, responsible for:
// 1. Managing multiple Hubs (user-side, admin-side, etc.)
// 2. Providing simple send interfaces for other modules
// 3. Initializing Redis cluster mode
//
// Usage:
//
//	// Initialize (called during boot phase)
//	service.InitWS()
//
//	// Send messages from other modules
//	service.PublishToUser(ctx, 123, &ws.Message{Type: "notify", Payload: xxx})
//	service.PublishToAdmin(ctx, 0, &ws.Message{Type: "broadcast", Payload: xxx})
//
// ============================================================

// Redis channel names
const (
	channelUser  = "ws:user"  // user-side channel
	channelAdmin = "ws:admin" // admin-side channel
)

var (
	// userHub is the user-side Hub
	userHub *ws.Hub

	// adminHub is the admin-side Hub
	adminHub *ws.Hub

	// ctx is the global context for canceling subscriptions
	ctx    context.Context
	cancel context.CancelFunc
)

// InitWS initializes the WebSocket service.
//
// Called during the boot phase to start all Hubs and Redis subscriptions.
// If Redis is not initialized, only local mode (single instance) is started.
func InitWS() {
	ctx, cancel = context.WithCancel(context.Background())

	// Create Hubs
	userHub = ws.NewHub()
	adminHub = ws.NewHub()

	// Start Hub event loops
	go userHub.Run()
	go adminHub.Run()

	// Enable cluster mode if Redis is available
	if rdb := redis.Get(); rdb != nil {
		rawClient := rdb.GetRaw()
		// Both standalone and cluster clients implement ws.RedisClient interface
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

// CloseWS closes the WebSocket service.
//
// Called during service shutdown to cancel all Redis subscriptions.
func CloseWS() {
	if cancel != nil {
		cancel()
	}
}

// ============================================================
// User-side Hub
// ============================================================

// GetUserHub returns the user-side Hub.
//
// Used by module/ws_user to register connections.
func GetUserHub() *ws.Hub {
	return userHub
}

// PublishToUser sends a message to a user.
//
// This is the main interface called by other modules.
// In cluster mode, messages are broadcast to all nodes via Redis.
//
// Parameters:
//   - ctx: context
//   - userID: target user ID (0 means broadcast to all users)
//   - msg: message to send
//
// Example:
//
//	// Send to a single user
//	service.PublishToUser(ctx, 123, &ws.Message{
//	    Type:    "order_paid",
//	    Payload: map[string]any{"order_id": "xxx"},
//	})
//
//	// Broadcast to all users
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

// IsUserOnline checks if a user is online (local node only).
//
// Note: In cluster mode, this only checks the local node, not other nodes.
func IsUserOnline(userID int64) bool {
	if userHub == nil {
		return false
	}
	return userHub.IsUserOnline(userID)
}

// GetUserOnlineCount returns the number of online users (local node only).
func GetUserOnlineCount() int {
	if userHub == nil {
		return 0
	}
	return userHub.UserCount()
}

// ============================================================
// Admin-side Hub
// ============================================================

// GetAdminHub returns the admin-side Hub.
//
// Used by module/ws_admin to register connections.
func GetAdminHub() *ws.Hub {
	return adminHub
}

// PublishToAdmin sends a message to an admin.
//
// Parameters:
//   - ctx: context
//   - adminID: target admin ID (0 means broadcast to all admins)
//   - msg: message to send
func PublishToAdmin(ctx context.Context, adminID int64, msg *ws.Message) error {
	if adminHub == nil {
		return nil
	}
	return adminHub.PublishToUser(ctx, adminID, msg)
}

// IsAdminOnline checks if an admin is online (local node only).
func IsAdminOnline(adminID int64) bool {
	if adminHub == nil {
		return false
	}
	return adminHub.IsUserOnline(adminID)
}

// GetAdminOnlineCount returns the number of online admins (local node only).
func GetAdminOnlineCount() int {
	if adminHub == nil {
		return 0
	}
	return adminHub.UserCount()
}
