package storage

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"jobconnect/proposal/internal/config"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type AttachmentStore struct {
	client       *minio.Client
	bucket       string
	createBucket bool
}

type ObjectInfo struct {
	SizeBytes   int64
	ContentType string
}

func NewAttachmentStore(ctx context.Context, cfg config.AttachmentStorageConfig) (*AttachmentStore, error) {
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
		return nil, fmt.Errorf("init proposal attachment store client: %w", err)
	}

	store := &AttachmentStore{
		client:       client,
		bucket:       cfg.Bucket,
		createBucket: cfg.CreateBucket,
	}
	if err := store.ensureBucket(ctx); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *AttachmentStore) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check proposal attachment bucket %q: %w", s.bucket, err)
	}
	if exists {
		return nil
	}
	if !s.createBucket {
		return fmt.Errorf("proposal attachment bucket %q does not exist", s.bucket)
	}
	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create proposal attachment bucket %q: %w", s.bucket, err)
	}
	return nil
}

func (s *AttachmentStore) BuildObjectKey(proposalID int64, fileName string) string {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(fileName)))
	if len(ext) > 10 {
		ext = ""
	}
	return path.Join("proposals", fmt.Sprintf("%d", proposalID), "attachments", uuid.NewString()+ext)
}

func (s *AttachmentStore) PresignPutObject(ctx context.Context, storageKey string, ttl time.Duration) (string, error) {
	if strings.TrimSpace(storageKey) == "" {
		return "", fmt.Errorf("proposal attachment storage_key is required")
	}
	if ttl <= 0 {
		return "", fmt.Errorf("proposal attachment presign ttl must be greater than 0")
	}
	u, err := s.client.PresignedPutObject(ctx, s.bucket, storageKey, ttl)
	if err != nil {
		return "", fmt.Errorf("presign proposal attachment upload: %w", err)
	}
	return u.String(), nil
}

func (s *AttachmentStore) PresignGetObject(ctx context.Context, storageKey string, ttl time.Duration) (string, error) {
	if strings.TrimSpace(storageKey) == "" {
		return "", fmt.Errorf("proposal attachment storage_key is required")
	}
	if ttl <= 0 {
		return "", fmt.Errorf("proposal attachment presign ttl must be greater than 0")
	}
	u, err := s.client.PresignedGetObject(ctx, s.bucket, storageKey, ttl, nil)
	if err != nil {
		return "", fmt.Errorf("presign proposal attachment download: %w", err)
	}
	return u.String(), nil
}

func (s *AttachmentStore) StatObject(ctx context.Context, storageKey string) (ObjectInfo, error) {
	if strings.TrimSpace(storageKey) == "" {
		return ObjectInfo{}, fmt.Errorf("proposal attachment storage_key is required")
	}
	obj, err := s.client.StatObject(ctx, s.bucket, storageKey, minio.StatObjectOptions{})
	if err != nil {
		return ObjectInfo{}, fmt.Errorf("stat proposal attachment object: %w", err)
	}
	return ObjectInfo{SizeBytes: obj.Size, ContentType: obj.ContentType}, nil
}
