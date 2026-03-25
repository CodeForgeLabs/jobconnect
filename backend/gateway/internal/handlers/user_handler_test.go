package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"jobconnect/gateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

func newJSONTestContext(method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(method, target, nil)
	return ctx, rec
}

func TestGetMeClientProfile_Unauthorized(t *testing.T) {
	h := &UserHandler{}
	ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/users/me/client-profile")

	h.GetMeClientProfile(ctx)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestListUsers_Unauthorized(t *testing.T) {
	h := &UserHandler{}
	ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/admin/users")

	h.ListUsers(ctx)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestGetUserAuditSummary_Unauthorized(t *testing.T) {
	h := &UserHandler{}
	ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/admin/users/abc/audit-summary")
	ctx.Params = gin.Params{{Key: "userId", Value: "abc"}}

	h.GetUserAuditSummary(ctx)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestCreateImpersonationToken_Unauthorized(t *testing.T) {
	h := &UserHandler{}
	ctx, rec := newJSONTestContext(http.MethodPost, "/api/v1/admin/users/abc/impersonation-token")
	ctx.Params = gin.Params{{Key: "userId", Value: "abc"}}

	h.CreateImpersonationToken(ctx)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestSetMeAvailability_Unauthorized(t *testing.T) {
	h := &UserHandler{}
	ctx, rec := newJSONTestContext(http.MethodPut, "/api/v1/users/me/availability")

	h.SetMeAvailability(ctx)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestListUsers_InvalidPageSize_ReturnsBadRequest(t *testing.T) {
	h := &UserHandler{}
	ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/admin/users?page_size=101")
	ctx.Set(middleware.ContextUserID, "admin-user-id")

	h.ListUsers(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestParsePagination_ValidInput(t *testing.T) {
	ctx, _ := newJSONTestContext(http.MethodGet, "/api/v1/admin/users?page_size=25&page_token=next-25")

	pageSize, pageToken, err := parsePagination(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pageSize != 25 {
		t.Fatalf("expected page size 25, got %d", pageSize)
	}
	if pageToken != "next-25" {
		t.Fatalf("expected page token next-25, got %q", pageToken)
	}
}

func TestParsePagination_Defaults(t *testing.T) {
	ctx, _ := newJSONTestContext(http.MethodGet, "/api/v1/users/me/saved-freelancers")

	pageSize, pageToken, err := parsePagination(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if pageSize != 20 {
		t.Fatalf("expected default page size 20, got %d", pageSize)
	}
	if pageToken != "" {
		t.Fatalf("expected empty page token, got %q", pageToken)
	}
}

func TestParsePagination_InvalidInteger(t *testing.T) {
	ctx, _ := newJSONTestContext(http.MethodGet, "/api/v1/admin/users?page_size=nope")

	_, _, err := parsePagination(ctx)
	if err == nil {
		t.Fatalf("expected error for invalid page_size")
	}
}
