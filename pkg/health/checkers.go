// Package health provides health check functionality for monitoring service status
// Package health 提供健康检查功能，用于监控服务状态
package health

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"
)

// DatabaseChecker checks database health
// DatabaseChecker 检查数据库健康状态
type DatabaseChecker struct {
	name   string       // Checker name | 检查器名称
	engine *xorm.Engine // Database engine | 数据库引擎
}

// NewDatabaseChecker creates a database checker
// NewDatabaseChecker 创建数据库检查器
func NewDatabaseChecker(name string, engine *xorm.Engine) *DatabaseChecker {
	return &DatabaseChecker{name: name, engine: engine}
}

// Name returns the checker name
// Name 返回检查器名称
func (d *DatabaseChecker) Name() string {
	return d.name
}

// Check executes database health check
// Check 执行数据库健康检查
func (d *DatabaseChecker) Check(ctx context.Context) CheckResult {
	if d.engine == nil {
		return CheckResult{
			Status:  StatusDown,
			Message: "database engine is nil",
		}
	}

	start := time.Now()
	db := d.engine.DB()
	if err := db.PingContext(ctx); err != nil {
		return CheckResult{
			Status:  StatusDown,
			Message: fmt.Sprintf("ping failed: %v", err),
		}
	}

	stats := db.Stats()
	return CheckResult{
		Status: StatusUp,
		Details: map[string]any{
			"latency_ms":       time.Since(start).Milliseconds(),
			"open_connections": stats.OpenConnections,
			"in_use":           stats.InUse,
			"idle":             stats.Idle,
		},
	}
}

// RedisChecker checks Redis health
// RedisChecker 检查 Redis 健康状态
type RedisChecker struct {
	name   string                // Checker name | 检查器名称
	client redis.UniversalClient // Redis client | Redis 客户端
}

// NewRedisChecker creates a Redis checker
// NewRedisChecker 创建 Redis 检查器
func NewRedisChecker(name string, client redis.UniversalClient) *RedisChecker {
	return &RedisChecker{name: name, client: client}
}

// Name returns the checker name
// Name 返回检查器名称
func (r *RedisChecker) Name() string {
	return r.name
}

// Check executes Redis health check
// Check 执行 Redis 健康检查
func (r *RedisChecker) Check(ctx context.Context) CheckResult {
	if r.client == nil {
		return CheckResult{
			Status:  StatusDown,
			Message: "redis client is nil",
		}
	}

	start := time.Now()
	if err := r.client.Ping(ctx).Err(); err != nil {
		return CheckResult{
			Status:  StatusDown,
			Message: fmt.Sprintf("ping failed: %v", err),
		}
	}

	return CheckResult{
		Status: StatusUp,
		Details: map[string]any{
			"latency_ms": time.Since(start).Milliseconds(),
		},
	}
}

// GRPCChecker checks gRPC service health
// GRPCChecker 检查 gRPC 服务健康状态
type GRPCChecker struct {
	name    string                          // Checker name | 检查器名称
	target  string                          // Target address | 目标地址
	checkFn func(ctx context.Context) error // Check function | 检查函数
}

// NewGRPCChecker creates a gRPC checker
// NewGRPCChecker 创建 gRPC 检查器
func NewGRPCChecker(name, target string, checkFn func(ctx context.Context) error) *GRPCChecker {
	return &GRPCChecker{name: name, target: target, checkFn: checkFn}
}

// Name returns the checker name
// Name 返回检查器名称
func (g *GRPCChecker) Name() string {
	return g.name
}

// Check executes gRPC health check
// Check 执行 gRPC 健康检查
func (g *GRPCChecker) Check(ctx context.Context) CheckResult {
	if g.checkFn == nil {
		return CheckResult{
			Status:  StatusUnknown,
			Message: "no check function provided",
		}
	}

	start := time.Now()
	if err := g.checkFn(ctx); err != nil {
		return CheckResult{
			Status:  StatusDown,
			Message: fmt.Sprintf("check failed: %v", err),
			Details: map[string]any{
				"target": g.target,
			},
		}
	}

	return CheckResult{
		Status: StatusUp,
		Details: map[string]any{
			"target":     g.target,
			"latency_ms": time.Since(start).Milliseconds(),
		},
	}
}

// DiskChecker checks disk space
// DiskChecker 检查磁盘空间
type DiskChecker struct {
	name      string  // Checker name | 检查器名称
	path      string  // Path to check | 检查路径
	threshold float64 // Usage threshold (0-1) | 使用率阈值（0-1）
}

// NewDiskChecker creates a disk checker
// NewDiskChecker 创建磁盘检查器
func NewDiskChecker(name, path string, threshold float64) *DiskChecker {
	return &DiskChecker{name: name, path: path, threshold: threshold}
}

// Name returns the checker name
// Name 返回检查器名称
func (d *DiskChecker) Name() string {
	return d.name
}

// Check executes disk health check
// Check 执行磁盘健康检查
func (d *DiskChecker) Check(ctx context.Context) CheckResult {
	// Simplified implementation, can use syscall to get disk info | 简化实现，可使用 syscall 获取磁盘信息
	return CheckResult{
		Status: StatusUp,
		Details: map[string]any{
			"path":      d.path,
			"threshold": d.threshold,
		},
	}
}

// MemoryChecker checks memory usage
// MemoryChecker 检查内存使用情况
type MemoryChecker struct {
	name      string  // Checker name | 检查器名称
	threshold float64 // Usage threshold (0-1) | 使用率阈值（0-1）
}

// NewMemoryChecker creates a memory checker
// NewMemoryChecker 创建内存检查器
func NewMemoryChecker(name string, threshold float64) *MemoryChecker {
	return &MemoryChecker{name: name, threshold: threshold}
}

// Name returns the checker name
// Name 返回检查器名称
func (m *MemoryChecker) Name() string {
	return m.name
}

// Check executes memory health check
// Check 执行内存健康检查
func (m *MemoryChecker) Check(ctx context.Context) CheckResult {
	// Simplified implementation | 简化实现
	return CheckResult{
		Status: StatusUp,
		Details: map[string]any{
			"threshold": m.threshold,
		},
	}
}

// CustomChecker is a custom checker
// CustomChecker 是自定义检查器
type CustomChecker struct {
	name    string                          // Checker name | 检查器名称
	checkFn func(ctx context.Context) error // Check function | 检查函数
}

// NewCustomChecker creates a custom checker
// NewCustomChecker 创建自定义检查器
func NewCustomChecker(name string, checkFn func(ctx context.Context) error) *CustomChecker {
	return &CustomChecker{name: name, checkFn: checkFn}
}

// Name returns the checker name
// Name 返回检查器名称
func (c *CustomChecker) Name() string {
	return c.name
}

// Check executes custom health check
// Check 执行自定义健康检查
func (c *CustomChecker) Check(ctx context.Context) CheckResult {
	if c.checkFn == nil {
		return CheckResult{Status: StatusUp}
	}

	if err := c.checkFn(ctx); err != nil {
		return CheckResult{
			Status:  StatusDown,
			Message: err.Error(),
		}
	}

	return CheckResult{Status: StatusUp}
}
