package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"jobconnect/recommendation/internal/domain"
)

type fakeJobClient struct {
	recentJobs       []domain.JobData
	skillJobsBySkill map[string][]domain.JobData
	jobsByID         map[int64]domain.JobData
	getJobErr        error
	getJobHits       int
}

func (f fakeJobClient) ListRecentPublicOpenJobs(ctx context.Context, pageSize int32) ([]domain.JobData, error) {
	return append([]domain.JobData(nil), f.recentJobs...), nil
}

func (f fakeJobClient) SearchPublicOpenJobsBySkill(ctx context.Context, skill string, pageSize int32) ([]domain.JobData, error) {
	return append([]domain.JobData(nil), f.skillJobsBySkill[skill]...), nil
}

func (f *fakeJobClient) GetJob(ctx context.Context, jobID int64) (domain.JobData, error) {
	f.getJobHits++
	if f.getJobErr != nil {
		return domain.JobData{}, f.getJobErr
	}
	return f.jobsByID[jobID], nil
}

type fakeUserClient struct {
	user             domain.UserData
	preferences      domain.WorkPreferences
	discoverable     []domain.FreelancerData
	discoverableErr  error
	discoverableHits int
}

func (f *fakeUserClient) GetFreelancer(ctx context.Context, userID string) (domain.UserData, error) {
	return f.user, nil
}

func (f *fakeUserClient) GetWorkPreferences(ctx context.Context, userID string) (domain.WorkPreferences, error) {
	return f.preferences, nil
}

func (f *fakeUserClient) ListDiscoverableFreelancers(ctx context.Context, skills []string, pageSize int32) ([]domain.FreelancerData, error) {
	f.discoverableHits++
	if f.discoverableErr != nil {
		return nil, f.discoverableErr
	}
	return append([]domain.FreelancerData(nil), f.discoverable...), nil
}

type fakeReviewClient struct {
	summaries map[string]domain.RatingSummary
	err       error
	hits      map[string]int
}

func (f *fakeReviewClient) GetUserRatingSummary(ctx context.Context, userID string) (domain.RatingSummary, error) {
	if f.hits == nil {
		f.hits = make(map[string]int)
	}
	f.hits[userID]++
	if f.err != nil {
		return domain.RatingSummary{}, f.err
	}
	return f.summaries[userID], nil
}

type fakeCache struct {
	recommendations    []domain.JobRecommendation
	hit                bool
	setCalled          bool
	freelancerStore    map[string][]domain.FreelancerRecommendation
	freelancerSetCount int
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

func (f *fakeCache) GetRecommendedFreelancers(key string) ([]domain.FreelancerRecommendation, bool) {
	if f.freelancerStore == nil {
		return nil, false
	}
	recs, ok := f.freelancerStore[key]
	if !ok {
		return nil, false
	}
	return append([]domain.FreelancerRecommendation(nil), recs...), true
}

func (f *fakeCache) SetRecommendedFreelancers(key string, recs []domain.FreelancerRecommendation) {
	if f.freelancerStore == nil {
		f.freelancerStore = make(map[string][]domain.FreelancerRecommendation)
	}
	f.freelancerStore[key] = append([]domain.FreelancerRecommendation(nil), recs...)
	f.freelancerSetCount++
}

func TestGetRecommendedJobsRanksSkillAndSemanticMatchFirst(t *testing.T) {
	now := time.Now().UTC()
	svc := NewRecommendationService(
		&fakeJobClient{
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
		&fakeUserClient{
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
		&fakeJobClient{
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
		&fakeUserClient{
			user: domain.UserData{
				ID:           "freelancer-1",
				Headline:     "Go freelancer",
				Skills:       []string{"Go"},
				CanApplyJobs: true,
			},
			preferences: domain.WorkPreferences{MinBudgetUSD: 300},
		},
		nil,
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

func TestGetRecommendedJobsUsesClientTrustInRanking(t *testing.T) {
	now := time.Now().UTC()
	lowTrustJob := domain.JobData{
		ID:             101,
		ClientID:       "client-low",
		Title:          "Go API engineer",
		Description:    "Build backend APIs in Go",
		RequiredSkills: []string{"Go"},
		JobType:        "hourly",
		HourlyRate:     60,
		Visibility:     "public",
		CreatedAt:      now,
	}
	highTrustJob := lowTrustJob
	highTrustJob.ID = 202
	highTrustJob.ClientID = "client-high"

	reviews := &fakeReviewClient{
		summaries: map[string]domain.RatingSummary{
			"client-low":  {AverageRating: 2.0, TotalReviews: 10},
			"client-high": {AverageRating: 5.0, TotalReviews: 20},
		},
	}

	svc := NewRecommendationService(
		&fakeJobClient{recentJobs: []domain.JobData{lowTrustJob, highTrustJob}},
		&fakeUserClient{
			user: domain.UserData{
				ID:           "freelancer-1",
				Headline:     "Go API engineer",
				Bio:          "Backend APIs in Go",
				Skills:       []string{"Go"},
				HourlyRate:   60,
				CanApplyJobs: true,
			},
		},
		reviews,
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
	if recs[0].JobID != highTrustJob.ID {
		t.Fatalf("expected high-trust client job first, got %d", recs[0].JobID)
	}
	if reviews.hits["client-high"] != 1 || reviews.hits["client-low"] != 1 {
		t.Fatalf("expected one review lookup per client, got %#v", reviews.hits)
	}
}

func TestGetRecommendedJobsTreatsNoReviewsAsNeutral(t *testing.T) {
	now := time.Now().UTC()
	firstJob := domain.JobData{
		ID:             101,
		ClientID:       "client-a",
		Title:          "Go API engineer",
		Description:    "Build backend APIs in Go",
		RequiredSkills: []string{"Go"},
		JobType:        "hourly",
		HourlyRate:     60,
		Visibility:     "public",
		CreatedAt:      now,
	}
	secondJob := firstJob
	secondJob.ID = 202
	secondJob.ClientID = "client-b"

	svc := NewRecommendationService(
		&fakeJobClient{recentJobs: []domain.JobData{secondJob, firstJob}},
		&fakeUserClient{
			user: domain.UserData{
				ID:           "freelancer-1",
				Headline:     "Go API engineer",
				Bio:          "Backend APIs in Go",
				Skills:       []string{"Go"},
				HourlyRate:   60,
				CanApplyJobs: true,
			},
		},
		&fakeReviewClient{summaries: map[string]domain.RatingSummary{}},
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
	if recs[0].JobID != firstJob.ID {
		t.Fatalf("expected no-review jobs to remain tie-broken by id, got %d", recs[0].JobID)
	}
}

func TestGetRecommendedJobsDegradesWhenReviewLookupFails(t *testing.T) {
	now := time.Now().UTC()
	svc := NewRecommendationService(
		&fakeJobClient{recentJobs: []domain.JobData{{
			ID:             101,
			ClientID:       "client-a",
			Title:          "Go API engineer",
			Description:    "Build backend APIs in Go",
			RequiredSkills: []string{"Go"},
			JobType:        "hourly",
			HourlyRate:     60,
			Visibility:     "public",
			CreatedAt:      now,
		}}},
		&fakeUserClient{
			user: domain.UserData{
				ID:           "freelancer-1",
				Headline:     "Go API engineer",
				Bio:          "Backend APIs in Go",
				Skills:       []string{"Go"},
				HourlyRate:   60,
				CanApplyJobs: true,
			},
		},
		&fakeReviewClient{err: errors.New("review unavailable")},
		nil,
		ServiceConfig{DefaultLimit: 10, MaxLimit: 25, CandidatePageSize: 20, PerSkillPageSize: 10, MaxSkillQueries: 3},
	)

	recs, err := svc.GetRecommendedJobs(context.Background(), "freelancer-1", 10)
	if err != nil {
		t.Fatalf("GetRecommendedJobs returned error: %v", err)
	}
	if len(recs) != 1 || recs[0].JobID != 101 {
		t.Fatalf("expected fallback recommendation for job 101, got %#v", recs)
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
		&fakeJobClient{},
		&fakeUserClient{},
		nil,
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

func newFreelancerTestConfig() ServiceConfig {
	return ServiceConfig{DefaultLimit: 10, MaxLimit: 25, CandidatePageSize: 20, PerSkillPageSize: 10, MaxSkillQueries: 3}
}

func TestGetRecommendedFreelancersRejectsInvalidJobID(t *testing.T) {
	svc := NewRecommendationService(&fakeJobClient{}, &fakeUserClient{}, nil, nil, newFreelancerTestConfig())
	if _, err := svc.GetRecommendedFreelancers(context.Background(), 0, 5, "caller-a"); err == nil {
		t.Fatal("expected error for zero job id")
	}
	if _, err := svc.GetRecommendedFreelancers(context.Background(), -1, 5, "caller-a"); err == nil {
		t.Fatal("expected error for negative job id")
	}
}

func TestGetRecommendedFreelancersJobFetchError(t *testing.T) {
	job := &fakeJobClient{getJobErr: errors.New("boom")}
	svc := NewRecommendationService(job, &fakeUserClient{}, nil, nil, newFreelancerTestConfig())
	if _, err := svc.GetRecommendedFreelancers(context.Background(), 7, 5, "caller-a"); err == nil {
		t.Fatal("expected propagated fetch error")
	}
}

func TestGetRecommendedFreelancersJobNotFound(t *testing.T) {
	job := &fakeJobClient{jobsByID: map[int64]domain.JobData{}}
	svc := NewRecommendationService(job, &fakeUserClient{}, nil, nil, newFreelancerTestConfig())
	if _, err := svc.GetRecommendedFreelancers(context.Background(), 42, 5, "caller-a"); err == nil {
		t.Fatal("expected not-found error for missing job")
	}
}

func TestGetRecommendedFreelancersRanksAndFilters(t *testing.T) {
	jobID := int64(100)
	job := &fakeJobClient{
		jobsByID: map[int64]domain.JobData{
			jobID: {
				ID:             jobID,
				Title:          "Senior Go Backend Engineer",
				Description:    "Build scalable microservices with Go and PostgreSQL",
				RequiredSkills: []string{"Go", "PostgreSQL", "gRPC"},
				HourlyRate:     80,
				JobType:        "hourly",
				Visibility:     "public",
			},
		},
	}

	strongMatch := domain.FreelancerData{
		ID:           "f-strong",
		Headline:     "Senior Go backend engineer",
		Bio:          "Microservices in Go with PostgreSQL and gRPC",
		Skills:       []string{"Go", "PostgreSQL", "gRPC"},
		HourlyRate:   70,
		Availability: availabilityFullTime,
		Rating:       4.8,
		TotalReviews: 25,
	}
	weakMatch := domain.FreelancerData{
		ID:           "f-weak",
		Headline:     "Graphic designer",
		Bio:          "Logo and brand design",
		Skills:       []string{"Illustrator", "Photoshop"},
		HourlyRate:   40,
		Availability: availabilityFullTime,
		Rating:       4.0,
	}
	unavailable := domain.FreelancerData{
		ID:           "f-unavailable",
		Headline:     "Go backend engineer",
		Skills:       []string{"Go", "gRPC"},
		HourlyRate:   60,
		Availability: availabilityUnavailable,
		Rating:       4.5,
	}
	rateMismatch := domain.FreelancerData{
		ID:           "f-overpriced",
		Headline:     "Go engineer",
		Skills:       []string{"Go", "PostgreSQL"},
		HourlyRate:   400,
		Availability: availabilityFullTime,
		Rating:       4.5,
	}

	user := &fakeUserClient{
		discoverable: []domain.FreelancerData{strongMatch, weakMatch, unavailable, rateMismatch},
	}
	cache := &fakeCache{}
	svc := NewRecommendationService(job, user, nil, cache, newFreelancerTestConfig())

	recs, err := svc.GetRecommendedFreelancers(context.Background(), jobID, 5, "caller-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) == 0 {
		t.Fatal("expected at least one recommendation")
	}
	if recs[0].UserID != strongMatch.ID {
		t.Fatalf("expected strong match first, got %q", recs[0].UserID)
	}
	for _, r := range recs {
		if r.UserID == unavailable.ID {
			t.Fatalf("unavailable freelancer should be filtered out, got %q", r.UserID)
		}
		if r.UserID == rateMismatch.ID {
			t.Fatalf("over-budget freelancer should be filtered out, got %q", r.UserID)
		}
	}

	if cache.freelancerSetCount != 1 {
		t.Fatalf("expected 1 cache write, got %d", cache.freelancerSetCount)
	}

	if _, err := svc.GetRecommendedFreelancers(context.Background(), jobID, 5, "caller-a"); err != nil {
		t.Fatalf("cached call errored: %v", err)
	}
	if cache.freelancerSetCount != 1 {
		t.Fatalf("cache should be reused on second call, writes=%d", cache.freelancerSetCount)
	}
	if user.discoverableHits != 1 {
		t.Fatalf("expected single user-service lookup on cache hit, got %d", user.discoverableHits)
	}
	if job.getJobHits != 2 {
		t.Fatalf("expected job authorization check before cache hit, got %d job lookups", job.getJobHits)
	}
}

func TestGetRecommendedFreelancersUsesFreelancerTrustInRanking(t *testing.T) {
	jobID := int64(200)
	job := &fakeJobClient{
		jobsByID: map[int64]domain.JobData{
			jobID: {
				ID:             jobID,
				Title:          "Go API Engineer",
				Description:    "Build backend APIs in Go",
				RequiredSkills: []string{"Go"},
				HourlyRate:     80,
				JobType:        "hourly",
				Visibility:     "public",
			},
		},
	}
	lowTrust := domain.FreelancerData{
		ID:           "f-aaa",
		Headline:     "Go API engineer",
		Bio:          "Backend APIs in Go",
		Skills:       []string{"Go"},
		HourlyRate:   70,
		Availability: availabilityFullTime,
	}
	highTrust := lowTrust
	highTrust.ID = "f-zzz"

	reviews := &fakeReviewClient{
		summaries: map[string]domain.RatingSummary{
			lowTrust.ID:  {AverageRating: 2.0, TotalReviews: 8},
			highTrust.ID: {AverageRating: 5.0, TotalReviews: 14},
		},
	}
	user := &fakeUserClient{discoverable: []domain.FreelancerData{lowTrust, highTrust}}
	svc := NewRecommendationService(job, user, reviews, nil, newFreelancerTestConfig())

	recs, err := svc.GetRecommendedFreelancers(context.Background(), jobID, 5, "caller-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("expected 2 recommendations, got %d", len(recs))
	}
	if recs[0].UserID != highTrust.ID {
		t.Fatalf("expected high-trust freelancer first, got %q", recs[0].UserID)
	}
	if reviews.hits[highTrust.ID] != 1 || reviews.hits[lowTrust.ID] != 1 {
		t.Fatalf("expected one review lookup per freelancer, got %#v", reviews.hits)
	}
}

func TestGetRecommendedFreelancersDegradesWhenReviewLookupFails(t *testing.T) {
	jobID := int64(201)
	job := &fakeJobClient{
		jobsByID: map[int64]domain.JobData{
			jobID: {
				ID:             jobID,
				Title:          "Go API Engineer",
				Description:    "Build backend APIs in Go",
				RequiredSkills: []string{"Go"},
				HourlyRate:     80,
				JobType:        "hourly",
				Visibility:     "public",
			},
		},
	}
	reviews := &fakeReviewClient{err: errors.New("review unavailable")}
	user := &fakeUserClient{discoverable: []domain.FreelancerData{{
		ID:           "f-go",
		Headline:     "Go API engineer",
		Bio:          "Backend APIs in Go",
		Skills:       []string{"Go"},
		HourlyRate:   70,
		Availability: availabilityFullTime,
	}}}
	svc := NewRecommendationService(job, user, reviews, nil, newFreelancerTestConfig())

	recs, err := svc.GetRecommendedFreelancers(context.Background(), jobID, 5, "caller-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) != 1 || recs[0].UserID != "f-go" {
		t.Fatalf("expected fallback recommendation for f-go, got %#v", recs)
	}
	if reviews.hits["f-go"] != 1 {
		t.Fatalf("expected one review lookup, got %#v", reviews.hits)
	}
}

func TestGetRecommendedFreelancersRespectsLimit(t *testing.T) {
	jobID := int64(77)
	job := &fakeJobClient{
		jobsByID: map[int64]domain.JobData{
			jobID: {
				ID:             jobID,
				Title:          "Go Engineer",
				RequiredSkills: []string{"Go"},
				JobType:        "hourly",
				HourlyRate:     80,
				Visibility:     "public",
			},
		},
	}
	user := &fakeUserClient{}
	for i := 0; i < 8; i++ {
		user.discoverable = append(user.discoverable, domain.FreelancerData{
			ID:           string(rune('a' + i)),
			Headline:     "Go engineer",
			Skills:       []string{"Go"},
			HourlyRate:   70,
			Availability: availabilityFullTime,
			Rating:       4.5,
		})
	}

	svc := NewRecommendationService(job, user, nil, nil, newFreelancerTestConfig())
	recs, err := svc.GetRecommendedFreelancers(context.Background(), jobID, 3, "caller-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) != 3 {
		t.Fatalf("expected 3 recommendations, got %d", len(recs))
	}
}

func TestGetRecommendedFreelancersCacheIsCallerScoped(t *testing.T) {
	jobID := int64(55)
	job := &fakeJobClient{
		jobsByID: map[int64]domain.JobData{
			jobID: {
				ID:             jobID,
				Title:          "Go Engineer",
				RequiredSkills: []string{"Go"},
				JobType:        "hourly",
				HourlyRate:     80,
				Visibility:     "public",
			},
		},
	}
	user := &fakeUserClient{
		discoverable: []domain.FreelancerData{{
			ID:           "f-go",
			Headline:     "Go engineer",
			Skills:       []string{"Go"},
			HourlyRate:   70,
			Availability: availabilityFullTime,
			Rating:       4.5,
		}},
	}
	cache := &fakeCache{}
	svc := NewRecommendationService(job, user, nil, cache, newFreelancerTestConfig())

	if _, err := svc.GetRecommendedFreelancers(context.Background(), jobID, 5, "caller-a"); err != nil {
		t.Fatalf("caller-a recommendations errored: %v", err)
	}
	if _, err := svc.GetRecommendedFreelancers(context.Background(), jobID, 5, "caller-b"); err != nil {
		t.Fatalf("caller-b recommendations errored: %v", err)
	}

	if cache.freelancerSetCount != 2 {
		t.Fatalf("expected separate cache writes per caller, got %d", cache.freelancerSetCount)
	}
	if user.discoverableHits != 2 {
		t.Fatalf("expected separate user lookups per caller, got %d", user.discoverableHits)
	}
}
