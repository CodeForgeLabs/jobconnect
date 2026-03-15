package config

import "testing"

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("WALLET_POSTGRES_URL", "postgres://wallet:wallet@localhost:5432/wallet")
	t.Setenv("AUTH_JWT_SECRET", "secret")
	t.Setenv("WALLET_GRPC_LISTEN_ADDR", ":55000")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}
	if cfg.GRPCListenAddr != ":55000" {
		t.Fatalf("unexpected listen addr: %s", cfg.GRPCListenAddr)
	}
	if cfg.PostgresURL == "" {
		t.Fatalf("expected postgres url")
	}
	if len(cfg.JWTSecret) == 0 {
		t.Fatalf("expected jwt secret")
	}
}

func TestLoadFromEnv_DefaultListenAddr(t *testing.T) {
	t.Setenv("WALLET_POSTGRES_URL", "postgres://wallet:wallet@localhost:5432/wallet")
	t.Setenv("AUTH_JWT_SECRET", "secret")
	t.Setenv("WALLET_GRPC_LISTEN_ADDR", "")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}
	if cfg.GRPCListenAddr != ":50059" {
		t.Fatalf("unexpected default listen addr: %s", cfg.GRPCListenAddr)
	}
}
