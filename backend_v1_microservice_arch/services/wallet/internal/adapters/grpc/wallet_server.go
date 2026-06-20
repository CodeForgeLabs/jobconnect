package grpcadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

	CreateWalletUC           *application.CreateWallet
	GetWalletUC              *application.GetWallet
	CreateDepositUC          *application.CreateDeposit
	CompleteDepositUC        *application.CompleteDeposit
	GetTransactionUC         *application.GetTransaction
	fetchWalletTransactionUc *application.FetchWalletTransactions

	TokenParser TokenParser
}

func NewWalletServer(
	createWallet *application.CreateWallet,
	getWallet *application.GetWallet,
	createDeposit *application.CreateDeposit,
	completeDeposit *application.CompleteDeposit,
	getTransaction *application.GetTransaction,
	fetchWalletTransactionUc *application.FetchWalletTransactions,
	tokenParser TokenParser,
) *WalletServer {
	return &WalletServer{
		CreateWalletUC:           createWallet,
		GetWalletUC:              getWallet,
		CreateDepositUC:          createDeposit,
		CompleteDepositUC:        completeDeposit,
		GetTransactionUC:         getTransaction,
		fetchWalletTransactionUc: fetchWalletTransactionUc,
		TokenParser:              tokenParser,
	}
}

// ==================== WALLET ====================

func (s *WalletServer) CreateWallet(
	ctx context.Context,
	req *walletv1.CreateWalletRequest,
) (*walletv1.CreateWalletResponse, error) {

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}

	// callerID, role, err := callerFromContext(ctx, s.TokenParser)
	// if err != nil {
	// 	return nil, err
	// }

	// ownerID := callerID

	// if strings.TrimSpace(req.GetOwnerId()) != "" {
	// 	parsed, parseErr := uuid.Parse(req.GetOwnerId())
	// 	if parseErr != nil {
	// 		return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
	// 	}

	// 	if parsed != callerID && !isInternalRole(role) {
	// 		return nil, status.Error(codes.PermissionDenied, "cannot create wallet for another owner")
	// 	}

	// 	ownerID = parsed
	// }

	ownerUUID, err := uuid.Parse(strings.TrimSpace(req.OwnerId))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
	}

	out, err := s.CreateWalletUC.Execute(
		ctx,
		application.CreateWalletInput{
			OwnerID: ownerUUID,
		},
	)

	if err != nil {
		return nil, toStatus(err)
	}

	return &walletv1.CreateWalletResponse{
		Wallet: toProtoWallet(out.Wallet),
	}, nil
}

func (s *WalletServer) GetWallet(
	ctx context.Context,
	req *walletv1.GetWalletRequest,
) (*walletv1.GetWalletResponse, error) {

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}

	// callerID, role, err := callerFromContext(ctx, s.TokenParser)
	// if err != nil {
	// 	return nil, err
	// }

	ownerID, err := uuid.Parse(strings.TrimSpace(req.GetOwnerId()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid owner_id")
	}

	// if ownerID != callerID && !isInternalRole(role) {
	// 	return nil, status.Error(codes.PermissionDenied, "wallet access denied")
	// }

	out, err := s.GetWalletUC.Execute(
		ctx,
		application.GetWalletInput{
			OwnerID: ownerID,
		},
	)

	if err != nil {
		return nil, toStatus(err)
	}

	return &walletv1.GetWalletResponse{
		Wallet: toProtoWallet(out.Wallet),
	}, nil
}

// ==================== DEPOSIT ====================

func (s *WalletServer) CreateDepositTransaction(
	ctx context.Context,
	req *walletv1.CreateDepositTransactionRequest,
) (*walletv1.CreateDepositTransactionResponse, error) {

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}

	// _, _, err := callerFromContext(ctx, s.TokenParser)
	// if err != nil {
	// 	return nil, err
	// }

	fmt.Printf("DEBUG INPUT: %+v\n", req)

	// Option B: Multi-line for better visibility
	fmt.Printf("--- Deposit Request Data ---\n")
	fmt.Printf("Wallet ID:    %d\n", req.GetWalletId())
	fmt.Printf("Amount Minor: %d\n", req.GetAmountMinor())
	fmt.Printf("Description:  %s\n", req.GetDescription())
	fmt.Printf("Phone:        %s\n", req.GetPhoneNumber())

	out, err := s.CreateDepositUC.Execute(
		ctx,
		application.CreateDepositInput{
			WalletID:    req.GetWalletId(),
			AmountMinor: req.GetAmountMinor(),
			Description: req.GetDescription(),
			Phone:       req.GetPhoneNumber(),
		},
	)

	if err != nil {
		return nil, toStatus(err)
	}

	return &walletv1.CreateDepositTransactionResponse{
		Transaction: toProtoTransaction(out.Transaction),
		PaymentUrl:  out.PaymentURL,
	}, nil
}

func (s *WalletServer) CompleteDeposit(
	ctx context.Context,
	req *walletv1.CompleteDepositRequest,
) (*walletv1.CompleteDepositResponse, error) {

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}

	// _, role, err := callerFromContext(ctx, s.TokenParser)
	// if err != nil {
	// 	return nil, err
	// }

	// if err := requireInternalRole(role); err != nil {
	// 	return nil, err
	// }

	out, err := s.CompleteDepositUC.Execute(
		ctx,
		application.CompleteDepositInput{
			TxRef:    req.GetTxRef(),
			ChapaRef: req.GetChapaRef(),
		},
	)

	if err != nil {
		return nil, toStatus(err)
	}

	return &walletv1.CompleteDepositResponse{
		Success: out.Success,
	}, nil
}

// ==================== TRANSACTION ====================

func (s *WalletServer) GetTransaction(
	ctx context.Context,
	req *walletv1.GetTransactionRequest,
) (*walletv1.GetTransactionResponse, error) {

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}

	// _, _, err := callerFromContext(ctx, s.TokenParser)
	// if err != nil {
	// 	return nil, err
	// }

	out, err := s.GetTransactionUC.Execute(
		ctx,
		application.GetTransactionInput{
			TxRef: req.GetTxRef(),
		},
	)

	if err != nil {
		return nil, toStatus(err)
	}

	return &walletv1.GetTransactionResponse{
		Transaction: toProtoTransaction(out.Transaction),
	}, nil
}

// ==================== FETCH TRANSACTION ====================
func (s *WalletServer) FetchWalletTransactions(
	ctx context.Context,
	req *walletv1.FetchWalletTransactionsRequest,
) (*walletv1.FetchWalletTransactionsResponse, error) {

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	out, err := s.fetchWalletTransactionUc.Execute(
		ctx,
		application.FetchWalletTransactionsInput{
			WalletID: req.GetWalletId(),
		},
	)

	if err != nil {
		return nil, toStatus(err)
	}

	protoTxs := make([]*walletv1.WalletTransaction, len(out.Transactions))
	for i, tx := range out.Transactions {
		protoTxs[i] = toProtoTransaction(tx)
	}

	return &walletv1.FetchWalletTransactionsResponse{
		Transactions: protoTxs,
	}, nil
}

// ==================== STATUS HANDLER ====================

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

	case errors.Is(err, domain.ErrConflict):
		return status.Error(codes.FailedPrecondition, err.Error())

	case strings.Contains(s, "required") || strings.Contains(s, "invalid"):
		return status.Error(codes.InvalidArgument, err.Error())

	default:
		return status.Error(codes.Internal, fmt.Sprintf("internal error: %v", err))
	}
}

// ==================== PROTO CONVERTERS ====================

func toProtoWallet(in domain.WalletAccount) *walletv1.Wallet {
	return &walletv1.Wallet{
		Id:                   in.ID,
		OwnerId:              in.OwnerID.String(),
		BalanceMinor:         in.BalanceMinor,
		CreatedAtUnixSeconds: in.CreatedAt.Unix(),
	}
}

func toProtoTransaction(in domain.WalletTransaction) *walletv1.WalletTransaction {
	return &walletv1.WalletTransaction{
		Id:                   in.ID,
		WalletId:             in.WalletID,
		TxRef:                in.TxRef,
		ChapaRef:             safeString(in.ChapaRef),
		AmountMinor:          in.AmountMinor,
		TxType:               toProtoTransactionType(in.TxType),
		Description:          in.Description,
		Status:               toProtoTransactionStatus(in.Status),
		CreatedAtUnixSeconds: in.CreatedAt.Unix(),
	}
}
func safeString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
func toProtoTransactionStatus(v string) walletv1.TransactionStatus {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case domain.TransactionPending:
		return walletv1.TransactionStatus_TRANSACTION_STATUS_PENDING
	case domain.TransactionSuccess:
		return walletv1.TransactionStatus_TRANSACTION_STATUS_SUCCESS
	case domain.TransactionFailed:
		return walletv1.TransactionStatus_TRANSACTION_STATUS_FAILED
	default:
		return walletv1.TransactionStatus_TRANSACTION_STATUS_UNSPECIFIED
	}
}

func toProtoTransactionType(v string) walletv1.TransactionType {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case domain.TransactionTypeDeposit:
		return walletv1.TransactionType_TRANSACTION_TYPE_DEPOSIT
	case domain.TransactionTypeWithdrawal:
		return walletv1.TransactionType_TRANSACTION_TYPE_WITHDRAWAL
	case domain.TransactionTypePayment:
		return walletv1.TransactionType_TRANSACTION_TYPE_PAYMENT
	default:
		return walletv1.TransactionType_TRANSACTION_TYPE_UNSPECIFIED
	}
}

func (s *WalletServer) ChapaWebhook(w http.ResponseWriter, r *http.Request) {

	var payload ChapaWebhookPayload

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// only process success payments
	if payload.Status != "success" {
		w.WriteHeader(http.StatusOK)
		return
	}

	_, err := s.CompleteDepositUC.Execute(r.Context(), application.CompleteDepositInput{
		TxRef:    payload.TrxRef,
		ChapaRef: payload.RefID,
	})

	if err != nil {
		// log error but still return 200 (important for webhooks)
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type ChapaWebhookPayload struct {
	TrxRef string `json:"trx_ref"`
	RefID  string `json:"ref_id"`
	Status string `json:"status"`
}
