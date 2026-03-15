package config

import (
	"fmt"
	"os"
)

type Config struct {
	GRPCListenAddr string
	PostgresURL    string
	JWTSecret      []byte
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		GRPCListenAddr: getEnv("WALLET_GRPC_LISTEN_ADDR", ":50059"),
		PostgresURL:    os.Getenv("WALLET_POSTGRES_URL"),
	}
	if cfg.PostgresURL == "" {
		return Config{}, fmt.Errorf("WALLET_POSTGRES_URL is required")
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
