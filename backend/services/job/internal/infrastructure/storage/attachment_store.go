package storage

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"jobconnect/job/internal/config"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type AttachmentStore struct {
	client       *minio.Client
	bucket       string
	endpoint     string
	useSSL       bool
	createBucket bool
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
		return nil, fmt.Errorf("init attachment object store client: %w", err)
	}

	store := &AttachmentStore{
		client:       client,
		bucket:       cfg.Bucket,
		endpoint:     cfg.Endpoint,
		useSSL:       cfg.UseSSL,
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
		return fmt.Errorf("check attachment bucket %q: %w", s.bucket, err)
	}
	if exists {
		return nil
	}
	if !s.createBucket {
		return fmt.Errorf("attachment bucket %q does not exist", s.bucket)
	}
	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create attachment bucket %q: %w", s.bucket, err)
	}
	return nil
}

func (s *AttachmentStore) BuildObjectKey(jobID int64, fileName string) string {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(fileName)))
	if len(ext) > 10 {
		ext = ""
	}
	return path.Join("jobs", fmt.Sprintf("%d", jobID), "attachments", uuid.NewString()+ext)
}

func (s *AttachmentStore) PutObject(ctx context.Context, objectKey string, content []byte, contentType string) (string, error) {
	if strings.TrimSpace(objectKey) == "" {
		return "", fmt.Errorf("attachment object_key is required")
	}
	if len(content) == 0 {
		return "", fmt.Errorf("attachment content is required")
	}
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/octet-stream"
	}

	_, err := s.client.PutObject(ctx, s.bucket, objectKey, bytes.NewReader(content), int64(len(content)), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("put attachment object: %w", err)
	}

	return s.PublicURL(objectKey), nil
}

func (s *AttachmentStore) DeleteObject(ctx context.Context, objectKey string) error {
	if strings.TrimSpace(objectKey) == "" {
		return fmt.Errorf("attachment object_key is required")
	}
	if err := s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete attachment object: %w", err)
	}
	return nil
}

func (s *AttachmentStore) PublicURL(objectKey string) string {
	scheme := "http"
	if s.useSSL {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/%s/%s", scheme, s.endpoint, s.bucket, strings.TrimPrefix(objectKey, "/"))
}
