package config

import "testing"

func TestLoadFromEnv_Minimal(t *testing.T) {
	t.Setenv("AUTH_POSTGRES_URL", "postgres://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("AUTH_JWT_SECRET", "secret")
	// Leave TTL envs unset to use defaults.

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv error: %v", err)
	}
	if cfg.PostgresURL == "" {
		t.Fatalf("expected PostgresURL to be set")
	}
	if len(cfg.JWTSecret) == 0 {
		t.Fatalf("expected JWTSecret to be set")
	}
}

func TestLoadFromEnv_SMTPOptionalWhenHostMissing(t *testing.T) {
	t.Setenv("AUTH_POSTGRES_URL", "postgres://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("AUTH_JWT_SECRET", "secret")
	t.Setenv("AUTH_SMTP_HOST", "")
	t.Setenv("AUTH_SMTP_FROM_ADDRESS", "")

	if _, err := LoadFromEnv(); err != nil {
		t.Fatalf("expected no error when SMTP host is not set, got: %v", err)
	}
}

func TestLoadFromEnv_SMTPRequiresFromAddress(t *testing.T) {
	t.Setenv("AUTH_POSTGRES_URL", "postgres://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("AUTH_JWT_SECRET", "secret")
	t.Setenv("AUTH_SMTP_HOST", "smtp.example.com")
	t.Setenv("AUTH_SMTP_PORT", "587")
	t.Setenv("AUTH_SMTP_FROM_ADDRESS", "")

	if _, err := LoadFromEnv(); err == nil {
		t.Fatalf("expected error when SMTP host is set but from address is missing")
	}
}

func TestLoadFromEnv_SMTPConfigLoaded(t *testing.T) {
	t.Setenv("AUTH_POSTGRES_URL", "postgres://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("AUTH_JWT_SECRET", "secret")
	t.Setenv("AUTH_SMTP_HOST", "smtp.example.com")
	t.Setenv("AUTH_SMTP_PORT", "2525")
	t.Setenv("AUTH_SMTP_TLS_MODE", "implicit")
	t.Setenv("AUTH_SMTP_USERNAME", "mailer")
	t.Setenv("AUTH_SMTP_PASSWORD", "mailer-pass")
	t.Setenv("AUTH_SMTP_FROM_ADDRESS", "no-reply@example.com")
	t.Setenv("AUTH_SMTP_FROM_NAME", "JobConnect")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv error: %v", err)
	}
	if cfg.SMTPHost != "smtp.example.com" {
		t.Fatalf("unexpected SMTPHost: %s", cfg.SMTPHost)
	}
	if cfg.SMTPPort != 2525 {
		t.Fatalf("unexpected SMTPPort: %d", cfg.SMTPPort)
	}
	if cfg.SMTPTLSMode != "implicit" {
		t.Fatalf("unexpected SMTPTLSMode: %s", cfg.SMTPTLSMode)
	}
	if cfg.SMTPFromAddress != "no-reply@example.com" {
		t.Fatalf("unexpected SMTPFromAddress: %s", cfg.SMTPFromAddress)
	}
	if cfg.SMTPFromName != "JobConnect" {
		t.Fatalf("unexpected SMTPFromName: %s", cfg.SMTPFromName)
	}
}

func TestLoadFromEnv_InvalidSMTPTLSMode(t *testing.T) {
	t.Setenv("AUTH_POSTGRES_URL", "postgres://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("AUTH_JWT_SECRET", "secret")
	t.Setenv("AUTH_SMTP_HOST", "smtp.example.com")
	t.Setenv("AUTH_SMTP_PORT", "587")
	t.Setenv("AUTH_SMTP_TLS_MODE", "badmode")
	t.Setenv("AUTH_SMTP_FROM_ADDRESS", "no-reply@example.com")

	if _, err := LoadFromEnv(); err == nil {
		t.Fatalf("expected error for invalid AUTH_SMTP_TLS_MODE")
	}
}
