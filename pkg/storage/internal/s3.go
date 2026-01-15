package internal

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Config AWS S3 configuration
type S3Config struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	Endpoint        string // Custom endpoint (for MinIO, etc.)
	BaseURL         string
}

// S3 AWS S3 storage
type S3Storage struct {
	client  *s3.Client
	bucket  string
	baseURL string
}

// NewS3 creates S3 storage
func NewS3(cfg S3Config) (*S3Storage, error) {
	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.Bucket == "" {
		return nil, fmt.Errorf("storage: s3 configuration incomplete")
	}

	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	// Create credentials
	creds := credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")

	// Load configuration
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("storage: s3 config load failed: %w", err)
	}

	// Create S3 client
	var client *s3.Client
	if cfg.Endpoint != "" {
		// Custom endpoint (MinIO, etc.)
		client = s3.NewFromConfig(awsCfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true // Required for MinIO
		})
	} else {
		client = s3.NewFromConfig(awsCfg)
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		if cfg.Endpoint != "" {
			baseURL = fmt.Sprintf("%s/%s", strings.TrimSuffix(cfg.Endpoint, "/"), cfg.Bucket)
		} else {
			baseURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", cfg.Bucket, cfg.Region)
		}
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	log.Printf("storage: s3 initialized, bucket: %s", cfg.Bucket)

	return &S3Storage{
		client:  client,
		bucket:  cfg.Bucket,
		baseURL: baseURL,
	}, nil
}

// Put uploads a file
func (s *S3Storage) Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   reader,
	}

	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("storage: s3 upload failed: %w", err)
	}

	return nil
}

// Get downloads a file
func (s *S3Storage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("storage: s3 download failed: %w", err)
	}

	return output.Body, nil
}

// Delete deletes a file
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("storage: s3 delete failed: %w", err)
	}

	return nil
}

// Exists checks if a file exists
func (s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		// Check if NotFound error
		return false, nil
	}

	return true, nil
}

// Info returns file information
func (s *S3Storage) Info(ctx context.Context, key string) (*FileInfo, error) {
	output, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("storage: s3 get info failed: %w", err)
	}

	var size int64
	if output.ContentLength != nil {
		size = *output.ContentLength
	}

	var contentType string
	if output.ContentType != nil {
		contentType = *output.ContentType
	}

	var lastModified int64
	if output.LastModified != nil {
		lastModified = output.LastModified.Unix()
	}

	return &FileInfo{
		Key:          key,
		Size:         size,
		ContentType:  contentType,
		LastModified: lastModified,
	}, nil
}

// URL returns the file access URL
func (s *S3Storage) URL(key string) string {
	return s.baseURL + "/" + key
}

// GetRaw returns the underlying client
func (s *S3Storage) GetRaw() any {
	return s.client
}
