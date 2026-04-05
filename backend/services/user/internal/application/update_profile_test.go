package application

import (
	"context"
	"testing"
	"time"

	"jobconnect/user/internal/domain"

	"github.com/google/uuid"
)

type updateProfileRepoMock struct {
	profile        domain.Profile
	client         *domain.ClientProfile
	freelancer     *domain.FreelancerProfile
	updatedProfile domain.Profile
	updatedClient  *domain.ClientProfile
	updatedFree    *domain.FreelancerProfile
	updateCalled   bool
}

func (m *updateProfileRepoMock) Create(context.Context, domain.Profile, *domain.ClientProfile, *domain.FreelancerProfile) (int64, error) {
	panic("not implemented")
}

func (m *updateProfileRepoMock) GetByUserID(context.Context, uuid.UUID) (domain.Profile, *domain.ClientProfile, *domain.FreelancerProfile, error) {
	return m.profile, m.client, m.freelancer, nil
}

func (m *updateProfileRepoMock) Update(_ context.Context, profile domain.Profile, client *domain.ClientProfile, freelancer *domain.FreelancerProfile) error {
	m.updateCalled = true
	m.updatedProfile = profile
	m.updatedClient = client
	m.updatedFree = freelancer
	return nil
}

func (m *updateProfileRepoMock) UpdateAccountState(context.Context, uuid.UUID, string, string, time.Time) (domain.Profile, *domain.ClientProfile, *domain.FreelancerProfile, error) {
	panic("not implemented")
}

func (m *updateProfileRepoMock) Delete(context.Context, uuid.UUID, bool, time.Time) error {
	panic("not implemented")
}

func (m *updateProfileRepoMock) SaveAvatar(context.Context, domain.Avatar) error {
	panic("not implemented")
}

func (m *updateProfileRepoMock) GetAvatar(context.Context, uuid.UUID) (domain.Avatar, error) {
	panic("not implemented")
}

func (m *updateProfileRepoMock) RemoveAvatar(context.Context, uuid.UUID) error {
	panic("not implemented")
}

func TestUpdateProfileFreelancerClearsSkillsWhenExplicitEmptySlice(t *testing.T) {
	userID := uuid.New()
	repo := &updateProfileRepoMock{
		profile: domain.Profile{UserID: userID, Role: domain.RoleFreelancer, DisplayName: "Jane", ContactEmail: "jane@example.com", AvatarURL: "avatar"},
		freelancer: &domain.FreelancerProfile{
			Headline: "Engineer",
			Skills:   []string{"go", "grpc"},
		},
	}
	uc := &UpdateProfile{Profiles: repo}

	_, err := uc.Execute(context.Background(), UpdateProfileInput{UserID: userID, Skills: []string{}})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !repo.updateCalled {
		t.Fatalf("expected Update to be called")
	}
	if repo.updatedFree == nil {
		t.Fatalf("expected freelancer profile to be updated")
	}
	if len(repo.updatedFree.Skills) != 0 {
		t.Fatalf("expected skills to be cleared, got %v", repo.updatedFree.Skills)
	}
}

func TestUpdateProfileClientRejectsFreelancerSkillsEvenWhenEmpty(t *testing.T) {
	userID := uuid.New()
	repo := &updateProfileRepoMock{
		profile: domain.Profile{UserID: userID, Role: domain.RoleClient, DisplayName: "Jane", ContactEmail: "jane@example.com", AvatarURL: "avatar"},
		client:  &domain.ClientProfile{CompanyName: "Acme"},
	}
	uc := &UpdateProfile{Profiles: repo}

	_, err := uc.Execute(context.Background(), UpdateProfileInput{UserID: userID, Skills: []string{}})
	if err == nil {
		t.Fatalf("expected role validation error")
	}
	if repo.updateCalled {
		t.Fatalf("did not expect Update to be called on validation failure")
	}
}
