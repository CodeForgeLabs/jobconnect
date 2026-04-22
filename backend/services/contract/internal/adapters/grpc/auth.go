package grpcadapter

import (
	"context"
	"os"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type TokenParser interface {
	ParseAccessToken(token string) (uuid.UUID, string, error)
}

func callerFromContext(ctx context.Context, parser TokenParser) (uuid.UUID, string, error) {
	if parser == nil {
		return uuid.Nil, "", status.Error(codes.Internal, "token parser is nil")
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, "", status.Error(codes.Unauthenticated, "missing metadata")
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return uuid.Nil, "", status.Error(codes.Unauthenticated, "missing authorization header")
	}
	bearer := strings.TrimSpace(vals[0])
	parts := strings.SplitN(bearer, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || strings.TrimSpace(parts[1]) == "" {
		return uuid.Nil, "", status.Error(codes.Unauthenticated, "invalid authorization header")
	}
	userID, role, err := parser.ParseAccessToken(strings.TrimSpace(parts[1]))
	if err != nil {
		return uuid.Nil, "", status.Error(codes.Unauthenticated, "invalid access token")
	}
	return userID, role, nil
}

func requireClientRole(role string) error {
	if strings.EqualFold(strings.TrimSpace(role), "client") {
		return nil
	}
	return status.Error(codes.PermissionDenied, "client role required")
}

func requireFreelancerRole(role string) error {
	if strings.EqualFold(strings.TrimSpace(role), "freelancer") {
		return nil
	}
	return status.Error(codes.PermissionDenied, "freelancer role required")
}

func requireInternalCaller(ctx context.Context, services ...string) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.PermissionDenied, "internal caller required")
	}
	vals := md.Get("x-jobconnect-internal")
	if len(vals) == 0 {
		return status.Error(codes.PermissionDenied, "internal caller required")
	}
	caller := strings.TrimSpace(vals[0])
	for _, service := range services {
		if strings.EqualFold(caller, service) {
			requiredSecret := strings.TrimSpace(os.Getenv("JOBCONNECT_INTERNAL_CALLER_SECRET"))
			if requiredSecret == "" {
				return nil
			}
			secretVals := md.Get("x-jobconnect-internal-secret")
			if len(secretVals) == 0 {
				return status.Error(codes.PermissionDenied, "internal caller secret required")
			}
			if strings.TrimSpace(secretVals[0]) != requiredSecret {
				return status.Error(codes.PermissionDenied, "invalid internal caller secret")
			}
			return nil
		}
	}
	return status.Error(codes.PermissionDenied, "internal caller required")
}
