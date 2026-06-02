package config

import (
	"fmt"
	"os"
)

type Config struct {
	GRPCListenAddr    string
	PostgresURL       string
	WalletServiceAddr string
	JWTSecret         []byte
}

func LoadFromEnv() (Config, error) {
	secret := os.Getenv("AUTH_JWT_SECRET")
	if secret == "" {
		return Config{}, fmt.Errorf("AUTH_JWT_SECRET is required")
	}
	pg := os.Getenv("DISPUTE_POSTGRES_URL")
	if pg == "" {
		pg = os.Getenv("POSTGRES_URL")
	}
	if pg == "" {
		return Config{}, fmt.Errorf("DISPUTE_POSTGRES_URL is required")
	}
	addr := os.Getenv("DISPUTE_SERVICE_GRPC_ADDR")
	if addr == "" {
		addr = ":50066"
	}
	walletAddr := os.Getenv("WALLET_SERVICE_ADDR")
	if walletAddr == "" {
		walletAddr = "wallet:50058"
	}
	return Config{
		GRPCListenAddr:    addr,
		PostgresURL:       pg,
		WalletServiceAddr: walletAddr,
		JWTSecret:         []byte(secret),
	}, nil
}
