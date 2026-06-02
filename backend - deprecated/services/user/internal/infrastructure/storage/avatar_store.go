package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"jobconnect/user/internal/application"
	"jobconnect/user/internal/config"
	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type AvatarStore struct {
	client       *minio.Client
	bucket       string
	createBucket bool
}

func NewAvatarStore(ctx context.Context, cfg config.AvatarStorageConfig) (*AvatarStore, error) {
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
		return nil, fmt.Errorf("init avatar object store client: %w", err)
	}

	store := &AvatarStore{
		client:       client,
		bucket:       cfg.Bucket,
		createBucket: cfg.CreateBucket,
	}
	if err := store.ensureBucket(ctx); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *AvatarStore) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check avatar bucket %q: %w", s.bucket, err)
	}
	if exists {
		return nil
	}
	if !s.createBucket {
		return fmt.Errorf("avatar bucket %q does not exist", s.bucket)
	}
	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create avatar bucket %q: %w", s.bucket, err)
	}
	return nil
}

func (s *AvatarStore) PutAvatar(ctx context.Context, avatar domain.AvatarObject) error {
	if avatar.UserID == uuid.Nil {
		return fmt.Errorf("avatar object user_id is required")
	}
	if avatar.StorageKey == "" {
		return fmt.Errorf("avatar object storage_key is required")
	}
	if len(avatar.Content) == 0 {
		return fmt.Errorf("avatar object content is required")
	}

	_, err := s.client.PutObject(ctx, s.bucket, avatar.StorageKey, bytes.NewReader(avatar.Content), int64(len(avatar.Content)), minio.PutObjectOptions{ContentType: avatar.ContentType})
	if err != nil {
		return fmt.Errorf("put avatar object: %w", err)
	}
	return nil
}

func (s *AvatarStore) GetAvatar(ctx context.Context, _ uuid.UUID, storageKey string) ([]byte, error) {
	if storageKey == "" {
		return nil, fmt.Errorf("avatar storage_key is required")
	}

	obj, err := s.client.GetObject(ctx, s.bucket, storageKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get avatar object: %w", err)
	}
	defer obj.Close()

	content, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("read avatar object: %w", err)
	}
	return content, nil
}

func (s *AvatarStore) DeleteAvatar(ctx context.Context, _ uuid.UUID, storageKey string) error {
	if storageKey == "" {
		return fmt.Errorf("avatar storage_key is required")
	}
	if err := s.client.RemoveObject(ctx, s.bucket, storageKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete avatar object: %w", err)
	}
	return nil
}

func (s *AvatarStore) PresignPutObject(ctx context.Context, storageKey string, contentType string, ttl time.Duration) (string, error) {
	if storageKey == "" {
		return "", fmt.Errorf("avatar storage_key is required")
	}
	if ttl <= 0 {
		return "", fmt.Errorf("avatar presign ttl must be greater than 0")
	}
	u, err := s.client.PresignedPutObject(ctx, s.bucket, storageKey, ttl)
	if err != nil {
		return "", fmt.Errorf("presign avatar object: %w", err)
	}
	if contentType == "" {
		return u.String(), nil
	}
	return u.String(), nil
}

func (s *AvatarStore) PresignGetObject(ctx context.Context, storageKey string, ttl time.Duration) (string, error) {
	if storageKey == "" {
		return "", fmt.Errorf("avatar storage_key is required")
	}
	if ttl <= 0 {
		return "", fmt.Errorf("avatar presign ttl must be greater than 0")
	}
	u, err := s.client.PresignedGetObject(ctx, s.bucket, storageKey, ttl, nil)
	if err != nil {
		return "", fmt.Errorf("presign avatar object: %w", err)
	}
	return u.String(), nil
}

func (s *AvatarStore) StatObject(ctx context.Context, storageKey string) (application.ObjectInfo, error) {
	if storageKey == "" {
		return application.ObjectInfo{}, fmt.Errorf("avatar storage_key is required")
	}
	obj, err := s.client.StatObject(ctx, s.bucket, storageKey, minio.StatObjectOptions{})
	if err != nil {
		return application.ObjectInfo{}, fmt.Errorf("stat avatar object: %w", err)
	}
	return application.ObjectInfo{SizeBytes: obj.Size, ContentType: obj.ContentType}, nil
}
