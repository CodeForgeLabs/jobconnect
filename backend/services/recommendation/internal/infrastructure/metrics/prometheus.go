package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var cacheHitDurationBuckets = []float64{
	0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25,
}

var recomputeDurationBuckets = []float64{
	0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10,
}

var countBuckets = []float64{0, 1, 2, 5, 10, 25, 50, 100, 250, 500, 1000}

// PrometheusRecorder implements application.MetricsRecorder and cache.MetricsRecorder
// against a Prometheus registry.
type PrometheusRecorder struct {
	registry *prometheus.Registry

	cacheLookups          *prometheus.CounterVec
	cacheHitDuration      *prometheus.HistogramVec
	recomputes            *prometheus.CounterVec
	recomputeDuration     *prometheus.HistogramVec
	candidateCount        *prometheus.HistogramVec
	rankedCount           *prometheus.HistogramVec
	returnedCount         *prometheus.HistogramVec
	invalidations         *prometheus.CounterVec
	invalidationDeletions *prometheus.CounterVec
	invalidationDuration  *prometheus.HistogramVec
	reviewLookupErrors    *prometheus.CounterVec
	redisErrors           *prometheus.CounterVec
}

func NewPrometheusRecorder() *PrometheusRecorder {
	registry := prometheus.NewRegistry()

	r := &PrometheusRecorder{
		registry: registry,
		cacheLookups: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "recommendation_cache_lookups_total",
			Help: "Recommendation cache lookups by type and result.",
		}, []string{"type", "result"}),
		cacheHitDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "recommendation_cache_hit_duration_seconds",
			Help:    "End-to-end duration of requests served from cache.",
			Buckets: cacheHitDurationBuckets,
		}, []string{"type"}),
		recomputes: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "recommendation_recomputes_total",
			Help: "Recommendation recomputes by type and result.",
		}, []string{"type", "result"}),
		recomputeDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "recommendation_recompute_duration_seconds",
			Help:    "End-to-end duration of recommendation recomputes.",
			Buckets: recomputeDurationBuckets,
		}, []string{"type", "result"}),
		candidateCount: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "recommendation_recompute_candidate_count",
			Help:    "Candidate count per recompute.",
			Buckets: countBuckets,
		}, []string{"type"}),
		rankedCount: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "recommendation_recompute_ranked_count",
			Help:    "Ranked count per recompute.",
			Buckets: countBuckets,
		}, []string{"type"}),
		returnedCount: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "recommendation_recompute_returned_count",
			Help:    "Returned count per recompute (after limit).",
			Buckets: countBuckets,
		}, []string{"type"}),
		invalidations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "recommendation_invalidations_total",
			Help: "Cache invalidation calls by scope.",
		}, []string{"scope"}),
		invalidationDeletions: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "recommendation_invalidation_deletions_total",
			Help: "Cache keys deleted by invalidation, by scope.",
		}, []string{"scope"}),
		invalidationDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "recommendation_invalidation_duration_seconds",
			Help:    "Duration of cache invalidation calls by scope.",
			Buckets: cacheHitDurationBuckets,
		}, []string{"scope"}),
		reviewLookupErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "recommendation_review_lookup_errors_total",
			Help: "Review service lookup errors by role.",
		}, []string{"role"}),
		redisErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "recommendation_cache_redis_errors_total",
			Help: "Redis cache adapter errors by operation.",
		}, []string{"op"}),
	}

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		r.cacheLookups,
		r.cacheHitDuration,
		r.recomputes,
		r.recomputeDuration,
		r.candidateCount,
		r.rankedCount,
		r.returnedCount,
		r.invalidations,
		r.invalidationDeletions,
		r.invalidationDuration,
		r.reviewLookupErrors,
		r.redisErrors,
	)

	return r
}

// Handler returns an http.Handler that serves the Prometheus text exposition
// for this recorder's registry.
func (r *PrometheusRecorder) Handler() http.Handler {
	return promhttp.HandlerFor(r.registry, promhttp.HandlerOpts{Registry: r.registry})
}

func (r *PrometheusRecorder) RecordCacheHit(recommendationType string, _, _ int, elapsed time.Duration) {
	r.cacheLookups.WithLabelValues(recommendationType, "hit").Inc()
	r.cacheHitDuration.WithLabelValues(recommendationType).Observe(elapsed.Seconds())
}

func (r *PrometheusRecorder) RecordCacheMiss(recommendationType string) {
	r.cacheLookups.WithLabelValues(recommendationType, "miss").Inc()
}

func (r *PrometheusRecorder) RecordCacheDisabled(recommendationType string) {
	r.cacheLookups.WithLabelValues(recommendationType, "disabled").Inc()
}

func (r *PrometheusRecorder) RecordRecomputeComplete(recommendationType string, candidateCount, rankedCount, returnedCount int, elapsed time.Duration) {
	r.recomputes.WithLabelValues(recommendationType, "complete").Inc()
	r.recomputeDuration.WithLabelValues(recommendationType, "complete").Observe(elapsed.Seconds())
	r.candidateCount.WithLabelValues(recommendationType).Observe(float64(candidateCount))
	r.rankedCount.WithLabelValues(recommendationType).Observe(float64(rankedCount))
	r.returnedCount.WithLabelValues(recommendationType).Observe(float64(returnedCount))
}

func (r *PrometheusRecorder) RecordRecomputeError(recommendationType string, elapsed time.Duration) {
	r.recomputes.WithLabelValues(recommendationType, "error").Inc()
	r.recomputeDuration.WithLabelValues(recommendationType, "error").Observe(elapsed.Seconds())
}

func (r *PrometheusRecorder) RecordInvalidation(scope string, deleted int, elapsed time.Duration) {
	r.invalidations.WithLabelValues(scope).Inc()
	if deleted > 0 {
		r.invalidationDeletions.WithLabelValues(scope).Add(float64(deleted))
	}
	r.invalidationDuration.WithLabelValues(scope).Observe(elapsed.Seconds())
}

func (r *PrometheusRecorder) RecordReviewLookupError(role string) {
	r.reviewLookupErrors.WithLabelValues(role).Inc()
}

func (r *PrometheusRecorder) RecordRedisError(op string) {
	r.redisErrors.WithLabelValues(op).Inc()
}
