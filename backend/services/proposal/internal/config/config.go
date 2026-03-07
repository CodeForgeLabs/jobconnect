package config

import (
	"fmt"
	"os"
)

type Config struct {
	GRPCListenAddr string
	PostgresURL    string
	JWTSecret      []byte
	JobServiceAddr string
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		GRPCListenAddr: getEnv("PROPOSAL_GRPC_LISTEN_ADDR", ":50054"),
		PostgresURL:    os.Getenv("PROPOSAL_POSTGRES_URL"),
		JobServiceAddr: os.Getenv("JOB_SERVICE_GRPC_ADDR"),
	}
	if cfg.PostgresURL == "" {
		return Config{}, fmt.Errorf("PROPOSAL_POSTGRES_URL is required")
	}
	if cfg.JobServiceAddr == "" {
		return Config{}, fmt.Errorf("JOB_SERVICE_GRPC_ADDR is required")
	}
	secret := os.Getenv("AUTH_JWT_SECRET")
	if secret == "" {
		return Config{}, fmt.Errorf("AUTH_JWT_SECRET is required")
	}
	cfg.JWTSecret = []byte(secret)
	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
