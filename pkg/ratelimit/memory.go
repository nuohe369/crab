package ratelimit

import (
	"context"
	"sync"
	"time"
)

// memoryLimiter memory rate limiter (sliding window)
type memoryLimiter struct {
	prefix  string
	buckets sync.Map // map[string]*bucket
}

type bucket struct {
	mu       sync.Mutex
	count    int       // Current window request count
	windowAt time.Time // Window start time
}

func newMemoryLimiter(prefix string) *memoryLimiter {
	m := &memoryLimiter{prefix: prefix}
	// Start cleanup goroutine
	go m.cleanup()
	return m
}

func (m *memoryLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Time) {
	fullKey := m.prefix + key
	now := time.Now()

	// Get or create bucket
	val, _ := m.buckets.LoadOrStore(fullKey, &bucket{windowAt: now})
	b := val.(*bucket)

	b.mu.Lock()
	defer b.mu.Unlock()

	// Check if window needs reset
	if now.Sub(b.windowAt) >= window {
		b.count = 0
		b.windowAt = now
	}

	resetAt := b.windowAt.Add(window)
	remaining := limit - b.count - 1

	// Check if limit exceeded
	if b.count >= limit {
		return false, 0, resetAt
	}

	b.count++
	if remaining < 0 {
		remaining = 0
	}

	return true, remaining, resetAt
}

// cleanup periodically cleans up expired buckets
func (m *memoryLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		m.buckets.Range(func(key, value any) bool {
			b := value.(*bucket)
			b.mu.Lock()
			// Delete buckets not accessed for over 10 minutes
			if now.Sub(b.windowAt) > 10*time.Minute {
				m.buckets.Delete(key)
			}
			b.mu.Unlock()
			return true
		})
	}
}
