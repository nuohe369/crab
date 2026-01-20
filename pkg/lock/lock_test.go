package lock

import (
	"testing"
	"time"
)

// 注意：这些测试需要真实的 Redis 实例
// miniredis 对 redsync 使用的 Lua 脚本支持有限
// 运行测试前请确保 Redis 在 localhost:6379 运行

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Expiry != 8*time.Second {
		t.Errorf("Expected expiry 8s, got %v", cfg.Expiry)
	}
	if cfg.Tries != 32 {
		t.Errorf("Expected tries 32, got %d", cfg.Tries)
	}
	if cfg.RetryDelay != 100*time.Millisecond {
		t.Errorf("Expected retry delay 100ms, got %v", cfg.RetryDelay)
	}
}

func TestWithExpiry_Option(t *testing.T) {
	cfg := DefaultConfig()
	opt := WithExpiry(5 * time.Second)
	opt(&cfg)

	if cfg.Expiry != 5*time.Second {
		t.Errorf("Expected expiry 5s, got %v", cfg.Expiry)
	}
}

func TestWithTries_Option(t *testing.T) {
	cfg := DefaultConfig()
	opt := WithTries(10)
	opt(&cfg)

	if cfg.Tries != 10 {
		t.Errorf("Expected tries 10, got %d", cfg.Tries)
	}
}

func TestWithRetryDelay_Option(t *testing.T) {
	cfg := DefaultConfig()
	opt := WithRetryDelay(200 * time.Millisecond)
	opt(&cfg)

	if cfg.RetryDelay != 200*time.Millisecond {
		t.Errorf("Expected retry delay 200ms, got %v", cfg.RetryDelay)
	}
}

func TestNewMutex_NotInitialized(t *testing.T) {
	// 确保未初始化时返回 nil 而不是 panic
	defaultLocker = nil

	mu := NewMutex("test")
	if mu != nil {
		t.Error("Expected nil when locker not initialized")
	}
}

// 以下测试需要真实 Redis，默认跳过
// 运行方式: go test -v -run TestIntegration -tags=integration

/*
func TestIntegration_LockUnlock(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	Init(client, DefaultConfig())

	mu := NewMutex("test:integration:1")

	err := mu.Lock()
	if err != nil {
		t.Fatalf("Failed to acquire lock: %v", err)
	}

	if !mu.IsLocked() {
		t.Error("Expected IsLocked to be true")
	}

	ok, err := mu.Unlock()
	if err != nil {
		t.Fatalf("Failed to release lock: %v", err)
	}
	if !ok {
		t.Error("Expected unlock to succeed")
	}
}
*/
