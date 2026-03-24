package grpcadapter

import (
	"context"
	"strings"
	"time"

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
	GetUserUC             *application.GetUser
	GetProfileUC          *application.GetProfile
	GetPublicProfileUC    *application.GetPublicProfile
	UpdateProfileUC       *application.UpdateProfile
	DeleteProfileUC       *application.DeleteProfile
	GetOnboardingStatusUC *application.GetOnboardingStatus
	UpdateAccountStatusUC *application.UpdateAccountStatus
	UploadAvatarUC        *application.UploadAvatar
	GetAvatarUC           *application.GetAvatar
	RemoveAvatarUC        *application.RemoveAvatar
	CapabilityPolicy      CapabilityPolicy
}

type CapabilityPolicy struct {
	MinSkillsForDiscovery        int
	RequireVerifiedForWithdraw   bool
	RequirePublicForDiscovery    bool
	RequireHeadlineForFreelancer bool
	RequireCompanyNameForClient  bool
	AllowMessagingWhenSuspended  bool
}

func (p CapabilityPolicy) withDefaults() CapabilityPolicy {
	if p.MinSkillsForDiscovery < 0 {
		p.MinSkillsForDiscovery = 0
	}
	if !p.RequireVerifiedForWithdraw && !p.RequirePublicForDiscovery && !p.RequireHeadlineForFreelancer && !p.RequireCompanyNameForClient && !p.AllowMessagingWhenSuspended && p.MinSkillsForDiscovery == 0 {
		return CapabilityPolicy{
			MinSkillsForDiscovery:        1,
			RequireVerifiedForWithdraw:   true,
			RequirePublicForDiscovery:    true,
			RequireHeadlineForFreelancer: true,
			RequireCompanyNameForClient:  true,
			AllowMessagingWhenSuspended:  false,
		}
	}
	return p
}

func NewUserServer(
	createProfile *application.CreateProfile,
	getUser *application.GetUser,
	getProfile *application.GetProfile,
	getPublicProfile *application.GetPublicProfile,
	updateProfile *application.UpdateProfile,
	deleteProfile *application.DeleteProfile,
	getOnboardingStatus *application.GetOnboardingStatus,
	updateAccountStatus *application.UpdateAccountStatus,
	uploadAvatar *application.UploadAvatar,
	getAvatar *application.GetAvatar,
	removeAvatar *application.RemoveAvatar,
	capabilityPolicy CapabilityPolicy,
) *UserServer {
	return &UserServer{
		CreateProfileUC:       createProfile,
		GetUserUC:             getUser,
		GetProfileUC:          getProfile,
		GetPublicProfileUC:    getPublicProfile,
		UpdateProfileUC:       updateProfile,
		DeleteProfileUC:       deleteProfile,
		GetOnboardingStatusUC: getOnboardingStatus,
		UpdateAccountStatusUC: updateAccountStatus,
		UploadAvatarUC:        uploadAvatar,
		GetAvatarUC:           getAvatar,
		RemoveAvatarUC:        removeAvatar,
		CapabilityPolicy:      capabilityPolicy.withDefaults(),
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
			VerificationStatus: mapVerificationStatusFromProto(req.GetClient().VerificationStatus),
		}
	}
	var freelancer *domain.FreelancerProfile
	if req.GetFreelancer() != nil {
		var lastActiveAt *time.Time
		if req.GetFreelancer().LastActiveAtUnix > 0 {
			t := fromUnix(req.GetFreelancer().LastActiveAtUnix)
			lastActiveAt = &t
		}
		freelancer = &domain.FreelancerProfile{
			Headline:           req.GetFreelancer().Headline,
			Bio:                req.GetFreelancer().Bio,
			Skills:             req.GetFreelancer().Skills,
			ExperienceLevel:    req.GetFreelancer().ExperienceLevel,
			Rating:             req.GetFreelancer().Rating,
			VerificationStatus: mapVerificationStatusFromProto(req.GetFreelancer().VerificationStatus),
			Reputation: domain.Reputation{
				JobSuccessScore:  req.GetFreelancer().GetReputation().GetJobSuccessScore(),
				AvgRating:        req.GetFreelancer().GetReputation().GetAvgRating(),
				TotalReviews:     req.GetFreelancer().GetReputation().GetTotalReviews(),
				TotalJobs:        req.GetFreelancer().GetReputation().GetTotalJobs(),
				TotalEarningsUSD: req.GetFreelancer().GetReputation().GetTotalEarningsUsd(),
			},
			HourlyRate:   req.GetFreelancer().HourlyRate,
			Availability: mapAvailabilityFromProto(req.GetFreelancer().Availability),
			Location:     req.GetFreelancer().Location,
			LastActiveAt: lastActiveAt,
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

func (s *UserServer) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.GetUserUC.Execute(ctx, application.GetUserInput{UserID: userID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.GetUserResponse{User: toProtoUser(out.Profile)}, nil
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
	return &userv1.GetProfileResponse{Profile: s.toProtoProfile(out.Profile, out.Client, out.Freelancer)}, nil
}

func (s *UserServer) GetPublicProfile(ctx context.Context, req *userv1.GetPublicProfileRequest) (*userv1.GetPublicProfileResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.GetPublicProfileUC.Execute(ctx, application.GetPublicProfileInput{UserID: userID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.GetPublicProfileResponse{Profile: toProtoPublicProfile(out.Profile, out.Freelancer)}, nil
}

func (s *UserServer) UpdateProfile(ctx context.Context, req *userv1.UpdateProfileRequest) (*userv1.UpdateProfileResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	var availability *string
	if req.Availability != nil {
		value := mapAvailabilityFromProto(req.GetAvailability())
		availability = &value
	}

	out, err := s.UpdateProfileUC.Execute(ctx, application.UpdateProfileInput{
		UserID:           userID,
		DisplayName:      req.DisplayName,
		AvatarURL:        req.AvatarUrl,
		Language:         req.Language,
		ContactEmail:     req.ContactEmail,
		ContactPhone:     req.ContactPhone,
		Bio:              req.Bio,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		CompanyName:      req.CompanyName,
		BillingAddress:   req.BillingAddress,
		TaxID:            req.TaxId,
		Headline:         req.Headline,
		Skills:           req.Skills,
		ExperienceLevel:  req.ExperienceLevel,
		HourlyRate:       req.HourlyRate,
		Availability:     availability,
		Location:         req.Location,
		LastActiveAtUnix: req.LastActiveAtUnix,
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &userv1.UpdateProfileResponse{
		Profile: s.toProtoProfile(out.Profile, out.Client, out.Freelancer),
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
		Steps: toProtoOnboardingSteps(out.Steps),
	}, nil
}

func (s *UserServer) UpdateAccountStatus(ctx context.Context, req *userv1.UpdateAccountStatusRequest) (*userv1.UpdateAccountStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	var suspensionReason *string
	if req.SuspensionReason != nil {
		s := req.GetSuspensionReason()
		suspensionReason = &s
	}

	out, err := s.UpdateAccountStatusUC.Execute(ctx, application.UpdateAccountStatusInput{
		UserID:           userID,
		Status:           req.GetStatus().String(),
		SuspensionReason: suspensionReason,
		Visibility:       req.GetVisibility().String(),
	})
	if err != nil {
		return nil, toStatus(err)
	}

	return &userv1.UpdateAccountStatusResponse{Profile: s.toProtoProfile(out.Profile, out.Client, out.Freelancer)}, nil
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

func (s *UserServer) CreatePortfolioItem(context.Context, *userv1.CreatePortfolioItemRequest) (*userv1.CreatePortfolioItemResponse, error) {
	return nil, status.Error(codes.Unimplemented, "CreatePortfolioItem not implemented")
}

func (s *UserServer) UpdatePortfolioItem(context.Context, *userv1.UpdatePortfolioItemRequest) (*userv1.UpdatePortfolioItemResponse, error) {
	return nil, status.Error(codes.Unimplemented, "UpdatePortfolioItem not implemented")
}

func (s *UserServer) DeletePortfolioItem(context.Context, *userv1.DeletePortfolioItemRequest) (*userv1.DeletePortfolioItemResponse, error) {
	return nil, status.Error(codes.Unimplemented, "DeletePortfolioItem not implemented")
}

func (s *UserServer) ListMyPortfolioItems(context.Context, *userv1.ListMyPortfolioItemsRequest) (*userv1.ListMyPortfolioItemsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListMyPortfolioItems not implemented")
}

func (s *UserServer) ListPublicPortfolioItems(context.Context, *userv1.ListPublicPortfolioItemsRequest) (*userv1.ListPublicPortfolioItemsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListPublicPortfolioItems not implemented")
}

func (s *UserServer) ReorderPortfolioItems(context.Context, *userv1.ReorderPortfolioItemsRequest) (*userv1.ReorderPortfolioItemsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ReorderPortfolioItems not implemented")
}

func (s *UserServer) toProtoProfile(p domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) *userv1.Profile {
	out := &userv1.Profile{
		Id:               p.ID,
		UserId:           p.UserID.String(),
		Role:             p.Role,
		FirstName:        p.FirstName,
		LastName:         p.LastName,
		DisplayName:      p.DisplayName,
		AvatarUrl:        p.AvatarURL,
		Language:         p.Language,
		ContactEmail:     p.ContactEmail,
		ContactPhone:     p.ContactPhone,
		Bio:              p.Bio,
		Deleted:          p.DeletedAt != nil,
		Status:           mapAccountStatusToProto(p.AccountStatus),
		SuspensionReason: p.SuspensionReason,
		Visibility:       mapVisibilityToProto(p.Visibility),
		Capabilities:     toProtoCapabilities(s.CapabilityPolicy, p, client, freelancer),
	}
	if client != nil {
		out.Client = &userv1.ClientProfileInput{
			CompanyName:        client.CompanyName,
			BillingAddress:     client.BillingAddress,
			TaxId:              client.TaxID,
			VerificationStatus: mapVerificationStatusToProto(client.VerificationStatus),
		}
	}
	if freelancer != nil {
		var lastActiveUnix int64
		if freelancer.LastActiveAt != nil {
			lastActiveUnix = freelancer.LastActiveAt.Unix()
		}
		out.Freelancer = &userv1.FreelancerProfileInput{
			Headline:           freelancer.Headline,
			Bio:                freelancer.Bio,
			Skills:             freelancer.Skills,
			ExperienceLevel:    freelancer.ExperienceLevel,
			Rating:             freelancer.Rating,
			VerificationStatus: mapVerificationStatusToProto(freelancer.VerificationStatus),
			Reputation: &userv1.Reputation{
				JobSuccessScore:  freelancer.Reputation.JobSuccessScore,
				AvgRating:        freelancer.Reputation.AvgRating,
				TotalReviews:     freelancer.Reputation.TotalReviews,
				TotalJobs:        freelancer.Reputation.TotalJobs,
				TotalEarningsUsd: freelancer.Reputation.TotalEarningsUSD,
			},
			HourlyRate:       freelancer.HourlyRate,
			Availability:     mapAvailabilityToProto(freelancer.Availability),
			Location:         freelancer.Location,
			LastActiveAtUnix: lastActiveUnix,
		}
	}
	return out
}

func toProtoUser(p domain.Profile) *userv1.User {
	return &userv1.User{
		UserId:        p.UserID.String(),
		Role:          p.Role,
		Status:        mapAccountStatusToProto(p.AccountStatus),
		Visibility:    mapVisibilityToProto(p.Visibility),
		FirstName:     p.FirstName,
		LastName:      p.LastName,
		DisplayName:   p.DisplayName,
		AvatarUrl:     p.AvatarURL,
		CreatedAtUnix: p.CreatedAt.Unix(),
		UpdatedAtUnix: p.UpdatedAt.Unix(),
	}
}

func toProtoPublicProfile(p domain.Profile, freelancer *domain.FreelancerProfile) *userv1.PublicProfile {
	if freelancer == nil {
		return &userv1.PublicProfile{UserId: p.UserID.String()}
	}
	var lastActiveUnix int64
	if freelancer.LastActiveAt != nil {
		lastActiveUnix = freelancer.LastActiveAt.Unix()
	}
	return &userv1.PublicProfile{
		UserId:          p.UserID.String(),
		DisplayName:     p.DisplayName,
		AvatarUrl:       p.AvatarURL,
		Headline:        freelancer.Headline,
		Bio:             freelancer.Bio,
		Skills:          freelancer.Skills,
		ExperienceLevel: freelancer.ExperienceLevel,
		Reputation: &userv1.Reputation{
			JobSuccessScore:  freelancer.Reputation.JobSuccessScore,
			AvgRating:        freelancer.Reputation.AvgRating,
			TotalReviews:     freelancer.Reputation.TotalReviews,
			TotalJobs:        freelancer.Reputation.TotalJobs,
			TotalEarningsUsd: freelancer.Reputation.TotalEarningsUSD,
		},
		HourlyRate:         freelancer.HourlyRate,
		Availability:       mapAvailabilityToProto(freelancer.Availability),
		Location:           freelancer.Location,
		LastActiveAtUnix:   lastActiveUnix,
		VerificationStatus: mapVerificationStatusToProto(freelancer.VerificationStatus),
	}
}

func toProtoOnboardingSteps(steps []application.OnboardingStep) []*userv1.OnboardingStep {
	out := make([]*userv1.OnboardingStep, 0, len(steps))
	for _, step := range steps {
		status := userv1.OnboardingStepStatus_ONBOARDING_STEP_STATUS_NOT_STARTED
		if step.Completed {
			status = userv1.OnboardingStepStatus_ONBOARDING_STEP_STATUS_COMPLETED
		}
		out = append(out, &userv1.OnboardingStep{Key: step.Key, Status: status})
	}
	return out
}

func mapAccountStatusToProto(status string) userv1.AccountStatus {
	switch strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(status), "ACCOUNT_STATUS_")) {
	case "SUSPENDED":
		return userv1.AccountStatus_ACCOUNT_STATUS_SUSPENDED
	case "DELETED":
		return userv1.AccountStatus_ACCOUNT_STATUS_DELETED
	default:
		return userv1.AccountStatus_ACCOUNT_STATUS_ACTIVE
	}
}

func mapVisibilityToProto(visibility string) userv1.ProfileVisibility {
	switch strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(visibility), "PROFILE_VISIBILITY_")) {
	case "PRIVATE":
		return userv1.ProfileVisibility_PROFILE_VISIBILITY_PRIVATE
	default:
		return userv1.ProfileVisibility_PROFILE_VISIBILITY_PUBLIC
	}
}

func mapVerificationStatusToProto(status string) userv1.VerificationStatus {
	switch strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(status), "VERIFICATION_STATUS_")) {
	case "VERIFIED":
		return userv1.VerificationStatus_VERIFICATION_STATUS_VERIFIED
	case "REJECTED":
		return userv1.VerificationStatus_VERIFICATION_STATUS_REJECTED
	case "EXPIRED":
		return userv1.VerificationStatus_VERIFICATION_STATUS_EXPIRED
	default:
		return userv1.VerificationStatus_VERIFICATION_STATUS_PENDING
	}
}

func mapVerificationStatusFromProto(status userv1.VerificationStatus) string {
	switch status {
	case userv1.VerificationStatus_VERIFICATION_STATUS_VERIFIED:
		return domain.VerificationStatusVerified
	case userv1.VerificationStatus_VERIFICATION_STATUS_REJECTED:
		return domain.VerificationStatusRejected
	case userv1.VerificationStatus_VERIFICATION_STATUS_EXPIRED:
		return domain.VerificationStatusExpired
	default:
		return domain.VerificationStatusPending
	}
}

func mapAvailabilityToProto(availability string) userv1.Availability {
	switch strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(availability), "AVAILABILITY_")) {
	case "FULL_TIME":
		return userv1.Availability_AVAILABILITY_FULL_TIME
	case "PART_TIME":
		return userv1.Availability_AVAILABILITY_PART_TIME
	case "UNAVAILABLE":
		return userv1.Availability_AVAILABILITY_UNAVAILABLE
	default:
		return userv1.Availability_AVAILABILITY_AS_NEEDED
	}
}

func mapAvailabilityFromProto(availability userv1.Availability) string {
	switch availability {
	case userv1.Availability_AVAILABILITY_FULL_TIME:
		return domain.AvailabilityFullTime
	case userv1.Availability_AVAILABILITY_PART_TIME:
		return domain.AvailabilityPartTime
	case userv1.Availability_AVAILABILITY_UNAVAILABLE:
		return domain.AvailabilityUnavailable
	default:
		return domain.AvailabilityAsNeeded
	}
}

func fromUnix(unixSec int64) time.Time {
	return time.Unix(unixSec, 0).UTC()
}

func toProtoCapabilities(policy CapabilityPolicy, p domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) *userv1.CapabilityFlags {
	active := strings.EqualFold(strings.TrimPrefix(strings.TrimSpace(p.AccountStatus), "ACCOUNT_STATUS_"), domain.AccountStatusActive)
	publicVisible := strings.EqualFold(strings.TrimPrefix(strings.TrimSpace(p.Visibility), "PROFILE_VISIBILITY_"), domain.ProfileVisibilityPublic)

	isFreelancer := p.Role == domain.RoleFreelancer
	isClient := p.Role == domain.RoleClient

	freelancerVerified := freelancer != nil && strings.EqualFold(strings.TrimSpace(freelancer.VerificationStatus), domain.VerificationStatusVerified)
	hasHeadline := freelancer != nil && strings.TrimSpace(freelancer.Headline) != ""
	hasEnoughSkills := freelancer != nil && len(freelancer.Skills) >= policy.MinSkillsForDiscovery
	hasDiscoverableFreelancerProfile := freelancer != nil && hasEnoughSkills && (!policy.RequireHeadlineForFreelancer || hasHeadline)
	hasDiscoverableClientProfile := client != nil && (!policy.RequireCompanyNameForClient || strings.TrimSpace(client.CompanyName) != "")

	canApplyJobs := active && isFreelancer
	canPostJobs := active && isClient
	canWithdrawFunds := active && isFreelancer && (!policy.RequireVerifiedForWithdraw || freelancerVerified)
	canMessage := active || policy.AllowMessagingWhenSuspended
	canBeDiscovered := active && (!policy.RequirePublicForDiscovery || publicVisible) && (hasDiscoverableFreelancerProfile || hasDiscoverableClientProfile)

	return &userv1.CapabilityFlags{
		CanApplyJobs:     canApplyJobs,
		CanPostJobs:      canPostJobs,
		CanWithdrawFunds: canWithdrawFunds,
		CanMessage:       canMessage,
		CanBeDiscovered:  canBeDiscovered,
	}
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
