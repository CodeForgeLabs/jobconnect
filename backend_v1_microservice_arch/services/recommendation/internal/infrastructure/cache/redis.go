package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"jobconnect/recommendation/internal/domain"
)

const (
	redisCachePrefix             = "recommendation:v1"
	defaultRedisOperationTimeout = 500 * time.Millisecond
)

type RedisConfig struct {
	Addr             string
	Password         string
	DB               int
	TTL              time.Duration
	OperationTimeout time.Duration
	Metrics          MetricsRecorder
}

// MetricsRecorder observes redis adapter error paths. A nil value is supported.
type MetricsRecorder interface {
	RecordRedisError(op string)
}

type noopMetricsRecorder struct{}

func (noopMetricsRecorder) RecordRedisError(string) {}

type RedisCache struct {
	client           redisClient
	ttl              time.Duration
	operationTimeout time.Duration
	metrics          MetricsRecorder
}

type redisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd
	Ping(ctx context.Context) *redis.StatusCmd
	Close() error
}

func NewRedisCache(cfg RedisConfig) *RedisCache {
	timeout := cfg.OperationTimeout
	if timeout <= 0 {
		timeout = defaultRedisOperationTimeout
	}
	metrics := cfg.Metrics
	if metrics == nil {
		metrics = noopMetricsRecorder{}
	}

	return &RedisCache{
		client: redis.NewClient(&redis.Options{
			Addr:     strings.TrimSpace(cfg.Addr),
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
		ttl:              cfg.TTL,
		operationTimeout: timeout,
		metrics:          metrics,
	}
}

func newRedisCacheWithClient(client redisClient, ttl time.Duration, operationTimeout time.Duration) *RedisCache {
	if operationTimeout <= 0 {
		operationTimeout = defaultRedisOperationTimeout
	}
	return &RedisCache{
		client:           client,
		ttl:              ttl,
		operationTimeout: operationTimeout,
		metrics:          noopMetricsRecorder{},
	}
}

func (c *RedisCache) Ping(ctx context.Context) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Ping(ctx).Err()
}

func (c *RedisCache) Close() error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Close()
}

func (c *RedisCache) GetRecommendedJobs(userID string) ([]domain.JobRecommendation, bool) {
	var recommendations []domain.JobRecommendation
	if !c.getJSON(jobRecommendationsKey(userID), &recommendations) {
		return nil, false
	}
	return recommendations, true
}

func (c *RedisCache) SetRecommendedJobs(userID string, recommendations []domain.JobRecommendation) {
	c.setJSON(jobRecommendationsKey(userID), recommendations)
}

func (c *RedisCache) DeleteRecommendedJobs(userID string) int {
	return c.deleteKeys(jobRecommendationsKey(userID))
}

func (c *RedisCache) GetRecommendedFreelancers(key string) ([]domain.FreelancerRecommendation, bool) {
	var recommendations []domain.FreelancerRecommendation
	if !c.getJSON(freelancerRecommendationsKey(key), &recommendations) {
		return nil, false
	}
	return recommendations, true
}

func (c *RedisCache) SetRecommendedFreelancers(key string, recommendations []domain.FreelancerRecommendation) {
	c.setJSON(freelancerRecommendationsKey(key), recommendations)
}

func (c *RedisCache) DeleteRecommendedFreelancersForJob(jobID int64) int {
	return c.deleteKeysByPattern(freelancerRecommendationsJobPattern(jobID))
}

func (c *RedisCache) Clear() int {
	return c.deleteKeysByPattern(redisCachePrefix + ":*")
}

func (c *RedisCache) getJSON(key string, dest any) bool {
	if c == nil || c.client == nil || c.ttl == 0 {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.operationTimeout)
	defer cancel()

	raw, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false
	}
	if err != nil {
		log.Printf("recommendation cache: redis get %q failed: %v", key, err)
		c.metrics.RecordRedisError("get")
		return false
	}
	if err := json.Unmarshal([]byte(raw), dest); err != nil {
		log.Printf("recommendation cache: redis decode %q failed: %v", key, err)
		c.metrics.RecordRedisError("decode")
		return false
	}
	return true
}

func (c *RedisCache) setJSON(key string, value any) {
	if c == nil || c.client == nil || c.ttl == 0 {
		return
	}

	payload, err := json.Marshal(value)
	if err != nil {
		log.Printf("recommendation cache: redis encode %q failed: %v", key, err)
		c.metrics.RecordRedisError("encode")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.operationTimeout)
	defer cancel()

	if err := c.client.Set(ctx, key, payload, c.ttl).Err(); err != nil {
		log.Printf("recommendation cache: redis set %q failed: %v", key, err)
		c.metrics.RecordRedisError("set")
	}
}

func (c *RedisCache) deleteKeys(keys ...string) int {
	if c == nil || c.client == nil || len(keys) == 0 {
		return 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.operationTimeout)
	defer cancel()

	deleted, err := c.client.Del(ctx, keys...).Result()
	if err != nil {
		log.Printf("recommendation cache: redis delete failed keys=%d: %v", len(keys), err)
		c.metrics.RecordRedisError("delete")
		return 0
	}
	return int(deleted)
}

func (c *RedisCache) deleteKeysByPattern(pattern string) int {
	if c == nil || c.client == nil {
		return 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.operationTimeout)
	defer cancel()

	var cursor uint64
	deleted := 0
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			log.Printf("recommendation cache: redis scan failed pattern=%q: %v", pattern, err)
			c.metrics.RecordRedisError("scan")
			return deleted
		}
		if len(keys) > 0 {
			count, err := c.client.Del(ctx, keys...).Result()
			if err != nil {
				log.Printf("recommendation cache: redis delete failed pattern=%q keys=%d: %v", pattern, len(keys), err)
				c.metrics.RecordRedisError("delete")
				return deleted
			}
			deleted += int(count)
		}
		if nextCursor == 0 {
			return deleted
		}
		cursor = nextCursor
	}
}

func jobRecommendationsKey(userID string) string {
	return redisCachePrefix + ":jobs:" + strings.TrimSpace(userID)
}

func freelancerRecommendationsKey(key string) string {
	return redisCachePrefix + ":freelancers:" + strings.TrimSpace(key)
}

func freelancerRecommendationsJobPattern(jobID int64) string {
	return redisCachePrefix + fmt.Sprintf(":freelancers:freelancers:%d:*", jobID)
}
