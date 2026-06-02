package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	userv1 "jobconnect/user/gen/user"
	verificationv1 "jobconnect/verification/gen/verification/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var protoJSON = protojson.MarshalOptions{
	UseProtoNames:   true,
	UseEnumNumbers:  false,
	EmitUnpopulated: false,
}

var protoJSONWithDefaults = protojson.MarshalOptions{
	UseProtoNames:   true,
	UseEnumNumbers:  false,
	EmitUnpopulated: true,
}

const (
	maxJSONBodyBytes   int64 = 1 << 20
	maxMultipartSlack  int64 = 1 << 20
	maxAvatarFileBytes int64 = 5 * 1024 * 1024
	maxCVFileBytes     int64 = 25 * 1024 * 1024
)

var avatarContentTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
}

var cvContentTypes = map[string]struct{}{
	"application/pdf":    {},
	"application/msword": {},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": {},
}

type UserHandler struct {
	client             userv1.UserServiceClient
	verificationClient verificationStatusClient
}

type verificationStatusClient interface {
	GetMyVerificationStatus(ctx context.Context, in *verificationv1.GetMyVerificationStatusRequest, opts ...grpc.CallOption) (*verificationv1.GetMyVerificationStatusResponse, error)
}

type updateProfileRequest struct {
	DisplayName      *string  `json:"display_name"`
	AvatarURL        *string  `json:"avatar_url"`
	Language         *string  `json:"language"`
	ContactEmail     *string  `json:"contact_email"`
	ContactPhone     *string  `json:"contact_phone"`
	Bio              *string  `json:"bio"`
	FirstName        *string  `json:"first_name"`
	LastName         *string  `json:"last_name"`
	CompanyName      *string  `json:"company_name"`
	TaxID            *string  `json:"tax_id"`
	Headline         *string  `json:"headline"`
	Skills           []string `json:"skills"`
	HourlyRate       *float64 `json:"hourly_rate"`
	Availability     *string  `json:"availability"`
	Location         *string  `json:"location"`
	LastActiveAtUnix *int64   `json:"last_active_at_unix"`
}

type updateAccountSettingsRequest struct {
	UILocale                  *string `json:"ui_locale"`
	EmailNotificationsEnabled *bool   `json:"email_notifications_enabled"`
	PushNotificationsEnabled  *bool   `json:"push_notifications_enabled"`
}

type UserProfileUpdateRequest struct {
	DisplayName      *string  `json:"display_name,omitempty"`
	AvatarURL        *string  `json:"avatar_url,omitempty"`
	Language         *string  `json:"language,omitempty"`
	ContactEmail     *string  `json:"contact_email,omitempty"`
	ContactPhone     *string  `json:"contact_phone,omitempty"`
	Bio              *string  `json:"bio,omitempty"`
	FirstName        *string  `json:"first_name,omitempty"`
	LastName         *string  `json:"last_name,omitempty"`
	CompanyName      *string  `json:"company_name,omitempty"`
	TaxID            *string  `json:"tax_id,omitempty"`
	Headline         *string  `json:"headline,omitempty"`
	Skills           []string `json:"skills,omitempty"`
	HourlyRate       *float64 `json:"hourly_rate,omitempty"`
	Availability     *string  `json:"availability,omitempty"`
	Location         *string  `json:"location,omitempty"`
	LastActiveAtUnix *int64   `json:"last_active_at_unix,omitempty"`
}

type UserAccountSettingsUpdateRequest struct {
	UILocale                  *string `json:"ui_locale,omitempty"`
	EmailNotificationsEnabled *bool   `json:"email_notifications_enabled,omitempty"`
	PushNotificationsEnabled  *bool   `json:"push_notifications_enabled,omitempty"`
}

type UserAvatarUploadURLRequest struct {
	FileName    string `json:"file_name,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

type UserAvatarFinalizeRequest struct {
	StorageKey  string `json:"storage_key"`
	FileName    string `json:"file_name,omitempty"`
	ContentType string `json:"content_type"`
	Width       *int32 `json:"width,omitempty"`
	Height      *int32 `json:"height,omitempty"`
}

type UserCVUploadURLRequest struct {
	FileName    string `json:"file_name,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

type UserCVFinalizeRequest struct {
	StorageKey  string `json:"storage_key"`
	FileName    string `json:"file_name,omitempty"`
	ContentType string `json:"content_type"`
}

type UserPortfolioMediaUploadURLRequest struct {
	FileName    string `json:"file_name,omitempty"`
	ContentType string `json:"content_type,omitempty"`
}

type PortfolioMediaInputDTO struct {
	StorageKey  string `json:"storage_key,omitempty"`
	ContentType string `json:"content_type,omitempty"`
	FileName    string `json:"file_name,omitempty"`
	MediaType   string `json:"media_type,omitempty"`
}

type UserPortfolioCreateRequest struct {
	Title           string                   `json:"title,omitempty"`
	Description     string                   `json:"description,omitempty"`
	ProjectURL      string                   `json:"project_url,omitempty"`
	RoleInProject   string                   `json:"role_in_project,omitempty"`
	CompletedAtUnix *int64                   `json:"completed_at_unix,omitempty"`
	Tags            []string                 `json:"tags,omitempty"`
	Media           []PortfolioMediaInputDTO `json:"media,omitempty"`
}

type UserPortfolioUpdateRequest struct {
	Title           *string                  `json:"title,omitempty"`
	Description     *string                  `json:"description,omitempty"`
	ProjectURL      *string                  `json:"project_url,omitempty"`
	RoleInProject   *string                  `json:"role_in_project,omitempty"`
	CompletedAtUnix *int64                   `json:"completed_at_unix,omitempty"`
	Tags            []string                 `json:"tags,omitempty"`
	Media           []PortfolioMediaInputDTO `json:"media,omitempty"`
}

type UserWorkPreferencesRequest struct {
	PreferredProjectLength *string  `json:"preferred_project_length,omitempty"`
	MinBudget              *float64 `json:"min_budget,omitempty"`
	MaxBudget              *float64 `json:"max_budget,omitempty"`
	ContractTypes          []string `json:"contract_types,omitempty"`
	WeeklyCapacityHours    *int64   `json:"weekly_capacity_hours,omitempty"`
}

type UserHiringPreferencesRequest struct {
	MinHourlyRate      *float64 `json:"min_hourly_rate,omitempty"`
	MaxHourlyRate      *float64 `json:"max_hourly_rate,omitempty"`
	PreferredLocations []string `json:"preferred_locations,omitempty"`
}

type UserFreelancerNoteRequest struct {
	Note string `json:"note,omitempty"`
}

type UserErrorResponse struct {
	Error string `json:"error"`
}

type UserProfileResponse struct {
	Profile      any `json:"profile"`
	Completeness any `json:"completeness"`
}

type UserOnboardingStatusResponse struct {
	Completeness any   `json:"completeness"`
	Readiness    any   `json:"readiness"`
	Steps        []any `json:"steps"`
}

type UserSettingsResponse struct {
	Settings any `json:"settings"`
}

type UploadURLResponse struct {
	StorageKey string `json:"storage_key"`
	UploadURL  string `json:"upload_url"`
}

type AvatarResponse struct {
	Avatar    any    `json:"avatar"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type UserDeletedResponse struct {
	Deleted bool `json:"deleted"`
}

type RemovedResponse struct {
	Removed bool `json:"removed"`
}

type CVResponse struct {
	Cv any `json:"cv"`
}

type PortfolioItemResponse struct {
	Item any `json:"item"`
}

type PortfolioItemsResponse struct {
	Items         []any  `json:"items"`
	NextPageToken string `json:"next_page_token,omitempty"`
}

type WorkPreferencesResponse struct {
	Settings any `json:"settings"`
}

type HiringPreferencesResponse struct {
	Preferences any `json:"preferences"`
}

type SavedFreelancerResponse struct {
	Saved any `json:"saved"`
}

type SavedFreelancersResponse struct {
	Freelancers   []any  `json:"freelancers"`
	NextPageToken string `json:"next_page_token,omitempty"`
}

type FreelancerNoteResponse struct {
	Note any `json:"note"`
}

func NewUserHandler(client userv1.UserServiceClient, verificationClient verificationStatusClient) *UserHandler {
	return &UserHandler{client: client, verificationClient: verificationClient}
}

// GetMe godoc
// @Summary Get current user profile
// @Description Returns the authenticated user's profile and completeness summary.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserProfileResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/profile [get]
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetMyProfile(c.Request.Context(), &userv1.GetMyProfileRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	profilePayload, err := protoToAny(resp.GetProfile())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	completenessPayload, err := protoToAny(resp.GetCompleteness())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"profile": profilePayload, "completeness": completenessPayload})
}

// UpdateMeProfile godoc
// @Summary Update my profile
// @Description Patch profile fields for the authenticated user.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UserProfileUpdateRequest true "Profile patch payload"
// @Success 200 {object} UserProfileResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/profile [patch]
func (h *UserHandler) UpdateMeProfile(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	limitRequestBody(c, maxJSONBodyBytes)
	var body updateProfileRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		if requestBodyTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.AvatarURL != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "avatar_url must be managed via avatar endpoints"})
		return
	}
	if body.Language != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "language must be updated via /users/me/settings"})
		return
	}
	if body.FirstName != nil || body.LastName != nil || body.LastActiveAtUnix != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported profile fields in this endpoint"})
		return
	}
	if body.TaxID != nil {
		locked, err := h.taxIDLocked(c.Request.Context(), userID)
		if err != nil {
			writeGRPCError(c, err)
			return
		}
		if locked {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tax_id update is not allowed after verification submission"})
			return
		}
	}

	var availability *userv1.Availability
	if body.Availability != nil {
		parsed, err := parseAvailability(*body.Availability)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		availability = &parsed
	}

	hasCore := body.DisplayName != nil || body.ContactEmail != nil || body.ContactPhone != nil || body.Bio != nil || body.TaxID != nil || body.Location != nil
	hasClient := body.CompanyName != nil
	hasFreelancer := body.Headline != nil || body.HourlyRate != nil || availability != nil || body.Skills != nil

	if !(hasCore || hasClient || hasFreelancer) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one updatable field is required"})
		return
	}
	if hasClient && hasFreelancer {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client and freelancer fields cannot be patched together"})
		return
	}

	req := &userv1.PatchMyProfileRequest{UserId: userID}
	if hasCore {
		req.Core = &userv1.PatchMyProfileCoreInput{
			DisplayName:  body.DisplayName,
			ContactEmail: body.ContactEmail,
			ContactPhone: body.ContactPhone,
			Bio:          body.Bio,
			TaxId:        body.TaxID,
			Location:     body.Location,
		}
	}
	if hasClient {
		req.RoleProfile = &userv1.PatchMyProfileRequest_Client{
			Client: &userv1.PatchMyClientProfileInput{CompanyName: body.CompanyName},
		}
	}
	if hasFreelancer {
		freelancer := &userv1.PatchMyFreelancerProfileInput{
			Headline:     body.Headline,
			HourlyRate:   body.HourlyRate,
			Availability: availability,
		}
		if body.Skills != nil {
			freelancer.Skills = &userv1.StringList{Values: body.Skills}
		}
		req.RoleProfile = &userv1.PatchMyProfileRequest_Freelancer{Freelancer: freelancer}
	}

	resp, err := h.client.PatchMyProfile(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	profilePayload, err := protoToAny(resp.GetProfile())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	completenessPayload, err := protoToAny(resp.GetCompleteness())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"profile": profilePayload, "completeness": completenessPayload})
}

func (h *UserHandler) taxIDLocked(ctx context.Context, userID string) (bool, error) {
	if h.verificationClient == nil {
		return false, fmt.Errorf("verification client unavailable")
	}
	resp, err := h.verificationClient.GetMyVerificationStatus(ctx, &verificationv1.GetMyVerificationStatusRequest{UserId: userID})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return false, nil
		}
		return false, err
	}
	request := resp.GetRequest()
	if request == nil {
		return false, nil
	}
	switch strings.TrimSpace(strings.ToLower(request.GetStatus())) {
	case "submitted", "pending_review", "verified":
		return true, nil
	default:
		return false, nil
	}
}

// DeleteMeProfile godoc
// @Summary Delete my profile
// @Description Deletes the authenticated user's profile.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Param hard_delete query boolean false "Hard delete flag"
// @Success 200 {object} UserDeletedResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/profile [delete]
func (h *UserHandler) DeleteMeProfile(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	hardDelete := false
	if v := strings.TrimSpace(c.Query("hard_delete")); v != "" {
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "hard_delete must be a boolean"})
			return
		}
		hardDelete = parsed
	}

	resp, err := h.client.DeleteMyProfile(c.Request.Context(), &userv1.DeleteMyProfileRequest{UserId: userID, HardDelete: hardDelete})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": resp.GetDeleted()})
}

// GetMeOnboardingStatus godoc
// @Summary Get my onboarding status
// @Description Returns onboarding completeness, readiness and steps for the authenticated user.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserOnboardingStatusResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/onboarding-status [get]
func (h *UserHandler) GetMeOnboardingStatus(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetMyOnboardingStatus(c.Request.Context(), &userv1.GetMyOnboardingStatusRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	completenessPayload, err := protoToAny(resp.GetCompleteness())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	stepsPayload, err := protoSliceToAny(resp.GetSteps())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	readinessPayload, err := protoToAny(resp.GetReadiness())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"completeness": completenessPayload, "readiness": readinessPayload, "steps": stepsPayload})
}

// GetMeAccountSettings godoc
// @Summary Get my account settings
// @Description Returns account settings for the authenticated user.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserSettingsResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/settings [get]
func (h *UserHandler) GetMeAccountSettings(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetMySettings(c.Request.Context(), &userv1.GetMySettingsRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	settingsPayload, err := protoToAnyWithDefaults(resp.GetSettings())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settingsPayload})
}

// UpdateMeAccountSettings godoc
// @Summary Update my account settings
// @Description Patch account and notification settings for the authenticated user.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UserAccountSettingsUpdateRequest true "Settings payload"
// @Success 200 {object} UserSettingsResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/settings [patch]
func (h *UserHandler) UpdateMeAccountSettings(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	limitRequestBody(c, maxJSONBodyBytes)
	var body updateAccountSettingsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		if requestBodyTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.UILocale == nil && body.EmailNotificationsEnabled == nil && body.PushNotificationsEnabled == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one updatable setting is required"})
		return
	}

	resp, err := h.client.PatchMySettings(c.Request.Context(), &userv1.PatchMySettingsRequest{
		UserId:                    userID,
		UiLocale:                  body.UILocale,
		EmailNotificationsEnabled: body.EmailNotificationsEnabled,
		PushNotificationsEnabled:  body.PushNotificationsEnabled,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	settingsPayload, err := protoToAnyWithDefaults(resp.GetSettings())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settingsPayload})
}

// GetMeAvatarUploadUrl godoc
// @Summary Reserve avatar upload URL
// @Description Returns a signed upload URL and storage key for avatar upload.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UserAvatarUploadURLRequest true "Avatar upload URL payload"
// @Success 200 {object} UploadURLResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/avatar/upload-url [post]
func (h *UserHandler) GetMeAvatarUploadUrl(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	req := &userv1.GetMyAvatarUploadUrlRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}
	req.UserId = userID
	resp, err := h.client.GetMyAvatarUploadUrl(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"storage_key": resp.GetStorageKey(), "upload_url": resp.GetUploadUrl()})
	return
}

// UploadMeAvatar godoc
// @Summary Upload avatar metadata
// @Description Registers an uploaded avatar for the authenticated user.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UserAvatarFinalizeRequest true "Avatar finalize payload"
// @Success 200 {object} AvatarResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/avatar [post]
func (h *UserHandler) UploadMeAvatar(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.UploadMyAvatarRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}
	req.UserId = userID

	if strings.TrimSpace(req.GetStorageKey()) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "storage_key is required"})
		return
	}
	if !allowedContentType(req.GetContentType(), avatarContentTypes) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported avatar content_type"})
		return
	}

	resp, err := h.client.UpsertMyAvatar(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	avatarPayload, err := protoToAny(resp.GetAvatar())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"avatar_url": resp.GetAvatarUrl(), "avatar": avatarPayload})
}

// GetMeAvatar godoc
// @Summary Get my avatar
// @Description Returns the authenticated user's avatar resource.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} AvatarResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/avatar [get]
func (h *UserHandler) GetMeAvatar(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetMyAvatar(c.Request.Context(), &userv1.GetMyAvatarRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	avatarPayload, err := protoToAny(resp.GetAvatar())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"avatar": avatarPayload})
}

// RemoveMeAvatar godoc
// @Summary Remove my avatar
// @Description Removes the authenticated user's avatar.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} RemovedResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/avatar [delete]
func (h *UserHandler) RemoveMeAvatar(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.RemoveMyAvatar(c.Request.Context(), &userv1.RemoveMyAvatarRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"removed": resp.GetRemoved()})
}

// GetMeCVUploadUrl godoc
// @Summary Reserve CV upload URL
// @Description Returns a signed upload URL and storage key for CV upload.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UserCVUploadURLRequest true "CV upload URL payload"
// @Success 200 {object} UploadURLResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/cv/upload-url [post]
func (h *UserHandler) GetMeCVUploadUrl(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	req := &userv1.GetMyCVUploadUrlRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}
	req.UserId = userID
	resp, err := h.client.GetMyCVUploadUrl(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"storage_key": resp.GetStorageKey(), "upload_url": resp.GetUploadUrl()})
	return
}

// UploadMeCV godoc
// @Summary Upload CV metadata
// @Description Registers an uploaded CV for the authenticated user.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UserCVFinalizeRequest true "CV finalize payload"
// @Success 200 {object} CVResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/cv [post]
func (h *UserHandler) UploadMeCV(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.UploadMyCVRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}
	req.UserId = userID

	if strings.TrimSpace(req.GetStorageKey()) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "storage_key is required"})
		return
	}
	if !allowedContentType(req.GetContentType(), cvContentTypes) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported cv content_type"})
		return
	}

	resp, err := h.client.UpsertMyCV(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	cvPayload, err := protoToAny(resp.GetCv())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cv": cvPayload})
}

// GetMeCV godoc
// @Summary Get my CV
// @Description Returns the authenticated user's CV resource.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} CVResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/cv [get]
func (h *UserHandler) GetMeCV(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetMyCV(c.Request.Context(), &userv1.GetMyCVRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	cvPayload, err := protoToAny(resp.GetCv())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cv": cvPayload})
}

// RemoveMeCV godoc
// @Summary Remove my CV
// @Description Removes the authenticated user's CV.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} RemovedResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/cv [delete]
func (h *UserHandler) RemoveMeCV(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.RemoveMyCV(c.Request.Context(), &userv1.RemoveMyCVRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"removed": resp.GetRemoved()})
}

// GetMePortfolioMediaUploadUrl godoc
// @Summary Reserve portfolio media upload URL
// @Description Returns a signed upload URL and storage key for portfolio media upload.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UserPortfolioMediaUploadURLRequest true "Portfolio media upload URL payload"
// @Success 200 {object} UploadURLResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/portfolio/media/upload-url [post]
func (h *UserHandler) GetMePortfolioMediaUploadUrl(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	req := &userv1.GetMyPortfolioMediaUploadUrlRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}
	req.UserId = userID

	resp, err := h.client.GetMyPortfolioMediaUploadUrl(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"storage_key": resp.GetStorageKey(), "upload_url": resp.GetUploadUrl()})
}

// CreateMePortfolioItem godoc
// @Summary Create portfolio item
// @Description Creates a new portfolio item for the authenticated user.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UserPortfolioCreateRequest true "Portfolio item payload"
// @Success 200 {object} PortfolioItemResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/portfolio [post]
func (h *UserHandler) CreateMePortfolioItem(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.CreateMyPortfolioItemRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}
	// Ensure authenticated caller's user ID is used (don't allow override from request body).
	req.UserId = userID

	resp, err := h.client.CreateMyPortfolioItem(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "item", resp.GetItem())
}

// ListMePortfolioItems godoc
// @Summary List my portfolio items
// @Description Returns paginated portfolio items for the authenticated user.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Param page_size query int false "Page size"
// @Param page_token query string false "Page token"
// @Success 200 {object} PortfolioItemsResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/portfolio [get]
func (h *UserHandler) ListMePortfolioItems(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	pageSize, pageToken, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.ListMyPortfolioItems(c.Request.Context(), &userv1.ListMyPortfolioItemsRequest{UserId: userID, Page: &userv1.PagingRequest{PageSize: pageSize, PageToken: pageToken}})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	itemsPayload, err := protoSliceToAny(resp.GetItems())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	nextPageToken := ""
	if resp.GetPage() != nil {
		nextPageToken = resp.GetPage().GetNextPageToken()
	}
	c.JSON(http.StatusOK, gin.H{"items": itemsPayload, "next_page_token": nextPageToken})
}

// GetMePortfolioItem godoc
// @Summary Get portfolio item
// @Description Returns a specific portfolio item for the authenticated user.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Param itemId path int true "Portfolio item ID"
// @Success 200 {object} PortfolioItemResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/portfolio/{itemId} [get]
func (h *UserHandler) GetMePortfolioItem(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	itemID, err := parseInt64PathParam(c, "itemId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.GetMyPortfolioItem(c.Request.Context(), &userv1.GetMyPortfolioItemRequest{UserId: userID, ItemId: itemID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "item", resp.GetItem())
}

// UpdateMePortfolioItem godoc
// @Summary Update portfolio item
// @Description Updates a portfolio item owned by the authenticated user.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param itemId path int true "Portfolio item ID"
// @Param request body UserPortfolioUpdateRequest true "Portfolio update payload"
// @Success 200 {object} PortfolioItemResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/portfolio/{itemId} [put]
func (h *UserHandler) UpdateMePortfolioItem(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	itemID, err := parseInt64PathParam(c, "itemId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req := &userv1.UpdateMyPortfolioItemRequest{UserId: userID, ItemId: itemID}
	if !bindProtoJSON(c, req) {
		return
	}
	// Ensure authenticated caller's user ID and path item ID are preserved.
	req.UserId = userID
	req.ItemId = itemID

	resp, err := h.client.UpdateMyPortfolioItem(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "item", resp.GetItem())
}

// DeleteMePortfolioItem godoc
// @Summary Delete portfolio item
// @Description Deletes a portfolio item owned by the authenticated user.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Param itemId path int true "Portfolio item ID"
// @Success 200 {object} UserDeletedResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/portfolio/{itemId} [delete]
func (h *UserHandler) DeleteMePortfolioItem(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	itemID, err := parseInt64PathParam(c, "itemId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.DeleteMyPortfolioItem(c.Request.Context(), &userv1.DeleteMyPortfolioItemRequest{UserId: userID, ItemId: itemID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": resp.GetDeleted()})
}

// SetMeWorkPreferences godoc
// @Summary Set work preferences
// @Description Updates the authenticated user's work preferences.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UserWorkPreferencesRequest true "Work preferences payload"
// @Success 200 {object} WorkPreferencesResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/work-preferences [patch]
func (h *UserHandler) SetMeWorkPreferences(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.PatchMyWorkPreferencesRequest{UserId: userID}
	limitRequestBody(c, maxJSONBodyBytes)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		if requestBodyTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read request body"})
		return
	}
	body, err = normalizeWorkPreferencesJSON(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(strings.TrimSpace(string(body))) > 0 {
		if err := protojson.Unmarshal(body, req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	req.UserId = userID

	resp, err := h.client.PatchMyWorkPreferences(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "settings", resp.GetSettings())
}

// GetMeWorkPreferences godoc
// @Summary Get work preferences
// @Description Returns the authenticated user's work preferences.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} WorkPreferencesResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/work-preferences [get]
func (h *UserHandler) GetMeWorkPreferences(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetMyWorkPreferences(c.Request.Context(), &userv1.GetMyWorkPreferencesRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "settings", resp.GetSettings())
}

// GetMeHiringPreferences godoc
// @Summary Get hiring preferences
// @Description Returns hiring preferences for the authenticated user.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} HiringPreferencesResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/hiring-preferences [get]
func (h *UserHandler) GetMeHiringPreferences(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetMyHiringPreferences(c.Request.Context(), &userv1.GetMyHiringPreferencesRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "preferences", resp.GetPreferences())
}

// UpdateMeHiringPreferences godoc
// @Summary Update hiring preferences
// @Description Updates hiring preferences for the authenticated user.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UserHiringPreferencesRequest true "Hiring preferences payload"
// @Success 200 {object} HiringPreferencesResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/hiring-preferences [patch]
func (h *UserHandler) UpdateMeHiringPreferences(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.PatchMyHiringPreferencesRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}
	req.UserId = userID

	resp, err := h.client.PatchMyHiringPreferences(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "preferences", resp.GetPreferences())
}

// SaveMeFreelancer godoc
// @Summary Save a freelancer
// @Description Shortlist a freelancer for the authenticated client.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Param freelancerId path string true "Freelancer user ID"
// @Success 200 {object} SavedFreelancerResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/saved-freelancers/{freelancerId} [post]
func (h *UserHandler) SaveMeFreelancer(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	freelancerID := strings.TrimSpace(c.Param("freelancerId"))
	if freelancerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "freelancerId is required"})
		return
	}

	resp, err := h.client.SaveFreelancer(c.Request.Context(), &userv1.SaveFreelancerRequest{UserId: userID, FreelancerUserId: freelancerID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "saved", resp.GetSaved())
}

// ListMeSavedFreelancers godoc
// @Summary List saved freelancers
// @Description Returns paginated saved freelancers for the authenticated user.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Param page_size query int false "Page size"
// @Param page_token query string false "Page token"
// @Success 200 {object} SavedFreelancersResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/saved-freelancers [get]
func (h *UserHandler) ListMeSavedFreelancers(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	pageSize, pageToken, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.ListSavedFreelancers(c.Request.Context(), &userv1.ListSavedFreelancersRequest{UserId: userID, Page: &userv1.PagingRequest{PageSize: pageSize, PageToken: pageToken}})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	itemsPayload, err := protoSliceToAny(resp.GetFreelancers())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	nextPageToken := ""
	if resp.GetPage() != nil {
		nextPageToken = resp.GetPage().GetNextPageToken()
	}
	c.JSON(http.StatusOK, gin.H{"freelancers": itemsPayload, "next_page_token": nextPageToken})
}

// RemoveMeSavedFreelancer godoc
// @Summary Remove saved freelancer
// @Description Removes a freelancer from the authenticated user's saved list.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Param freelancerId path string true "Freelancer user ID"
// @Success 200 {object} RemovedResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/saved-freelancers/{freelancerId} [delete]
func (h *UserHandler) RemoveMeSavedFreelancer(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	freelancerID := strings.TrimSpace(c.Param("freelancerId"))
	if freelancerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "freelancerId is required"})
		return
	}

	resp, err := h.client.RemoveSavedFreelancer(c.Request.Context(), &userv1.RemoveSavedFreelancerRequest{UserId: userID, FreelancerUserId: freelancerID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"removed": resp.GetRemoved()})
}

// UpsertMeFreelancerNote godoc
// @Summary Upsert freelancer note
// @Description Creates or updates a recruiter note about a freelancer.
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param freelancerId path string true "Freelancer user ID"
// @Param request body UserFreelancerNoteRequest true "Freelancer note payload"
// @Success 200 {object} FreelancerNoteResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/freelancer-notes/{freelancerId} [put]
func (h *UserHandler) UpsertMeFreelancerNote(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	freelancerID := strings.TrimSpace(c.Param("freelancerId"))
	if freelancerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "freelancerId is required"})
		return
	}

	req := &userv1.UpsertFreelancerNoteRequest{UserId: userID, FreelancerUserId: freelancerID}
	if !bindProtoJSON(c, req) {
		return
	}
	req.UserId = userID
	req.FreelancerUserId = freelancerID

	resp, err := h.client.UpsertFreelancerNote(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "note", resp.GetNote())
}

// GetMeFreelancerNote godoc
// @Summary Get freelancer note
// @Description Returns the recruiter's note for a specific freelancer.
// @Tags User
// @Produce json
// @Security BearerAuth
// @Param freelancerId path string true "Freelancer user ID"
// @Success 200 {object} FreelancerNoteResponse
// @Failure 400 {object} UserErrorResponse
// @Failure 401 {object} UserErrorResponse
// @Failure 500 {object} UserErrorResponse
// @Router /api/v1/users/me/freelancer-notes/{freelancerId} [get]
func (h *UserHandler) GetMeFreelancerNote(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	freelancerID := strings.TrimSpace(c.Param("freelancerId"))
	if freelancerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "freelancerId is required"})
		return
	}

	resp, err := h.client.GetFreelancerNote(c.Request.Context(), &userv1.GetFreelancerNoteRequest{UserId: userID, FreelancerUserId: freelancerID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "note", resp.GetNote())
}

func parseInt64PathParam(c *gin.Context, param string) (int64, error) {
	raw := strings.TrimSpace(c.Param(param))
	if raw == "" {
		return 0, fmt.Errorf("%s is required", param)
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", param)
	}
	return value, nil
}

func parsePagination(c *gin.Context) (uint32, string, error) {
	pageSize := uint32(20)
	if raw := strings.TrimSpace(c.Query("page_size")); raw != "" {
		parsed, err := strconv.ParseUint(raw, 10, 32)
		if err != nil {
			return 0, "", fmt.Errorf("page_size must be an integer")
		}
		if parsed == 0 || parsed > 100 {
			return 0, "", fmt.Errorf("page_size must be between 1 and 100")
		}
		pageSize = uint32(parsed)
	}
	return pageSize, strings.TrimSpace(c.Query("page_token")), nil
}

func bindProtoJSON(c *gin.Context, msg proto.Message) bool {
	limitRequestBody(c, maxJSONBodyBytes)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		if requestBodyTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
			return false
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read request body"})
		return false
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return true
	}
	if err := protojson.Unmarshal(body, msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return false
	}
	return true
}

func limitRequestBody(c *gin.Context, maxBytes int64) {
	if c == nil || c.Request == nil || maxBytes <= 0 {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
}

func requestBodyTooLarge(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "request body too large")
}

func allowedContentType(contentType string, allowed map[string]struct{}) bool {
	normalized := strings.ToLower(strings.TrimSpace(contentType))
	_, ok := allowed[normalized]
	return ok
}

func multipartUploadLimit(fileBytes int64) int64 {
	if fileBytes <= 0 {
		return maxMultipartSlack
	}
	return fileBytes + maxMultipartSlack
}

func writeProtoEnvelope(c *gin.Context, statusCode int, key string, msg proto.Message) {
	payload, err := protoToAny(msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(statusCode, gin.H{key: payload})
}

func protoToAny(msg proto.Message) (any, error) {
	return protoToAnyWithOptions(msg, protoJSON)
}

func protoToAnyWithDefaults(msg proto.Message) (any, error) {
	return protoToAnyWithOptions(msg, protoJSONWithDefaults)
}

func protoToAnyWithOptions(msg proto.Message, marshaler protojson.MarshalOptions) (any, error) {
	if msg == nil {
		return nil, nil
	}
	raw, err := marshaler.Marshal(msg)
	if err != nil {
		return nil, err
	}
	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func protoSliceToAny[T proto.Message](messages []T) ([]any, error) {
	out := make([]any, 0, len(messages))
	for _, msg := range messages {
		payload, err := protoToAny(msg)
		if err != nil {
			return nil, err
		}
		out = append(out, payload)
	}
	return out, nil
}

func parseAvailability(raw string) (userv1.Availability, error) {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	normalized = strings.TrimPrefix(normalized, "AVAILABILITY_")
	switch normalized {
	case "FULL_TIME":
		return userv1.Availability_AVAILABILITY_FULL_TIME, nil
	case "PART_TIME":
		return userv1.Availability_AVAILABILITY_PART_TIME, nil
	case "AS_NEEDED":
		return userv1.Availability_AVAILABILITY_AS_NEEDED, nil
	case "UNAVAILABLE":
		return userv1.Availability_AVAILABILITY_UNAVAILABLE, nil
	default:
		return userv1.Availability_AVAILABILITY_UNSPECIFIED, fmt.Errorf("invalid availability")
	}
}

func parseProjectLength(raw string) (userv1.ProjectLength, error) {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	normalized = strings.TrimPrefix(normalized, "PROJECT_LENGTH_")
	switch normalized {
	case "", "UNSPECIFIED", "NO_PREFERENCE", "NONE":
		return userv1.ProjectLength_PROJECT_LENGTH_UNSPECIFIED, nil
	case "SHORT", "SHORT_TERM", "SHORT-TERM":
		return userv1.ProjectLength_PROJECT_LENGTH_SHORT_TERM, nil
	case "MEDIUM", "MEDIUM_TERM", "MEDIUM-TERM":
		return userv1.ProjectLength_PROJECT_LENGTH_MEDIUM_TERM, nil
	case "LONG", "LONG_TERM", "LONG-TERM":
		return userv1.ProjectLength_PROJECT_LENGTH_LONG_TERM, nil
	default:
		return userv1.ProjectLength_PROJECT_LENGTH_UNSPECIFIED, fmt.Errorf("invalid preferred_project_length")
	}
}

func normalizeWorkPreferencesJSON(body []byte) ([]byte, error) {
	if len(strings.TrimSpace(string(body))) == 0 {
		return body, nil
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	rawValue, ok := payload["preferred_project_length"]
	if !ok {
		return body, nil
	}

	var asString string
	if err := json.Unmarshal(rawValue, &asString); err == nil {
		parsed, parseErr := parseProjectLength(asString)
		if parseErr != nil {
			return nil, parseErr
		}
		normalized, _ := json.Marshal(parsed.String())
		payload["preferred_project_length"] = normalized
		return json.Marshal(payload)
	}

	return body, nil
}

func parseUserRole(raw string) (userv1.UserRole, error) {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	normalized = strings.TrimPrefix(normalized, "USER_ROLE_")
	switch normalized {
	case "CLIENT":
		return userv1.UserRole_USER_ROLE_CLIENT, nil
	case "FREELANCER":
		return userv1.UserRole_USER_ROLE_FREELANCER, nil
	case "ADMIN":
		return userv1.UserRole_USER_ROLE_ADMIN, nil
	default:
		return userv1.UserRole_USER_ROLE_UNSPECIFIED, fmt.Errorf("invalid role")
	}
}

func parseAccountStatus(raw string) (userv1.AccountStatus, error) {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	normalized = strings.TrimPrefix(normalized, "ACCOUNT_STATUS_")
	switch normalized {
	case "ACTIVE":
		return userv1.AccountStatus_ACCOUNT_STATUS_ACTIVE, nil
	case "SUSPENDED":
		return userv1.AccountStatus_ACCOUNT_STATUS_SUSPENDED, nil
	case "DELETED":
		return userv1.AccountStatus_ACCOUNT_STATUS_DELETED, nil
	default:
		return userv1.AccountStatus_ACCOUNT_STATUS_UNSPECIFIED, fmt.Errorf("invalid status")
	}
}
