// Package config provides application configuration management
// config 包提供应用程序配置管理
package config

import (
	"github.com/nuohe369/crab/pkg/config"
	"github.com/nuohe369/crab/pkg/jwt"
	"github.com/nuohe369/crab/pkg/metrics"
	"github.com/nuohe369/crab/pkg/mq"
	"github.com/nuohe369/crab/pkg/pgsql"
	"github.com/nuohe369/crab/pkg/redis"
	"github.com/nuohe369/crab/pkg/storage"
	"github.com/nuohe369/crab/pkg/trace"
)

// Config represents the application configuration
// Config 表示应用程序配置
type Config struct {
	App       App                     `toml:"app"`
	Server    Server                  `toml:"server"`
	Snowflake Snowflake               `toml:"snowflake"`
	Database  map[string]pgsql.Config `toml:"database"`
	Redis     map[string]redis.Config `toml:"redis"`
	MQ        mq.Config               `toml:"mq"`
	JWT       jwt.Config              `toml:"jwt"`
	Trace     trace.Config            `toml:"trace"`
	Metrics   metrics.Config          `toml:"metrics"`
	Storage   storage.Config          `toml:"storage"`
	Services  []Service               `toml:"services"`
}

// App represents application configuration
// App 表示应用程序配置
type App struct {
	Name                  string `toml:"name"`                    // Application name | 应用名称
	Version               string `toml:"version"`                 // Application version | 应用版本
	Env                   string `toml:"env"`                     // Environment: dev (development), prod (production) | 环境: dev (开发), prod (生产)
	StrictDependencyCheck bool   `toml:"strict_dependency_check"` // Strict module dependency checking | 严格模块依赖检查
}

// Snowflake represents Snowflake ID generator configuration
// Snowflake 表示雪花 ID 生成器配置
type Snowflake struct {
	MachineID int64 `toml:"machine_id"` // Machine ID (0-1023), default 1 | 机器 ID (0-1023)，默认 1
}

// IsDev returns true if the environment is development
// IsDev 返回环境是否为开发环境
func IsDev() bool {
	return cfg != nil && cfg.App.Env == "dev"
}

// IsProd returns true if the environment is production
// IsProd 返回环境是否为生产环境
func IsProd() bool {
	return cfg != nil && cfg.App.Env == "prod"
}

// IsStrictDependencyCheck returns true if strict dependency checking is enabled
// IsStrictDependencyCheck 返回是否启用严格依赖检查
// Priority: explicit config > production mode > false
// 优先级: 显式配置 > 生产模式 > false
func IsStrictDependencyCheck() bool {
	if cfg == nil {
		return false
	}
	// If explicitly configured, use that value
	// 如果显式配置，使用该值
	if cfg.App.StrictDependencyCheck {
		return true
	}
	// Otherwise, use production mode as default
	// 否则，默认使用生产模式
	return IsProd()
}

// Server represents server configuration
// Server 表示服务器配置
type Server struct {
	Addr string `toml:"addr"` // Listen address | 监听地址
}

// Service defines a service configuration
// Service 定义服务配置
type Service struct {
	Name    string   `toml:"name"`    // Service name | 服务名称
	Addr    string   `toml:"addr"`    // Listen address | 监听地址
	Modules []string `toml:"modules"` // Included modules | 包含的模块
}

var cfg *Config

// Load loads configuration from the specified path
// Load 从指定路径加载配置
func Load(path string) error {
	cfg = &Config{}
	return config.Load(path, cfg)
}

// MustLoad loads configuration from the specified path, panics on failure
// MustLoad 从指定路径加载配置，失败时 panic
func MustLoad(path string) {
	cfg = &Config{}
	config.MustLoad(path, cfg)
}

// SetDecryptKey sets the decryption key for encrypted configuration values
// SetDecryptKey 设置加密配置值的解密密钥
func SetDecryptKey(key string) {
	config.SetDecryptKey(key)
}

// Get returns the global configuration
// Get 返回全局配置
func Get() *Config {
	return cfg
}

// GetApp returns the application configuration
// GetApp 返回应用程序配置
func GetApp() App {
	return cfg.App
}

// GetSnowflake returns the Snowflake configuration
// GetSnowflake 返回雪花 ID 配置
func GetSnowflake() Snowflake {
	return cfg.Snowflake
}

// GetServer returns the server configuration
// GetServer 返回服务器配置
func GetServer() Server {
	return cfg.Server
}

// GetDatabase returns the database configuration (deprecated, use GetDatabases)
// GetDatabase 返回数据库配置（已弃用，请使用 GetDatabases）
func GetDatabase() pgsql.Config {
	// For backward compatibility, return first database if exists | 为了向后兼容，返回第一个数据库（如果存在）
	for _, db := range cfg.Database {
		return db
	}
	return pgsql.Config{}
}

// GetDatabases returns all database configurations
// GetDatabases 返回所有数据库配置
func GetDatabases() map[string]pgsql.Config {
	return cfg.Database
}

// GetRedis returns the Redis configuration (deprecated, use GetRedisInstances)
// GetRedis 返回 Redis 配置（已弃用，请使用 GetRedisInstances）
func GetRedis() redis.Config {
	// For backward compatibility, return first Redis if exists | 为了向后兼容，返回第一个 Redis（如果存在）
	for _, rdb := range cfg.Redis {
		return rdb
	}
	return redis.Config{}
}

// GetRedisInstances returns all Redis configurations
// GetRedisInstances 返回所有 Redis 配置
func GetRedisInstances() map[string]redis.Config {
	return cfg.Redis
}

// GetJWT returns the JWT configuration
// GetJWT 返回 JWT 配置
func GetJWT() jwt.Config {
	return cfg.JWT
}

// GetTrace returns the tracing configuration
// GetTrace 返回追踪配置
func GetTrace() trace.Config {
	return cfg.Trace
}

// GetServices returns the list of service configurations
// GetServices 返回服务配置列表
func GetServices() []Service {
	return cfg.Services
}

// GetMQ returns the message queue configuration
// GetMQ 返回消息队列配置
func GetMQ() mq.Config {
	return cfg.MQ
}

// GetService returns a service configuration by name
// GetService 根据名称返回服务配置
func GetService(name string) *Service {
	for i := range cfg.Services {
		if cfg.Services[i].Name == name {
			return &cfg.Services[i]
		}
	}
	return nil
}

// GetMetrics returns the metrics configuration
// GetMetrics 返回指标配置
func GetMetrics() metrics.Config {
	return cfg.Metrics
}

// GetStorage returns the storage configuration
// GetStorage 返回存储配置
func GetStorage() storage.Config {
	return cfg.Storage
}
