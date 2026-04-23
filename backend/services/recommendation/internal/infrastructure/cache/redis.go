package cache

import (
	"context"
	"encoding/json"
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
}

type RedisCache struct {
	client           redisClient
	ttl              time.Duration
	operationTimeout time.Duration
}

type redisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Ping(ctx context.Context) *redis.StatusCmd
	Close() error
}

func NewRedisCache(cfg RedisConfig) *RedisCache {
	timeout := cfg.OperationTimeout
	if timeout <= 0 {
		timeout = defaultRedisOperationTimeout
	}

	return &RedisCache{
		client: redis.NewClient(&redis.Options{
			Addr:     strings.TrimSpace(cfg.Addr),
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
		ttl:              cfg.TTL,
		operationTimeout: timeout,
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
		return false
	}
	if err := json.Unmarshal([]byte(raw), dest); err != nil {
		log.Printf("recommendation cache: redis decode %q failed: %v", key, err)
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
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.operationTimeout)
	defer cancel()

	if err := c.client.Set(ctx, key, payload, c.ttl).Err(); err != nil {
		log.Printf("recommendation cache: redis set %q failed: %v", key, err)
	}
}

func jobRecommendationsKey(userID string) string {
	return redisCachePrefix + ":jobs:" + strings.TrimSpace(userID)
}

func freelancerRecommendationsKey(key string) string {
	return redisCachePrefix + ":freelancers:" + strings.TrimSpace(key)
}
