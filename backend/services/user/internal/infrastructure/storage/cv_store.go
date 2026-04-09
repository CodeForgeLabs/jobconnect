package storage

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"jobconnect/user/internal/config"
	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type CVStore struct {
	client       *minio.Client
	bucket       string
	createBucket bool
}

func NewCVStore(ctx context.Context, cfg config.CVStorageConfig) (*CVStore, error) {
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
		return nil, fmt.Errorf("init cv object store client: %w", err)
	}

	store := &CVStore{client: client, bucket: cfg.Bucket, createBucket: cfg.CreateBucket}
	if err := store.ensureBucket(ctx); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *CVStore) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check cv bucket %q: %w", s.bucket, err)
	}
	if exists {
		return nil
	}
	if !s.createBucket {
		return fmt.Errorf("cv bucket %q does not exist", s.bucket)
	}
	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create cv bucket %q: %w", s.bucket, err)
	}
	return nil
}

func (s *CVStore) PutCV(ctx context.Context, cv domain.CVObject) error {
	if cv.UserID == uuid.Nil {
		return fmt.Errorf("cv object user_id is required")
	}
	if cv.StorageKey == "" {
		return fmt.Errorf("cv storage_key is required")
	}
	if len(cv.Content) == 0 {
		return fmt.Errorf("cv object content is required")
	}
	_, err := s.client.PutObject(ctx, s.bucket, cv.StorageKey, bytes.NewReader(cv.Content), int64(len(cv.Content)), minio.PutObjectOptions{ContentType: cv.ContentType})
	if err != nil {
		return fmt.Errorf("put cv object: %w", err)
	}
	return nil
}

func (s *CVStore) DeleteCV(ctx context.Context, _ uuid.UUID, storageKey string) error {
	if storageKey == "" {
		return fmt.Errorf("cv storage_key is required")
	}
	if err := s.client.RemoveObject(ctx, s.bucket, storageKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete cv object: %w", err)
	}
	return nil
}

func (s *CVStore) PresignGetObject(ctx context.Context, storageKey string, ttl time.Duration) (string, error) {
	if storageKey == "" {
		return "", fmt.Errorf("cv storage_key is required")
	}
	if ttl <= 0 {
		return "", fmt.Errorf("cv presign ttl must be greater than 0")
	}
	u, err := s.client.PresignedGetObject(ctx, s.bucket, storageKey, ttl, nil)
	if err != nil {
		return "", fmt.Errorf("presign cv object: %w", err)
	}
	return u.String(), nil
}
