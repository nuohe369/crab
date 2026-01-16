package transaction

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryableSagaStep_Success(t *testing.T) {
	attempts := 0
	step := NewRetryableSagaStep(SagaStep{
		Name: "test_step",
		Execute: func(ctx context.Context) error {
			attempts++
			return nil
		},
	}, DefaultRetryConfig())

	err := step.ExecuteWithRetry(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestRetryableSagaStep_RetryOnFailure(t *testing.T) {
	attempts := 0
	testErr := errors.New("temporary error")

	step := NewRetryableSagaStep(SagaStep{
		Name: "test_step",
		Execute: func(ctx context.Context) error {
			attempts++
			if attempts < 3 {
				return testErr
			}
			return nil
		},
	}, RetryConfig{
		MaxAttempts:     5,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     100 * time.Millisecond,
		Multiplier:      2.0,
		Jitter:          0,
	})

	err := step.ExecuteWithRetry(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetryableSagaStep_MaxAttemptsExceeded(t *testing.T) {
	attempts := 0
	testErr := errors.New("persistent error")

	step := NewRetryableSagaStep(SagaStep{
		Name: "test_step",
		Execute: func(ctx context.Context) error {
			attempts++
			return testErr
		},
	}, RetryConfig{
		MaxAttempts:     3,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     100 * time.Millisecond,
		Multiplier:      2.0,
		Jitter:          0,
	})

	err := step.ExecuteWithRetry(context.Background())
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetryableSagaStep_NonRetryableError(t *testing.T) {
	attempts := 0
	retryableErr := errors.New("retryable error")
	nonRetryableErr := errors.New("non-retryable error")

	step := NewRetryableSagaStep(SagaStep{
		Name: "test_step",
		Execute: func(ctx context.Context) error {
			attempts++
			return nonRetryableErr
		},
	}, RetryConfig{
		MaxAttempts:     5,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     100 * time.Millisecond,
		Multiplier:      2.0,
		Jitter:          0,
	}).WithRetryableErrors(retryableErr)

	err := step.ExecuteWithRetry(context.Background())
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt (non-retryable), got %d", attempts)
	}
}

func TestRetryableSagaStep_ContextCancellation(t *testing.T) {
	attempts := 0
	testErr := errors.New("error")

	ctx, cancel := context.WithCancel(context.Background())

	step := NewRetryableSagaStep(SagaStep{
		Name: "test_step",
		Execute: func(ctx context.Context) error {
			attempts++
			if attempts == 2 {
				cancel() // 第二次尝试后取消
			}
			return testErr
		},
	}, RetryConfig{
		MaxAttempts:     5,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     100 * time.Millisecond,
		Multiplier:      2.0,
		Jitter:          0,
	})

	err := step.ExecuteWithRetry(ctx)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

func TestSaga_WithDefaultRetry(t *testing.T) {
	attempts := 0

	saga := NewSaga().
		WithDefaultRetry(SagaStep{
			Name: "retry_step",
			Execute: func(ctx context.Context) error {
				attempts++
				if attempts < 2 {
					return errors.New("temporary error")
				}
				return nil
			},
			Compensate: func(ctx context.Context) error {
				return nil
			},
		})

	err := saga.Execute(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestSaga_AddRetryableStep(t *testing.T) {
	attempts := 0

	saga := NewSaga().
		AddRetryableStep(SagaStep{
			Name: "custom_retry_step",
			Execute: func(ctx context.Context) error {
				attempts++
				if attempts < 4 {
					return errors.New("temporary error")
				}
				return nil
			},
		}, RetryConfig{
			MaxAttempts:     5,
			InitialInterval: 5 * time.Millisecond,
			MaxInterval:     50 * time.Millisecond,
			Multiplier:      1.5,
			Jitter:          0,
		})

	err := saga.Execute(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if attempts != 4 {
		t.Errorf("Expected 4 attempts, got %d", attempts)
	}
}

func TestRetry_SimpleFunction(t *testing.T) {
	attempts := 0

	err := Retry(context.Background(), func(ctx context.Context) error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary error")
		}
		return nil
	}, RetryConfig{
		MaxAttempts:     3,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     100 * time.Millisecond,
		Multiplier:      2.0,
		Jitter:          0,
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryWithDefault(t *testing.T) {
	attempts := 0

	err := RetryWithDefault(context.Background(), func(ctx context.Context) error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary error")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}
