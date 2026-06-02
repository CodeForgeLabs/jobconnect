package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireRolesRejectsMissingRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	RequireRoles("freelancer")(ctx)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestRequireRolesRejectsWrongRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Set(ContextRole, "client")

	RequireRoles("freelancer")(ctx)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestRequireRolesAllowsMatchingRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Set(ContextRole, "freelancer")

	nextCalled := false
	h := RequireRoles("freelancer")
	h(ctx)
	if !ctx.IsAborted() {
		nextCalled = true
	}
	if !nextCalled {
		t.Fatalf("expected request to continue for matching role")
	}
}
