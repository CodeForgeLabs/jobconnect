package reviewgrpc

import (
	"context"

	"google.golang.org/grpc"

	"jobconnect/recommendation/internal/domain"
	reviewv1 "jobconnect/review/gen/review/v1"
)

type Client struct {
	grpcClient reviewv1.ReviewServiceClient
}

func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{grpcClient: reviewv1.NewReviewServiceClient(conn)}
}

func (c *Client) GetUserRatingSummary(ctx context.Context, userID string) (domain.RatingSummary, error) {
	resp, err := c.grpcClient.GetUserRatingSummary(ctx, &reviewv1.GetUserRatingSummaryRequest{UserId: userID})
	if err != nil {
		return domain.RatingSummary{}, err
	}
	return domain.RatingSummary{
		AverageRating: resp.GetAverageRating(),
		TotalReviews:  resp.GetTotalReviews(),
	}, nil
}
