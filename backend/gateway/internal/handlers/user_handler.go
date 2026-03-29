package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	userv1 "jobconnect/user/gen/user"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var protoJSON = protojson.MarshalOptions{
	UseProtoNames:   true,
	UseEnumNumbers:  false,
	EmitUnpopulated: false,
}

type UserHandler struct {
	client userv1.UserServiceClient
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

func NewUserHandler(client userv1.UserServiceClient) *UserHandler {
	return &UserHandler{client: client}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetProfile(c.Request.Context(), &userv1.GetProfileRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "profile", resp.GetProfile())
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("userId"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	resp, err := h.client.GetProfile(c.Request.Context(), &userv1.GetProfileRequest{UserId: userID})
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

	var availability *userv1.Availability
	if body.Availability != nil {
		parsed, err := parseAvailability(*body.Availability)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		availability = &parsed
	}

	resp, err := h.client.UpdateProfile(c.Request.Context(), &userv1.UpdateProfileRequest{
		UserId:           userID,
		DisplayName:      body.DisplayName,
		AvatarUrl:        body.AvatarURL,
		Language:         body.Language,
		ContactEmail:     body.ContactEmail,
		ContactPhone:     body.ContactPhone,
		Bio:              body.Bio,
		FirstName:        body.FirstName,
		LastName:         body.LastName,
		CompanyName:      body.CompanyName,
		BillingAddress:   body.BillingAddress,
		TaxId:            body.TaxID,
		Headline:         body.Headline,
		Skills:           body.Skills,
		ExperienceLevel:  body.ExperienceLevel,
		HourlyRate:       body.HourlyRate,
		Availability:     availability,
		Location:         body.Location,
		LastActiveAtUnix: body.LastActiveAtUnix,
	})
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

	resp, err := h.client.DeleteProfile(c.Request.Context(), &userv1.DeleteProfileRequest{UserId: userID, HardDelete: hardDelete})
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

	resp, err := h.client.GetOnboardingStatus(c.Request.Context(), &userv1.GetOnboardingStatusRequest{UserId: userID})
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

	resp, err := h.client.GetAccountSettings(c.Request.Context(), &userv1.GetAccountSettingsRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "settings", resp.GetSettings())
}

func (h *UserHandler) UpdateMeAccountSettings(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.UpdateAccountSettingsRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.UpdateAccountSettings(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "settings", resp.GetSettings())
}

func (h *UserHandler) GetMePrivacySettings(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetPrivacySettings(c.Request.Context(), &userv1.GetPrivacySettingsRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "settings", resp.GetSettings())
}

func (h *UserHandler) UpdateMePrivacySettings(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.UpdatePrivacySettingsRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.UpdatePrivacySettings(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "settings", resp.GetSettings())
}

func (h *UserHandler) GetMeNotificationSettings(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetNotificationSettings(c.Request.Context(), &userv1.GetNotificationSettingsRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "settings", resp.GetSettings())
}

func (h *UserHandler) UpdateMeNotificationSettings(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.UpdateNotificationSettingsRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.UpdateNotificationSettings(c.Request.Context(), req)
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "settings", resp.GetSettings())
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

	resp, err := h.client.UploadAvatar(c.Request.Context(), &userv1.UploadAvatarRequest{
		UserId:      userID,
		FileName:    fileHeader.Filename,
		ContentType: contentType,
		Content:     content,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"avatar_url":   resp.GetAvatarUrl(),
		"preview_url":  resp.GetPreviewUrl(),
		"content_type": resp.GetContentType(),
		"size_bytes":   resp.GetSizeBytes(),
		"width":        resp.GetWidth(),
		"height":       resp.GetHeight(),
	})
}

func (h *UserHandler) GetMeAvatar(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetAvatar(c.Request.Context(), &userv1.GetAvatarRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	if fileName := strings.TrimSpace(resp.GetFileName()); fileName != "" {
		c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", fileName))
	}
	c.Data(http.StatusOK, resp.GetContentType(), resp.GetContent())
}

func (h *UserHandler) RemoveMeAvatar(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.RemoveAvatar(c.Request.Context(), &userv1.RemoveAvatarRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"removed": resp.GetRemoved()})
}

func (h *UserHandler) UpdateAccountStatus(c *gin.Context) {
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
	visibility, err := parseVisibility(body.Visibility)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.UpdateAccountStatus(c.Request.Context(), &userv1.UpdateAccountStatusRequest{
		UserId:           userID,
		Status:           status,
		Visibility:       visibility,
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

	resp, err := h.client.ListUsers(c.Request.Context(), &userv1.ListUsersRequest{
		RequesterUserId: requesterID,
		Q:               strings.TrimSpace(c.Query("q")),
		Role:            strings.TrimSpace(c.Query("role")),
		Status:          strings.TrimSpace(c.Query("status")),
		PageSize:        pageSize,
		PageToken:       pageToken,
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
	c.JSON(http.StatusOK, gin.H{"users": itemsPayload, "next_page_token": resp.GetNextPageToken()})
}

func (h *UserHandler) CreateMePortfolioItem(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.CreatePortfolioItemRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.CreatePortfolioItem(c.Request.Context(), req)
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

	req := &userv1.UpdatePortfolioItemRequest{UserId: userID, ItemId: itemID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.UpdatePortfolioItem(c.Request.Context(), req)
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

	resp, err := h.client.DeletePortfolioItem(c.Request.Context(), &userv1.DeletePortfolioItemRequest{UserId: userID, ItemId: itemID})
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

	resp, err := h.client.ListPublicPortfolioItems(c.Request.Context(), &userv1.ListPublicPortfolioItemsRequest{UserId: userID, PageSize: pageSize, PageToken: pageToken})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	itemsPayload, err := protoSliceToAny(resp.GetItems())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": itemsPayload, "next_page_token": resp.GetNextPageToken()})
}

func (h *UserHandler) CreateMeEmployment(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.CreateEmploymentRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.CreateEmployment(c.Request.Context(), req)
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

	req := &userv1.UpdateEmploymentRequest{UserId: userID, EmploymentId: employmentID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.UpdateEmployment(c.Request.Context(), req)
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

	resp, err := h.client.DeleteEmployment(c.Request.Context(), &userv1.DeleteEmploymentRequest{UserId: userID, EmploymentId: employmentID})
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

	resp, err := h.client.ListPublicEmployment(c.Request.Context(), &userv1.ListPublicEmploymentRequest{UserId: userID, PageSize: pageSize, PageToken: pageToken})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	itemsPayload, err := protoSliceToAny(resp.GetEmployment())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"employment": itemsPayload, "next_page_token": resp.GetNextPageToken()})
}

func (h *UserHandler) CreateMeEducation(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.CreateEducationRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.CreateEducation(c.Request.Context(), req)
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

	req := &userv1.UpdateEducationRequest{UserId: userID, EducationId: educationID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.UpdateEducation(c.Request.Context(), req)
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

	resp, err := h.client.DeleteEducation(c.Request.Context(), &userv1.DeleteEducationRequest{UserId: userID, EducationId: educationID})
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

	resp, err := h.client.ListPublicEducation(c.Request.Context(), &userv1.ListPublicEducationRequest{UserId: userID, PageSize: pageSize, PageToken: pageToken})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	itemsPayload, err := protoSliceToAny(resp.GetEducation())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"education": itemsPayload, "next_page_token": resp.GetNextPageToken()})
}

func (h *UserHandler) CreateMeCertification(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.CreateCertificationRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.CreateCertification(c.Request.Context(), req)
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

	req := &userv1.UpdateCertificationRequest{UserId: userID, CertificationId: certificationID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.UpdateCertification(c.Request.Context(), req)
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

	resp, err := h.client.DeleteCertification(c.Request.Context(), &userv1.DeleteCertificationRequest{UserId: userID, CertificationId: certificationID})
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

	resp, err := h.client.ListPublicCertifications(c.Request.Context(), &userv1.ListPublicCertificationsRequest{UserId: userID, PageSize: pageSize, PageToken: pageToken})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	itemsPayload, err := protoSliceToAny(resp.GetCertifications())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"certifications": itemsPayload, "next_page_token": resp.GetNextPageToken()})
}

func (h *UserHandler) UpsertMeLanguages(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	req := &userv1.UpsertLanguagesRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.UpsertLanguages(c.Request.Context(), req)
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

	req := &userv1.SetAvailabilityRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.SetAvailability(c.Request.Context(), req)
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

	resp, err := h.client.GetAvailability(c.Request.Context(), &userv1.GetAvailabilityRequest{UserId: userID})
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

	req := &userv1.SetRatesRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.SetRates(c.Request.Context(), req)
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

	resp, err := h.client.GetRates(c.Request.Context(), &userv1.GetRatesRequest{UserId: userID})
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

	req := &userv1.SetWorkPreferencesRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.SetWorkPreferences(c.Request.Context(), req)
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

	resp, err := h.client.GetWorkPreferences(c.Request.Context(), &userv1.GetWorkPreferencesRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "settings", resp.GetSettings())
}

func (h *UserHandler) GetMeClientProfile(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetProfile(c.Request.Context(), &userv1.GetProfileRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	profile := resp.GetProfile()
	if profile == nil || !strings.EqualFold(strings.TrimSpace(profile.GetRole()), "client") || profile.GetClient() == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client profile not found"})
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "profile", profile.GetClient())
}

func (h *UserHandler) UpdateMeClientProfile(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var body struct {
		CompanyName    *string `json:"company_name"`
		BillingAddress *string `json:"billing_address"`
		TaxID          *string `json:"tax_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.UpdateProfile(c.Request.Context(), &userv1.UpdateProfileRequest{
		UserId:         userID,
		CompanyName:    body.CompanyName,
		BillingAddress: body.BillingAddress,
		TaxId:          body.TaxID,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	profile := resp.GetProfile()
	if profile == nil || !strings.EqualFold(strings.TrimSpace(profile.GetRole()), "client") || profile.GetClient() == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "client profile not found"})
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "profile", profile.GetClient())
}

func (h *UserHandler) GetMeHiringPreferences(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetHiringPreferences(c.Request.Context(), &userv1.GetHiringPreferencesRequest{UserId: userID})
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

	req := &userv1.UpdateHiringPreferencesRequest{UserId: userID}
	if !bindProtoJSON(c, req) {
		return
	}

	resp, err := h.client.UpdateHiringPreferences(c.Request.Context(), req)
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

	resp, err := h.client.ListSavedFreelancers(c.Request.Context(), &userv1.ListSavedFreelancersRequest{UserId: userID, PageSize: pageSize, PageToken: pageToken})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	itemsPayload, err := protoSliceToAny(resp.GetFreelancers())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"freelancers": itemsPayload, "next_page_token": resp.GetNextPageToken()})
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

func parseVisibility(raw string) (userv1.ProfileVisibility, error) {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	normalized = strings.TrimPrefix(normalized, "PROFILE_VISIBILITY_")
	switch normalized {
	case "PUBLIC":
		return userv1.ProfileVisibility_PROFILE_VISIBILITY_PUBLIC, nil
	case "PRIVATE":
		return userv1.ProfileVisibility_PROFILE_VISIBILITY_PRIVATE, nil
	default:
		return userv1.ProfileVisibility_PROFILE_VISIBILITY_UNSPECIFIED, fmt.Errorf("invalid visibility")
	}
}
