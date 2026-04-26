package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	GRPCListenAddr      string
	PostgresURL         string
	ProposalServiceAddr string
	JobServiceAddr      string
	UserServiceAddr     string
	WalletServiceAddr   string
	DisputeServiceAddr  string
	HourlyEvidenceStore HourlyEvidenceStorageConfig
	JWTSecret           []byte
}

type HourlyEvidenceStorageConfig struct {
	Provider      string
	Bucket        string
	Endpoint      string
	Region        string
	AccessKey     string
	SecretKey     string
	UseSSL        bool
	PathStyle     bool
	CreateBucket  bool
	PresignPutTTL string
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
		HourlyEvidenceStore: HourlyEvidenceStorageConfig{
			Provider:      strings.ToLower(strings.TrimSpace(getEnv("CONTRACT_HOURLY_EVIDENCE_STORAGE_PROVIDER", "minio"))),
			Bucket:        strings.TrimSpace(getEnv("CONTRACT_HOURLY_EVIDENCE_STORAGE_BUCKET", "jobconnect-contract-hourly-evidence")),
			Endpoint:      strings.TrimSpace(getEnv("CONTRACT_HOURLY_EVIDENCE_STORAGE_ENDPOINT", "localhost:9000")),
			Region:        strings.TrimSpace(getEnv("CONTRACT_HOURLY_EVIDENCE_STORAGE_REGION", "us-east-1")),
			AccessKey:     strings.TrimSpace(getEnv("CONTRACT_HOURLY_EVIDENCE_STORAGE_ACCESS_KEY", "minioadmin")),
			SecretKey:     strings.TrimSpace(getEnv("CONTRACT_HOURLY_EVIDENCE_STORAGE_SECRET_KEY", "minioadmin")),
			UseSSL:        getEnvBool("CONTRACT_HOURLY_EVIDENCE_STORAGE_USE_SSL", false),
			PathStyle:     getEnvBool("CONTRACT_HOURLY_EVIDENCE_STORAGE_PATH_STYLE", true),
			CreateBucket:  getEnvBool("CONTRACT_HOURLY_EVIDENCE_STORAGE_CREATE_BUCKET", true),
			PresignPutTTL: strings.TrimSpace(getEnv("CONTRACT_HOURLY_EVIDENCE_STORAGE_PRESIGN_PUT_TTL", "15m")),
		},
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
	if err := cfg.HourlyEvidenceStore.Validate(); err != nil {
		return Config{}, err
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

func getEnvBool(key string, def bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return def
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return parsed
}

func (c HourlyEvidenceStorageConfig) Validate() error {
	if c.Provider != "minio" {
		return fmt.Errorf("CONTRACT_HOURLY_EVIDENCE_STORAGE_PROVIDER must be 'minio'")
	}
	if c.Bucket == "" {
		return fmt.Errorf("CONTRACT_HOURLY_EVIDENCE_STORAGE_BUCKET is required")
	}
	if c.Endpoint == "" {
		return fmt.Errorf("CONTRACT_HOURLY_EVIDENCE_STORAGE_ENDPOINT is required")
	}
	if c.Region == "" {
		return fmt.Errorf("CONTRACT_HOURLY_EVIDENCE_STORAGE_REGION is required")
	}
	if c.AccessKey == "" {
		return fmt.Errorf("CONTRACT_HOURLY_EVIDENCE_STORAGE_ACCESS_KEY is required")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("CONTRACT_HOURLY_EVIDENCE_STORAGE_SECRET_KEY is required")
	}
	if _, err := time.ParseDuration(c.PresignPutTTL); err != nil {
		return fmt.Errorf("invalid CONTRACT_HOURLY_EVIDENCE_STORAGE_PRESIGN_PUT_TTL: %w", err)
	}
	return nil
}
