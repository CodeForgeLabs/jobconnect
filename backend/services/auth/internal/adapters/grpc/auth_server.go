package grpcadapter

import (
	"context"
	"fmt"
	"strings"
	"time"

	authv1 "jobconnect/auth/gen/auth/v1"
	"jobconnect/auth/internal/application"

	"github.com/google/uuid"

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
	ForgotPasswordUC   *application.ForgotPassword
	ResetPasswordUC    *application.ResetPassword
	ListSessionsUC     *application.ListSessions
	RevokeSessionUC    *application.RevokeSession
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

func (s *AuthServer) ForgotPassword(ctx context.Context, req *authv1.ForgotPasswordRequest) (*authv1.ForgotPasswordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	if s.ForgotPasswordUC == nil {
		return nil, status.Error(codes.Unimplemented, "forgot password is not configured")
	}
	out, err := s.ForgotPasswordUC.Execute(ctx, application.ForgotPasswordInput{Email: req.Email})
	if err != nil {
		return nil, toStatus(err)
	}
	return &authv1.ForgotPasswordResponse{Accepted: out.Accepted}, nil
}

func (s *AuthServer) ResetPassword(ctx context.Context, req *authv1.ResetPasswordRequest) (*authv1.ResetPasswordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	if s.ResetPasswordUC == nil {
		return nil, status.Error(codes.Unimplemented, "reset password is not configured")
	}
	err := s.ResetPasswordUC.Execute(ctx, application.ResetPasswordInput{
		Email:       req.Email,
		OTP:         req.Otp,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &authv1.ResetPasswordResponse{Ok: true}, nil
}

func (s *AuthServer) ListSessions(ctx context.Context, req *authv1.ListSessionsRequest) (*authv1.ListSessionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	if s.ListSessionsUC == nil {
		return nil, status.Error(codes.Unimplemented, "list sessions is not configured")
	}
	userID, err := uuid.Parse(strings.TrimSpace(req.UserId))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	out, err := s.ListSessionsUC.Execute(ctx, application.ListSessionsInput{UserID: userID})
	if err != nil {
		return nil, toStatus(err)
	}

	resp := &authv1.ListSessionsResponse{Sessions: make([]*authv1.Session, 0, len(out.Sessions))}
	for _, sess := range out.Sessions {
		lastUsedAt := ""
		if sess.LastUsedAt != nil {
			lastUsedAt = sess.LastUsedAt.UTC().Format(time.RFC3339)
		}
		resp.Sessions = append(resp.Sessions, &authv1.Session{
			SessionId:  sess.ID.String(),
			CreatedAt:  sess.CreatedAt.UTC().Format(time.RFC3339),
			ExpiresAt:  sess.ExpiresAt.UTC().Format(time.RFC3339),
			LastUsedAt: lastUsedAt,
		})
	}
	return resp, nil
}

func (s *AuthServer) RevokeSession(ctx context.Context, req *authv1.RevokeSessionRequest) (*authv1.RevokeSessionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	if s.RevokeSessionUC == nil {
		return nil, status.Error(codes.Unimplemented, "revoke session is not configured")
	}
	userID, err := uuid.Parse(strings.TrimSpace(req.UserId))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	sessionID, err := uuid.Parse(strings.TrimSpace(req.SessionId))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid session_id")
	}
	err = s.RevokeSessionUC.Execute(ctx, application.RevokeSessionInput{UserID: userID, SessionID: sessionID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &authv1.RevokeSessionResponse{Ok: true}, nil
}

func toStatus(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case contains(msg, "already registered"), contains(msg, "email already"):
		return status.Error(codes.AlreadyExists, msg)
	case contains(msg, "invalid email"), contains(msg, "password"), contains(msg, "display name"), contains(msg, "first name"), contains(msg, "last name"), contains(msg, "role"), contains(msg, "terms"), contains(msg, "refresh token required"), contains(msg, "otp is required"), contains(msg, "invalid user_id"), contains(msg, "invalid session_id"):
		return status.Error(codes.InvalidArgument, msg)
	case contains(msg, "invalid refresh token"), contains(msg, "refresh token expired"), contains(msg, "session revoked"), contains(msg, "invalid email or password"), contains(msg, "invalid reset credentials"), contains(msg, "invalid or expired otp"):
		return status.Error(codes.Unauthenticated, msg)
	case contains(msg, "forbidden session access"):
		return status.Error(codes.PermissionDenied, msg)
	case contains(msg, "session not found"):
		return status.Error(codes.NotFound, msg)
	default:
		return status.Error(codes.Internal, fmt.Sprintf("internal error: %v", err))
	}
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
