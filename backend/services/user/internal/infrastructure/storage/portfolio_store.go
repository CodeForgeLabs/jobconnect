package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"jobconnect/user/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type PortfolioStore struct {
	client       *minio.Client
	bucket       string
	createBucket bool
}

func NewPortfolioStore(ctx context.Context, cfg config.PortfolioStorageConfig) (*PortfolioStore, error) {
	options := &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	}
	if cfg.PathStyle {
		options.BucketLookup = minio.BucketLookupPath
	}

	client, err := minio.New(cfg.Endpoint, options)
	if err != nil {
		return nil, fmt.Errorf("init portfolio object store client: %w", err)
	}

	store := &PortfolioStore{
		client:       client,
		bucket:       cfg.Bucket,
		createBucket: cfg.CreateBucket,
	}
	if err := store.ensureBucket(ctx); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *PortfolioStore) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check portfolio bucket %q: %w", s.bucket, err)
	}
	if exists {
		return nil
	}
	if !s.createBucket {
		return fmt.Errorf("portfolio bucket %q does not exist", s.bucket)
	}
	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create portfolio bucket %q: %w", s.bucket, err)
	}
	return nil
}

func (s *PortfolioStore) PutObject(ctx context.Context, storageKey string, content []byte, contentType string) error {
	if storageKey == "" {
		return fmt.Errorf("portfolio storage_key is required")
	}
	if len(content) == 0 {
		return fmt.Errorf("portfolio object content is required")
	}

	_, err := s.client.PutObject(ctx, s.bucket, storageKey, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("put portfolio object: %w", err)
	}
	return nil
}

func (s *PortfolioStore) GetObject(ctx context.Context, storageKey string) ([]byte, error) {
	if storageKey == "" {
		return nil, fmt.Errorf("portfolio storage_key is required")
	}

	obj, err := s.client.GetObject(ctx, s.bucket, storageKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get portfolio object: %w", err)
	}
	defer obj.Close()

	content, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("read portfolio object: %w", err)
	}
	return content, nil
}

func (s *PortfolioStore) DeleteObject(ctx context.Context, storageKey string) error {
	if storageKey == "" {
		return fmt.Errorf("portfolio storage_key is required")
	}
	if err := s.client.RemoveObject(ctx, s.bucket, storageKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete portfolio object: %w", err)
	}
	return nil
}

func (s *PortfolioStore) PresignGetObject(ctx context.Context, storageKey string, ttl time.Duration) (string, error) {
	if storageKey == "" {
		return "", fmt.Errorf("portfolio storage_key is required")
	}
	if ttl <= 0 {
		return "", fmt.Errorf("portfolio presign ttl must be greater than 0")
	}

	u, err := s.client.PresignedGetObject(ctx, s.bucket, storageKey, ttl, nil)
	if err != nil {
		return "", fmt.Errorf("presign portfolio object: %w", err)
	}
	return u.String(), nil
}
