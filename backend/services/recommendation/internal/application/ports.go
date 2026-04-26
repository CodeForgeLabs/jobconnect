package application

import (
	"context"
	"errors"
	"time"

	"jobconnect/recommendation/internal/domain"
)

// ErrEmbedderUnavailable signals that the semantic embedder could not produce
// vectors for the requested texts. Callers should treat this as a soft failure
// and fall back to the non-semantic ranking path; a request must never fail
// just because the embedder is down.
var ErrEmbedderUnavailable = errors.New("recommendation embedder unavailable")

// Embedder converts free-form text into dense vectors for semantic similarity.
// Implementations must be safe for concurrent use. Returning
// ErrEmbedderUnavailable asks callers to fall back to token cosine.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

type noopEmbedder struct{}

func (noopEmbedder) Embed(context.Context, []string) ([][]float32, error) {
	return nil, ErrEmbedderUnavailable
}

// EmbeddingSourceType identifies what kind of entity an embedding represents.
// Used as part of the store key so freelancer and job IDs cannot collide.
type EmbeddingSourceType string

const (
	EmbeddingSourceTypeFreelancer EmbeddingSourceType = "freelancer"
	EmbeddingSourceTypeJob        EmbeddingSourceType = "job"
)

// StoredEmbedding is a persisted vector plus the hash of the text it was
// computed from. The hash lets callers detect when the source text has
// changed and the vector needs to be recomputed.
type StoredEmbedding struct {
	SourceType EmbeddingSourceType
	SourceID   string
	TextHash   string
	Vector     []float32
}

// EmbeddingStore persists and retrieves embeddings keyed by (sourceType,
// sourceID). Get returns (_, false, nil) on a clean miss; the error channel
// is reserved for transport/encoding failures. Implementations must be safe
// for concurrent use.
type EmbeddingStore interface {
	Get(ctx context.Context, sourceType EmbeddingSourceType, sourceID string) (StoredEmbedding, bool, error)
	Upsert(ctx context.Context, embedding StoredEmbedding) error
}

type noopEmbeddingStore struct{}

func (noopEmbeddingStore) Get(context.Context, EmbeddingSourceType, string) (StoredEmbedding, bool, error) {
	return StoredEmbedding{}, false, nil
}

func (noopEmbeddingStore) Upsert(context.Context, StoredEmbedding) error { return nil }

type JobServiceClient interface {
	ListRecentPublicOpenJobs(ctx context.Context, pageSize int32) ([]domain.JobData, error)
	SearchPublicOpenJobsBySkill(ctx context.Context, skill string, pageSize int32) ([]domain.JobData, error)
	GetJob(ctx context.Context, jobID int64) (domain.JobData, error)
}

type UserServiceClient interface {
	GetFreelancer(ctx context.Context, userID string) (domain.UserData, error)
	GetWorkPreferences(ctx context.Context, userID string) (domain.WorkPreferences, error)
	ListDiscoverableFreelancers(ctx context.Context, skills []string, pageSize int32) ([]domain.FreelancerData, error)
}

type ReviewServiceClient interface {
	GetUserRatingSummary(ctx context.Context, userID string) (domain.RatingSummary, error)
}

type RecommendationCache interface {
	GetRecommendedJobs(userID string) ([]domain.JobRecommendation, bool)
	SetRecommendedJobs(userID string, recommendations []domain.JobRecommendation)
	DeleteRecommendedJobs(userID string) int
	GetRecommendedFreelancers(key string) ([]domain.FreelancerRecommendation, bool)
	SetRecommendedFreelancers(key string, recommendations []domain.FreelancerRecommendation)
	DeleteRecommendedFreelancersForJob(jobID int64) int
	Clear() int
}

type MetricsRecorder interface {
	RecordCacheHit(recommendationType string, cachedCount, returnedCount int, elapsed time.Duration)
	RecordCacheMiss(recommendationType string)
	RecordCacheDisabled(recommendationType string)
	RecordRecomputeComplete(recommendationType string, candidateCount, rankedCount, returnedCount int, elapsed time.Duration)
	RecordRecomputeError(recommendationType string, elapsed time.Duration)
	RecordInvalidation(scope string, deleted int, elapsed time.Duration)
	RecordReviewLookupError(role string)
	RecordRedisError(op string)
}

type noopMetricsRecorder struct{}

func (noopMetricsRecorder) RecordCacheHit(string, int, int, time.Duration) {}
func (noopMetricsRecorder) RecordCacheMiss(string)                         {}
func (noopMetricsRecorder) RecordCacheDisabled(string)                     {}
func (noopMetricsRecorder) RecordRecomputeComplete(string, int, int, int, time.Duration) {
}
func (noopMetricsRecorder) RecordRecomputeError(string, time.Duration) {}
func (noopMetricsRecorder) RecordInvalidation(string, int, time.Duration) {
}
func (noopMetricsRecorder) RecordReviewLookupError(string) {}
func (noopMetricsRecorder) RecordRedisError(string)        {}
