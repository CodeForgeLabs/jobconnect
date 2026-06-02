package storage

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"jobconnect/contract/internal/config"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type HourlyEvidenceStore struct {
	client       *minio.Client
	bucket       string
	createBucket bool
}

func NewHourlyEvidenceStore(ctx context.Context, cfg config.HourlyEvidenceStorageConfig) (*HourlyEvidenceStore, error) {
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
		return nil, fmt.Errorf("init hourly evidence store client: %w", err)
	}
	store := &HourlyEvidenceStore{client: client, bucket: cfg.Bucket, createBucket: cfg.CreateBucket}
	if err := store.ensureBucket(ctx); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *HourlyEvidenceStore) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check hourly evidence bucket %q: %w", s.bucket, err)
	}
	if exists {
		return nil
	}
	if !s.createBucket {
		return fmt.Errorf("hourly evidence bucket %q does not exist", s.bucket)
	}
	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create hourly evidence bucket %q: %w", s.bucket, err)
	}
	return nil
}

func (s *HourlyEvidenceStore) BuildObjectKey(contractID int64, fileName string) string {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(fileName)))
	if len(ext) > 10 {
		ext = ""
	}
	return path.Join("contracts", fmt.Sprintf("%d", contractID), "hourly-evidence", uuid.NewString()+ext)
}

func (s *HourlyEvidenceStore) PresignPutObject(ctx context.Context, storageKey string, ttl time.Duration) (string, error) {
	if strings.TrimSpace(storageKey) == "" {
		return "", fmt.Errorf("hourly evidence storage_key is required")
	}
	if ttl <= 0 {
		return "", fmt.Errorf("hourly evidence presign ttl must be greater than 0")
	}
	u, err := s.client.PresignedPutObject(ctx, s.bucket, storageKey, ttl)
	if err != nil {
		return "", fmt.Errorf("presign hourly evidence upload: %w", err)
	}
	return u.String(), nil
}
