package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioConfig struct {
	Endpoint     string
	AccessKey    string
	SecretKey    string
	UseSSL       bool
	Region       string
	Bucket       string
	CreateBucket bool
	PathStyle    bool
}

type ReceiptStore struct {
	client       *minio.Client
	bucket       string
	createBucket bool
}

func NewReceiptStore(ctx context.Context, cfg MinioConfig) (*ReceiptStore, error) {
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
		return nil, fmt.Errorf("init receipt store client: %w", err)
	}

	store := &ReceiptStore{
		client:       client,
		bucket:       cfg.Bucket,
		createBucket: cfg.CreateBucket,
	}

	if err := store.ensureBucket(ctx); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *ReceiptStore) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check receipt bucket %q: %w", s.bucket, err)
	}
	if exists {
		return nil
	}
	if !s.createBucket {
		return fmt.Errorf("receipt bucket %q does not exist", s.bucket)
	}
	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create receipt bucket %q: %w", s.bucket, err)
	}
	return nil
}

func (s *ReceiptStore) PutReceipt(ctx context.Context, key string, data []byte, contentType string) error {
	if key == "" {
		return fmt.Errorf("receipt storage_key is required")
	}
	if len(data) == 0 {
		return fmt.Errorf("receipt data is required")
	}

	opts := minio.PutObjectOptions{ContentType: contentType}
	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(data), int64(len(data)), opts)
	if err != nil {
		return fmt.Errorf("put receipt object: %w", err)
	}
	return nil
}

func (s *ReceiptStore) GetReceipt(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("receipt storage_key is required")
	}

	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get receipt object: %w", err)
	}
	defer obj.Close()

	content, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("read receipt object: %w", err)
	}
	return content, nil
}

func (s *ReceiptStore) DeleteReceipt(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("receipt storage_key is required")
	}
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete receipt object: %w", err)
	}
	return nil
}
