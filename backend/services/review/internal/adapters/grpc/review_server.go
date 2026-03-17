package grpcadapter

import (
	"context"
	"strings"

	reviewv1 "jobconnect/review/gen/review/v1"
	"jobconnect/review/internal/application"
	"jobconnect/review/internal/domain"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ReviewServer struct {
	reviewv1.UnimplementedReviewServiceServer
	CreateReviewUC           *application.CreateReview
	GetReviewUC              *application.GetReview
	ListReviewsByUserUC      *application.ListReviewsByUser
	ListReviewsByContractUC  *application.ListReviewsByContract
	GetUserRatingSummaryUC   *application.GetUserRatingSummary
	UpdateReviewUC           *application.UpdateReview
	DeleteReviewUC           *application.DeleteReview
	ReplyToReviewUC          *application.ReplyToReview
	TokenParser              TokenParser
}

func NewReviewServer(
	createReview *application.CreateReview,
	getReview *application.GetReview,
	listByUser *application.ListReviewsByUser,
	listByContract *application.ListReviewsByContract,
	ratingSummary *application.GetUserRatingSummary,
	updateReview *application.UpdateReview,
	deleteReview *application.DeleteReview,
	replyToReview *application.ReplyToReview,
	tokenParser TokenParser,
) *ReviewServer {
	return &ReviewServer{
		CreateReviewUC:          createReview,
		GetReviewUC:             getReview,
		ListReviewsByUserUC:     listByUser,
		ListReviewsByContractUC: listByContract,
		GetUserRatingSummaryUC:  ratingSummary,
		UpdateReviewUC:          updateReview,
		DeleteReviewUC:          deleteReview,
		ReplyToReviewUC:         replyToReview,
		TokenParser:             tokenParser,
	}
}

func (s *ReviewServer) CreateReview(ctx context.Context, req *reviewv1.CreateReviewRequest) (*reviewv1.CreateReviewResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}

	revieweeID, err := uuid.Parse(req.RevieweeId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid reviewee_id")
	}

	out, err := s.CreateReviewUC.Execute(ctx, application.CreateReviewInput{
		ContractID:   req.ContractId,
		ReviewerID:   callerID,
		ReviewerRole: role,
		RevieweeID:   revieweeID,
		Rating:       req.Rating,
		Title:        req.Title,
		Comment:      req.Comment,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &reviewv1.CreateReviewResponse{Review: toProtoReview(out.Review)}, nil
}

func (s *ReviewServer) GetReview(ctx context.Context, req *reviewv1.GetReviewRequest) (*reviewv1.GetReviewResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	out, err := s.GetReviewUC.Execute(ctx, application.GetReviewInput{ReviewID: req.ReviewId})
	if err != nil {
		return nil, toStatus(err)
	}
	return &reviewv1.GetReviewResponse{Review: toProtoReview(out.Review)}, nil
}

func (s *ReviewServer) ListReviewsByUser(ctx context.Context, req *reviewv1.ListReviewsByUserRequest) (*reviewv1.ListReviewsByUserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.ListReviewsByUserUC.Execute(ctx, application.ListReviewsByUserInput{
		UserID:    userID,
		PageSize:  req.PageSize,
		PageToken: req.PageToken,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	reviews := make([]*reviewv1.Review, 0, len(out.Reviews))
	for _, r := range out.Reviews {
		reviews = append(reviews, toProtoReview(r))
	}
	return &reviewv1.ListReviewsByUserResponse{Reviews: reviews, NextPageToken: out.NextPageToken}, nil
}

func (s *ReviewServer) ListReviewsByContract(ctx context.Context, req *reviewv1.ListReviewsByContractRequest) (*reviewv1.ListReviewsByContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	out, err := s.ListReviewsByContractUC.Execute(ctx, application.ListReviewsByContractInput{
		ContractID: req.ContractId,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	reviews := make([]*reviewv1.Review, 0, len(out.Reviews))
	for _, r := range out.Reviews {
		reviews = append(reviews, toProtoReview(r))
	}
	return &reviewv1.ListReviewsByContractResponse{Reviews: reviews}, nil
}

func (s *ReviewServer) GetUserRatingSummary(ctx context.Context, req *reviewv1.GetUserRatingSummaryRequest) (*reviewv1.GetUserRatingSummaryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.GetUserRatingSummaryUC.Execute(ctx, application.GetUserRatingSummaryInput{UserID: userID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &reviewv1.GetUserRatingSummaryResponse{
		AverageRating: out.AverageRating,
		TotalReviews:  out.TotalReviews,
	}, nil
}

func (s *ReviewServer) UpdateReview(ctx context.Context, req *reviewv1.UpdateReviewRequest) (*reviewv1.UpdateReviewResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, _, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}

	out, err := s.UpdateReviewUC.Execute(ctx, application.UpdateReviewInput{
		ReviewID:    req.ReviewId,
		RequesterID: callerID,
		Rating:      req.Rating,
		Title:       req.Title,
		Comment:     req.Comment,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &reviewv1.UpdateReviewResponse{Review: toProtoReview(out.Review)}, nil
}

func (s *ReviewServer) DeleteReview(ctx context.Context, req *reviewv1.DeleteReviewRequest) (*reviewv1.DeleteReviewResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, _, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}

	err = s.DeleteReviewUC.Execute(ctx, application.DeleteReviewInput{
		ReviewID:    req.ReviewId,
		RequesterID: callerID,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &reviewv1.DeleteReviewResponse{Deleted: true}, nil
}

func (s *ReviewServer) ReplyToReview(ctx context.Context, req *reviewv1.ReplyToReviewRequest) (*reviewv1.ReplyToReviewResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, _, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}

	out, err := s.ReplyToReviewUC.Execute(ctx, application.ReplyToReviewInput{
		ReviewID:     req.ReviewId,
		RequesterID:  callerID,
		ReplyComment: req.ReplyComment,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &reviewv1.ReplyToReviewResponse{Review: toProtoReview(out.Review)}, nil
}

func toProtoReview(r domain.Review) *reviewv1.Review {
	p := &reviewv1.Review{
		Id:                   r.ID,
		ContractId:           r.ContractID,
		ReviewerId:           r.ReviewerID.String(),
		RevieweeId:           r.RevieweeID.String(),
		ReviewerRole:         r.ReviewerRole,
		Rating:               r.Rating,
		Title:                r.Title,
		Comment:              r.Comment,
		CreatedAtUnixSeconds: r.CreatedAt.Unix(),
	}
	if r.UpdatedAt != nil {
		unix := r.UpdatedAt.Unix()
		p.UpdatedAtUnixSeconds = &unix
	}
	if r.ReplyComment != nil {
		p.ReplyComment = r.ReplyComment
	}
	if r.RepliedAt != nil {
		unix := r.RepliedAt.Unix()
		p.RepliedAtUnixSeconds = &unix
	}
	return p
}

func toStatus(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := status.FromError(err); ok {
		return err
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "not found"):
		return status.Error(codes.NotFound, err.Error())
	case strings.Contains(msg, "required"), strings.Contains(msg, "invalid"),
		strings.Contains(msg, "too long"), strings.Contains(msg, "must"),
		strings.Contains(msg, "already reviewed"):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
