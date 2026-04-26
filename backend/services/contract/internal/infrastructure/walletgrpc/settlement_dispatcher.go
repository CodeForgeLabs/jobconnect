package walletgrpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	walletv1 "jobconnect/contract/gen/wallet/v1"
	"jobconnect/contract/internal/application"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

const serviceActorID = "00000000-0000-0000-0000-00000000c0de"

type tokenIssuer interface {
	IssueAccessToken(userID uuid.UUID, role string, ttl time.Duration) (string, error)
}

type SettlementDispatcher struct {
	grpc   walletv1.WalletServiceClient
	issuer tokenIssuer
}

func NewSettlementDispatcher(grpc walletv1.WalletServiceClient, issuer tokenIssuer) *SettlementDispatcher {
	return &SettlementDispatcher{grpc: grpc, issuer: issuer}
}

func (d *SettlementDispatcher) DispatchMilestoneApproved(ctx context.Context, cmd application.MilestoneApprovedSettlementCommand) error {
	if d.grpc == nil || d.issuer == nil {
		return fmt.Errorf("wallet settlement dependencies are not configured")
	}
	actorID, _ := uuid.Parse(serviceActorID)
	token, err := d.issuer.IssueAccessToken(actorID, "service", 2*time.Minute)
	if err != nil {
		return err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
	ctx = metadata.AppendToOutgoingContext(ctx, "x-jobconnect-internal", "contract-service")

	holdResp, err := d.grpc.GetHoldByReference(ctx, &walletv1.GetHoldByReferenceRequest{
		ReferenceType: "milestone",
		ReferenceId:   strings.TrimSpace(cmd.ReferenceID),
	})
	if err != nil {
		return err
	}
	hold := holdResp.GetHold()
	if hold == nil || hold.GetId() <= 0 {
		return fmt.Errorf("hold not found for milestone reference")
	}
	_, err = d.grpc.CaptureHold(ctx, &walletv1.CaptureHoldRequest{
		HoldId:             hold.GetId(),
		CaptureAmountMinor: cmd.AmountMinor,
		IdempotencyKey:     cmd.EventID,
		ReferenceType:      "milestone",
		ReferenceId:        strings.TrimSpace(cmd.ReferenceID),
		Note:               fmt.Sprintf("milestone approved settlement for contract %d milestone %d", cmd.ContractID, cmd.MilestoneID),
	})
	return err
}

var _ application.MilestoneSettlementDispatcher = (*SettlementDispatcher)(nil)
