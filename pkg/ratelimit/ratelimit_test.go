package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestNewMemory(t *testing.T) {
	limiter := NewMemory()
	if limiter == nil {
		t.Fatal("Expected non-nil limiter")
	}

	ctx := context.Background()
	key := "test-key"
	limit := 5
	window := time.Second

	// First request should be allowed
	allowed, remaining, resetAt := limiter.Allow(ctx, key, limit, window)
	if !allowed {
		t.Error("First request should be allowed")
	}

	if remaining != limit-1 {
		t.Errorf("Expected remaining %d, got %d", limit-1, remaining)
	}

	if resetAt.IsZero() {
		t.Error("Reset time should not be zero")
	}
}

func TestMemoryLimiterExceedLimit(t *testing.T) {
	limiter := NewMemory()
	ctx := context.Background()
	key := "test-key-exceed"
	limit := 3
	window := time.Second

	// Make requests up to limit
	for i := 0; i < limit; i++ {
		allowed, _, _ := limiter.Allow(ctx, key, limit, window)
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Next request should be denied
	allowed, remaining, _ := limiter.Allow(ctx, key, limit, window)
	if allowed {
		t.Error("Request exceeding limit should be denied")
	}

	if remaining != 0 {
		t.Errorf("Expected remaining 0, got %d", remaining)
	}
}

func TestMemoryLimiterWindowReset(t *testing.T) {
	limiter := NewMemory()
	ctx := context.Background()
	key := "test-key-reset"
	limit := 2
	window := 100 * time.Millisecond

	// Exhaust limit
	limiter.Allow(ctx, key, limit, window)
	limiter.Allow(ctx, key, limit, window)

	// Should be denied
	allowed, _, _ := limiter.Allow(ctx, key, limit, window)
	if allowed {
		t.Error("Should be denied before window reset")
	}

	// Wait for window to reset
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again
	allowed, remaining, _ := limiter.Allow(ctx, key, limit, window)
	if !allowed {
		t.Error("Should be allowed after window reset")
	}

	if remaining != limit-1 {
		t.Errorf("Expected remaining %d after reset, got %d", limit-1, remaining)
	}
}

func TestNewWithConfig(t *testing.T) {
	cfg := Config{
		Store:     "memory",
		KeyPrefix: "custom:",
	}

	limiter := New(cfg)
	if limiter == nil {
		t.Fatal("Expected non-nil limiter")
	}

	// Test basic functionality
	ctx := context.Background()
	allowed, _, _ := limiter.Allow(ctx, "test", 10, time.Second)
	if !allowed {
		t.Error("First request should be allowed")
	}
}

func TestDifferentKeys(t *testing.T) {
	limiter := NewMemory()
	ctx := context.Background()
	limit := 1
	window := time.Second

	// Use up limit for key1
	allowed, _, _ := limiter.Allow(ctx, "key1", limit, window)
	if !allowed {
		t.Error("First request for key1 should be allowed")
	}

	allowed, _, _ = limiter.Allow(ctx, "key1", limit, window)
	if allowed {
		t.Error("Second request for key1 should be denied")
	}

	// key2 should still be allowed
	allowed, _, _ = limiter.Allow(ctx, "key2", limit, window)
	if !allowed {
		t.Error("First request for key2 should be allowed")
	}
}
