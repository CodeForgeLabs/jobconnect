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

	CapabilityMinSkillsForDiscovery        int
	CapabilityRequireVerifiedForWithdraw   bool
	CapabilityRequirePublicForDiscovery    bool
	CapabilityRequireHeadlineForFreelancer bool
	CapabilityRequireCompanyNameForClient  bool
	CapabilityAllowMessagingWhenSuspended  bool
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		GRPCListenAddr:                         getEnv("USER_GRPC_LISTEN_ADDR", ":50052"),
		PostgresURL:                            os.Getenv("USER_POSTGRES_URL"),
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
