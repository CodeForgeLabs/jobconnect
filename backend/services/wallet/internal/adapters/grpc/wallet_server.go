package grpcadapter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	walletv1 "jobconnect/wallet/gen/wallet/v1"
	"jobconnect/wallet/internal/application"
	"jobconnect/wallet/internal/domain"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WalletServer struct {
	walletv1.UnimplementedWalletServiceServer

	CreateWalletUC         *application.CreateWallet
	GetWalletUC            *application.GetWallet
	GetBalanceUC           *application.GetBalance
	CreditWalletInternalUC *application.CreditWalletInternal
	DebitWalletInternalUC  *application.DebitWalletInternal
	PlaceHoldUC            *application.PlaceHold
	ReleaseHoldUC          *application.ReleaseHold
	CaptureHoldUC          *application.CaptureHold
	ListTransactionsUC     *application.ListTransactions

	TokenParser TokenParser
}

func NewWalletServer(
	createWallet *application.CreateWallet,
	getWallet *application.GetWallet,
	getBalance *application.GetBalance,
	creditWalletInternal *application.CreditWalletInternal,
	debitWalletInternal *application.DebitWalletInternal,
	placeHold *application.PlaceHold,
	releaseHold *application.ReleaseHold,
	captureHold *application.CaptureHold,
	listTransactions *application.ListTransactions,
	tokenParser TokenParser,
) *WalletServer {
	return &WalletServer{
		CreateWalletUC:         createWallet,
		GetWalletUC:            getWallet,
		GetBalanceUC:           getBalance,
		CreditWalletInternalUC: creditWalletInternal,
		DebitWalletInternalUC:  debitWalletInternal,
		PlaceHoldUC:            placeHold,
		ReleaseHoldUC:          releaseHold,
		CaptureHoldUC:          captureHold,
		ListTransactionsUC:     listTransactions,
		TokenParser:            tokenParser,
	}
}

func (s *WalletServer) CreateWallet(ctx context.Context, req *walletv1.CreateWalletRequest) (*walletv1.CreateWalletResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}

	ownerID := callerID
	if strings.TrimSpace(req.GetOwnerId()) != "" {
		parsed, parseErr := uuid.Parse(req.GetOwnerId())
		if parseErr != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
		}
		if parsed != callerID && !isInternalRole(role) {
			return nil, status.Error(codes.PermissionDenied, "cannot create wallet for another owner")
		}
		ownerID = parsed
	}

	out, err := s.CreateWalletUC.Execute(ctx, application.CreateWalletInput{OwnerID: ownerID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &walletv1.CreateWalletResponse{Wallet: toProtoWallet(out.Wallet)}, nil
}

func (s *WalletServer) GetWallet(ctx context.Context, req *walletv1.GetWalletRequest) (*walletv1.GetWalletResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}

	in := application.GetWalletInput{}
	switch target := req.GetTarget().(type) {
	case *walletv1.GetWalletRequest_WalletId:
		in.WalletID = target.WalletId
	case *walletv1.GetWalletRequest_OwnerId:
		ownerID, parseErr := uuid.Parse(strings.TrimSpace(target.OwnerId))
		if parseErr != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
		}
		in.OwnerID = ownerID
	default:
		return nil, status.Error(codes.InvalidArgument, "target is required")
	}

	out, err := s.GetWalletUC.Execute(ctx, in)
	if err != nil {
		return nil, toStatus(err)
	}
	if out.Wallet.OwnerID != callerID && !isInternalRole(role) {
		return nil, status.Error(codes.PermissionDenied, "wallet access denied")
	}
	return &walletv1.GetWalletResponse{Wallet: toProtoWallet(out.Wallet)}, nil
}

func (s *WalletServer) GetBalance(ctx context.Context, req *walletv1.GetBalanceRequest) (*walletv1.GetBalanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	walletOut, err := s.GetWalletUC.Execute(ctx, application.GetWalletInput{WalletID: req.GetWalletId()})
	if err != nil {
		return nil, toStatus(err)
	}
	if walletOut.Wallet.OwnerID != callerID && !isInternalRole(role) {
		return nil, status.Error(codes.PermissionDenied, "wallet access denied")
	}
	out, err := s.GetBalanceUC.Execute(ctx, application.GetBalanceInput{WalletID: req.GetWalletId()})
	if err != nil {
		return nil, toStatus(err)
	}
	return &walletv1.GetBalanceResponse{Balance: toProtoBalance(out.Balance)}, nil
}

func (s *WalletServer) CreditWalletInternal(ctx context.Context, req *walletv1.CreditWalletInternalRequest) (*walletv1.CreditWalletInternalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	_, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireInternalRole(role); err != nil {
		return nil, err
	}
	out, err := s.CreditWalletInternalUC.Execute(ctx, application.CreditWalletInternalInput{
		WalletID:       req.GetWalletId(),
		AmountMinor:    req.GetAmountMinor(),
		IdempotencyKey: req.GetIdempotencyKey(),
		ReferenceType:  req.GetReferenceType(),
		ReferenceID:    req.GetReferenceId(),
		Note:           req.GetNote(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &walletv1.CreditWalletInternalResponse{Wallet: toProtoWallet(out.Result.Wallet), Transaction: toProtoEntry(out.Result.Entry)}, nil
}

func (s *WalletServer) DebitWalletInternal(ctx context.Context, req *walletv1.DebitWalletInternalRequest) (*walletv1.DebitWalletInternalResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	_, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireInternalRole(role); err != nil {
		return nil, err
	}
	out, err := s.DebitWalletInternalUC.Execute(ctx, application.DebitWalletInternalInput{
		WalletID:       req.GetWalletId(),
		AmountMinor:    req.GetAmountMinor(),
		IdempotencyKey: req.GetIdempotencyKey(),
		ReferenceType:  req.GetReferenceType(),
		ReferenceID:    req.GetReferenceId(),
		Note:           req.GetNote(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &walletv1.DebitWalletInternalResponse{Wallet: toProtoWallet(out.Result.Wallet), Transaction: toProtoEntry(out.Result.Entry)}, nil
}

func (s *WalletServer) PlaceHold(ctx context.Context, req *walletv1.PlaceHoldRequest) (*walletv1.PlaceHoldResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	_, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireInternalRole(role); err != nil {
		return nil, err
	}
	out, err := s.PlaceHoldUC.Execute(ctx, application.PlaceHoldCommand{
		WalletID:       req.GetWalletId(),
		AmountMinor:    req.GetAmountMinor(),
		IdempotencyKey: req.GetIdempotencyKey(),
		ReferenceType:  req.GetReferenceType(),
		ReferenceID:    req.GetReferenceId(),
		ExpiresAtUnix:  req.GetExpiresAtUnixSeconds(),
		Note:           req.GetNote(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &walletv1.PlaceHoldResponse{Wallet: toProtoWallet(out.Result.Wallet), Hold: toProtoHold(out.Result.Hold), Transaction: toProtoEntry(out.Result.Entry)}, nil
}

func (s *WalletServer) ReleaseHold(ctx context.Context, req *walletv1.ReleaseHoldRequest) (*walletv1.ReleaseHoldResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	_, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireInternalRole(role); err != nil {
		return nil, err
	}
	out, err := s.ReleaseHoldUC.Execute(ctx, application.ReleaseHoldCommand{
		HoldID:         req.GetHoldId(),
		IdempotencyKey: req.GetIdempotencyKey(),
		Note:           req.GetNote(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &walletv1.ReleaseHoldResponse{Wallet: toProtoWallet(out.Result.Wallet), Hold: toProtoHold(out.Result.Hold), Transaction: toProtoEntry(out.Result.Entry)}, nil
}

func (s *WalletServer) CaptureHold(ctx context.Context, req *walletv1.CaptureHoldRequest) (*walletv1.CaptureHoldResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	_, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireInternalRole(role); err != nil {
		return nil, err
	}
	out, err := s.CaptureHoldUC.Execute(ctx, application.CaptureHoldCommand{
		HoldID:             req.GetHoldId(),
		CaptureAmountMinor: req.GetCaptureAmountMinor(),
		IdempotencyKey:     req.GetIdempotencyKey(),
		ReferenceType:      req.GetReferenceType(),
		ReferenceID:        req.GetReferenceId(),
		Note:               req.GetNote(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &walletv1.CaptureHoldResponse{Wallet: toProtoWallet(out.Result.Wallet), Hold: toProtoHold(out.Result.Hold), Transaction: toProtoEntry(out.Result.Entry)}, nil
}

func (s *WalletServer) ListTransactions(ctx context.Context, req *walletv1.ListTransactionsRequest) (*walletv1.ListTransactionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	walletOut, err := s.GetWalletUC.Execute(ctx, application.GetWalletInput{WalletID: req.GetWalletId()})
	if err != nil {
		return nil, toStatus(err)
	}
	if walletOut.Wallet.OwnerID != callerID && !isInternalRole(role) {
		return nil, status.Error(codes.PermissionDenied, "wallet access denied")
	}

	out, err := s.ListTransactionsUC.Execute(ctx, application.ListTransactionsInput{
		WalletID:  req.GetWalletId(),
		PageSize:  req.GetPageSize(),
		PageToken: req.GetPageToken(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	items := make([]*walletv1.LedgerEntry, 0, len(out.Transactions))
	for _, item := range out.Transactions {
		items = append(items, toProtoEntry(item))
	}
	return &walletv1.ListTransactionsResponse{Transactions: items, NextPageToken: out.NextPageToken}, nil
}

func toStatus(err error) error {
	if err == nil {
		return nil
	}
	s := strings.ToLower(strings.TrimSpace(err.Error()))
	switch {
	case errors.Is(err, domain.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrInsufficientFunds):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrConflict):
		return status.Error(codes.FailedPrecondition, err.Error())
	case strings.Contains(s, "required") || strings.Contains(s, "invalid"):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, fmt.Sprintf("internal error: %v", err))
	}
}

func toProtoWallet(in domain.WalletAccount) *walletv1.Wallet {
	return &walletv1.Wallet{
		Id:                   in.ID,
		OwnerId:              in.OwnerID.String(),
		Status:               toProtoWalletStatus(in.Status),
		AvailableMinor:       in.AvailableMinor,
		HeldMinor:            in.HeldMinor,
		CreatedAtUnixSeconds: in.CreatedAt.Unix(),
		UpdatedAtUnixSeconds: in.UpdatedAt.Unix(),
	}
}

func toProtoBalance(in domain.BalanceSnapshot) *walletv1.Balance {
	return &walletv1.Balance{
		AvailableMinor: in.AvailableMinor,
		HeldMinor:      in.HeldMinor,
		TotalMinor:     in.TotalMinor(),
	}
}

func toProtoEntry(in domain.LedgerEntry) *walletv1.LedgerEntry {
	return &walletv1.LedgerEntry{
		Id:                   in.ID,
		WalletId:             in.WalletID,
		Type:                 toProtoLedgerType(in.EntryType),
		AmountMinor:          in.AmountMinor,
		IdempotencyKey:       in.IdempotencyKey,
		ReferenceType:        in.ReferenceType,
		ReferenceId:          in.ReferenceID,
		Note:                 in.Note,
		AvailableAfterMinor:  in.AvailableAfterMinor,
		HeldAfterMinor:       in.HeldAfterMinor,
		CreatedAtUnixSeconds: in.CreatedAt.Unix(),
	}
}

func toProtoHold(in domain.Hold) *walletv1.Hold {
	expiresAt := int64(0)
	if in.ExpiresAt != nil {
		expiresAt = in.ExpiresAt.Unix()
	}
	return &walletv1.Hold{
		Id:                   in.ID,
		WalletId:             in.WalletID,
		ReferenceType:        in.ReferenceType,
		ReferenceId:          in.ReferenceID,
		AmountMinor:          in.AmountMinor,
		CapturedMinor:        in.CapturedMinor,
		Status:               toProtoHoldStatus(in.Status),
		ExpiresAtUnixSeconds: expiresAt,
		CreatedAtUnixSeconds: in.CreatedAt.Unix(),
		UpdatedAtUnixSeconds: in.UpdatedAt.Unix(),
	}
}

func toProtoWalletStatus(v string) walletv1.WalletStatus {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case domain.WalletStatusActive:
		return walletv1.WalletStatus_WALLET_STATUS_ACTIVE
	case domain.WalletStatusFrozen:
		return walletv1.WalletStatus_WALLET_STATUS_FROZEN
	default:
		return walletv1.WalletStatus_WALLET_STATUS_UNSPECIFIED
	}
}

func toProtoLedgerType(v string) walletv1.LedgerEntryType {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case domain.LedgerTypeCreditInternal:
		return walletv1.LedgerEntryType_LEDGER_ENTRY_TYPE_CREDIT_INTERNAL
	case domain.LedgerTypeDebitInternal:
		return walletv1.LedgerEntryType_LEDGER_ENTRY_TYPE_DEBIT_INTERNAL
	case domain.LedgerTypeHoldPlaced:
		return walletv1.LedgerEntryType_LEDGER_ENTRY_TYPE_HOLD_PLACED
	case domain.LedgerTypeHoldReleased:
		return walletv1.LedgerEntryType_LEDGER_ENTRY_TYPE_HOLD_RELEASED
	case domain.LedgerTypeHoldCaptured:
		return walletv1.LedgerEntryType_LEDGER_ENTRY_TYPE_HOLD_CAPTURED
	default:
		return walletv1.LedgerEntryType_LEDGER_ENTRY_TYPE_UNSPECIFIED
	}
}

func toProtoHoldStatus(v string) walletv1.HoldStatus {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case domain.HoldStatusActive:
		return walletv1.HoldStatus_HOLD_STATUS_ACTIVE
	case domain.HoldStatusReleased:
		return walletv1.HoldStatus_HOLD_STATUS_RELEASED
	case domain.HoldStatusCaptured:
		return walletv1.HoldStatus_HOLD_STATUS_CAPTURED
	default:
		return walletv1.HoldStatus_HOLD_STATUS_UNSPECIFIED
	}
}
