package grpcadapter

import (
	"context"
	"strings"

	userv1 "jobconnect/user/gen/user"
	"jobconnect/user/internal/application"
	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServer struct {
	userv1.UnimplementedUserServiceServer
	CreateProfileUC       *application.CreateProfile
	GetProfileUC          *application.GetProfile
	UpdateProfileUC       *application.UpdateProfile
	DeleteProfileUC       *application.DeleteProfile
	GetOnboardingStatusUC *application.GetOnboardingStatus
	UploadAvatarUC        *application.UploadAvatar
	GetAvatarUC           *application.GetAvatar
	RemoveAvatarUC        *application.RemoveAvatar
}

func NewUserServer(
	createProfile *application.CreateProfile,
	getProfile *application.GetProfile,
	updateProfile *application.UpdateProfile,
	deleteProfile *application.DeleteProfile,
	getOnboardingStatus *application.GetOnboardingStatus,
	uploadAvatar *application.UploadAvatar,
	getAvatar *application.GetAvatar,
	removeAvatar *application.RemoveAvatar,
) *UserServer {
	return &UserServer{
		CreateProfileUC:       createProfile,
		GetProfileUC:          getProfile,
		UpdateProfileUC:       updateProfile,
		DeleteProfileUC:       deleteProfile,
		GetOnboardingStatusUC: getOnboardingStatus,
		UploadAvatarUC:        uploadAvatar,
		GetAvatarUC:           getAvatar,
		RemoveAvatarUC:        removeAvatar,
	}
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

func (s *UserServer) GetProfile(ctx context.Context, req *userv1.GetProfileRequest) (*userv1.GetProfileResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.GetProfileUC.Execute(ctx, application.GetProfileInput{UserID: userID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.GetProfileResponse{Profile: toProtoProfile(out.Profile, out.Client, out.Freelancer)}, nil
}

func (s *UserServer) UpdateProfile(ctx context.Context, req *userv1.UpdateProfileRequest) (*userv1.UpdateProfileResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	out, err := s.UpdateProfileUC.Execute(ctx, application.UpdateProfileInput{
		UserID:          userID,
		DisplayName:     req.DisplayName,
		AvatarURL:       req.AvatarUrl,
		Language:        req.Language,
		ContactEmail:    req.ContactEmail,
		ContactPhone:    req.ContactPhone,
		Bio:             req.Bio,
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		CompanyName:     req.CompanyName,
		BillingAddress:  req.BillingAddress,
		TaxID:           req.TaxId,
		Headline:        req.Headline,
		Skills:          req.Skills,
		ExperienceLevel: req.ExperienceLevel,
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &userv1.UpdateProfileResponse{
		Profile: toProtoProfile(out.Profile, out.Client, out.Freelancer),
		Completeness: &userv1.ProfileCompleteness{
			Percent:               out.Completeness,
			MissingRequiredFields: out.Missing,
		},
	}, nil
}

func (s *UserServer) DeleteProfile(ctx context.Context, req *userv1.DeleteProfileRequest) (*userv1.DeleteProfileResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.DeleteProfileUC.Execute(ctx, application.DeleteProfileInput{UserID: userID, HardDelete: req.HardDelete})
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.DeleteProfileResponse{Deleted: out.Deleted}, nil
}

func (s *UserServer) GetOnboardingStatus(ctx context.Context, req *userv1.GetOnboardingStatusRequest) (*userv1.GetOnboardingStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.GetOnboardingStatusUC.Execute(ctx, application.GetOnboardingStatusInput{UserID: userID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.GetOnboardingStatusResponse{
		Completeness: &userv1.ProfileCompleteness{
			Percent:               out.Percent,
			MissingRequiredFields: out.Missing,
		},
	}, nil
}

func (s *UserServer) UploadAvatar(ctx context.Context, req *userv1.UploadAvatarRequest) (*userv1.UploadAvatarResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.UploadAvatarUC.Execute(ctx, application.UploadAvatarInput{
		UserID:      userID,
		FileName:    req.FileName,
		ContentType: req.ContentType,
		Content:     req.Content,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.UploadAvatarResponse{
		AvatarUrl:   out.AvatarURL,
		PreviewUrl:  out.PreviewURL,
		ContentType: out.ContentType,
		SizeBytes:   out.SizeBytes,
		Width:       out.Width,
		Height:      out.Height,
	}, nil
}

func (s *UserServer) GetAvatar(ctx context.Context, req *userv1.GetAvatarRequest) (*userv1.GetAvatarResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.GetAvatarUC.Execute(ctx, application.GetAvatarInput{UserID: userID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.GetAvatarResponse{FileName: out.FileName, ContentType: out.ContentType, Content: out.Content}, nil
}

func (s *UserServer) RemoveAvatar(ctx context.Context, req *userv1.RemoveAvatarRequest) (*userv1.RemoveAvatarResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.RemoveAvatarUC.Execute(ctx, application.RemoveAvatarInput{UserID: userID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.RemoveAvatarResponse{Removed: out.Removed}, nil
}

func toProtoProfile(p domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) *userv1.Profile {
	out := &userv1.Profile{
		Id:           p.ID,
		UserId:       p.UserID.String(),
		Role:         p.Role,
		FirstName:    p.FirstName,
		LastName:     p.LastName,
		DisplayName:  p.DisplayName,
		AvatarUrl:    p.AvatarURL,
		Language:     p.Language,
		ContactEmail: p.ContactEmail,
		ContactPhone: p.ContactPhone,
		Bio:          p.Bio,
		Deleted:      p.DeletedAt != nil,
	}
	if client != nil {
		out.Client = &userv1.ClientProfileInput{
			CompanyName:        client.CompanyName,
			BillingAddress:     client.BillingAddress,
			TaxId:              client.TaxID,
			VerificationStatus: client.VerificationStatus,
		}
	}
	if freelancer != nil {
		out.Freelancer = &userv1.FreelancerProfileInput{
			Headline:           freelancer.Headline,
			Bio:                freelancer.Bio,
			Skills:             freelancer.Skills,
			ExperienceLevel:    freelancer.ExperienceLevel,
			Rating:             freelancer.Rating,
			VerificationStatus: freelancer.VerificationStatus,
		}
	}
	return out
}

func toStatus(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case contains(msg, "not found"):
		return status.Error(codes.NotFound, msg)
	case contains(msg, "required"), contains(msg, "invalid"), contains(msg, "not allowed"):
		return status.Error(codes.InvalidArgument, msg)
	case contains(msg, "unsupported"), contains(msg, "exceeds"), contains(msg, "too small"):
		return status.Error(codes.InvalidArgument, msg)
	default:
		// Surface the underlying error for debugging via gRPC clients (e.g., Postman).
		return status.Error(codes.Internal, msg)
	}
}

func contains(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}
