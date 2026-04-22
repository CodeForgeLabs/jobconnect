package clients

import (
	"context"
	"fmt"
	"log"

	userv1 "jobconnect/job/gen/user"
	"jobconnect/job/internal/application"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type UserClient struct {
	client userv1.UserServiceClient
}

func NewUserClient(address string) (*UserClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to user service at %s", address)
	return &UserClient{client: userv1.NewUserServiceClient(conn)}, nil
}

func (c *UserClient) EnsureClientCanHire(ctx context.Context, userID uuid.UUID) error {
	profile, err := c.getProfile(ctx, userID)
	if err != nil {
		return err
	}
	if profile.GetCore().GetRole() != userv1.UserRole_USER_ROLE_CLIENT {
		return fmt.Errorf("client role required")
	}
	if profile.GetCore().GetAccountStatus() == userv1.AccountStatus_ACCOUNT_STATUS_SUSPENDED ||
		profile.GetCore().GetAccountStatus() == userv1.AccountStatus_ACCOUNT_STATUS_DELETED {
		return fmt.Errorf("client account is not eligible to hire")
	}
	if profile.GetCapabilities() == nil || !profile.GetCapabilities().GetCanPostJobs() {
		return fmt.Errorf("client account cannot hire")
	}
	return nil
}

func (c *UserClient) EnsureFreelancerCanWork(ctx context.Context, userID uuid.UUID) error {
	profile, err := c.getProfile(ctx, userID)
	if err != nil {
		return err
	}
	if profile.GetCore().GetRole() != userv1.UserRole_USER_ROLE_FREELANCER {
		return fmt.Errorf("freelancer role required")
	}
	if profile.GetCore().GetAccountStatus() == userv1.AccountStatus_ACCOUNT_STATUS_SUSPENDED ||
		profile.GetCore().GetAccountStatus() == userv1.AccountStatus_ACCOUNT_STATUS_DELETED {
		return fmt.Errorf("freelancer account is not eligible to work")
	}
	if profile.GetCapabilities() == nil || !profile.GetCapabilities().GetCanApplyJobs() {
		return fmt.Errorf("freelancer account cannot accept work")
	}
	return nil
}

func (c *UserClient) getProfile(ctx context.Context, userID uuid.UUID) (*userv1.UserProfile, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("user client dependencies are not configured")
	}
	res, err := c.client.GetMyProfile(ctx, &userv1.GetMyProfileRequest{UserId: userID.String()})
	if err != nil {
		if _, ok := status.FromError(err); ok {
			return nil, err
		}
		return nil, fmt.Errorf("user service: %w", err)
	}
	if res.GetProfile() == nil || res.GetProfile().GetCore() == nil {
		return nil, fmt.Errorf("user profile not found")
	}
	return res.GetProfile(), nil
}

var _ application.ActorPolicy = (*UserClient)(nil)
