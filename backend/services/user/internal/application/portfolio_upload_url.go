package application

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type GetPortfolioMediaUploadURLInput struct {
	UserID      uuid.UUID
	FileName    string
	ContentType string
}

type GetPortfolioMediaUploadURLOutput struct {
	StorageKey string
	UploadURL  string
}

type PortfolioMediaUploadStore interface {
	PresignPutObject(ctx context.Context, storageKey string, contentType string, ttl time.Duration) (string, error)
}

type GetPortfolioMediaUploadURL struct {
	Store PortfolioMediaUploadStore
	TTL   time.Duration
}

func (uc *GetPortfolioMediaUploadURL) Execute(ctx context.Context, in GetPortfolioMediaUploadURLInput) (GetPortfolioMediaUploadURLOutput, error) {
	if in.UserID == uuid.Nil {
		return GetPortfolioMediaUploadURLOutput{}, fmt.Errorf("user_id is required")
	}
	if uc.Store == nil {
		return GetPortfolioMediaUploadURLOutput{}, fmt.Errorf("portfolio object store is not configured")
	}
	fileName := strings.TrimSpace(in.FileName)
	if fileName == "" {
		return GetPortfolioMediaUploadURLOutput{}, fmt.Errorf("file_name is required")
	}
	contentType := strings.TrimSpace(strings.ToLower(in.ContentType))
	if contentType == "" {
		return GetPortfolioMediaUploadURLOutput{}, fmt.Errorf("content_type is required")
	}

	storageKey := buildPortfolioMediaStorageKey(in.UserID, fileName)
	ttl := uc.TTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	uploadURL, err := uc.Store.PresignPutObject(ctx, storageKey, contentType, ttl)
	if err != nil {
		return GetPortfolioMediaUploadURLOutput{}, err
	}
	return GetPortfolioMediaUploadURLOutput{StorageKey: storageKey, UploadURL: uploadURL}, nil
}

func buildPortfolioMediaStorageKey(userID uuid.UUID, fileName string) string {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(fileName)))
	return fmt.Sprintf("portfolio/%s/%s%s", userID.String(), uuid.NewString(), ext)
}
