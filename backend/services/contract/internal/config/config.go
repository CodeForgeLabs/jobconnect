package config

import (
	"fmt"
	"os"
)

type Config struct {
	GRPCListenAddr      string
	PostgresURL         string
	ProposalServiceAddr string
	JobServiceAddr      string
	UserServiceAddr     string
	WalletServiceAddr   string
	DisputeServiceAddr  string
	JWTSecret           []byte
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		GRPCListenAddr:      getEnv("CONTRACT_GRPC_LISTEN_ADDR", ":50055"),
		PostgresURL:         os.Getenv("CONTRACT_POSTGRES_URL"),
		ProposalServiceAddr: os.Getenv("PROPOSAL_SERVICE_GRPC_ADDR"),
		JobServiceAddr:      os.Getenv("JOB_SERVICE_GRPC_ADDR"),
		UserServiceAddr:     os.Getenv("USER_SERVICE_GRPC_ADDR"),
		WalletServiceAddr:   getEnv("WALLET_SERVICE_GRPC_ADDR", "wallet:50058"),
		DisputeServiceAddr:  getEnv("DISPUTE_SERVICE_GRPC_ADDR", "dispute:50066"),
	}
	if cfg.PostgresURL == "" {
		return Config{}, fmt.Errorf("CONTRACT_POSTGRES_URL is required")
	}
	if cfg.ProposalServiceAddr == "" {
		return Config{}, fmt.Errorf("PROPOSAL_SERVICE_GRPC_ADDR is required")
	}
	if cfg.JobServiceAddr == "" {
		return Config{}, fmt.Errorf("JOB_SERVICE_GRPC_ADDR is required")
	}
	if cfg.UserServiceAddr == "" {
		return Config{}, fmt.Errorf("USER_SERVICE_GRPC_ADDR is required")
	}
	if cfg.WalletServiceAddr == "" {
		return Config{}, fmt.Errorf("WALLET_SERVICE_GRPC_ADDR is required")
	}
	if cfg.DisputeServiceAddr == "" {
		return Config{}, fmt.Errorf("DISPUTE_SERVICE_GRPC_ADDR is required")
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
