package storage

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"jobconnect/verification/internal/config"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type EvidenceStore struct {
	client       *minio.Client
	bucket       string
	createBucket bool
}

func NewEvidenceStore(ctx context.Context, cfg config.VerificationEvidenceStorageConfig) (*EvidenceStore, error) {
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
		return nil, fmt.Errorf("init verification evidence store client: %w", err)
	}
	store := &EvidenceStore{client: client, bucket: cfg.Bucket, createBucket: cfg.CreateBucket}
	if err := store.ensureBucket(ctx); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *EvidenceStore) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check verification evidence bucket %q: %w", s.bucket, err)
	}
	if exists {
		return nil
	}
	if !s.createBucket {
		return fmt.Errorf("verification evidence bucket %q does not exist", s.bucket)
	}
	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create verification evidence bucket %q: %w", s.bucket, err)
	}
	return nil
}

func (s *EvidenceStore) BuildObjectKey(userID uuid.UUID, fileName string) string {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(fileName)))
	if len(ext) > 10 {
		ext = ""
	}
	return path.Join("verifications", userID.String(), uuid.NewString()+ext)
}

func (s *EvidenceStore) PresignPutObject(ctx context.Context, storageKey string, ttl time.Duration) (string, error) {
	if strings.TrimSpace(storageKey) == "" {
		return "", fmt.Errorf("verification evidence storage_key is required")
	}
	if ttl <= 0 {
		return "", fmt.Errorf("verification evidence presign ttl must be greater than 0")
	}
	u, err := s.client.PresignedPutObject(ctx, s.bucket, storageKey, ttl)
	if err != nil {
		return "", fmt.Errorf("presign verification evidence upload: %w", err)
	}
	return u.String(), nil
}

