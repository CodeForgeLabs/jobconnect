package grpcadapter

import (
	"context"
	"jobconnect/reviews/internal/domain"

	reviewsv1 "jobconnect/reviews/gen/reviews/v1"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func mapReview(r domain.Review) *reviewsv1.Review {
	return &reviewsv1.Review{
		Id:           r.ID,
		ContractId:   r.ContractID,
		ClientId:     r.ClientID,
		FreelancerId: r.FreelancerID,
		ReviewerRole: mapProtoRole(r.ReviewerRole),
		Rating:       int32(r.Rating),
		Title:        r.Title,
		Comment:      r.Comment,
		CreatedAt:    timestamppb.New(r.CreatedAt),
		UpdatedAt:    timestamppb.New(r.UpdatedAt),
	}
}

func mapProtoRole(r domain.ReviewerRole) reviewsv1.ReviewerRole {
	switch r {
	case domain.RoleClient:
		return reviewsv1.ReviewerRole_CLIENT
	case domain.RoleFreelancer:
		return reviewsv1.ReviewerRole_FREELANCER
	default:
		return reviewsv1.ReviewerRole_REVIEWER_ROLE_UNSPECIFIED
	}
}

func mapRole(r reviewsv1.ReviewerRole) domain.ReviewerRole {
	switch r {
	case reviewsv1.ReviewerRole_CLIENT:
		return domain.RoleClient
	case reviewsv1.ReviewerRole_FREELANCER:
		return domain.RoleFreelancer
	default:
		return domain.RoleClient
	}

}

func getUserID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get("user_id")
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func getUserRole(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get("role")
	if len(values) == 0 {
		return ""
	}

	return values[0]
}
