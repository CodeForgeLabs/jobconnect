package usergrpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"jobconnect/recommendation/internal/domain"
	userv1 "jobconnect/user/gen/user"
)

type Client struct {
	grpcClient userv1.UserServiceClient
}

func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{grpcClient: userv1.NewUserServiceClient(conn)}
}

func (c *Client) GetFreelancer(ctx context.Context, userID string) (domain.UserData, error) {
	resp, err := c.grpcClient.GetMyProfile(ctx, &userv1.GetMyProfileRequest{UserId: userID})
	if err != nil {
		return domain.UserData{}, err
	}
	if resp.GetProfile() == nil || resp.GetProfile().GetFreelancer() == nil {
		return domain.UserData{}, fmt.Errorf("freelancer profile not found")
	}

	freelancer := resp.GetProfile().GetFreelancer()
	return domain.UserData{
		ID:           resp.GetProfile().GetCore().GetUserId(),
		Headline:     freelancer.GetHeadline(),
		Bio:          resp.GetProfile().GetCore().GetBio(),
		Skills:       append([]string(nil), freelancer.GetSkills()...),
		HourlyRate:   freelancer.GetHourlyRate(),
		Availability: freelancer.GetAvailability().String(),
		Rating:       freelancer.GetMetrics().GetRating(),
		CanApplyJobs: resp.GetProfile().GetCapabilities().GetCanApplyJobs(),
	}, nil
}

func (c *Client) GetWorkPreferences(ctx context.Context, userID string) (domain.WorkPreferences, error) {
	resp, err := c.grpcClient.GetMyWorkPreferences(ctx, &userv1.GetMyWorkPreferencesRequest{UserId: userID})
	if err != nil {
		return domain.WorkPreferences{}, err
	}

	settings := resp.GetSettings()
	return domain.WorkPreferences{
		PreferredProjectLength: settings.GetPreferredProjectLength().String(),
		MinBudgetUSD:           settings.GetMinBudget(),
		MaxBudgetUSD:           settings.GetMaxBudget(),
		ContractTypes:          append([]string(nil), settings.GetContractTypes()...),
		WeeklyCapacityHours:    settings.GetWeeklyCapacityHours(),
	}, nil
}
