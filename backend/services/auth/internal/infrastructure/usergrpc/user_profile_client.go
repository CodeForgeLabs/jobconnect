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

	req := &userv1.CreateProfileRequest{
		UserId:      in.UserID.String(),
		Role:        in.Role,
		FirstName:   in.FirstName,
		LastName:    in.LastName,
		DisplayName: in.DisplayName,
		AvatarUrl:   in.AvatarURL,
	}

	switch in.Role {
	case roleClient:
		req.RoleDetails = &userv1.CreateProfileRequest_Client{
			Client: &userv1.ClientProfileInput{
				VerificationStatus: userv1.VerificationStatus_VERIFICATION_STATUS_PENDING,
			},
		}
	case roleFreelancer:
		req.RoleDetails = &userv1.CreateProfileRequest_Freelancer{
			Freelancer: &userv1.FreelancerProfileInput{
				VerificationStatus: userv1.VerificationStatus_VERIFICATION_STATUS_PENDING,
			},
		}
	}

	resp, err := c.client.CreateProfile(ctx, req)
	if err != nil {
		return fmt.Errorf("create profile: %w", err)
	}
	if resp == nil || !resp.Success {
		return fmt.Errorf("create profile failed")
	}

	contactEmail := strings.TrimSpace(in.Email)
	if contactEmail != "" {
		_, err = c.client.UpdateProfile(ctx, &userv1.UpdateProfileRequest{
			UserId:       in.UserID.String(),
			ContactEmail: &contactEmail,
		})
		if err != nil {
			return fmt.Errorf("autofill contact email: %w", err)
		}
	}

	return nil
}
