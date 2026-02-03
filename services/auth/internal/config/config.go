package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	GRPCListenAddr string
	PostgresURL    string

	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		GRPCListenAddr:  getEnv("AUTH_GRPC_LISTEN_ADDR", ":50051"),
		PostgresURL:     os.Getenv("AUTH_POSTGRES_URL"),
		AccessTokenTTL:  getEnvDurationSeconds("AUTH_ACCESS_TOKEN_TTL_SECONDS", 15*60),
		RefreshTokenTTL: getEnvDurationSeconds("AUTH_REFRESH_TOKEN_TTL_SECONDS", 30*24*60*60),
	}
	if cfg.PostgresURL == "" {
		return Config{}, fmt.Errorf("AUTH_POSTGRES_URL is required")
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvDurationSeconds(key string, defSeconds int) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return time.Duration(defSeconds) * time.Second
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return time.Duration(defSeconds) * time.Second
	}
	return time.Duration(n) * time.Second
}
