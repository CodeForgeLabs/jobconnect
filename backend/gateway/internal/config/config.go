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
	HTTPListenAddr                string
	AuthServiceGRPCAddr           string
	UserServiceGRPCAddr           string
	VerificationServiceGRPCAddr   string
	JobServiceGRPCAddr            string
	ProposalServiceGRPCAddr       string
	ContractServiceGRPCAddr       string
	RecommendationServiceGRPCAddr string
	JWTSecret                     []byte
	OAuthStateSecret              []byte
	ChallengeProofSecret          []byte
	ChallengeProofTTL             time.Duration
	RecaptchaSecretKey            string
	RecaptchaMinScore             float64
	RecaptchaDevBypass            bool
	RecaptchaBypassToken          string
	ChatServiceGRPCAddr           string
	OAuthGoogleClientID           string
	OAuthGoogleClientSecret       string
	OAuthGoogleRedirectURI        string

	OAuthGitHubClientID     string
	OAuthGitHubClientSecret string
	OAuthGitHubRedirectURI  string

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
		HTTPListenAddr:                getEnv("GATEWAY_HTTP_LISTEN_ADDR", ":8080"),
		AuthServiceGRPCAddr:           getEnv("AUTH_SERVICE_GRPC_ADDR", "auth:50051"),
		UserServiceGRPCAddr:           getEnv("USER_SERVICE_GRPC_ADDR", "user:50052"),
		VerificationServiceGRPCAddr:   getEnv("VERIFICATION_SERVICE_GRPC_ADDR", "verification:50060"),
		JobServiceGRPCAddr:            getEnv("JOB_SERVICE_GRPC_ADDR", "job:50053"),
		ChatServiceGRPCAddr:           getEnv("CHAT_SERVICE_GRPC_ADDR", "chat:50054"),
		ProposalServiceGRPCAddr:       getEnv("PROPOSAL_SERVICE_GRPC_ADDR", "proposal:50054"),
		ContractServiceGRPCAddr:       getEnv("CONTRACT_SERVICE_GRPC_ADDR", "contract:50055"),
		RecommendationServiceGRPCAddr: getEnv("RECOMMENDATION_SERVICE_GRPC_ADDR", "recommendation:50064"),
		JWTSecret:                     []byte(secret),
		OAuthStateSecret:              []byte(getEnv("GATEWAY_OAUTH_STATE_SECRET", secret)),
		ChallengeProofSecret:          []byte(getEnv("GATEWAY_CHALLENGE_PROOF_SECRET", secret)),
		ChallengeProofTTL:             getEnvDurationSeconds("GATEWAY_CHALLENGE_PROOF_TTL_SECONDS", 120),
		RecaptchaSecretKey:            os.Getenv("GATEWAY_RECAPTCHA_SECRET_KEY"),
		RecaptchaMinScore:             getEnvFloat("GATEWAY_RECAPTCHA_MIN_SCORE", 0.5),
		RecaptchaDevBypass:            getEnvBool("GATEWAY_RECAPTCHA_DEV_BYPASS", false),
		RecaptchaBypassToken:          getEnv("GATEWAY_RECAPTCHA_BYPASS_TOKEN", "dev-human"),
		OAuthGoogleClientID:           os.Getenv("GATEWAY_OAUTH_GOOGLE_CLIENT_ID"),
		OAuthGoogleClientSecret:       os.Getenv("GATEWAY_OAUTH_GOOGLE_CLIENT_SECRET"),
		OAuthGoogleRedirectURI:        os.Getenv("GATEWAY_OAUTH_GOOGLE_REDIRECT_URI"),
		OAuthGitHubClientID:           os.Getenv("GATEWAY_OAUTH_GITHUB_CLIENT_ID"),
		OAuthGitHubClientSecret:       os.Getenv("GATEWAY_OAUTH_GITHUB_CLIENT_SECRET"),
		OAuthGitHubRedirectURI:        os.Getenv("GATEWAY_OAUTH_GITHUB_REDIRECT_URI"),
		RefreshCookieName:             getEnv("GATEWAY_REFRESH_COOKIE_NAME", "jc_refresh_token"),
		RefreshCookieDomain:           os.Getenv("GATEWAY_REFRESH_COOKIE_DOMAIN"),
		RefreshCookieSecure:           getEnvBool("GATEWAY_REFRESH_COOKIE_SECURE", false),
		RefreshCookieHTTPOnly:         getEnvBool("GATEWAY_REFRESH_COOKIE_HTTP_ONLY", true),
		RefreshCookiePath:             getEnv("GATEWAY_REFRESH_COOKIE_PATH", "/"),
		RefreshCookieSameSite:         parseSameSite(getEnv("GATEWAY_REFRESH_COOKIE_SAME_SITE", "lax")),
		RefreshCookieMaxAge:           getEnvDurationSeconds("AUTH_REFRESH_TOKEN_TTL_SECONDS", 30*24*60*60),
	}

	if cfg.AuthServiceGRPCAddr == "" {
		return Config{}, fmt.Errorf("AUTH_SERVICE_GRPC_ADDR is required")
	}
	if cfg.UserServiceGRPCAddr == "" {
		return Config{}, fmt.Errorf("USER_SERVICE_GRPC_ADDR is required")
	}
	if cfg.VerificationServiceGRPCAddr == "" {
		return Config{}, fmt.Errorf("VERIFICATION_SERVICE_GRPC_ADDR is required")
	}
	if cfg.JobServiceGRPCAddr == "" {
		return Config{}, fmt.Errorf("JOB_SERVICE_GRPC_ADDR is required")
	}
	if cfg.ProposalServiceGRPCAddr == "" {
		return Config{}, fmt.Errorf("PROPOSAL_SERVICE_GRPC_ADDR is required")
	}
	if cfg.ContractServiceGRPCAddr == "" {
		return Config{}, fmt.Errorf("CONTRACT_SERVICE_GRPC_ADDR is required")
	}
	if cfg.RecommendationServiceGRPCAddr == "" {
		return Config{}, fmt.Errorf("RECOMMENDATION_SERVICE_GRPC_ADDR is required")
	}
	if cfg.ChatServiceGRPCAddr == "" {
		return Config{}, fmt.Errorf("CHAT_SERVICE_GRPC_ADDR is required")
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

func getEnvFloat(key string, def float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return n
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
