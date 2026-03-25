package grpc

import (
	"context"

	"github.com/google/uuid"
	paymentv1 "jobconnect/payment/gen/payment/v1"
	"jobconnect/payment/internal/application"
	"jobconnect/payment/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	paymentv1.UnimplementedPaymentServiceServer
	initiateDeposit   *application.InitiateDeposit
	verifyDeposit     *application.VerifyDeposit
	requestWithdrawal *application.RequestWithdrawal
	getSession        *application.GetSession
	listSessions      *application.ListSessions
	uploadReceipt     *application.UploadReceipt
}

func NewServer(
	initiateDeposit *application.InitiateDeposit,
	verifyDeposit *application.VerifyDeposit,
	requestWithdrawal *application.RequestWithdrawal,
	getSession *application.GetSession,
	listSessions *application.ListSessions,
	uploadReceipt *application.UploadReceipt,
) *Server {
	return &Server{
		initiateDeposit:   initiateDeposit,
		verifyDeposit:     verifyDeposit,
		requestWithdrawal: requestWithdrawal,
		getSession:        getSession,
		listSessions:      listSessions,
		uploadReceipt:     uploadReceipt,
	}
}

func (s *Server) InitiateDeposit(ctx context.Context, req *paymentv1.InitiateDepositRequest) (*paymentv1.InitiateDepositResponse, error) {
	// In a real app, UserID would come from gRPC metadata/context via interceptors.
	// For this exercise, since the proto lacks it, we use a dummy UUID.
	userId := uuid.Nil

	res, checkoutURL, err := s.initiateDeposit.Execute(ctx, application.InitiateDepositInput{
		UserID:      userId,
		Provider:    req.Provider.String(),
		AmountMinor: req.AmountMinor,
		Currency:    req.Currency,
		ReferenceID: req.ReferenceId,
		ReferenceType: req.ReferenceType,
		ReturnURL:   req.ReturnUrl,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to initiate deposit: %v", err)
	}

	return &paymentv1.InitiateDepositResponse{
		Session:     mapToProtoSession(res),
		CheckoutUrl: checkoutURL,
	}, nil
}

func (s *Server) VerifyDeposit(ctx context.Context, req *paymentv1.VerifyDepositRequest) (*paymentv1.VerifyDepositResponse, error) {
	session, err := s.verifyDeposit.Execute(ctx, req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify deposit: %v", err)
	}

	return &paymentv1.VerifyDepositResponse{
		Session: mapToProtoSession(session),
	}, nil
}

func (s *Server) RequestWithdrawal(ctx context.Context, req *paymentv1.RequestWithdrawalRequest) (*paymentv1.RequestWithdrawalResponse, error) {
	userId := uuid.Nil

	res, err := s.requestWithdrawal.Execute(ctx, application.RequestWithdrawalInput{
		UserID:            userId,
		Provider:          req.Provider.String(),
		AmountMinor:       req.AmountMinor,
		Currency:          req.Currency,
		BankCode:          req.BankCode,
		AccountNumber:     req.AccountNumber,
		AccountHolderName: req.AccountHolderName,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to request withdrawal: %v", err)
	}

	return &paymentv1.RequestWithdrawalResponse{
		Session: mapToProtoSession(res),
	}, nil
}

func (s *Server) UploadReceipt(ctx context.Context, req *paymentv1.UploadReceiptRequest) (*paymentv1.UploadReceiptResponse, error) {
	userId := uuid.Nil

	session, resString, err := s.uploadReceipt.Execute(ctx, application.UploadReceiptInput{
		SessionID:   req.SessionId,
		UserID:      userId,
		ReceiptData: req.ReceiptData,
		ContentType: req.ContentType,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to upload receipt: %v", err)
	}

	// Fetch updated session to return
	session, err = s.getSession.Execute(ctx, req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "receipt uploaded but failed to fetch session: %v", err)
	}

	return &paymentv1.UploadReceiptResponse{
		Session:    mapToProtoSession(session),
		ReceiptUrl: resString,
	}, nil
}

// Queries...
func (s *Server) GetPaymentSession(ctx context.Context, req *paymentv1.GetPaymentSessionRequest) (*paymentv1.GetPaymentSessionResponse, error) {
	session, err := s.getSession.Execute(ctx, req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "session not found")
	}

	return &paymentv1.GetPaymentSessionResponse{
		Session: mapToProtoSession(session),
	}, nil
}

func (s *Server) ListPaymentSessions(ctx context.Context, req *paymentv1.ListPaymentSessionsRequest) (*paymentv1.ListPaymentSessionsResponse, error) {
	userId := uuid.Nil

	var statusFilter *string
	if req.Status != paymentv1.PaymentSessionStatus_PAYMENT_SESSION_STATUS_UNSPECIFIED {
		sStr := req.Status.String()
		statusFilter = &sStr
	}

	var typeFilter *string
	if req.PaymentType != paymentv1.PaymentType_PAYMENT_TYPE_UNSPECIFIED {
		tStr := req.PaymentType.String()
		typeFilter = &tStr
	}

	sessions, err := s.listSessions.Execute(ctx, application.ListSessionsInput{
		UserID:      userId,
		Status:      statusFilter,
		PaymentType: typeFilter,
		PageSize:    int(req.PageSize),
		Offset:      0, // Using 0 for simplicity instead of parsing PageToken
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list sessions: %v", err)
	}

	var protos []*paymentv1.PaymentSession
	for _, sess := range sessions {
		protos = append(protos, mapToProtoSession(sess))
	}

	return &paymentv1.ListPaymentSessionsResponse{
		Sessions:      protos,
		NextPageToken: "", // Simplification for now
	}, nil
}

func mapToProtoSession(s domain.PaymentSession) *paymentv1.PaymentSession {
	var protoProvider paymentv1.PaymentProvider
	if s.Provider == "CHAPA" || s.Provider == "PAYMENT_PROVIDER_CHAPA" {
		protoProvider = paymentv1.PaymentProvider_PAYMENT_PROVIDER_CHAPA
	} else if s.Provider == "TELEBIRR" || s.Provider == "PAYMENT_PROVIDER_TELEBIRR"{
		protoProvider = paymentv1.PaymentProvider_PAYMENT_PROVIDER_TELEBIRR
	} else {
		protoProvider = paymentv1.PaymentProvider_PAYMENT_PROVIDER_UNSPECIFIED
	}

	var protoType paymentv1.PaymentType
	switch s.PaymentType {
	case domain.TypeDeposit:
		protoType = paymentv1.PaymentType_PAYMENT_TYPE_DEPOSIT
	case domain.TypeWithdrawal:
		protoType = paymentv1.PaymentType_PAYMENT_TYPE_WITHDRAWAL
	}

	var protoStatus paymentv1.PaymentSessionStatus
	switch s.Status {
	case domain.StatusPending:
		protoStatus = paymentv1.PaymentSessionStatus_PAYMENT_SESSION_STATUS_PENDING
	case domain.StatusCompleted:
		protoStatus = paymentv1.PaymentSessionStatus_PAYMENT_SESSION_STATUS_COMPLETED
	case domain.StatusFailed:
		protoStatus = paymentv1.PaymentSessionStatus_PAYMENT_SESSION_STATUS_FAILED
	case domain.StatusRefunded:
		protoStatus = paymentv1.PaymentSessionStatus_PAYMENT_SESSION_STATUS_REFUNDED
	}

	return &paymentv1.PaymentSession{
		Id:                 s.ID,
		UserId:             s.UserID.String(),
		Provider:           protoProvider,
		PaymentType:        protoType,
		Status:             protoStatus,
		AmountMinor:        s.AmountMinor,
		Currency:           s.Currency,
		IdempotencyKey:     s.IdempotencyKey,
		ExternalRef:        s.ExternalRef,
		ReceiptStorageKey:  s.ReceiptKey,
		ReferenceType:      s.ReferenceType,
		ReferenceId:        s.ReferenceID,
		ErrorMessage:       s.ErrorMessage,
		CreatedAtUnixSeconds: s.CreatedAt.Unix(),
		UpdatedAtUnixSeconds: s.UpdatedAt.Unix(),
	}
}
