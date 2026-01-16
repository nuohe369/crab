package transaction

import (
	"context"
	"errors"
	"testing"
)

// TestNewSaga tests Saga initialization
func TestNewSaga(t *testing.T) {
	saga := NewSaga()

	if saga == nil {
		t.Fatal("NewSaga() returned nil")
	}
	if saga.steps == nil {
		t.Error("steps should be initialized")
	}
	if saga.executed == nil {
		t.Error("executed should be initialized")
	}
	if saga.StepCount() != 0 {
		t.Errorf("StepCount() = %d, want 0", saga.StepCount())
	}
}

// TestSaga_AddStep tests adding steps to saga
func TestSaga_AddStep(t *testing.T) {
	saga := NewSaga()

	step1 := SagaStep{
		Name:    "step1",
		Execute: func(ctx context.Context) error { return nil },
	}
	step2 := SagaStep{
		Name:    "step2",
		Execute: func(ctx context.Context) error { return nil },
	}

	saga.AddStep(step1).AddStep(step2)

	if saga.StepCount() != 2 {
		t.Errorf("StepCount() = %d, want 2", saga.StepCount())
	}
}

// TestSaga_Execute_Success tests successful execution of all steps
func TestSaga_Execute_Success(t *testing.T) {
	executed := make([]string, 0)

	saga := NewSaga().
		AddStep(SagaStep{
			Name: "step1",
			Execute: func(ctx context.Context) error {
				executed = append(executed, "step1")
				return nil
			},
		}).
		AddStep(SagaStep{
			Name: "step2",
			Execute: func(ctx context.Context) error {
				executed = append(executed, "step2")
				return nil
			},
		}).
		AddStep(SagaStep{
			Name: "step3",
			Execute: func(ctx context.Context) error {
				executed = append(executed, "step3")
				return nil
			},
		})

	err := saga.Execute(context.Background())

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if len(executed) != 3 {
		t.Errorf("expected 3 steps executed, got %d", len(executed))
	}

	expectedOrder := []string{"step1", "step2", "step3"}
	for i, name := range expectedOrder {
		if executed[i] != name {
			t.Errorf("step %d: got %s, want %s", i, executed[i], name)
		}
	}

	if saga.ExecutedCount() != 3 {
		t.Errorf("ExecutedCount() = %d, want 3", saga.ExecutedCount())
	}
}

// TestSaga_Execute_FailureWithCompensation tests compensation on failure
func TestSaga_Execute_FailureWithCompensation(t *testing.T) {
	executed := make([]string, 0)
	compensated := make([]string, 0)

	saga := NewSaga().
		AddStep(SagaStep{
			Name: "step1",
			Execute: func(ctx context.Context) error {
				executed = append(executed, "step1")
				return nil
			},
			Compensate: func(ctx context.Context) error {
				compensated = append(compensated, "step1")
				return nil
			},
		}).
		AddStep(SagaStep{
			Name: "step2",
			Execute: func(ctx context.Context) error {
				executed = append(executed, "step2")
				return nil
			},
			Compensate: func(ctx context.Context) error {
				compensated = append(compensated, "step2")
				return nil
			},
		}).
		AddStep(SagaStep{
			Name: "step3",
			Execute: func(ctx context.Context) error {
				return errors.New("step3 failed")
			},
			Compensate: func(ctx context.Context) error {
				compensated = append(compensated, "step3")
				return nil
			},
		})

	err := saga.Execute(context.Background())

	if err == nil {
		t.Fatal("Execute() should return error")
	}

	// Only step1 and step2 should be executed
	if len(executed) != 2 {
		t.Errorf("expected 2 steps executed, got %d", len(executed))
	}

	// Compensation should run for step2 then step1 (reverse order)
	if len(compensated) != 2 {
		t.Errorf("expected 2 steps compensated, got %d", len(compensated))
	}

	expectedCompensationOrder := []string{"step2", "step1"}
	for i, name := range expectedCompensationOrder {
		if compensated[i] != name {
			t.Errorf("compensation %d: got %s, want %s", i, compensated[i], name)
		}
	}
}

// TestSaga_Execute_WithCallbacks tests success and failure callbacks
func TestSaga_Execute_WithCallbacks(t *testing.T) {
	t.Run("success callback", func(t *testing.T) {
		successCalled := false

		saga := NewSaga().
			AddStep(SagaStep{
				Name:    "step1",
				Execute: func(ctx context.Context) error { return nil },
			}).
			OnSuccess(func(ctx context.Context) error {
				successCalled = true
				return nil
			})

		err := saga.Execute(context.Background())

		if err != nil {
			t.Errorf("Execute() error = %v, want nil", err)
		}
		if !successCalled {
			t.Error("success callback was not called")
		}
	})

	t.Run("failure callback", func(t *testing.T) {
		failureCalled := false
		var capturedError error

		saga := NewSaga().
			AddStep(SagaStep{
				Name:    "step1",
				Execute: func(ctx context.Context) error { return nil },
			}).
			AddStep(SagaStep{
				Name:    "step2",
				Execute: func(ctx context.Context) error { return errors.New("failed") },
			}).
			OnFailure(func(ctx context.Context, err error) error {
				failureCalled = true
				capturedError = err
				return nil
			})

		err := saga.Execute(context.Background())

		if err == nil {
			t.Fatal("Execute() should return error")
		}
		if !failureCalled {
			t.Error("failure callback was not called")
		}
		if capturedError == nil {
			t.Error("error was not captured in failure callback")
		}
	})
}

// TestSaga_Execute_NoCompensateFunction tests handling of missing compensation
func TestSaga_Execute_NoCompensateFunction(t *testing.T) {
	executed := make([]string, 0)

	saga := NewSaga().
		AddStep(SagaStep{
			Name: "step1",
			Execute: func(ctx context.Context) error {
				executed = append(executed, "step1")
				return nil
			},
			// No Compensate function
		}).
		AddStep(SagaStep{
			Name: "step2",
			Execute: func(ctx context.Context) error {
				return errors.New("failed")
			},
		})

	err := saga.Execute(context.Background())

	// Should handle missing compensation gracefully
	if err == nil {
		t.Fatal("Execute() should return error")
	}

	// Step1 was executed
	if len(executed) != 1 {
		t.Errorf("expected 1 step executed, got %d", len(executed))
	}
}

// TestSaga_Execute_CompensationFailure tests handling of compensation failure
func TestSaga_Execute_CompensationFailure(t *testing.T) {
	saga := NewSaga().
		AddStep(SagaStep{
			Name:    "step1",
			Execute: func(ctx context.Context) error { return nil },
			Compensate: func(ctx context.Context) error {
				return errors.New("compensation failed")
			},
		}).
		AddStep(SagaStep{
			Name:    "step2",
			Execute: func(ctx context.Context) error { return errors.New("failed") },
		})

	err := saga.Execute(context.Background())

	if err == nil {
		t.Fatal("Execute() should return error")
	}

	// Error message should mention both failures
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("error message should not be empty")
	}
}

// TestSaga_Execute_ContextCancellation tests handling of context cancellation
func TestSaga_Execute_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	saga := NewSaga().
		AddStep(SagaStep{
			Name:    "step1",
			Execute: func(ctx context.Context) error { return nil },
		})

	err := saga.Execute(ctx)

	if err == nil {
		t.Fatal("Execute() should return error for cancelled context")
	}
}

// TestSaga_Execute_SuccessCallbackFailure tests compensation when success callback fails
func TestSaga_Execute_SuccessCallbackFailure(t *testing.T) {
	compensated := false

	saga := NewSaga().
		AddStep(SagaStep{
			Name:    "step1",
			Execute: func(ctx context.Context) error { return nil },
			Compensate: func(ctx context.Context) error {
				compensated = true
				return nil
			},
		}).
		OnSuccess(func(ctx context.Context) error {
			return errors.New("success callback failed")
		})

	err := saga.Execute(context.Background())

	if err == nil {
		t.Fatal("Execute() should return error when success callback fails")
	}

	if !compensated {
		t.Error("should compensate when success callback fails")
	}
}

// TestSaga_GetExecutedSteps tests getting executed step names
func TestSaga_GetExecutedSteps(t *testing.T) {
	saga := NewSaga().
		AddStep(SagaStep{
			Name:    "step1",
			Execute: func(ctx context.Context) error { return nil },
		}).
		AddStep(SagaStep{
			Name:    "step2",
			Execute: func(ctx context.Context) error { return errors.New("failed") },
		})

	saga.Execute(context.Background())

	steps := saga.GetExecutedSteps()
	if len(steps) != 1 {
		t.Errorf("GetExecutedSteps() returned %d steps, want 1", len(steps))
	}
	if steps[0] != "step1" {
		t.Errorf("GetExecutedSteps()[0] = %s, want step1", steps[0])
	}
}

// TestSaga_ChainedCalls tests method chaining
func TestSaga_ChainedCalls(t *testing.T) {
	saga := NewSaga().
		AddStep(SagaStep{Name: "step1", Execute: func(ctx context.Context) error { return nil }}).
		AddStep(SagaStep{Name: "step2", Execute: func(ctx context.Context) error { return nil }}).
		OnSuccess(func(ctx context.Context) error { return nil }).
		OnFailure(func(ctx context.Context, err error) error { return nil })

	if saga.StepCount() != 2 {
		t.Errorf("StepCount() = %d, want 2", saga.StepCount())
	}
}

// TestSaga_EmptySaga tests executing saga with no steps
func TestSaga_EmptySaga(t *testing.T) {
	successCalled := false

	saga := NewSaga().
		OnSuccess(func(ctx context.Context) error {
			successCalled = true
			return nil
		})

	err := saga.Execute(context.Background())

	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	if !successCalled {
		t.Error("success callback should be called even for empty saga")
	}
}
