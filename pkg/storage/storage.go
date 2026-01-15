package storage

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/nuohe369/crab/pkg/storage/internal"
)

// FileInfo represents file information
type FileInfo struct {
	Key          string // File path/key
	Size         int64  // File size (bytes)
	ContentType  string // MIME type
	LastModified int64  // Last modified time (Unix timestamp)
}

// Storage is the storage interface
type Storage interface {
	// Put uploads file
	Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error

	// Get downloads file
	Get(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete deletes file
	Delete(ctx context.Context, key string) error

	// Exists checks if file exists
	Exists(ctx context.Context, key string) (bool, error)

	// Info gets file information
	Info(ctx context.Context, key string) (*FileInfo, error)

	// URL gets file access URL
	// For local storage returns relative path, for OSS/S3 returns full URL
	URL(key string) string

	// GetRaw gets underlying client (for special scenarios)
	GetRaw() any
}

// Config represents storage configuration
type Config struct {
	Driver string `toml:"driver"` // local, oss, s3

	// Local storage configuration
	Local LocalConfig `toml:"local"`

	// Alibaba Cloud OSS configuration
	OSS OSSConfig `toml:"oss"`

	// AWS S3 configuration
	S3 S3Config `toml:"s3"`
}

// LocalConfig represents local storage configuration
type LocalConfig struct {
	Root    string `toml:"root"`     // Storage root directory
	BaseURL string `toml:"base_url"` // Access URL prefix
}

// OSSConfig represents Alibaba Cloud OSS configuration
type OSSConfig struct {
	Endpoint        string `toml:"endpoint"`
	AccessKeyID     string `toml:"access_key_id"`
	AccessKeySecret string `toml:"access_key_secret"`
	Bucket          string `toml:"bucket"`
	BaseURL         string `toml:"base_url"` // CDN or custom domain
}

// S3Config represents AWS S3 configuration
type S3Config struct {
	Region          string `toml:"region"`
	AccessKeyID     string `toml:"access_key_id"`
	SecretAccessKey string `toml:"secret_access_key"`
	Bucket          string `toml:"bucket"`
	Endpoint        string `toml:"endpoint"` // Custom endpoint (for MinIO etc.)
	BaseURL         string `toml:"base_url"` // CDN or custom domain
}

var defaultStorage Storage

// Init initializes default storage
func Init(cfg Config) error {
	if cfg.Driver == "" {
		log.Println("storage: driver not configured, skip initialization")
		return nil
	}

	client, err := New(cfg)
	if err != nil {
		return err
	}
	defaultStorage = client
	return nil
}

// MustInit initializes and panics on error
func MustInit(cfg Config) {
	if err := Init(cfg); err != nil {
		log.Fatalf("storage initialization failed: %v", err)
	}
}

// Get returns default storage instance
func Get() Storage {
	return defaultStorage
}

// New creates storage instance (auto select implementation based on config)
func New(cfg Config) (Storage, error) {
	switch cfg.Driver {
	case "local":
		impl, err := internal.NewLocal(internal.LocalConfig{
			Root:    cfg.Local.Root,
			BaseURL: cfg.Local.BaseURL,
		})
		if err != nil {
			return nil, err
		}
		return &storageWrapper{impl: impl}, nil

	case "oss":
		impl, err := internal.NewOSS(internal.OSSConfig{
			Endpoint:        cfg.OSS.Endpoint,
			AccessKeyID:     cfg.OSS.AccessKeyID,
			AccessKeySecret: cfg.OSS.AccessKeySecret,
			Bucket:          cfg.OSS.Bucket,
			BaseURL:         cfg.OSS.BaseURL,
		})
		if err != nil {
			return nil, err
		}
		return &storageWrapper{impl: impl}, nil

	case "s3":
		impl, err := internal.NewS3(internal.S3Config{
			Region:          cfg.S3.Region,
			AccessKeyID:     cfg.S3.AccessKeyID,
			SecretAccessKey: cfg.S3.SecretAccessKey,
			Bucket:          cfg.S3.Bucket,
			Endpoint:        cfg.S3.Endpoint,
			BaseURL:         cfg.S3.BaseURL,
		})
		if err != nil {
			return nil, err
		}
		return &storageWrapper{impl: impl}, nil

	default:
		return nil, fmt.Errorf("storage: unsupported driver: %s", cfg.Driver)
	}
}

// storageWrapper wraps internal implementation
type storageWrapper struct {
	impl interface {
		Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
		Get(ctx context.Context, key string) (io.ReadCloser, error)
		Delete(ctx context.Context, key string) error
		Exists(ctx context.Context, key string) (bool, error)
		Info(ctx context.Context, key string) (*internal.FileInfo, error)
		URL(key string) string
		GetRaw() any
	}
}

func (w *storageWrapper) Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	return w.impl.Put(ctx, key, reader, size, contentType)
}

func (w *storageWrapper) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return w.impl.Get(ctx, key)
}

func (w *storageWrapper) Delete(ctx context.Context, key string) error {
	return w.impl.Delete(ctx, key)
}

func (w *storageWrapper) Exists(ctx context.Context, key string) (bool, error) {
	return w.impl.Exists(ctx, key)
}

func (w *storageWrapper) Info(ctx context.Context, key string) (*FileInfo, error) {
	info, err := w.impl.Info(ctx, key)
	if err != nil {
		return nil, err
	}
	return &FileInfo{
		Key:          info.Key,
		Size:         info.Size,
		ContentType:  info.ContentType,
		LastModified: info.LastModified,
	}, nil
}

func (w *storageWrapper) URL(key string) string {
	return w.impl.URL(key)
}

func (w *storageWrapper) GetRaw() any {
	return w.impl.GetRaw()
}

// ============ Convenience methods (using default storage) ============

// Enabled checks if storage is enabled
func Enabled() bool {
	return defaultStorage != nil
}

// Put uploads file
func Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	if defaultStorage == nil {
		return fmt.Errorf("storage: not initialized")
	}
	return defaultStorage.Put(ctx, key, reader, size, contentType)
}

// Download downloads file
func Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if defaultStorage == nil {
		return nil, fmt.Errorf("storage: not initialized")
	}
	return defaultStorage.Get(ctx, key)
}

// Delete deletes file
func Delete(ctx context.Context, key string) error {
	if defaultStorage == nil {
		return fmt.Errorf("storage: not initialized")
	}
	return defaultStorage.Delete(ctx, key)
}

// Exists checks if file exists
func Exists(ctx context.Context, key string) (bool, error) {
	if defaultStorage == nil {
		return false, fmt.Errorf("storage: not initialized")
	}
	return defaultStorage.Exists(ctx, key)
}

// Info gets file information
func Info(ctx context.Context, key string) (*FileInfo, error) {
	if defaultStorage == nil {
		return nil, fmt.Errorf("storage: not initialized")
	}
	return defaultStorage.Info(ctx, key)
}

// URL gets file access URL
func URL(key string) string {
	if defaultStorage == nil {
		return ""
	}
	return defaultStorage.URL(key)
}

// GetRaw gets underlying client
func GetRaw() any {
	if defaultStorage == nil {
		return nil
	}
	return defaultStorage.GetRaw()
}
