package grpcadapter

import (
	"context"
	"net/url"
	"strings"
	"time"

	userv1 "jobconnect/user/gen/user"
	"jobconnect/user/internal/application"
	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type portfolioURLPresigner interface {
	PresignGetObject(ctx context.Context, storageKey string, ttl time.Duration) (string, error)
}

type UserServer struct {
	userv1.UnimplementedUserServiceServer
	CreateProfileUC       *application.CreateProfile
	GetProfileUC          *application.GetProfile
	UpdateProfileUC       *application.UpdateProfile
	DeleteProfileUC       *application.DeleteProfile
	GetOnboardingStatusUC *application.GetOnboardingStatus
	GetSettingsUC         *application.GetSettings
	PatchSettingsUC       *application.PatchSettingsUseCase
	GetAvatarUploadURLUC  *application.GetAvatarUploadURL
	UploadAvatarUC        *application.UploadAvatar
	GetAvatarUC           *application.GetAvatar
	RemoveAvatarUC        *application.RemoveAvatar
	GetCVUploadURLUC      *application.GetCVUploadURL
	UpsertCVUC            *application.UpsertCV
	GetCVUC               *application.GetCV
	RemoveCVUC            *application.RemoveCV
	PortfolioStore        portfolioURLPresigner
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
	getProfile *application.GetProfile,
	updateProfile *application.UpdateProfile,
	deleteProfile *application.DeleteProfile,
	getOnboardingStatus *application.GetOnboardingStatus,
	getSettings *application.GetSettings,
	patchSettings *application.PatchSettingsUseCase,
	getAvatarUploadURL *application.GetAvatarUploadURL,
	uploadAvatar *application.UploadAvatar,
	getAvatar *application.GetAvatar,
	removeAvatar *application.RemoveAvatar,
	getCVUploadURL *application.GetCVUploadURL,
	upsertCV *application.UpsertCV,
	getCV *application.GetCV,
	removeCV *application.RemoveCV,
	portfolioStore portfolioURLPresigner,
	profileDetailsRepo application.ProfileDetailsRepository,
	capabilityPolicy CapabilityPolicy,
) *UserServer {
	return &UserServer{
		CreateProfileUC:       createProfile,
		GetProfileUC:          getProfile,
		UpdateProfileUC:       updateProfile,
		DeleteProfileUC:       deleteProfile,
		GetOnboardingStatusUC: getOnboardingStatus,
		GetSettingsUC:         getSettings,
		PatchSettingsUC:       patchSettings,
		GetAvatarUploadURLUC:  getAvatarUploadURL,
		UploadAvatarUC:        uploadAvatar,
		GetAvatarUC:           getAvatar,
		RemoveAvatarUC:        removeAvatar,
		GetCVUploadURLUC:      getCVUploadURL,
		UpsertCVUC:            upsertCV,
		GetCVUC:               getCV,
		RemoveCVUC:            removeCV,
		PortfolioStore:        portfolioStore,
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
		}
	}

	_, err = s.CreateProfileUC.Execute(ctx, application.CreateProfileInput{
		UserID:       userID,
		Role:         mapRoleFromProto(req.Role),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		DisplayName:  req.DisplayName,
		Location:     req.Location,
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
		in.ContactEmail = req.Core.ContactEmail
		in.ContactPhone = req.Core.ContactPhone
		in.Bio = req.Core.Bio
		in.TaxID = req.Core.TaxId
		in.Location = req.Core.Location
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
	}

	if hasClearField(req.ClearFields, "contact_phone") {
		empty := ""
		in.ContactPhone = &empty
	}
	if hasClearField(req.ClearFields, "bio") {
		empty := ""
		in.Bio = &empty
	}
	if hasClearField(req.ClearFields, "tax_id") {
		empty := ""
		in.TaxID = &empty
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
		Readiness: &userv1.ProfileReadiness{
			Percent:               out.ReadinessPercent,
			MissingRequiredFields: out.ReadinessMissing,
			Recommendations:       out.ReadinessRecommendations,
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
	out, err := s.GetSettingsUC.Execute(ctx, application.GetSettingsInput{UserID: userID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.GetMySettingsResponse{Settings: toProtoSettings(out.Settings)}, nil
}

func (s *UserServer) PatchMySettings(ctx context.Context, req *userv1.PatchMySettingsRequest) (*userv1.PatchMySettingsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	out, err := s.PatchSettingsUC.Execute(ctx, application.PatchSettingsInput{
		UserID: userID,
		Patch: application.PatchSettings{
			UILocale:                  req.UiLocale,
			EmailNotificationsEnabled: req.EmailNotificationsEnabled,
			PushNotificationsEnabled:  req.PushNotificationsEnabled,
		},
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.PatchMySettingsResponse{Settings: toProtoSettings(out.Settings)}, nil
}

func (s *UserServer) PatchMyWorkPreferences(ctx context.Context, req *userv1.PatchMyWorkPreferencesRequest) (*userv1.PatchMyWorkPreferencesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	current, err := s.ProfileDetailsRepo.GetWorkPreferences(ctx, userID)
	if err != nil {
		return nil, toStatus(err)
	}

	if req.PreferredProjectLength != nil {
		projectLength, ok := fromProtoProjectLength(req.GetPreferredProjectLength())
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "invalid preferred_project_length")
		}
		current.PreferredProjectLength = projectLength
	}
	if req.MinBudget != nil {
		current.MinBudgetUSD = req.GetMinBudget()
	}
	if req.MaxBudget != nil {
		current.MaxBudgetUSD = req.GetMaxBudget()
	}
	if req.ContractTypes != nil {
		current.ContractTypes = req.ContractTypes.GetValues()
	}
	if req.WeeklyCapacityHours != nil {
		current.WeeklyCapacityHours = req.GetWeeklyCapacityHours()
	}
	if current.MinBudgetUSD < 0 || current.MaxBudgetUSD < 0 {
		return nil, status.Error(codes.InvalidArgument, "budget values must be greater than or equal to 0")
	}
	if current.MaxBudgetUSD > 0 && current.MinBudgetUSD > current.MaxBudgetUSD {
		return nil, status.Error(codes.InvalidArgument, "min_budget cannot be greater than max_budget")
	}

	updated, err := s.ProfileDetailsRepo.SetWorkPreferences(ctx, userID, current)
	if err != nil {
		return nil, toStatus(err)
	}

	return &userv1.PatchMyWorkPreferencesResponse{Settings: toProtoWorkPreferences(updated)}, nil
}

func (s *UserServer) GetMyWorkPreferences(ctx context.Context, req *userv1.GetMyWorkPreferencesRequest) (*userv1.GetMyWorkPreferencesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	out, err := s.ProfileDetailsRepo.GetWorkPreferences(ctx, userID)
	if err != nil {
		return nil, toStatus(err)
	}

	return &userv1.GetMyWorkPreferencesResponse{Settings: toProtoWorkPreferences(out)}, nil
}

func (s *UserServer) GetMyHiringPreferences(ctx context.Context, req *userv1.GetMyHiringPreferencesRequest) (*userv1.GetMyHiringPreferencesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	out, err := s.ProfileDetailsRepo.GetHiringPreferences(ctx, userID)
	if err != nil {
		return nil, toStatus(err)
	}

	return &userv1.GetMyHiringPreferencesResponse{Preferences: toProtoHiringPreferences(out)}, nil
}

func (s *UserServer) PatchMyHiringPreferences(ctx context.Context, req *userv1.PatchMyHiringPreferencesRequest) (*userv1.PatchMyHiringPreferencesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	current, err := s.ProfileDetailsRepo.GetHiringPreferences(ctx, userID)
	if err != nil {
		return nil, toStatus(err)
	}

	if req.MinHourlyRate != nil {
		current.MinHourlyRate = req.GetMinHourlyRate()
	}
	if req.MaxHourlyRate != nil {
		current.MaxHourlyRate = req.GetMaxHourlyRate()
	}
	if req.PreferredLocations != nil {
		current.PreferredLocations = req.PreferredLocations.GetValues()
	}
	if current.MinHourlyRate < 0 || current.MaxHourlyRate < 0 {
		return nil, status.Error(codes.InvalidArgument, "hourly rates must be greater than or equal to 0")
	}
	if current.MaxHourlyRate > 0 && current.MinHourlyRate > current.MaxHourlyRate {
		return nil, status.Error(codes.InvalidArgument, "min_hourly_rate cannot be greater than max_hourly_rate")
	}

	updated, err := s.ProfileDetailsRepo.UpdateHiringPreferences(ctx, userID, current)
	if err != nil {
		return nil, toStatus(err)
	}

	return &userv1.PatchMyHiringPreferencesResponse{Preferences: toProtoHiringPreferences(updated)}, nil
}

func (s *UserServer) SaveFreelancer(ctx context.Context, req *userv1.SaveFreelancerRequest) (*userv1.SaveFreelancerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(strings.TrimSpace(req.GetUserId()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	freelancerUserID, err := uuid.Parse(strings.TrimSpace(req.GetFreelancerUserId()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid freelancer_user_id")
	}

	out, err := s.ProfileDetailsRepo.SaveFreelancer(ctx, userID, freelancerUserID)
	if err != nil {
		return nil, toStatus(err)
	}

	return &userv1.SaveFreelancerResponse{Saved: toProtoSavedFreelancer(out)}, nil
}

func (s *UserServer) ListSavedFreelancers(ctx context.Context, req *userv1.ListSavedFreelancersRequest) (*userv1.ListSavedFreelancersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(strings.TrimSpace(req.GetUserId()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	pageSize := uint32(20)
	pageToken := ""
	if req.GetPage() != nil {
		if req.GetPage().GetPageSize() > 0 {
			pageSize = req.GetPage().GetPageSize()
		}
		pageToken = strings.TrimSpace(req.GetPage().GetPageToken())
	}

	out, err := s.ProfileDetailsRepo.ListSavedFreelancers(ctx, userID, pageSize, pageToken)
	if err != nil {
		return nil, toStatus(err)
	}

	items := make([]*userv1.SavedFreelancer, 0, len(out.Items))
	for _, item := range out.Items {
		items = append(items, toProtoSavedFreelancer(item))
	}

	return &userv1.ListSavedFreelancersResponse{
		Freelancers: items,
		Page:        &userv1.PagingResponse{NextPageToken: out.NextPageToken},
	}, nil
}

func (s *UserServer) RemoveSavedFreelancer(ctx context.Context, req *userv1.RemoveSavedFreelancerRequest) (*userv1.RemoveSavedFreelancerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(strings.TrimSpace(req.GetUserId()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	freelancerUserID, err := uuid.Parse(strings.TrimSpace(req.GetFreelancerUserId()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid freelancer_user_id")
	}

	removed, err := s.ProfileDetailsRepo.RemoveSavedFreelancer(ctx, userID, freelancerUserID)
	if err != nil {
		return nil, toStatus(err)
	}

	return &userv1.RemoveSavedFreelancerResponse{Removed: removed}, nil
}

func (s *UserServer) UpsertFreelancerNote(ctx context.Context, req *userv1.UpsertFreelancerNoteRequest) (*userv1.UpsertFreelancerNoteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(strings.TrimSpace(req.GetUserId()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	freelancerUserID, err := uuid.Parse(strings.TrimSpace(req.GetFreelancerUserId()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid freelancer_user_id")
	}
	note := strings.TrimSpace(req.GetNote())
	if len(note) > 100 {
		return nil, status.Error(codes.InvalidArgument, "note exceeds max length of 100 characters")
	}

	out, err := s.ProfileDetailsRepo.UpsertFreelancerNote(ctx, userID, freelancerUserID, note)
	if err != nil {
		return nil, toStatus(err)
	}

	return &userv1.UpsertFreelancerNoteResponse{Note: toProtoFreelancerNote(out)}, nil
}

func (s *UserServer) GetFreelancerNote(ctx context.Context, req *userv1.GetFreelancerNoteRequest) (*userv1.GetFreelancerNoteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(strings.TrimSpace(req.GetUserId()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	freelancerUserID, err := uuid.Parse(strings.TrimSpace(req.GetFreelancerUserId()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid freelancer_user_id")
	}

	out, err := s.ProfileDetailsRepo.GetFreelancerNote(ctx, userID, freelancerUserID)
	if err != nil {
		return nil, toStatus(err)
	}

	return &userv1.GetFreelancerNoteResponse{Note: toProtoFreelancerNote(out)}, nil
}

func (s *UserServer) GetMyAvatarUploadUrl(ctx context.Context, req *userv1.GetMyAvatarUploadUrlRequest) (*userv1.GetMyAvatarUploadUrlResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	if s.GetAvatarUploadURLUC == nil {
		return nil, status.Error(codes.Internal, "avatar upload url use-case not configured")
	}
	out, err := s.GetAvatarUploadURLUC.Execute(ctx, application.GetAvatarUploadURLInput{
		UserID:      userID,
		FileName:    req.GetFileName(),
		ContentType: req.GetContentType(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.GetMyAvatarUploadUrlResponse{StorageKey: out.StorageKey, UploadUrl: out.UploadURL}, nil
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
		StorageKey:  req.GetStorageKey(),
		Width:       req.GetWidth(),
		Height:      req.GetHeight(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.UploadMyAvatarResponse{
		AvatarUrl: out.DownloadURL,
		Avatar: &userv1.ProfileAvatar{
			UserId:        req.UserId,
			FileName:      req.FileName,
			ContentType:   out.ContentType,
			StorageKey:    req.GetStorageKey(),
			SizeBytes:     out.SizeBytes,
			Width:         out.Width,
			Height:        out.Height,
			UpdatedAtUnix: time.Now().UTC().Unix(),
			DownloadUrl:   out.DownloadURL,
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
			SizeBytes:     out.SizeBytes,
			Width:         0,
			Height:        0,
			UpdatedAtUnix: 0,
			DownloadUrl:   out.DownloadURL,
		},
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

func (s *UserServer) GetMyCVUploadUrl(ctx context.Context, req *userv1.GetMyCVUploadUrlRequest) (*userv1.GetMyCVUploadUrlResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	if s.GetCVUploadURLUC == nil {
		return nil, status.Error(codes.Internal, "cv upload url use-case not configured")
	}
	out, err := s.GetCVUploadURLUC.Execute(ctx, application.GetCVUploadURLInput{
		UserID:      userID,
		FileName:    req.GetFileName(),
		ContentType: req.GetContentType(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.GetMyCVUploadUrlResponse{StorageKey: out.StorageKey, UploadUrl: out.UploadURL}, nil
}

func (s *UserServer) UpsertMyCV(ctx context.Context, req *userv1.UploadMyCVRequest) (*userv1.UploadMyCVResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	if s.UpsertCVUC == nil {
		return nil, status.Error(codes.Internal, "cv use-case not configured")
	}
	out, err := s.UpsertCVUC.Execute(ctx, application.UpsertCVInput{
		UserID:      userID,
		FileName:    req.GetFileName(),
		ContentType: req.GetContentType(),
		StorageKey:  req.GetStorageKey(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.UploadMyCVResponse{Cv: s.toProtoCV(out.CV, out.DownloadURL)}, nil
}

func (s *UserServer) GetMyCV(ctx context.Context, req *userv1.GetMyCVRequest) (*userv1.GetMyCVResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	if s.GetCVUC == nil {
		return nil, status.Error(codes.Internal, "cv use-case not configured")
	}
	out, err := s.GetCVUC.Execute(ctx, application.GetCVInput{UserID: userID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.GetMyCVResponse{Cv: s.toProtoCV(out.CV, out.DownloadURL)}, nil
}

func (s *UserServer) RemoveMyCV(ctx context.Context, req *userv1.RemoveMyCVRequest) (*userv1.RemoveMyCVResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	if s.RemoveCVUC == nil {
		return nil, status.Error(codes.Internal, "cv use-case not configured")
	}
	out, err := s.RemoveCVUC.Execute(ctx, application.RemoveCVInput{UserID: userID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &userv1.RemoveMyCVResponse{Removed: out.Removed}, nil
}

func (s *UserServer) CreateMyPortfolioItem(ctx context.Context, req *userv1.CreateMyPortfolioItemRequest) (*userv1.CreateMyPortfolioItemResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	if err := validatePortfolioItemInput(req.GetTitle(), req.GetDescription()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	media, err := toAppPortfolioMediaInputs(req.GetMedia())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	item, err := s.ProfileDetailsRepo.CreatePortfolioItem(ctx, userID, application.PortfolioItem{
		Title:         strings.TrimSpace(req.GetTitle()),
		Description:   strings.TrimSpace(req.GetDescription()),
		ProjectURL:    strings.TrimSpace(req.GetProjectUrl()),
		RoleInProject: strings.TrimSpace(req.GetRoleInProject()),
		CompletedAt:   unixPtr(req.GetCompletedAtUnix()),
		Tags:          req.GetTags(),
		Media:         media,
	})
	if err != nil {
		return nil, toStatus(err)
	}

	itemProto, err := s.toProtoPortfolioItem(ctx, item)
	if err != nil {
		return nil, err
	}
	return &userv1.CreateMyPortfolioItemResponse{Item: itemProto}, nil
}

func (s *UserServer) GetMyPortfolioItem(ctx context.Context, req *userv1.GetMyPortfolioItemRequest) (*userv1.GetMyPortfolioItemResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	item, err := s.ProfileDetailsRepo.GetPortfolioItem(ctx, userID, req.GetItemId())
	if err != nil {
		return nil, toStatus(err)
	}

	itemProto, err := s.toProtoPortfolioItem(ctx, item)
	if err != nil {
		return nil, err
	}
	return &userv1.GetMyPortfolioItemResponse{Item: itemProto}, nil
}

func (s *UserServer) UpdateMyPortfolioItem(ctx context.Context, req *userv1.UpdateMyPortfolioItemRequest) (*userv1.UpdateMyPortfolioItemResponse, error) {
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
		current.Title = strings.TrimSpace(req.GetTitle())
	}
	if req.Description != nil {
		current.Description = strings.TrimSpace(req.GetDescription())
	}
	if req.ProjectUrl != nil {
		current.ProjectURL = strings.TrimSpace(req.GetProjectUrl())
	}
	if req.RoleInProject != nil {
		current.RoleInProject = strings.TrimSpace(req.GetRoleInProject())
	}
	if req.CompletedAtUnix != nil {
		current.CompletedAt = unixPtr(req.GetCompletedAtUnix())
	}
	if req.Tags != nil {
		current.Tags = req.GetTags().GetValues()
	}
	if req.Media != nil {
		media, mapErr := toAppPortfolioMediaInputs(req.GetMedia().GetValues())
		if mapErr != nil {
			return nil, status.Error(codes.InvalidArgument, mapErr.Error())
		}
		current.Media = media
	}

	if err := validatePortfolioItemInput(current.Title, current.Description); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	item, err := s.ProfileDetailsRepo.UpdatePortfolioItem(ctx, userID, req.GetItemId(), current)
	if err != nil {
		return nil, toStatus(err)
	}

	itemProto, err := s.toProtoPortfolioItem(ctx, item)
	if err != nil {
		return nil, err
	}
	return &userv1.UpdateMyPortfolioItemResponse{Item: itemProto}, nil
}

func (s *UserServer) DeleteMyPortfolioItem(ctx context.Context, req *userv1.DeleteMyPortfolioItemRequest) (*userv1.DeleteMyPortfolioItemResponse, error) {
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

	return &userv1.DeleteMyPortfolioItemResponse{Deleted: deleted}, nil
}

func (s *UserServer) ListMyPortfolioItems(ctx context.Context, req *userv1.ListMyPortfolioItemsRequest) (*userv1.ListMyPortfolioItemsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}

	pageSize := uint32(20)
	pageToken := ""
	if req.GetPage() != nil {
		if req.GetPage().GetPageSize() > 0 {
			pageSize = req.GetPage().GetPageSize()
		}
		pageToken = strings.TrimSpace(req.GetPage().GetPageToken())
	}

	out, err := s.ProfileDetailsRepo.ListMyPortfolioItems(ctx, userID, pageSize, pageToken)
	if err != nil {
		return nil, toStatus(err)
	}

	items := make([]*userv1.PortfolioItem, 0, len(out.Items))
	for _, item := range out.Items {
		itemProto, mapErr := s.toProtoPortfolioItem(ctx, item)
		if mapErr != nil {
			return nil, toStatus(mapErr)
		}
		items = append(items, itemProto)
	}

	return &userv1.ListMyPortfolioItemsResponse{
		Items: items,
		Page:  &userv1.PagingResponse{NextPageToken: out.NextPageToken},
	}, nil
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
			ContactEmail:       profile.ContactEmail,
			ContactPhone:       profile.ContactPhone,
			Bio:                profile.Bio,
			Location:           profile.Location,
			AccountStatus:      toProtoAccountStatus(profile.AccountStatus),
			SuspensionReason:   profile.SuspensionReason,
			TaxId:              profile.TaxID,
			VerificationStatus: toProtoVerificationStatus(profile.VerificationStatus),
			CreatedAtUnix:      profile.CreatedAt.Unix(),
			UpdatedAtUnix:      profile.UpdatedAt.Unix(),
		},
		Capabilities: toProtoCapabilities(s.CapabilityPolicy, profile, client, freelancer),
	}

	if client != nil {
		out.Client = &userv1.ClientProfile{CompanyName: client.CompanyName}
	}
	if freelancer != nil {
		out.Freelancer = &userv1.FreelancerProfile{
			Headline:     freelancer.Headline,
			Skills:       freelancer.Skills,
			HourlyRate:   freelancer.HourlyRate,
			Availability: toProtoAvailability(freelancer.Availability),
			Metrics: &userv1.FreelancerMetrics{
				Rating:          freelancer.Rating,
				JobSuccessScore: freelancer.Reputation.JobSuccessScore,
				TotalReviews:    freelancer.Reputation.TotalReviews,
				TotalJobs:       freelancer.Reputation.TotalJobs,
				TotalEarnings:   freelancer.Reputation.TotalEarningsUSD,
			},
		}
		if profile.LastActiveAt != nil {
			lastActive := profile.LastActiveAt.Unix()
			out.Freelancer.Metrics.LastActiveAtUnix = &lastActive
		}
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
	case domain.VerificationStatusSubmitted:
		return userv1.VerificationStatus_VERIFICATION_STATUS_SUBMITTED
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
	case contains(msg, "role required"):
		return status.Error(codes.PermissionDenied, msg)
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

func toProtoSettings(in application.UserSettings) *userv1.UserSettings {
	return &userv1.UserSettings{
		UiLocale:                  in.UILocale,
		EmailNotificationsEnabled: in.EmailNotificationsEnabled,
		PushNotificationsEnabled:  in.PushNotificationsEnabled,
	}
}

func toProtoWorkPreferences(in application.WorkPreferences) *userv1.WorkPreferences {
	return &userv1.WorkPreferences{
		PreferredProjectLength: toProtoProjectLength(in.PreferredProjectLength),
		MinBudget:              in.MinBudgetUSD,
		MaxBudget:              in.MaxBudgetUSD,
		ContractTypes:          in.ContractTypes,
		WeeklyCapacityHours:    in.WeeklyCapacityHours,
	}
}

func fromProtoProjectLength(value userv1.ProjectLength) (string, bool) {
	switch value {
	case userv1.ProjectLength_PROJECT_LENGTH_UNSPECIFIED:
		return application.ProjectLengthUnspecified, true
	case userv1.ProjectLength_PROJECT_LENGTH_SHORT_TERM:
		return application.ProjectLengthShortTerm, true
	case userv1.ProjectLength_PROJECT_LENGTH_MEDIUM_TERM:
		return application.ProjectLengthMediumTerm, true
	case userv1.ProjectLength_PROJECT_LENGTH_LONG_TERM:
		return application.ProjectLengthLongTerm, true
	default:
		return "", false
	}
}

func toProtoProjectLength(value string) userv1.ProjectLength {
	canonical := application.CanonicalProjectLengthOrUnspecified(value)
	switch canonical {
	case application.ProjectLengthShortTerm:
		return userv1.ProjectLength_PROJECT_LENGTH_SHORT_TERM
	case application.ProjectLengthMediumTerm:
		return userv1.ProjectLength_PROJECT_LENGTH_MEDIUM_TERM
	case application.ProjectLengthLongTerm:
		return userv1.ProjectLength_PROJECT_LENGTH_LONG_TERM
	default:
		return userv1.ProjectLength_PROJECT_LENGTH_UNSPECIFIED
	}
}

func toProtoHiringPreferences(in application.HiringPreferences) *userv1.HiringPreferences {
	return &userv1.HiringPreferences{
		MinHourlyRate:      in.MinHourlyRate,
		MaxHourlyRate:      in.MaxHourlyRate,
		PreferredLocations: in.PreferredLocations,
	}
}

func toProtoSavedFreelancer(in application.SavedFreelancer) *userv1.SavedFreelancer {
	return &userv1.SavedFreelancer{
		FreelancerUserId: in.FreelancerUserID.String(),
		SavedAtUnix:      in.SavedAt.Unix(),
	}
}

func toProtoFreelancerNote(in application.FreelancerNote) *userv1.FreelancerNote {
	return &userv1.FreelancerNote{
		FreelancerUserId: in.FreelancerUserID.String(),
		Note:             in.Note,
		UpdatedAtUnix:    in.UpdatedAt.Unix(),
	}
}

func validatePortfolioItemInput(title, description string) error {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	if title == "" {
		return status.Error(codes.InvalidArgument, "title is required")
	}
	if len(title) > 50 {
		return status.Error(codes.InvalidArgument, "title exceeds max length of 50 characters")
	}
	if len(description) > 200 {
		return status.Error(codes.InvalidArgument, "description exceeds max length of 200 characters")
	}
	return nil
}

func toAppPortfolioMediaInputs(items []*userv1.PortfolioMediaInput) ([]application.PortfolioMedia, error) {
	out := make([]application.PortfolioMedia, 0, len(items))
	for _, media := range items {
		if media == nil {
			continue
		}
		storageKey := strings.TrimSpace(media.GetStorageKey())
		externalURL := strings.TrimSpace(media.GetExternalUrl())
		mediaType := media.GetMediaType()

		switch mediaType {
		case userv1.PortfolioMediaType_PORTFOLIO_MEDIA_TYPE_LINK:
			if externalURL == "" {
				return nil, status.Error(codes.InvalidArgument, "external_url is required for LINK media")
			}
			if storageKey != "" {
				return nil, status.Error(codes.InvalidArgument, "storage_key must be empty for LINK media")
			}
			if _, err := url.ParseRequestURI(externalURL); err != nil {
				return nil, status.Error(codes.InvalidArgument, "external_url must be a valid URL")
			}
		case userv1.PortfolioMediaType_PORTFOLIO_MEDIA_TYPE_IMAGE,
			userv1.PortfolioMediaType_PORTFOLIO_MEDIA_TYPE_VIDEO,
			userv1.PortfolioMediaType_PORTFOLIO_MEDIA_TYPE_FILE:
			if storageKey == "" {
				return nil, status.Error(codes.InvalidArgument, "storage_key is required for upload media")
			}
			if externalURL != "" {
				return nil, status.Error(codes.InvalidArgument, "external_url is not allowed for upload media")
			}
		default:
			return nil, status.Error(codes.InvalidArgument, "unsupported portfolio media_type")
		}

		out = append(out, application.PortfolioMedia{
			MediaType:   mediaType.String(),
			StorageKey:  storageKey,
			ExternalURL: externalURL,
			FileName:    strings.TrimSpace(media.GetFileName()),
			ContentType: strings.TrimSpace(media.GetContentType()),
			SizeBytes:   media.GetSizeBytes(),
			Width:       media.GetWidth(),
			Height:      media.GetHeight(),
		})
	}
	return out, nil
}

func (s *UserServer) toProtoCV(cv application.CV, downloadURL string) *userv1.ProfileCV {
	return &userv1.ProfileCV{
		UserId:        cv.UserID.String(),
		FileName:      cv.FileName,
		ContentType:   cv.ContentType,
		SizeBytes:     cv.SizeBytes,
		UpdatedAtUnix: cv.UpdatedAt.Unix(),
		DownloadUrl:   strings.TrimSpace(downloadURL),
	}
}

func (s *UserServer) toProtoPortfolioItem(ctx context.Context, item application.PortfolioItem) (*userv1.PortfolioItem, error) {
	media := make([]*userv1.PortfolioMedia, 0, len(item.Media))
	for _, m := range item.Media {
		externalURL := strings.TrimSpace(m.ExternalURL)
		storageKey := strings.TrimSpace(m.StorageKey)
		if storageKey != "" {
			if s.PortfolioStore == nil {
				return nil, status.Error(codes.Internal, "portfolio store not configured")
			}
			presigned, err := s.PortfolioStore.PresignGetObject(ctx, storageKey, time.Hour)
			if err != nil {
				return nil, err
			}
			externalURL = presigned
			storageKey = ""
		}
		media = append(media, &userv1.PortfolioMedia{
			Id:          m.ID,
			MediaType:   mapPortfolioMediaTypeToProto(m.MediaType),
			StorageKey:  storageKey,
			ExternalUrl: externalURL,
			FileName:    m.FileName,
			ContentType: m.ContentType,
			SizeBytes:   m.SizeBytes,
			Width:       m.Width,
			Height:      m.Height,
		})
	}

	resp := &userv1.PortfolioItem{
		Id:            item.ID,
		UserId:        item.UserID.String(),
		Title:         item.Title,
		Description:   item.Description,
		ProjectUrl:    item.ProjectURL,
		RoleInProject: item.RoleInProject,
		Tags:          item.Tags,
		Media:         media,
		CreatedAtUnix: item.CreatedAt.Unix(),
		UpdatedAtUnix: item.UpdatedAt.Unix(),
	}
	if item.CompletedAt != nil {
		completed := item.CompletedAt.Unix()
		resp.CompletedAtUnix = &completed
	}
	return resp, nil
}

func mapPortfolioMediaTypeToProto(value string) userv1.PortfolioMediaType {
	canonical := strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(value)), "PORTFOLIO_MEDIA_TYPE_")
	switch canonical {
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

func unixPtr(v int64) *time.Time {
	t := time.Unix(v, 0).UTC()
	return &t
}

func hasClearField(fields []string, target string) bool {
	for _, field := range fields {
		if strings.EqualFold(strings.TrimSpace(field), target) {
			return true
		}
	}
	return false
}
