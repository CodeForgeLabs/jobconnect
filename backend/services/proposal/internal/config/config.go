package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	GRPCListenAddr string
	PostgresURL    string
	JWTSecret      []byte
	JobServiceAddr string
	AttachmentStorage AttachmentStorageConfig
}

type AttachmentStorageConfig struct {
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
	PresignGetTTL string
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		GRPCListenAddr: getEnv("PROPOSAL_GRPC_LISTEN_ADDR", ":50054"),
		PostgresURL:    os.Getenv("PROPOSAL_POSTGRES_URL"),
		JobServiceAddr: os.Getenv("JOB_SERVICE_GRPC_ADDR"),
		AttachmentStorage: AttachmentStorageConfig{
			Provider:      strings.ToLower(strings.TrimSpace(getEnv("PROPOSAL_ATTACHMENT_STORAGE_PROVIDER", "minio"))),
			Bucket:        strings.TrimSpace(getEnv("PROPOSAL_ATTACHMENT_STORAGE_BUCKET", "jobconnect-proposal-attachments")),
			Endpoint:      strings.TrimSpace(getEnv("PROPOSAL_ATTACHMENT_STORAGE_ENDPOINT", "localhost:9000")),
			Region:        strings.TrimSpace(getEnv("PROPOSAL_ATTACHMENT_STORAGE_REGION", "us-east-1")),
			AccessKey:     strings.TrimSpace(os.Getenv("PROPOSAL_ATTACHMENT_STORAGE_ACCESS_KEY")),
			SecretKey:     strings.TrimSpace(os.Getenv("PROPOSAL_ATTACHMENT_STORAGE_SECRET_KEY")),
			UseSSL:        getEnvBool("PROPOSAL_ATTACHMENT_STORAGE_USE_SSL", false),
			PathStyle:     getEnvBool("PROPOSAL_ATTACHMENT_STORAGE_PATH_STYLE", true),
			CreateBucket:  getEnvBool("PROPOSAL_ATTACHMENT_STORAGE_CREATE_BUCKET", true),
			PresignPutTTL: strings.TrimSpace(getEnv("PROPOSAL_ATTACHMENT_STORAGE_PRESIGN_PUT_TTL", "15m")),
			PresignGetTTL: strings.TrimSpace(getEnv("PROPOSAL_ATTACHMENT_STORAGE_PRESIGN_GET_TTL", "30m")),
		},
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
	if err := cfg.AttachmentStorage.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c AttachmentStorageConfig) Validate() error {
	if c.Provider != "minio" {
		return fmt.Errorf("PROPOSAL_ATTACHMENT_STORAGE_PROVIDER must be 'minio'")
	}
	if c.Bucket == "" {
		return fmt.Errorf("PROPOSAL_ATTACHMENT_STORAGE_BUCKET is required")
	}
	if c.Endpoint == "" {
		return fmt.Errorf("PROPOSAL_ATTACHMENT_STORAGE_ENDPOINT is required")
	}
	if c.Region == "" {
		return fmt.Errorf("PROPOSAL_ATTACHMENT_STORAGE_REGION is required")
	}
	if c.AccessKey == "" {
		return fmt.Errorf("PROPOSAL_ATTACHMENT_STORAGE_ACCESS_KEY is required")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("PROPOSAL_ATTACHMENT_STORAGE_SECRET_KEY is required")
	}
	if _, err := time.ParseDuration(c.PresignPutTTL); err != nil {
		return fmt.Errorf("invalid PROPOSAL_ATTACHMENT_STORAGE_PRESIGN_PUT_TTL: %w", err)
	}
	if _, err := time.ParseDuration(c.PresignGetTTL); err != nil {
		return fmt.Errorf("invalid PROPOSAL_ATTACHMENT_STORAGE_PRESIGN_GET_TTL: %w", err)
	}
	return nil
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
