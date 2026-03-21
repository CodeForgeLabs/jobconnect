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
	AvatarStorage  AvatarStorageConfig

	CapabilityMinSkillsForDiscovery        int
	CapabilityRequireVerifiedForWithdraw   bool
	CapabilityRequirePublicForDiscovery    bool
	CapabilityRequireHeadlineForFreelancer bool
	CapabilityRequireCompanyNameForClient  bool
	CapabilityAllowMessagingWhenSuspended  bool
}

type AvatarStorageConfig struct {
	Provider   string
	Bucket     string
	Endpoint   string
	Region     string
	AccessKey  string
	SecretKey  string
	UseSSL     bool
	PathStyle  bool
	CreateBucket bool
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		GRPCListenAddr:                         getEnv("USER_GRPC_LISTEN_ADDR", ":50052"),
		PostgresURL:                            os.Getenv("USER_POSTGRES_URL"),
		AvatarStorage: AvatarStorageConfig{
			Provider:     strings.ToLower(strings.TrimSpace(getEnv("USER_AVATAR_STORAGE_PROVIDER", "minio"))),
			Bucket:       strings.TrimSpace(getEnv("USER_AVATAR_STORAGE_BUCKET", "jobconnect-avatars")),
			Endpoint:     strings.TrimSpace(getEnv("USER_AVATAR_STORAGE_ENDPOINT", "localhost:9000")),
			Region:       strings.TrimSpace(getEnv("USER_AVATAR_STORAGE_REGION", "us-east-1")),
			AccessKey:    strings.TrimSpace(os.Getenv("USER_AVATAR_STORAGE_ACCESS_KEY")),
			SecretKey:    strings.TrimSpace(os.Getenv("USER_AVATAR_STORAGE_SECRET_KEY")),
			UseSSL:       getEnvBool("USER_AVATAR_STORAGE_USE_SSL", false),
			PathStyle:    getEnvBool("USER_AVATAR_STORAGE_PATH_STYLE", true),
			CreateBucket: getEnvBool("USER_AVATAR_STORAGE_CREATE_BUCKET", true),
		},
		CapabilityMinSkillsForDiscovery:        getEnvInt("USER_CAP_MIN_SKILLS_DISCOVERY", 1),
		CapabilityRequireVerifiedForWithdraw:   getEnvBool("USER_CAP_REQUIRE_VERIFIED_WITHDRAW", true),
		CapabilityRequirePublicForDiscovery:    getEnvBool("USER_CAP_REQUIRE_PUBLIC_DISCOVERY", true),
		CapabilityRequireHeadlineForFreelancer: getEnvBool("USER_CAP_REQUIRE_HEADLINE_DISCOVERY", true),
		CapabilityRequireCompanyNameForClient:  getEnvBool("USER_CAP_REQUIRE_COMPANY_DISCOVERY", true),
		CapabilityAllowMessagingWhenSuspended:  getEnvBool("USER_CAP_ALLOW_MSG_SUSPENDED", false),
	}
	if cfg.PostgresURL == "" {
		return Config{}, fmt.Errorf("USER_POSTGRES_URL is required")
	}
	if cfg.CapabilityMinSkillsForDiscovery < 0 {
		cfg.CapabilityMinSkillsForDiscovery = 0
	}
	if err := cfg.AvatarStorage.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c AvatarStorageConfig) Validate() error {
	if c.Provider != "minio" {
		return fmt.Errorf("USER_AVATAR_STORAGE_PROVIDER must be 'minio'")
	}
	if c.Bucket == "" {
		return fmt.Errorf("USER_AVATAR_STORAGE_BUCKET is required")
	}
	if c.Endpoint == "" {
		return fmt.Errorf("USER_AVATAR_STORAGE_ENDPOINT is required")
	}
	if c.AccessKey == "" {
		return fmt.Errorf("USER_AVATAR_STORAGE_ACCESS_KEY is required")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("USER_AVATAR_STORAGE_SECRET_KEY is required")
	}
	if c.Region == "" {
		return fmt.Errorf("USER_AVATAR_STORAGE_REGION is required")
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

func getEnvInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	parsed, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return parsed
}
