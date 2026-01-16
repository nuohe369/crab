package transaction

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

// RetryConfig represents retry configuration
type RetryConfig struct {
	MaxAttempts     int           // Max retry attempts (default 3)
	InitialInterval time.Duration // Initial interval (default 100ms)
	MaxInterval     time.Duration // Max interval (default 10s)
	Multiplier      float64       // Backoff multiplier (default 2.0)
	Jitter          float64       // Jitter factor 0-1 (default 0.1)
}

// DefaultRetryConfig returns default retry configuration
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
type RetryableSagaStep struct {
	SagaStep
	RetryConfig   RetryConfig
	RetryableErrs []error // Retryable error types, nil means all errors are retryable
}

// NewRetryableSagaStep creates a retryable step
func NewRetryableSagaStep(step SagaStep, config RetryConfig) *RetryableSagaStep {
	return &RetryableSagaStep{
		SagaStep:    step,
		RetryConfig: config,
	}
}

// WithRetryableErrors sets retryable error types
func (s *RetryableSagaStep) WithRetryableErrors(errs ...error) *RetryableSagaStep {
	s.RetryableErrs = errs
	return s
}

// ExecuteWithRetry executes the step with retry
func (s *RetryableSagaStep) ExecuteWithRetry(ctx context.Context) error {
	var lastErr error
	interval := s.RetryConfig.InitialInterval

	for attempt := 1; attempt <= s.RetryConfig.MaxAttempts; attempt++ {
		// Check if context is cancelled
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = s.Execute(ctx)
		if lastErr == nil {
			return nil
		}

		// Check if error is retryable
		if !s.isRetryable(lastErr) {
			return lastErr
		}

		// Don't wait on last attempt
		if attempt == s.RetryConfig.MaxAttempts {
			break
		}

		// Calculate next wait time (exponential backoff + jitter)
		waitTime := s.calculateBackoff(interval)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}

		// Update interval
		interval = time.Duration(float64(interval) * s.RetryConfig.Multiplier)
		if interval > s.RetryConfig.MaxInterval {
			interval = s.RetryConfig.MaxInterval
		}
	}

	return lastErr
}

func (s *RetryableSagaStep) calculateBackoff(base time.Duration) time.Duration {
	if s.RetryConfig.Jitter <= 0 {
		return base
	}
	jitter := (rand.Float64()*2 - 1) * s.RetryConfig.Jitter * float64(base)
	return base + time.Duration(jitter)
}

func (s *RetryableSagaStep) isRetryable(err error) bool {
	if s.RetryableErrs == nil || len(s.RetryableErrs) == 0 {
		return true // All errors are retryable
	}
	for _, retryableErr := range s.RetryableErrs {
		if errors.Is(err, retryableErr) {
			return true
		}
	}
	return false
}

// RetryableFunc is a retryable function wrapper
type RetryableFunc struct {
	fn     func(ctx context.Context) error
	config RetryConfig
	errs   []error
}

// NewRetryableFunc creates a retryable function
func NewRetryableFunc(fn func(ctx context.Context) error, config RetryConfig) *RetryableFunc {
	return &RetryableFunc{
		fn:     fn,
		config: config,
	}
}

// WithRetryableErrors sets retryable error types
func (r *RetryableFunc) WithRetryableErrors(errs ...error) *RetryableFunc {
	r.errs = errs
	return r
}

// Execute executes the function with retry
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
func Retry(ctx context.Context, fn func(ctx context.Context) error, config RetryConfig) error {
	return NewRetryableFunc(fn, config).Execute(ctx)
}

// RetryWithDefault retries with default configuration
func RetryWithDefault(ctx context.Context, fn func(ctx context.Context) error) error {
	return Retry(ctx, fn, DefaultRetryConfig())
}
