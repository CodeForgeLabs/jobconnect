package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type GetVerificationEvidenceUploadURL struct {
	Store  VerificationEvidenceObjectStore
	PutTTL time.Duration
}

type GetVerificationEvidenceUploadURLInput struct {
	UserID      uuid.UUID
	FileName    string
	ContentType string
}

type GetVerificationEvidenceUploadURLOutput struct {
	StorageKey string
	UploadURL  string
}

func (uc *GetVerificationEvidenceUploadURL) Execute(ctx context.Context, in GetVerificationEvidenceUploadURLInput) (GetVerificationEvidenceUploadURLOutput, error) {
	if uc.Store == nil {
		return GetVerificationEvidenceUploadURLOutput{}, fmt.Errorf("verification evidence upload dependencies are not configured")
	}
	if in.UserID == uuid.Nil {
		return GetVerificationEvidenceUploadURLOutput{}, fmt.Errorf("user_id is required")
	}
	if strings.TrimSpace(in.FileName) == "" {
		return GetVerificationEvidenceUploadURLOutput{}, fmt.Errorf("file_name is required")
	}
	if strings.TrimSpace(in.ContentType) == "" {
		return GetVerificationEvidenceUploadURLOutput{}, fmt.Errorf("content_type is required")
	}
	if uc.PutTTL <= 0 {
		return GetVerificationEvidenceUploadURLOutput{}, fmt.Errorf("invalid verification evidence upload presign ttl")
	}
	storageKey := uc.Store.BuildObjectKey(in.UserID, in.FileName)
	uploadURL, err := uc.Store.PresignPutObject(ctx, storageKey, uc.PutTTL)
	if err != nil {
		return GetVerificationEvidenceUploadURLOutput{}, err
	}
	return GetVerificationEvidenceUploadURLOutput{StorageKey: storageKey, UploadURL: uploadURL}, nil
}

