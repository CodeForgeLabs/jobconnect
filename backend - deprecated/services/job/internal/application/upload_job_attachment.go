package application

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/job/internal/domain"

	"github.com/google/uuid"
)

const maxAttachmentBytes = 25 * 1024 * 1024

type UploadJobAttachment struct {
	Jobs    JobRepository
	Storage AttachmentObjectStore
}

type UploadJobAttachmentInput struct {
	JobID        int64
	ClientID     uuid.UUID
	FileName     string
	ContentType  string
	ContentBytes []byte
}

type UploadJobAttachmentOutput struct {
	Attachment domain.Attachment
}

func (uc *UploadJobAttachment) Execute(ctx context.Context, in UploadJobAttachmentInput) (UploadJobAttachmentOutput, error) {
	if in.JobID <= 0 {
		return UploadJobAttachmentOutput{}, fmt.Errorf("job_id is required")
	}
	if in.ClientID == uuid.Nil {
		return UploadJobAttachmentOutput{}, fmt.Errorf("client_id is required")
	}
	if strings.TrimSpace(in.FileName) == "" {
		return UploadJobAttachmentOutput{}, fmt.Errorf("file_name is required")
	}
	if len(in.ContentBytes) == 0 {
		return UploadJobAttachmentOutput{}, fmt.Errorf("content is required")
	}
	if len(in.ContentBytes) > maxAttachmentBytes {
		return UploadJobAttachmentOutput{}, fmt.Errorf("attachment content exceeds 25MB limit")
	}

	job, err := uc.Jobs.GetByIDForClient(ctx, in.JobID, in.ClientID)
	if err != nil {
		return UploadJobAttachmentOutput{}, err
	}
	if job.Status != domain.JobStatusOpen {
		return UploadJobAttachmentOutput{}, fmt.Errorf("only open jobs can receive attachments")
	}

	objectKey := uc.Storage.BuildObjectKey(in.JobID, in.FileName)
	publicURL, err := uc.Storage.PutObject(ctx, objectKey, in.ContentBytes, in.ContentType)
	if err != nil {
		return UploadJobAttachmentOutput{}, err
	}

	attachment, err := uc.Jobs.AddAttachment(ctx, in.JobID, in.ClientID, domain.Attachment{
		FileName:    strings.TrimSpace(in.FileName),
		ContentType: strings.TrimSpace(in.ContentType),
		StorageKey:  objectKey,
		URL:         publicURL,
		SizeBytes:   int64(len(in.ContentBytes)),
	})
	if err != nil {
		_ = uc.Storage.DeleteObject(ctx, objectKey)
		return UploadJobAttachmentOutput{}, err
	}

	return UploadJobAttachmentOutput{Attachment: attachment}, nil
}
