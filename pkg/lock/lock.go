package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/nuohe369/crab/pkg/logger"
	"github.com/redis/go-redis/v9"
)

var log = logger.NewSystem("lock")

// Config represents distributed lock configuration
type Config struct {
	Expiry        time.Duration `toml:"expiry"`         // Lock expiry time (default 8s)
	Tries         int           `toml:"tries"`          // Max retry attempts (default 32)
	RetryDelay    time.Duration `toml:"retry_delay"`    // Retry interval (default 100ms)
	DriftFactor   float64       `toml:"drift_factor"`   // Clock drift factor (default 0.01)
	TimeoutFactor float64       `toml:"timeout_factor"` // Timeout factor (default 0.05)
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Expiry:        8 * time.Second,
		Tries:         32,
		RetryDelay:    100 * time.Millisecond,
		DriftFactor:   0.01,
		TimeoutFactor: 0.05,
	}
}

// Locker is the distributed lock manager
type Locker struct {
	rs     *redsync.Redsync
	config Config
}

var defaultLocker *Locker

// Init initializes distributed lock
func Init(client redis.UniversalClient, cfg Config) {
	pool := goredis.NewPool(client)
	defaultLocker = &Locker{
		rs:     redsync.New(pool),
		config: cfg,
	}
	log.Info("Distributed lock initialized: expiry=%v, tries=%d", cfg.Expiry, cfg.Tries)
}

// Get returns the default Locker
func Get() *Locker {
	return defaultLocker
}

// Mutex represents a distributed mutex
type Mutex struct {
	mu     *redsync.Mutex
	name   string
	locked bool
}

// NewMutex creates a mutex
func (l *Locker) NewMutex(name string, opts ...Option) *Mutex {
	options := l.config
	for _, opt := range opts {
		opt(&options)
	}

	rsOpts := []redsync.Option{
		redsync.WithExpiry(options.Expiry),
		redsync.WithTries(options.Tries),
		redsync.WithRetryDelay(options.RetryDelay),
		redsync.WithDriftFactor(options.DriftFactor),
		redsync.WithTimeoutFactor(options.TimeoutFactor),
		redsync.WithGenValueFunc(func() (string, error) {
			return fmt.Sprintf("%d", time.Now().UnixNano()), nil
		}),
	}

	mu := l.rs.NewMutex(name, rsOpts...)

	return &Mutex{
		mu:   mu,
		name: name,
	}
}

// Lock acquires the lock
func (m *Mutex) Lock() error {
	if err := m.mu.Lock(); err != nil {
		log.Debug("Failed to acquire lock [%s]: %v", m.name, err)
		return err
	}
	m.locked = true
	log.Debug("Acquired lock [%s]", m.name)
	return nil
}

// LockContext acquires the lock with context
func (m *Mutex) LockContext(ctx context.Context) error {
	if err := m.mu.LockContext(ctx); err != nil {
		log.Debug("Failed to acquire lock [%s]: %v", m.name, err)
		return err
	}
	m.locked = true
	log.Debug("Acquired lock [%s]", m.name)
	return nil
}

// TryLock tries to acquire the lock (non-blocking)
func (m *Mutex) TryLock() (bool, error) {
	err := m.mu.TryLock()
	if err != nil {
		if err == redsync.ErrFailed {
			return false, nil // Lock is held by another
		}
		return false, err
	}
	m.locked = true
	log.Debug("Acquired lock [%s]", m.name)
	return true, nil
}

// TryLockContext tries to acquire the lock with context
func (m *Mutex) TryLockContext(ctx context.Context) (bool, error) {
	err := m.mu.TryLockContext(ctx)
	if err != nil {
		if err == redsync.ErrFailed {
			return false, nil
		}
		return false, err
	}
	m.locked = true
	log.Debug("Acquired lock [%s]", m.name)
	return true, nil
}

// Unlock releases the lock
func (m *Mutex) Unlock() (bool, error) {
	if !m.locked {
		return false, nil
	}
	ok, err := m.mu.Unlock()
	if ok {
		m.locked = false
		log.Debug("Released lock [%s]", m.name)
	}
	return ok, err
}

// UnlockContext releases the lock with context
func (m *Mutex) UnlockContext(ctx context.Context) (bool, error) {
	if !m.locked {
		return false, nil
	}
	ok, err := m.mu.UnlockContext(ctx)
	if ok {
		m.locked = false
		log.Debug("Released lock [%s]", m.name)
	}
	return ok, err
}

// Extend extends the lock expiry time
func (m *Mutex) Extend() (bool, error) {
	return m.mu.Extend()
}

// ExtendContext extends the lock expiry time with context
func (m *Mutex) ExtendContext(ctx context.Context) (bool, error) {
	return m.mu.ExtendContext(ctx)
}

// Name returns the lock name
func (m *Mutex) Name() string {
	return m.name
}

// IsLocked checks if the lock is held
func (m *Mutex) IsLocked() bool {
	return m.locked
}

// Until returns the lock expiry time
func (m *Mutex) Until() time.Time {
	return m.mu.Until()
}

// Option is a configuration option
type Option func(*Config)

// WithExpiry sets the expiry time
func WithExpiry(d time.Duration) Option {
	return func(c *Config) {
		c.Expiry = d
	}
}

// WithTries sets the retry attempts
func WithTries(n int) Option {
	return func(c *Config) {
		c.Tries = n
	}
}

// WithRetryDelay sets the retry interval
func WithRetryDelay(d time.Duration) Option {
	return func(c *Config) {
		c.RetryDelay = d
	}
}

// NewMutex creates a mutex using default Locker
func NewMutex(name string, opts ...Option) *Mutex {
	if defaultLocker == nil {
		panic("lock: not initialized, call lock.Init() first")
	}
	return defaultLocker.NewMutex(name, opts...)
}

// WithLock executes function under lock protection
func WithLock(ctx context.Context, name string, fn func() error, opts ...Option) error {
	mu := NewMutex(name, opts...)
	if err := mu.LockContext(ctx); err != nil {
		return err
	}
	defer mu.UnlockContext(ctx)
	return fn()
}

// TryWithLock tries to execute function under lock protection (non-blocking)
func TryWithLock(ctx context.Context, name string, fn func() error, opts ...Option) (bool, error) {
	mu := NewMutex(name, opts...)
	ok, err := mu.TryLockContext(ctx)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil // Lock is held by another
	}
	defer mu.UnlockContext(ctx)
	return true, fn()
}
