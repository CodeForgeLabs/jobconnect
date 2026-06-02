package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/job/internal/domain"

	"github.com/google/uuid"
)

type ListJobAttachments struct {
	Jobs JobRepository
}

type ListJobAttachmentsInput struct {
	JobID    int64
	ClientID uuid.UUID
}

type ListJobAttachmentsOutput struct {
	Attachments []domain.Attachment
}

func (uc *ListJobAttachments) Execute(ctx context.Context, in ListJobAttachmentsInput) (ListJobAttachmentsOutput, error) {
	if in.JobID <= 0 {
		return ListJobAttachmentsOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return ListJobAttachmentsOutput{}, fmt.Errorf("client_id is required")
	}
	attachments, err := uc.Jobs.ListAttachments(ctx, in.JobID, in.ClientID)
	if err != nil {
		return ListJobAttachmentsOutput{}, err
	}
	return ListJobAttachmentsOutput{Attachments: attachments}, nil
}

type GetJobAttachmentDownloadURL struct {
	Jobs JobRepository
}

type GetJobAttachmentDownloadURLInput struct {
	JobID        int64
	AttachmentID int64
	ClientID     uuid.UUID
}

type GetJobAttachmentDownloadURLOutput struct {
	URL string
}

func (uc *GetJobAttachmentDownloadURL) Execute(ctx context.Context, in GetJobAttachmentDownloadURLInput) (GetJobAttachmentDownloadURLOutput, error) {
	if in.JobID <= 0 {
		return GetJobAttachmentDownloadURLOutput{}, fmt.Errorf("job_id is required")
	}
	if in.AttachmentID <= 0 {
		return GetJobAttachmentDownloadURLOutput{}, fmt.Errorf("attachment_id is required")
	}
	if in.ClientID == uuid.Nil {
		return GetJobAttachmentDownloadURLOutput{}, fmt.Errorf("client_id is required")
	}
	att, err := uc.Jobs.GetAttachment(ctx, in.JobID, in.AttachmentID, in.ClientID)
	if err != nil {
		return GetJobAttachmentDownloadURLOutput{}, err
	}
	if strings.TrimSpace(att.URL) == "" {
		return GetJobAttachmentDownloadURLOutput{}, fmt.Errorf("attachment url is empty")
	}
	return GetJobAttachmentDownloadURLOutput{URL: att.URL}, nil
}
