package internal

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

// LocalConfig local storage configuration
type LocalConfig struct {
	Root    string // Storage root directory
	BaseURL string // Access URL prefix
}

// Local local file storage
type Local struct {
	root    string
	baseURL string
}

// NewLocal creates local storage
func NewLocal(cfg LocalConfig) (*Local, error) {
	if cfg.Root == "" {
		cfg.Root = "./uploads"
	}

	// Ensure directory exists
	if err := os.MkdirAll(cfg.Root, 0755); err != nil {
		return nil, fmt.Errorf("storage: failed to create directory: %w", err)
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "/uploads"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	log.Printf("storage: local initialized, directory: %s", cfg.Root)

	return &Local{
		root:    cfg.Root,
		baseURL: baseURL,
	}, nil
}

// fullPath returns the full file path
func (l *Local) fullPath(key string) string {
	return filepath.Join(l.root, key)
}

// Put uploads a file
func (l *Local) Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	path := l.fullPath(key)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("storage: failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("storage: failed to create file: %w", err)
	}
	defer file.Close()

	// Write content
	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("storage: failed to write file: %w", err)
	}

	return nil
}

// Get downloads a file
func (l *Local) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	path := l.fullPath(key)

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("storage: file does not exist: %s", key)
		}
		return nil, fmt.Errorf("storage: failed to open file: %w", err)
	}

	return file, nil
}

// Delete deletes a file
func (l *Local) Delete(ctx context.Context, key string) error {
	path := l.fullPath(key)

	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("storage: failed to delete file: %w", err)
	}

	return nil
}

// Exists checks if a file exists
func (l *Local) Exists(ctx context.Context, key string) (bool, error) {
	path := l.fullPath(key)

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("storage: failed to check file: %w", err)
	}

	return true, nil
}

// Info returns file information
func (l *Local) Info(ctx context.Context, key string) (*FileInfo, error) {
	path := l.fullPath(key)

	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("storage: file does not exist: %s", key)
		}
		return nil, fmt.Errorf("storage: failed to get file info: %w", err)
	}

	// Infer MIME type from extension
	contentType := mime.TypeByExtension(filepath.Ext(key))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return &FileInfo{
		Key:          key,
		Size:         stat.Size(),
		ContentType:  contentType,
		LastModified: stat.ModTime().Unix(),
	}, nil
}

// URL returns the file access URL
func (l *Local) URL(key string) string {
	return l.baseURL + "/" + key
}

// GetRaw returns the underlying (local storage returns root directory)
func (l *Local) GetRaw() any {
	return l.root
}
