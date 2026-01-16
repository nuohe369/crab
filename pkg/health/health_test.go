package health

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestHealth_Check_Empty(t *testing.T) {
	h := &Health{
		checkers: make([]Checker, 0),
		timeout:  5 * time.Second,
	}

	result := h.Check(context.Background())
	if result.Status != StatusUp {
		t.Errorf("Expected status UP, got %s", result.Status)
	}
}

func TestHealth_Check_AllUp(t *testing.T) {
	h := &Health{
		checkers: make([]Checker, 0),
		timeout:  5 * time.Second,
	}

	h.RegisterFunc("test1", func(ctx context.Context) CheckResult {
		return CheckResult{Status: StatusUp}
	})
	h.RegisterFunc("test2", func(ctx context.Context) CheckResult {
		return CheckResult{Status: StatusUp}
	})

	result := h.Check(context.Background())
	if result.Status != StatusUp {
		t.Errorf("Expected status UP, got %s", result.Status)
	}
	if len(result.Checks) != 2 {
		t.Errorf("Expected 2 checks, got %d", len(result.Checks))
	}
}

func TestHealth_Check_OneDown(t *testing.T) {
	h := &Health{
		checkers: make([]Checker, 0),
		timeout:  5 * time.Second,
	}

	h.RegisterFunc("healthy", func(ctx context.Context) CheckResult {
		return CheckResult{Status: StatusUp}
	})
	h.RegisterFunc("unhealthy", func(ctx context.Context) CheckResult {
		return CheckResult{Status: StatusDown, Message: "service unavailable"}
	})

	result := h.Check(context.Background())
	if result.Status != StatusDown {
		t.Errorf("Expected status DOWN, got %s", result.Status)
	}
}

func TestHealth_Liveness(t *testing.T) {
	h := &Health{
		checkers: make([]Checker, 0),
		timeout:  5 * time.Second,
	}

	// 即使有不健康的检查器，liveness 也应该返回 UP
	h.RegisterFunc("unhealthy", func(ctx context.Context) CheckResult {
		return CheckResult{Status: StatusDown}
	})

	result := h.Liveness(context.Background())
	if result.Status != StatusUp {
		t.Errorf("Expected liveness status UP, got %s", result.Status)
	}
}

func TestHealth_Readiness(t *testing.T) {
	h := &Health{
		checkers: make([]Checker, 0),
		timeout:  5 * time.Second,
	}

	h.RegisterFunc("unhealthy", func(ctx context.Context) CheckResult {
		return CheckResult{Status: StatusDown}
	})

	result := h.Readiness(context.Background())
	if result.Status != StatusDown {
		t.Errorf("Expected readiness status DOWN, got %s", result.Status)
	}
}

func TestCustomChecker(t *testing.T) {
	checker := NewCustomChecker("custom", func(ctx context.Context) error {
		return nil
	})

	result := checker.Check(context.Background())
	if result.Status != StatusUp {
		t.Errorf("Expected status UP, got %s", result.Status)
	}

	// 测试失败情况
	failChecker := NewCustomChecker("fail", func(ctx context.Context) error {
		return errors.New("something went wrong")
	})

	result = failChecker.Check(context.Background())
	if result.Status != StatusDown {
		t.Errorf("Expected status DOWN, got %s", result.Status)
	}
}

func TestFuncChecker(t *testing.T) {
	checker := NewChecker("func-test", func(ctx context.Context) CheckResult {
		return CheckResult{
			Status: StatusUp,
			Details: map[string]any{
				"key": "value",
			},
		}
	})

	if checker.Name() != "func-test" {
		t.Errorf("Expected name 'func-test', got %s", checker.Name())
	}

	result := checker.Check(context.Background())
	if result.Status != StatusUp {
		t.Errorf("Expected status UP, got %s", result.Status)
	}
	if result.Details["key"] != "value" {
		t.Error("Expected details to contain key=value")
	}
}

func TestHealth_Timeout(t *testing.T) {
	h := &Health{
		checkers: make([]Checker, 0),
		timeout:  100 * time.Millisecond,
	}

	h.RegisterFunc("slow", func(ctx context.Context) CheckResult {
		select {
		case <-ctx.Done():
			return CheckResult{Status: StatusDown, Message: "timeout"}
		case <-time.After(200 * time.Millisecond):
			return CheckResult{Status: StatusUp}
		}
	})

	result := h.Check(context.Background())
	// 检查应该因为超时而返回 DOWN
	if result.Checks["slow"].Status != StatusDown {
		t.Errorf("Expected slow check to be DOWN due to timeout, got %s", result.Checks["slow"].Status)
	}
}

func TestDefaultHealth(t *testing.T) {
	// 重置默认实例
	defaultHealth = nil

	// 应该自动初始化
	h := Get()
	if h == nil {
		t.Error("Expected Get() to return non-nil")
	}

	// 注册检查器
	RegisterFunc("default-test", func(ctx context.Context) CheckResult {
		return CheckResult{Status: StatusUp}
	})

	result := Check(context.Background())
	if result.Status != StatusUp {
		t.Errorf("Expected status UP, got %s", result.Status)
	}
}
