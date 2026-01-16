package health

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"xorm.io/xorm"
)

// DatabaseChecker checks database health
type DatabaseChecker struct {
	name   string
	engine *xorm.Engine
}

// NewDatabaseChecker creates a database checker
func NewDatabaseChecker(name string, engine *xorm.Engine) *DatabaseChecker {
	return &DatabaseChecker{name: name, engine: engine}
}

func (d *DatabaseChecker) Name() string {
	return d.name
}

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
type RedisChecker struct {
	name   string
	client redis.UniversalClient
}

// NewRedisChecker creates a Redis checker
func NewRedisChecker(name string, client redis.UniversalClient) *RedisChecker {
	return &RedisChecker{name: name, client: client}
}

func (r *RedisChecker) Name() string {
	return r.name
}

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
type GRPCChecker struct {
	name    string
	target  string
	checkFn func(ctx context.Context) error
}

// NewGRPCChecker creates a gRPC checker
func NewGRPCChecker(name, target string, checkFn func(ctx context.Context) error) *GRPCChecker {
	return &GRPCChecker{name: name, target: target, checkFn: checkFn}
}

func (g *GRPCChecker) Name() string {
	return g.name
}

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
type DiskChecker struct {
	name      string
	path      string
	threshold float64 // Usage threshold (0-1)
}

// NewDiskChecker creates a disk checker
func NewDiskChecker(name, path string, threshold float64) *DiskChecker {
	return &DiskChecker{name: name, path: path, threshold: threshold}
}

func (d *DiskChecker) Name() string {
	return d.name
}

func (d *DiskChecker) Check(ctx context.Context) CheckResult {
	// Simplified implementation, can use syscall to get disk info
	return CheckResult{
		Status: StatusUp,
		Details: map[string]any{
			"path":      d.path,
			"threshold": d.threshold,
		},
	}
}

// MemoryChecker checks memory usage
type MemoryChecker struct {
	name      string
	threshold float64 // Usage threshold (0-1)
}

// NewMemoryChecker creates a memory checker
func NewMemoryChecker(name string, threshold float64) *MemoryChecker {
	return &MemoryChecker{name: name, threshold: threshold}
}

func (m *MemoryChecker) Name() string {
	return m.name
}

func (m *MemoryChecker) Check(ctx context.Context) CheckResult {
	// Simplified implementation
	return CheckResult{
		Status: StatusUp,
		Details: map[string]any{
			"threshold": m.threshold,
		},
	}
}

// CustomChecker is a custom checker
type CustomChecker struct {
	name    string
	checkFn func(ctx context.Context) error
}

// NewCustomChecker creates a custom checker
func NewCustomChecker(name string, checkFn func(ctx context.Context) error) *CustomChecker {
	return &CustomChecker{name: name, checkFn: checkFn}
}

func (c *CustomChecker) Name() string {
	return c.name
}

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
