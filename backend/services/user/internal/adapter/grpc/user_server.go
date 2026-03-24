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
	ProfileDetailsRepo    application.ProfileDetailsRepository
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
	profileDetailsRepo application.ProfileDetailsRepository,
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
		ProfileDetailsRepo:    profileDetailsRepo,
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

func (s *UserServer) GetAccountSettings(ctx context.Context, req *userv1.GetAccountSettingsRequest) (*userv1.GetAccountSettingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	out, err := s.GetProfileUC.Execute(ctx, application.GetProfileInput{UserID: userID})
	if err != nil {
		return nil, toStatus(err)
	}

	settings := &userv1.AccountSettings{
		Language:     out.Profile.Language,
		ContactEmail: out.Profile.ContactEmail,
		ContactPhone: out.Profile.ContactPhone,
	}

	return &userv1.GetAccountSettingsResponse{Settings: settings}, nil
}

func (s *UserServer) UpdateAccountSettings(ctx context.Context, req *userv1.UpdateAccountSettingsRequest) (*userv1.UpdateAccountSettingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	if req.Locale != nil || req.Timezone != nil {
		return nil, status.Error(codes.InvalidArgument, "locale and timezone are not supported yet")
	}

	out, err := s.UpdateProfileUC.Execute(ctx, application.UpdateProfileInput{
		UserID:       userID,
		Language:     req.Language,
		ContactEmail: req.ContactEmail,
		ContactPhone: req.ContactPhone,
	})
	if err != nil {
		return nil, toStatus(err)
	}

	settings := &userv1.AccountSettings{
		Language:     out.Profile.Language,
		ContactEmail: out.Profile.ContactEmail,
		ContactPhone: out.Profile.ContactPhone,
	}

	return &userv1.UpdateAccountSettingsResponse{Settings: settings}, nil
}

func (s *UserServer) GetPrivacySettings(ctx context.Context, req *userv1.GetPrivacySettingsRequest) (*userv1.GetPrivacySettingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	out, err := s.GetProfileUC.Execute(ctx, application.GetProfileInput{UserID: userID})
	if err != nil {
		return nil, toStatus(err)
	}

	settings := &userv1.PrivacySettings{
		Discoverable:   strings.EqualFold(out.Profile.Visibility, "PUBLIC"),
		ShowLastActive: true,
		ShowEarnings:   false,
	}

	return &userv1.GetPrivacySettingsResponse{Settings: settings}, nil
}

func (s *UserServer) UpdatePrivacySettings(ctx context.Context, req *userv1.UpdatePrivacySettingsRequest) (*userv1.UpdatePrivacySettingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	if req.ShowLastActive != nil || req.ShowEarnings != nil {
		return nil, status.Error(codes.InvalidArgument, "show_last_active and show_earnings are not supported yet")
	}

	profileOut, err := s.GetProfileUC.Execute(ctx, application.GetProfileInput{UserID: userID})
	if err != nil {
		return nil, toStatus(err)
	}

	discoverable := strings.EqualFold(profileOut.Profile.Visibility, "PUBLIC")
	if req.Discoverable != nil {
		visibility := "PRIVATE"
		if req.GetDiscoverable() {
			visibility = "PUBLIC"
		}

		statusOut, err := s.UpdateAccountStatusUC.Execute(ctx, application.UpdateAccountStatusInput{
			UserID:           userID,
			Status:           profileOut.Profile.AccountStatus,
			SuspensionReason: stringPtrOrNil(profileOut.Profile.SuspensionReason),
			Visibility:       visibility,
		})
		if err != nil {
			return nil, toStatus(err)
		}
		discoverable = strings.EqualFold(statusOut.Profile.Visibility, "PUBLIC")
	}

	return &userv1.UpdatePrivacySettingsResponse{Settings: &userv1.PrivacySettings{
		Discoverable:   discoverable,
		ShowLastActive: true,
		ShowEarnings:   false,
	}}, nil
}

func (s *UserServer) GetNotificationSettings(ctx context.Context, req *userv1.GetNotificationSettingsRequest) (*userv1.GetNotificationSettingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	if _, err := uuid.Parse(req.GetUserId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	settings := defaultNotificationSettings()
	return &userv1.GetNotificationSettingsResponse{Settings: &settings}, nil
}

func (s *UserServer) UpdateNotificationSettings(ctx context.Context, req *userv1.UpdateNotificationSettingsRequest) (*userv1.UpdateNotificationSettingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	if _, err := uuid.Parse(req.GetUserId()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	settings := defaultNotificationSettings()
	if req.EmailJobAlerts != nil {
		settings.EmailJobAlerts = req.GetEmailJobAlerts()
	}
	if req.EmailMessages != nil {
		settings.EmailMessages = req.GetEmailMessages()
	}
	if req.EmailBilling != nil {
		settings.EmailBilling = req.GetEmailBilling()
	}
	if req.EmailSecurity != nil {
		settings.EmailSecurity = req.GetEmailSecurity()
	}
	if req.PushJobAlerts != nil {
		settings.PushJobAlerts = req.GetPushJobAlerts()
	}
	if req.PushMessages != nil {
		settings.PushMessages = req.GetPushMessages()
	}
	if req.PushBilling != nil {
		settings.PushBilling = req.GetPushBilling()
	}
	if req.PushSecurity != nil {
		settings.PushSecurity = req.GetPushSecurity()
	}

	return &userv1.UpdateNotificationSettingsResponse{Settings: &settings}, nil
}

func (s *UserServer) CreatePortfolioItem(ctx context.Context, req *userv1.CreatePortfolioItemRequest) (*userv1.CreatePortfolioItemResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	item, err := s.ProfileDetailsRepo.CreatePortfolioItem(ctx, userID, toAppPortfolioItem(req))
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.CreatePortfolioItemResponse{Item: toProtoPortfolioItem(item)}, nil
}

func (s *UserServer) UpdatePortfolioItem(ctx context.Context, req *userv1.UpdatePortfolioItemRequest) (*userv1.UpdatePortfolioItemResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	current, err := s.ProfileDetailsRepo.GetPortfolioItem(ctx, userID, req.GetItemId())
	if err != nil {
		return nil, toStatus(err)
	}
	if req.Title != nil {
		current.Title = req.GetTitle()
	}
	if req.Description != nil {
		current.Description = req.GetDescription()
	}
	if req.ProjectUrl != nil {
		current.ProjectURL = req.GetProjectUrl()
	}
	if req.RoleInProject != nil {
		current.RoleInProject = req.GetRoleInProject()
	}
	if req.CompletedAtUnix != nil {
		current.CompletedAt = unixPtr(req.GetCompletedAtUnix())
	}
	if req.Visibility != nil {
		current.Visibility = mapVisibilityFromProto(req.GetVisibility())
	}
	if len(req.GetTags()) > 0 {
		current.Tags = req.GetTags()
	}
	if len(req.GetMedia()) > 0 {
		current.Media = toAppPortfolioMediaSlice(req.GetMedia())
	}
	item, err := s.ProfileDetailsRepo.UpdatePortfolioItem(ctx, userID, req.GetItemId(), current)
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.UpdatePortfolioItemResponse{Item: toProtoPortfolioItem(item)}, nil
}

func (s *UserServer) DeletePortfolioItem(ctx context.Context, req *userv1.DeletePortfolioItemRequest) (*userv1.DeletePortfolioItemResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	deleted, err := s.ProfileDetailsRepo.DeletePortfolioItem(ctx, userID, req.GetItemId())
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.DeletePortfolioItemResponse{Deleted: deleted}, nil
}

func (s *UserServer) ListMyPortfolioItems(ctx context.Context, req *userv1.ListMyPortfolioItemsRequest) (*userv1.ListMyPortfolioItemsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.ProfileDetailsRepo.ListMyPortfolioItems(ctx, userID, req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, toStatus(err)
	}
	items := make([]*userv1.PortfolioItem, 0, len(out.Items))
	for _, item := range out.Items {
		items = append(items, toProtoPortfolioItem(item))
	}
	return &userv1.ListMyPortfolioItemsResponse{Items: items, NextPageToken: out.NextPageToken}, nil
}

func (s *UserServer) ListPublicPortfolioItems(ctx context.Context, req *userv1.ListPublicPortfolioItemsRequest) (*userv1.ListPublicPortfolioItemsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.ProfileDetailsRepo.ListPublicPortfolioItems(ctx, userID, req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, toStatus(err)
	}
	items := make([]*userv1.PortfolioItem, 0, len(out.Items))
	for _, item := range out.Items {
		items = append(items, toProtoPortfolioItem(item))
	}
	return &userv1.ListPublicPortfolioItemsResponse{Items: items, NextPageToken: out.NextPageToken}, nil
}

func (s *UserServer) ReorderPortfolioItems(ctx context.Context, req *userv1.ReorderPortfolioItemsRequest) (*userv1.ReorderPortfolioItemsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	items, err := s.ProfileDetailsRepo.ReorderPortfolioItems(ctx, userID, req.GetOrderedItemIds())
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*userv1.PortfolioItem, 0, len(items))
	for _, item := range items {
		out = append(out, toProtoPortfolioItem(item))
	}
	return &userv1.ReorderPortfolioItemsResponse{Items: out}, nil
}

func (s *UserServer) CreateEmployment(ctx context.Context, req *userv1.CreateEmploymentRequest) (*userv1.CreateEmploymentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	in := application.Employment{CompanyName: req.GetCompanyName(), Title: req.GetTitle(), EmploymentType: req.GetEmploymentType(), Location: req.GetLocation(), IsCurrent: req.GetIsCurrent(), StartDate: unixPtr(req.GetStartDateUnix()), EndDate: unixPtr(req.GetEndDateUnix()), Description: req.GetDescription(), Visibility: mapVisibilityFromProto(req.GetVisibility())}
	out, err := s.ProfileDetailsRepo.CreateEmployment(ctx, userID, in)
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.CreateEmploymentResponse{Employment: toProtoEmployment(out)}, nil
}

func (s *UserServer) UpdateEmployment(ctx context.Context, req *userv1.UpdateEmploymentRequest) (*userv1.UpdateEmploymentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	current, err := s.ProfileDetailsRepo.GetEmployment(ctx, userID, req.GetEmploymentId())
	if err != nil {
		return nil, toStatus(err)
	}
	if req.CompanyName != nil {
		current.CompanyName = req.GetCompanyName()
	}
	if req.Title != nil {
		current.Title = req.GetTitle()
	}
	if req.EmploymentType != nil {
		current.EmploymentType = req.GetEmploymentType()
	}
	if req.Location != nil {
		current.Location = req.GetLocation()
	}
	if req.IsCurrent != nil {
		current.IsCurrent = req.GetIsCurrent()
	}
	if req.StartDateUnix != nil {
		current.StartDate = unixPtr(req.GetStartDateUnix())
	}
	if req.EndDateUnix != nil {
		current.EndDate = unixPtr(req.GetEndDateUnix())
	}
	if req.Description != nil {
		current.Description = req.GetDescription()
	}
	if req.Visibility != nil {
		current.Visibility = mapVisibilityFromProto(req.GetVisibility())
	}
	out, err := s.ProfileDetailsRepo.UpdateEmployment(ctx, userID, req.GetEmploymentId(), current)
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.UpdateEmploymentResponse{Employment: toProtoEmployment(out)}, nil
}

func (s *UserServer) DeleteEmployment(ctx context.Context, req *userv1.DeleteEmploymentRequest) (*userv1.DeleteEmploymentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	deleted, err := s.ProfileDetailsRepo.DeleteEmployment(ctx, userID, req.GetEmploymentId())
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.DeleteEmploymentResponse{Deleted: deleted}, nil
}

func (s *UserServer) ListMyEmployment(ctx context.Context, req *userv1.ListMyEmploymentRequest) (*userv1.ListMyEmploymentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.ProfileDetailsRepo.ListMyEmployment(ctx, userID, req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, toStatus(err)
	}
	items := make([]*userv1.Employment, 0, len(out.Items))
	for _, item := range out.Items {
		items = append(items, toProtoEmployment(item))
	}
	return &userv1.ListMyEmploymentResponse{Employment: items, NextPageToken: out.NextPageToken}, nil
}

func (s *UserServer) ListPublicEmployment(ctx context.Context, req *userv1.ListPublicEmploymentRequest) (*userv1.ListPublicEmploymentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.ProfileDetailsRepo.ListPublicEmployment(ctx, userID, req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, toStatus(err)
	}
	items := make([]*userv1.Employment, 0, len(out.Items))
	for _, item := range out.Items {
		items = append(items, toProtoEmployment(item))
	}
	return &userv1.ListPublicEmploymentResponse{Employment: items, NextPageToken: out.NextPageToken}, nil
}

func (s *UserServer) CreateEducation(ctx context.Context, req *userv1.CreateEducationRequest) (*userv1.CreateEducationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	in := application.Education{SchoolName: req.GetSchoolName(), Degree: req.GetDegree(), FieldOfStudy: req.GetFieldOfStudy(), IsCurrent: req.GetIsCurrent(), StartDate: unixPtr(req.GetStartDateUnix()), EndDate: unixPtr(req.GetEndDateUnix()), Grade: req.GetGrade(), Description: req.GetDescription(), Visibility: mapVisibilityFromProto(req.GetVisibility())}
	out, err := s.ProfileDetailsRepo.CreateEducation(ctx, userID, in)
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.CreateEducationResponse{Education: toProtoEducation(out)}, nil
}

func (s *UserServer) UpdateEducation(ctx context.Context, req *userv1.UpdateEducationRequest) (*userv1.UpdateEducationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	current, err := s.ProfileDetailsRepo.GetEducation(ctx, userID, req.GetEducationId())
	if err != nil {
		return nil, toStatus(err)
	}
	if req.SchoolName != nil {
		current.SchoolName = req.GetSchoolName()
	}
	if req.Degree != nil {
		current.Degree = req.GetDegree()
	}
	if req.FieldOfStudy != nil {
		current.FieldOfStudy = req.GetFieldOfStudy()
	}
	if req.IsCurrent != nil {
		current.IsCurrent = req.GetIsCurrent()
	}
	if req.StartDateUnix != nil {
		current.StartDate = unixPtr(req.GetStartDateUnix())
	}
	if req.EndDateUnix != nil {
		current.EndDate = unixPtr(req.GetEndDateUnix())
	}
	if req.Grade != nil {
		current.Grade = req.GetGrade()
	}
	if req.Description != nil {
		current.Description = req.GetDescription()
	}
	if req.Visibility != nil {
		current.Visibility = mapVisibilityFromProto(req.GetVisibility())
	}
	out, err := s.ProfileDetailsRepo.UpdateEducation(ctx, userID, req.GetEducationId(), current)
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.UpdateEducationResponse{Education: toProtoEducation(out)}, nil
}

func (s *UserServer) DeleteEducation(ctx context.Context, req *userv1.DeleteEducationRequest) (*userv1.DeleteEducationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	deleted, err := s.ProfileDetailsRepo.DeleteEducation(ctx, userID, req.GetEducationId())
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.DeleteEducationResponse{Deleted: deleted}, nil
}

func (s *UserServer) ListMyEducation(ctx context.Context, req *userv1.ListMyEducationRequest) (*userv1.ListMyEducationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.ProfileDetailsRepo.ListMyEducation(ctx, userID, req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, toStatus(err)
	}
	items := make([]*userv1.Education, 0, len(out.Items))
	for _, item := range out.Items {
		items = append(items, toProtoEducation(item))
	}
	return &userv1.ListMyEducationResponse{Education: items, NextPageToken: out.NextPageToken}, nil
}

func (s *UserServer) ListPublicEducation(ctx context.Context, req *userv1.ListPublicEducationRequest) (*userv1.ListPublicEducationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.ProfileDetailsRepo.ListPublicEducation(ctx, userID, req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, toStatus(err)
	}
	items := make([]*userv1.Education, 0, len(out.Items))
	for _, item := range out.Items {
		items = append(items, toProtoEducation(item))
	}
	return &userv1.ListPublicEducationResponse{Education: items, NextPageToken: out.NextPageToken}, nil
}

func (s *UserServer) CreateCertification(ctx context.Context, req *userv1.CreateCertificationRequest) (*userv1.CreateCertificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	in := application.Certification{Name: req.GetName(), IssuingOrganization: req.GetIssuingOrganization(), CredentialID: req.GetCredentialId(), CredentialURL: req.GetCredentialUrl(), IssueDate: unixPtr(req.GetIssueDateUnix()), ExpirationDate: unixPtr(req.GetExpirationDateUnix()), DoesNotExpire: req.GetDoesNotExpire(), Visibility: mapVisibilityFromProto(req.GetVisibility())}
	out, err := s.ProfileDetailsRepo.CreateCertification(ctx, userID, in)
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.CreateCertificationResponse{Certification: toProtoCertification(out)}, nil
}

func (s *UserServer) UpdateCertification(ctx context.Context, req *userv1.UpdateCertificationRequest) (*userv1.UpdateCertificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	current, err := s.ProfileDetailsRepo.GetCertification(ctx, userID, req.GetCertificationId())
	if err != nil {
		return nil, toStatus(err)
	}
	if req.Name != nil {
		current.Name = req.GetName()
	}
	if req.IssuingOrganization != nil {
		current.IssuingOrganization = req.GetIssuingOrganization()
	}
	if req.CredentialId != nil {
		current.CredentialID = req.GetCredentialId()
	}
	if req.CredentialUrl != nil {
		current.CredentialURL = req.GetCredentialUrl()
	}
	if req.IssueDateUnix != nil {
		current.IssueDate = unixPtr(req.GetIssueDateUnix())
	}
	if req.ExpirationDateUnix != nil {
		current.ExpirationDate = unixPtr(req.GetExpirationDateUnix())
	}
	if req.DoesNotExpire != nil {
		current.DoesNotExpire = req.GetDoesNotExpire()
	}
	if req.Visibility != nil {
		current.Visibility = mapVisibilityFromProto(req.GetVisibility())
	}
	out, err := s.ProfileDetailsRepo.UpdateCertification(ctx, userID, req.GetCertificationId(), current)
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.UpdateCertificationResponse{Certification: toProtoCertification(out)}, nil
}

func (s *UserServer) DeleteCertification(ctx context.Context, req *userv1.DeleteCertificationRequest) (*userv1.DeleteCertificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	deleted, err := s.ProfileDetailsRepo.DeleteCertification(ctx, userID, req.GetCertificationId())
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.DeleteCertificationResponse{Deleted: deleted}, nil
}

func (s *UserServer) ListMyCertifications(ctx context.Context, req *userv1.ListMyCertificationsRequest) (*userv1.ListMyCertificationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.ProfileDetailsRepo.ListMyCertifications(ctx, userID, req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, toStatus(err)
	}
	items := make([]*userv1.Certification, 0, len(out.Items))
	for _, item := range out.Items {
		items = append(items, toProtoCertification(item))
	}
	return &userv1.ListMyCertificationsResponse{Certifications: items, NextPageToken: out.NextPageToken}, nil
}

func (s *UserServer) ListPublicCertifications(ctx context.Context, req *userv1.ListPublicCertificationsRequest) (*userv1.ListPublicCertificationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.ProfileDetailsRepo.ListPublicCertifications(ctx, userID, req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, toStatus(err)
	}
	items := make([]*userv1.Certification, 0, len(out.Items))
	for _, item := range out.Items {
		items = append(items, toProtoCertification(item))
	}
	return &userv1.ListPublicCertificationsResponse{Certifications: items, NextPageToken: out.NextPageToken}, nil
}

func (s *UserServer) UpsertLanguages(ctx context.Context, req *userv1.UpsertLanguagesRequest) (*userv1.UpsertLanguagesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	langs := make([]application.LanguageProficiency, 0, len(req.GetLanguages()))
	for _, l := range req.GetLanguages() {
		langs = append(langs, application.LanguageProficiency{LanguageCode: l.GetLanguageCode(), Proficiency: l.GetProficiency(), Visibility: mapVisibilityFromProto(l.GetVisibility())})
	}
	out, err := s.ProfileDetailsRepo.UpsertLanguages(ctx, userID, langs)
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.UpsertLanguagesResponse{Languages: toProtoLanguages(out)}, nil
}

func (s *UserServer) GetMyLanguages(ctx context.Context, req *userv1.GetMyLanguagesRequest) (*userv1.GetMyLanguagesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.ProfileDetailsRepo.GetMyLanguages(ctx, userID)
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.GetMyLanguagesResponse{Languages: toProtoLanguages(out)}, nil
}

func (s *UserServer) GetPublicLanguages(ctx context.Context, req *userv1.GetPublicLanguagesRequest) (*userv1.GetPublicLanguagesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.ProfileDetailsRepo.GetPublicLanguages(ctx, userID)
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.GetPublicLanguagesResponse{Languages: toProtoLanguages(out)}, nil
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

func toAppPortfolioItem(req *userv1.CreatePortfolioItemRequest) application.PortfolioItem {
	return application.PortfolioItem{
		Title:         req.GetTitle(),
		Description:   req.GetDescription(),
		ProjectURL:    req.GetProjectUrl(),
		RoleInProject: req.GetRoleInProject(),
		CompletedAt:   unixPtr(req.GetCompletedAtUnix()),
		Visibility:    mapVisibilityFromProto(req.GetVisibility()),
		Tags:          req.GetTags(),
		Media:         toAppPortfolioMediaSlice(req.GetMedia()),
	}
}

func toAppPortfolioMediaSlice(items []*userv1.PortfolioMedia) []application.PortfolioMedia {
	out := make([]application.PortfolioMedia, 0, len(items))
	for _, m := range items {
		out = append(out, application.PortfolioMedia{
			ID:          m.GetId(),
			MediaType:   mapPortfolioMediaTypeFromProto(m.GetMediaType()),
			StorageKey:  m.GetStorageKey(),
			ExternalURL: m.GetExternalUrl(),
			FileName:    m.GetFileName(),
			ContentType: m.GetContentType(),
			SizeBytes:   m.GetSizeBytes(),
			Width:       m.GetWidth(),
			Height:      m.GetHeight(),
			SortOrder:   m.GetSortOrder(),
		})
	}
	return out
}

func toProtoPortfolioItem(item application.PortfolioItem) *userv1.PortfolioItem {
	media := make([]*userv1.PortfolioMedia, 0, len(item.Media))
	for _, m := range item.Media {
		media = append(media, &userv1.PortfolioMedia{
			Id:          m.ID,
			MediaType:   mapPortfolioMediaTypeToProto(m.MediaType),
			StorageKey:  m.StorageKey,
			ExternalUrl: m.ExternalURL,
			FileName:    m.FileName,
			ContentType: m.ContentType,
			SizeBytes:   m.SizeBytes,
			Width:       m.Width,
			Height:      m.Height,
			SortOrder:   m.SortOrder,
		})
	}
	var completedAt int64
	if item.CompletedAt != nil {
		completedAt = item.CompletedAt.Unix()
	}
	return &userv1.PortfolioItem{
		Id:              item.ID,
		UserId:          item.UserID.String(),
		Title:           item.Title,
		Description:     item.Description,
		ProjectUrl:      item.ProjectURL,
		RoleInProject:   item.RoleInProject,
		CompletedAtUnix: completedAt,
		SortOrder:       item.SortOrder,
		Visibility:      mapVisibilityToProto(item.Visibility),
		Tags:            item.Tags,
		Media:           media,
		CreatedAtUnix:   item.CreatedAt.Unix(),
		UpdatedAtUnix:   item.UpdatedAt.Unix(),
	}
}

func toProtoEmployment(in application.Employment) *userv1.Employment {
	return &userv1.Employment{
		Id:             in.ID,
		UserId:         in.UserID.String(),
		CompanyName:    in.CompanyName,
		Title:          in.Title,
		EmploymentType: in.EmploymentType,
		Location:       in.Location,
		IsCurrent:      in.IsCurrent,
		StartDateUnix:  timePtrUnix(in.StartDate),
		EndDateUnix:    timePtrUnix(in.EndDate),
		Description:    in.Description,
		SortOrder:      in.SortOrder,
		Visibility:     mapVisibilityToProto(in.Visibility),
		CreatedAtUnix:  in.CreatedAt.Unix(),
		UpdatedAtUnix:  in.UpdatedAt.Unix(),
	}
}

func toProtoEducation(in application.Education) *userv1.Education {
	return &userv1.Education{
		Id:            in.ID,
		UserId:        in.UserID.String(),
		SchoolName:    in.SchoolName,
		Degree:        in.Degree,
		FieldOfStudy:  in.FieldOfStudy,
		IsCurrent:     in.IsCurrent,
		StartDateUnix: timePtrUnix(in.StartDate),
		EndDateUnix:   timePtrUnix(in.EndDate),
		Grade:         in.Grade,
		Description:   in.Description,
		SortOrder:     in.SortOrder,
		Visibility:    mapVisibilityToProto(in.Visibility),
		CreatedAtUnix: in.CreatedAt.Unix(),
		UpdatedAtUnix: in.UpdatedAt.Unix(),
	}
}

func toProtoCertification(in application.Certification) *userv1.Certification {
	return &userv1.Certification{
		Id:                  in.ID,
		UserId:              in.UserID.String(),
		Name:                in.Name,
		IssuingOrganization: in.IssuingOrganization,
		CredentialId:        in.CredentialID,
		CredentialUrl:       in.CredentialURL,
		IssueDateUnix:       timePtrUnix(in.IssueDate),
		ExpirationDateUnix:  timePtrUnix(in.ExpirationDate),
		DoesNotExpire:       in.DoesNotExpire,
		Visibility:          mapVisibilityToProto(in.Visibility),
		CreatedAtUnix:       in.CreatedAt.Unix(),
		UpdatedAtUnix:       in.UpdatedAt.Unix(),
	}
}

func toProtoLanguages(items []application.LanguageProficiency) []*userv1.LanguageProficiency {
	out := make([]*userv1.LanguageProficiency, 0, len(items))
	for _, item := range items {
		out = append(out, &userv1.LanguageProficiency{
			LanguageCode: item.LanguageCode,
			Proficiency:  item.Proficiency,
			Visibility:   mapVisibilityToProto(item.Visibility),
		})
	}
	return out
}

func unixPtr(v int64) *time.Time {
	if v <= 0 {
		return nil
	}
	t := time.Unix(v, 0).UTC()
	return &t
}

func timePtrUnix(v *time.Time) int64 {
	if v == nil {
		return 0
	}
	return v.Unix()
}

func stringPtrOrNil(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}

func defaultNotificationSettings() userv1.NotificationSettings {
	return userv1.NotificationSettings{
		EmailJobAlerts: true,
		EmailMessages:  true,
		EmailBilling:   true,
		EmailSecurity:  true,
		PushJobAlerts:  true,
		PushMessages:   true,
		PushBilling:    false,
		PushSecurity:   true,
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

func mapVisibilityFromProto(visibility userv1.ProfileVisibility) string {
	switch visibility {
	case userv1.ProfileVisibility_PROFILE_VISIBILITY_PRIVATE:
		return domain.ProfileVisibilityPrivate
	case userv1.ProfileVisibility_PROFILE_VISIBILITY_PUBLIC:
		return domain.ProfileVisibilityPublic
	default:
		return domain.ProfileVisibilityPublic
	}
}

func mapPortfolioMediaTypeToProto(v string) userv1.PortfolioMediaType {
	switch strings.ToUpper(strings.TrimPrefix(strings.TrimSpace(v), "PORTFOLIO_MEDIA_TYPE_")) {
	case "IMAGE":
		return userv1.PortfolioMediaType_PORTFOLIO_MEDIA_TYPE_IMAGE
	case "VIDEO":
		return userv1.PortfolioMediaType_PORTFOLIO_MEDIA_TYPE_VIDEO
	case "FILE":
		return userv1.PortfolioMediaType_PORTFOLIO_MEDIA_TYPE_FILE
	case "LINK":
		return userv1.PortfolioMediaType_PORTFOLIO_MEDIA_TYPE_LINK
	default:
		return userv1.PortfolioMediaType_PORTFOLIO_MEDIA_TYPE_UNSPECIFIED
	}
}

func mapPortfolioMediaTypeFromProto(v userv1.PortfolioMediaType) string {
	switch v {
	case userv1.PortfolioMediaType_PORTFOLIO_MEDIA_TYPE_IMAGE:
		return "IMAGE"
	case userv1.PortfolioMediaType_PORTFOLIO_MEDIA_TYPE_VIDEO:
		return "VIDEO"
	case userv1.PortfolioMediaType_PORTFOLIO_MEDIA_TYPE_FILE:
		return "FILE"
	case userv1.PortfolioMediaType_PORTFOLIO_MEDIA_TYPE_LINK:
		return "LINK"
	default:
		return "UNSPECIFIED"
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
