package grpcadapter

import (
	"context"
	"fmt"
	reviewsv1 "jobconnect/reviews/gen/reviews/v1"
	"jobconnect/reviews/internal/applications"
)

type ReviewServer struct {
	reviewsv1.UnimplementedReviewServiceServer
	createReview *applications.CreateReview
	deleteReview *applications.DeleteReview
	getReview    *applications.GetReview
	listReviews  *applications.ListReviews
	updateReview *applications.UpdateReview
}

func NewReviewServer(
	createReview *applications.CreateReview,
	deleteReview *applications.DeleteReview,
	getReview *applications.GetReview,
	listReviews *applications.ListReviews,
	updateReview *applications.UpdateReview,
) *ReviewServer {
	return &ReviewServer{
		createReview: createReview,
		deleteReview: deleteReview,
		getReview:    getReview,
		listReviews:  listReviews,
		updateReview: updateReview,
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
	// contract := s.contractClient.GetContract(...)
	// clientID := contract.ClientId
	// freelancerID := contract.FreelancerId

	input := applications.CreateReviewInput{
		ContractID:   contractId,
		ClientID:     "550e8400-e29b-41d4-a716-446655440000", // TODO from contract service
		FreelancerID: "550e8400-e29b-41d4-a716-445755440000", // TODO from contract service
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

	// TODO: get user from context (JWT middleware)
	// userID := getUserID(ctx)
	userID := "550e8400-e29b-41d4-a716-446655440000" // TODO: remove hardcoded user ID

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

	// userID := getUserID(ctx)
	userID := "550e8400-e29b-41d4-a716-446655440000" // TODO: remove hardcoded user ID

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
