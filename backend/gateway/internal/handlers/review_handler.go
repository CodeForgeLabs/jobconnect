package handlers

import (
	reviewsv1 "jobconnect/reviews/gen/reviews/v1"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	client reviewsv1.ReviewServiceClient
}

func NewReviewHandler(client reviewsv1.ReviewServiceClient) *ReviewHandler {
	return &ReviewHandler{client: client}
}

func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}

	var req reviewsv1.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := attachUserID(c.Request.Context(), userID)

	resp, err := h.client.CreateReview(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ReviewHandler) GetReview(c *gin.Context) {
	// return h.client.GetReview(ctx, req)
}

func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	// return h.client.UpdateReview(ctx, req)
}

func (h *ReviewHandler) ListReviews(c *gin.Context) {
	// return h.client.ListReviews(ctx, req)
}

func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	// return h.client.DeleteReview(ctx, req)
}
