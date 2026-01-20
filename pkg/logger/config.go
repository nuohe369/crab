package logger

import "time"

// Config represents logger configuration
// Config 表示日志器配置
type Config struct {
	BufferSize    int           `toml:"buffer_size"`    // Buffer size in bytes, default 64KB | 缓冲区大小（字节），默认 64KB
	FlushInterval time.Duration `toml:"flush_interval"` // Flush interval, default 3s | 刷新间隔，默认 3 秒
	Enabled       bool          `toml:"enabled"`        // Enable file logging, default true | 启用文件日志，默认 true
}

// DefaultConfig returns default logger configuration
// DefaultConfig 返回默认日志器配置
func DefaultConfig() Config {
	return Config{
		BufferSize:    64 * 1024,      // 64KB
		FlushInterval: 3 * time.Second, // 3 seconds
		Enabled:       true,
	}
}

var globalConfig = DefaultConfig()

// SetConfig sets global logger configuration
// SetConfig 设置全局日志器配置
func SetConfig(cfg Config) {
	// Apply defaults for zero values | 为零值应用默认值
	if cfg.BufferSize <= 0 {
		cfg.BufferSize = 64 * 1024
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = 3 * time.Second
	}
	globalConfig = cfg
}

// GetConfig returns current global logger configuration
// GetConfig 返回当前全局日志器配置
func GetConfig() Config {
	return globalConfig
}
