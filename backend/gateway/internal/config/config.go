package config

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPListenAddr      string
	AuthServiceGRPCAddr string
	JWTSecret           []byte

	RefreshCookieName     string
	RefreshCookieDomain   string
	RefreshCookieSecure   bool
	RefreshCookieHTTPOnly bool
	RefreshCookiePath     string
	RefreshCookieSameSite http.SameSite
	RefreshCookieMaxAge   time.Duration
}

func LoadFromEnv() (Config, error) {
	secret := os.Getenv("AUTH_JWT_SECRET")
	if secret == "" {
		return Config{}, fmt.Errorf("AUTH_JWT_SECRET is required")
	}

	cfg := Config{
		HTTPListenAddr:        getEnv("GATEWAY_HTTP_LISTEN_ADDR", ":8080"),
		AuthServiceGRPCAddr:   getEnv("AUTH_SERVICE_GRPC_ADDR", "auth:50051"),
		JWTSecret:             []byte(secret),
		RefreshCookieName:     getEnv("GATEWAY_REFRESH_COOKIE_NAME", "jc_refresh_token"),
		RefreshCookieDomain:   os.Getenv("GATEWAY_REFRESH_COOKIE_DOMAIN"),
		RefreshCookieSecure:   getEnvBool("GATEWAY_REFRESH_COOKIE_SECURE", false),
		RefreshCookieHTTPOnly: getEnvBool("GATEWAY_REFRESH_COOKIE_HTTP_ONLY", true),
		RefreshCookiePath:     getEnv("GATEWAY_REFRESH_COOKIE_PATH", "/"),
		RefreshCookieSameSite: parseSameSite(getEnv("GATEWAY_REFRESH_COOKIE_SAME_SITE", "lax")),
		RefreshCookieMaxAge:   getEnvDurationSeconds("AUTH_REFRESH_TOKEN_TTL_SECONDS", 30*24*60*60),
	}

	if cfg.AuthServiceGRPCAddr == "" {
		return Config{}, fmt.Errorf("AUTH_SERVICE_GRPC_ADDR is required")
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

func parseSameSite(v string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	case "default":
		return http.SameSiteDefaultMode
	default:
		return http.SameSiteLaxMode
	}
}
