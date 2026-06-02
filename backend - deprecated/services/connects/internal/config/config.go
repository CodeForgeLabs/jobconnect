package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	GrpcListenAddr string
	PostgresURL    string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found; relies on environment variables")
	}

	cfg := &Config{
		GrpcListenAddr: os.Getenv("CONNECTS_GRPC_LISTEN_ADDR"),
		PostgresURL:    os.Getenv("CONNECTS_POSTGRES_URL"),
	}

	if cfg.GrpcListenAddr == "" {
		cfg.GrpcListenAddr = ":50058" // Default port for connects
	}

	if cfg.PostgresURL == "" {
		log.Fatal("CONNECTS_POSTGRES_URL is required")
	}

	return cfg
}
