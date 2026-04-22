package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"jobconnect/gateway/internal/auth"
	"jobconnect/gateway/internal/config"
	"jobconnect/gateway/internal/handlers"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestUserPortfolioUpdateRouteUsesPut(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		JWTSecret: []byte("test-secret"),
	}

	engine := New(
		cfg,
		&handlers.AuthHandler{},
		&handlers.VerificationHandler{},
		&handlers.UserHandler{},
		&handlers.JobHandler{},
		&handlers.ProposalHandler{},
		nil,
		&handlers.RecommendationHandler{},
		&handlers.ChatHandler{},
	)

	var foundPut bool
	var foundPatch bool
	for _, route := range engine.Routes() {
		if route.Path == "/api/v1/users/me/portfolio/:itemId" {
			switch route.Method {
			case "PUT":
				foundPut = true
			case "PATCH":
				foundPatch = true
			}
		}
	}

	if !foundPut {
		t.Fatalf("expected PUT /api/v1/users/me/portfolio/:itemId to be registered")
	}
	if foundPatch {
		t.Fatalf("did not expect PATCH /api/v1/users/me/portfolio/:itemId to be registered")
	}
}

func TestUserRoutesDoNotExposePublicAdminInternalUserRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		JWTSecret: []byte("test-secret"),
	}

	engine := New(
		cfg,
		&handlers.AuthHandler{},
		&handlers.VerificationHandler{},
		&handlers.UserHandler{},
		&handlers.JobHandler{},
		&handlers.ProposalHandler{},
		nil,
		&handlers.RecommendationHandler{},
		&handlers.ChatHandler{},
	)

	for _, route := range engine.Routes() {
		switch route.Path {
		case "/api/v1/public/users/:userId/profile",
			"/api/v1/public/users/:userId/portfolio",
			"/api/v1/admin/users",
			"/api/v1/admin/users/:userId/profile",
			"/api/v1/admin/users/:userId/account-status",
			"/api/v1/internal/users/:userId/basic",
			"/api/v1/internal/users/:userId/profile":
			t.Fatalf("did not expect removed route to remain registered: %s %s", route.Method, route.Path)
		}
	}
}

func TestUserRoutes_DoNotExposeCreateProfileButExposeGetSinglePortfolioEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		JWTSecret: []byte("test-secret"),
	}

	engine := New(
		cfg,
		&handlers.AuthHandler{},
		&handlers.VerificationHandler{},
		&handlers.UserHandler{},
		&handlers.JobHandler{},
		&handlers.ProposalHandler{},
		nil,
		&handlers.RecommendationHandler{},
		&handlers.ChatHandler{},
	)

	var foundGetPortfolioItem bool
	for _, route := range engine.Routes() {
		if route.Path == "/api/v1/users/me/profile" && route.Method == "POST" {
			t.Fatalf("did not expect POST /api/v1/users/me/profile to be registered")
		}
		if route.Path == "/api/v1/users/me/portfolio/:itemId" && route.Method == "GET" {
			foundGetPortfolioItem = true
		}
	}

	if !foundGetPortfolioItem {
		t.Fatalf("expected GET /api/v1/users/me/portfolio/:itemId to be registered")
	}
}

func TestUserRoutesExposePortfolioMediaUploadURLRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		JWTSecret: []byte("test-secret"),
	}

	engine := New(
		cfg,
		&handlers.AuthHandler{},
		&handlers.VerificationHandler{},
		&handlers.UserHandler{},
		&handlers.JobHandler{},
		&handlers.ProposalHandler{},
		nil,
		&handlers.RecommendationHandler{},
		&handlers.ChatHandler{},
	)

	for _, route := range engine.Routes() {
		if route.Path == "/api/v1/users/me/portfolio/media/upload-url" && route.Method == "POST" {
			return
		}
	}

	t.Fatalf("expected POST /api/v1/users/me/portfolio/media/upload-url to be registered")
}

func TestUserPortfolioRoutesRejectNonFreelancerRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := []byte("test-secret")
	cfg := config.Config{JWTSecret: secret}

	engine := New(
		cfg,
		&handlers.AuthHandler{},
		&handlers.VerificationHandler{},
		&handlers.UserHandler{},
		&handlers.JobHandler{},
		&handlers.ProposalHandler{},
		nil,
		&handlers.RecommendationHandler{},
		&handlers.ChatHandler{},
	)

	clientToken := signTestAccessToken(t, secret, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "client")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/portfolio", nil)
	req.Header.Set("Authorization", "Bearer "+clientToken)
	rec := httptest.NewRecorder()

	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected %d for non-freelancer role, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestProposalDecisionRouteIsCanonicalAndApplicantStageRouteIsRemoved(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{
		JWTSecret: []byte("test-secret"),
	}

	engine := New(
		cfg,
		&handlers.AuthHandler{},
		&handlers.VerificationHandler{},
		&handlers.UserHandler{},
		&handlers.JobHandler{},
		&handlers.ProposalHandler{},
		nil,
		&handlers.RecommendationHandler{},
		&handlers.ChatHandler{},
	)

	var hasProposalDecision bool
	var hasApplicantStage bool
	for _, route := range engine.Routes() {
		if route.Method == http.MethodPost && route.Path == "/api/v1/proposals/:proposalId/decision" {
			hasProposalDecision = true
		}
		if route.Method == http.MethodPost && route.Path == "/api/v1/jobs/applicants/:proposalId/stage" {
			hasApplicantStage = true
		}
	}

	if !hasProposalDecision {
		t.Fatalf("expected POST /api/v1/proposals/:proposalId/decision to be registered")
	}
	if hasApplicantStage {
		t.Fatalf("did not expect POST /api/v1/jobs/applicants/:proposalId/stage to be registered")
	}
}

func TestContractRoutesExposeLifecycleAndBootstrap(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.Config{JWTSecret: []byte("test-secret")}

	engine := New(
		cfg,
		&handlers.AuthHandler{},
		&handlers.VerificationHandler{},
		&handlers.UserHandler{},
		&handlers.JobHandler{},
		&handlers.ProposalHandler{},
		&handlers.ContractHandler{},
		&handlers.RecommendationHandler{},
		&handlers.ChatHandler{},
	)

	want := map[string]bool{
		"GET /api/v1/contracts/bootstrap":            false,
		"POST /api/v1/contracts":                     false,
		"GET /api/v1/contracts":                      false,
		"GET /api/v1/contracts/:contractId":          false,
		"POST /api/v1/contracts/:contractId/accept":  false,
		"POST /api/v1/contracts/:contractId/decline": false,
		"POST /api/v1/contracts/:contractId/revoke":  false,
	}

	for _, route := range engine.Routes() {
		key := route.Method + " " + route.Path
		if _, ok := want[key]; ok {
			want[key] = true
		}
	}

	for key, found := range want {
		if !found {
			t.Fatalf("expected route to be registered: %s", key)
		}
	}
}

func signTestAccessToken(t *testing.T, secret []byte, userID string, role string) string {
	t.Helper()
	claims := &auth.AccessClaims{UserID: userID, Role: role}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}
