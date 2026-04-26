package application

import (
	"context"
	"fmt"
	"jobconnect/contract/internal/domain"
	"strings"
	"time"

	"github.com/google/uuid"
)

type GetHourlyLogEvidenceUploadURL struct {
	Contracts ContractRepository
	Store     HourlyEvidenceObjectStore
	PutTTL    time.Duration
}

type GetHourlyLogEvidenceUploadURLInput struct {
	ContractID   int64
	FreelancerID uuid.UUID
	FileName     string
	ContentType  string
}

type GetHourlyLogEvidenceUploadURLOutput struct {
	StorageKey string
	UploadURL  string
}

func (uc *GetHourlyLogEvidenceUploadURL) Execute(ctx context.Context, in GetHourlyLogEvidenceUploadURLInput) (GetHourlyLogEvidenceUploadURLOutput, error) {
	if uc.Contracts == nil || uc.Store == nil {
		return GetHourlyLogEvidenceUploadURLOutput{}, fmt.Errorf("hourly evidence upload dependencies are not configured")
	}
	if in.ContractID <= 0 {
		return GetHourlyLogEvidenceUploadURLOutput{}, fmt.Errorf("contract_id is required")
	}
	if in.FreelancerID == uuid.Nil {
		return GetHourlyLogEvidenceUploadURLOutput{}, fmt.Errorf("freelancer_id is required")
	}
	if strings.TrimSpace(in.FileName) == "" {
		return GetHourlyLogEvidenceUploadURLOutput{}, fmt.Errorf("file_name is required")
	}
	if strings.TrimSpace(in.ContentType) == "" {
		return GetHourlyLogEvidenceUploadURLOutput{}, fmt.Errorf("content_type is required")
	}
	if uc.PutTTL <= 0 {
		return GetHourlyLogEvidenceUploadURLOutput{}, fmt.Errorf("invalid hourly evidence upload presign ttl")
	}
	contract, err := uc.Contracts.GetByIDForActor(ctx, in.ContractID, in.FreelancerID)
	if err != nil {
		return GetHourlyLogEvidenceUploadURLOutput{}, err
	}
	if contract.ContractType != domain.TypeHourly {
		return GetHourlyLogEvidenceUploadURLOutput{}, fmt.Errorf("hourly evidence uploads are only allowed for hourly contracts")
	}
	storageKey := uc.Store.BuildObjectKey(in.ContractID, in.FileName)
	uploadURL, err := uc.Store.PresignPutObject(ctx, storageKey, uc.PutTTL)
	if err != nil {
		return GetHourlyLogEvidenceUploadURLOutput{}, err
	}
	return GetHourlyLogEvidenceUploadURLOutput{StorageKey: storageKey, UploadURL: uploadURL}, nil
}
