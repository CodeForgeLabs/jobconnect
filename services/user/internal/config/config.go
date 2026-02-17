package config

import (
	"fmt"
	"os"
)

type Config struct {
	GRPCListenAddr string
	PostgresURL    string
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		GRPCListenAddr: getEnv("USER_GRPC_LISTEN_ADDR", ":50052"),
		PostgresURL:    os.Getenv("USER_POSTGRES_URL"),
	}
	if cfg.PostgresURL == "" {
		return Config{}, fmt.Errorf("USER_POSTGRES_URL is required")
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
