package cache

import (
	"sync"
	"time"

	"jobconnect/recommendation/internal/domain"
)

type MemoryCache struct {
	mu                 sync.RWMutex
	ttl                time.Duration
	jobEntries         map[string]jobEntry
	freelancerEntries  map[string]freelancerEntry
	clockNow           func() time.Time
}

type jobEntry struct {
	recommendations []domain.JobRecommendation
	expiresAt       time.Time
}

type freelancerEntry struct {
	recommendations []domain.FreelancerRecommendation
	expiresAt       time.Time
}

func NewMemoryCache(ttl time.Duration) *MemoryCache {
	return &MemoryCache{
		ttl:               ttl,
		jobEntries:        make(map[string]jobEntry),
		freelancerEntries: make(map[string]freelancerEntry),
		clockNow:          time.Now,
	}
}

func (c *MemoryCache) GetRecommendedJobs(userID string) ([]domain.JobRecommendation, bool) {
	if c == nil || c.ttl == 0 {
		return nil, false
	}

	c.mu.RLock()
	cached, ok := c.jobEntries[userID]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if c.clockNow().After(cached.expiresAt) {
		c.mu.Lock()
		delete(c.jobEntries, userID)
		c.mu.Unlock()
		return nil, false
	}

	out := make([]domain.JobRecommendation, len(cached.recommendations))
	copy(out, cached.recommendations)
	return out, true
}

func (c *MemoryCache) SetRecommendedJobs(userID string, recommendations []domain.JobRecommendation) {
	if c == nil || c.ttl == 0 {
		return
	}

	copied := make([]domain.JobRecommendation, len(recommendations))
	copy(copied, recommendations)

	c.mu.Lock()
	c.jobEntries[userID] = jobEntry{
		recommendations: copied,
		expiresAt:       c.clockNow().Add(c.ttl),
	}
	c.mu.Unlock()
}

func (c *MemoryCache) GetRecommendedFreelancers(key string) ([]domain.FreelancerRecommendation, bool) {
	if c == nil || c.ttl == 0 {
		return nil, false
	}

	c.mu.RLock()
	cached, ok := c.freelancerEntries[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if c.clockNow().After(cached.expiresAt) {
		c.mu.Lock()
		delete(c.freelancerEntries, key)
		c.mu.Unlock()
		return nil, false
	}

	out := make([]domain.FreelancerRecommendation, len(cached.recommendations))
	copy(out, cached.recommendations)
	return out, true
}

func (c *MemoryCache) SetRecommendedFreelancers(key string, recommendations []domain.FreelancerRecommendation) {
	if c == nil || c.ttl == 0 {
		return
	}

	copied := make([]domain.FreelancerRecommendation, len(recommendations))
	copy(copied, recommendations)

	c.mu.Lock()
	c.freelancerEntries[key] = freelancerEntry{
		recommendations: copied,
		expiresAt:       c.clockNow().Add(c.ttl),
	}
	c.mu.Unlock()
}
