package handlers

import (
	"net/http"

	recommendationv1 "jobconnect/recommendation/gen/recommendation/v1"

	"github.com/gin-gonic/gin"
)

type RecommendationHandler struct {
	client recommendationv1.RecommendationServiceClient
}

func NewRecommendationHandler(client recommendationv1.RecommendationServiceClient) *RecommendationHandler {
	return &RecommendationHandler{client: client}
}

func (h *RecommendationHandler) GetRecommendedJobs(c *gin.Context) {
	userID, ok := callerUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	resp, err := h.client.GetRecommendedJobs(c.Request.Context(), &recommendationv1.GetRecommendedJobsRequest{
		UserId: userID,
		Limit:  int32(parseIntQuery(c, "limit", 10)),
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	payload, convErr := protoSliceToAny(resp.GetRecommendations())
	if convErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize response"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"recommendations": payload})
}
