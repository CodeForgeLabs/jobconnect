package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type DeleteJobAttachment struct {
	Jobs    JobRepository
	Storage AttachmentObjectStore
}

type DeleteJobAttachmentInput struct {
	JobID        int64
	AttachmentID int64
	ClientID     uuid.UUID
}

type DeleteJobAttachmentOutput struct {
	Deleted bool
}

func (uc *DeleteJobAttachment) Execute(ctx context.Context, in DeleteJobAttachmentInput) (DeleteJobAttachmentOutput, error) {
	if in.JobID <= 0 {
		return DeleteJobAttachmentOutput{}, fmt.Errorf("job_id is required")
	}
	if in.AttachmentID <= 0 {
		return DeleteJobAttachmentOutput{}, fmt.Errorf("attachment_id is required")
	}
	if in.ClientID == uuid.Nil {
		return DeleteJobAttachmentOutput{}, fmt.Errorf("client_id is required")
	}

	attachment, err := uc.Jobs.DeleteAttachment(ctx, in.JobID, in.AttachmentID, in.ClientID)
	if err != nil {
		return DeleteJobAttachmentOutput{}, err
	}

	if attachment.StorageKey != "" {
		_ = uc.Storage.DeleteObject(ctx, attachment.StorageKey)
	}
	return DeleteJobAttachmentOutput{Deleted: true}, nil
}
