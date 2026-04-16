package application

import (
	"context"
	"testing"
	"time"

	"jobconnect/recommendation/internal/domain"
)

type fakeJobClient struct {
	recentJobs       []domain.JobData
	skillJobsBySkill map[string][]domain.JobData
}

func (f fakeJobClient) ListRecentPublicOpenJobs(ctx context.Context, pageSize int32) ([]domain.JobData, error) {
	return append([]domain.JobData(nil), f.recentJobs...), nil
}

func (f fakeJobClient) SearchPublicOpenJobsBySkill(ctx context.Context, skill string, pageSize int32) ([]domain.JobData, error) {
	return append([]domain.JobData(nil), f.skillJobsBySkill[skill]...), nil
}

type fakeUserClient struct {
	user        domain.UserData
	preferences domain.WorkPreferences
}

func (f fakeUserClient) GetFreelancer(ctx context.Context, userID string) (domain.UserData, error) {
	return f.user, nil
}

func (f fakeUserClient) GetWorkPreferences(ctx context.Context, userID string) (domain.WorkPreferences, error) {
	return f.preferences, nil
}

type fakeCache struct {
	recommendations []domain.JobRecommendation
	hit             bool
	setCalled       bool
}

func (f *fakeCache) GetRecommendedJobs(userID string) ([]domain.JobRecommendation, bool) {
	if !f.hit {
		return nil, false
	}
	return append([]domain.JobRecommendation(nil), f.recommendations...), true
}

func (f *fakeCache) SetRecommendedJobs(userID string, recommendations []domain.JobRecommendation) {
	f.setCalled = true
	f.recommendations = append([]domain.JobRecommendation(nil), recommendations...)
}

func TestGetRecommendedJobsRanksSkillAndSemanticMatchFirst(t *testing.T) {
	now := time.Now().UTC()
	svc := NewRecommendationService(
		fakeJobClient{
			recentJobs: []domain.JobData{
				{
					ID:             101,
					Title:          "Senior Go backend engineer",
					Description:    "Build gRPC APIs and PostgreSQL services",
					RequiredSkills: []string{"Go", "gRPC", "PostgreSQL"},
					JobType:        "hourly",
					HourlyRate:     55,
					Visibility:     "public",
					CreatedAt:      now.Add(-2 * time.Hour),
				},
				{
					ID:             202,
					Title:          "Frontend React project",
					Description:    "Need a React specialist",
					RequiredSkills: []string{"React"},
					JobType:        "hourly",
					HourlyRate:     45,
					Visibility:     "public",
					CreatedAt:      now.Add(-1 * time.Hour),
				},
			},
			skillJobsBySkill: map[string][]domain.JobData{
				"Go": {
					{
						ID:             101,
						Title:          "Senior Go backend engineer",
						Description:    "Build gRPC APIs and PostgreSQL services",
						RequiredSkills: []string{"Go", "gRPC", "PostgreSQL"},
						JobType:        "hourly",
						HourlyRate:     55,
						Visibility:     "public",
						CreatedAt:      now.Add(-2 * time.Hour),
					},
				},
			},
		},
		fakeUserClient{
			user: domain.UserData{
				ID:           "freelancer-1",
				Headline:     "Backend Go engineer",
				Bio:          "I build PostgreSQL-backed gRPC APIs",
				Skills:       []string{"Go", "gRPC", "PostgreSQL"},
				HourlyRate:   50,
				CanApplyJobs: true,
			},
		},
		nil,
		ServiceConfig{DefaultLimit: 10, MaxLimit: 25, CandidatePageSize: 20, PerSkillPageSize: 10, MaxSkillQueries: 3},
	)

	recs, err := svc.GetRecommendedJobs(context.Background(), "freelancer-1", 10)
	if err != nil {
		t.Fatalf("GetRecommendedJobs returned error: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("expected 2 recommendations, got %d", len(recs))
	}
	if recs[0].JobID != 101 {
		t.Fatalf("expected best match job 101, got %d", recs[0].JobID)
	}
	if recs[0].MatchReason == "" {
		t.Fatal("expected match reason for top recommendation")
	}
}

func TestGetRecommendedJobsFiltersBudgetMismatch(t *testing.T) {
	now := time.Now().UTC()
	svc := NewRecommendationService(
		fakeJobClient{
			recentJobs: []domain.JobData{
				{
					ID:             101,
					Title:          "Tiny fixed-price API fix",
					Description:    "Small job",
					RequiredSkills: []string{"Go"},
					JobType:        "fixed",
					BudgetMin:      50,
					BudgetMax:      75,
					Visibility:     "public",
					CreatedAt:      now,
				},
				{
					ID:             202,
					Title:          "Well-funded backend project",
					Description:    "Medium sized backend build",
					RequiredSkills: []string{"Go"},
					JobType:        "fixed",
					BudgetMin:      500,
					BudgetMax:      900,
					Visibility:     "public",
					CreatedAt:      now,
				},
			},
		},
		fakeUserClient{
			user: domain.UserData{
				ID:           "freelancer-1",
				Headline:     "Go freelancer",
				Skills:       []string{"Go"},
				CanApplyJobs: true,
			},
			preferences: domain.WorkPreferences{MinBudgetUSD: 300},
		},
		nil,
		ServiceConfig{DefaultLimit: 10, MaxLimit: 25, CandidatePageSize: 20, PerSkillPageSize: 10, MaxSkillQueries: 3},
	)

	recs, err := svc.GetRecommendedJobs(context.Background(), "freelancer-1", 10)
	if err != nil {
		t.Fatalf("GetRecommendedJobs returned error: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 recommendation, got %d", len(recs))
	}
	if recs[0].JobID != 202 {
		t.Fatalf("expected budget-compatible job 202, got %d", recs[0].JobID)
	}
}

func TestGetRecommendedJobsUsesCacheWhenWarm(t *testing.T) {
	cache := &fakeCache{
		hit: true,
		recommendations: []domain.JobRecommendation{
			{JobID: 7, MatchScore: 0.9, MatchReason: "cached"},
			{JobID: 8, MatchScore: 0.8, MatchReason: "cached"},
		},
	}

	svc := NewRecommendationService(
		fakeJobClient{},
		fakeUserClient{},
		cache,
		ServiceConfig{DefaultLimit: 1, MaxLimit: 2, CandidatePageSize: 20, PerSkillPageSize: 10, MaxSkillQueries: 3},
	)

	recs, err := svc.GetRecommendedJobs(context.Background(), "freelancer-1", 1)
	if err != nil {
		t.Fatalf("GetRecommendedJobs returned error: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 cached recommendation, got %d", len(recs))
	}
	if recs[0].JobID != 7 {
		t.Fatalf("expected cached job 7, got %d", recs[0].JobID)
	}
	if cache.setCalled {
		t.Fatal("did not expect cache set on cache hit")
	}
}
