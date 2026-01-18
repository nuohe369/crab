// Package transaction provides transaction management utilities for database operations
// Package transaction 提供数据库操作的事务管理工具
package transaction

import (
	"context"
	"fmt"
	"time"
)

// SagaStep represents a single step in a Saga transaction.
// Each step has a forward operation (Execute) and a compensating operation (Compensate).
// SagaStep 表示 Saga 事务中的单个步骤
// 每个步骤都有一个正向操作（Execute）和一个补偿操作（Compensate）
type SagaStep struct {
	Name       string                          // Step name for logging and debugging | 步骤名称，用于日志和调试
	Execute    func(ctx context.Context) error // Forward operation | 正向操作
	Compensate func(ctx context.Context) error // Compensating operation (rollback) | 补偿操作（回滚）
}

// Saga represents a distributed transaction using the Saga pattern.
// It manages a sequence of steps with automatic compensation on failure.
// Saga 表示使用 Saga 模式的分布式事务
// 它管理一系列步骤，失败时自动补偿
type Saga struct {
	steps     []SagaStep                                 // All steps to execute | 要执行的所有步骤
	executed  []SagaStep                                 // Successfully executed steps | 已成功执行的步骤
	onSuccess func(ctx context.Context) error            // Success callback | 成功回调
	onFailure func(ctx context.Context, err error) error // Failure callback | 失败回调
}

// NewSaga creates a new Saga transaction coordinator.
// NewSaga 创建新的 Saga 事务协调器
func NewSaga() *Saga {
	return &Saga{
		steps:    make([]SagaStep, 0),
		executed: make([]SagaStep, 0),
	}
}

// AddStep adds a step to the saga.
// Steps are executed in the order they are added.
// AddStep 向 saga 添加步骤
// 步骤按添加顺序执行
func (s *Saga) AddStep(step SagaStep) *Saga {
	s.steps = append(s.steps, step)
	return s
}

// OnSuccess sets a callback to be executed when all steps succeed.
// The callback is optional and will be called after all steps complete successfully.
// OnSuccess 设置所有步骤成功时执行的回调
// 回调是可选的，将在所有步骤成功完成后调用
func (s *Saga) OnSuccess(fn func(ctx context.Context) error) *Saga {
	s.onSuccess = fn
	return s
}

// OnFailure sets a callback to be executed when any step fails.
// The callback is called after compensation is complete.
// The error parameter contains the original failure error.
// OnFailure 设置任何步骤失败时执行的回调
// 回调在补偿完成后调用
// error 参数包含原始失败错误
func (s *Saga) OnFailure(fn func(ctx context.Context, err error) error) *Saga {
	s.onFailure = fn
	return s
}

// Execute runs the saga transaction.
// It executes steps sequentially. If any step fails, it compensates all previously executed steps in reverse order.
// Execute 运行 saga 事务
// 它按顺序执行步骤。如果任何步骤失败，它会按相反顺序补偿所有先前执行的步骤
//
// Execution flow:
//   - Success: Execute Step1 → Execute Step2 → Execute Step3 → OnSuccess
//   - Failure (Step2 fails): Execute Step1 → Execute Step2 (fails) → Compensate Step1 → OnFailure
//
// Returns:
//   - nil if all steps succeed
//   - error if any step fails (compensation is automatically triggered)
func (s *Saga) Execute(ctx context.Context) error {
	// Forward phase: execute all steps | 正向阶段：执行所有步骤
	for i, step := range s.steps {
		if err := s.executeStep(ctx, step, i+1, len(s.steps)); err != nil {
			// Step failed, trigger compensation | 步骤失败，触发补偿
			compensateErr := s.compensate(ctx)

			// Call failure callback if set | 如果设置了失败回调，则调用
			if s.onFailure != nil {
				if cbErr := s.onFailure(ctx, err); cbErr != nil {
					// Log callback error but don't override original error | 记录回调错误但不覆盖原始错误
					return fmt.Errorf("step %d failed: %w (failure callback error: %v)", i+1, err, cbErr)
				}
			}

			// Return original error with compensation info | 返回带补偿信息的原始错误
			if compensateErr != nil {
				return fmt.Errorf("step %d failed: %w (compensation also failed: %v)", i+1, err, compensateErr)
			}
			return fmt.Errorf("step %d failed: %w (compensated successfully)", i+1, err)
		}

		// Track successfully executed step | 跟踪成功执行的步骤
		s.executed = append(s.executed, step)
	}

	// All steps succeeded, call success callback | 所有步骤成功，调用成功回调
	if s.onSuccess != nil {
		if err := s.onSuccess(ctx); err != nil {
			// Success callback failed, compensate | 成功回调失败，补偿
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
// executeStep 执行单个步骤并记录日志
func (s *Saga) executeStep(ctx context.Context, step SagaStep, current, total int) error {
	start := time.Now()

	// Check context cancellation before execution | 执行前检查上下文取消
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled before step '%s': %w", step.Name, err)
	}

	// Execute the step | 执行步骤
	if err := step.Execute(ctx); err != nil {
		return fmt.Errorf("step '%s' execution failed: %w", step.Name, err)
	}

	// Log execution time (can be replaced with actual logger) | 记录执行时间（可替换为实际日志器）
	_ = time.Since(start) // Duration available for logging

	return nil
}

// compensate runs compensation operations in reverse order.
// compensate 按相反顺序运行补偿操作
func (s *Saga) compensate(ctx context.Context) error {
	// Compensate in reverse order | 按相反顺序补偿
	for i := len(s.executed) - 1; i >= 0; i-- {
		step := s.executed[i]

		// Skip if no compensation function defined | 如果未定义补偿函数则跳过
		if step.Compensate == nil {
			continue
		}

		// Execute compensation | 执行补偿
		if err := step.Compensate(ctx); err != nil {
			// Compensation failed - this is a critical error | 补偿失败 - 这是一个严重错误
			return fmt.Errorf("compensation for step '%s' failed: %w", step.Name, err)
		}
	}

	return nil
}

// GetExecutedSteps returns the list of successfully executed steps.
// Useful for debugging and monitoring.
// GetExecutedSteps 返回成功执行的步骤列表
// 用于调试和监控
func (s *Saga) GetExecutedSteps() []string {
	names := make([]string, len(s.executed))
	for i, step := range s.executed {
		names[i] = step.Name
	}
	return names
}

// StepCount returns the total number of steps in the saga.
// StepCount 返回 saga 中的步骤总数
func (s *Saga) StepCount() int {
	return len(s.steps)
}

// ExecutedCount returns the number of successfully executed steps.
// ExecutedCount 返回成功执行的步骤数
func (s *Saga) ExecutedCount() int {
	return len(s.executed)
}

// AddRetryableStep adds a retryable step
// AddRetryableStep 添加可重试的步骤
func (s *Saga) AddRetryableStep(step SagaStep, config RetryConfig) *Saga {
	retryable := NewRetryableSagaStep(step, config)
	// Wrap Execute function | 包装 Execute 函数
	wrappedStep := SagaStep{
		Name:       step.Name,
		Execute:    retryable.ExecuteWithRetry,
		Compensate: step.Compensate,
	}
	s.steps = append(s.steps, wrappedStep)
	return s
}

// WithDefaultRetry adds a step with default retry configuration
// WithDefaultRetry 使用默认重试配置添加步骤
func (s *Saga) WithDefaultRetry(step SagaStep) *Saga {
	return s.AddRetryableStep(step, DefaultRetryConfig())
}

// AddRetryableStepWithErrors adds a retryable step with specified retryable error types
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
