package grpcadapter

import (
	"context"
	"fmt"
	"strings"

	authv1 "jobconnect/auth/gen/auth/v1"
	"jobconnect/auth/internal/application"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServer implements authv1.AuthServiceServer by delegating to use-cases.
type AuthServer struct {
	authv1.UnimplementedAuthServiceServer

	RegisterUC         *application.RegisterUser
	VerifyEmailOTPUC   *application.VerifyEmailOTP
	LoginUC            *application.Login
	RefreshUC          *application.Refresh
	LogoutEverywhereUC *application.Logout
}

// NewAuthServer returns an AuthServer with the given use-cases.
func NewAuthServer(
	register *application.RegisterUser,
	verifyOTP *application.VerifyEmailOTP,
	login *application.Login,
	refresh *application.Refresh,
	logout *application.Logout,
) *AuthServer {
	return &AuthServer{
		RegisterUC:         register,
		VerifyEmailOTPUC:   verifyOTP,
		LoginUC:            login,
		RefreshUC:          refresh,
		LogoutEverywhereUC: logout,
	}
}

func (s *AuthServer) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	out, err := s.RegisterUC.Execute(ctx, application.RegisterUserInput{
		Email:       req.Email,
		Password:    req.Password,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Role:        req.Role,
		AcceptTerms: req.AcceptTerms,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &authv1.RegisterResponse{
		UserId:  out.UserID.String(),
		OtpSent: out.OTPSent,
	}, nil
}

func (s *AuthServer) VerifyEmailOTP(ctx context.Context, req *authv1.VerifyEmailOTPRequest) (*authv1.VerifyEmailOTPResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	verified, err := s.VerifyEmailOTPUC.Execute(ctx, application.VerifyEmailOTPInput{
		Email: req.Email,
		OTP:   req.Otp,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &authv1.VerifyEmailOTPResponse{Verified: verified}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	out, err := s.LoginUC.Execute(ctx, application.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &authv1.LoginResponse{
		AccessToken:                 out.AccessToken,
		RefreshToken:                out.RefreshToken,
		AccessTokenExpiresInSeconds: out.ExpiresInSec,
	}, nil
}

func (s *AuthServer) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	out, err := s.RefreshUC.Execute(ctx, application.RefreshInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &authv1.RefreshResponse{
		AccessToken:                 out.AccessToken,
		RefreshToken:                out.RefreshToken,
		AccessTokenExpiresInSeconds: out.ExpiresInSec,
	}, nil
}

func (s *AuthServer) LogoutEverywhere(ctx context.Context, req *authv1.LogoutEverywhereRequest) (*authv1.LogoutEverywhereResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	err := s.LogoutEverywhereUC.Execute(ctx, application.LogoutInput{RefreshToken: req.RefreshToken})
	if err != nil {
		return nil, toStatus(err)
	}
	return &authv1.LogoutEverywhereResponse{Ok: true}, nil
}

func toStatus(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case contains(msg, "already registered"), contains(msg, "email already"):
		return status.Error(codes.AlreadyExists, msg)
	case contains(msg, "invalid email"), contains(msg, "password"), contains(msg, "display name"), contains(msg, "first name"), contains(msg, "last name"), contains(msg, "role"), contains(msg, "terms"), contains(msg, "refresh token required"):
		return status.Error(codes.InvalidArgument, msg)
	case contains(msg, "invalid refresh token"), contains(msg, "refresh token expired"), contains(msg, "session revoked"), contains(msg, "invalid email or password"):
		return status.Error(codes.Unauthenticated, msg)
	default:
		return status.Error(codes.Internal, fmt.Sprintf("internal error: %v", err))
	}
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
