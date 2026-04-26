package cache

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"jobconnect/recommendation/internal/domain"
)

type fakeRedisClient struct {
	values      map[string]string
	expirations map[string]time.Duration
	err         error
	closed      bool
}

func newFakeRedisClient() *fakeRedisClient {
	return &fakeRedisClient{
		values:      make(map[string]string),
		expirations: make(map[string]time.Duration),
	}
}

func (f *fakeRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	if f.err != nil {
		return redis.NewStringResult("", f.err)
	}
	value, ok := f.values[key]
	if !ok {
		return redis.NewStringResult("", redis.Nil)
	}
	return redis.NewStringResult(value, nil)
}

func (f *fakeRedisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	if f.err != nil {
		return redis.NewStatusResult("", f.err)
	}
	switch v := value.(type) {
	case []byte:
		f.values[key] = string(v)
	case string:
		f.values[key] = v
	default:
		f.values[key] = ""
	}
	f.expirations[key] = expiration
	return redis.NewStatusResult("OK", nil)
}

func (f *fakeRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	if f.err != nil {
		return redis.NewIntResult(0, f.err)
	}
	var deleted int64
	for _, key := range keys {
		if _, ok := f.values[key]; ok {
			delete(f.values, key)
			delete(f.expirations, key)
			deleted++
		}
	}
	return redis.NewIntResult(deleted, nil)
}

func (f *fakeRedisClient) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	if f.err != nil {
		return redis.NewScanCmdResult(nil, 0, f.err)
	}
	prefix := strings.TrimSuffix(match, "*")
	keys := make([]string, 0)
	for key := range f.values {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	return redis.NewScanCmdResult(keys, 0, nil)
}

func (f *fakeRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	if f.err != nil {
		return redis.NewStatusResult("", f.err)
	}
	return redis.NewStatusResult("PONG", nil)
}

func (f *fakeRedisClient) Close() error {
	f.closed = true
	return nil
}

func TestRedisCacheRoundTripsJobRecommendations(t *testing.T) {
	client := newFakeRedisClient()
	cache := newRedisCacheWithClient(client, 2*time.Minute, time.Second)
	recs := []domain.JobRecommendation{
		{JobID: 101, MatchScore: 0.9, MatchReason: "skill match"},
		{JobID: 202, MatchScore: 0.7, MatchReason: "trust match"},
	}

	cache.SetRecommendedJobs(" freelancer-1 ", recs)

	key := jobRecommendationsKey("freelancer-1")
	if client.expirations[key] != 2*time.Minute {
		t.Fatalf("expected TTL to be preserved, got %s", client.expirations[key])
	}

	got, ok := cache.GetRecommendedJobs("freelancer-1")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(got) != len(recs) || got[0].JobID != 101 || got[1].JobID != 202 {
		t.Fatalf("unexpected cached jobs: %#v", got)
	}
}

func TestRedisCacheRoundTripsFreelancerRecommendations(t *testing.T) {
	client := newFakeRedisClient()
	cache := newRedisCacheWithClient(client, time.Minute, time.Second)
	recs := []domain.FreelancerRecommendation{
		{UserID: "f-1", MatchScore: 0.8, MatchReason: "skill match"},
	}

	cache.SetRecommendedFreelancers(" freelancers:77:caller ", recs)

	key := freelancerRecommendationsKey("freelancers:77:caller")
	if client.expirations[key] != time.Minute {
		t.Fatalf("expected TTL to be preserved, got %s", client.expirations[key])
	}

	got, ok := cache.GetRecommendedFreelancers("freelancers:77:caller")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(got) != 1 || got[0].UserID != "f-1" {
		t.Fatalf("unexpected cached freelancers: %#v", got)
	}
}

func TestRedisCacheMissesOnDecodeAndClientErrors(t *testing.T) {
	client := newFakeRedisClient()
	cache := newRedisCacheWithClient(client, time.Minute, time.Second)
	client.values[jobRecommendationsKey("freelancer-1")] = "not-json"

	if _, ok := cache.GetRecommendedJobs("freelancer-1"); ok {
		t.Fatal("expected malformed payload to miss")
	}

	client.err = errors.New("redis unavailable")
	if _, ok := cache.GetRecommendedJobs("freelancer-1"); ok {
		t.Fatal("expected client error to miss")
	}

	cache.SetRecommendedJobs("freelancer-2", []domain.JobRecommendation{{JobID: 2}})
	if _, ok := client.values[jobRecommendationsKey("freelancer-2")]; ok {
		t.Fatal("expected set failure to skip write")
	}
}

func TestRedisCacheDisabledWhenTTLZero(t *testing.T) {
	client := newFakeRedisClient()
	cache := newRedisCacheWithClient(client, 0, time.Second)

	cache.SetRecommendedJobs("freelancer-1", []domain.JobRecommendation{{JobID: 1}})
	if _, ok := client.values[jobRecommendationsKey("freelancer-1")]; ok {
		t.Fatal("expected disabled cache to skip writes")
	}
	if _, ok := cache.GetRecommendedJobs("freelancer-1"); ok {
		t.Fatal("expected disabled cache to miss")
	}
}

func TestRedisCacheDeletesJobRecommendations(t *testing.T) {
	client := newFakeRedisClient()
	cache := newRedisCacheWithClient(client, time.Minute, time.Second)

	cache.SetRecommendedJobs("freelancer-1", []domain.JobRecommendation{{JobID: 1}})
	if deleted := cache.DeleteRecommendedJobs("freelancer-1"); deleted != 1 {
		t.Fatalf("expected 1 deleted job recommendation cache entry, got %d", deleted)
	}
	if _, ok := cache.GetRecommendedJobs("freelancer-1"); ok {
		t.Fatal("expected job recommendation cache to be deleted")
	}
}

func TestRedisCacheDeletesFreelancerRecommendationsForJob(t *testing.T) {
	client := newFakeRedisClient()
	cache := newRedisCacheWithClient(client, time.Minute, time.Second)

	cache.SetRecommendedFreelancers("freelancers:77:caller-a", []domain.FreelancerRecommendation{{UserID: "f-1"}})
	cache.SetRecommendedFreelancers("freelancers:77:caller-b", []domain.FreelancerRecommendation{{UserID: "f-2"}})
	cache.SetRecommendedFreelancers("freelancers:88:caller-a", []domain.FreelancerRecommendation{{UserID: "f-3"}})

	if deleted := cache.DeleteRecommendedFreelancersForJob(77); deleted != 2 {
		t.Fatalf("expected 2 deleted freelancer recommendation cache entries, got %d", deleted)
	}
	if _, ok := cache.GetRecommendedFreelancers("freelancers:77:caller-a"); ok {
		t.Fatal("expected caller-a cache entry to be deleted")
	}
	if _, ok := cache.GetRecommendedFreelancers("freelancers:77:caller-b"); ok {
		t.Fatal("expected caller-b cache entry to be deleted")
	}
	if _, ok := cache.GetRecommendedFreelancers("freelancers:88:caller-a"); !ok {
		t.Fatal("expected other job cache entry to remain")
	}
}

func TestRedisCacheClearDeletesAllRecommendationEntries(t *testing.T) {
	client := newFakeRedisClient()
	cache := newRedisCacheWithClient(client, time.Minute, time.Second)

	cache.SetRecommendedJobs("freelancer-1", []domain.JobRecommendation{{JobID: 1}})
	cache.SetRecommendedFreelancers("freelancers:77:caller-a", []domain.FreelancerRecommendation{{UserID: "f-1"}})
	client.values["other:key"] = "keep"

	if deleted := cache.Clear(); deleted != 2 {
		t.Fatalf("expected 2 deleted recommendation cache entries, got %d", deleted)
	}
	if _, ok := client.values[jobRecommendationsKey("freelancer-1")]; ok {
		t.Fatal("expected job recommendation cache to be deleted")
	}
	if _, ok := client.values[freelancerRecommendationsKey("freelancers:77:caller-a")]; ok {
		t.Fatal("expected freelancer recommendation cache to be deleted")
	}
	if _, ok := client.values["other:key"]; !ok {
		t.Fatal("expected unrelated Redis key to remain")
	}
}
