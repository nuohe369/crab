package internal

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSSConfig Alibaba Cloud OSS configuration
type OSSConfig struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	Bucket          string
	BaseURL         string
}

// OSS Alibaba Cloud OSS storage
type OSS struct {
	client  *oss.Client
	bucket  *oss.Bucket
	baseURL string
}

// NewOSS creates OSS storage
func NewOSS(cfg OSSConfig) (*OSS, error) {
	if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.AccessKeySecret == "" || cfg.Bucket == "" {
		return nil, fmt.Errorf("storage: oss configuration incomplete")
	}

	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("storage: oss client creation failed: %w", err)
	}

	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("storage: oss bucket get failed: %w", err)
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		// Default use bucket domain
		baseURL = fmt.Sprintf("https://%s.%s", cfg.Bucket, cfg.Endpoint)
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	log.Printf("storage: oss initialized, bucket: %s", cfg.Bucket)

	return &OSS{
		client:  client,
		bucket:  bucket,
		baseURL: baseURL,
	}, nil
}

// Put uploads a file
func (o *OSS) Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	options := []oss.Option{}
	if contentType != "" {
		options = append(options, oss.ContentType(contentType))
	}

	err := o.bucket.PutObject(key, reader, options...)
	if err != nil {
		return fmt.Errorf("storage: oss upload failed: %w", err)
	}

	return nil
}

// Get downloads a file
func (o *OSS) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	body, err := o.bucket.GetObject(key)
	if err != nil {
		return nil, fmt.Errorf("storage: oss download failed: %w", err)
	}

	return body, nil
}

// Delete deletes a file
func (o *OSS) Delete(ctx context.Context, key string) error {
	err := o.bucket.DeleteObject(key)
	if err != nil {
		return fmt.Errorf("storage: oss delete failed: %w", err)
	}

	return nil
}

// Exists checks if a file exists
func (o *OSS) Exists(ctx context.Context, key string) (bool, error) {
	exist, err := o.bucket.IsObjectExist(key)
	if err != nil {
		return false, fmt.Errorf("storage: oss check failed: %w", err)
	}

	return exist, nil
}

// Info returns file information
func (o *OSS) Info(ctx context.Context, key string) (*FileInfo, error) {
	meta, err := o.bucket.GetObjectDetailedMeta(key)
	if err != nil {
		return nil, fmt.Errorf("storage: oss get info failed: %w", err)
	}

	var size int64
	if sizeStr := meta.Get("Content-Length"); sizeStr != "" {
		fmt.Sscanf(sizeStr, "%d", &size)
	}

	var lastModified int64
	if lm := meta.Get("Last-Modified"); lm != "" {
		// Parse time (simplified handling)
		// RFC1123 format: Mon, 02 Jan 2006 15:04:05 GMT
	}

	return &FileInfo{
		Key:          key,
		Size:         size,
		ContentType:  meta.Get("Content-Type"),
		LastModified: lastModified,
	}, nil
}

// URL returns the file access URL
func (o *OSS) URL(key string) string {
	return o.baseURL + "/" + key
}

// GetRaw returns the underlying bucket
func (o *OSS) GetRaw() any {
	return o.bucket
}
