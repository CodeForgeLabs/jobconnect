package router

import (
	"testing"

	"jobconnect/gateway/internal/config"
	"jobconnect/gateway/internal/handlers"

	"github.com/gin-gonic/gin"
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

func TestUserRoutes_DoNotExposeCreateOrGetSinglePortfolioEndpoints(t *testing.T) {
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
	)

	for _, route := range engine.Routes() {
		if route.Path == "/api/v1/users/me/profile" && route.Method == "POST" {
			t.Fatalf("did not expect POST /api/v1/users/me/profile to be registered")
		}
		if route.Path == "/api/v1/users/me/portfolio/:itemId" && route.Method == "GET" {
			t.Fatalf("did not expect GET /api/v1/users/me/portfolio/:itemId to be registered")
		}
	}
}
