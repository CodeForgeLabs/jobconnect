package grpcadapter

import (
	"context"
	"fmt"
	reviewsv1 "jobconnect/reviews/gen/reviews/v1"
	"jobconnect/reviews/internal/applications"
)

type ReviewServer struct {
	reviewsv1.UnimplementedReviewServiceServer
	createReview         *applications.CreateReview
	deleteReview         *applications.DeleteReview
	getReview            *applications.GetReview
	listReviews          *applications.ListReviews
	updateReview         *applications.UpdateReview
	getContractsUsers    *applications.GetContractUsers
	getUserRatingSummary *applications.GetUserRatingSummary
}

func NewReviewServer(
	createReview *applications.CreateReview,
	deleteReview *applications.DeleteReview,
	getReview *applications.GetReview,
	listReviews *applications.ListReviews,
	updateReview *applications.UpdateReview,
	getContractsUsers *applications.GetContractUsers,
	getUserRatingSummary *applications.GetUserRatingSummary,
) *ReviewServer {
	return &ReviewServer{
		createReview:         createReview,
		deleteReview:         deleteReview,
		getReview:            getReview,
		listReviews:          listReviews,
		updateReview:         updateReview,
		getContractsUsers:    getContractsUsers,
		getUserRatingSummary: getUserRatingSummary,
	}
}

func (s *ReviewServer) CreateReview(
	ctx context.Context,
	req *reviewsv1.CreateReviewRequest,
) (*reviewsv1.CreateReviewResponse, error) {

	contractId := req.GetContractId()
	if contractId <= 0 {
		return nil, fmt.Errorf("invalid contract ID: %d", contractId)
	}

	// ⚠️ TODO: CALL CONTRACT SERVICE HERE
	contract, err := s.getContractsUsers.Execute(ctx, applications.GetContractUsersInput{
		ContractID: contractId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get contract users: %w", err)
	}
	// clientID := contract.ClientId
	// freelancerID := contract.FreelancerId

	input := applications.CreateReviewInput{
		ContractID:   contractId,
		ClientID:     contract.ClientID,
		FreelancerID: contract.FreelancerID,
		ReviewerRole: mapRole(req.GetReviewerRole()),
		Rating:       int(req.GetRating()),
		Title:        req.GetTitle(),
		Comment:      req.GetComment(),
	}

	output, err := s.createReview.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &reviewsv1.CreateReviewResponse{
		Review: mapReview(output.Review),
	}, nil
}
func (s *ReviewServer) GetReview(
	ctx context.Context,
	req *reviewsv1.GetReviewRequest,
) (*reviewsv1.GetReviewResponse, error) {

	if req.GetId() <= 0 {
		return nil, fmt.Errorf("invalid review ID: %d", req.GetId())
	}
	getReviewInput := applications.GetReviewInput{
		ID: req.GetId(),
	}
	output, err := s.getReview.Execute(ctx, getReviewInput)
	if err != nil {
		return nil, err
	}

	return &reviewsv1.GetReviewResponse{
		Review: mapReview(output.Review),
	}, nil
}
func (s *ReviewServer) UpdateReview(
	ctx context.Context,
	req *reviewsv1.UpdateReviewRequest,
) (*reviewsv1.UpdateReviewResponse, error) {

	if req.GetId() <= 0 {
		return nil, fmt.Errorf("invalid review ID: %d", req.GetId())
	}

	userID := getUserID(ctx)

	output, err := s.updateReview.Execute(ctx, applications.UpdateReviewInput{
		ID:      req.GetId(),
		Rating:  int(req.GetRating()),
		Title:   req.GetTitle(),
		Comment: req.GetComment(),
		UserID:  userID,
	})
	if err != nil {
		return nil, err
	}

	return &reviewsv1.UpdateReviewResponse{
		Review: mapReview(output.Review),
	}, nil
}
func (s *ReviewServer) DeleteReview(
	ctx context.Context,
	req *reviewsv1.DeleteReviewRequest,
) (*reviewsv1.DeleteReviewResponse, error) {

	if req.GetId() <= 0 {
		return nil, fmt.Errorf("invalid review ID: %d", req.GetId())
	}

	userID := getUserID(ctx)

	output, err := s.deleteReview.Execute(ctx, applications.DeleteReviewInput{
		ID:     req.GetId(),
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	return &reviewsv1.DeleteReviewResponse{
		Success: output.Success,
	}, nil
}
func (s *ReviewServer) ListReviewsByUser(
	ctx context.Context,
	req *reviewsv1.ListReviewsByUserRequest,
) (*reviewsv1.ListReviewsResponse, error) {

	if req.GetUserId() == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	output, err := s.listReviews.Execute(ctx, applications.ListReviewsInput{
		UserID: req.GetUserId(),
		Role:   mapRole(req.GetRole()),
		Limit:  int(req.GetLimit()),
		Offset: int(req.GetOffset()),
	})
	if err != nil {
		return nil, err
	}

	reviews := make([]*reviewsv1.Review, 0, len(output.Reviews))
	for _, r := range output.Reviews {
		reviews = append(reviews, mapReview(r))
	}

	return &reviewsv1.ListReviewsResponse{
		Reviews: reviews,
	}, nil
}

func (s *ReviewServer) GetUserRatingSummary(
	ctx context.Context,
	req *reviewsv1.GetUserRatingSummaryRequest,
) (*reviewsv1.GetUserRatingSummaryResponse, error) {

	if req.GetUserId() == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	avg, err := s.getUserRatingSummary.Execute(ctx, applications.GetUserRatingSummaryInput{
		UserID: req.GetUserId(),
	})
	if err != nil {
		return nil, err
	}

	return &reviewsv1.GetUserRatingSummaryResponse{
		AverageRating: avg.AverageRating,
		TotalReviews:  avg.TotalReviews,
	}, nil
}
