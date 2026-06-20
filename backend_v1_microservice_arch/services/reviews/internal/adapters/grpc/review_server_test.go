package grpcadapter

import (
	"context"
	"testing"

	reviewsv1 "jobconnect/reviews/gen/reviews/v1"
	"jobconnect/reviews/internal/applications"
	"jobconnect/reviews/internal/domain"

	"google.golang.org/grpc/metadata"
)

func TestDeleteReviewUsesCallerMetadataUserID(t *testing.T) {
	repo := &reviewRepoStub{
		review: domain.Review{
			ID:           7,
			ClientID:     "caller-123",
			FreelancerID: "freelancer-456",
			ReviewerRole: domain.RoleClient,
		},
	}
	server := NewReviewServer(nil, &applications.DeleteReview{Reviews: repo}, nil, nil, nil, nil, nil)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("user_id", "caller-123"))

	resp, err := server.DeleteReview(ctx, &reviewsv1.DeleteReviewRequest{Id: 7})
	if err != nil {
		t.Fatalf("DeleteReview returned error: %v", err)
	}
	if !resp.GetSuccess() {
		t.Fatalf("expected successful delete response")
	}
	if repo.deletedID != 7 {
		t.Fatalf("expected review 7 to be deleted, got %d", repo.deletedID)
	}
}

type reviewRepoStub struct {
	review    domain.Review
	deletedID int64
}

func (r *reviewRepoStub) Create(context.Context, domain.Review) (domain.Review, error) {
	return domain.Review{}, nil
}

func (r *reviewRepoStub) GetByID(context.Context, int64) (domain.Review, error) {
	return r.review, nil
}

func (r *reviewRepoStub) Update(context.Context, domain.Review) (domain.Review, error) {
	return domain.Review{}, nil
}

func (r *reviewRepoStub) Delete(_ context.Context, id int64) error {
	r.deletedID = id
	return nil
}

func (r *reviewRepoStub) ListByUser(context.Context, string, domain.ReviewerRole, int, int) ([]domain.Review, error) {
	return nil, nil
}

func (r *reviewRepoStub) GetContractUsers(context.Context, int64) (applications.GetContractUsersOutput, error) {
	return applications.GetContractUsersOutput{}, nil
}

func (r *reviewRepoStub) GetUserRatingSummary(context.Context, string) (float64, int64, error) {
	return 0, 0, nil
}
