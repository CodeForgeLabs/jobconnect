package config

import "testing"

func TestLoadFromEnvReadsRedisCacheSettings(t *testing.T) {
	t.Setenv("RECOMMENDATION_CACHE_BACKEND", "Redis")
	t.Setenv("RECOMMENDATION_REDIS_ADDR", "redis:6379")
	t.Setenv("RECOMMENDATION_REDIS_PASSWORD", "secret")
	t.Setenv("RECOMMENDATION_REDIS_DB", "2")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv returned error: %v", err)
	}
	if cfg.RecommendationCacheBackend != "redis" {
		t.Fatalf("expected redis backend, got %q", cfg.RecommendationCacheBackend)
	}
	if cfg.RecommendationRedisAddr != "redis:6379" {
		t.Fatalf("expected redis address, got %q", cfg.RecommendationRedisAddr)
	}
	if cfg.RecommendationRedisPassword != "secret" {
		t.Fatalf("expected redis password, got %q", cfg.RecommendationRedisPassword)
	}
	if cfg.RecommendationRedisDB != 2 {
		t.Fatalf("expected redis db 2, got %d", cfg.RecommendationRedisDB)
	}
}

func TestLoadFromEnvRejectsInvalidCacheBackend(t *testing.T) {
	t.Setenv("RECOMMENDATION_CACHE_BACKEND", "postgres")

	if _, err := LoadFromEnv(); err == nil {
		t.Fatal("expected invalid cache backend error")
	}
}
