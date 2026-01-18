package cache

import (
	"context"
	"testing"
	"time"
)

func TestLocalCache(t *testing.T) {
	cache := newLocalCache(100, time.Minute)

	// Test set and get
	cache.set("key1", []byte("value1"))

	val, ok := cache.get("key1")
	if !ok {
		t.Error("Expected to find key1")
	}
	if string(val) != "value1" {
		t.Errorf("Value mismatch: got %s, want value1", string(val))
	}

	// Test non-existent key
	_, ok = cache.get("nonexistent")
	if ok {
		t.Error("Should not find nonexistent key")
	}

	// Test delete
	cache.del("key1")
	_, ok = cache.get("key1")
	if ok {
		t.Error("Key should be deleted")
	}
}

func TestLocalCacheExpiration(t *testing.T) {
	cache := newLocalCache(100, 50*time.Millisecond)

	cache.set("key1", []byte("value1"))

	// Should exist immediately
	_, ok := cache.get("key1")
	if !ok {
		t.Error("Key should exist immediately after set")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	_, ok = cache.get("key1")
	if ok {
		t.Error("Key should be expired")
	}
}

func TestLocalCacheLen(t *testing.T) {
	cache := newLocalCache(100, time.Minute)

	if cache.Len() != 0 {
		t.Error("New cache should be empty")
	}

	cache.set("key1", []byte("value1"))
	cache.set("key2", []byte("value2"))

	if cache.Len() != 2 {
		t.Errorf("Expected 2 items, got %d", cache.Len())
	}

	cache.del("key1")

	if cache.Len() != 1 {
		t.Errorf("Expected 1 item, got %d", cache.Len())
	}
}

func TestLocalCacheEviction(t *testing.T) {
	cache := newLocalCache(3, 10*time.Millisecond) // Short TTL for eviction

	cache.set("key1", []byte("value1"))
	cache.set("key2", []byte("value2"))
	cache.set("key3", []byte("value3"))

	if cache.Len() != 3 {
		t.Errorf("Expected 3 items, got %d", cache.Len())
	}

	// Wait for items to expire
	time.Sleep(20 * time.Millisecond)

	// Adding 4th item should trigger eviction of expired items
	cache.set("key4", []byte("value4"))

	// After eviction, only key4 should remain (others expired)
	if cache.Len() > 3 {
		t.Errorf("Cache should not exceed maxSize after eviction, got %d", cache.Len())
	}
}

// Mock Redis client for testing
type mockRedisClient struct {
	data map[string]string
}

func newMockRedis() *mockRedisClient {
	return &mockRedisClient{data: make(map[string]string)}
}

func (m *mockRedisClient) Get(ctx context.Context, key string) (string, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return "", &notFoundError{}
}

func (m *mockRedisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	m.data[key] = value.(string)
	return nil
}

func (m *mockRedisClient) Del(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		delete(m.data, key)
	}
	return nil
}

type notFoundError struct{}

func (e *notFoundError) Error() string { return "redis: nil" }

func TestCacheWithRedis(t *testing.T) {
	redis := newMockRedis()
	cache := New(redis, Config{
		LocalTTL:    time.Minute,
		LocalSize:   100,
		EnableLocal: true,
	})

	ctx := context.Background()

	// Test SetValue and GetValue
	type testData struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	original := testData{Name: "test", Age: 25}
	err := cache.SetValue(ctx, "user:1", original, time.Hour)
	if err != nil {
		t.Fatalf("SetValue failed: %v", err)
	}

	var result testData
	err = cache.GetValue(ctx, "user:1", &result)
	if err != nil {
		t.Fatalf("GetValue failed: %v", err)
	}

	if result.Name != original.Name || result.Age != original.Age {
		t.Errorf("Data mismatch: got %+v, want %+v", result, original)
	}
}

func TestCacheGetOrSet(t *testing.T) {
	redis := newMockRedis()
	cache := New(redis, Config{
		LocalTTL:    time.Minute,
		LocalSize:   100,
		EnableLocal: true,
	})

	ctx := context.Background()
	loadCount := 0

	loader := func() (any, error) {
		loadCount++
		return map[string]string{"loaded": "data"}, nil
	}

	var result map[string]string

	// First call should trigger loader
	err := cache.GetOrSet(ctx, "key1", &result, time.Hour, loader)
	if err != nil {
		t.Fatalf("GetOrSet failed: %v", err)
	}

	if loadCount != 1 {
		t.Errorf("Loader should be called once, got %d", loadCount)
	}

	if result["loaded"] != "data" {
		t.Errorf("Result mismatch: %+v", result)
	}

	// Second call should use cache
	var result2 map[string]string
	err = cache.GetOrSet(ctx, "key1", &result2, time.Hour, loader)
	if err != nil {
		t.Fatalf("GetOrSet failed: %v", err)
	}

	if loadCount != 1 {
		t.Errorf("Loader should not be called again, got %d", loadCount)
	}
}

func TestCacheDel(t *testing.T) {
	redis := newMockRedis()
	cache := New(redis, Config{
		LocalTTL:    time.Minute,
		LocalSize:   100,
		EnableLocal: true,
	})

	ctx := context.Background()

	cache.SetValue(ctx, "key1", "value1", time.Hour)
	cache.SetValue(ctx, "key2", "value2", time.Hour)

	err := cache.Del(ctx, "key1", "key2")
	if err != nil {
		t.Fatalf("Del failed: %v", err)
	}

	var result string
	err = cache.GetValue(ctx, "key1", &result)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}
