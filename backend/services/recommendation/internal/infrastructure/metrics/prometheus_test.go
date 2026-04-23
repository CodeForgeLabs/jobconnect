package metrics

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPrometheusRecorderScrapeExposesInstrumentedMetrics(t *testing.T) {
	r := NewPrometheusRecorder()

	r.RecordCacheHit("jobs", 1, 2, 3*time.Millisecond)
	r.RecordCacheMiss("jobs")
	r.RecordCacheDisabled("freelancers")
	r.RecordRecomputeComplete("jobs", 20, 15, 10, 120*time.Millisecond)
	r.RecordRecomputeError("freelancers", 80*time.Millisecond)
	r.RecordInvalidation("user", 2, 1*time.Millisecond)
	r.RecordReviewLookupError("freelancer")
	r.RecordRedisError("get")

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	r.Handler().ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	out := string(body)

	wants := []string{
		`recommendation_cache_lookups_total{result="hit",type="jobs"} 1`,
		`recommendation_cache_lookups_total{result="miss",type="jobs"} 1`,
		`recommendation_cache_lookups_total{result="disabled",type="freelancers"} 1`,
		`recommendation_recomputes_total{result="complete",type="jobs"} 1`,
		`recommendation_recomputes_total{result="error",type="freelancers"} 1`,
		`recommendation_invalidations_total{scope="user"} 1`,
		`recommendation_invalidation_deletions_total{scope="user"} 2`,
		`recommendation_review_lookup_errors_total{role="freelancer"} 1`,
		`recommendation_cache_redis_errors_total{op="get"} 1`,
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Errorf("scrape missing %q", w)
		}
	}
}

func TestPrometheusRecorderNoInvalidationDeletionsWhenZero(t *testing.T) {
	r := NewPrometheusRecorder()
	r.RecordInvalidation("job", 0, time.Millisecond)

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	r.Handler().ServeHTTP(rec, req)

	body, _ := io.ReadAll(rec.Body)
	if strings.Contains(string(body), `recommendation_invalidation_deletions_total{scope="job"}`) {
		t.Error("deletions counter should not be incremented when deleted=0")
	}
}
