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

// Config represents the application configuration.
type Config struct {
	App      App            `toml:"app"`
	Server   Server         `toml:"server"`
	Database pgsql.Config   `toml:"database"`
	Redis    redis.Config   `toml:"redis"`
	MQ       mq.Config      `toml:"mq"`
	JWT      jwt.Config     `toml:"jwt"`
	Trace    trace.Config   `toml:"trace"`
	Metrics  metrics.Config `toml:"metrics"`
	Storage  storage.Config `toml:"storage"`
	Services []Service      `toml:"services"`
}

// App represents application configuration.
type App struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
	Env     string `toml:"env"` // dev: development, prod: production
}

// IsDev returns true if the environment is development.
func IsDev() bool {
	return cfg != nil && cfg.App.Env == "dev"
}

// IsProd returns true if the environment is production.
func IsProd() bool {
	return cfg != nil && cfg.App.Env == "prod"
}

// Server represents server configuration.
type Server struct {
	Addr string `toml:"addr"`
}

// Service defines a service configuration.
type Service struct {
	Name    string   `toml:"name"`    // service name
	Addr    string   `toml:"addr"`    // listen address
	Modules []string `toml:"modules"` // included modules
}

var cfg *Config

// Load loads configuration from the specified path.
func Load(path string) error {
	cfg = &Config{}
	return config.Load(path, cfg)
}

// MustLoad loads configuration from the specified path, panics on failure.
func MustLoad(path string) {
	cfg = &Config{}
	config.MustLoad(path, cfg)
}

// SetDecryptKey sets the decryption key for encrypted configuration values.
func SetDecryptKey(key string) {
	config.SetDecryptKey(key)
}

// Get returns the global configuration.
func Get() *Config {
	return cfg
}

// GetApp returns the application configuration.
func GetApp() App {
	return cfg.App
}

// GetServer returns the server configuration.
func GetServer() Server {
	return cfg.Server
}

// GetDatabase returns the database configuration.
func GetDatabase() pgsql.Config {
	return cfg.Database
}

// GetRedis returns the Redis configuration.
func GetRedis() redis.Config {
	return cfg.Redis
}

// GetJWT returns the JWT configuration.
func GetJWT() jwt.Config {
	return cfg.JWT
}

// GetTrace returns the tracing configuration.
func GetTrace() trace.Config {
	return cfg.Trace
}

// GetServices returns the list of service configurations.
func GetServices() []Service {
	return cfg.Services
}

// GetMQ returns the message queue configuration.
func GetMQ() mq.Config {
	return cfg.MQ
}

// GetService returns a service configuration by name.
func GetService(name string) *Service {
	for i := range cfg.Services {
		if cfg.Services[i].Name == name {
			return &cfg.Services[i]
		}
	}
	return nil
}

// GetMetrics returns the metrics configuration.
func GetMetrics() metrics.Config {
	return cfg.Metrics
}

// GetStorage returns the storage configuration.
func GetStorage() storage.Config {
	return cfg.Storage
}
