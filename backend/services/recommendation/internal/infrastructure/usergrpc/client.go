package usergrpc

import (
	"context"

	"google.golang.org/grpc"

	"jobconnect/recommendation/internal/domain"
	userv1 "jobconnect/user/gen/user"
)

type Client struct {
	grpcClient userv1.UserServiceClient
}

func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{
		grpcClient: userv1.NewUserServiceClient(conn),
	}
}

func (c *Client) GetFreelancer(ctx context.Context, userID string) (domain.UserData, error) {
	resp, err := c.grpcClient.GetProfile(ctx, &userv1.GetProfileRequest{
		UserId: userID,
	})
	if err != nil {
		return domain.UserData{}, err
	}

	var skills []string
	var rating float32
	if resp.Profile.Freelancer != nil {
		skills = resp.Profile.Freelancer.Skills
		rating = float32(resp.Profile.Freelancer.Rating)
	}

	return domain.UserData{
		ID:     resp.Profile.UserId,
		Skills: skills,
		Rating: rating,
	}, nil
}

func (c *Client) GetFreelancers(ctx context.Context) ([]domain.UserData, error) {
	resp, err := c.grpcClient.ListProfiles(ctx, &userv1.ListProfilesRequest{
		Role:     "freelancer",
		PageSize: 100, // Phase 1 default limit
	})
	if err != nil {
		return nil, err
	}

	var users []domain.UserData
	for _, p := range resp.Profiles {
		var skills []string
		var rating float32
		if p.Freelancer != nil {
			skills = p.Freelancer.Skills
			rating = float32(p.Freelancer.Rating)
		}
		users = append(users, domain.UserData{
			ID:     p.UserId,
			Skills: skills,
			Rating: rating,
		})
	}
	return users, nil
}
