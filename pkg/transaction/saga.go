package transaction

import (
	"context"
	"fmt"
	"time"
)

// SagaStep represents a single step in a Saga transaction.
// Each step has a forward operation (Execute) and a compensating operation (Compensate).
type SagaStep struct {
	Name       string                          // Step name for logging and debugging
	Execute    func(ctx context.Context) error // Forward operation
	Compensate func(ctx context.Context) error // Compensating operation (rollback)
}

// Saga represents a distributed transaction using the Saga pattern.
// It manages a sequence of steps with automatic compensation on failure.
type Saga struct {
	steps     []SagaStep                                  // All steps to execute
	executed  []SagaStep                                  // Successfully executed steps
	onSuccess func(ctx context.Context) error            // Success callback
	onFailure func(ctx context.Context, err error) error // Failure callback
}

// NewSaga creates a new Saga transaction coordinator.
func NewSaga() *Saga {
	return &Saga{
		steps:    make([]SagaStep, 0),
		executed: make([]SagaStep, 0),
	}
}

// AddStep adds a step to the saga.
// Steps are executed in the order they are added.
func (s *Saga) AddStep(step SagaStep) *Saga {
	s.steps = append(s.steps, step)
	return s
}

// OnSuccess sets a callback to be executed when all steps succeed.
// The callback is optional and will be called after all steps complete successfully.
func (s *Saga) OnSuccess(fn func(ctx context.Context) error) *Saga {
	s.onSuccess = fn
	return s
}

// OnFailure sets a callback to be executed when any step fails.
// The callback is called after compensation is complete.
// The error parameter contains the original failure error.
func (s *Saga) OnFailure(fn func(ctx context.Context, err error) error) *Saga {
	s.onFailure = fn
	return s
}

// Execute runs the saga transaction.
// It executes steps sequentially. If any step fails, it compensates all previously executed steps in reverse order.
//
// Execution flow:
//   - Success: Execute Step1 → Execute Step2 → Execute Step3 → OnSuccess
//   - Failure (Step2 fails): Execute Step1 → Execute Step2 (fails) → Compensate Step1 → OnFailure
//
// Returns:
//   - nil if all steps succeed
//   - error if any step fails (compensation is automatically triggered)
func (s *Saga) Execute(ctx context.Context) error {
	// Forward phase: execute all steps
	for i, step := range s.steps {
		if err := s.executeStep(ctx, step, i+1, len(s.steps)); err != nil {
			// Step failed, trigger compensation
			compensateErr := s.compensate(ctx)

			// Call failure callback if set
			if s.onFailure != nil {
				if cbErr := s.onFailure(ctx, err); cbErr != nil {
					// Log callback error but don't override original error
					return fmt.Errorf("step %d failed: %w (failure callback error: %v)", i+1, err, cbErr)
				}
			}

			// Return original error with compensation info
			if compensateErr != nil {
				return fmt.Errorf("step %d failed: %w (compensation also failed: %v)", i+1, err, compensateErr)
			}
			return fmt.Errorf("step %d failed: %w (compensated successfully)", i+1, err)
		}

		// Track successfully executed step
		s.executed = append(s.executed, step)
	}

	// All steps succeeded, call success callback
	if s.onSuccess != nil {
		if err := s.onSuccess(ctx); err != nil {
			// Success callback failed, compensate
			compensateErr := s.compensate(ctx)

			if compensateErr != nil {
				return fmt.Errorf("success callback failed: %w (compensation also failed: %v)", err, compensateErr)
			}
			return fmt.Errorf("success callback failed: %w (compensated successfully)", err)
		}
	}

	return nil
}

// executeStep executes a single step with logging.
func (s *Saga) executeStep(ctx context.Context, step SagaStep, current, total int) error {
	start := time.Now()

	// Check context cancellation before execution
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled before step '%s': %w", step.Name, err)
	}

	// Execute the step
	if err := step.Execute(ctx); err != nil {
		return fmt.Errorf("step '%s' execution failed: %w", step.Name, err)
	}

	// Log execution time (can be replaced with actual logger)
	_ = time.Since(start) // Duration available for logging

	return nil
}

// compensate runs compensation operations in reverse order.
func (s *Saga) compensate(ctx context.Context) error {
	// Compensate in reverse order
	for i := len(s.executed) - 1; i >= 0; i-- {
		step := s.executed[i]

		// Skip if no compensation function defined
		if step.Compensate == nil {
			continue
		}

		// Execute compensation
		if err := step.Compensate(ctx); err != nil {
			// Compensation failed - this is a critical error
			return fmt.Errorf("compensation for step '%s' failed: %w", step.Name, err)
		}
	}

	return nil
}

// GetExecutedSteps returns the list of successfully executed steps.
// Useful for debugging and monitoring.
func (s *Saga) GetExecutedSteps() []string {
	names := make([]string, len(s.executed))
	for i, step := range s.executed {
		names[i] = step.Name
	}
	return names
}

// StepCount returns the total number of steps in the saga.
func (s *Saga) StepCount() int {
	return len(s.steps)
}

// ExecutedCount returns the number of successfully executed steps.
func (s *Saga) ExecutedCount() int {
	return len(s.executed)
}

// AddRetryableStep 添加可重试的步骤
func (s *Saga) AddRetryableStep(step SagaStep, config RetryConfig) *Saga {
	retryable := NewRetryableSagaStep(step, config)
	// 包装 Execute 函数
	wrappedStep := SagaStep{
		Name:       step.Name,
		Execute:    retryable.ExecuteWithRetry,
		Compensate: step.Compensate,
	}
	s.steps = append(s.steps, wrappedStep)
	return s
}

// WithDefaultRetry 使用默认重试配置添加步骤
func (s *Saga) WithDefaultRetry(step SagaStep) *Saga {
	return s.AddRetryableStep(step, DefaultRetryConfig())
}

// AddRetryableStepWithErrors 添加可重试的步骤，指定可重试的错误类型
func (s *Saga) AddRetryableStepWithErrors(step SagaStep, config RetryConfig, retryableErrs ...error) *Saga {
	retryable := NewRetryableSagaStep(step, config).WithRetryableErrors(retryableErrs...)
	wrappedStep := SagaStep{
		Name:       step.Name,
		Execute:    retryable.ExecuteWithRetry,
		Compensate: step.Compensate,
	}
	s.steps = append(s.steps, wrappedStep)
	return s
}
