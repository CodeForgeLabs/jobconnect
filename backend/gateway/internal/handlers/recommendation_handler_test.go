package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"jobconnect/gateway/internal/middleware"
	recommendationv1 "jobconnect/recommendation/gen/recommendation/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type recommendationServiceClientStub struct {
	lastReq *recommendationv1.GetRecommendedJobsRequest
}

func (s *recommendationServiceClientStub) GetRecommendedJobs(ctx context.Context, in *recommendationv1.GetRecommendedJobsRequest, opts ...grpc.CallOption) (*recommendationv1.GetRecommendedJobsResponse, error) {
	s.lastReq = in
	return &recommendationv1.GetRecommendedJobsResponse{
		Recommendations: []*recommendationv1.JobRecommendation{
			{JobId: 11, MatchScore: 0.92, MatchReason: "Matches your skills in Go"},
		},
	}, nil
}

func (s *recommendationServiceClientStub) GetRecommendedFreelancers(ctx context.Context, in *recommendationv1.GetRecommendedFreelancersRequest, opts ...grpc.CallOption) (*recommendationv1.GetRecommendedFreelancersResponse, error) {
	return &recommendationv1.GetRecommendedFreelancersResponse{}, nil
}

func TestRecommendationHandlerGetRecommendedJobs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := &recommendationServiceClientStub{}
	handler := NewRecommendationHandler(client)

	router := gin.New()
	router.GET("/recommendations/jobs", func(c *gin.Context) {
		c.Set(middleware.ContextUserID, "user-123")
		handler.GetRecommendedJobs(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/recommendations/jobs?limit=3", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if client.lastReq == nil {
		t.Fatal("expected recommendation request to be sent")
	}
	if client.lastReq.UserId != "user-123" {
		t.Fatalf("expected user_id user-123, got %q", client.lastReq.UserId)
	}
	if client.lastReq.Limit != 3 {
		t.Fatalf("expected limit 3, got %d", client.lastReq.Limit)
	}

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	recommendations, ok := body["recommendations"].([]any)
	if !ok || len(recommendations) != 1 {
		t.Fatalf("expected one recommendation, got %#v", body["recommendations"])
	}
}
