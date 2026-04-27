package config

import (
	"os"
)

type Config struct {
	GRPCListenAddr string
	PostgresURL    string
	JWTSecret      []byte
	ChapaSecretKey string
	ChapaBaseURL   string
}

// func LoadFromEnv() (Config, error) {
// 	cfg := Config{
// 		GRPCListenAddr: getEnv("WALLET_GRPC_LISTEN_ADDR", ":50059"),
// 		PostgresURL:    os.Getenv("WALLET_POSTGRES_URL"),
// 	}
// 	if cfg.PostgresURL == "" {
// 		return Config{}, fmt.Errorf("WALLET_POSTGRES_URL is required")
// 	}
// 	secret := os.Getenv("AUTH_JWT_SECRET")
// 	if secret == "" {
// 		return Config{}, fmt.Errorf("AUTH_JWT_SECRET is required")
// 	}
// 	cfg.JWTSecret = []byte(secret)
// 	return cfg, nil
// }

func LoadFromEnv() (Config, error) {
	// Hardcoded for local testing
	// format: postgres://<user>:<password>@localhost:<port>/<dbname>?sslmode=disable
	postgresURL := "postgres://postgres:password@localhost:5432/wallet_db?sslmode=disable"

	cfg := Config{
		GRPCListenAddr: getEnv("WALLET_GRPC_LISTEN_ADDR", ":50059"),
		PostgresURL:    postgresURL,
		JWTSecret:      []byte("my-super-secret-test-key"),
		ChapaSecretKey: "CHASECK_TEST-YF6dUEjoX7ubK1DStKOKZjDz8kxW4jyo", // Use your actual test key
		ChapaBaseURL:   "https://api.chapa.co",
	}

	// Comment out the requirement checks for now while you test
	/*
	   if cfg.PostgresURL == "" {
	       return Config{}, fmt.Errorf("WALLET_POSTGRES_URL is required")
	   }
	*/

	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
