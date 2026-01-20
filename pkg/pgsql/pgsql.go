package pgsql

import (
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"xorm.io/xorm"
	"xorm.io/xorm/contexts"
)

// Config represents PostgreSQL configuration
// Config 表示 PostgreSQL 配置
type Config struct {
	Host        string `toml:"host"`
	Port        int    `toml:"port"`
	User        string `toml:"user"`
	Password    string `toml:"password"`
	DBName      string `toml:"db_name"`
	AutoMigrate bool   `toml:"auto_migrate"` // Auto migrate database schema | 自动迁移数据库架构
	ShowSQL     bool   `toml:"show_sql"`     // Show SQL logs | 显示 SQL 日志
}

// DSN generates connection string
// DSN 生成连接字符串
func (c Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DBName)
}

// Client wraps PostgreSQL client
// Client 封装 PostgreSQL 客户端
type Client struct {
	engine     *xorm.Engine
	xormLogger *XormLogger
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
	mu.Unlock()
	return nil
}

// InitNamed initializes a named database client
// InitNamed 初始化命名数据库客户端
func InitNamed(name string, cfg Config) error {
	client, err := New(cfg)
	if err != nil {
		return fmt.Errorf("failed to init named client %s: %w", name, err)
	}
	mu.Lock()
	clients[name] = client
	mu.Unlock()
	return nil
}

// InitMultiple initializes multiple database clients
// InitMultiple 初始化多个数据库客户端
// Usage | 用法:
//
//	configs := map[string]Config{
//	    "usercenter": {...},
//	    "organization": {...},
//	}
//	err := pgsql.InitMultiple(configs)
func InitMultiple(configs map[string]Config) error {
	for name, cfg := range configs {
		if err := InitNamed(name, cfg); err != nil {
			return err
		}
	}
	return nil
}

// Get returns database client
// Get 返回数据库客户端
// Usage | 用法:
//
//	pgsql.Get()              // returns default database | 返回默认数据库
//	pgsql.Get("usercenter")  // returns usercenter database | 返回 usercenter 数据库
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
	if defaultClient != nil {
		defaultClient.Close()
	}
	mu.Lock()
	for _, client := range clients {
		if client != nil {
			client.Close()
		}
	}
	clients = make(map[string]*Client)
	mu.Unlock()
}

// SetLogger sets the logger for all pgsql clients
// SetLogger 设置所有 pgsql 客户端的日志器
// Call this after logger is initialized | 在日志器初始化后调用
func SetLogger(logger Logger) {
	// Set logger for default client | 设置默认客户端的日志器
	if defaultClient != nil && defaultClient.xormLogger != nil {
		defaultClient.xormLogger.SetLogger(logger)
	}

	// Set logger for all named clients | 设置所有命名客户端的日志器
	mu.RLock()
	for _, client := range clients {
		if client != nil && client.xormLogger != nil {
			client.xormLogger.SetLogger(logger)
		}
	}
	mu.RUnlock()
}

// MustInit initializes and panics on error
// MustInit 初始化，出错时 panic
func MustInit(cfg Config) {
	if err := Init(cfg); err != nil {
		log.Fatalf("pgsql initialization failed: %v", err)
	}
}

// New creates PostgreSQL client
// New 创建 PostgreSQL 客户端
func New(cfg Config) (*Client, error) {
	engine, err := xorm.NewEngine("postgres", cfg.DSN())
	if err != nil {
		return nil, err
	}

	// Configure connection pool | 配置连接池
	engine.SetMaxOpenConns(100)                      // Maximum open connections | 最大连接数
	engine.SetMaxIdleConns(20)                       // Maximum idle connections | 最大空闲连接
	engine.SetConnMaxLifetime(time.Hour)             // Connection max lifetime | 连接最大生命周期
	engine.SetConnMaxIdleTime(10 * time.Minute)      // Idle connection timeout | 空闲连接超时

	xormLogger := NewXormLogger(cfg.ShowSQL)
	engine.SetLogger(xormLogger)

	// Add slow query monitoring hook | 添加慢查询监控钩子
	slowQueryHook := NewSlowQueryHook(time.Second, nil)
	engine.AddHook(slowQueryHook)

	return &Client{engine: engine, xormLogger: xormLogger}, nil
}

// AddHook adds hook for external injection (e.g., tracing)
// AddHook 添加钩子用于外部注入（例如追踪）
func (c *Client) AddHook(hook contexts.Hook) {
	c.engine.AddHook(hook)
}

// Engine returns xorm engine
// Engine 返回 xorm 引擎
func (c *Client) Engine() *xorm.Engine {
	return c.engine
}

// Sync synchronizes table structure
// Sync 同步表结构
func (c *Client) Sync(beans ...any) error {
	return c.engine.Sync(beans...)
}

// Close closes connection
// Close 关闭连接
func (c *Client) Close() error {
	return c.engine.Close()
}
