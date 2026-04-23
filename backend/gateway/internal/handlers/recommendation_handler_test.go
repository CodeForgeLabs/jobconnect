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
	"google.golang.org/grpc/metadata"
)

type recommendationServiceClientStub struct {
	lastJobsReq         *recommendationv1.GetRecommendedJobsRequest
	lastJobsAuth        string
	lastFreelancersReq  *recommendationv1.GetRecommendedFreelancersRequest
	lastFreelancersAuth string
}

func (s *recommendationServiceClientStub) GetRecommendedJobs(ctx context.Context, in *recommendationv1.GetRecommendedJobsRequest, opts ...grpc.CallOption) (*recommendationv1.GetRecommendedJobsResponse, error) {
	s.lastJobsReq = in
	s.lastJobsAuth = outgoingAuth(ctx)
	return &recommendationv1.GetRecommendedJobsResponse{
		Recommendations: []*recommendationv1.JobRecommendation{
			{JobId: 11, MatchScore: 0.92, MatchReason: "Matches your skills in Go"},
		},
	}, nil
}

func (s *recommendationServiceClientStub) GetRecommendedFreelancers(ctx context.Context, in *recommendationv1.GetRecommendedFreelancersRequest, opts ...grpc.CallOption) (*recommendationv1.GetRecommendedFreelancersResponse, error) {
	s.lastFreelancersReq = in
	s.lastFreelancersAuth = outgoingAuth(ctx)
	return &recommendationv1.GetRecommendedFreelancersResponse{
		Recommendations: []*recommendationv1.FreelancerRecommendation{
			{UserId: "freelancer-1", MatchScore: 0.88, MatchReason: "Matches required skills: Go"},
		},
	}, nil
}

func outgoingAuth(ctx context.Context) string {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get("authorization")
	if len(values) == 0 {
		return ""
	}
	return values[0]
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
	req.Header.Set("Authorization", "Bearer test-token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if client.lastJobsReq == nil {
		t.Fatal("expected recommendation request to be sent")
	}
	if client.lastJobsReq.UserId != "user-123" {
		t.Fatalf("expected user_id user-123, got %q", client.lastJobsReq.UserId)
	}
	if client.lastJobsReq.Limit != 3 {
		t.Fatalf("expected limit 3, got %d", client.lastJobsReq.Limit)
	}
	if client.lastJobsAuth != "Bearer test-token" {
		t.Fatalf("expected forwarded authorization metadata, got %q", client.lastJobsAuth)
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

func TestRecommendationHandlerGetRecommendedFreelancers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	client := &recommendationServiceClientStub{}
	handler := NewRecommendationHandler(client)

	router := gin.New()
	router.GET("/recommendations/jobs/:jobId/freelancers", handler.GetRecommendedFreelancers)

	req := httptest.NewRequest(http.MethodGet, "/recommendations/jobs/42/freelancers?limit=4", nil)
	req.Header.Set("Authorization", "Bearer client-token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if client.lastFreelancersReq == nil {
		t.Fatal("expected freelancer recommendation request to be sent")
	}
	if client.lastFreelancersReq.JobId != 42 {
		t.Fatalf("expected job_id 42, got %d", client.lastFreelancersReq.JobId)
	}
	if client.lastFreelancersReq.Limit != 4 {
		t.Fatalf("expected limit 4, got %d", client.lastFreelancersReq.Limit)
	}
	if client.lastFreelancersAuth != "Bearer client-token" {
		t.Fatalf("expected forwarded authorization metadata, got %q", client.lastFreelancersAuth)
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
