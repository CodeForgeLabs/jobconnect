package cache

import (
	"testing"
	"time"

	"jobconnect/recommendation/internal/domain"
)

func TestMemoryCacheDeletesRecommendedJobs(t *testing.T) {
	cache := NewMemoryCache(time.Minute)
	cache.SetRecommendedJobs("freelancer-1", []domain.JobRecommendation{{JobID: 1}})

	if deleted := cache.DeleteRecommendedJobs("freelancer-1"); deleted != 1 {
		t.Fatalf("expected 1 deleted job recommendation cache entry, got %d", deleted)
	}
	if _, ok := cache.GetRecommendedJobs("freelancer-1"); ok {
		t.Fatal("expected job recommendation cache to be deleted")
	}
}

func TestMemoryCacheDeletesFreelancersForJob(t *testing.T) {
	cache := NewMemoryCache(time.Minute)
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

func TestMemoryCacheClearDeletesAllEntries(t *testing.T) {
	cache := NewMemoryCache(time.Minute)
	cache.SetRecommendedJobs("freelancer-1", []domain.JobRecommendation{{JobID: 1}})
	cache.SetRecommendedFreelancers("freelancers:77:caller-a", []domain.FreelancerRecommendation{{UserID: "f-1"}})

	if deleted := cache.Clear(); deleted != 2 {
		t.Fatalf("expected 2 deleted cache entries, got %d", deleted)
	}
	if _, ok := cache.GetRecommendedJobs("freelancer-1"); ok {
		t.Fatal("expected job recommendation cache to be deleted")
	}
	if _, ok := cache.GetRecommendedFreelancers("freelancers:77:caller-a"); ok {
		t.Fatal("expected freelancer recommendation cache to be deleted")
	}
}
