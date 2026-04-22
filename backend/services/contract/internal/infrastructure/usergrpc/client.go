package usergrpc

import (
	"context"
	"fmt"

	userv1 "jobconnect/contract/gen/user"
	"jobconnect/contract/internal/application"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Client struct {
	client userv1.UserServiceClient
}

func NewClient(client userv1.UserServiceClient) *Client {
	return &Client{client: client}
}

func (c *Client) EnsureClientCanHire(ctx context.Context, userID uuid.UUID) error {
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

func (c *Client) EnsureFreelancerCanWork(ctx context.Context, userID uuid.UUID) error {
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

func (c *Client) getProfile(ctx context.Context, userID uuid.UUID) (*userv1.UserProfile, error) {
	if c == nil || c.client == nil {
		return nil, fmt.Errorf("user client dependencies are not configured")
	}
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user_id is required")
	}
	res, err := c.client.GetMyProfile(ctx, &userv1.GetMyProfileRequest{UserId: userID.String()})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound || st.Code() == codes.InvalidArgument || st.Code() == codes.PermissionDenied {
				return nil, err
			}
		}
		return nil, fmt.Errorf("user service: %w", err)
	}
	if res.GetProfile() == nil || res.GetProfile().GetCore() == nil {
		return nil, fmt.Errorf("user profile not found")
	}
	return res.GetProfile(), nil
}

var _ application.ActorPolicy = (*Client)(nil)
