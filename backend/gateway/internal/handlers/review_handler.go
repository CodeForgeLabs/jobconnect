package handlers

import (
	reviewsv1 "jobconnect/reviews/gen/reviews/v1"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ReviewHandler struct {
	client reviewsv1.ReviewServiceClient
}

// Swagger-only replicas for request/response docs.
type CreateReviewSwaggerDTO struct {
	ContractId   int64  `json:"contract_id" example:"101"`
	ReviewerRole string `json:"reviewer_role" example:"CLIENT" enums:"CLIENT,FREELANCER"`
	Rating       int32  `json:"rating" example:"5"`
	Title        string `json:"title" example:"Great work!"`
	Comment      string `json:"comment" example:"The freelancer delivered on time and exceeded expectations."`
}

type UpdateReviewSwaggerDTO struct {
	Rating  int32  `json:"rating" example:"5"`
	Title   string `json:"title" example:"Great work!"`
	Comment string `json:"comment" example:"The freelancer delivered on time and exceeded expectations."`
}

type ReviewSwaggerDTO struct {
	Id           int64  `json:"id" example:"1"`
	ContractId   int64  `json:"contract_id" example:"101"`
	ReviewerRole string `json:"reviewer_role" example:"CLIENT"`
	Rating       int32  `json:"rating" example:"5"`
	Title        string `json:"title" example:"Great work!"`
	Comment      string `json:"comment" example:"The freelancer delivered on time and exceeded expectations."`
}

type ReviewsListSwaggerDTO struct {
	Reviews []ReviewSwaggerDTO `json:"reviews"`
}

type DeleteReviewSwaggerDTO struct {
	Message string `json:"message" example:"review deleted successfully"`
}

// CreateReviewDTO represents the request body for creating a review
type CreateReviewDTO struct {
	ContractId   int64  `json:"contract_id" binding:"required" example:"101"`
	ReviewerRole string `json:"reviewer_role" binding:"required" enums:"CLIENT,FREELANCER" example:"CLIENT"`
	Rating       int32  `json:"rating" binding:"required,min=1,max=5" example:"5"`
	Title        string `json:"title" example:"Great work!"`
	Comment      string `json:"comment" example:"The freelancer delivered on time and exceeded expectations."`
}

func NewReviewHandler(client reviewsv1.ReviewServiceClient) *ReviewHandler {
	return &ReviewHandler{client: client}
}

// CreateReview godoc
// @Summary      Create a new review
// @Description  Allows an authenticated user to create a review for a specific contract.
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Security BearerAuth
// @Param        review  body      CreateReviewSwaggerDTO  true  "Review Details"
// @Success      200     {object}  ReviewSwaggerDTO
// @Failure      400     {object}  map[string]string      "Invalid request body"
// @Failure      401     {object}  map[string]string      "Unauthorized"
// @Failure      500     {object}  map[string]string      "Internal server error"
// @Router       /api/v1/reviews [post]
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}

	var dto CreateReviewDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Map DTO to the Proto Request for the gRPC call
	req := &reviewsv1.CreateReviewRequest{
		ContractId: dto.ContractId,
		Rating:     dto.Rating,
		Title:      dto.Title,
		Comment:    dto.Comment,
		// You'll need a helper to map the string role to the Enum type
		ReviewerRole: MapStringToRole(dto.ReviewerRole),
	}

	ctx := attachUserID(c.Request.Context(), userID)
	resp, err := h.client.CreateReview(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetReview godoc
// @Summary      Get a review by ID
// @Tags         reviews
// @Produce      json
// @Security BearerAuth
// @Param        id   path      int  true  "Review ID"
// @Success      200  {object}  ReviewSwaggerDTO
// @Failure      400  {object}  map[string]string  "Invalid review id"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/reviews/{id} [get]
func (h *ReviewHandler) GetReview(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}

	idParam := c.Param("id")
	reviewID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	// Map DTO to the Proto Request for the gRPC call
	req := &reviewsv1.GetReviewRequest{
		Id: reviewID,
	}
	ctx := attachUserID(c.Request.Context(), userID)
	resp, err := h.client.GetReview(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)

}

// UpdateReview godoc
// @Summary      Update a review
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Security BearerAuth
// @Param        id      path      int                   true  "Review ID"
// @Param        review   body      UpdateReviewSwaggerDTO true  "Review Details"
// @Success      200      {object}  ReviewSwaggerDTO
// @Failure      400      {object}  map[string]string      "Invalid request"
// @Failure      401      {object}  map[string]string      "Unauthorized"
// @Failure      500      {object}  map[string]string      "Internal server error"
// @Router       /api/v1/reviews/{id} [put]
func (h *ReviewHandler) UpdateReview(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}

	// Request body structure
	var req struct {
		Rating  int32  `json:"rating"`
		Title   string `json:"title"`
		Comment string `json:"comment"`
	}

	// Get review ID from URL param
	idParam := c.Param("id")
	reviewID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	// Bind JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Optional validation
	if req.Rating < 1 || req.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rating must be between 1 and 5"})
		return
	}

	// Attach authenticated user context
	ctx := attachUserID(c.Request.Context(), userID)

	// Call gRPC service
	resp, err := h.client.UpdateReview(ctx, &reviewsv1.UpdateReviewRequest{
		Id:      reviewID,
		Rating:  req.Rating,
		Title:   req.Title,
		Comment: req.Comment,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListReviews godoc
// @Summary      List reviews for the authenticated user
// @Tags         reviews
// @Produce      json
// @Security BearerAuth
// @Param        offset  query     int  false  "Pagination offset"  default(0)
// @Param        limit   query     int  false  "Pagination limit"   default(10)
// @Success      200     {object}  ReviewsListSwaggerDTO
// @Failure      401     {object}  map[string]string  "Unauthorized"
// @Failure      500     {object}  map[string]string  "Internal server error"
// @Router       /api/v1/reviews [get]
func (h *ReviewHandler) ListReviews(c *gin.Context) {
	offset := int32(parseIntQuery(c, "offset", 0))
	limit := int32(parseIntQuery(c, "limit", 10))

	userID, ok := callerUserID(c)
	role, roleOk := callerUserRole(c)

	if !ok || !roleOk {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}

	// Attach authenticated user info
	ctx := attachUserIdAndRole(c.Request.Context(), userID, role)

	// Call gRPC service
	resp, err := h.client.ListReviewsByUser(ctx, &reviewsv1.ListReviewsByUserRequest{
		UserId: userID,
		Role:   MapStringToRole(role),
		Limit:  limit,
		Offset: offset,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteReview godoc
// @Summary      Delete a review
// @Tags         reviews
// @Produce      json
// @Security BearerAuth
// @Param        id   path      int  true  "Review ID"
// @Success      200  {object}  DeleteReviewSwaggerDTO
// @Failure      400  {object}  map[string]string  "Invalid review id"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /api/v1/reviews/{id} [delete]
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "auth required"})
		return
	}

	idParam := c.Param("id")
	reviewID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}

	ctx := attachUserID(c.Request.Context(), userID)

	resp, err := h.client.DeleteReview(ctx, &reviewsv1.DeleteReviewRequest{
		Id: reviewID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func MapStringToRole(roleStr string) reviewsv1.ReviewerRole {
	// Normalize the input to uppercase
	normalized := strings.ToUpper(strings.TrimSpace(roleStr))

	roleMap := map[string]reviewsv1.ReviewerRole{
		"CLIENT":     reviewsv1.ReviewerRole_CLIENT,
		"FREELANCER": reviewsv1.ReviewerRole_FREELANCER,
	}

	if role, exists := roleMap[normalized]; exists {
		return role
	}

	// Default to unspecified if the input is garbage
	return reviewsv1.ReviewerRole_REVIEWER_ROLE_UNSPECIFIED
}
