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

func (s *UserServer) CreateMyProfile(ctx context.Context, req *userv1.CreateMyProfileRequest) (*userv1.CreateMyProfileResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	var client *domain.ClientProfile
	if req.GetClient() != nil {
		client = &domain.ClientProfile{CompanyName: req.GetClient().CompanyName}
	}

	var freelancer *domain.FreelancerProfile
	if req.GetFreelancer() != nil {
		freelancer = &domain.FreelancerProfile{
			Headline:     req.GetFreelancer().Headline,
			Skills:       req.GetFreelancer().Skills,
			HourlyRate:   req.GetFreelancer().HourlyRate,
			Availability: mapAvailabilityFromProto(req.GetFreelancer().Availability),
			Location:     req.GetFreelancer().Location,
		}
	}

	_, err = s.CreateProfileUC.Execute(ctx, application.CreateProfileInput{
		UserID:       userID,
		Role:         mapRoleFromProto(req.Role),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		DisplayName:  req.DisplayName,
		ContactEmail: req.ContactEmail,
		AvatarURL:    "",
		Client:       client,
		Freelancer:   freelancer,
	})
	if err != nil {
		return nil, toStatus(err)
	}

	out, err := s.GetProfileUC.Execute(ctx, application.GetProfileInput{UserID: userID})
	if err != nil {
		return nil, toStatus(err)
	}

	return &userv1.CreateMyProfileResponse{
		Success: true,
		Profile: s.toProtoUserProfile(out.Profile, out.Client, out.Freelancer),
	}, nil
}

func (s *UserServer) GetMyProfile(ctx context.Context, req *userv1.GetMyProfileRequest) (*userv1.GetMyProfileResponse, error) {
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
	statusOut := s.GetOnboardingStatusUC.Build(out.Profile, out.Client, out.Freelancer)
	return &userv1.GetMyProfileResponse{
		Profile: s.toProtoUserProfile(out.Profile, out.Client, out.Freelancer),
		Completeness: &userv1.ProfileCompleteness{
			Percent:               statusOut.Percent,
			MissingRequiredFields: statusOut.Missing,
		},
	}, nil
}

func (s *UserServer) PatchMyProfile(ctx context.Context, req *userv1.PatchMyProfileRequest) (*userv1.PatchMyProfileResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	in := application.UpdateProfileInput{UserID: userID}

	if req.Core != nil {
		in.DisplayName = req.Core.DisplayName
		if req.Core.Language != nil {
			return nil, status.Error(codes.InvalidArgument, "language must be updated via PatchMySettings")
		}
		in.ContactEmail = req.Core.ContactEmail
		in.ContactPhone = req.Core.ContactPhone
		in.Bio = req.Core.Bio
	}

	if req.GetClient() != nil {
		in.CompanyName = req.GetClient().CompanyName
	}

	if req.GetFreelancer() != nil {
		in.Headline = req.GetFreelancer().Headline
		if req.GetFreelancer().Skills != nil {
			in.Skills = req.GetFreelancer().Skills.GetValues()
		}
		in.HourlyRate = req.GetFreelancer().HourlyRate
		in.Availability = stringPtrFromAvailability(req.GetFreelancer().Availability)
		in.Location = req.GetFreelancer().Location
	}

	if hasClearField(req.ClearFields, "contact_phone") {
		empty := ""
		in.ContactPhone = &empty
	}
	if hasClearField(req.ClearFields, "bio") {
		empty := ""
		in.Bio = &empty
	}
	if hasClearField(req.ClearFields, "company_name") {
		empty := ""
		in.CompanyName = &empty
	}
	if hasClearField(req.ClearFields, "headline") {
		empty := ""
		in.Headline = &empty
	}
	if hasClearField(req.ClearFields, "location") {
		empty := ""
		in.Location = &empty
	}
	if hasClearField(req.ClearFields, "skills") {
		in.Skills = []string{}
	}

	out, err := s.UpdateProfileUC.Execute(ctx, in)
	if err != nil {
		return nil, toStatus(err)
	}

	return &userv1.PatchMyProfileResponse{
		Profile: s.toProtoUserProfile(out.Profile, out.Client, out.Freelancer),
		Completeness: &userv1.ProfileCompleteness{
			Percent:               out.Completeness,
			MissingRequiredFields: out.Missing,
		},
	}, nil
}

func (s *UserServer) DeleteMyProfile(ctx context.Context, req *userv1.DeleteMyProfileRequest) (*userv1.DeleteMyProfileResponse, error) {
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
	return &userv1.DeleteMyProfileResponse{Deleted: out.Deleted}, nil
}

func (s *UserServer) GetMyOnboardingStatus(ctx context.Context, req *userv1.GetMyOnboardingStatusRequest) (*userv1.GetMyOnboardingStatusResponse, error) {
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
	return &userv1.GetMyOnboardingStatusResponse{
		Completeness: &userv1.ProfileCompleteness{
			Percent:               out.Percent,
			MissingRequiredFields: out.Missing,
		},
		Steps: toProtoOnboardingSteps(out.Steps),
	}, nil
}

func (s *UserServer) GetMySettings(ctx context.Context, req *userv1.GetMySettingsRequest) (*userv1.GetMySettingsResponse, error) {
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
	return &userv1.GetMySettingsResponse{Settings: toDefaultSettings(out.Profile.Language)}, nil
}

func (s *UserServer) PatchMySettings(ctx context.Context, req *userv1.PatchMySettingsRequest) (*userv1.PatchMySettingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	updateIn := application.UpdateProfileInput{UserID: userID}
	if req.UiLocale != nil {
		locale := strings.TrimSpace(req.GetUiLocale())
		if locale == "" {
			return nil, status.Error(codes.InvalidArgument, "ui_locale cannot be empty")
		}
		updateIn.Language = &locale
	}

	if req.EmailNotificationsEnabled != nil || req.PushNotificationsEnabled != nil {
		return nil, status.Error(codes.Unimplemented, "notification settings are not implemented yet")
	}

	if updateIn.Language != nil {
		if _, err := s.UpdateProfileUC.Execute(ctx, updateIn); err != nil {
			return nil, toStatus(err)
		}
	}

	out, err := s.GetProfileUC.Execute(ctx, application.GetProfileInput{UserID: userID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.PatchMySettingsResponse{Settings: toDefaultSettings(out.Profile.Language)}, nil
}

func (s *UserServer) UpsertMyAvatar(ctx context.Context, req *userv1.UploadMyAvatarRequest) (*userv1.UploadMyAvatarResponse, error) {
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
	return &userv1.UploadMyAvatarResponse{
		AvatarUrl: out.AvatarURL,
		Avatar: &userv1.ProfileAvatar{
			UserId:        req.UserId,
			FileName:      req.FileName,
			ContentType:   out.ContentType,
			StorageKey:    "",
			SizeBytes:     out.SizeBytes,
			Width:         out.Width,
			Height:        out.Height,
			UpdatedAtUnix: time.Now().UTC().Unix(),
		},
	}, nil
}

func (s *UserServer) GetMyAvatar(ctx context.Context, req *userv1.GetMyAvatarRequest) (*userv1.GetMyAvatarResponse, error) {
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
	return &userv1.GetMyAvatarResponse{
		Avatar: &userv1.ProfileAvatar{
			UserId:        req.UserId,
			FileName:      out.FileName,
			ContentType:   out.ContentType,
			StorageKey:    "",
			SizeBytes:     int64(len(out.Content)),
			Width:         0,
			Height:        0,
			UpdatedAtUnix: 0,
		},
		Content: out.Content,
	}, nil
}

func (s *UserServer) RemoveMyAvatar(ctx context.Context, req *userv1.RemoveMyAvatarRequest) (*userv1.RemoveMyAvatarResponse, error) {
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
	return &userv1.RemoveMyAvatarResponse{Removed: out.Removed}, nil
}

func (s *UserServer) toProtoUserProfile(profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) *userv1.UserProfile {
	out := &userv1.UserProfile{
		Core: &userv1.UserCore{
			ProfileId:          profile.ID,
			UserId:             profile.UserID.String(),
			Role:               toProtoRole(profile.Role),
			FirstName:          profile.FirstName,
			LastName:           profile.LastName,
			DisplayName:        profile.DisplayName,
			AvatarUrl:          profile.AvatarURL,
			Language:           profile.Language,
			ContactEmail:       profile.ContactEmail,
			ContactPhone:       profile.ContactPhone,
			Bio:                profile.Bio,
			AccountStatus:      toProtoAccountStatus(profile.AccountStatus),
			SuspensionReason:   profile.SuspensionReason,
			TaxId:              "",
			VerificationStatus: toProtoVerificationStatus(""),
			CreatedAtUnix:      profile.CreatedAt.Unix(),
			UpdatedAtUnix:      profile.UpdatedAt.Unix(),
		},
		Capabilities: toProtoCapabilities(s.CapabilityPolicy, profile, client, freelancer),
	}

	if client != nil {
		out.Client = &userv1.ClientProfile{CompanyName: client.CompanyName}
		out.Core.TaxId = client.TaxID
		out.Core.VerificationStatus = toProtoVerificationStatus(client.VerificationStatus)
	}
	if freelancer != nil {
		out.Freelancer = &userv1.FreelancerProfile{
			Headline:     freelancer.Headline,
			Skills:       freelancer.Skills,
			HourlyRate:   freelancer.HourlyRate,
			Availability: toProtoAvailability(freelancer.Availability),
			Location:     freelancer.Location,
			Metrics: &userv1.FreelancerMetrics{
				Rating:             freelancer.Rating,
				JobSuccessScore:    freelancer.Reputation.JobSuccessScore,
				TotalReviews:       freelancer.Reputation.TotalReviews,
				TotalJobs:          freelancer.Reputation.TotalJobs,
				TotalEarnings:      freelancer.Reputation.TotalEarningsUSD,
				VerificationStatus: toProtoVerificationStatus(freelancer.VerificationStatus),
			},
		}
		if profile.LastActiveAt != nil {
			lastActive := profile.LastActiveAt.Unix()
			out.Freelancer.Metrics.LastActiveAtUnix = &lastActive
		}
		out.Core.VerificationStatus = toProtoVerificationStatus(freelancer.VerificationStatus)
	}
	if profile.DeletedAt != nil && !profile.DeletedAt.IsZero() {
		deletedAt := profile.DeletedAt.Unix()
		out.Core.DeletedAtUnix = &deletedAt
	}
	return out
}

func toProtoOnboardingSteps(steps []application.OnboardingStep) []*userv1.OnboardingStep {
	out := make([]*userv1.OnboardingStep, 0, len(steps))
	for _, step := range steps {
		statusValue := userv1.OnboardingStepStatus_ONBOARDING_STEP_STATUS_NOT_STARTED
		if step.Completed {
			statusValue = userv1.OnboardingStepStatus_ONBOARDING_STEP_STATUS_COMPLETED
		}
		out = append(out, &userv1.OnboardingStep{Key: step.Key, Status: statusValue})
	}
	return out
}

func toProtoCapabilities(policy CapabilityPolicy, p domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) *userv1.CapabilityFlags {
	active := strings.EqualFold(strings.TrimPrefix(strings.TrimSpace(p.AccountStatus), "ACCOUNT_STATUS_"), domain.AccountStatusActive)
	isFreelancer := p.Role == domain.RoleFreelancer
	isClient := p.Role == domain.RoleClient
	freelancerVerified := freelancer != nil && strings.EqualFold(strings.TrimSpace(freelancer.VerificationStatus), domain.VerificationStatusVerified)
	hasHeadline := freelancer != nil && strings.TrimSpace(freelancer.Headline) != ""
	hasEnoughSkills := freelancer != nil && len(freelancer.Skills) >= policy.MinSkillsForDiscovery
	hasDiscoverableFreelancerProfile := freelancer != nil && hasEnoughSkills && (!policy.RequireHeadlineForFreelancer || hasHeadline)
	hasDiscoverableClientProfile := client != nil && (!policy.RequireCompanyNameForClient || strings.TrimSpace(client.CompanyName) != "")

	return &userv1.CapabilityFlags{
		CanApplyJobs:     active && isFreelancer,
		CanPostJobs:      active && isClient,
		CanWithdrawFunds: active && isFreelancer && (!policy.RequireVerifiedForWithdraw || freelancerVerified),
		CanMessage:       active || policy.AllowMessagingWhenSuspended,
		CanBeDiscovered:  active && (hasDiscoverableFreelancerProfile || hasDiscoverableClientProfile),
	}
}

func mapRoleFromProto(role userv1.UserRole) string {
	switch role {
	case userv1.UserRole_USER_ROLE_CLIENT:
		return domain.RoleClient
	case userv1.UserRole_USER_ROLE_FREELANCER:
		return domain.RoleFreelancer
	case userv1.UserRole_USER_ROLE_ADMIN:
		return domain.RoleAdmin
	default:
		return ""
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

func stringPtrFromAvailability(availability *userv1.Availability) *string {
	if availability == nil {
		return nil
	}
	mapped := mapAvailabilityFromProto(*availability)
	return &mapped
}

func toProtoRole(role string) userv1.UserRole {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case domain.RoleClient:
		return userv1.UserRole_USER_ROLE_CLIENT
	case domain.RoleFreelancer:
		return userv1.UserRole_USER_ROLE_FREELANCER
	case domain.RoleAdmin:
		return userv1.UserRole_USER_ROLE_ADMIN
	default:
		return userv1.UserRole_USER_ROLE_UNSPECIFIED
	}
}

func toProtoAccountStatus(statusValue string) userv1.AccountStatus {
	s := strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(statusValue)), "ACCOUNT_STATUS_")
	switch s {
	case domain.AccountStatusActive:
		return userv1.AccountStatus_ACCOUNT_STATUS_ACTIVE
	case domain.AccountStatusSuspended:
		return userv1.AccountStatus_ACCOUNT_STATUS_SUSPENDED
	case domain.AccountStatusDeleted:
		return userv1.AccountStatus_ACCOUNT_STATUS_DELETED
	default:
		return userv1.AccountStatus_ACCOUNT_STATUS_UNSPECIFIED
	}
}

func toProtoVerificationStatus(value string) userv1.VerificationStatus {
	s := strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(value)), "VERIFICATION_STATUS_")
	switch s {
	case domain.VerificationStatusPending:
		return userv1.VerificationStatus_VERIFICATION_STATUS_PENDING
	case domain.VerificationStatusVerified:
		return userv1.VerificationStatus_VERIFICATION_STATUS_VERIFIED
	case domain.VerificationStatusRejected:
		return userv1.VerificationStatus_VERIFICATION_STATUS_REJECTED
	case domain.VerificationStatusExpired:
		return userv1.VerificationStatus_VERIFICATION_STATUS_EXPIRED
	default:
		return userv1.VerificationStatus_VERIFICATION_STATUS_UNSPECIFIED
	}
}

func toProtoAvailability(value string) userv1.Availability {
	s := strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(value)), "AVAILABILITY_")
	switch s {
	case strings.TrimPrefix(domain.AvailabilityFullTime, "AVAILABILITY_"):
		return userv1.Availability_AVAILABILITY_FULL_TIME
	case strings.TrimPrefix(domain.AvailabilityPartTime, "AVAILABILITY_"):
		return userv1.Availability_AVAILABILITY_PART_TIME
	case strings.TrimPrefix(domain.AvailabilityUnavailable, "AVAILABILITY_"):
		return userv1.Availability_AVAILABILITY_UNAVAILABLE
	default:
		return userv1.Availability_AVAILABILITY_AS_NEEDED
	}
}

func toStatus(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case contains(msg, "duplicate key"), contains(msg, "23505"):
		return status.Error(codes.AlreadyExists, "profile already exists for user_id")
	case contains(msg, "not found"):
		return status.Error(codes.NotFound, msg)
	case contains(msg, "required"), contains(msg, "invalid"), contains(msg, "not allowed"):
		return status.Error(codes.InvalidArgument, msg)
	case contains(msg, "unsupported"), contains(msg, "exceeds"), contains(msg, "too small"):
		return status.Error(codes.InvalidArgument, msg)
	default:
		return status.Error(codes.Internal, msg)
	}
}

func contains(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

func toDefaultSettings(uiLocale string) *userv1.UserSettings {
	locale := strings.TrimSpace(uiLocale)
	if locale == "" {
		locale = "en"
	}
	return &userv1.UserSettings{
		UiLocale:                  locale,
		EmailNotificationsEnabled: true,
		PushNotificationsEnabled:  true,
	}
}

func hasClearField(fields []string, target string) bool {
	for _, field := range fields {
		if strings.EqualFold(strings.TrimSpace(field), target) {
			return true
		}
	}
	return false
}
