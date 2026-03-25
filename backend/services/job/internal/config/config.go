package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	GRPCListenAddr string
	PostgresURL    string
	JWTSecret      []byte
	AttachmentStorage AttachmentStorageConfig
}

type AttachmentStorageConfig struct {
	Provider     string
	Bucket       string
	Endpoint     string
	Region       string
	AccessKey    string
	SecretKey    string
	UseSSL       bool
	PathStyle    bool
	CreateBucket bool
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		GRPCListenAddr: getEnv("JOB_GRPC_LISTEN_ADDR", ":50053"),
		PostgresURL:    os.Getenv("JOB_POSTGRES_URL"),
		AttachmentStorage: AttachmentStorageConfig{
			Provider:     strings.ToLower(strings.TrimSpace(getEnv("JOB_ATTACHMENT_STORAGE_PROVIDER", "minio"))),
			Bucket:       strings.TrimSpace(getEnv("JOB_ATTACHMENT_STORAGE_BUCKET", "jobconnect-job-attachments")),
			Endpoint:     strings.TrimSpace(getEnv("JOB_ATTACHMENT_STORAGE_ENDPOINT", "localhost:9000")),
			Region:       strings.TrimSpace(getEnv("JOB_ATTACHMENT_STORAGE_REGION", "us-east-1")),
			AccessKey:    strings.TrimSpace(os.Getenv("JOB_ATTACHMENT_STORAGE_ACCESS_KEY")),
			SecretKey:    strings.TrimSpace(os.Getenv("JOB_ATTACHMENT_STORAGE_SECRET_KEY")),
			UseSSL:       getEnvBool("JOB_ATTACHMENT_STORAGE_USE_SSL", false),
			PathStyle:    getEnvBool("JOB_ATTACHMENT_STORAGE_PATH_STYLE", true),
			CreateBucket: getEnvBool("JOB_ATTACHMENT_STORAGE_CREATE_BUCKET", true),
		},
	}
	if cfg.PostgresURL == "" {
		return Config{}, fmt.Errorf("JOB_POSTGRES_URL is required")
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
		return fmt.Errorf("JOB_ATTACHMENT_STORAGE_PROVIDER must be 'minio'")
	}
	if c.Bucket == "" {
		return fmt.Errorf("JOB_ATTACHMENT_STORAGE_BUCKET is required")
	}
	if c.Endpoint == "" {
		return fmt.Errorf("JOB_ATTACHMENT_STORAGE_ENDPOINT is required")
	}
	if c.AccessKey == "" {
		return fmt.Errorf("JOB_ATTACHMENT_STORAGE_ACCESS_KEY is required")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("JOB_ATTACHMENT_STORAGE_SECRET_KEY is required")
	}
	if c.Region == "" {
		return fmt.Errorf("JOB_ATTACHMENT_STORAGE_REGION is required")
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
