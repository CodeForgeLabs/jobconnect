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

type UploadAvatarInput struct {
	UserID      uuid.UUID
	FileName    string
	ContentType string
	Content     []byte
}

type UploadAvatarOutput struct {
	AvatarURL   string
	PreviewURL  string
	ContentType string
	SizeBytes   int64
	Width       int32
	Height      int32
}

type UploadAvatar struct {
	Profiles  ProfileRepository
	Store     AvatarObjectStore
	Processor AvatarProcessor
	Moderator AvatarModerator
	Clock     Clock
}

func (uc *UploadAvatar) Execute(ctx context.Context, in UploadAvatarInput) (UploadAvatarOutput, error) {
	if in.UserID == uuid.Nil {
		return UploadAvatarOutput{}, fmt.Errorf("user_id is required")
	}
	if uc.Processor == nil {
		return UploadAvatarOutput{}, fmt.Errorf("avatar processor is not configured")
	}
	if uc.Store == nil {
		return UploadAvatarOutput{}, fmt.Errorf("avatar object store is not configured")
	}
	if err := domain.ValidateAvatarSize(len(in.Content)); err != nil {
		return UploadAvatarOutput{}, err
	}

	normalized, contentType, width, height, err := uc.Processor.Process(in.Content, in.ContentType)
	if err != nil {
		return UploadAvatarOutput{}, err
	}
	if err := domain.ValidateAvatarContentType(contentType); err != nil {
		return UploadAvatarOutput{}, err
	}
	if uc.Moderator != nil {
		if err := uc.Moderator.Moderate(ctx, normalized, contentType); err != nil {
			return UploadAvatarOutput{}, err
		}
	}

	fileName := sanitizeAvatarFileName(in.FileName, contentType)
	storageKey := buildAvatarStorageKey(in.UserID)
	previousStorageKey := ""
	if prior, err := uc.Profiles.GetAvatar(ctx, in.UserID); err == nil {
		previousStorageKey = strings.TrimSpace(prior.StorageKey)
	}
	if err := uc.Store.PutAvatar(ctx, domain.AvatarObject{
		UserID:      in.UserID,
		StorageKey:  storageKey,
		ContentType: contentType,
		Content:     normalized,
	}); err != nil {
		return UploadAvatarOutput{}, err
	}
	avatarURL := buildAvatarURL(in.UserID)
	updatedAt := time.Now().UTC()
	if uc.Clock != nil {
		updatedAt = uc.Clock.Now()
	}
	avatar := domain.Avatar{
		UserID:      in.UserID,
		FileName:    fileName,
		ContentType: contentType,
		StorageKey:  storageKey,
		Width:       width,
		Height:      height,
		SizeBytes:   int64(len(normalized)),
		UpdatedAt:   updatedAt,
	}

	if err := uc.Profiles.SaveAvatar(ctx, avatar); err != nil {
		return UploadAvatarOutput{}, err
	}

	profile, client, freelancer, err := uc.Profiles.GetByUserID(ctx, in.UserID)
	if err != nil {
		return UploadAvatarOutput{}, err
	}
	profile.AvatarURL = avatarURL
	if err := uc.Profiles.Update(ctx, profile, client, freelancer); err != nil {
		return UploadAvatarOutput{}, err
	}
	if previousStorageKey != "" && previousStorageKey != storageKey {
		if err := uc.Store.DeleteAvatar(ctx, in.UserID, previousStorageKey); err != nil {
			log.Printf("avatar cleanup warning user_id=%s key=%s: %v", in.UserID.String(), previousStorageKey, err)
		}
	}

	return UploadAvatarOutput{
		AvatarURL:   avatarURL,
		PreviewURL:  avatarURL,
		ContentType: contentType,
		SizeBytes:   int64(len(normalized)),
		Width:       int32(width),
		Height:      int32(height),
	}, nil
}

type GetAvatarInput struct {
	UserID uuid.UUID
}

type GetAvatarOutput struct {
	FileName    string
	ContentType string
	Content     []byte
}

type GetAvatar struct {
	Profiles ProfileRepository
	Store    AvatarObjectStore
}

func (uc *GetAvatar) Execute(ctx context.Context, in GetAvatarInput) (GetAvatarOutput, error) {
	if in.UserID == uuid.Nil {
		return GetAvatarOutput{}, fmt.Errorf("user_id is required")
	}
	if uc.Store == nil {
		return GetAvatarOutput{}, fmt.Errorf("avatar object store is not configured")
	}
	avatar, err := uc.Profiles.GetAvatar(ctx, in.UserID)
	if err != nil {
		return GetAvatarOutput{}, err
	}
	content, err := uc.Store.GetAvatar(ctx, in.UserID, avatar.StorageKey)
	if err != nil {
		return GetAvatarOutput{}, err
	}
	return GetAvatarOutput{FileName: avatar.FileName, ContentType: avatar.ContentType, Content: content}, nil
}

type RemoveAvatarInput struct {
	UserID uuid.UUID
}

type RemoveAvatarOutput struct {
	Removed bool
}

type RemoveAvatar struct {
	Profiles ProfileRepository
	Store    AvatarObjectStore
}

func (uc *RemoveAvatar) Execute(ctx context.Context, in RemoveAvatarInput) (RemoveAvatarOutput, error) {
	if in.UserID == uuid.Nil {
		return RemoveAvatarOutput{}, fmt.Errorf("user_id is required")
	}
	if uc.Store == nil {
		return RemoveAvatarOutput{}, fmt.Errorf("avatar object store is not configured")
	}
	avatar, err := uc.Profiles.GetAvatar(ctx, in.UserID)
	if err != nil {
		return RemoveAvatarOutput{}, err
	}
	if err := uc.Store.DeleteAvatar(ctx, in.UserID, avatar.StorageKey); err != nil {
		return RemoveAvatarOutput{}, err
	}
	if err := uc.Profiles.RemoveAvatar(ctx, in.UserID); err != nil {
		return RemoveAvatarOutput{}, err
	}

	profile, client, freelancer, err := uc.Profiles.GetByUserID(ctx, in.UserID)
	if err != nil {
		return RemoveAvatarOutput{}, err
	}
	profile.AvatarURL = ""
	if err := uc.Profiles.Update(ctx, profile, client, freelancer); err != nil {
		return RemoveAvatarOutput{}, err
	}
	return RemoveAvatarOutput{Removed: true}, nil
}

func buildAvatarURL(userID uuid.UUID) string {
	return "/profiles/" + userID.String() + "/avatar"
}

func buildAvatarStorageKey(userID uuid.UUID) string {
	return "avatars/" + userID.String() + "/current"
}

func avatarFileExtension(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ".bin"
	}
}

func sanitizeAvatarFileName(name, contentType string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		trimmed = "avatar"
	}
	ext := strings.ToLower(filepath.Ext(trimmed))
	switch contentType {
	case "image/jpeg":
		if ext != ".jpg" && ext != ".jpeg" {
			trimmed = trimmed + ".jpg"
		}
	case "image/png":
		if ext != ".png" {
			trimmed = trimmed + ".png"
		}
	case "image/webp":
		if ext != ".webp" {
			trimmed = trimmed + ".webp"
		}
	}
	return trimmed
}
