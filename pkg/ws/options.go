package ws

import "time"

// Default configuration values
const (
	defaultReadTimeout    = 60 * time.Second // Read timeout (disconnect if no message)
	defaultWriteTimeout   = 10 * time.Second // Write timeout
	defaultPingInterval   = 30 * time.Second // Ping interval
	defaultMaxMessageSize = 512 * 1024       // Max message size 512KB
	defaultSendBuffer     = 256              // Send buffer size
)

// Options represents Hub configuration options
type Options struct {
	// ReadTimeout is the read timeout duration.
	// Connection will be closed if no message (including ping) is received within this time.
	// Default: 60 seconds
	ReadTimeout time.Duration

	// WriteTimeout is the write timeout duration for sending messages.
	// Default: 10 seconds
	WriteTimeout time.Duration

	// PingInterval is the interval for server to send ping messages.
	// Default: 30 seconds
	PingInterval time.Duration

	// MaxMessageSize is the maximum message size in bytes.
	// Messages exceeding this size will be rejected.
	// Default: 512KB
	MaxMessageSize int64

	// SendBuffer is the capacity of Client.Send channel.
	// Default: 256
	SendBuffer int
}

// Option is a function type for configuring Options
type Option func(*Options)

// defaultOptions returns default configuration
func defaultOptions() *Options {
	return &Options{
		ReadTimeout:    defaultReadTimeout,
		WriteTimeout:   defaultWriteTimeout,
		PingInterval:   defaultPingInterval,
		MaxMessageSize: defaultMaxMessageSize,
		SendBuffer:     defaultSendBuffer,
	}
}

// WithReadTimeout sets read timeout
func WithReadTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.ReadTimeout = d
	}
}

// WithWriteTimeout sets write timeout
func WithWriteTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.WriteTimeout = d
	}
}

// WithPingInterval sets ping interval
func WithPingInterval(d time.Duration) Option {
	return func(o *Options) {
		o.PingInterval = d
	}
}

// WithMaxMessageSize sets maximum message size
func WithMaxMessageSize(size int64) Option {
	return func(o *Options) {
		o.MaxMessageSize = size
	}
}

// WithSendBuffer sets send buffer size
func WithSendBuffer(size int) Option {
	return func(o *Options) {
		o.SendBuffer = size
	}
}
