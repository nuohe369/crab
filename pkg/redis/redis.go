package redis

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// UniversalClient is the universal Redis client interface (supports standalone and cluster)
// UniversalClient 是通用的 Redis 客户端接口（支持单机和集群模式）
type UniversalClient interface {
	redis.Cmdable
	Close() error
}

// Client wraps Redis client
// Client 封装 Redis 客户端
type Client struct {
	client UniversalClient
}

// Config represents Redis configuration
// Config 表示 Redis 配置
type Config struct {
	// Standalone mode | 单机模式
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`

	// Cluster mode (comma-separated address list, e.g. "host1:6379,host2:6379,host3:6379")
	// 集群模式（逗号分隔的地址列表，例如 "host1:6379,host2:6379,host3:6379"）
	Cluster string `toml:"cluster"`
}

var (
	clients       = make(map[string]*Client)
	defaultClient *Client
	mu            sync.RWMutex
)

// Init initializes default client
// Init 初始化默认客户端
func Init(cfg Config) error {
	client, err := New(cfg)
	if err != nil {
		return err
	}
	mu.Lock()
	defaultClient = client
	clients[""] = client        // Store with empty key | 用空键存储
	clients["default"] = client // Store with "default" key | 用 "default" 键存储
	mu.Unlock()
	return nil
}

// MustInit initializes and panics on error
// MustInit 初始化，出错时 panic
func MustInit(cfg Config) {
	if err := Init(cfg); err != nil {
		log.Fatalf("redis initialization failed: %v", err)
	}
}

// InitNamed initializes a named Redis client
// InitNamed 初始化命名的 Redis 客户端
func InitNamed(name string, cfg Config) error {
	client, err := New(cfg)
	if err != nil {
		return fmt.Errorf("failed to init named client %s: %w", name, err)
	}
	mu.Lock()
	clients[name] = client
	// If name is "default" or "", also set as default client
	// 如果名称是 "default" 或 ""，也设置为默认客户端
	if name == "default" || name == "" {
		defaultClient = client
	}
	mu.Unlock()
	return nil
}

// InitMultiple initializes multiple Redis clients
// InitMultiple 初始化多个 Redis 客户端
// Usage | 用法:
//
//	configs := map[string]Config{
//	    "default": {...},
//	    "cache": {...},
//	}
//	err := redis.InitMultiple(configs)
func InitMultiple(configs map[string]Config) error {
	for name, cfg := range configs {
		if err := InitNamed(name, cfg); err != nil {
			return err
		}
	}
	return nil
}

// MustInitMultiple initializes multiple Redis clients and panics on error
// MustInitMultiple 初始化多个 Redis 客户端，出错时 panic
func MustInitMultiple(configs map[string]Config) {
	if err := InitMultiple(configs); err != nil {
		log.Fatalf("redis initialization failed: %v", err)
	}
}

// Get returns Redis client
// Get 返回 Redis 客户端
// Usage | 用法:
//
//	redis.Get()           // returns default Redis | 返回默认 Redis
//	redis.Get("cache")    // returns cache Redis | 返回 cache Redis
func Get(name ...string) *Client {
	if len(name) == 0 {
		return defaultClient
	}
	mu.RLock()
	defer mu.RUnlock()
	return clients[name[0]]
}

// Close closes default client and all named clients
// Close 关闭默认客户端和所有命名客户端
func Close() {
	mu.Lock()
	for _, client := range clients {
		if client != nil {
			client.Close()
		}
	}
	clients = make(map[string]*Client)
	defaultClient = nil
	mu.Unlock()
}

// New creates Redis client (auto detect standalone/cluster mode)
// New 创建 Redis 客户端（自动检测单机/集群模式）
func New(cfg Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var client UniversalClient

	if cfg.Cluster != "" {
		// Cluster mode | 集群模式
		addrs := strings.Split(cfg.Cluster, ",")
		for i := range addrs {
			addrs[i] = strings.TrimSpace(addrs[i])
		}

		clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    addrs,
			Password: cfg.Password,
			
			// Connection pool configuration | 连接池配置
			PoolSize:           100,                   // Connection pool size | 连接池大小
			MinIdleConns:       10,                    // Minimum idle connections | 最小空闲连接
			MaxIdleConns:       50,                    // Maximum idle connections | 最大空闲连接
			ConnMaxIdleTime:    10 * time.Minute,      // Idle connection timeout | 空闲连接超时
			ConnMaxLifetime:    time.Hour,             // Connection max lifetime | 连接最大生命周期
			PoolTimeout:        4 * time.Second,       // Get connection timeout | 获取连接超时
			
			// Retry configuration | 重试配置
			MaxRetries:         3,                     // Max retry attempts | 最大重试次数
			MinRetryBackoff:    8 * time.Millisecond,  // Min retry backoff | 最小重试间隔
			MaxRetryBackoff:    512 * time.Millisecond, // Max retry backoff | 最大重试间隔
		})

		if err := clusterClient.Ping(ctx).Err(); err != nil {
			return nil, err
		}

		client = clusterClient
		log.Printf("redis: cluster mode connected, nodes: %v", addrs)
	} else {
		// Standalone mode | 单机模式
		standaloneClient := redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
			
			// Connection pool configuration | 连接池配置
			PoolSize:           100,                   // Connection pool size | 连接池大小
			MinIdleConns:       10,                    // Minimum idle connections | 最小空闲连接
			MaxIdleConns:       50,                    // Maximum idle connections | 最大空闲连接
			ConnMaxIdleTime:    10 * time.Minute,      // Idle connection timeout | 空闲连接超时
			ConnMaxLifetime:    time.Hour,             // Connection max lifetime | 连接最大生命周期
			PoolTimeout:        4 * time.Second,       // Get connection timeout | 获取连接超时
			
			// Retry configuration | 重试配置
			MaxRetries:         3,                     // Max retry attempts | 最大重试次数
			MinRetryBackoff:    8 * time.Millisecond,  // Min retry backoff | 最小重试间隔
			MaxRetryBackoff:    512 * time.Millisecond, // Max retry backoff | 最大重试间隔
		})

		if err := standaloneClient.Ping(ctx).Err(); err != nil {
			return nil, err
		}

		client = standaloneClient
		log.Printf("redis: standalone mode connected, address: %s", cfg.Addr)
	}

	return &Client{client: client}, nil
}

// GetRaw returns underlying Redis client
// GetRaw 返回底层 Redis 客户端
//
// For scenarios requiring direct access to underlying client (e.g. Pub/Sub).
// Returns type may be *redis.Client or *redis.ClusterClient.
// 用于需要直接访问底层客户端的场景（例如 Pub/Sub）。
// 返回类型可能是 *redis.Client 或 *redis.ClusterClient。
func (c *Client) GetRaw() any {
	return c.client
}

// Set sets value
// Set 设置值
func (c *Client) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// SetNX sets value (only if key does not exist)
// SetNX 设置值（仅当键不存在时）
func (c *Client) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, expiration).Result()
}

// Get gets value
// Get 获取值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Del deletes keys
// Del 删除键
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists checks if key exists
// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	return n > 0, err
}

// HSet sets Hash field
// HSet 设置 Hash 字段
func (c *Client) HSet(ctx context.Context, key, field string, value any) error {
	return c.client.HSet(ctx, key, field, value).Err()
}

// HGet gets Hash field
// HGet 获取 Hash 字段
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return c.client.HGet(ctx, key, field).Result()
}

// HGetAll gets all Hash fields
// HGetAll 获取所有 Hash 字段
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

// HDel deletes Hash fields
// HDel 删除 Hash 字段
func (c *Client) HDel(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, key, fields...).Err()
}

// HExists checks if Hash field exists
// HExists 检查 Hash 字段是否存在
func (c *Client) HExists(ctx context.Context, key, field string) (bool, error) {
	return c.client.HExists(ctx, key, field).Result()
}

// Close closes connection
// Close 关闭连接
func (c *Client) Close() error {
	return c.client.Close()
}

// Keys finds matching keys (supports wildcards)
// Note: Use with caution in production with large data, recommend using Scan
// Keys 查找匹配的键（支持通配符）
// 注意：在生产环境中处理大量数据时谨慎使用，建议使用 Scan
func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
}
