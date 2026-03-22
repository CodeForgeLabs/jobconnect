package config

import (
	"fmt"
	"os"
)

type Config struct {
	GRPCListenAddr string
	HTTPListenAddr string
	PostgresURL    string
	JWTSecret      []byte
	
	ChapaSecretKey string
	TelebirrAppKey string
	TelebirrAppID  string
	
	WalletSvcAddr       string
	ContractSvcAddr     string
	VerificationSvcAddr string
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		GRPCListenAddr: getEnv("PAYMENT_GRPC_LISTEN_ADDR", ":50061"),
		HTTPListenAddr: getEnv("PAYMENT_HTTP_LISTEN_ADDR", ":8081"),
		PostgresURL:    os.Getenv("PAYMENT_POSTGRES_URL"),
		ChapaSecretKey: getEnv("CHAPA_SECRET_KEY", "sandbox_secret"),
		TelebirrAppKey: getEnv("TELEBIRR_APP_KEY", "sandbox_key"),
		TelebirrAppID:  getEnv("TELEBIRR_APP_ID", "sandbox_id"),
		WalletSvcAddr:       getEnv("WALLET_SVC_ADDR", "wallet:50059"),
		ContractSvcAddr:     getEnv("CONTRACT_SVC_ADDR", "contract:50055"),
		VerificationSvcAddr: getEnv("VERIFICATION_SVC_ADDR", "verification:50060"),
	}
	if cfg.PostgresURL == "" {
		return Config{}, fmt.Errorf("PAYMENT_POSTGRES_URL is required")
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
