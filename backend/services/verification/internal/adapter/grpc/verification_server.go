package grpcadapter

import (
	"context"
	"strings"
	"time"

	verificationv1 "jobconnect/verification/gen/verification/v1"
	"jobconnect/verification/internal/application"
	"jobconnect/verification/internal/domain"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type VerificationServer struct {
	verificationv1.UnimplementedVerificationServiceServer
	GetEvidenceUploadURLUC    *application.GetVerificationEvidenceUploadURL
	SubmitVerificationUC     *application.SubmitVerification
	GetMyStatusUC            *application.GetMyVerificationStatus
	ListPendingUC            *application.ListPendingVerifications
	GetVerificationRequestUC *application.GetVerificationRequest
	ReviewVerificationUC     *application.ReviewVerification
	RequestReverificationUC  *application.RequestReverification
}

func NewVerificationServer(
	getEvidenceUploadURL *application.GetVerificationEvidenceUploadURL,
	submit *application.SubmitVerification,
	getMyStatus *application.GetMyVerificationStatus,
	listPending *application.ListPendingVerifications,
	getRequest *application.GetVerificationRequest,
	review *application.ReviewVerification,
	requestReverification *application.RequestReverification,
) *VerificationServer {
	return &VerificationServer{
		GetEvidenceUploadURLUC:    getEvidenceUploadURL,
		SubmitVerificationUC:     submit,
		GetMyStatusUC:            getMyStatus,
		ListPendingUC:            listPending,
		GetVerificationRequestUC: getRequest,
		ReviewVerificationUC:     review,
		RequestReverificationUC:  requestReverification,
	}
}

func (s *VerificationServer) GetVerificationEvidenceUploadUrl(ctx context.Context, req *verificationv1.GetVerificationEvidenceUploadUrlRequest) (*verificationv1.GetVerificationEvidenceUploadUrlResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	if s.GetEvidenceUploadURLUC == nil {
		return nil, status.Error(codes.Internal, "verification evidence upload url use-case not configured")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.GetEvidenceUploadURLUC.Execute(ctx, application.GetVerificationEvidenceUploadURLInput{
		UserID:      userID,
		FileName:    req.GetFileName(),
		ContentType: req.GetContentType(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &verificationv1.GetVerificationEvidenceUploadUrlResponse{
		StorageKey: out.StorageKey,
		UploadUrl:  out.UploadURL,
	}, nil
}

func (s *VerificationServer) SubmitVerification(ctx context.Context, req *verificationv1.SubmitVerificationRequest) (*verificationv1.SubmitVerificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.SubmitVerificationUC.Execute(ctx, application.SubmitVerificationInput{
		UserID:               userID,
		LegalName:            req.GetLegalName(),
		CountryCode:          req.GetCountryCode(),
		DocumentType:         req.GetDocumentType(),
		DocumentNumberMasked: req.GetDocumentNumberMasked(),
		EvidenceURL:          req.GetEvidenceUrl(),
		SubmissionNote:       req.GetSubmissionNote(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &verificationv1.SubmitVerificationResponse{Request: toProto(out)}, nil
}

func (s *VerificationServer) GetMyVerificationStatus(ctx context.Context, req *verificationv1.GetMyVerificationStatusRequest) (*verificationv1.GetMyVerificationStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.GetMyStatusUC.Execute(ctx, application.GetMyVerificationStatusInput{UserID: userID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &verificationv1.GetMyVerificationStatusResponse{Request: toProto(out)}, nil
}

func (s *VerificationServer) ListPendingVerifications(ctx context.Context, req *verificationv1.ListPendingVerificationsRequest) (*verificationv1.ListPendingVerificationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	out, err := s.ListPendingUC.Execute(ctx, application.ListPendingVerificationsInput{PageSize: req.GetPageSize(), Page: req.GetPage()})
	if err != nil {
		return nil, toStatus(err)
	}
	items := make([]*verificationv1.VerificationRequest, 0, len(out))
	for _, item := range out {
		items = append(items, toProto(item))
	}
	return &verificationv1.ListPendingVerificationsResponse{Requests: items}, nil
}

func (s *VerificationServer) GetVerificationRequest(ctx context.Context, req *verificationv1.GetVerificationRequestRequest) (*verificationv1.GetVerificationRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	out, err := s.GetVerificationRequestUC.Execute(ctx, application.GetVerificationRequestInput{RequestID: req.GetRequestId()})
	if err != nil {
		return nil, toStatus(err)
	}
	return &verificationv1.GetVerificationRequestResponse{Request: toProto(out)}, nil
}

func (s *VerificationServer) ReviewVerification(ctx context.Context, req *verificationv1.ReviewVerificationRequest) (*verificationv1.ReviewVerificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	reviewerID, err := uuid.Parse(req.GetReviewerUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid reviewer_user_id")
	}
	out, err := s.ReviewVerificationUC.Execute(ctx, application.ReviewVerificationInput{
		RequestID:       req.GetRequestId(),
		ReviewerUserID:  reviewerID,
		Decision:        req.GetDecision(),
		RejectionReason: req.GetRejectionReason(),
		InternalNote:    req.GetInternalNote(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &verificationv1.ReviewVerificationResponse{Request: toProto(out)}, nil
}

func (s *VerificationServer) RequestReverification(ctx context.Context, req *verificationv1.RequestReverificationRequest) (*verificationv1.RequestReverificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	reviewerID, err := uuid.Parse(req.GetReviewerUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid reviewer_user_id")
	}
	due := time.Unix(req.GetReverifyDueAtUnix(), 0).UTC()
	out, err := s.RequestReverificationUC.Execute(ctx, application.RequestReverificationInput{
		UserID:         userID,
		ReviewerUserID: reviewerID,
		Reason:         req.GetReason(),
		ReverifyDueAt:  due,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &verificationv1.RequestReverificationResponse{Request: toProto(out)}, nil
}

func toProto(in domain.VerificationRequest) *verificationv1.VerificationRequest {
	out := &verificationv1.VerificationRequest{
		Id:                   in.ID,
		UserId:               in.UserID.String(),
		RequestVersion:       in.RequestVersion,
		Status:               in.Status,
		LegalName:            in.LegalName,
		CountryCode:          in.CountryCode,
		DocumentType:         in.DocumentType,
		DocumentNumberMasked: in.DocumentNumberMasked,
		EvidenceUrl:          in.EvidenceURL,
		SubmissionNote:       in.SubmissionNote,
		RejectionReason:      in.RejectionReason,
		InternalNote:         in.InternalNote,
		SubmittedAtUnix:      in.SubmittedAt.Unix(),
	}
	if in.ReviewerUserID != nil {
		out.ReviewerUserId = in.ReviewerUserID.String()
	}
	if in.ReviewedAt != nil {
		out.ReviewedAtUnix = in.ReviewedAt.Unix()
	}
	if in.ReverifyDueAt != nil {
		out.ReverifyDueAtUnix = in.ReverifyDueAt.Unix()
	}
	return out
}

func toStatus(err error) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "required"), strings.Contains(msg, "invalid"):
		return status.Error(codes.InvalidArgument, err.Error())
	case strings.Contains(msg, "not found"):
		return status.Error(codes.NotFound, err.Error())
	case strings.Contains(msg, "forbidden"), strings.Contains(msg, "not allowed"):
		return status.Error(codes.PermissionDenied, err.Error())
	case strings.Contains(msg, "pending"):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
