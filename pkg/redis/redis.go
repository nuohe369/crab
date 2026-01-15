package redis

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// UniversalClient is the universal Redis client interface (supports standalone and cluster)
type UniversalClient interface {
	redis.Cmdable
	Close() error
}

// Client wraps Redis client
type Client struct {
	client UniversalClient
}

// Config represents Redis configuration
type Config struct {
	// Standalone mode
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`

	// Cluster mode (comma-separated address list, e.g. "host1:6379,host2:6379,host3:6379")
	Cluster string `toml:"cluster"`
}

var defaultClient *Client

// Init initializes default client
func Init(cfg Config) error {
	client, err := New(cfg)
	if err != nil {
		return err
	}
	defaultClient = client
	return nil
}

// MustInit initializes and panics on error
func MustInit(cfg Config) {
	if err := Init(cfg); err != nil {
		log.Fatalf("redis initialization failed: %v", err)
	}
}

// Get returns default client
func Get() *Client {
	return defaultClient
}

// Close closes default client
func Close() {
	if defaultClient != nil {
		defaultClient.Close()
	}
}

// New creates Redis client (auto detect standalone/cluster mode)
func New(cfg Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var client UniversalClient

	if cfg.Cluster != "" {
		// Cluster mode
		addrs := strings.Split(cfg.Cluster, ",")
		for i := range addrs {
			addrs[i] = strings.TrimSpace(addrs[i])
		}

		clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    addrs,
			Password: cfg.Password,
		})

		if err := clusterClient.Ping(ctx).Err(); err != nil {
			return nil, err
		}

		client = clusterClient
		log.Printf("redis: cluster mode connected, nodes: %v", addrs)
	} else {
		// Standalone mode
		standaloneClient := redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
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
//
// For scenarios requiring direct access to underlying client (e.g. Pub/Sub).
// Returns type may be *redis.Client or *redis.ClusterClient.
func (c *Client) GetRaw() any {
	return c.client
}

// Set sets value
func (c *Client) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// SetNX sets value (only if key does not exist)
func (c *Client) SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, expiration).Result()
}

// Get gets value
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Del deletes keys
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists checks if key exists
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	return n > 0, err
}

// HSet sets Hash field
func (c *Client) HSet(ctx context.Context, key, field string, value any) error {
	return c.client.HSet(ctx, key, field, value).Err()
}

// HGet gets Hash field
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	return c.client.HGet(ctx, key, field).Result()
}

// HGetAll gets all Hash fields
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

// HDel deletes Hash fields
func (c *Client) HDel(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, key, fields...).Err()
}

// HExists checks if Hash field exists
func (c *Client) HExists(ctx context.Context, key, field string) (bool, error) {
	return c.client.HExists(ctx, key, field).Result()
}

// Close closes connection
func (c *Client) Close() error {
	return c.client.Close()
}

// Keys finds matching keys (supports wildcards)
// Note: Use with caution in production with large data, recommend using Scan
func (c *Client) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
}
