package health

import (
	"context"
	"sync"
	"time"

	"github.com/nuohe369/crab/pkg/logger"
)

var log = logger.NewSystem("health")

// Status represents health status
type Status string

const (
	StatusUp      Status = "UP"      // Healthy
	StatusDown    Status = "DOWN"    // Unhealthy
	StatusUnknown Status = "UNKNOWN" // Unknown
)

// CheckResult represents a single check result
type CheckResult struct {
	Status  Status         `json:"status"`
	Message string         `json:"message,omitempty"`
	Details map[string]any `json:"details,omitempty"`
}

// Checker is the health checker interface
type Checker interface {
	Name() string
	Check(ctx context.Context) CheckResult
}

// CheckFunc is the check function type
type CheckFunc func(ctx context.Context) CheckResult

// FuncChecker is a function-based checker
type FuncChecker struct {
	name string
	fn   CheckFunc
}

func (f *FuncChecker) Name() string {
	return f.name
}

func (f *FuncChecker) Check(ctx context.Context) CheckResult {
	return f.fn(ctx)
}

// NewChecker creates a function-based checker
func NewChecker(name string, fn CheckFunc) Checker {
	return &FuncChecker{name: name, fn: fn}
}

// Health is the health check manager
type Health struct {
	checkers []Checker
	timeout  time.Duration
	mu       sync.RWMutex
}

// Result represents the overall health check result
type Result struct {
	Status  Status                 `json:"status"`
	Checks  map[string]CheckResult `json:"checks,omitempty"`
	Version string                 `json:"version,omitempty"`
}

var defaultHealth *Health

// Init initializes health check
func Init(timeout time.Duration) {
	defaultHealth = &Health{
		checkers: make([]Checker, 0),
		timeout:  timeout,
	}
	log.Info("Health check initialized with timeout %v", timeout)
}

// Get returns the default instance
func Get() *Health {
	if defaultHealth == nil {
		Init(5 * time.Second)
	}
	return defaultHealth
}

// Register registers a checker
func Register(checker Checker) {
	Get().Register(checker)
}

// RegisterFunc registers a check function
func RegisterFunc(name string, fn CheckFunc) {
	Get().RegisterFunc(name, fn)
}

// Register registers a checker
func (h *Health) Register(checker Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers = append(h.checkers, checker)
	log.Debug("Registered health checker: %s", checker.Name())
}

// RegisterFunc registers a check function
func (h *Health) RegisterFunc(name string, fn CheckFunc) {
	h.Register(NewChecker(name, fn))
}

// Check executes all health checks
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

	// Create context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	// Execute checks concurrently
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
func (h *Health) Liveness(ctx context.Context) Result {
	return Result{Status: StatusUp}
}

// Readiness returns readiness check result (K8s readiness probe)
func (h *Health) Readiness(ctx context.Context) Result {
	return h.Check(ctx)
}

// Check executes check using default instance
func Check(ctx context.Context) Result {
	return Get().Check(ctx)
}

// Liveness executes liveness check using default instance
func Liveness(ctx context.Context) Result {
	return Get().Liveness(ctx)
}

// Readiness executes readiness check using default instance
func Readiness(ctx context.Context) Result {
	return Get().Readiness(ctx)
}
