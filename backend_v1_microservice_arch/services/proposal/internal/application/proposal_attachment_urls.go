package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
)

type GetProposalAttachmentUploadURL struct {
	Proposals ProposalRepository
	Store     AttachmentObjectStore
	PutTTL    time.Duration
}

type GetProposalAttachmentUploadURLInput struct {
	FreelancerID uuid.UUID
	ProposalID   int64
	FileName     string
	ContentType  string
}

type GetProposalAttachmentUploadURLOutput struct {
	StorageKey string
	UploadURL  string
}

func (uc *GetProposalAttachmentUploadURL) Execute(ctx context.Context, in GetProposalAttachmentUploadURLInput) (GetProposalAttachmentUploadURLOutput, error) {
	if in.FreelancerID == uuid.Nil {
		return GetProposalAttachmentUploadURLOutput{}, fmt.Errorf("freelancer_id is required")
	}
	if in.ProposalID <= 0 {
		return GetProposalAttachmentUploadURLOutput{}, fmt.Errorf("proposal_id is required")
	}
	if strings.TrimSpace(in.FileName) == "" {
		return GetProposalAttachmentUploadURLOutput{}, fmt.Errorf("file_name is required")
	}
	if strings.TrimSpace(in.ContentType) == "" {
		return GetProposalAttachmentUploadURLOutput{}, fmt.Errorf("content_type is required")
	}
	if uc.PutTTL <= 0 {
		return GetProposalAttachmentUploadURLOutput{}, fmt.Errorf("invalid upload presign ttl")
	}

	if _, err := uc.Proposals.GetByIDForFreelancer(ctx, in.ProposalID, in.FreelancerID); err != nil {
		return GetProposalAttachmentUploadURLOutput{}, err
	}

	storageKey := uc.Store.BuildObjectKey(in.ProposalID, in.FileName)
	uploadURL, err := uc.Store.PresignPutObject(ctx, storageKey, uc.PutTTL)
	if err != nil {
		return GetProposalAttachmentUploadURLOutput{}, err
	}

	return GetProposalAttachmentUploadURLOutput{StorageKey: storageKey, UploadURL: uploadURL}, nil
}

type GetProposalAttachmentDownloadURL struct {
	Proposals ProposalRepository
	Store     AttachmentObjectStore
	GetTTL    time.Duration
}

type GetProposalAttachmentDownloadURLInput struct {
	ProposalID   int64
	AttachmentID int64
	ActorID      uuid.UUID
	ActorRole    string
}

type GetProposalAttachmentDownloadURLOutput struct {
	DownloadURL string
}

func (uc *GetProposalAttachmentDownloadURL) Execute(ctx context.Context, in GetProposalAttachmentDownloadURLInput) (GetProposalAttachmentDownloadURLOutput, error) {
	if in.ProposalID <= 0 {
		return GetProposalAttachmentDownloadURLOutput{}, fmt.Errorf("proposal_id is required")
	}
	if in.AttachmentID <= 0 {
		return GetProposalAttachmentDownloadURLOutput{}, fmt.Errorf("attachment_id is required")
	}
	if in.ActorID == uuid.Nil {
		return GetProposalAttachmentDownloadURLOutput{}, fmt.Errorf("actor_id is required")
	}
	if uc.GetTTL <= 0 {
		return GetProposalAttachmentDownloadURLOutput{}, fmt.Errorf("invalid download presign ttl")
	}

	var (
		proposal domain.Proposal
		err      error
	)
	switch strings.ToLower(strings.TrimSpace(in.ActorRole)) {
	case "client":
		proposal, err = uc.Proposals.GetByIDForClient(ctx, in.ProposalID, in.ActorID)
	case "freelancer":
		proposal, err = uc.Proposals.GetByIDForFreelancer(ctx, in.ProposalID, in.ActorID)
	default:
		return GetProposalAttachmentDownloadURLOutput{}, fmt.Errorf("unsupported actor role")
	}
	if err != nil {
		return GetProposalAttachmentDownloadURLOutput{}, err
	}

	for _, a := range proposal.Attachments {
		if a.ID != in.AttachmentID {
			continue
		}
		if strings.TrimSpace(a.StorageKey) == "" {
			return GetProposalAttachmentDownloadURLOutput{}, fmt.Errorf("attachment does not have storage key")
		}
		downloadURL, err := uc.Store.PresignGetObject(ctx, a.StorageKey, uc.GetTTL)
		if err != nil {
			return GetProposalAttachmentDownloadURLOutput{}, err
		}
		return GetProposalAttachmentDownloadURLOutput{DownloadURL: downloadURL}, nil
	}

	return GetProposalAttachmentDownloadURLOutput{}, fmt.Errorf("attachment not found")
}
