package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	userv1 "jobconnect/user/gen/user"
	verificationv1 "jobconnect/verification/gen/verification/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var protoJSON = protojson.MarshalOptions{
	UseProtoNames:   true,
	UseEnumNumbers:  false,
	EmitUnpopulated: false,
}

type UserHandler struct {
	client             userv1.UserServiceClient
	verificationClient verificationv1.VerificationServiceClient
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
	BillingAddress   *string  `json:"billing_address"`
	TaxID            *string  `json:"tax_id"`
	Headline         *string  `json:"headline"`
	Skills           []string `json:"skills"`
	ExperienceLevel  *string  `json:"experience_level"`
	HourlyRate       *float64 `json:"hourly_rate"`
	Availability     *string  `json:"availability"`
	Location         *string  `json:"location"`
	LastActiveAtUnix *int64   `json:"last_active_at_unix"`
}

type updateAccountStatusRequest struct {
	Status           string  `json:"status" binding:"required"`
	Visibility       string  `json:"visibility" binding:"required"`
	SuspensionReason *string `json:"suspension_reason"`
}

type updateAccountSettingsRequest struct {
	UILocale *string `json:"ui_locale"`
}

func NewUserHandler(client userv1.UserServiceClient, verificationClient verificationv1.VerificationServiceClient) *UserHandler {
	return &UserHandler{client: client, verificationClient: verificationClient}
}

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

	verificationPayload := any(nil)
	if h.verificationClient != nil {
		verificationResp, err := h.verificationClient.GetMyVerificationStatus(c.Request.Context(), &verificationv1.GetMyVerificationStatusRequest{UserId: userID})
		if err != nil {
			if st, ok := grpcstatus.FromError(err); !ok || st.Code() != codes.NotFound {
				writeGRPCError(c, err)
				return
			}
		}
		if verificationResp != nil && verificationResp.GetRequest() != nil {
			verificationPayload, err = protoToAny(verificationResp.GetRequest())
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"profile": profilePayload, "completeness": completenessPayload, "verification": verificationPayload})
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	requesterID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userID := strings.TrimSpace(c.Param("userId"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	resp, err := h.client.GetUserProfile(c.Request.Context(), &userv1.GetUserProfileRequest{RequesterUserId: requesterID, TargetUserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "profile", resp.GetProfile())
}

func (h *UserHandler) GetInternalUserBasic(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("userId"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	resp, err := h.client.GetInternalUserBasic(c.Request.Context(), &userv1.GetInternalUserBasicRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "user", resp.GetUser())
}

func (h *UserHandler) GetInternalUserProfile(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("userId"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	resp, err := h.client.GetInternalUserProfile(c.Request.Context(), &userv1.GetInternalUserProfileRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "profile", resp.GetProfile())
}

func (h *UserHandler) GetPublicProfile(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("userId"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	resp, err := h.client.GetPublicProfile(c.Request.Context(), &userv1.GetPublicProfileRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "profile", resp.GetProfile())
}

func (h *UserHandler) UpdateMeProfile(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var body updateProfileRequest
	if err := c.ShouldBindJSON(&body); err != nil {
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
	if body.FirstName != nil || body.LastName != nil || body.BillingAddress != nil || body.TaxID != nil || body.ExperienceLevel != nil || body.LastActiveAtUnix != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported profile fields in this endpoint"})
		return
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

	hasCore := body.DisplayName != nil || body.ContactEmail != nil || body.ContactPhone != nil || body.Bio != nil
	hasClient := body.CompanyName != nil
	hasFreelancer := body.Headline != nil || body.HourlyRate != nil || availability != nil || body.Location != nil || body.Skills != nil

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
			Location:     body.Location,
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
	c.JSON(http.StatusOK, gin.H{"completeness": completenessPayload, "steps": stepsPayload})
}

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

	settingsPayload, err := protoToAny(resp.GetSettings())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settingsPayload})
}

func (h *UserHandler) UpdateMeAccountSettings(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var body updateAccountSettingsRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.UILocale == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one updatable setting is required"})
		return
	}

	resp, err := h.client.PatchMySettings(c.Request.Context(), &userv1.PatchMySettingsRequest{
		UserId:   userID,
		UiLocale: body.UILocale,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	settingsPayload, err := protoToAny(resp.GetSettings())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settingsPayload})
}

func (h *UserHandler) GetMePrivacySettings(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "endpoint not implemented in current gateway phase"})
}

func (h *UserHandler) UpdateMePrivacySettings(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "endpoint not implemented in current gateway phase"})
}

func (h *UserHandler) GetMeNotificationSettings(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "endpoint not implemented in current gateway phase"})
}

func (h *UserHandler) UpdateMeNotificationSettings(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "endpoint not implemented in current gateway phase"})
}

func (h *UserHandler) UploadMeAvatar(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to open uploaded file"})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read uploaded file"})
		return
	}

	contentType := strings.TrimSpace(fileHeader.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}

	resp, err := h.client.UpsertMyAvatar(c.Request.Context(), &userv1.UploadMyAvatarRequest{
		UserId:      userID,
		FileName:    fileHeader.Filename,
		ContentType: contentType,
		Content:     content,
	})
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

	avatar := resp.GetAvatar()
	if fileName := strings.TrimSpace(avatar.GetFileName()); fileName != "" {
		c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", fileName))
	}
	contentType := strings.TrimSpace(avatar.GetContentType())
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Data(http.StatusOK, contentType, resp.GetContent())
}

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

func (h *UserHandler) UpdateAccountStatus(c *gin.Context) {
	requesterID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userID := strings.TrimSpace(c.Param("userId"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	var body updateAccountStatusRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status, err := parseAccountStatus(body.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	statusValue := status

	resp, err := h.client.PatchUserGovernance(c.Request.Context(), &userv1.PatchUserGovernanceRequest{
		RequesterUserId:  requesterID,
		TargetUserId:     userID,
		AccountStatus:    &statusValue,
		SuspensionReason: body.SuspensionReason,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "profile", resp.GetProfile())
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	requesterID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	pageSize, pageToken, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role *userv1.UserRole
	if rawRole := strings.TrimSpace(c.Query("role")); rawRole != "" {
		parsedRole, err := parseUserRole(rawRole)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		role = &parsedRole
	}

	var accountStatus *userv1.AccountStatus
	if rawStatus := strings.TrimSpace(c.Query("status")); rawStatus != "" {
		parsedStatus, err := parseAccountStatus(rawStatus)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		accountStatus = &parsedStatus
	}

	resp, err := h.client.ListUsers(c.Request.Context(), &userv1.ListUsersRequest{
		RequesterUserId: requesterID,
		Q:               strings.TrimSpace(c.Query("q")),
		Role:            role,
		AccountStatus:   accountStatus,
		Page:            &userv1.PagingRequest{PageSize: pageSize, PageToken: pageToken},
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	itemsPayload, err := protoSliceToAny(resp.GetUsers())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	nextPageToken := ""
	if resp.GetPage() != nil {
		nextPageToken = resp.GetPage().GetNextPageToken()
	}
	c.JSON(http.StatusOK, gin.H{"users": itemsPayload, "next_page_token": nextPageToken})
}

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

	resp, err := h.client.CreateMyPortfolioItem(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "item", resp.GetItem())
}

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

	resp, err := h.client.UpdateMyPortfolioItem(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "item", resp.GetItem())
}

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

func (h *UserHandler) ListPublicPortfolioItems(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("userId"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	pageSize, pageToken, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.ListPublicPortfolioItems(c.Request.Context(), &userv1.ListPublicPortfolioItemsRequest{UserId: userID, Page: &userv1.PagingRequest{PageSize: pageSize, PageToken: pageToken}})
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

func (h *UserHandler) CreateMeEmployment(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.CreateMyEmploymentRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.CreateMyEmployment(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "employment", resp.GetEmployment())
}

func (h *UserHandler) UpdateMeEmployment(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	employmentID, err := parseInt64PathParam(c, "employmentId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req := &userv1.UpdateMyEmploymentRequest{UserId: userID, EmploymentId: employmentID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.UpdateMyEmployment(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "employment", resp.GetEmployment())
}

func (h *UserHandler) DeleteMeEmployment(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	employmentID, err := parseInt64PathParam(c, "employmentId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.DeleteMyEmployment(c.Request.Context(), &userv1.DeleteMyEmploymentRequest{UserId: userID, EmploymentId: employmentID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": resp.GetDeleted()})
}

func (h *UserHandler) ListPublicEmployment(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("userId"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	pageSize, pageToken, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.ListPublicEmployment(c.Request.Context(), &userv1.ListPublicEmploymentRequest{UserId: userID, Page: &userv1.PagingRequest{PageSize: pageSize, PageToken: pageToken}})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	itemsPayload, err := protoSliceToAny(resp.GetEmployment())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	nextPageToken := ""
	if resp.GetPage() != nil {
		nextPageToken = resp.GetPage().GetNextPageToken()
	}
	c.JSON(http.StatusOK, gin.H{"employment": itemsPayload, "next_page_token": nextPageToken})
}

func (h *UserHandler) CreateMeEducation(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.CreateMyEducationRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.CreateMyEducation(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "education", resp.GetEducation())
}

func (h *UserHandler) UpdateMeEducation(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	educationID, err := parseInt64PathParam(c, "educationId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req := &userv1.UpdateMyEducationRequest{UserId: userID, EducationId: educationID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.UpdateMyEducation(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "education", resp.GetEducation())
}

func (h *UserHandler) DeleteMeEducation(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	educationID, err := parseInt64PathParam(c, "educationId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.DeleteMyEducation(c.Request.Context(), &userv1.DeleteMyEducationRequest{UserId: userID, EducationId: educationID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": resp.GetDeleted()})
}

func (h *UserHandler) ListPublicEducation(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("userId"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	pageSize, pageToken, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.ListPublicEducation(c.Request.Context(), &userv1.ListPublicEducationRequest{UserId: userID, Page: &userv1.PagingRequest{PageSize: pageSize, PageToken: pageToken}})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	itemsPayload, err := protoSliceToAny(resp.GetEducation())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	nextPageToken := ""
	if resp.GetPage() != nil {
		nextPageToken = resp.GetPage().GetNextPageToken()
	}
	c.JSON(http.StatusOK, gin.H{"education": itemsPayload, "next_page_token": nextPageToken})
}

func (h *UserHandler) CreateMeCertification(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.CreateMyCertificationRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.CreateMyCertification(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "certification", resp.GetCertification())
}

func (h *UserHandler) UpdateMeCertification(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	certificationID, err := parseInt64PathParam(c, "certificationId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req := &userv1.UpdateMyCertificationRequest{UserId: userID, CertificationId: certificationID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.UpdateMyCertification(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "certification", resp.GetCertification())
}

func (h *UserHandler) DeleteMeCertification(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	certificationID, err := parseInt64PathParam(c, "certificationId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.DeleteMyCertification(c.Request.Context(), &userv1.DeleteMyCertificationRequest{UserId: userID, CertificationId: certificationID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": resp.GetDeleted()})
}

func (h *UserHandler) ListPublicCertifications(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("userId"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	pageSize, pageToken, err := parsePagination(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.ListPublicCertifications(c.Request.Context(), &userv1.ListPublicCertificationsRequest{UserId: userID, Page: &userv1.PagingRequest{PageSize: pageSize, PageToken: pageToken}})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	itemsPayload, err := protoSliceToAny(resp.GetCertifications())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	nextPageToken := ""
	if resp.GetPage() != nil {
		nextPageToken = resp.GetPage().GetNextPageToken()
	}
	c.JSON(http.StatusOK, gin.H{"certifications": itemsPayload, "next_page_token": nextPageToken})
}

func (h *UserHandler) UpsertMeLanguages(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.UpsertMyLanguagesRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.UpsertMyLanguages(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	itemsPayload, err := protoSliceToAny(resp.GetLanguages())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"languages": itemsPayload})
}

func (h *UserHandler) GetPublicLanguages(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("userId"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	resp, err := h.client.GetPublicLanguages(c.Request.Context(), &userv1.GetPublicLanguagesRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	itemsPayload, err := protoSliceToAny(resp.GetLanguages())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"languages": itemsPayload})
}

func (h *UserHandler) SetMeAvailability(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.PatchMyAvailabilityRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.PatchMyAvailability(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "settings", resp.GetSettings())
}

func (h *UserHandler) GetMeAvailability(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetMyAvailability(c.Request.Context(), &userv1.GetMyAvailabilityRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "settings", resp.GetSettings())
}

func (h *UserHandler) SetMeRates(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.PatchMyRatesRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.PatchMyRates(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "settings", resp.GetSettings())
}

func (h *UserHandler) GetMeRates(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetMyRates(c.Request.Context(), &userv1.GetMyRatesRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "settings", resp.GetSettings())
}

func (h *UserHandler) SetMeWorkPreferences(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.PatchMyWorkPreferencesRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.PatchMyWorkPreferences(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "settings", resp.GetSettings())
}

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

	resp, err := h.client.PatchMyHiringPreferences(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "preferences", resp.GetPreferences())
}

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

	resp, err := h.client.UpsertFreelancerNote(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "note", resp.GetNote())
}

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
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
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

func writeProtoEnvelope(c *gin.Context, statusCode int, key string, msg proto.Message) {
	payload, err := protoToAny(msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(statusCode, gin.H{key: payload})
}

func protoToAny(msg proto.Message) (any, error) {
	if msg == nil {
		return nil, nil
	}
	raw, err := protoJSON.Marshal(msg)
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
