package grpcadapter

import (
	"context"
	"strings"

	userv1 "jobconnect/user/gen/user/v1"
	"jobconnect/user/internal/application"
	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServer struct {
	userv1.UnimplementedUserServiceServer
	CreateProfileUC *application.CreateProfile
}

func NewUserServer(createProfile *application.CreateProfile) *UserServer {
	return &UserServer{CreateProfileUC: createProfile}
}

func (s *UserServer) CreateProfile(ctx context.Context, req *userv1.CreateProfileRequest) (*userv1.CreateProfileResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	var client *domain.ClientProfile
	if req.GetClient() != nil {
		client = &domain.ClientProfile{
			CompanyName:        req.GetClient().CompanyName,
			BillingAddress:     req.GetClient().BillingAddress,
			TaxID:              req.GetClient().TaxId,
			VerificationStatus: req.GetClient().VerificationStatus,
		}
	}
	var freelancer *domain.FreelancerProfile
	if req.GetFreelancer() != nil {
		freelancer = &domain.FreelancerProfile{
			Headline:           req.GetFreelancer().Headline,
			Bio:                req.GetFreelancer().Bio,
			Skills:             req.GetFreelancer().Skills,
			ExperienceLevel:    req.GetFreelancer().ExperienceLevel,
			Rating:             req.GetFreelancer().Rating,
			VerificationStatus: req.GetFreelancer().VerificationStatus,
		}
	}

	out, err := s.CreateProfileUC.Execute(ctx, application.CreateProfileInput{
		UserID:      userID,
		Role:        req.Role,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DisplayName: req.DisplayName,
		AvatarURL:   req.AvatarUrl,
		Client:      client,
		Freelancer:  freelancer,
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &userv1.CreateProfileResponse{Success: true, ProfileId: out.ProfileID}, nil
}

func toStatus(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case contains(msg, "required"), contains(msg, "invalid"), contains(msg, "not allowed"):
		return status.Error(codes.InvalidArgument, msg)
	default:
		// Surface the underlying error for debugging via gRPC clients (e.g., Postman).
		return status.Error(codes.Internal, msg)
	}
}

func contains(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}
