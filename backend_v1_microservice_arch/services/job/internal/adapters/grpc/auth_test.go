package grpcadapter

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type fakeParser struct {
	userID uuid.UUID
	role   string
	err    error
}

func (f fakeParser) ParseAccessToken(token string) (uuid.UUID, string, error) {
	return f.userID, f.role, f.err
}

func TestCallerFromContextSuccess(t *testing.T) {
	uid := uuid.New()
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer abc"))
	outID, role, err := callerFromContext(ctx, fakeParser{userID: uid, role: "client"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if outID != uid {
		t.Fatalf("unexpected user id: %v", outID)
	}
	if role != "client" {
		t.Fatalf("unexpected role: %s", role)
	}
}

func TestRequireClientRole(t *testing.T) {
	if err := requireClientRole("client"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	err := requireClientRole("freelancer")
	if err == nil {
		t.Fatal("expected permission denied")
	}
	s, ok := status.FromError(err)
	if !ok || s.Code() != codes.PermissionDenied {
		t.Fatalf("expected permission denied, got %v", err)
	}
}
