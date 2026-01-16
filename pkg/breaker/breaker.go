package breaker

import (
	"sync"
	"time"

	"github.com/nuohe369/crab/pkg/logger"
	"github.com/sony/gobreaker/v2"
)

var log = logger.NewSystem("breaker")

// Config represents circuit breaker configuration
type Config struct {
	MaxRequests  uint32        `toml:"max_requests"`  // Max requests in half-open state (default 3)
	Interval     time.Duration `toml:"interval"`      // Statistics interval, 0 means no reset (default 0)
	Timeout      time.Duration `toml:"timeout"`       // Circuit breaker timeout duration (default 30s)
	FailureRatio float64       `toml:"failure_ratio"` // Failure ratio to trigger circuit breaker (default 0.6)
	MinRequests  uint32        `toml:"min_requests"`  // Min requests to trigger circuit breaker (default 5)
}

// DefaultConfig returns default configuration
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
type State = gobreaker.State

const (
	StateClosed   = gobreaker.StateClosed
	StateHalfOpen = gobreaker.StateHalfOpen
	StateOpen     = gobreaker.StateOpen
)

// CircuitBreaker wraps gobreaker circuit breaker
type CircuitBreaker struct {
	cb   *gobreaker.CircuitBreaker[any]
	name string
}

// New creates a new circuit breaker
func New(name string, config Config) *CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Not enough requests, don't trip
			if counts.Requests < config.MinRequests {
				return false
			}
			// Failure ratio exceeds threshold, trip the breaker
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
func (c *CircuitBreaker) Execute(fn func() error) error {
	_, err := c.cb.Execute(func() (any, error) {
		return nil, fn()
	})
	return err
}

// ExecuteWithResult executes the function and returns the result
func (c *CircuitBreaker) ExecuteWithResult(fn func() (any, error)) (any, error) {
	return c.cb.Execute(fn)
}

// State returns current state
func (c *CircuitBreaker) State() State {
	return c.cb.State()
}

// Name returns the name
func (c *CircuitBreaker) Name() string {
	return c.name
}

// Counts returns statistics
func (c *CircuitBreaker) Counts() gobreaker.Counts {
	return c.cb.Counts()
}

// Manager manages multiple circuit breakers
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
func GetManager() *Manager {
	if defaultManager == nil {
		Init(DefaultConfig())
	}
	return defaultManager
}

// GetBreaker gets or creates a circuit breaker
func GetBreaker(name string) *CircuitBreaker {
	return GetManager().Get(name)
}

// Get gets or creates a circuit breaker
func (m *Manager) Get(name string) *CircuitBreaker {
	m.mu.RLock()
	if cb, ok := m.breakers[name]; ok {
		m.mu.RUnlock()
		return cb
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double check
	if cb, ok := m.breakers[name]; ok {
		return cb
	}

	cb := New(name, m.config)
	m.breakers[name] = cb
	log.Debug("Created circuit breaker: %s", name)
	return cb
}

// GetAll returns all circuit breakers
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
type Stats struct {
	Name     string
	State    State
	Requests uint32
	Failures uint32
	Success  uint32
}

// GetAllStats returns statistics for all circuit breakers
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
func (m *Manager) SetConfig(config Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config
}

// IsOpen checks if the error is circuit breaker open error
func IsOpen(err error) bool {
	return err == gobreaker.ErrOpenState
}

// IsTooManyRequests checks if the error is too many requests in half-open state
func IsTooManyRequests(err error) bool {
	return err == gobreaker.ErrTooManyRequests
}
