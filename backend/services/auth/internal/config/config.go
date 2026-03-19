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
	OTPTTL          time.Duration

	JWTSecret       []byte
	UserServiceAddr string

	SMTPHost        string
	SMTPPort        int
	SMTPTLSMode     string
	SMTPUsername    string
	SMTPPassword    string
	SMTPFromAddress string
	SMTPFromName    string
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		GRPCListenAddr:  getEnv("AUTH_GRPC_LISTEN_ADDR", ":50051"),
		PostgresURL:     os.Getenv("AUTH_POSTGRES_URL"),
		AccessTokenTTL:  getEnvDurationSeconds("AUTH_ACCESS_TOKEN_TTL_SECONDS", 15*60),
		RefreshTokenTTL: getEnvDurationSeconds("AUTH_REFRESH_TOKEN_TTL_SECONDS", 30*24*60*60),
		OTPTTL:          getEnvDurationSeconds("AUTH_OTP_TTL_SECONDS", 15*60),
		UserServiceAddr: os.Getenv("USER_SERVICE_GRPC_ADDR"),
		SMTPHost:        os.Getenv("AUTH_SMTP_HOST"),
		SMTPPort:        getEnvInt("AUTH_SMTP_PORT", 587),
		SMTPTLSMode:     getEnv("AUTH_SMTP_TLS_MODE", "starttls"),
		SMTPUsername:    os.Getenv("AUTH_SMTP_USERNAME"),
		SMTPPassword:    os.Getenv("AUTH_SMTP_PASSWORD"),
		SMTPFromAddress: os.Getenv("AUTH_SMTP_FROM_ADDRESS"),
		SMTPFromName:    os.Getenv("AUTH_SMTP_FROM_NAME"),
	}
	if cfg.PostgresURL == "" {
		return Config{}, fmt.Errorf("AUTH_POSTGRES_URL is required")
	}
	secret := os.Getenv("AUTH_JWT_SECRET")
	if secret == "" {
		return Config{}, fmt.Errorf("AUTH_JWT_SECRET is required")
	}
	if cfg.SMTPHost != "" {
		if cfg.SMTPPort <= 0 {
			return Config{}, fmt.Errorf("AUTH_SMTP_PORT must be a positive integer when AUTH_SMTP_HOST is set")
		}
		switch cfg.SMTPTLSMode {
		case "starttls", "implicit", "none":
		default:
			return Config{}, fmt.Errorf("AUTH_SMTP_TLS_MODE must be one of: starttls, implicit, none")
		}
		if cfg.SMTPFromAddress == "" {
			return Config{}, fmt.Errorf("AUTH_SMTP_FROM_ADDRESS is required when AUTH_SMTP_HOST is set")
		}
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

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
