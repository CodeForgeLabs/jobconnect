package cache

import (
	"sync"
	"time"

	"jobconnect/recommendation/internal/domain"
)

type MemoryCache struct {
	mu       sync.RWMutex
	ttl      time.Duration
	entries  map[string]entry
	clockNow func() time.Time
}

type entry struct {
	recommendations []domain.JobRecommendation
	expiresAt       time.Time
}

func NewMemoryCache(ttl time.Duration) *MemoryCache {
	return &MemoryCache{
		ttl:      ttl,
		entries:  make(map[string]entry),
		clockNow: time.Now,
	}
}

func (c *MemoryCache) GetRecommendedJobs(userID string) ([]domain.JobRecommendation, bool) {
	if c == nil || c.ttl == 0 {
		return nil, false
	}

	c.mu.RLock()
	cached, ok := c.entries[userID]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if c.clockNow().After(cached.expiresAt) {
		c.mu.Lock()
		delete(c.entries, userID)
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
	c.entries[userID] = entry{
		recommendations: copied,
		expiresAt:       c.clockNow().Add(c.ttl),
	}
	c.mu.Unlock()
}
