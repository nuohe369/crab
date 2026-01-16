package pgsql

import (
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"
	"xorm.io/xorm"
	"xorm.io/xorm/contexts"
)

// Config represents PostgreSQL configuration
type Config struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DBName   string `toml:"db_name"`
}

// DSN generates connection string
func (c Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DBName)
}

// Client wraps PostgreSQL client
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
// Usage:
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
// Usage:
//
//	pgsql.Get()              // returns default database
//	pgsql.Get("usercenter")  // returns usercenter database
func Get(name ...string) *Client {
	if len(name) == 0 {
		return defaultClient
	}
	mu.RLock()
	defer mu.RUnlock()
	return clients[name[0]]
}

// Close closes default client and all named clients
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

// SetLogger sets the logger for pgsql.
// Call this after logger is initialized.
func SetLogger(logger Logger) {
	if defaultClient != nil && defaultClient.xormLogger != nil {
		defaultClient.xormLogger.SetLogger(logger)
	}
}

// MustInit initializes and panics on error
func MustInit(cfg Config) {
	if err := Init(cfg); err != nil {
		log.Fatalf("pgsql initialization failed: %v", err)
	}
}

// New creates PostgreSQL client
func New(cfg Config) (*Client, error) {
	engine, err := xorm.NewEngine("postgres", cfg.DSN())
	if err != nil {
		return nil, err
	}

	xormLogger := NewXormLogger(true)
	engine.SetLogger(xormLogger)

	return &Client{engine: engine, xormLogger: xormLogger}, nil
}

// AddHook adds hook for external injection (e.g., tracing)
func (c *Client) AddHook(hook contexts.Hook) {
	c.engine.AddHook(hook)
}

// Engine returns xorm engine
func (c *Client) Engine() *xorm.Engine {
	return c.engine
}

// Sync synchronizes table structure
func (c *Client) Sync(beans ...any) error {
	return c.engine.Sync(beans...)
}

// Close closes connection
func (c *Client) Close() error {
	return c.engine.Close()
}
