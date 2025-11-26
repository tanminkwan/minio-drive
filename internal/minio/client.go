package minio

import (
	"context"
	"fmt"
	"path/filepath"
	"simple-uploader/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client *minio.Client
	bucket string
}

// NewClient creates a new MinIO client from config
func NewClient(cfg *config.MinIOConfig) (*Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &Client{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// UploadFile uploads a single file to the bucket root
func (c *Client) UploadFile(ctx context.Context, filePath string) error {
	// Use only the filename, upload to bucket root
	objectName := filepath.Base(filePath)

	_, err := c.client.FPutObject(ctx, c.bucket, objectName, filePath, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to upload %s: %w", filePath, err)
	}

	return nil
}

// UploadFiles uploads multiple files to the bucket root
func (c *Client) UploadFiles(ctx context.Context, filePaths []string) (successes []string, failures map[string]error) {
	failures = make(map[string]error)

	for _, path := range filePaths {
		if err := c.UploadFile(ctx, path); err != nil {
			failures[path] = err
		} else {
			successes = append(successes, path)
		}
	}

	return successes, failures
}

// EnsureBucket checks if bucket exists, creates if not
func (c *Client) EnsureBucket(ctx context.Context) error {
	exists, err := c.client.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		err = c.client.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}
