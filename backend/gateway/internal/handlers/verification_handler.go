package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"jobconnect/gateway/internal/middleware"
	verificationv1 "jobconnect/verification/gen/verification/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

type VerificationHandler struct {
	client verificationv1.VerificationServiceClient
}

func NewVerificationHandler(client verificationv1.VerificationServiceClient) *VerificationHandler {
	return &VerificationHandler{client: client}
}

type submitVerificationRequest struct {
	LegalName            string   `json:"legal_name" binding:"required"`
	CountryCode          string   `json:"country_code" binding:"required,len=2"`
	DocumentType         string   `json:"document_type" binding:"required"`
	DocumentNumberMasked string   `json:"document_number_masked" binding:"required"`
	EvidenceURLs         []string `json:"evidence_urls"`
	SubmissionNote       string   `json:"submission_note"`
}

type reviewVerificationRequest struct {
	Decision        string `json:"decision" binding:"required,oneof=approve reject"`
	RejectionReason string `json:"rejection_reason"`
	InternalNote    string `json:"internal_note"`
}

type requestReverificationRequest struct {
	UserID            string `json:"user_id" binding:"required,uuid"`
	Reason            string `json:"reason" binding:"required"`
	ReverifyDueAtUnix int64  `json:"reverify_due_at_unix" binding:"required"`
}

func (h *VerificationHandler) Submit(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req submitVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.SubmitVerification(c.Request.Context(), &verificationv1.SubmitVerificationRequest{
		UserId:               userID,
		LegalName:            req.LegalName,
		CountryCode:          strings.ToUpper(req.CountryCode),
		DocumentType:         req.DocumentType,
		DocumentNumberMasked: req.DocumentNumberMasked,
		EvidenceUrls:         req.EvidenceURLs,
		SubmissionNote:       req.SubmissionNote,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"request": resp.GetRequest()})
}

func (h *VerificationHandler) GetMyStatus(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetMyVerificationStatus(c.Request.Context(), &verificationv1.GetMyVerificationStatusRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"request": resp.GetRequest()})
}

func (h *VerificationHandler) ListPending(c *gin.Context) {
	pageSize := int32(parseIntQuery(c, "page_size", 20))
	page := int32(parseIntQuery(c, "page", 1))

	resp, err := h.client.ListPendingVerifications(c.Request.Context(), &verificationv1.ListPendingVerificationsRequest{
		PageSize: pageSize,
		Page:     page,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"requests": resp.GetRequests()})
}

func (h *VerificationHandler) GetByID(c *gin.Context) {
	requestID, err := strconv.ParseInt(c.Param("requestId"), 10, 64)
	if err != nil || requestID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid requestId"})
		return
	}

	resp, rpcErr := h.client.GetVerificationRequest(c.Request.Context(), &verificationv1.GetVerificationRequestRequest{RequestId: requestID})
	if rpcErr != nil {
		writeGRPCError(c, rpcErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{"request": resp.GetRequest()})
}

func (h *VerificationHandler) Review(c *gin.Context) {
	reviewerID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	requestID, err := strconv.ParseInt(c.Param("requestId"), 10, 64)
	if err != nil || requestID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid requestId"})
		return
	}

	var req reviewVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, rpcErr := h.client.ReviewVerification(c.Request.Context(), &verificationv1.ReviewVerificationRequest{
		RequestId:       requestID,
		ReviewerUserId:  reviewerID,
		Decision:        req.Decision,
		RejectionReason: req.RejectionReason,
		InternalNote:    req.InternalNote,
	})
	if rpcErr != nil {
		writeGRPCError(c, rpcErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{"request": resp.GetRequest()})
}

func (h *VerificationHandler) RequestReverification(c *gin.Context) {
	reviewerID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req requestReverificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ReverifyDueAtUnix <= time.Now().UTC().Unix() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reverify_due_at_unix must be in the future"})
		return
	}

	resp, err := h.client.RequestReverification(c.Request.Context(), &verificationv1.RequestReverificationRequest{
		UserId:            req.UserID,
		ReviewerUserId:    reviewerID,
		Reason:            req.Reason,
		ReverifyDueAtUnix: req.ReverifyDueAtUnix,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"request": resp.GetRequest()})
}

func callerUserID(c *gin.Context) (string, bool) {
	v, ok := c.Get(middleware.ContextUserID)
	if !ok {
		return "", false
	}
	id, ok := v.(string)
	if !ok || strings.TrimSpace(id) == "" {
		return "", false
	}
	return id, true
}
func callerUserRole(c *gin.Context) (string, bool) {
	v, ok := c.Get(middleware.ContextRole)
	if !ok {
		return "", false
	}
	role, ok := v.(string)
	if !ok || strings.TrimSpace(role) == "" {
		return "", false
	}
	return role, true
}
func parseIntQuery(c *gin.Context, key string, def int) int {
	v := strings.TrimSpace(c.Query(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func attachUserID(ctx context.Context, userID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "user_id", userID)
}

func attachUserIdAndRole(ctx context.Context, userID, role string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "user_id", userID, "role", role)
}
