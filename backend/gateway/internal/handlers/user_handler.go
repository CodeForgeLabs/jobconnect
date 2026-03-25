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

func (h *UserHandler) GetMeUser(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetUser(c.Request.Context(), &userv1.GetUserRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "user", resp.GetUser())
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("userId"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	resp, err := h.client.GetUser(c.Request.Context(), &userv1.GetUserRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	writeProtoEnvelope(c, http.StatusOK, "user", resp.GetUser())
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
