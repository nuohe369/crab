// Package health provides health check functionality for monitoring service status
// Package health 提供健康检查功能，用于监控服务状态
package health

import (
	"context"
	"sync"
	"time"

	"github.com/nuohe369/crab/pkg/logger"
)

var log = logger.NewSystem("health")

// Status represents health status
// Status 表示健康状态
type Status string

const (
	StatusUp      Status = "UP"      // Healthy | 健康
	StatusDown    Status = "DOWN"    // Unhealthy | 不健康
	StatusUnknown Status = "UNKNOWN" // Unknown | 未知
)

// CheckResult represents a single check result
// CheckResult 表示单个检查结果
type CheckResult struct {
	Status  Status         `json:"status"`            // Check status | 检查状态
	Message string         `json:"message,omitempty"` // Error message | 错误消息
	Details map[string]any `json:"details,omitempty"` // Additional details | 附加详情
}

// Checker is the health checker interface
// Checker 是健康检查器接口
type Checker interface {
	// Name returns the checker name
	// Name 返回检查器名称
	Name() string
	// Check executes the health check
	// Check 执行健康检查
	Check(ctx context.Context) CheckResult
}

// CheckFunc is the check function type
// CheckFunc 是检查函数类型
type CheckFunc func(ctx context.Context) CheckResult

// FuncChecker is a function-based checker
// FuncChecker 是基于函数的检查器
type FuncChecker struct {
	name string    // Checker name | 检查器名称
	fn   CheckFunc // Check function | 检查函数
}

// Name returns the checker name
// Name 返回检查器名称
func (f *FuncChecker) Name() string {
	return f.name
}

// Check executes the check function
// Check 执行检查函数
func (f *FuncChecker) Check(ctx context.Context) CheckResult {
	return f.fn(ctx)
}

// NewChecker creates a function-based checker
// NewChecker 创建基于函数的检查器
func NewChecker(name string, fn CheckFunc) Checker {
	return &FuncChecker{name: name, fn: fn}
}

// Health is the health check manager
// Health 是健康检查管理器
type Health struct {
	checkers []Checker     // Registered checkers | 已注册的检查器
	timeout  time.Duration // Check timeout | 检查超时时间
	mu       sync.RWMutex  // Mutex for concurrent access | 并发访问互斥锁
}

// Result represents the overall health check result
// Result 表示整体健康检查结果
type Result struct {
	Status  Status                 `json:"status"`            // Overall status | 整体状态
	Checks  map[string]CheckResult `json:"checks,omitempty"`  // Individual check results | 各个检查结果
	Version string                 `json:"version,omitempty"` // Service version | 服务版本
}

var defaultHealth *Health // Default health check instance | 默认健康检查实例

// Init initializes health check
// Init 初始化健康检查
func Init(timeout time.Duration) {
	defaultHealth = &Health{
		checkers: make([]Checker, 0),
		timeout:  timeout,
	}
	log.Info("Health check initialized with timeout %v", timeout)
}

// Get returns the default instance
// Get 返回默认实例
func Get() *Health {
	if defaultHealth == nil {
		Init(5 * time.Second)
	}
	return defaultHealth
}

// Register registers a checker
// Register 注册检查器
func Register(checker Checker) {
	Get().Register(checker)
}

// RegisterFunc registers a check function
// RegisterFunc 注册检查函数
func RegisterFunc(name string, fn CheckFunc) {
	Get().RegisterFunc(name, fn)
}

// Register registers a checker
// Register 注册检查器
func (h *Health) Register(checker Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers = append(h.checkers, checker)
	log.Debug("Registered health checker: %s", checker.Name())
}

// RegisterFunc registers a check function
// RegisterFunc 注册检查函数
func (h *Health) RegisterFunc(name string, fn CheckFunc) {
	h.Register(NewChecker(name, fn))
}

// Check executes all health checks
// Check 执行所有健康检查
func (h *Health) Check(ctx context.Context) Result {
	h.mu.RLock()
	checkers := make([]Checker, len(h.checkers))
	copy(checkers, h.checkers)
	h.mu.RUnlock()

	result := Result{
		Status: StatusUp,
		Checks: make(map[string]CheckResult),
	}

	if len(checkers) == 0 {
		return result
	}

	// Create context with timeout | 创建带超时的上下文
	checkCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	// Execute checks concurrently | 并发执行检查
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, checker := range checkers {
		wg.Add(1)
		go func(c Checker) {
			defer wg.Done()

			checkResult := c.Check(checkCtx)

			mu.Lock()
			result.Checks[c.Name()] = checkResult
			if checkResult.Status != StatusUp {
				result.Status = StatusDown
			}
			mu.Unlock()
		}(checker)
	}

	wg.Wait()
	return result
}

// Liveness returns liveness check result (K8s liveness probe)
// Liveness 返回存活检查结果（K8s 存活探针）
func (h *Health) Liveness(ctx context.Context) Result {
	return Result{Status: StatusUp}
}

// Readiness returns readiness check result (K8s readiness probe)
// Readiness 返回就绪检查结果（K8s 就绪探针）
func (h *Health) Readiness(ctx context.Context) Result {
	return h.Check(ctx)
}

// Check executes check using default instance
// Check 使用默认实例执行检查
func Check(ctx context.Context) Result {
	return Get().Check(ctx)
}

// Liveness executes liveness check using default instance
// Liveness 使用默认实例执行存活检查
func Liveness(ctx context.Context) Result {
	return Get().Liveness(ctx)
}

// Readiness executes readiness check using default instance
// Readiness 使用默认实例执行就绪检查
func Readiness(ctx context.Context) Result {
	return Get().Readiness(ctx)
}
