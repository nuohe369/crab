// Package transaction provides transaction management utilities for database operations
// Package transaction 提供数据库操作的事务管理工具
package transaction

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

// RetryConfig represents retry configuration
// RetryConfig 表示重试配置
type RetryConfig struct {
	MaxAttempts     int           // Max retry attempts (default 3) | 最大重试次数（默认 3）
	InitialInterval time.Duration // Initial interval (default 100ms) | 初始间隔（默认 100ms）
	MaxInterval     time.Duration // Max interval (default 10s) | 最大间隔（默认 10s）
	Multiplier      float64       // Backoff multiplier (default 2.0) | 退避乘数（默认 2.0）
	Jitter          float64       // Jitter factor 0-1 (default 0.1) | 抖动因子 0-1（默认 0.1）
}

// DefaultRetryConfig returns default retry configuration
// DefaultRetryConfig 返回默认重试配置
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:     3,
		InitialInterval: 100 * time.Millisecond,
		MaxInterval:     10 * time.Second,
		Multiplier:      2.0,
		Jitter:          0.1,
	}
}

// RetryableSagaStep represents a retryable Saga step
// RetryableSagaStep 表示可重试的 Saga 步骤
type RetryableSagaStep struct {
	SagaStep
	RetryConfig   RetryConfig // Retry configuration | 重试配置
	RetryableErrs []error     // Retryable error types, nil means all errors are retryable | 可重试的错误类型，nil 表示所有错误都可重试
}

// NewRetryableSagaStep creates a retryable step
// NewRetryableSagaStep 创建可重试的步骤
func NewRetryableSagaStep(step SagaStep, config RetryConfig) *RetryableSagaStep {
	return &RetryableSagaStep{
		SagaStep:    step,
		RetryConfig: config,
	}
}

// WithRetryableErrors sets retryable error types
// WithRetryableErrors 设置可重试的错误类型
func (s *RetryableSagaStep) WithRetryableErrors(errs ...error) *RetryableSagaStep {
	s.RetryableErrs = errs
	return s
}

// ExecuteWithRetry executes the step with retry
// ExecuteWithRetry 执行带重试的步骤
func (s *RetryableSagaStep) ExecuteWithRetry(ctx context.Context) error {
	var lastErr error
	interval := s.RetryConfig.InitialInterval

	for attempt := 1; attempt <= s.RetryConfig.MaxAttempts; attempt++ {
		// Check if context is cancelled | 检查上下文是否已取消
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = s.Execute(ctx)
		if lastErr == nil {
			return nil
		}

		// Check if error is retryable | 检查错误是否可重试
		if !s.isRetryable(lastErr) {
			return lastErr
		}

		// Don't wait on last attempt | 最后一次尝试不等待
		if attempt == s.RetryConfig.MaxAttempts {
			break
		}

		// Calculate next wait time (exponential backoff + jitter) | 计算下次等待时间（指数退避 + 抖动）
		waitTime := s.calculateBackoff(interval)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}

		// Update interval | 更新间隔
		interval = time.Duration(float64(interval) * s.RetryConfig.Multiplier)
		if interval > s.RetryConfig.MaxInterval {
			interval = s.RetryConfig.MaxInterval
		}
	}

	return lastErr
}

// calculateBackoff calculates backoff time with jitter
// calculateBackoff 计算带抖动的退避时间
func (s *RetryableSagaStep) calculateBackoff(base time.Duration) time.Duration {
	if s.RetryConfig.Jitter <= 0 {
		return base
	}
	jitter := (rand.Float64()*2 - 1) * s.RetryConfig.Jitter * float64(base)
	return base + time.Duration(jitter)
}

// isRetryable checks if error is retryable
// isRetryable 检查错误是否可重试
func (s *RetryableSagaStep) isRetryable(err error) bool {
	if s.RetryableErrs == nil || len(s.RetryableErrs) == 0 {
		return true // All errors are retryable | 所有错误都可重试
	}
	for _, retryableErr := range s.RetryableErrs {
		if errors.Is(err, retryableErr) {
			return true
		}
	}
	return false
}

// RetryableFunc is a retryable function wrapper
// RetryableFunc 是可重试的函数包装器
type RetryableFunc struct {
	fn     func(ctx context.Context) error // Function to execute | 要执行的函数
	config RetryConfig                     // Retry configuration | 重试配置
	errs   []error                         // Retryable error types | 可重试的错误类型
}

// NewRetryableFunc creates a retryable function
// NewRetryableFunc 创建可重试的函数
func NewRetryableFunc(fn func(ctx context.Context) error, config RetryConfig) *RetryableFunc {
	return &RetryableFunc{
		fn:     fn,
		config: config,
	}
}

// WithRetryableErrors sets retryable error types
// WithRetryableErrors 设置可重试的错误类型
func (r *RetryableFunc) WithRetryableErrors(errs ...error) *RetryableFunc {
	r.errs = errs
	return r
}

// Execute executes the function with retry
// Execute 执行带重试的函数
func (r *RetryableFunc) Execute(ctx context.Context) error {
	step := &RetryableSagaStep{
		SagaStep: SagaStep{
			Name:    "retryable_func",
			Execute: r.fn,
		},
		RetryConfig:   r.config,
		RetryableErrs: r.errs,
	}
	return step.ExecuteWithRetry(ctx)
}

// Retry is a simple retry function
// Retry 是简单的重试函数
func Retry(ctx context.Context, fn func(ctx context.Context) error, config RetryConfig) error {
	return NewRetryableFunc(fn, config).Execute(ctx)
}

// RetryWithDefault retries with default configuration
// RetryWithDefault 使用默认配置重试
func RetryWithDefault(ctx context.Context, fn func(ctx context.Context) error) error {
	return Retry(ctx, fn, DefaultRetryConfig())
}
