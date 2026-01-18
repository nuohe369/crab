package redis

import (
	"testing"
)

// TestMultipleRedisConfig tests multiple Redis configuration
// TestMultipleRedisConfig 测试多 Redis 配置
func TestMultipleRedisConfig(t *testing.T) {
	configs := map[string]Config{
		"default": {
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		"cache": {
			Addr:     "localhost:6380",
			Password: "",
			DB:       0,
		},
		"session": {
			Addr:     "localhost:6381",
			Password: "",
			DB:       1,
		},
	}

	// Test that InitMultiple doesn't panic with valid config
	// 测试 InitMultiple 在有效配置下不会 panic
	err := InitMultiple(configs)
	if err == nil {
		t.Log("InitMultiple succeeded (Redis servers must be running)")

		// Test Get with default
		// 测试获取默认 Redis
		defaultClient := Get()
		if defaultClient == nil {
			t.Error("Expected default client to be non-nil")
		}

		// Test Get with "default" name
		// 测试用 "default" 名称获取
		defaultByName := Get("default")
		if defaultByName == nil {
			t.Error("Expected default client by name to be non-nil")
		}
		if defaultClient != defaultByName {
			t.Error("Expected Get() and Get('default') to return same instance")
		}

		// Test Get with name
		// 测试获取命名 Redis
		cacheClient := Get("cache")
		if cacheClient == nil {
			t.Error("Expected cache client to be non-nil")
		}

		sessionClient := Get("session")
		if sessionClient == nil {
			t.Error("Expected session client to be non-nil")
		}

		// Test Get with non-existent name
		// 测试获取不存在的 Redis
		nonExistent := Get("nonexistent")
		if nonExistent != nil {
			t.Error("Expected non-existent client to be nil")
		}

		// Cleanup
		Close()
	} else {
		t.Logf("InitMultiple failed (expected if Redis is not running): %v", err)
	}
}

// TestSingleRedisBackwardCompatibility tests backward compatibility with single Redis
// TestSingleRedisBackwardCompatibility 测试单 Redis 向后兼容性
func TestSingleRedisBackwardCompatibility(t *testing.T) {
	cfg := Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	// Test that Init doesn't panic with valid config
	// 测试 Init 在有效配置下不会 panic
	err := Init(cfg)
	if err == nil {
		t.Log("Init succeeded (Redis server must be running)")

		// Test Get without name
		// 测试不带名称的 Get
		client := Get()
		if client == nil {
			t.Error("Expected client to be non-nil")
		}

		// Test Get with "default" name should return same instance
		// 测试用 "default" 名称获取应返回相同实例
		defaultByName := Get("default")
		if defaultByName == nil {
			t.Error("Expected default client by name to be non-nil")
		}
		if client != defaultByName {
			t.Error("Expected Get() and Get('default') to return same instance")
		}

		// Cleanup
		Close()
	} else {
		t.Logf("Init failed (expected if Redis is not running): %v", err)
	}
}

// TestInitNamed tests InitNamed function
// TestInitNamed 测试 InitNamed 函数
func TestInitNamed(t *testing.T) {
	// Reset state
	// 重置状态
	defaultClient = nil
	clients = make(map[string]*Client)

	cfg := Config{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	// Test InitNamed with "default" name
	// 测试用 "default" 名称初始化
	err := InitNamed("default", cfg)
	if err == nil {
		t.Log("InitNamed succeeded (Redis server must be running)")

		// Should set as default client
		// 应该设置为默认客户端
		if defaultClient == nil {
			t.Error("Expected defaultClient to be set when InitNamed with 'default'")
		}

		// Should be accessible by Get()
		// 应该可以通过 Get() 访问
		client := Get()
		if client == nil {
			t.Error("Expected Get() to return client after InitNamed('default')")
		}

		// Should be accessible by Get("default")
		// 应该可以通过 Get("default") 访问
		clientByName := Get("default")
		if clientByName == nil {
			t.Error("Expected Get('default') to return client")
		}

		if client != clientByName {
			t.Error("Expected Get() and Get('default') to return same instance")
		}

		// Cleanup
		Close()
	} else {
		t.Logf("InitNamed failed (expected if Redis is not running): %v", err)
	}
}

// TestGetWithoutInit tests Get behavior before initialization
// TestGetWithoutInit 测试初始化前的 Get 行为
func TestGetWithoutInit(t *testing.T) {
	// Reset state
	// 重置状态
	defaultClient = nil
	clients = make(map[string]*Client)

	// Get should return nil before initialization
	// 初始化前 Get 应该返回 nil
	client := Get()
	if client != nil {
		t.Error("Expected nil client before initialization")
	}

	namedClient := Get("cache")
	if namedClient != nil {
		t.Error("Expected nil named client before initialization")
	}
}

// TestCloseAllClients tests that Close closes all clients
// TestCloseAllClients 测试 Close 关闭所有客户端
func TestCloseAllClients(t *testing.T) {
	// Reset state
	// 重置状态
	defaultClient = nil
	clients = make(map[string]*Client)

	configs := map[string]Config{
		"default": {
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		},
		"cache": {
			Addr:     "localhost:6380",
			Password: "",
			DB:       0,
		},
	}

	err := InitMultiple(configs)
	if err == nil {
		t.Log("InitMultiple succeeded")

		// Verify clients exist
		// 验证客户端存在
		if Get() == nil {
			t.Error("Expected default client to exist")
		}
		if Get("cache") == nil {
			t.Error("Expected cache client to exist")
		}

		// Close all
		// 关闭所有
		Close()

		// Verify all clients are closed
		// 验证所有客户端已关闭
		if Get() != nil {
			t.Error("Expected default client to be nil after Close()")
		}
		if Get("cache") != nil {
			t.Error("Expected cache client to be nil after Close()")
		}
		if Get("default") != nil {
			t.Error("Expected default client by name to be nil after Close()")
		}
	} else {
		t.Logf("InitMultiple failed (expected if Redis is not running): %v", err)
	}
}
