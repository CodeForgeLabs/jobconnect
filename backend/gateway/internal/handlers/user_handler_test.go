package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"jobconnect/gateway/internal/middleware"
	userv1 "jobconnect/user/gen/user"
	verificationv1 "jobconnect/verification/gen/verification/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func newJSONTestContext(method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(method, target, nil)
	return ctx, rec
}

func newJSONBodyTestContext(method, target, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest(method, target, bytes.NewBufferString(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	return ctx, rec
}

func TestListUsers_Unauthorized(t *testing.T) {
	h := &UserHandler{}
	ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/admin/users")

	h.ListUsers(ctx)

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

func TestUpdateMeProfile_Unauthorized(t *testing.T) {
	h := &UserHandler{}
	ctx, rec := newJSONTestContext(http.MethodPatch, "/api/v1/users/me/profile")

	h.UpdateMeProfile(ctx)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestUpdateMeProfile_MixedRolePayloadRejected(t *testing.T) {
	h := &UserHandler{}
	body := `{"company_name":"Acme","headline":"Builder"}`
	ctx, rec := newJSONBodyTestContext(http.MethodPatch, "/api/v1/users/me/profile", body)
	ctx.Set(middleware.ContextUserID, "11111111-1111-1111-1111-111111111111")

	h.UpdateMeProfile(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUpdateMeProfile_UnsupportedLegacyFieldsRejected(t *testing.T) {
	h := &UserHandler{}
	body := `{"first_name":"Jane"}`
	ctx, rec := newJSONBodyTestContext(http.MethodPatch, "/api/v1/users/me/profile", body)
	ctx.Set(middleware.ContextUserID, "11111111-1111-1111-1111-111111111111")

	h.UpdateMeProfile(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUpdateMeProfile_LanguageRejectedInFavorOfSettings(t *testing.T) {
	h := &UserHandler{}
	body := `{"language":"fr"}`
	ctx, rec := newJSONBodyTestContext(http.MethodPatch, "/api/v1/users/me/profile", body)
	ctx.Set(middleware.ContextUserID, "11111111-1111-1111-1111-111111111111")

	h.UpdateMeProfile(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

type submittedVerificationClientStub struct {
	called bool
}

func (s *submittedVerificationClientStub) GetMyVerificationStatus(ctx context.Context, in *verificationv1.GetMyVerificationStatusRequest, opts ...grpc.CallOption) (*verificationv1.GetMyVerificationStatusResponse, error) {
	s.called = true
	return &verificationv1.GetMyVerificationStatusResponse{
		Request: &verificationv1.VerificationRequest{Status: "submitted"},
	}, nil
}

func TestUpdateMeProfile_TaxIDRejectedWhenVerificationSubmitted(t *testing.T) {
	verificationClient := &submittedVerificationClientStub{}
	h := &UserHandler{verificationClient: verificationClient}
	body := `{"tax_id":"TIN-123"}`
	ctx, rec := newJSONBodyTestContext(http.MethodPatch, "/api/v1/users/me/profile", body)
	ctx.Set(middleware.ContextUserID, "11111111-1111-1111-1111-111111111111")

	h.UpdateMeProfile(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if !verificationClient.called {
		t.Fatalf("expected verification status check to be called")
	}
}

func TestGetMeAccountSettings_Unauthorized(t *testing.T) {
	h := &UserHandler{}
	ctx, rec := newJSONTestContext(http.MethodGet, "/api/v1/users/me/settings")

	h.GetMeAccountSettings(ctx)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestUpdateMeAccountSettings_RequiresPayload(t *testing.T) {
	h := &UserHandler{}
	body := `{}`
	ctx, rec := newJSONBodyTestContext(http.MethodPatch, "/api/v1/users/me/settings", body)
	ctx.Set(middleware.ContextUserID, "11111111-1111-1111-1111-111111111111")

	h.UpdateMeAccountSettings(ctx)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestProtoToAny_ProfileReadiness(t *testing.T) {
	msg := &userv1.ProfileReadiness{
		Percent:               80,
		MissingRequiredFields: []string{"hiring_preferences"},
		Recommendations:       []string{"Add hiring preferences to improve client matching"},
	}

	payload, err := protoToAny(msg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	obj, ok := payload.(map[string]any)
	if !ok {
		t.Fatalf("expected map payload, got %T", payload)
	}
	if got, ok := obj["percent"].(float64); !ok || got != 80 {
		t.Fatalf("expected percent=80, got %v", obj["percent"])
	}

	missing, ok := obj["missing_required_fields"].([]any)
	if !ok || len(missing) != 1 || missing[0] != "hiring_preferences" {
		t.Fatalf("expected missing_required_fields with hiring_preferences, got %#v", obj["missing_required_fields"])
	}

	recs, ok := obj["recommendations"].([]any)
	if !ok || len(recs) != 1 {
		t.Fatalf("expected recommendations array, got %#v", obj["recommendations"])
	}
}
