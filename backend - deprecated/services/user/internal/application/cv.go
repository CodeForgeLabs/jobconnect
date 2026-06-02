package application

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
)

type UpsertCVInput struct {
	UserID      uuid.UUID
	FileName    string
	ContentType string
	StorageKey  string
	Content     []byte
}

type UpsertCVOutput struct {
	CV          CV
	DownloadURL string
}

type GetCVUploadURLInput struct {
	UserID      uuid.UUID
	FileName    string
	ContentType string
}

type GetCVUploadURLOutput struct {
	StorageKey string
	UploadURL  string
}

type GetCVUploadURL struct {
	Store CVObjectStore
	TTL   time.Duration
}

func (uc *GetCVUploadURL) Execute(ctx context.Context, in GetCVUploadURLInput) (GetCVUploadURLOutput, error) {
	if in.UserID == uuid.Nil {
		return GetCVUploadURLOutput{}, fmt.Errorf("user_id is required")
	}
	if uc.Store == nil {
		return GetCVUploadURLOutput{}, fmt.Errorf("cv object store is not configured")
	}
	if strings.TrimSpace(in.FileName) == "" {
		return GetCVUploadURLOutput{}, fmt.Errorf("file_name is required")
	}
	if err := domain.ValidateCVContentType(strings.TrimSpace(in.ContentType)); err != nil {
		return GetCVUploadURLOutput{}, err
	}
	storageKey := buildCVStorageKey(in.UserID)
	ttl := uc.TTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	uploadURL, err := uc.Store.PresignPutObject(ctx, storageKey, strings.TrimSpace(in.ContentType), ttl)
	if err != nil {
		return GetCVUploadURLOutput{}, err
	}
	return GetCVUploadURLOutput{StorageKey: storageKey, UploadURL: uploadURL}, nil
}

type CVRepository interface {
	SaveCV(ctx context.Context, cv CV) error
	GetCV(ctx context.Context, userID uuid.UUID) (CV, error)
	RemoveCV(ctx context.Context, userID uuid.UUID) error
}

type CVRoleRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (domain.Profile, *domain.ClientProfile, *domain.FreelancerProfile, error)
}

type UpsertCV struct {
	Profiles     CVRepository
	RoleProfiles CVRoleRepository
	Store        CVObjectStore
	Clock        Clock
}

func (uc *UpsertCV) Execute(ctx context.Context, in UpsertCVInput) (UpsertCVOutput, error) {
	if in.UserID == uuid.Nil {
		return UpsertCVOutput{}, fmt.Errorf("user_id is required")
	}
	if strings.TrimSpace(in.FileName) == "" {
		return UpsertCVOutput{}, fmt.Errorf("file_name is required")
	}
	if uc.Store == nil {
		return UpsertCVOutput{}, fmt.Errorf("cv object store is not configured")
	}
	if err := requireFreelancerProfile(ctx, in.UserID, uc.RoleProfiles); err != nil {
		return UpsertCVOutput{}, err
	}

	storageKey := buildCVStorageKey(in.UserID)
	fileName := sanitizeCVFileName(in.FileName, in.ContentType)
	previousStorageKey := ""
	if prior, err := uc.Profiles.GetCV(ctx, in.UserID); err == nil {
		previousStorageKey = strings.TrimSpace(prior.StorageKey)
	} else if !isNotFoundError(err) {
		return UpsertCVOutput{}, err
	}

	if strings.TrimSpace(in.StorageKey) != "" {
		if err := domain.ValidateCVContentType(in.ContentType); err != nil {
			return UpsertCVOutput{}, err
		}
		info, err := uc.Store.StatObject(ctx, in.StorageKey)
		if err != nil {
			return UpsertCVOutput{}, err
		}
		updatedAt := time.Now().UTC()
		if uc.Clock != nil {
			updatedAt = uc.Clock.Now()
		}
		cv := CV{
			UserID:      in.UserID,
			FileName:    fileName,
			ContentType: strings.TrimSpace(in.ContentType),
			StorageKey:  in.StorageKey,
			SizeBytes:   info.SizeBytes,
			UpdatedAt:   updatedAt,
		}
		if err := uc.Profiles.SaveCV(ctx, cv); err != nil {
			return UpsertCVOutput{}, err
		}
		downloadURL, err := uc.Store.PresignGetObject(ctx, in.StorageKey, 15*time.Minute)
		if err != nil {
			return UpsertCVOutput{}, err
		}
		if previousStorageKey != "" && previousStorageKey != storageKey {
			if err := uc.Store.DeleteCV(ctx, in.UserID, previousStorageKey); err != nil {
				log.Printf("cv cleanup warning user_id=%s key=%s: %v", in.UserID.String(), previousStorageKey, err)
			}
		}
		return UpsertCVOutput{CV: cv, DownloadURL: downloadURL}, nil
	}

	if len(in.Content) == 0 {
		return UpsertCVOutput{}, fmt.Errorf("content is required")
	}
	if err := domain.ValidateCVSize(len(in.Content)); err != nil {
		return UpsertCVOutput{}, err
	}
	if err := domain.ValidateCVContentType(in.ContentType); err != nil {
		return UpsertCVOutput{}, err
	}

	if err := uc.Store.PutCV(ctx, domain.CVObject{
		UserID:      in.UserID,
		StorageKey:  storageKey,
		ContentType: strings.TrimSpace(in.ContentType),
		Content:     in.Content,
	}); err != nil {
		return UpsertCVOutput{}, err
	}

	updatedAt := time.Now().UTC()
	if uc.Clock != nil {
		updatedAt = uc.Clock.Now()
	}
	cv := CV{
		UserID:      in.UserID,
		FileName:    fileName,
		ContentType: strings.TrimSpace(in.ContentType),
		StorageKey:  storageKey,
		SizeBytes:   int64(len(in.Content)),
		UpdatedAt:   updatedAt,
	}
	if err := uc.Profiles.SaveCV(ctx, cv); err != nil {
		_ = uc.Store.DeleteCV(ctx, in.UserID, storageKey)
		return UpsertCVOutput{}, err
	}

	if previousStorageKey != "" && previousStorageKey != storageKey {
		if err := uc.Store.DeleteCV(ctx, in.UserID, previousStorageKey); err != nil {
			log.Printf("cv cleanup warning user_id=%s key=%s: %v", in.UserID.String(), previousStorageKey, err)
		}
	}

	downloadURL, err := uc.Store.PresignGetObject(ctx, storageKey, 15*time.Minute)
	if err != nil {
		return UpsertCVOutput{}, err
	}
	return UpsertCVOutput{CV: cv, DownloadURL: downloadURL}, nil
}

type GetCVInput struct {
	UserID uuid.UUID
}

type GetCVOutput struct {
	CV          CV
	DownloadURL string
}

type GetCV struct {
	Profiles     CVRepository
	RoleProfiles CVRoleRepository
	Store        CVObjectStore
}

func (uc *GetCV) Execute(ctx context.Context, in GetCVInput) (GetCVOutput, error) {
	if in.UserID == uuid.Nil {
		return GetCVOutput{}, fmt.Errorf("user_id is required")
	}
	if uc.Store == nil {
		return GetCVOutput{}, fmt.Errorf("cv object store is not configured")
	}
	if err := requireFreelancerProfile(ctx, in.UserID, uc.RoleProfiles); err != nil {
		return GetCVOutput{}, err
	}
	cv, err := uc.Profiles.GetCV(ctx, in.UserID)
	if err != nil {
		return GetCVOutput{}, err
	}
	if strings.TrimSpace(cv.StorageKey) == "" {
		return GetCVOutput{}, fmt.Errorf("cv storage_key is required")
	}
	info, err := uc.Store.StatObject(ctx, cv.StorageKey)
	if err != nil {
		return GetCVOutput{}, err
	}
	cv.SizeBytes = info.SizeBytes
	downloadURL, err := uc.Store.PresignGetObject(ctx, cv.StorageKey, 15*time.Minute)
	if err != nil {
		return GetCVOutput{}, err
	}
	return GetCVOutput{CV: cv, DownloadURL: downloadURL}, nil
}

type RemoveCVInput struct {
	UserID uuid.UUID
}

type RemoveCVOutput struct {
	Removed bool
}

type RemoveCV struct {
	Profiles     CVRepository
	RoleProfiles CVRoleRepository
	Store        CVObjectStore
}

func (uc *RemoveCV) Execute(ctx context.Context, in RemoveCVInput) (RemoveCVOutput, error) {
	if in.UserID == uuid.Nil {
		return RemoveCVOutput{}, fmt.Errorf("user_id is required")
	}
	if uc.Store == nil {
		return RemoveCVOutput{}, fmt.Errorf("cv object store is not configured")
	}
	if err := requireFreelancerProfile(ctx, in.UserID, uc.RoleProfiles); err != nil {
		return RemoveCVOutput{}, err
	}
	cv, err := uc.Profiles.GetCV(ctx, in.UserID)
	if err != nil {
		return RemoveCVOutput{}, err
	}
	if err := uc.Store.DeleteCV(ctx, in.UserID, cv.StorageKey); err != nil {
		return RemoveCVOutput{}, err
	}
	if err := uc.Profiles.RemoveCV(ctx, in.UserID); err != nil {
		return RemoveCVOutput{}, err
	}
	return RemoveCVOutput{Removed: true}, nil
}

func buildCVStorageKey(userID uuid.UUID) string {
	return "cvs/" + userID.String() + "/current"
}

func sanitizeCVFileName(name, contentType string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		trimmed = "cv"
	}
	ext := strings.ToLower(filepath.Ext(trimmed))
	switch strings.TrimSpace(strings.ToLower(contentType)) {
	case "application/pdf":
		if ext != ".pdf" {
			trimmed += ".pdf"
		}
	case "application/msword":
		if ext != ".doc" {
			trimmed += ".doc"
		}
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		if ext != ".docx" {
			trimmed += ".docx"
		}
	}
	return trimmed
}

func isNotFoundError(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "not found")
}

func requireFreelancerProfile(ctx context.Context, userID uuid.UUID, repo CVRoleRepository) error {
	if repo == nil {
		return fmt.Errorf("profile repository is not configured")
	}
	p, _, _, err := repo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if strings.ToLower(strings.TrimSpace(p.Role)) != domain.RoleFreelancer {
		return fmt.Errorf("freelancer role required")
	}
	return nil
}
