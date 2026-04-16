package application

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"jobconnect/recommendation/internal/domain"
)

const (
	publicVisibility = "public"
	jobTypeFixed     = "fixed"
	jobTypeHourly    = "hourly"
)

type ServiceConfig struct {
	DefaultLimit      int32
	MaxLimit          int32
	CandidatePageSize int32
	PerSkillPageSize  int32
	MaxSkillQueries   int
}

type RecommendationService struct {
	jobClient  JobServiceClient
	userClient UserServiceClient
	cache      RecommendationCache
	cfg        ServiceConfig
}

type scoredJobRecommendation struct {
	recommendation domain.JobRecommendation
	score          float64
	skillMatches   []string
	semanticScore  float64
	budgetScore    float64
	freshnessScore float64
}

func NewRecommendationService(
	jobClient JobServiceClient,
	userClient UserServiceClient,
	cache RecommendationCache,
	cfg ServiceConfig,
) *RecommendationService {
	if cfg.DefaultLimit <= 0 {
		cfg.DefaultLimit = 10
	}
	if cfg.MaxLimit < cfg.DefaultLimit {
		cfg.MaxLimit = cfg.DefaultLimit
	}
	if cfg.CandidatePageSize <= 0 {
		cfg.CandidatePageSize = 100
	}
	if cfg.PerSkillPageSize <= 0 {
		cfg.PerSkillPageSize = 25
	}
	if cfg.MaxSkillQueries <= 0 {
		cfg.MaxSkillQueries = 5
	}

	return &RecommendationService{
		jobClient:  jobClient,
		userClient: userClient,
		cache:      cache,
		cfg:        cfg,
	}
}

func (s *RecommendationService) GetRecommendedJobs(ctx context.Context, userID string, limit int32) ([]domain.JobRecommendation, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	normalizedLimit := s.normalizeLimit(limit)
	if s.cache != nil {
		if cached, ok := s.cache.GetRecommendedJobs(userID); ok {
			return limitRecommendations(cached, normalizedLimit), nil
		}
	}

	user, err := s.userClient.GetFreelancer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetch freelancer profile: %w", err)
	}
	if !user.CanApplyJobs {
		return nil, nil
	}

	preferences, err := s.userClient.GetWorkPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetch freelancer work preferences: %w", err)
	}

	candidates, err := s.collectCandidates(ctx, user, preferences)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	ranked := s.rankJobs(user, preferences, candidates)
	recommendations := make([]domain.JobRecommendation, 0, len(ranked))
	for _, rankedJob := range ranked {
		recommendations = append(recommendations, rankedJob.recommendation)
	}

	if s.cache != nil {
		s.cache.SetRecommendedJobs(userID, recommendations)
	}
	return limitRecommendations(recommendations, normalizedLimit), nil
}

func (s *RecommendationService) GetRecommendedFreelancers(ctx context.Context, jobID int64, limit int32) ([]domain.FreelancerRecommendation, error) {
	return nil, fmt.Errorf("freelancer recommendations are not implemented in phase 1")
}

func (s *RecommendationService) collectCandidates(ctx context.Context, user domain.UserData, preferences domain.WorkPreferences) ([]domain.JobData, error) {
	candidateMap := make(map[int64]domain.JobData)

	baseJobs, err := s.jobClient.ListRecentPublicOpenJobs(ctx, s.cfg.CandidatePageSize)
	if err != nil {
		return nil, fmt.Errorf("list recent public jobs: %w", err)
	}
	addEligibleJobs(candidateMap, baseJobs, user, preferences)

	for _, skill := range topSkills(user.Skills, s.cfg.MaxSkillQueries) {
		skillJobs, skillErr := s.jobClient.SearchPublicOpenJobsBySkill(ctx, skill, s.cfg.PerSkillPageSize)
		if skillErr != nil {
			log.Printf("recommendation: skill candidate query failed for %q: %v", skill, skillErr)
			continue
		}
		addEligibleJobs(candidateMap, skillJobs, user, preferences)
	}

	candidates := make([]domain.JobData, 0, len(candidateMap))
	for _, job := range candidateMap {
		candidates = append(candidates, job)
	}
	return candidates, nil
}

func addEligibleJobs(candidateMap map[int64]domain.JobData, jobs []domain.JobData, user domain.UserData, preferences domain.WorkPreferences) {
	for _, job := range jobs {
		if !isEligibleJob(job, user, preferences) {
			continue
		}
		candidateMap[job.ID] = job
	}
}

func isEligibleJob(job domain.JobData, user domain.UserData, preferences domain.WorkPreferences) bool {
	if job.ID == 0 || strings.TrimSpace(job.Visibility) != publicVisibility {
		return false
	}

	if len(preferences.ContractTypes) > 0 && !matchesContractType(job.JobType, preferences.ContractTypes) {
		return false
	}

	switch strings.TrimSpace(strings.ToLower(job.JobType)) {
	case jobTypeFixed:
		if preferences.MinBudgetUSD > 0 && job.BudgetMax > 0 && job.BudgetMax < preferences.MinBudgetUSD {
			return false
		}
		if preferences.MaxBudgetUSD > 0 && job.BudgetMin > 0 && job.BudgetMin > preferences.MaxBudgetUSD {
			return false
		}
	case jobTypeHourly:
		if user.HourlyRate > 0 && job.HourlyRate > 0 && job.HourlyRate < user.HourlyRate*0.5 {
			return false
		}
	}

	return true
}

func matchesContractType(jobType string, contractTypes []string) bool {
	jobType = strings.TrimSpace(strings.ToLower(jobType))
	for _, contractType := range contractTypes {
		if strings.TrimSpace(strings.ToLower(contractType)) == jobType {
			return true
		}
	}
	return false
}

func (s *RecommendationService) rankJobs(user domain.UserData, preferences domain.WorkPreferences, jobs []domain.JobData) []scoredJobRecommendation {
	profileText := strings.Join([]string{
		user.Headline,
		user.Bio,
		strings.Join(user.Skills, " "),
	}, " ")
	profileVector := buildTokenVector(profileText)

	scored := make([]scoredJobRecommendation, 0, len(jobs))
	for _, job := range jobs {
		jobText := strings.Join([]string{
			job.Title,
			job.Description,
			strings.Join(job.RequiredSkills, " "),
		}, " ")

		skillMatches, skillOverlap := calculateSkillOverlap(user.Skills, job.RequiredSkills)
		semanticScore := cosineSimilarity(profileVector, buildTokenVector(jobText))
		budgetScore := calculateBudgetScore(user, preferences, job)
		freshnessScore := calculateFreshnessScore(job.CreatedAt, time.Now())

		score := 0.55*semanticScore + 0.25*skillOverlap + 0.15*budgetScore + 0.05*freshnessScore
		if score <= 0 {
			continue
		}

		scored = append(scored, scoredJobRecommendation{
			recommendation: domain.JobRecommendation{
				JobID:       job.ID,
				MatchScore:  float32(math.Min(score, 0.999)),
				MatchReason: buildMatchReason(skillMatches, semanticScore, budgetScore, freshnessScore),
			},
			score:          score,
			skillMatches:   skillMatches,
			semanticScore:  semanticScore,
			budgetScore:    budgetScore,
			freshnessScore: freshnessScore,
		})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if math.Abs(scored[i].score-scored[j].score) > 0.0001 {
			return scored[i].score > scored[j].score
		}
		return scored[i].recommendation.JobID < scored[j].recommendation.JobID
	})

	return scored
}

func calculateSkillOverlap(userSkills, requiredSkills []string) ([]string, float64) {
	if len(userSkills) == 0 || len(requiredSkills) == 0 {
		return nil, 0
	}

	userSkillSet := make(map[string]string, len(userSkills))
	for _, skill := range userSkills {
		normalized := normalizeSkill(skill)
		if normalized == "" {
			continue
		}
		if _, exists := userSkillSet[normalized]; !exists {
			userSkillSet[normalized] = strings.TrimSpace(skill)
		}
	}

	matches := make([]string, 0, len(requiredSkills))
	for _, skill := range requiredSkills {
		normalized := normalizeSkill(skill)
		if normalized == "" {
			continue
		}
		if original, ok := userSkillSet[normalized]; ok {
			matches = append(matches, original)
		}
	}

	if len(matches) == 0 {
		return nil, 0
	}
	return uniqueStrings(matches), float64(len(matches)) / float64(len(requiredSkills))
}

func calculateBudgetScore(user domain.UserData, preferences domain.WorkPreferences, job domain.JobData) float64 {
	switch strings.TrimSpace(strings.ToLower(job.JobType)) {
	case jobTypeFixed:
		if preferences.MinBudgetUSD == 0 && preferences.MaxBudgetUSD == 0 {
			if job.BudgetMin > 0 || job.BudgetMax > 0 {
				return 0.7
			}
			return 0.5
		}
		if preferences.MinBudgetUSD > 0 && job.BudgetMax > 0 && job.BudgetMax < preferences.MinBudgetUSD {
			return 0
		}
		if preferences.MaxBudgetUSD > 0 && job.BudgetMin > 0 && job.BudgetMin > preferences.MaxBudgetUSD {
			return 0
		}
		return 1
	case jobTypeHourly:
		if user.HourlyRate <= 0 || job.HourlyRate <= 0 {
			return 0.5
		}
		ratio := job.HourlyRate / user.HourlyRate
		switch {
		case ratio >= 1:
			return 1
		case ratio >= 0.85:
			return 0.85
		case ratio >= 0.7:
			return 0.65
		case ratio >= 0.5:
			return 0.4
		default:
			return 0
		}
	default:
		return 0.5
	}
}

func calculateFreshnessScore(createdAt, now time.Time) float64 {
	if createdAt.IsZero() {
		return 0.5
	}
	age := now.Sub(createdAt)
	switch {
	case age <= 24*time.Hour:
		return 1
	case age <= 7*24*time.Hour:
		return 0.8
	case age <= 30*24*time.Hour:
		return 0.5
	default:
		return 0.2
	}
}

func buildMatchReason(skillMatches []string, semanticScore, budgetScore, freshnessScore float64) string {
	if len(skillMatches) > 0 {
		preview := skillMatches
		if len(preview) > 3 {
			preview = preview[:3]
		}
		return fmt.Sprintf("Matches your skills in %s", strings.Join(preview, ", "))
	}
	if semanticScore >= 0.35 {
		return "Strong semantic match for your profile"
	}
	if budgetScore >= 0.85 {
		return "Fits your current rate and work preferences"
	}
	if freshnessScore >= 0.8 {
		return "Recent opportunity aligned with your profile"
	}
	return "Broad match based on your profile and preferences"
}

func topSkills(skills []string, limit int) []string {
	if limit <= 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(skills))
	out := make([]string, 0, limit)
	for _, skill := range skills {
		normalized := normalizeSkill(skill)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, strings.TrimSpace(skill))
		if len(out) == limit {
			break
		}
	}
	return out
}

func limitRecommendations(recommendations []domain.JobRecommendation, limit int32) []domain.JobRecommendation {
	if int32(len(recommendations)) <= limit {
		return recommendations
	}
	return recommendations[:limit]
}

func (s *RecommendationService) normalizeLimit(limit int32) int32 {
	if limit <= 0 {
		return s.cfg.DefaultLimit
	}
	if limit > s.cfg.MaxLimit {
		return s.cfg.MaxLimit
	}
	return limit
}

func buildTokenVector(text string) map[string]float64 {
	vector := make(map[string]float64)
	for _, token := range tokenize(text) {
		vector[token]++
	}
	return vector
}

func tokenize(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

func cosineSimilarity(left, right map[string]float64) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}

	var dot, leftNorm, rightNorm float64
	for token, leftValue := range left {
		leftNorm += leftValue * leftValue
		if rightValue, ok := right[token]; ok {
			dot += leftValue * rightValue
		}
	}
	for _, rightValue := range right {
		rightNorm += rightValue * rightValue
	}
	if leftNorm == 0 || rightNorm == 0 {
		return 0
	}
	return dot / (math.Sqrt(leftNorm) * math.Sqrt(rightNorm))
}

func normalizeSkill(skill string) string {
	return strings.ToLower(strings.TrimSpace(skill))
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, strings.TrimSpace(value))
	}
	return out
}
