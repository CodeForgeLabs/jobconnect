package main

import (
	"bufio"
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	walletv1 "jobconnect/dispute/gen/wallet/v1"
	grpcadapter "jobconnect/dispute/internal/adapters/grpc"
	"jobconnect/dispute/internal/application"
	"jobconnect/dispute/internal/config"
	"jobconnect/dispute/internal/infrastructure/db"
	"jobconnect/dispute/internal/infrastructure/tokens"
	"jobconnect/dispute/internal/infrastructure/walletgrpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := loadDotEnv(".env", "../../.env", "../../../.env"); err != nil {
		log.Fatalf("load .env: %v", err)
	}
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	pool, err := db.NewPool(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	walletConn, err := grpc.NewClient(cfg.WalletServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("wallet service dial: %v", err)
	}
	defer walletConn.Close()

	repo := db.NewRepo(pool)
	issuer := tokens.NewJWTIssuer(cfg.JWTSecret)
	walletClient := walletgrpc.NewClient(walletv1.NewWalletServiceClient(walletConn), issuer)
	app := &application.Service{
		Repo:   repo,
		Wallet: walletClient,
		Clock:  realClock{},
	}
	tokenParser := tokens.NewJWTParser(cfg.JWTSecret)
	server := grpcadapter.NewServer(app, tokenParser)

	lis, err := net.Listen("tcp", cfg.GRPCListenAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	grpcadapter.Register(grpcServer, server)

	go func() {
		log.Printf("dispute gRPC listening on %s", cfg.GRPCListenAddr)
		if serveErr := grpcServer.Serve(lis); serveErr != nil {
			log.Printf("serve: %v", serveErr)
			cancel()
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigCh:
	case <-ctx.Done():
	}

	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(10 * time.Second):
		grpcServer.Stop()
	}
}

func loadDotEnv(paths ...string) error {
	for _, path := range paths {
		if err := loadDotEnvFile(path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
	}
	return nil
}

func loadDotEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), "\"'")
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
	return scanner.Err()
}
