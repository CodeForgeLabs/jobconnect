package events

import (
	"context"

	shared "jobconnect/events"
	"jobconnect/review/internal/domain"
)

type ReviewPublisher struct {
	publisher *shared.Publisher
}

func NewReviewPublisher(p *shared.Publisher) *ReviewPublisher {
	return &ReviewPublisher{publisher: p}
}

func (p *ReviewPublisher) PublishReviewCreated(ctx context.Context, review domain.Review) error {
	env, err := shared.NewEnvelope("review.created", review.RevieweeID.String(), "review-service", 1, map[string]any{
		"review_id":   review.ID,
		"contract_id": review.ContractID,
		"reviewer_id": review.ReviewerID.String(),
		"reviewee_id": review.RevieweeID.String(),
	}, "review-created:"+review.RevieweeID.String(), review.RevieweeID.String())
	if err != nil {
		return err
	}
	return p.publisher.Publish(ctx, env)
}

func (p *ReviewPublisher) PublishReviewUpdated(ctx context.Context, review domain.Review) error {
	env, err := shared.NewEnvelope("review.updated", review.RevieweeID.String(), "review-service", 1, map[string]any{
		"review_id":   review.ID,
		"contract_id": review.ContractID,
		"reviewee_id": review.RevieweeID.String(),
	}, "review-updated:"+review.RevieweeID.String(), review.RevieweeID.String())
	if err != nil {
		return err
	}
	return p.publisher.Publish(ctx, env)
}

func (p *ReviewPublisher) PublishReviewDeleted(ctx context.Context, reviewID int64, revieweeID string) error {
	env, err := shared.NewEnvelope("review.deleted", revieweeID, "review-service", 1, map[string]any{
		"review_id":   reviewID,
		"reviewee_id": revieweeID,
	}, "review-deleted:"+revieweeID, revieweeID)
	if err != nil {
		return err
	}
	return p.publisher.Publish(ctx, env)
}
