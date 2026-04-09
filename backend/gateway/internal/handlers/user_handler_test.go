package handlers

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"path/filepath"
	"strings"
	"testing"

	"jobconnect/gateway/internal/middleware"
	userv1 "jobconnect/user/gen/user"
	verificationv1 "jobconnect/verification/gen/verification/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func newJSONTestContext(method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(method, target, nil)
	return ctx, rec
}

func newJSONBodyTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(method, target, bytes.NewBufferString(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	return ctx, rec
}

func newMultipartBodyTestContext(method, target, fieldName, fileName string, content []byte) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	contentType := mime.TypeByExtension(filepath.Ext(fileName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": {fmt.Sprintf(`form-data; name=%q; filename=%q`, fieldName, fileName)},
		"Content-Type":        {contentType},
	})
	if err == nil {
		_, _ = part.Write(content)
	}
	_ = writer.Close()

	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(method, target, body)
	ctx.Request.Header.Set("Content-Type", writer.FormDataContentType())
	return ctx, rec
}

type fullUserServiceServerStub struct {
	userv1.UnimplementedUserServiceServer

	getMyProfileReq          *userv1.GetMyProfileRequest
	patchMyProfileReq        *userv1.PatchMyProfileRequest
	deleteMyProfileReq       *userv1.DeleteMyProfileRequest
	getMyOnboardingStatusReq *userv1.GetMyOnboardingStatusRequest

	getMySettingsReq        *userv1.GetMySettingsRequest
	patchMySettingsReq      *userv1.PatchMySettingsRequest
	getMyAvatarUploadUrlReq *userv1.GetMyAvatarUploadUrlRequest

	upsertMyAvatarReq *userv1.UploadMyAvatarRequest
	getMyAvatarReq    *userv1.GetMyAvatarRequest
	removeMyAvatarReq *userv1.RemoveMyAvatarRequest

	getMyCVUploadUrlReq *userv1.GetMyCVUploadUrlRequest

	upsertMyCVReq *userv1.UploadMyCVRequest
	getMyCVReq    *userv1.GetMyCVRequest
	removeMyCVReq *userv1.RemoveMyCVRequest

	createMyPortfolioItemReq *userv1.CreateMyPortfolioItemRequest
	listMyPortfolioItemsReq  *userv1.ListMyPortfolioItemsRequest
	getMyPortfolioItemReq    *userv1.GetMyPortfolioItemRequest
	updateMyPortfolioItemReq *userv1.UpdateMyPortfolioItemRequest
	deleteMyPortfolioItemReq *userv1.DeleteMyPortfolioItemRequest

	patchMyWorkPreferencesReq *userv1.PatchMyWorkPreferencesRequest
	getMyWorkPreferencesReq   *userv1.GetMyWorkPreferencesRequest

	getMyHiringPreferencesReq   *userv1.GetMyHiringPreferencesRequest
	patchMyHiringPreferencesReq *userv1.PatchMyHiringPreferencesRequest

	saveFreelancerReq        *userv1.SaveFreelancerRequest
	listSavedFreelancersReq  *userv1.ListSavedFreelancersRequest
	removeSavedFreelancerReq *userv1.RemoveSavedFreelancerRequest
	upsertFreelancerNoteReq  *userv1.UpsertFreelancerNoteRequest
	getFreelancerNoteReq     *userv1.GetFreelancerNoteRequest
}

func (s *fullUserServiceServerStub) GetMyProfile(ctx context.Context, in *userv1.GetMyProfileRequest) (*userv1.GetMyProfileResponse, error) {
	s.getMyProfileReq = in
	return &userv1.GetMyProfileResponse{
		Profile:      &userv1.UserProfile{Core: &userv1.UserCore{UserId: in.GetUserId(), DisplayName: "Tester"}},
		Completeness: &userv1.ProfileCompleteness{Percent: 80},
	}, nil
}

func (s *fullUserServiceServerStub) PatchMyProfile(ctx context.Context, in *userv1.PatchMyProfileRequest) (*userv1.PatchMyProfileResponse, error) {
	s.patchMyProfileReq = in
	return &userv1.PatchMyProfileResponse{
		Profile:      &userv1.UserProfile{Core: &userv1.UserCore{UserId: in.GetUserId(), DisplayName: in.GetCore().GetDisplayName()}},
		Completeness: &userv1.ProfileCompleteness{Percent: 85},
	}, nil
}

func (s *fullUserServiceServerStub) DeleteMyProfile(ctx context.Context, in *userv1.DeleteMyProfileRequest) (*userv1.DeleteMyProfileResponse, error) {
	s.deleteMyProfileReq = in
	return &userv1.DeleteMyProfileResponse{Deleted: true}, nil
}

func (s *fullUserServiceServerStub) GetMyOnboardingStatus(ctx context.Context, in *userv1.GetMyOnboardingStatusRequest) (*userv1.GetMyOnboardingStatusResponse, error) {
	s.getMyOnboardingStatusReq = in
	return &userv1.GetMyOnboardingStatusResponse{
		Completeness: &userv1.ProfileCompleteness{Percent: 90},
		Steps: []*userv1.OnboardingStep{{
			Key:    "profile",
			Status: userv1.OnboardingStepStatus_ONBOARDING_STEP_STATUS_COMPLETED,
		}},
		Readiness: &userv1.ProfileReadiness{Percent: 90},
	}, nil
}

func (s *fullUserServiceServerStub) GetMySettings(ctx context.Context, in *userv1.GetMySettingsRequest) (*userv1.GetMySettingsResponse, error) {
	s.getMySettingsReq = in
	return &userv1.GetMySettingsResponse{Settings: &userv1.UserSettings{UiLocale: "en-US", EmailNotificationsEnabled: true, PushNotificationsEnabled: false}}, nil
}

func (s *fullUserServiceServerStub) PatchMySettings(ctx context.Context, in *userv1.PatchMySettingsRequest) (*userv1.PatchMySettingsResponse, error) {
	s.patchMySettingsReq = in
	return &userv1.PatchMySettingsResponse{Settings: &userv1.UserSettings{UiLocale: in.GetUiLocale(), EmailNotificationsEnabled: in.GetEmailNotificationsEnabled(), PushNotificationsEnabled: in.GetPushNotificationsEnabled()}}, nil
}

func (s *fullUserServiceServerStub) GetMyAvatarUploadUrl(ctx context.Context, in *userv1.GetMyAvatarUploadUrlRequest) (*userv1.GetMyAvatarUploadUrlResponse, error) {
	s.getMyAvatarUploadUrlReq = in
	return &userv1.GetMyAvatarUploadUrlResponse{StorageKey: "avatars/test/current", UploadUrl: "https://upload.test/avatar"}, nil
}

func (s *fullUserServiceServerStub) UpsertMyAvatar(ctx context.Context, in *userv1.UploadMyAvatarRequest) (*userv1.UploadMyAvatarResponse, error) {
	s.upsertMyAvatarReq = in
	return &userv1.UploadMyAvatarResponse{AvatarUrl: "https://cdn.test/avatar.png", Avatar: &userv1.ProfileAvatar{UserId: in.GetUserId(), FileName: in.GetFileName(), ContentType: in.GetContentType(), StorageKey: in.GetStorageKey(), SizeBytes: 123, DownloadUrl: "https://cdn.test/avatar.png"}}, nil
}

func (s *fullUserServiceServerStub) GetMyAvatar(ctx context.Context, in *userv1.GetMyAvatarRequest) (*userv1.GetMyAvatarResponse, error) {
	s.getMyAvatarReq = in
	return &userv1.GetMyAvatarResponse{Avatar: &userv1.ProfileAvatar{FileName: "avatar.png", ContentType: "image/png", DownloadUrl: "https://cdn.test/avatar.png"}}, nil
}

func (s *fullUserServiceServerStub) RemoveMyAvatar(ctx context.Context, in *userv1.RemoveMyAvatarRequest) (*userv1.RemoveMyAvatarResponse, error) {
	s.removeMyAvatarReq = in
	return &userv1.RemoveMyAvatarResponse{Removed: true}, nil
}

func (s *fullUserServiceServerStub) GetMyCVUploadUrl(ctx context.Context, in *userv1.GetMyCVUploadUrlRequest) (*userv1.GetMyCVUploadUrlResponse, error) {
	s.getMyCVUploadUrlReq = in
	return &userv1.GetMyCVUploadUrlResponse{StorageKey: "cvs/test/current", UploadUrl: "https://upload.test/cv"}, nil
}

func (s *fullUserServiceServerStub) UpsertMyCV(ctx context.Context, in *userv1.UploadMyCVRequest) (*userv1.UploadMyCVResponse, error) {
	s.upsertMyCVReq = in
	return &userv1.UploadMyCVResponse{Cv: &userv1.ProfileCV{UserId: in.GetUserId(), FileName: in.GetFileName(), ContentType: in.GetContentType(), SizeBytes: 123, DownloadUrl: "https://cdn.test/resume.pdf"}}, nil
}

func (s *fullUserServiceServerStub) GetMyCV(ctx context.Context, in *userv1.GetMyCVRequest) (*userv1.GetMyCVResponse, error) {
	s.getMyCVReq = in
	return &userv1.GetMyCVResponse{Cv: &userv1.ProfileCV{UserId: in.GetUserId(), FileName: "resume.pdf", DownloadUrl: "https://cdn.test/resume.pdf"}}, nil
}

func (s *fullUserServiceServerStub) RemoveMyCV(ctx context.Context, in *userv1.RemoveMyCVRequest) (*userv1.RemoveMyCVResponse, error) {
	s.removeMyCVReq = in
	return &userv1.RemoveMyCVResponse{Removed: true}, nil
}

func (s *fullUserServiceServerStub) CreateMyPortfolioItem(ctx context.Context, in *userv1.CreateMyPortfolioItemRequest) (*userv1.CreateMyPortfolioItemResponse, error) {
	s.createMyPortfolioItemReq = in
	return &userv1.CreateMyPortfolioItemResponse{Item: &userv1.PortfolioItem{Id: 101, UserId: in.GetUserId(), Title: in.GetTitle()}}, nil
}

func (s *fullUserServiceServerStub) ListMyPortfolioItems(ctx context.Context, in *userv1.ListMyPortfolioItemsRequest) (*userv1.ListMyPortfolioItemsResponse, error) {
	s.listMyPortfolioItemsReq = in
	return &userv1.ListMyPortfolioItemsResponse{Items: []*userv1.PortfolioItem{{Id: 101, UserId: in.GetUserId(), Title: "One"}, {Id: 102, UserId: in.GetUserId(), Title: "Two"}}, Page: &userv1.PagingResponse{NextPageToken: "next-token"}}, nil
}

func (s *fullUserServiceServerStub) GetMyPortfolioItem(ctx context.Context, in *userv1.GetMyPortfolioItemRequest) (*userv1.GetMyPortfolioItemResponse, error) {
	s.getMyPortfolioItemReq = in
	return &userv1.GetMyPortfolioItemResponse{Item: &userv1.PortfolioItem{Id: in.GetItemId(), UserId: in.GetUserId(), Title: "Single Item"}}, nil
}

func (s *fullUserServiceServerStub) UpdateMyPortfolioItem(ctx context.Context, in *userv1.UpdateMyPortfolioItemRequest) (*userv1.UpdateMyPortfolioItemResponse, error) {
	s.updateMyPortfolioItemReq = in
	return &userv1.UpdateMyPortfolioItemResponse{Item: &userv1.PortfolioItem{Id: in.GetItemId(), UserId: in.GetUserId(), Title: in.GetTitle()}}, nil
}

func (s *fullUserServiceServerStub) DeleteMyPortfolioItem(ctx context.Context, in *userv1.DeleteMyPortfolioItemRequest) (*userv1.DeleteMyPortfolioItemResponse, error) {
	s.deleteMyPortfolioItemReq = in
	return &userv1.DeleteMyPortfolioItemResponse{Deleted: true}, nil
}

func (s *fullUserServiceServerStub) PatchMyWorkPreferences(ctx context.Context, in *userv1.PatchMyWorkPreferencesRequest) (*userv1.PatchMyWorkPreferencesResponse, error) {
	s.patchMyWorkPreferencesReq = in
	return &userv1.PatchMyWorkPreferencesResponse{Settings: &userv1.WorkPreferences{PreferredProjectLength: in.GetPreferredProjectLength(), MinBudget: in.GetMinBudget(), MaxBudget: in.GetMaxBudget(), ContractTypes: in.GetContractTypes().GetValues(), WeeklyCapacityHours: in.GetWeeklyCapacityHours()}}, nil
}

func (s *fullUserServiceServerStub) GetMyWorkPreferences(ctx context.Context, in *userv1.GetMyWorkPreferencesRequest) (*userv1.GetMyWorkPreferencesResponse, error) {
	s.getMyWorkPreferencesReq = in
	return &userv1.GetMyWorkPreferencesResponse{Settings: &userv1.WorkPreferences{PreferredProjectLength: userv1.ProjectLength_PROJECT_LENGTH_SHORT_TERM}}, nil
}

func (s *fullUserServiceServerStub) GetMyHiringPreferences(ctx context.Context, in *userv1.GetMyHiringPreferencesRequest) (*userv1.GetMyHiringPreferencesResponse, error) {
	s.getMyHiringPreferencesReq = in
	return &userv1.GetMyHiringPreferencesResponse{Preferences: &userv1.HiringPreferences{MinHourlyRate: 20, MaxHourlyRate: 80, PreferredLocations: []string{"US"}}}, nil
}

func (s *fullUserServiceServerStub) PatchMyHiringPreferences(ctx context.Context, in *userv1.PatchMyHiringPreferencesRequest) (*userv1.PatchMyHiringPreferencesResponse, error) {
	s.patchMyHiringPreferencesReq = in
	return &userv1.PatchMyHiringPreferencesResponse{Preferences: &userv1.HiringPreferences{MinHourlyRate: in.GetMinHourlyRate(), MaxHourlyRate: in.GetMaxHourlyRate(), PreferredLocations: in.GetPreferredLocations().GetValues()}}, nil
}

func (s *fullUserServiceServerStub) SaveFreelancer(ctx context.Context, in *userv1.SaveFreelancerRequest) (*userv1.SaveFreelancerResponse, error) {
	s.saveFreelancerReq = in
	return &userv1.SaveFreelancerResponse{Saved: &userv1.SavedFreelancer{FreelancerUserId: in.GetFreelancerUserId(), SavedAtUnix: 1}}, nil
}

func (s *fullUserServiceServerStub) ListSavedFreelancers(ctx context.Context, in *userv1.ListSavedFreelancersRequest) (*userv1.ListSavedFreelancersResponse, error) {
	s.listSavedFreelancersReq = in
	return &userv1.ListSavedFreelancersResponse{Freelancers: []*userv1.SavedFreelancer{{FreelancerUserId: "f1"}, {FreelancerUserId: "f2"}}, Page: &userv1.PagingResponse{NextPageToken: "saved-next"}}, nil
}

func (s *fullUserServiceServerStub) RemoveSavedFreelancer(ctx context.Context, in *userv1.RemoveSavedFreelancerRequest) (*userv1.RemoveSavedFreelancerResponse, error) {
	s.removeSavedFreelancerReq = in
	return &userv1.RemoveSavedFreelancerResponse{Removed: true}, nil
}

func (s *fullUserServiceServerStub) UpsertFreelancerNote(ctx context.Context, in *userv1.UpsertFreelancerNoteRequest) (*userv1.UpsertFreelancerNoteResponse, error) {
	s.upsertFreelancerNoteReq = in
	return &userv1.UpsertFreelancerNoteResponse{Note: &userv1.FreelancerNote{FreelancerUserId: in.GetFreelancerUserId(), Note: in.GetNote(), UpdatedAtUnix: 1}}, nil
}

func (s *fullUserServiceServerStub) GetFreelancerNote(ctx context.Context, in *userv1.GetFreelancerNoteRequest) (*userv1.GetFreelancerNoteResponse, error) {
	s.getFreelancerNoteReq = in
	return &userv1.GetFreelancerNoteResponse{Note: &userv1.FreelancerNote{FreelancerUserId: in.GetFreelancerUserId(), Note: "saved note", UpdatedAtUnix: 2}}, nil
}

func TestUserHandler_PostmanStyleEndpointCoverage(t *testing.T) {
	stub := &fullUserServiceServerStub{}
	client, cleanup := newBufConnUserClient(t, stub)
	defer cleanup()

	h := NewUserHandler(client, nil)
	const userID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	t.Run("GetMyProfile", func(t *testing.T) {
		ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/users/me/profile")
		ctx.Set(middleware.ContextUserID, userID)
		h.GetMe(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.getMyProfileReq == nil || stub.getMyProfileReq.GetUserId() != userID {
			t.Fatalf("expected GetMyProfile request user_id=%q", userID)
		}
	})

	t.Run("PatchMyProfile", func(t *testing.T) {
		ctx, rec := newJSONBodyTestContext(http.MethodPatch, "/api/v1/users/me/profile", `{"display_name":"New Name"}`)
		ctx.Set(middleware.ContextUserID, userID)
		h.UpdateMeProfile(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.patchMyProfileReq == nil || stub.patchMyProfileReq.GetUserId() != userID {
			t.Fatalf("expected PatchMyProfile request user_id=%q", userID)
		}
	})

	t.Run("DeleteMyProfile", func(t *testing.T) {
		ctx, rec := newJSONTestContext(http.MethodDelete, "/api/v1/users/me/profile?hard_delete=true")
		ctx.Set(middleware.ContextUserID, userID)
		h.DeleteMeProfile(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.deleteMyProfileReq == nil || !stub.deleteMyProfileReq.GetHardDelete() {
			t.Fatalf("expected DeleteMyProfile hard_delete=true")
		}
	})

	t.Run("GetMyOnboardingStatus", func(t *testing.T) {
		ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/users/me/onboarding-status")
		ctx.Set(middleware.ContextUserID, userID)
		h.GetMeOnboardingStatus(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.getMyOnboardingStatusReq == nil || stub.getMyOnboardingStatusReq.GetUserId() != userID {
			t.Fatalf("expected GetMyOnboardingStatus request user_id=%q", userID)
		}
	})

	t.Run("GetMySettings", func(t *testing.T) {
		ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/users/me/settings")
		ctx.Set(middleware.ContextUserID, userID)
		h.GetMeAccountSettings(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.getMySettingsReq == nil || stub.getMySettingsReq.GetUserId() != userID {
			t.Fatalf("expected GetMySettings request user_id=%q", userID)
		}
	})

	t.Run("GetMyAvatarUploadUrl", func(t *testing.T) {
		ctx, rec := newJSONBodyTestContext(http.MethodPost, "/api/v1/users/me/avatar/upload-url", `{"file_name":"avatar.png","content_type":"image/png"}`)
		ctx.Set(middleware.ContextUserID, userID)
		h.GetMeAvatarUploadUrl(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.getMyAvatarUploadUrlReq == nil || stub.getMyAvatarUploadUrlReq.GetUserId() != userID {
			t.Fatalf("expected GetMyAvatarUploadUrl request with trusted user_id")
		}
	})

	t.Run("PatchMySettings", func(t *testing.T) {
		ctx, rec := newJSONBodyTestContext(http.MethodPatch, "/api/v1/users/me/settings", `{"ui_locale":"vi-VN","email_notifications_enabled":true}`)
		ctx.Set(middleware.ContextUserID, userID)
		h.UpdateMeAccountSettings(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.patchMySettingsReq == nil || stub.patchMySettingsReq.GetUserId() != userID || stub.patchMySettingsReq.GetUiLocale() != "vi-VN" {
			t.Fatalf("expected PatchMySettings request to be captured with trusted user_id and ui_locale")
		}
	})

	t.Run("UpsertMyAvatar", func(t *testing.T) {
		ctx, rec := newJSONBodyTestContext(http.MethodPost, "/api/v1/users/me/avatar", `{"storage_key":"avatars/test/current","file_name":"avatar.png","content_type":"image/png","width":64,"height":64}`)
		ctx.Set(middleware.ContextUserID, userID)
		h.UploadMeAvatar(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.upsertMyAvatarReq == nil || stub.upsertMyAvatarReq.GetUserId() != userID || stub.upsertMyAvatarReq.GetStorageKey() == "" {
			t.Fatalf("expected UpsertMyAvatar request with storage_key and trusted user_id")
		}
	})

	t.Run("GetMyAvatar", func(t *testing.T) {
		ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/users/me/avatar")
		ctx.Set(middleware.ContextUserID, userID)
		h.GetMeAvatar(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.getMyAvatarReq == nil || stub.getMyAvatarReq.GetUserId() != userID {
			t.Fatalf("expected GetMyAvatar request user_id=%q", userID)
		}
		if !strings.Contains(rec.Body.String(), "download_url") {
			t.Fatalf("expected avatar json body, got %q", rec.Body.String())
		}
	})

	t.Run("RemoveMyAvatar", func(t *testing.T) {
		ctx, rec := newJSONTestContext(http.MethodDelete, "/api/v1/users/me/avatar")
		ctx.Set(middleware.ContextUserID, userID)
		h.RemoveMeAvatar(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.removeMyAvatarReq == nil || stub.removeMyAvatarReq.GetUserId() != userID {
			t.Fatalf("expected RemoveMyAvatar request user_id=%q", userID)
		}
	})

	t.Run("GetMyCVUploadUrl", func(t *testing.T) {
		ctx, rec := newJSONBodyTestContext(http.MethodPost, "/api/v1/users/me/cv/upload-url", `{"file_name":"resume.pdf","content_type":"application/pdf"}`)
		ctx.Set(middleware.ContextUserID, userID)
		h.GetMeCVUploadUrl(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.getMyCVUploadUrlReq == nil || stub.getMyCVUploadUrlReq.GetUserId() != userID {
			t.Fatalf("expected GetMyCVUploadUrl request with trusted user_id")
		}
	})

	t.Run("UpsertMyCV", func(t *testing.T) {
		ctx, rec := newJSONBodyTestContext(http.MethodPost, "/api/v1/users/me/cv", `{"storage_key":"cvs/test/current","file_name":"resume.pdf","content_type":"application/pdf"}`)
		ctx.Set(middleware.ContextUserID, userID)
		h.UploadMeCV(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.upsertMyCVReq == nil || stub.upsertMyCVReq.GetUserId() != userID || stub.upsertMyCVReq.GetStorageKey() == "" {
			t.Fatalf("expected UpsertMyCV request with storage_key and trusted user_id")
		}
	})

	t.Run("GetMyCV", func(t *testing.T) {
		ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/users/me/cv")
		ctx.Set(middleware.ContextUserID, userID)
		h.GetMeCV(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.getMyCVReq == nil || stub.getMyCVReq.GetUserId() != userID {
			t.Fatalf("expected GetMyCV request user_id=%q", userID)
		}
	})

	t.Run("RemoveMyCV", func(t *testing.T) {
		ctx, rec := newJSONTestContext(http.MethodDelete, "/api/v1/users/me/cv")
		ctx.Set(middleware.ContextUserID, userID)
		h.RemoveMeCV(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.removeMyCVReq == nil || stub.removeMyCVReq.GetUserId() != userID {
			t.Fatalf("expected RemoveMyCV request user_id=%q", userID)
		}
	})

	t.Run("CreateMyPortfolioItem", func(t *testing.T) {
		ctx, rec := newJSONBodyTestContext(http.MethodPost, "/api/v1/users/me/portfolio", `{"user_id":"override","title":"Case Study","description":"Desc"}`)
		ctx.Set(middleware.ContextUserID, userID)
		h.CreateMePortfolioItem(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.createMyPortfolioItemReq == nil || stub.createMyPortfolioItemReq.GetUserId() != userID {
			t.Fatalf("expected trusted user_id for CreateMyPortfolioItem")
		}
	})

	t.Run("ListMyPortfolioItems", func(t *testing.T) {
		ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/users/me/portfolio?page_size=10&page_token=abc")
		ctx.Set(middleware.ContextUserID, userID)
		h.ListMePortfolioItems(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.listMyPortfolioItemsReq == nil || stub.listMyPortfolioItemsReq.GetUserId() != userID || stub.listMyPortfolioItemsReq.GetPage().GetPageSize() != 10 {
			t.Fatalf("expected ListMyPortfolioItems request with pagination and trusted user_id")
		}
	})

	t.Run("GetMyPortfolioItem", func(t *testing.T) {
		ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/users/me/portfolio/123")
		ctx.Set(middleware.ContextUserID, userID)
		ctx.Params = gin.Params{{Key: "itemId", Value: "123"}}
		h.GetMePortfolioItem(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.getMyPortfolioItemReq == nil || stub.getMyPortfolioItemReq.GetUserId() != userID || stub.getMyPortfolioItemReq.GetItemId() != 123 {
			t.Fatalf("expected GetMyPortfolioItem request with trusted user_id and item_id")
		}
	})

	t.Run("UpdateMyPortfolioItem", func(t *testing.T) {
		ctx, rec := newJSONBodyTestContext(http.MethodPut, "/api/v1/users/me/portfolio/123", `{"user_id":"override","item_id":999,"title":"Updated"}`)
		ctx.Set(middleware.ContextUserID, userID)
		ctx.Params = gin.Params{{Key: "itemId", Value: "123"}}
		h.UpdateMePortfolioItem(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.updateMyPortfolioItemReq == nil || stub.updateMyPortfolioItemReq.GetUserId() != userID || stub.updateMyPortfolioItemReq.GetItemId() != 123 {
			t.Fatalf("expected trusted user_id and trusted path item_id for UpdateMyPortfolioItem")
		}
	})

	t.Run("DeleteMyPortfolioItem", func(t *testing.T) {
		ctx, rec := newJSONTestContext(http.MethodDelete, "/api/v1/users/me/portfolio/123")
		ctx.Set(middleware.ContextUserID, userID)
		ctx.Params = gin.Params{{Key: "itemId", Value: "123"}}
		h.DeleteMePortfolioItem(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.deleteMyPortfolioItemReq == nil || stub.deleteMyPortfolioItemReq.GetUserId() != userID || stub.deleteMyPortfolioItemReq.GetItemId() != 123 {
			t.Fatalf("expected DeleteMyPortfolioItem request with trusted fields")
		}
	})

	t.Run("PatchMyWorkPreferences", func(t *testing.T) {
		ctx, rec := newJSONBodyTestContext(http.MethodPatch, "/api/v1/users/me/work-preferences", `{"preferred_project_length":"short_term","min_budget":100,"max_budget":250,"contract_types":{"values":["fixed"]},"weekly_capacity_hours":20}`)
		ctx.Set(middleware.ContextUserID, userID)
		h.SetMeWorkPreferences(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.patchMyWorkPreferencesReq == nil || stub.patchMyWorkPreferencesReq.GetUserId() != userID || stub.patchMyWorkPreferencesReq.GetPreferredProjectLength() != userv1.ProjectLength_PROJECT_LENGTH_SHORT_TERM {
			t.Fatalf("expected normalized preferred_project_length and trusted user_id")
		}
	})

	t.Run("GetMyWorkPreferences", func(t *testing.T) {
		ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/users/me/work-preferences")
		ctx.Set(middleware.ContextUserID, userID)
		h.GetMeWorkPreferences(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.getMyWorkPreferencesReq == nil || stub.getMyWorkPreferencesReq.GetUserId() != userID {
			t.Fatalf("expected GetMyWorkPreferences request user_id=%q", userID)
		}
	})

	t.Run("GetMyHiringPreferences", func(t *testing.T) {
		ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/users/me/hiring-preferences")
		ctx.Set(middleware.ContextUserID, userID)
		h.GetMeHiringPreferences(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.getMyHiringPreferencesReq == nil || stub.getMyHiringPreferencesReq.GetUserId() != userID {
			t.Fatalf("expected GetMyHiringPreferences request user_id=%q", userID)
		}
	})

	t.Run("PatchMyHiringPreferences", func(t *testing.T) {
		ctx, rec := newJSONBodyTestContext(http.MethodPatch, "/api/v1/users/me/hiring-preferences", `{"min_hourly_rate":15,"max_hourly_rate":65,"preferred_locations":{"values":["US","VN"]}}`)
		ctx.Set(middleware.ContextUserID, userID)
		h.UpdateMeHiringPreferences(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.patchMyHiringPreferencesReq == nil || stub.patchMyHiringPreferencesReq.GetUserId() != userID || stub.patchMyHiringPreferencesReq.GetMinHourlyRate() != 15 {
			t.Fatalf("expected PatchMyHiringPreferences request with trusted user_id")
		}
	})

	t.Run("SaveFreelancer", func(t *testing.T) {
		ctx, rec := newJSONTestContext(http.MethodPost, "/api/v1/users/me/saved-freelancers/f-1")
		ctx.Set(middleware.ContextUserID, userID)
		ctx.Params = gin.Params{{Key: "freelancerId", Value: "f-1"}}
		h.SaveMeFreelancer(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.saveFreelancerReq == nil || stub.saveFreelancerReq.GetUserId() != userID || stub.saveFreelancerReq.GetFreelancerUserId() != "f-1" {
			t.Fatalf("expected SaveFreelancer request with trusted user and path freelancer")
		}
	})

	t.Run("ListSavedFreelancers", func(t *testing.T) {
		ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/users/me/saved-freelancers?page_size=5&page_token=cursor")
		ctx.Set(middleware.ContextUserID, userID)
		h.ListMeSavedFreelancers(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.listSavedFreelancersReq == nil || stub.listSavedFreelancersReq.GetUserId() != userID || stub.listSavedFreelancersReq.GetPage().GetPageSize() != 5 {
			t.Fatalf("expected ListSavedFreelancers request with pagination and trusted user_id")
		}
	})

	t.Run("RemoveSavedFreelancer", func(t *testing.T) {
		ctx, rec := newJSONTestContext(http.MethodDelete, "/api/v1/users/me/saved-freelancers/f-2")
		ctx.Set(middleware.ContextUserID, userID)
		ctx.Params = gin.Params{{Key: "freelancerId", Value: "f-2"}}
		h.RemoveMeSavedFreelancer(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.removeSavedFreelancerReq == nil || stub.removeSavedFreelancerReq.GetUserId() != userID || stub.removeSavedFreelancerReq.GetFreelancerUserId() != "f-2" {
			t.Fatalf("expected RemoveSavedFreelancer request with trusted user and path freelancer")
		}
	})

	t.Run("UpsertFreelancerNote", func(t *testing.T) {
		ctx, rec := newJSONBodyTestContext(http.MethodPut, "/api/v1/users/me/freelancer-notes/f-3", `{"user_id":"override","freelancer_user_id":"override","note":"Strong communication"}`)
		ctx.Set(middleware.ContextUserID, userID)
		ctx.Params = gin.Params{{Key: "freelancerId", Value: "f-3"}}
		h.UpsertMeFreelancerNote(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.upsertFreelancerNoteReq == nil || stub.upsertFreelancerNoteReq.GetUserId() != userID || stub.upsertFreelancerNoteReq.GetFreelancerUserId() != "f-3" {
			t.Fatalf("expected UpsertFreelancerNote request with trusted user and path freelancer")
		}
	})

	t.Run("GetFreelancerNote", func(t *testing.T) {
		ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/users/me/freelancer-notes/f-3")
		ctx.Set(middleware.ContextUserID, userID)
		ctx.Params = gin.Params{{Key: "freelancerId", Value: "f-3"}}
		h.GetMeFreelancerNote(ctx)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		if stub.getFreelancerNoteReq == nil || stub.getFreelancerNoteReq.GetUserId() != userID || stub.getFreelancerNoteReq.GetFreelancerUserId() != "f-3" {
			t.Fatalf("expected GetFreelancerNote request with trusted user and path freelancer")
		}
	})
}

func TestParsePagination_ValidInput(t *testing.T) {
	ctx, _ := newJSONTestContext(http.MethodGet, "/api/v1/admin/users?page_size=25&page_token=next-25")

	pageSize, pageToken, err := parsePagination(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pageSize != 25 {
		t.Fatalf("expected page size 25, got %d", pageSize)
	}
	if pageToken != "next-25" {
		t.Fatalf("expected page token next-25, got %q", pageToken)
	}
}

func TestParsePagination_Defaults(t *testing.T) {
	ctx, _ := newJSONTestContext(http.MethodGet, "/api/v1/users/me/saved-freelancers")

	pageSize, pageToken, err := parsePagination(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pageSize != 20 {
		t.Fatalf("expected default page size 20, got %d", pageSize)
	}
	if pageToken != "" {
		t.Fatalf("expected empty page token, got %q", pageToken)
	}
}

func TestParsePagination_InvalidInteger(t *testing.T) {
	ctx, _ := newJSONTestContext(http.MethodGet, "/api/v1/admin/users?page_size=nope")

	_, _, err := parsePagination(ctx)
	if err == nil {
		t.Fatalf("expected error for invalid page_size")
	}
}

func TestUpdateMeProfile_Unauthorized(t *testing.T) {
	h := &UserHandler{}
	ctx, rec := newJSONTestContext(http.MethodPatch, "/api/v1/users/me/profile")

	h.UpdateMeProfile(ctx)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestUpdateMeProfile_MixedRolePayloadRejected(t *testing.T) {
	h := &UserHandler{}
	body := `{"company_name":"Acme","headline":"Builder"}`
	ctx, rec := newJSONBodyTestContext(http.MethodPatch, "/api/v1/users/me/profile", body)
	ctx.Set(middleware.ContextUserID, "11111111-1111-1111-1111-111111111111")

	h.UpdateMeProfile(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUpdateMeProfile_UnsupportedLegacyFieldsRejected(t *testing.T) {
	h := &UserHandler{}
	body := `{"first_name":"Jane"}`
	ctx, rec := newJSONBodyTestContext(http.MethodPatch, "/api/v1/users/me/profile", body)
	ctx.Set(middleware.ContextUserID, "11111111-1111-1111-1111-111111111111")

	h.UpdateMeProfile(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUpdateMeProfile_LanguageRejectedInFavorOfSettings(t *testing.T) {
	h := &UserHandler{}
	body := `{"language":"fr"}`
	ctx, rec := newJSONBodyTestContext(http.MethodPatch, "/api/v1/users/me/profile", body)
	ctx.Set(middleware.ContextUserID, "11111111-1111-1111-1111-111111111111")

	h.UpdateMeProfile(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUpdateMeProfile_TooLargeBodyRejected(t *testing.T) {
	h := &UserHandler{}
	largeBody := `{"display_name":"` + strings.Repeat("a", 1<<20) + `"}`
	ctx, rec := newJSONBodyTestContext(http.MethodPatch, "/api/v1/users/me/profile", largeBody)
	ctx.Set(middleware.ContextUserID, "11111111-1111-1111-1111-111111111111")

	h.UpdateMeProfile(ctx)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d", http.StatusRequestEntityTooLarge, rec.Code)
	}
}

type submittedVerificationClientStub struct {
	called bool
}

func (s *submittedVerificationClientStub) GetMyVerificationStatus(ctx context.Context, in *verificationv1.GetMyVerificationStatusRequest, opts ...grpc.CallOption) (*verificationv1.GetMyVerificationStatusResponse, error) {
	s.called = true
	return &verificationv1.GetMyVerificationStatusResponse{
		Request: &verificationv1.VerificationRequest{Status: "submitted"},
	}, nil
}

func TestUpdateMeProfile_TaxIDRejectedWhenVerificationSubmitted(t *testing.T) {
	verificationClient := &submittedVerificationClientStub{}
	h := &UserHandler{verificationClient: verificationClient}
	body := `{"tax_id":"TIN-123"}`
	ctx, rec := newJSONBodyTestContext(http.MethodPatch, "/api/v1/users/me/profile", body)
	ctx.Set(middleware.ContextUserID, "11111111-1111-1111-1111-111111111111")

	h.UpdateMeProfile(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if !verificationClient.called {
		t.Fatalf("expected verification status check to be called")
	}
}

func TestUploadMeAvatar_InvalidContentTypeRejected(t *testing.T) {
	h := &UserHandler{}
	ctx, rec := newJSONBodyTestContext(http.MethodPost, "/api/v1/users/me/avatar", `{"storage_key":"avatars/test/current","file_name":"avatar.txt","content_type":"text/plain"}`)
	ctx.Set(middleware.ContextUserID, "11111111-1111-1111-1111-111111111111")

	h.UploadMeAvatar(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUploadMeAvatar_TooLargeFileRejected(t *testing.T) {
	h := &UserHandler{}
	largeBody := `{"storage_key":"avatars/test/current","file_name":"avatar.png","content_type":"image/png","padding":"` + strings.Repeat("a", 1<<20) + `"}`
	ctx, rec := newJSONBodyTestContext(http.MethodPost, "/api/v1/users/me/avatar", largeBody)
	ctx.Set(middleware.ContextUserID, "11111111-1111-1111-1111-111111111111")

	h.UploadMeAvatar(ctx)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d", http.StatusRequestEntityTooLarge, rec.Code)
	}
}

type updatePortfolioCaptureServer struct {
	userv1.UnimplementedUserServiceServer
	lastReq *userv1.UpdateMyPortfolioItemRequest
}

func (s *updatePortfolioCaptureServer) UpdateMyPortfolioItem(ctx context.Context, in *userv1.UpdateMyPortfolioItemRequest) (*userv1.UpdateMyPortfolioItemResponse, error) {
	s.lastReq = in
	return &userv1.UpdateMyPortfolioItemResponse{
		Item: &userv1.PortfolioItem{
			Id:            in.GetItemId(),
			UserId:        in.GetUserId(),
			Title:         in.GetTitle(),
			Description:   in.GetDescription(),
			ProjectUrl:    in.GetProjectUrl(),
			RoleInProject: in.GetRoleInProject(),
			Tags:          in.GetTags().GetValues(),
		},
	}, nil
}

func TestUpdateMePortfolioItem_UsesPutAndTrustedIdentity(t *testing.T) {
	srv := &updatePortfolioCaptureServer{}
	client, cleanup := newBufConnUserClient(t, srv)
	defer cleanup()

	h := NewUserHandler(client, nil)
	body := `{"user_id":"22222222-2222-2222-2222-222222222222","item_id":999,"title":"Design System Work","description":"Portfolio examples across product and marketing sites.","project_url":"https://example.com/design-system","role_in_project":"Frontend Engineer","completed_at_unix":1738368000,"tags":{"values":["react","ui","design-system"]},"media":{"values":[{"media_type":"PORTFOLIO_MEDIA_TYPE_LINK","external_url":"https://www.behance.net/gallery/one","file_name":"Case Study 1"},{"media_type":"PORTFOLIO_MEDIA_TYPE_LINK","external_url":"https://www.behance.net/gallery/two","file_name":"Case Study 2","content_type":"","size_bytes":0,"width":0,"height":0}]}}`
	ctx, rec := newJSONBodyTestContext(http.MethodPut, "/api/v1/users/me/portfolio/123", body)
	ctx.Set(middleware.ContextUserID, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	ctx.Params = gin.Params{{Key: "itemId", Value: "123"}}

	h.UpdateMePortfolioItem(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if srv.lastReq == nil {
		t.Fatalf("expected grpc request to be captured")
	}
	if got := srv.lastReq.GetUserId(); got != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("expected trusted auth user_id, got %q", got)
	}
	if got := srv.lastReq.GetItemId(); got != 123 {
		t.Fatalf("expected trusted path item_id, got %d", got)
	}
	if got := srv.lastReq.GetTitle(); got != "Design System Work" {
		t.Fatalf("expected title to be preserved, got %q", got)
	}
	if got := srv.lastReq.GetTags().GetValues(); len(got) != 3 || got[0] != "react" || got[1] != "ui" || got[2] != "design-system" {
		t.Fatalf("expected tags values to be preserved, got %#v", got)
	}
	if got := srv.lastReq.GetMedia().GetValues(); len(got) != 2 {
		t.Fatalf("expected 2 media entries, got %d", len(got))
	}
}

func TestGetMeAccountSettings_Unauthorized(t *testing.T) {
	h := &UserHandler{}
	ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/users/me/settings")

	h.GetMeAccountSettings(ctx)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestUpdateMeAccountSettings_RequiresPayload(t *testing.T) {
	h := &UserHandler{}
	body := `{}`
	ctx, rec := newJSONBodyTestContext(http.MethodPatch, "/api/v1/users/me/settings", body)
	ctx.Set(middleware.ContextUserID, "11111111-1111-1111-1111-111111111111")

	h.UpdateMeAccountSettings(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestProtoToAny_ProfileReadiness(t *testing.T) {
	msg := &userv1.ProfileReadiness{
		Percent:               80,
		MissingRequiredFields: []string{"hiring_preferences"},
		Recommendations:       []string{"Add hiring preferences to improve client matching"},
	}

	payload, err := protoToAny(msg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	obj, ok := payload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", payload)
	}
	if got, ok := obj["percent"].(float64); !ok || got != 80 {
		t.Fatalf("expected percent=80, got %v", obj["percent"])
	}

	missing, ok := obj["missing_required_fields"].([]any)
	if !ok || len(missing) != 1 || missing[0] != "hiring_preferences" {
		t.Fatalf("expected missing_required_fields with hiring_preferences, got %#v", obj["missing_required_fields"])
	}

	recs, ok := obj["recommendations"].([]any)
	if !ok || len(recs) != 1 {
		t.Fatalf("expected recommendations array, got %#v", obj["recommendations"])
	}
}

type upsertFreelancerNoteCaptureServer struct {
	userv1.UnimplementedUserServiceServer
	lastReq *userv1.UpsertFreelancerNoteRequest
}

func (s *upsertFreelancerNoteCaptureServer) UpsertFreelancerNote(ctx context.Context, in *userv1.UpsertFreelancerNoteRequest) (*userv1.UpsertFreelancerNoteResponse, error) {
	s.lastReq = in
	return &userv1.UpsertFreelancerNoteResponse{
		Note: &userv1.FreelancerNote{
			FreelancerUserId: in.GetFreelancerUserId(),
			Note:             in.GetNote(),
			UpdatedAtUnix:    1,
		},
	}, nil
}

func newBufConnUserClient(t *testing.T, srv userv1.UserServiceServer) (userv1.UserServiceClient, func()) {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	userv1.RegisterUserServiceServer(grpcServer, srv)

	go func() {
		_ = grpcServer.Serve(listener)
	}()

	ctx := context.Background()
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithInsecure(),
	)
	if err != nil {
		listener.Close()
		grpcServer.Stop()
		t.Fatalf("failed to dial bufconn grpc server: %v", err)
	}

	cleanup := func() {
		_ = conn.Close()
		grpcServer.Stop()
		_ = listener.Close()
		_ = ctx
	}

	return userv1.NewUserServiceClient(conn), cleanup
}

func TestUpsertMeFreelancerNote_IdentityFieldsCannotBeOverriddenByBody(t *testing.T) {
	srv := &upsertFreelancerNoteCaptureServer{}
	client, cleanup := newBufConnUserClient(t, srv)
	defer cleanup()

	h := NewUserHandler(client, nil)
	body := `{"user_id":"22222222-2222-2222-2222-222222222222","freelancer_user_id":"33333333-3333-3333-3333-333333333333","note":"hello"}`
	ctx, rec := newJSONBodyTestContext(http.MethodPut, "/api/v1/users/me/freelancer-notes/11111111-1111-1111-1111-111111111111", body)
	ctx.Set(middleware.ContextUserID, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	ctx.Params = gin.Params{{Key: "freelancerId", Value: "11111111-1111-1111-1111-111111111111"}}

	h.UpsertMeFreelancerNote(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if srv.lastReq == nil {
		t.Fatalf("expected grpc request to be captured")
	}
	if got := srv.lastReq.GetUserId(); got != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("expected trusted auth user_id, got %q", got)
	}
	if got := srv.lastReq.GetFreelancerUserId(); got != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("expected trusted path freelancer_user_id, got %q", got)
	}
	if got := srv.lastReq.GetNote(); got != "hello" {
		t.Fatalf("expected note to be preserved, got %q", got)
	}
}
