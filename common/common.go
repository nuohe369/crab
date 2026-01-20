// Package common provides common business layer functionality
// common 包提供通用业务层功能
package common

import (
	"github.com/nuohe369/crab/common/middleware"
	"github.com/nuohe369/crab/common/service"
	"github.com/nuohe369/crab/pkg/logger"
)

var log = logger.NewSystem("common")

// Init initializes the common business layer
// Init 初始化通用业务层
func Init() {
	log.Info("Initializing common business layer...")

	// Initialize rate limiter with Redis if available | 如果 Redis 可用，则初始化限流器
	middleware.InitRateLimiter()

	// Initialize WebSocket service | 初始化 WebSocket 服务
	service.InitWS()
}
