package config

import "testing"

func TestLoadFromEnv_Minimal(t *testing.T) {
	t.Setenv("AUTH_POSTGRES_URL", "postgres://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("AUTH_JWT_SECRET", "secret")
	// Leave TTL envs unset to use defaults.

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv error: %v", err)
	}
	if cfg.PostgresURL == "" {
		t.Fatalf("expected PostgresURL to be set")
	}
	if len(cfg.JWTSecret) == 0 {
		t.Fatalf("expected JWTSecret to be set")
	}
}
