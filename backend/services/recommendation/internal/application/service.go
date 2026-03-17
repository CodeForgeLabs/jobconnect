package application

import (
	"context"
	"fmt"
	"jobconnect/recommendation/internal/domain"
)

type RecommendationService struct {
	jobClient  JobServiceClient
	userClient UserServiceClient
}

func NewRecommendationService(jobClient JobServiceClient, userClient UserServiceClient) *RecommendationService {
	return &RecommendationService{
		jobClient:  jobClient,
		userClient: userClient,
	}
}

func (s *RecommendationService) GetRecommendedJobs(ctx context.Context, userID string, limit int32) ([]domain.JobRecommendation, error) {
	// 1. Fetch freelancer skills
	user, err := s.userClient.GetFreelancer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// 2. Fetch all open jobs
	jobs, err := s.jobClient.GetOpenJobs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch jobs: %w", err)
	}

	// 3. Score jobs based on skills overlap
	var recommendations []domain.JobRecommendation
	for _, job := range jobs {
		score, reason := calculateSkillMatch(user.Skills, job.RequiredSkills)
		if score > 0 {
			recommendations = append(recommendations, domain.JobRecommendation{
				JobID:       job.ID,
				MatchScore:  score,
				MatchReason: reason,
			})
		}
	}

	// 4. Sort and limit (Sorting omitted for Phase 1 skeleton)
	if int32(len(recommendations)) > limit && limit > 0 {
		recommendations = recommendations[:limit]
	}

	return recommendations, nil
}

func (s *RecommendationService) GetRecommendedFreelancers(ctx context.Context, jobID int64, limit int32) ([]domain.FreelancerRecommendation, error) {
	// 1. Fetch job required skills
	job, err := s.jobClient.GetJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch job: %w", err)
	}

	// 2. Fetch all freelancers
	freelancers, err := s.userClient.GetFreelancers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch freelancers: %w", err)
	}

	// 3. Score freelancers
	var recommendations []domain.FreelancerRecommendation
	for _, f := range freelancers {
		score, reason := calculateSkillMatch(job.RequiredSkills, f.Skills)
		if score > 0 {
			recommendations = append(recommendations, domain.FreelancerRecommendation{
				UserID:      f.ID,
				MatchScore:  score,
				MatchReason: reason,
			})
		}
	}

	if int32(len(recommendations)) > limit && limit > 0 {
		recommendations = recommendations[:limit]
	}

	return recommendations, nil
}

// calculateSkillMatch implements the basic heuristic for Phase 1
func calculateSkillMatch(targetSkills, providedSkills []string) (float32, string) {
	if len(targetSkills) == 0 || len(providedSkills) == 0 {
		return 0, ""
	}

	matchCount := 0
	providedMap := make(map[string]bool)
	for _, s := range providedSkills {
		providedMap[s] = true
	}

	for _, s := range targetSkills {
		if providedMap[s] {
			matchCount++
		}
	}

	if matchCount == 0 {
		return 0, ""
	}

	score := float32(matchCount) / float32(len(targetSkills))
	return score, fmt.Sprintf("Matches %d required skills", matchCount)
}
