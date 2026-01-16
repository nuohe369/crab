package breaker

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestCircuitBreaker_ClosedState(t *testing.T) {
	cb := New("test-closed", Config{
		MaxRequests:  3,
		Timeout:      100 * time.Millisecond,
		FailureRatio: 0.6,
		MinRequests:  5,
	})

	// 正常请求应该通过
	err := cb.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cb.State() != StateClosed {
		t.Errorf("Expected state Closed, got %v", cb.State())
	}
}

func TestCircuitBreaker_OpenAfterFailures(t *testing.T) {
	cb := New("test-open", Config{
		MaxRequests:  3,
		Timeout:      100 * time.Millisecond,
		FailureRatio: 0.5,
		MinRequests:  4,
	})

	testErr := errors.New("test error")

	// 触发 4 次失败（满足 MinRequests，且失败率 100% > 50%）
	for i := 0; i < 4; i++ {
		cb.Execute(func() error {
			return testErr
		})
	}

	// 应该进入 Open 状态
	if cb.State() != StateOpen {
		t.Errorf("Expected state Open, got %v", cb.State())
	}

	// 后续请求应该被拒绝
	err := cb.Execute(func() error {
		return nil
	})
	if !IsOpen(err) {
		t.Errorf("Expected ErrOpenState, got %v", err)
	}
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := New("test-halfopen", Config{
		MaxRequests:  2,
		Timeout:      50 * time.Millisecond,
		FailureRatio: 0.5,
		MinRequests:  2,
	})

	testErr := errors.New("test error")

	// 触发熔断
	for i := 0; i < 2; i++ {
		cb.Execute(func() error {
			return testErr
		})
	}

	if cb.State() != StateOpen {
		t.Errorf("Expected state Open, got %v", cb.State())
	}

	// 等待超时
	time.Sleep(60 * time.Millisecond)

	// 下一个请求应该被允许（进入半开状态）
	err := cb.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// gobreaker 需要 MaxRequests 次成功才能恢复
	// 半开状态下再成功一次
	cb.Execute(func() error {
		return nil
	})

	// 现在应该恢复到 Closed
	if cb.State() != StateClosed {
		t.Errorf("Expected state Closed after successes, got %v", cb.State())
	}
}

func TestCircuitBreaker_ExecuteWithResult(t *testing.T) {
	cb := New("test-result", DefaultConfig())

	result, err := cb.ExecuteWithResult(func() (any, error) {
		return "hello", nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "hello" {
		t.Errorf("Expected 'hello', got %v", result)
	}
}

func TestManager_GetBreaker(t *testing.T) {
	// 重新初始化
	defaultManager = nil
	managerOnce = sync.Once{}
	Init(DefaultConfig())

	cb1 := GetBreaker("service1")
	cb2 := GetBreaker("service1")
	cb3 := GetBreaker("service2")

	if cb1 != cb2 {
		t.Error("Expected same breaker for same name")
	}

	if cb1 == cb3 {
		t.Error("Expected different breaker for different name")
	}
}

func TestManager_GetAllStats(t *testing.T) {
	// 重新初始化
	defaultManager = nil
	managerOnce = sync.Once{}
	Init(DefaultConfig())

	GetBreaker("stats-test-1")
	GetBreaker("stats-test-2")

	stats := GetManager().GetAllStats()
	if len(stats) < 2 {
		t.Errorf("Expected at least 2 stats, got %d", len(stats))
	}
}
