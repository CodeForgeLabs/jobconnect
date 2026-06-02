package usergrpc

import (
	"context"
	"fmt"
	"strings"

	"jobconnect/auth/internal/application"
	userv1 "jobconnect/user/gen/user"
)

const (
	roleClient     = "client"
	roleFreelancer = "freelancer"
)

// ProfileClient implements application.UserProfileService via gRPC.
type ProfileClient struct {
	client userv1.UserServiceClient
}

func NewProfileClient(client userv1.UserServiceClient) *ProfileClient {
	return &ProfileClient{client: client}
}

func (c *ProfileClient) CreateProfile(ctx context.Context, in application.CreateProfileInput) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("user profile client is nil")
	}

	role, err := toProtoRole(in.Role)
	if err != nil {
		return err
	}

	req := &userv1.CreateMyProfileRequest{
		UserId:       in.UserID.String(),
		Role:         role,
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		DisplayName:  in.DisplayName,
		ContactEmail: strings.TrimSpace(in.Email),
	}

	switch in.Role {
	case roleClient:
		req.RoleProfile = &userv1.CreateMyProfileRequest_Client{
			Client: &userv1.ClientProfileCreateInput{},
		}
	case roleFreelancer:
		req.RoleProfile = &userv1.CreateMyProfileRequest_Freelancer{
			Freelancer: &userv1.FreelancerProfileCreateInput{},
		}
	}

	resp, err := c.client.CreateMyProfile(ctx, req)
	if err != nil {
		return fmt.Errorf("create profile: %w", err)
	}
	if resp == nil || !resp.Success {
		return fmt.Errorf("create profile failed")
	}

	return nil
}

func toProtoRole(role string) (userv1.UserRole, error) {
	switch strings.TrimSpace(strings.ToLower(role)) {
	case roleClient:
		return userv1.UserRole_USER_ROLE_CLIENT, nil
	case roleFreelancer:
		return userv1.UserRole_USER_ROLE_FREELANCER, nil
	case "admin":
		return userv1.UserRole_USER_ROLE_ADMIN, nil
	default:
		return userv1.UserRole_USER_ROLE_UNSPECIFIED, fmt.Errorf("unsupported role: %s", role)
	}
}
