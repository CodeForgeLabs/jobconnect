package application

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"jobconnect/recommendation/internal/domain"
)

const (
	publicVisibility = "public"
	jobTypeFixed     = "fixed"
	jobTypeHourly    = "hourly"

	availabilityFullTime    = "AVAILABILITY_FULL_TIME"
	availabilityPartTime    = "AVAILABILITY_PART_TIME"
	availabilityAsNeeded    = "AVAILABILITY_AS_NEEDED"
	availabilityUnavailable = "AVAILABILITY_UNAVAILABLE"

	trustEnrichmentLimit = 50
	neutralTrustScore    = 0.5

	recommendationTypeJobs        = "jobs"
	recommendationTypeFreelancers = "freelancers"
)

type ServiceConfig struct {
	DefaultLimit      int32
	MaxLimit          int32
	CandidatePageSize int32
	PerSkillPageSize  int32
	MaxSkillQueries   int
}

type RecommendationService struct {
	jobClient      JobServiceClient
	userClient     UserServiceClient
	reviewClient   ReviewServiceClient
	cache          RecommendationCache
	metrics        MetricsRecorder
	embedder       Embedder
	embeddingStore EmbeddingStore
	cfg            ServiceConfig
}

type scoredJobRecommendation struct {
	recommendation domain.JobRecommendation
	score          float64
	clientID       string
	skillMatches   []string
	skillScore     float64
	semanticScore  float64
	budgetScore    float64
	freshnessScore float64
	trustScore     float64
	ratingSummary  domain.RatingSummary
}

func NewRecommendationService(
	jobClient JobServiceClient,
	userClient UserServiceClient,
	reviewClient ReviewServiceClient,
	cache RecommendationCache,
	metrics MetricsRecorder,
	embedder Embedder,
	embeddingStore EmbeddingStore,
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
	if metrics == nil {
		metrics = noopMetricsRecorder{}
	}
	if embedder == nil {
		embedder = noopEmbedder{}
	}
	if embeddingStore == nil {
		embeddingStore = noopEmbeddingStore{}
	}

	return &RecommendationService{
		jobClient:      jobClient,
		userClient:     userClient,
		reviewClient:   reviewClient,
		cache:          cache,
		metrics:        metrics,
		embedder:       embedder,
		embeddingStore: embeddingStore,
		cfg:            cfg,
	}
}

func (s *RecommendationService) GetRecommendedJobs(ctx context.Context, userID string, limit int32) ([]domain.JobRecommendation, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	startedAt := time.Now()
	normalizedLimit := s.normalizeLimit(limit)
	if s.cache != nil {
		if cached, ok := s.cache.GetRecommendedJobs(userID); ok {
			limited := limitRecommendations(cached, normalizedLimit)
			s.emitCacheHit(recommendationTypeJobs, len(cached), len(limited), normalizedLimit, time.Since(startedAt))
			return limited, nil
		}
		s.emitCacheMiss(recommendationTypeJobs, normalizedLimit)
	} else {
		s.emitCacheDisabled(recommendationTypeJobs, normalizedLimit)
	}

	user, err := s.userClient.GetFreelancer(ctx, userID)
	if err != nil {
		s.emitRecomputeError(recommendationTypeJobs, time.Since(startedAt), err)
		return nil, fmt.Errorf("fetch freelancer profile: %w", err)
	}
	if !user.CanApplyJobs {
		s.emitRecomputeComplete(recommendationTypeJobs, 0, 0, 0, normalizedLimit, time.Since(startedAt))
		return nil, nil
	}

	preferences, err := s.userClient.GetWorkPreferences(ctx, userID)
	if err != nil {
		s.emitRecomputeError(recommendationTypeJobs, time.Since(startedAt), err)
		return nil, fmt.Errorf("fetch freelancer work preferences: %w", err)
	}

	candidates, err := s.collectCandidates(ctx, user, preferences)
	if err != nil {
		s.emitRecomputeError(recommendationTypeJobs, time.Since(startedAt), err)
		return nil, err
	}
	if len(candidates) == 0 {
		s.emitRecomputeComplete(recommendationTypeJobs, 0, 0, 0, normalizedLimit, time.Since(startedAt))
		return nil, nil
	}

	ranked := s.rankJobs(ctx, user, preferences, candidates)
	recommendations := make([]domain.JobRecommendation, 0, len(ranked))
	for _, rankedJob := range ranked {
		recommendations = append(recommendations, rankedJob.recommendation)
	}

	if s.cache != nil {
		s.cache.SetRecommendedJobs(userID, recommendations)
	}
	limited := limitRecommendations(recommendations, normalizedLimit)
	s.emitRecomputeComplete(recommendationTypeJobs, len(candidates), len(recommendations), len(limited), normalizedLimit, time.Since(startedAt))
	return limited, nil
}

func (s *RecommendationService) GetRecommendedFreelancers(ctx context.Context, jobID int64, limit int32, callerScope string) ([]domain.FreelancerRecommendation, error) {
	if jobID <= 0 {
		return nil, fmt.Errorf("job_id is required")
	}

	startedAt := time.Now()
	normalizedLimit := s.normalizeLimit(limit)
	job, err := s.jobClient.GetJob(ctx, jobID)
	if err != nil {
		s.emitRecomputeError(recommendationTypeFreelancers, time.Since(startedAt), err)
		return nil, fmt.Errorf("fetch job: %w", err)
	}
	if job.ID == 0 {
		s.emitRecomputeComplete(recommendationTypeFreelancers, 0, 0, 0, normalizedLimit, time.Since(startedAt))
		return nil, fmt.Errorf("job %d not found", jobID)
	}

	cacheKey := freelancerCacheKey(jobID, callerScope)
	if s.cache != nil {
		if cached, ok := s.cache.GetRecommendedFreelancers(cacheKey); ok {
			limited := limitFreelancerRecommendations(cached, normalizedLimit)
			s.emitCacheHit(recommendationTypeFreelancers, len(cached), len(limited), normalizedLimit, time.Since(startedAt))
			return limited, nil
		}
		s.emitCacheMiss(recommendationTypeFreelancers, normalizedLimit)
	} else {
		s.emitCacheDisabled(recommendationTypeFreelancers, normalizedLimit)
	}

	candidates, err := s.collectFreelancerCandidates(ctx, job)
	if err != nil {
		s.emitRecomputeError(recommendationTypeFreelancers, time.Since(startedAt), err)
		return nil, err
	}
	if len(candidates) == 0 {
		s.emitRecomputeComplete(recommendationTypeFreelancers, 0, 0, 0, normalizedLimit, time.Since(startedAt))
		return nil, nil
	}

	ranked := s.rankFreelancers(ctx, job, candidates)
	recommendations := make([]domain.FreelancerRecommendation, 0, len(ranked))
	for _, entry := range ranked {
		recommendations = append(recommendations, entry.recommendation)
	}

	if s.cache != nil {
		s.cache.SetRecommendedFreelancers(cacheKey, recommendations)
	}
	limited := limitFreelancerRecommendations(recommendations, normalizedLimit)
	s.emitRecomputeComplete(recommendationTypeFreelancers, len(candidates), len(recommendations), len(limited), normalizedLimit, time.Since(startedAt))
	return limited, nil
}

func (s *RecommendationService) emitCacheHit(recommendationType string, cachedCount, returnedCount int, limit int32, elapsed time.Duration) {
	log.Printf(
		"recommendation: type=%s cache=hit cached_count=%d returned_count=%d limit=%d elapsed=%s",
		recommendationType,
		cachedCount,
		returnedCount,
		limit,
		elapsed,
	)
	s.metrics.RecordCacheHit(recommendationType, cachedCount, returnedCount, elapsed)
}

func (s *RecommendationService) emitCacheMiss(recommendationType string, limit int32) {
	log.Printf("recommendation: type=%s cache=miss limit=%d", recommendationType, limit)
	s.metrics.RecordCacheMiss(recommendationType)
}

func (s *RecommendationService) emitCacheDisabled(recommendationType string, limit int32) {
	log.Printf("recommendation: type=%s cache=disabled limit=%d", recommendationType, limit)
	s.metrics.RecordCacheDisabled(recommendationType)
}

func (s *RecommendationService) emitRecomputeComplete(recommendationType string, candidateCount, rankedCount, returnedCount int, limit int32, elapsed time.Duration) {
	log.Printf(
		"recommendation: type=%s recompute=complete candidate_count=%d ranked_count=%d returned_count=%d limit=%d elapsed=%s",
		recommendationType,
		candidateCount,
		rankedCount,
		returnedCount,
		limit,
		elapsed,
	)
	s.metrics.RecordRecomputeComplete(recommendationType, candidateCount, rankedCount, returnedCount, elapsed)
}

func (s *RecommendationService) emitRecomputeError(recommendationType string, elapsed time.Duration, err error) {
	log.Printf("recommendation: type=%s recompute=error elapsed=%s error=%v", recommendationType, elapsed, err)
	s.metrics.RecordRecomputeError(recommendationType, elapsed)
}

func (s *RecommendationService) InvalidateRecommendationCache(ctx context.Context, userIDs []string, jobIDs []int64, all bool) (int, error) {
	_ = ctx
	if s.cache == nil {
		log.Printf("recommendation: cache invalidation skipped cache=disabled users=%d jobs=%d all=%t", len(userIDs), len(jobIDs), all)
		return 0, nil
	}

	startedAt := time.Now()
	if all {
		deleted := s.cache.Clear()
		s.emitInvalidated("all", deleted, len(userIDs), len(jobIDs), time.Since(startedAt))
		return deleted, nil
	}

	deleted := 0
	for _, userID := range uniqueNonEmptyStrings(userIDs) {
		deleted += s.cache.DeleteRecommendedJobs(userID)
	}
	for _, jobID := range uniquePositiveInt64s(jobIDs) {
		deleted += s.cache.DeleteRecommendedFreelancersForJob(jobID)
	}

	s.emitInvalidated("targeted", deleted, len(userIDs), len(jobIDs), time.Since(startedAt))
	return deleted, nil
}

func (s *RecommendationService) emitInvalidated(scope string, deleted, userIDCount, jobIDCount int, elapsed time.Duration) {
	log.Printf(
		"recommendation: cache=invalidate scope=%s deleted_count=%d user_id_count=%d job_id_count=%d elapsed=%s",
		scope,
		deleted,
		userIDCount,
		jobIDCount,
		elapsed,
	)
	s.metrics.RecordInvalidation(scope, deleted, elapsed)
}

func freelancerCacheKey(jobID int64, callerScope string) string {
	return fmt.Sprintf("freelancers:%d:%s", jobID, strings.TrimSpace(callerScope))
}

type scoredFreelancerRecommendation struct {
	recommendation    domain.FreelancerRecommendation
	score             float64
	skillMatches      []string
	skillScore        float64
	semanticScore     float64
	rateScore         float64
	availabilityScore float64
	trustScore        float64
	ratingSummary     domain.RatingSummary
}

func (s *RecommendationService) collectFreelancerCandidates(ctx context.Context, job domain.JobData) ([]domain.FreelancerData, error) {
	candidateMap := make(map[string]domain.FreelancerData)

	topJobSkills := topSkills(job.RequiredSkills, s.cfg.MaxSkillQueries)
	base, err := s.userClient.ListDiscoverableFreelancers(ctx, topJobSkills, s.cfg.CandidatePageSize)
	if err != nil {
		return nil, fmt.Errorf("list discoverable freelancers: %w", err)
	}
	for _, f := range base {
		if f.ID == "" {
			continue
		}
		candidateMap[f.ID] = f
	}

	if len(candidateMap) == 0 && len(topJobSkills) == 0 {
		return nil, nil
	}

	candidates := make([]domain.FreelancerData, 0, len(candidateMap))
	for _, f := range candidateMap {
		if !isEligibleFreelancer(f, job) {
			continue
		}
		candidates = append(candidates, f)
	}
	return candidates, nil
}

func isEligibleFreelancer(f domain.FreelancerData, job domain.JobData) bool {
	if strings.EqualFold(strings.TrimSpace(f.Availability), availabilityUnavailable) {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(job.JobType), jobTypeHourly) {
		if f.HourlyRate > 0 && job.HourlyRate > 0 && f.HourlyRate > job.HourlyRate*2 {
			return false
		}
	}
	return true
}

func (s *RecommendationService) rankFreelancers(ctx context.Context, job domain.JobData, candidates []domain.FreelancerData) []scoredFreelancerRecommendation {
	jobText := strings.Join([]string{
		job.Title,
		job.Description,
		strings.Join(job.RequiredSkills, " "),
	}, " ")
	jobVector := buildTokenVector(jobText)

	jobIDStr := strconv.FormatInt(job.ID, 10)
	embedReqs := make([]EmbeddingRequest, 0, len(candidates)+1)
	embedReqs = append(embedReqs, EmbeddingRequest{
		SourceType: EmbeddingSourceTypeJob,
		SourceID:   jobIDStr,
		Text:       jobText,
	})
	candidateTexts := make([]string, len(candidates))
	for i, f := range candidates {
		profileText := strings.Join([]string{
			f.Headline,
			f.Bio,
			strings.Join(f.Skills, " "),
		}, " ")
		candidateTexts[i] = profileText
		embedReqs = append(embedReqs, EmbeddingRequest{
			SourceType: EmbeddingSourceTypeFreelancer,
			SourceID:   f.ID,
			Text:       profileText,
		})
	}
	vectors := s.resolveEmbeddings(ctx, embedReqs)
	jobVec, jobVecOK := vectors.lookup(EmbeddingSourceTypeJob, jobIDStr)

	scored := make([]scoredFreelancerRecommendation, 0, len(candidates))
	for i, f := range candidates {
		profileText := candidateTexts[i]

		skillMatches, skillOverlap := calculateSkillOverlap(f.Skills, job.RequiredSkills)
		semanticScore := 0.0
		semanticPath := "token"
		if jobVecOK {
			if candVec, ok := vectors.lookup(EmbeddingSourceTypeFreelancer, f.ID); ok {
				semanticScore = denseCosineSimilarity(jobVec, candVec)
				semanticPath = "embedding"
			}
		}
		if semanticPath == "token" {
			semanticScore = cosineSimilarity(jobVector, buildTokenVector(profileText))
		}
		s.metrics.RecordSemanticPath("freelancers", semanticPath)
		rateScore := calculateFreelancerRateScore(f, job)
		availabilityScore := calculateAvailabilityScore(f.Availability)
		ratingScore := clamp01(f.Rating / 5)

		preTrustScore := 0.40*semanticScore +
			0.30*skillOverlap +
			0.15*rateScore +
			0.05*availabilityScore +
			0.10*ratingScore

		if preTrustScore <= 0 {
			continue
		}
		if len(skillMatches) == 0 && semanticScore < 0.2 {
			continue
		}

		scored = append(scored, scoredFreelancerRecommendation{
			recommendation: domain.FreelancerRecommendation{
				UserID:      f.ID,
				MatchScore:  float32(math.Min(preTrustScore, 0.999)),
				MatchReason: buildFreelancerMatchReason(skillMatches, semanticScore, rateScore, neutralTrustScore, domain.RatingSummary{}),
			},
			score:             preTrustScore,
			skillMatches:      skillMatches,
			skillScore:        skillOverlap,
			semanticScore:     semanticScore,
			rateScore:         rateScore,
			availabilityScore: availabilityScore,
			trustScore:        neutralTrustScore,
		})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if math.Abs(scored[i].score-scored[j].score) > 0.0001 {
			return scored[i].score > scored[j].score
		}
		return scored[i].recommendation.UserID < scored[j].recommendation.UserID
	})

	s.enrichFreelancerTrust(ctx, scored)
	for i := range scored {
		scored[i].score = 0.35*scored[i].semanticScore +
			0.30*scored[i].skillScore +
			0.15*scored[i].rateScore +
			0.15*scored[i].trustScore +
			0.05*scored[i].availabilityScore
		scored[i].recommendation.MatchScore = float32(math.Min(scored[i].score, 0.999))
		scored[i].recommendation.MatchReason = buildFreelancerMatchReason(
			scored[i].skillMatches,
			scored[i].semanticScore,
			scored[i].rateScore,
			scored[i].trustScore,
			scored[i].ratingSummary,
		)
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if math.Abs(scored[i].score-scored[j].score) > 0.0001 {
			return scored[i].score > scored[j].score
		}
		return scored[i].recommendation.UserID < scored[j].recommendation.UserID
	})
	return scored
}

func (s *RecommendationService) enrichFreelancerTrust(ctx context.Context, scored []scoredFreelancerRecommendation) {
	for i := range scored {
		scored[i].trustScore = neutralTrustScore
	}
	if s.reviewClient == nil {
		return
	}

	limit := min(len(scored), trustEnrichmentLimit)
	summaries := make(map[string]domain.RatingSummary, limit)
	for i := 0; i < limit; i++ {
		userID := strings.TrimSpace(scored[i].recommendation.UserID)
		if userID == "" {
			continue
		}
		summary, ok := summaries[userID]
		if !ok {
			var err error
			summary, err = s.reviewClient.GetUserRatingSummary(ctx, userID)
			if err != nil {
				log.Printf("recommendation: review summary lookup failed for freelancer %q: %v", userID, err)
				s.metrics.RecordReviewLookupError("freelancer")
				continue
			}
			summaries[userID] = summary
		}
		scored[i].ratingSummary = summary
		scored[i].trustScore = calculateTrustScore(summary)
	}
}

func calculateFreelancerRateScore(f domain.FreelancerData, job domain.JobData) float64 {
	switch strings.TrimSpace(strings.ToLower(job.JobType)) {
	case jobTypeHourly:
		if f.HourlyRate <= 0 || job.HourlyRate <= 0 {
			return 0.5
		}
		ratio := job.HourlyRate / f.HourlyRate
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
	case jobTypeFixed:
		return 0.6
	default:
		return 0.5
	}
}

func calculateAvailabilityScore(availability string) float64 {
	switch strings.ToUpper(strings.TrimSpace(availability)) {
	case availabilityFullTime:
		return 1
	case availabilityPartTime:
		return 0.75
	case availabilityAsNeeded:
		return 0.5
	case availabilityUnavailable:
		return 0
	default:
		return 0.4
	}
}

func buildFreelancerMatchReason(skillMatches []string, semanticScore, rateScore, trustScore float64, summary domain.RatingSummary) string {
	if len(skillMatches) > 0 {
		preview := skillMatches
		if len(preview) > 3 {
			preview = preview[:3]
		}
		return fmt.Sprintf("Matches required skills: %s", strings.Join(preview, ", "))
	}
	if trustScore >= 0.8 && summary.TotalReviews > 0 {
		return fmt.Sprintf("Top-rated freelancer (%.1f avg, %d reviews)", summary.AverageRating, summary.TotalReviews)
	}
	if semanticScore >= 0.35 {
		return "Strong profile match for this job"
	}
	if rateScore >= 0.85 {
		return "Rate aligns with job budget"
	}
	return "Broad match based on profile and availability"
}

func limitFreelancerRecommendations(recommendations []domain.FreelancerRecommendation, limit int32) []domain.FreelancerRecommendation {
	if int32(len(recommendations)) <= limit {
		return recommendations
	}
	return recommendations[:limit]
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

func (s *RecommendationService) rankJobs(ctx context.Context, user domain.UserData, preferences domain.WorkPreferences, jobs []domain.JobData) []scoredJobRecommendation {
	profileText := strings.Join([]string{
		user.Headline,
		user.Bio,
		strings.Join(user.Skills, " "),
	}, " ")
	profileVector := buildTokenVector(profileText)

	embedReqs := make([]EmbeddingRequest, 0, len(jobs)+1)
	embedReqs = append(embedReqs, EmbeddingRequest{
		SourceType: EmbeddingSourceTypeFreelancer,
		SourceID:   user.ID,
		Text:       profileText,
	})
	jobTexts := make([]string, len(jobs))
	jobIDStrs := make([]string, len(jobs))
	for i, job := range jobs {
		jobText := strings.Join([]string{
			job.Title,
			job.Description,
			strings.Join(job.RequiredSkills, " "),
		}, " ")
		jobTexts[i] = jobText
		jobIDStrs[i] = strconv.FormatInt(job.ID, 10)
		embedReqs = append(embedReqs, EmbeddingRequest{
			SourceType: EmbeddingSourceTypeJob,
			SourceID:   jobIDStrs[i],
			Text:       jobText,
		})
	}
	vectors := s.resolveEmbeddings(ctx, embedReqs)
	profileVec, profileVecOK := vectors.lookup(EmbeddingSourceTypeFreelancer, user.ID)

	scored := make([]scoredJobRecommendation, 0, len(jobs))
	for i, job := range jobs {
		jobText := jobTexts[i]

		skillMatches, skillOverlap := calculateSkillOverlap(user.Skills, job.RequiredSkills)
		semanticScore := 0.0
		semanticPath := "token"
		if profileVecOK {
			if jobVec, ok := vectors.lookup(EmbeddingSourceTypeJob, jobIDStrs[i]); ok {
				semanticScore = denseCosineSimilarity(profileVec, jobVec)
				semanticPath = "embedding"
			}
		}
		if semanticPath == "token" {
			semanticScore = cosineSimilarity(profileVector, buildTokenVector(jobText))
		}
		s.metrics.RecordSemanticPath("jobs", semanticPath)
		budgetScore := calculateBudgetScore(user, preferences, job)
		freshnessScore := calculateFreshnessScore(job.CreatedAt, time.Now())

		preTrustScore := 0.55*semanticScore + 0.25*skillOverlap + 0.15*budgetScore + 0.05*freshnessScore
		if preTrustScore <= 0 {
			continue
		}

		scored = append(scored, scoredJobRecommendation{
			recommendation: domain.JobRecommendation{
				JobID:       job.ID,
				MatchScore:  float32(math.Min(preTrustScore, 0.999)),
				MatchReason: buildMatchReason(skillMatches, semanticScore, budgetScore, freshnessScore, neutralTrustScore, domain.RatingSummary{}),
			},
			score:          preTrustScore,
			clientID:       job.ClientID,
			skillMatches:   skillMatches,
			skillScore:     skillOverlap,
			semanticScore:  semanticScore,
			budgetScore:    budgetScore,
			freshnessScore: freshnessScore,
			trustScore:     neutralTrustScore,
		})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if math.Abs(scored[i].score-scored[j].score) > 0.0001 {
			return scored[i].score > scored[j].score
		}
		return scored[i].recommendation.JobID < scored[j].recommendation.JobID
	})

	s.enrichJobTrust(ctx, scored)
	for i := range scored {
		scored[i].score = 0.45*scored[i].semanticScore +
			0.25*scored[i].skillScore +
			0.15*scored[i].budgetScore +
			0.05*scored[i].freshnessScore +
			0.10*scored[i].trustScore
		scored[i].recommendation.MatchScore = float32(math.Min(scored[i].score, 0.999))
		scored[i].recommendation.MatchReason = buildMatchReason(
			scored[i].skillMatches,
			scored[i].semanticScore,
			scored[i].budgetScore,
			scored[i].freshnessScore,
			scored[i].trustScore,
			scored[i].ratingSummary,
		)
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if math.Abs(scored[i].score-scored[j].score) > 0.0001 {
			return scored[i].score > scored[j].score
		}
		return scored[i].recommendation.JobID < scored[j].recommendation.JobID
	})

	return scored
}

func (s *RecommendationService) enrichJobTrust(ctx context.Context, scored []scoredJobRecommendation) {
	for i := range scored {
		scored[i].trustScore = neutralTrustScore
	}
	if s.reviewClient == nil {
		return
	}

	limit := min(len(scored), trustEnrichmentLimit)
	summaries := make(map[string]domain.RatingSummary, limit)
	for i := 0; i < limit; i++ {
		clientID := strings.TrimSpace(scored[i].clientID)
		if clientID == "" {
			continue
		}
		summary, ok := summaries[clientID]
		if !ok {
			var err error
			summary, err = s.reviewClient.GetUserRatingSummary(ctx, clientID)
			if err != nil {
				log.Printf("recommendation: review summary lookup failed for client %q: %v", clientID, err)
				s.metrics.RecordReviewLookupError("client")
				continue
			}
			summaries[clientID] = summary
		}
		scored[i].ratingSummary = summary
		scored[i].trustScore = calculateTrustScore(summary)
	}
}

func calculateTrustScore(summary domain.RatingSummary) float64 {
	if summary.TotalReviews <= 0 {
		return neutralTrustScore
	}
	return clamp01(summary.AverageRating / 5)
}

func clamp01(value float64) float64 {
	if math.IsNaN(value) || value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
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

func buildMatchReason(skillMatches []string, semanticScore, budgetScore, freshnessScore, trustScore float64, summary domain.RatingSummary) string {
	if len(skillMatches) > 0 {
		preview := skillMatches
		if len(preview) > 3 {
			preview = preview[:3]
		}
		return fmt.Sprintf("Matches your skills in %s", strings.Join(preview, ", "))
	}
	if trustScore >= 0.8 && summary.TotalReviews > 0 {
		return fmt.Sprintf("Highly rated client (%.1f avg, %d reviews)", summary.AverageRating, summary.TotalReviews)
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

func denseCosineSimilarity(left, right []float32) float64 {
	if len(left) == 0 || len(right) == 0 || len(left) != len(right) {
		return 0
	}
	var dot, leftNorm, rightNorm float64
	for i := range left {
		l := float64(left[i])
		r := float64(right[i])
		dot += l * r
		leftNorm += l * l
		rightNorm += r * r
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

func uniqueNonEmptyStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func uniquePositiveInt64s(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	out := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
