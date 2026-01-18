// Package storage provides unified object storage interface supporting local, OSS, and S3
// Package storage 提供统一的对象存储接口，支持本地、OSS 和 S3
package storage

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/nuohe369/crab/pkg/storage/internal"
)

// FileInfo represents file information
// FileInfo 表示文件信息
type FileInfo struct {
	Key          string // File path/key | 文件路径/键
	Size         int64  // File size (bytes) | 文件大小（字节）
	ContentType  string // MIME type | MIME 类型
	LastModified int64  // Last modified time (Unix timestamp) | 最后修改时间（Unix 时间戳）
}

// Storage is the storage interface
// Storage 是存储接口
type Storage interface {
	// Put uploads file | Put 上传文件
	Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error

	// Get downloads file | Get 下载文件
	Get(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete deletes file | Delete 删除文件
	Delete(ctx context.Context, key string) error

	// Exists checks if file exists | Exists 检查文件是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// Info gets file information | Info 获取文件信息
	Info(ctx context.Context, key string) (*FileInfo, error)

	// URL gets file access URL
	// For local storage returns relative path, for OSS/S3 returns full URL
	// URL 获取文件访问 URL
	// 本地存储返回相对路径，OSS/S3 返回完整 URL
	URL(key string) string

	// GetRaw gets underlying client (for special scenarios)
	// GetRaw 获取底层客户端（用于特殊场景）
	GetRaw() any
}

// Config represents storage configuration
// Config 表示存储配置
type Config struct {
	Driver string      `toml:"driver"` // local, oss, s3 | 本地、OSS、S3
	Local  LocalConfig `toml:"local"`  // Local storage configuration | 本地存储配置
	OSS    OSSConfig   `toml:"oss"`    // Alibaba Cloud OSS configuration | 阿里云 OSS 配置
	S3     S3Config    `toml:"s3"`     // AWS S3 configuration | AWS S3 配置
}

// LocalConfig represents local storage configuration
// LocalConfig 表示本地存储配置
type LocalConfig struct {
	Root    string `toml:"root"`     // Storage root directory | 存储根目录
	BaseURL string `toml:"base_url"` // Access URL prefix | 访问 URL 前缀
}

// OSSConfig represents Alibaba Cloud OSS configuration
// OSSConfig 表示阿里云 OSS 配置
type OSSConfig struct {
	Endpoint        string `toml:"endpoint"`          // OSS endpoint | OSS 端点
	AccessKeyID     string `toml:"access_key_id"`     // Access key ID | 访问密钥 ID
	AccessKeySecret string `toml:"access_key_secret"` // Access key secret | 访问密钥
	Bucket          string `toml:"bucket"`            // Bucket name | 存储桶名称
	BaseURL         string `toml:"base_url"`          // CDN or custom domain | CDN 或自定义域名
}

// S3Config represents AWS S3 configuration
// S3Config 表示 AWS S3 配置
type S3Config struct {
	Region          string `toml:"region"`            // AWS region | AWS 区域
	AccessKeyID     string `toml:"access_key_id"`     // Access key ID | 访问密钥 ID
	SecretAccessKey string `toml:"secret_access_key"` // Secret access key | 密钥
	Bucket          string `toml:"bucket"`            // Bucket name | 存储桶名称
	Endpoint        string `toml:"endpoint"`          // Custom endpoint (for MinIO etc.) | 自定义端点（用于 MinIO 等）
	BaseURL         string `toml:"base_url"`          // CDN or custom domain | CDN 或自定义域名
}

var defaultStorage Storage // Default storage instance | 默认存储实例

// Init initializes default storage
// Init 初始化默认存储
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
// MustInit 初始化，失败时 panic
func MustInit(cfg Config) {
	if err := Init(cfg); err != nil {
		log.Fatalf("storage initialization failed: %v", err)
	}
}

// Get returns default storage instance
// Get 返回默认存储实例
func Get() Storage {
	return defaultStorage
}

// New creates storage instance (auto select implementation based on config)
// New 创建存储实例（根据配置自动选择实现）
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

// ============ Convenience methods (using default storage) | 便捷方法（使用默认存储）============

// Enabled checks if storage is enabled
// Enabled 检查存储是否已启用
func Enabled() bool {
	return defaultStorage != nil
}

// Put uploads file
// Put 上传文件
func Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	if defaultStorage == nil {
		return fmt.Errorf("storage: not initialized")
	}
	return defaultStorage.Put(ctx, key, reader, size, contentType)
}

// Download downloads file
// Download 下载文件
func Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if defaultStorage == nil {
		return nil, fmt.Errorf("storage: not initialized")
	}
	return defaultStorage.Get(ctx, key)
}

// Delete deletes file
// Delete 删除文件
func Delete(ctx context.Context, key string) error {
	if defaultStorage == nil {
		return fmt.Errorf("storage: not initialized")
	}
	return defaultStorage.Delete(ctx, key)
}

// Exists checks if file exists
// Exists 检查文件是否存在
func Exists(ctx context.Context, key string) (bool, error) {
	if defaultStorage == nil {
		return false, fmt.Errorf("storage: not initialized")
	}
	return defaultStorage.Exists(ctx, key)
}

// Info gets file information
// Info 获取文件信息
func Info(ctx context.Context, key string) (*FileInfo, error) {
	if defaultStorage == nil {
		return nil, fmt.Errorf("storage: not initialized")
	}
	return defaultStorage.Info(ctx, key)
}

// URL gets file access URL
// URL 获取文件访问 URL
func URL(key string) string {
	if defaultStorage == nil {
		return ""
	}
	return defaultStorage.URL(key)
}

// GetRaw gets underlying client
// GetRaw 获取底层客户端
func GetRaw() any {
	if defaultStorage == nil {
		return nil
	}
	return defaultStorage.GetRaw()
}
