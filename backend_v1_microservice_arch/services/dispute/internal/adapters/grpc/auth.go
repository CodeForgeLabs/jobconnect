package grpcadapter

import (
	"context"
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
