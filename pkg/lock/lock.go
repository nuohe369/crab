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
// Config 表示分布式锁配置
type Config struct {
	Expiry        time.Duration `toml:"expiry"`         // Lock expiry time (default 8s) | 锁过期时间（默认 8 秒）
	Tries         int           `toml:"tries"`          // Max retry attempts (default 32) | 最大重试次数（默认 32）
	RetryDelay    time.Duration `toml:"retry_delay"`    // Retry interval (default 100ms) | 重试间隔（默认 100 毫秒）
	DriftFactor   float64       `toml:"drift_factor"`   // Clock drift factor (default 0.01) | 时钟漂移因子（默认 0.01）
	TimeoutFactor float64       `toml:"timeout_factor"` // Timeout factor (default 0.05) | 超时因子（默认 0.05）
}

// DefaultConfig returns default configuration
// DefaultConfig 返回默认配置
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
// Locker 是分布式锁管理器
type Locker struct {
	rs     *redsync.Redsync
	config Config
}

var defaultLocker *Locker

// Init initializes distributed lock
// Init 初始化分布式锁
func Init(client redis.UniversalClient, cfg Config) {
	pool := goredis.NewPool(client)
	defaultLocker = &Locker{
		rs:     redsync.New(pool),
		config: cfg,
	}
	log.Info("Distributed lock initialized: expiry=%v, tries=%d", cfg.Expiry, cfg.Tries)
}

// Get returns the default Locker
// Get 返回默认 Locker
func Get() *Locker {
	return defaultLocker
}

// Mutex represents a distributed mutex
// Mutex 表示分布式互斥锁
type Mutex struct {
	mu     *redsync.Mutex
	name   string
	locked bool
}

// NewMutex creates a mutex
// NewMutex 创建互斥锁
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
// Lock 获取锁
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
// LockContext 使用 context 获取锁
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
// TryLock 尝试获取锁（非阻塞）
func (m *Mutex) TryLock() (bool, error) {
	err := m.mu.TryLock()
	if err != nil {
		if err == redsync.ErrFailed {
			return false, nil // Lock is held by another | 锁被其他持有
		}
		return false, err
	}
	m.locked = true
	log.Debug("Acquired lock [%s]", m.name)
	return true, nil
}

// TryLockContext tries to acquire the lock with context
// TryLockContext 使用 context 尝试获取锁
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
// Unlock 释放锁
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
// UnlockContext 使用 context 释放锁
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
// Extend 延长锁过期时间
func (m *Mutex) Extend() (bool, error) {
	return m.mu.Extend()
}

// ExtendContext extends the lock expiry time with context
// ExtendContext 使用 context 延长锁过期时间
func (m *Mutex) ExtendContext(ctx context.Context) (bool, error) {
	return m.mu.ExtendContext(ctx)
}

// Name returns the lock name
// Name 返回锁名称
func (m *Mutex) Name() string {
	return m.name
}

// IsLocked checks if the lock is held
// IsLocked 检查锁是否被持有
func (m *Mutex) IsLocked() bool {
	return m.locked
}

// Until returns the lock expiry time
// Until 返回锁过期时间
func (m *Mutex) Until() time.Time {
	return m.mu.Until()
}

// Option is a configuration option
// Option 是配置选项
type Option func(*Config)

// WithExpiry sets the expiry time
// WithExpiry 设置过期时间
func WithExpiry(d time.Duration) Option {
	return func(c *Config) {
		c.Expiry = d
	}
}

// WithTries sets the retry attempts
// WithTries 设置重试次数
func WithTries(n int) Option {
	return func(c *Config) {
		c.Tries = n
	}
}

// WithRetryDelay sets the retry interval
// WithRetryDelay 设置重试间隔
func WithRetryDelay(d time.Duration) Option {
	return func(c *Config) {
		c.RetryDelay = d
	}
}

// NewMutex creates a mutex using default Locker
// NewMutex 使用默认 Locker 创建互斥锁
func NewMutex(name string, opts ...Option) *Mutex {
	if defaultLocker == nil {
		panic("lock: not initialized, call lock.Init() first")
	}
	return defaultLocker.NewMutex(name, opts...)
}

// WithLock executes function under lock protection
// WithLock 在锁保护下执行函数
func WithLock(ctx context.Context, name string, fn func() error, opts ...Option) error {
	mu := NewMutex(name, opts...)
	if err := mu.LockContext(ctx); err != nil {
		return err
	}
	defer mu.UnlockContext(ctx)
	return fn()
}

// TryWithLock tries to execute function under lock protection (non-blocking)
// TryWithLock 尝试在锁保护下执行函数（非阻塞）
func TryWithLock(ctx context.Context, name string, fn func() error, opts ...Option) (bool, error) {
	mu := NewMutex(name, opts...)
	ok, err := mu.TryLockContext(ctx)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil // Lock is held by another | 锁被其他持有
	}
	defer mu.UnlockContext(ctx)
	return true, fn()
}
