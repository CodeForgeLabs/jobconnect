package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type keyLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type InMemoryLimiter struct {
	mu      sync.Mutex
	entries map[string]*keyLimiter
	rate    rate.Limit
	burst   int
	ttl     time.Duration
}

func NewInMemoryLimiter(rps rate.Limit, burst int, ttl time.Duration) *InMemoryLimiter {
	return &InMemoryLimiter{
		entries: make(map[string]*keyLimiter),
		rate:    rps,
		burst:   burst,
		ttl:     ttl,
	}
}

func (l *InMemoryLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if key == "" {
			key = "unknown"
		}

		if !l.allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":              "too many requests",
				"challenge_required": true,
				"challenge_endpoint": "/api/v1/auth/challenge",
			})
			return
		}

		c.Next()
	}
}

func (l *InMemoryLimiter) allow(key string) bool {
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	for k, entry := range l.entries {
		if now.Sub(entry.lastSeen) > l.ttl {
			delete(l.entries, k)
		}
	}

	entry, ok := l.entries[key]
	if !ok {
		entry = &keyLimiter{limiter: rate.NewLimiter(l.rate, l.burst)}
		l.entries[key] = entry
	}
	entry.lastSeen = now

	return entry.limiter.Allow()
}
