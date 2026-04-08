package handlers

import (
	"bytes"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"jobconnect/gateway/internal/middleware"
	userv1 "jobconnect/user/gen/user"
	verificationv1 "jobconnect/verification/gen/verification/v1"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
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

type upsertFreelancerNoteCaptureServer struct {
	userv1.UnimplementedUserServiceServer
	lastReq *userv1.UpsertFreelancerNoteRequest
}

func (s *upsertFreelancerNoteCaptureServer) UpsertFreelancerNote(ctx context.Context, in *userv1.UpsertFreelancerNoteRequest) (*userv1.UpsertFreelancerNoteResponse, error) {
	s.lastReq = in
	return &userv1.UpsertFreelancerNoteResponse{
		Note: &userv1.FreelancerNote{
			FreelancerUserId: in.GetFreelancerUserId(),
			Note:             in.GetNote(),
			UpdatedAtUnix:    1,
		},
	}, nil
}

func newBufConnUserClient(t *testing.T, srv userv1.UserServiceServer) (userv1.UserServiceClient, func()) {
	t.Helper()

	listener := bufconn.Listen(1024 * 1024)
	grpcServer := grpc.NewServer()
	userv1.RegisterUserServiceServer(grpcServer, srv)

	go func() {
		_ = grpcServer.Serve(listener)
	}()

	ctx := context.Background()
	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithInsecure(),
	)
	if err != nil {
		listener.Close()
		grpcServer.Stop()
		t.Fatalf("failed to dial bufconn grpc server: %v", err)
	}

	cleanup := func() {
		_ = conn.Close()
		grpcServer.Stop()
		_ = listener.Close()
		_ = ctx
	}

	return userv1.NewUserServiceClient(conn), cleanup
}

func TestUpsertMeFreelancerNote_IdentityFieldsCannotBeOverriddenByBody(t *testing.T) {
	srv := &upsertFreelancerNoteCaptureServer{}
	client, cleanup := newBufConnUserClient(t, srv)
	defer cleanup()

	h := NewUserHandler(client, nil)
	body := `{"user_id":"22222222-2222-2222-2222-222222222222","freelancer_user_id":"33333333-3333-3333-3333-333333333333","note":"hello"}`
	ctx, rec := newJSONBodyTestContext(http.MethodPut, "/api/v1/users/me/freelancer-notes/11111111-1111-1111-1111-111111111111", body)
	ctx.Set(middleware.ContextUserID, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	ctx.Params = gin.Params{{Key: "freelancerId", Value: "11111111-1111-1111-1111-111111111111"}}

	h.UpsertMeFreelancerNote(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if srv.lastReq == nil {
		t.Fatalf("expected grpc request to be captured")
	}
	if got := srv.lastReq.GetUserId(); got != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("expected trusted auth user_id, got %q", got)
	}
	if got := srv.lastReq.GetFreelancerUserId(); got != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("expected trusted path freelancer_user_id, got %q", got)
	}
	if got := srv.lastReq.GetNote(); got != "hello" {
		t.Fatalf("expected note to be preserved, got %q", got)
	}
}
