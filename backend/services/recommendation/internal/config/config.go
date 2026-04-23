package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	GRPCListenAddr              string
	JobServiceAddr              string
	UserServiceAddr             string
	ReviewServiceAddr           string
	RecommendationCacheBackend  string
	RecommendationRedisAddr     string
	RecommendationRedisPassword string
	RecommendationRedisDB       int
	DefaultRecommendationLimit  int32
	MaxRecommendationLimit      int32
	CandidatePageSize           int32
	PerSkillPageSize            int32
	MaxSkillQueries             int
	RecommendationCacheTTL      time.Duration
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		GRPCListenAddr:              getEnv("RECOMMENDATION_GRPC_LISTEN_ADDR", ":50064"),
		JobServiceAddr:              getEnv("JOB_SERVICE_ADDR", "localhost:50053"),
		UserServiceAddr:             getEnv("USER_SERVICE_ADDR", "localhost:50052"),
		ReviewServiceAddr:           getEnv("REVIEW_SERVICE_ADDR", "localhost:50056"),
		RecommendationCacheBackend:  strings.ToLower(getEnv("RECOMMENDATION_CACHE_BACKEND", "memory")),
		RecommendationRedisAddr:     getEnv("RECOMMENDATION_REDIS_ADDR", "localhost:6379"),
		RecommendationRedisPassword: getEnv("RECOMMENDATION_REDIS_PASSWORD", ""),
		RecommendationRedisDB:       getIntEnv("RECOMMENDATION_REDIS_DB", 0),
		DefaultRecommendationLimit:  int32(getIntEnv("RECOMMENDATION_DEFAULT_LIMIT", 10)),
		MaxRecommendationLimit:      int32(getIntEnv("RECOMMENDATION_MAX_LIMIT", 25)),
		CandidatePageSize:           int32(getIntEnv("RECOMMENDATION_CANDIDATE_PAGE_SIZE", 100)),
		PerSkillPageSize:            int32(getIntEnv("RECOMMENDATION_PER_SKILL_PAGE_SIZE", 25)),
		MaxSkillQueries:             getIntEnv("RECOMMENDATION_MAX_SKILL_QUERIES", 5),
		RecommendationCacheTTL:      getDurationEnv("RECOMMENDATION_CACHE_TTL", 2*time.Minute),
	}

	if cfg.DefaultRecommendationLimit <= 0 {
		return Config{}, fmt.Errorf("RECOMMENDATION_DEFAULT_LIMIT must be greater than zero")
	}
	if cfg.MaxRecommendationLimit < cfg.DefaultRecommendationLimit {
		return Config{}, fmt.Errorf("RECOMMENDATION_MAX_LIMIT must be greater than or equal to RECOMMENDATION_DEFAULT_LIMIT")
	}
	if cfg.CandidatePageSize <= 0 {
		return Config{}, fmt.Errorf("RECOMMENDATION_CANDIDATE_PAGE_SIZE must be greater than zero")
	}
	if cfg.PerSkillPageSize <= 0 {
		return Config{}, fmt.Errorf("RECOMMENDATION_PER_SKILL_PAGE_SIZE must be greater than zero")
	}
	if cfg.MaxSkillQueries <= 0 {
		return Config{}, fmt.Errorf("RECOMMENDATION_MAX_SKILL_QUERIES must be greater than zero")
	}
	if cfg.RecommendationCacheTTL < 0 {
		return Config{}, fmt.Errorf("RECOMMENDATION_CACHE_TTL must be greater than or equal to zero")
	}
	if cfg.RecommendationCacheBackend != "memory" && cfg.RecommendationCacheBackend != "redis" {
		return Config{}, fmt.Errorf("RECOMMENDATION_CACHE_BACKEND must be memory or redis")
	}
	if cfg.RecommendationCacheBackend == "redis" && cfg.RecommendationRedisAddr == "" {
		return Config{}, fmt.Errorf("RECOMMENDATION_REDIS_ADDR is required when RECOMMENDATION_CACHE_BACKEND=redis")
	}
	if cfg.RecommendationRedisDB < 0 {
		return Config{}, fmt.Errorf("RECOMMENDATION_REDIS_DB must be greater than or equal to zero")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getIntEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
