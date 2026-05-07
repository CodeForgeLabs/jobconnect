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
	EvidenceStore  VerificationEvidenceStorageConfig
}

type VerificationEvidenceStorageConfig struct {
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
		GRPCListenAddr: getEnv("VERIFICATION_GRPC_LISTEN_ADDR", ":50060"),
		PostgresURL:    os.Getenv("VERIFICATION_POSTGRES_URL"),
		EvidenceStore: VerificationEvidenceStorageConfig{
			Provider:      strings.ToLower(strings.TrimSpace(getEnv("VERIFICATION_EVIDENCE_STORAGE_PROVIDER", "minio"))),
			Bucket:        strings.TrimSpace(getEnv("VERIFICATION_EVIDENCE_STORAGE_BUCKET", "jobconnect-verification-evidence")),
			Endpoint:      strings.TrimSpace(getEnv("VERIFICATION_EVIDENCE_STORAGE_ENDPOINT", "localhost:9000")),
			Region:        strings.TrimSpace(getEnv("VERIFICATION_EVIDENCE_STORAGE_REGION", "us-east-1")),
			AccessKey:     strings.TrimSpace(getEnv("VERIFICATION_EVIDENCE_STORAGE_ACCESS_KEY", "minioadmin")),
			SecretKey:     strings.TrimSpace(getEnv("VERIFICATION_EVIDENCE_STORAGE_SECRET_KEY", "minioadmin")),
			UseSSL:        getEnvBool("VERIFICATION_EVIDENCE_STORAGE_USE_SSL", false),
			PathStyle:     getEnvBool("VERIFICATION_EVIDENCE_STORAGE_PATH_STYLE", true),
			CreateBucket:  getEnvBool("VERIFICATION_EVIDENCE_STORAGE_CREATE_BUCKET", true),
			PresignPutTTL: strings.TrimSpace(getEnv("VERIFICATION_EVIDENCE_STORAGE_PRESIGN_PUT_TTL", "15m")),
		},
	}
	if cfg.PostgresURL == "" {
		return Config{}, fmt.Errorf("VERIFICATION_POSTGRES_URL is required")
	}
	if err := cfg.EvidenceStore.Validate(); err != nil {
		return Config{}, err
	}
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

func (c VerificationEvidenceStorageConfig) Validate() error {
	if c.Provider != "minio" {
		return fmt.Errorf("VERIFICATION_EVIDENCE_STORAGE_PROVIDER must be 'minio'")
	}
	if c.Bucket == "" {
		return fmt.Errorf("VERIFICATION_EVIDENCE_STORAGE_BUCKET is required")
	}
	if c.Endpoint == "" {
		return fmt.Errorf("VERIFICATION_EVIDENCE_STORAGE_ENDPOINT is required")
	}
	if c.Region == "" {
		return fmt.Errorf("VERIFICATION_EVIDENCE_STORAGE_REGION is required")
	}
	if c.AccessKey == "" {
		return fmt.Errorf("VERIFICATION_EVIDENCE_STORAGE_ACCESS_KEY is required")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("VERIFICATION_EVIDENCE_STORAGE_SECRET_KEY is required")
	}
	if strings.TrimSpace(c.PresignPutTTL) == "" {
		return fmt.Errorf("VERIFICATION_EVIDENCE_STORAGE_PRESIGN_PUT_TTL is required")
	}
	return nil
}
