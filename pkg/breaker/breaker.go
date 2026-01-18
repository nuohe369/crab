package breaker

import (
	"sync"
	"time"

	"github.com/nuohe369/crab/pkg/logger"
	"github.com/sony/gobreaker/v2"
)

var log = logger.NewSystem("breaker")

// Config represents circuit breaker configuration
// Config 表示熔断器配置
type Config struct {
	MaxRequests  uint32        `toml:"max_requests"`  // Max requests in half-open state (default 3) | 半开状态最大请求数（默认 3）
	Interval     time.Duration `toml:"interval"`      // Statistics interval, 0 means no reset (default 0) | 统计间隔，0 表示不重置（默认 0）
	Timeout      time.Duration `toml:"timeout"`       // Circuit breaker timeout duration (default 30s) | 熔断器超时时间（默认 30 秒）
	FailureRatio float64       `toml:"failure_ratio"` // Failure ratio to trigger circuit breaker (default 0.6) | 触发熔断的失败率（默认 0.6）
	MinRequests  uint32        `toml:"min_requests"`  // Min requests to trigger circuit breaker (default 5) | 触发熔断的最小请求数（默认 5）
}

// DefaultConfig returns default configuration
// DefaultConfig 返回默认配置
func DefaultConfig() Config {
	return Config{
		MaxRequests:  3,
		Interval:     0,
		Timeout:      30 * time.Second,
		FailureRatio: 0.6,
		MinRequests:  5,
	}
}

// State represents circuit breaker state
// State 表示熔断器状态
type State = gobreaker.State

const (
	StateClosed   = gobreaker.StateClosed   // Closed state | 关闭状态
	StateHalfOpen = gobreaker.StateHalfOpen // Half-open state | 半开状态
	StateOpen     = gobreaker.StateOpen     // Open state | 开启状态
)

// CircuitBreaker wraps gobreaker circuit breaker
// CircuitBreaker 封装 gobreaker 熔断器
type CircuitBreaker struct {
	cb   *gobreaker.CircuitBreaker[any]
	name string
}

// New creates a new circuit breaker
// New 创建新的熔断器
func New(name string, config Config) *CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Not enough requests, don't trip | 请求数不足，不触发熔断
			if counts.Requests < config.MinRequests {
				return false
			}
			// Failure ratio exceeds threshold, trip the breaker | 失败率超过阈值，触发熔断
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= config.FailureRatio
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Info("Circuit breaker [%s] state changed: %s -> %s", name, from, to)
		},
	}

	return &CircuitBreaker{
		cb:   gobreaker.NewCircuitBreaker[any](settings),
		name: name,
	}
}

// Execute executes the function with circuit breaker protection
// Execute 在熔断器保护下执行函数
func (c *CircuitBreaker) Execute(fn func() error) error {
	_, err := c.cb.Execute(func() (any, error) {
		return nil, fn()
	})
	return err
}

// ExecuteWithResult executes the function and returns the result
// ExecuteWithResult 执行函数并返回结果
func (c *CircuitBreaker) ExecuteWithResult(fn func() (any, error)) (any, error) {
	return c.cb.Execute(fn)
}

// State returns current state
// State 返回当前状态
func (c *CircuitBreaker) State() State {
	return c.cb.State()
}

// Name returns the name
// Name 返回名称
func (c *CircuitBreaker) Name() string {
	return c.name
}

// Counts returns statistics
// Counts 返回统计信息
func (c *CircuitBreaker) Counts() gobreaker.Counts {
	return c.cb.Counts()
}

// Manager manages multiple circuit breakers
// Manager 管理多个熔断器
type Manager struct {
	breakers map[string]*CircuitBreaker
	config   Config
	mu       sync.RWMutex
}

var (
	defaultManager *Manager
	managerOnce    sync.Once
)

// Init initializes the circuit breaker manager
// Init 初始化熔断器管理器
func Init(config Config) {
	managerOnce.Do(func() {
		defaultManager = &Manager{
			breakers: make(map[string]*CircuitBreaker),
			config:   config,
		}
		log.Info("Circuit breaker manager initialized (gobreaker): maxRequests=%d, timeout=%v, failureRatio=%.2f",
			config.MaxRequests, config.Timeout, config.FailureRatio)
	})
}

// GetManager returns the manager instance
// GetManager 返回管理器实例
func GetManager() *Manager {
	if defaultManager == nil {
		Init(DefaultConfig())
	}
	return defaultManager
}

// GetBreaker gets or creates a circuit breaker
// GetBreaker 获取或创建熔断器
func GetBreaker(name string) *CircuitBreaker {
	return GetManager().Get(name)
}

// Get gets or creates a circuit breaker
// Get 获取或创建熔断器
func (m *Manager) Get(name string) *CircuitBreaker {
	m.mu.RLock()
	if cb, ok := m.breakers[name]; ok {
		m.mu.RUnlock()
		return cb
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double check | 双重检查
	if cb, ok := m.breakers[name]; ok {
		return cb
	}

	cb := New(name, m.config)
	m.breakers[name] = cb
	log.Debug("Created circuit breaker: %s", name)
	return cb
}

// GetAll returns all circuit breakers
// GetAll 返回所有熔断器
func (m *Manager) GetAll() map[string]*CircuitBreaker {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*CircuitBreaker, len(m.breakers))
	for k, v := range m.breakers {
		result[k] = v
	}
	return result
}

// Stats represents circuit breaker statistics
// Stats 表示熔断器统计信息
type Stats struct {
	Name     string
	State    State
	Requests uint32
	Failures uint32
	Success  uint32
}

// GetAllStats returns statistics for all circuit breakers
// GetAllStats 返回所有熔断器的统计信息
func (m *Manager) GetAllStats() []Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make([]Stats, 0, len(m.breakers))
	for _, cb := range m.breakers {
		counts := cb.Counts()
		stats = append(stats, Stats{
			Name:     cb.Name(),
			State:    cb.State(),
			Requests: counts.Requests,
			Failures: counts.TotalFailures,
			Success:  counts.TotalSuccesses,
		})
	}
	return stats
}

// SetConfig updates configuration (only affects newly created circuit breakers)
// SetConfig 更新配置（仅影响新创建的熔断器）
func (m *Manager) SetConfig(config Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config
}

// IsOpen checks if the error is circuit breaker open error
// IsOpen 检查错误是否为熔断器开启错误
func IsOpen(err error) bool {
	return err == gobreaker.ErrOpenState
}

// IsTooManyRequests checks if the error is too many requests in half-open state
// IsTooManyRequests 检查错误是否为半开状态请求过多
func IsTooManyRequests(err error) bool {
	return err == gobreaker.ErrTooManyRequests
}
