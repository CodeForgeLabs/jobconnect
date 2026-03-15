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

	grpcadapter "jobconnect/wallet/internal/adapters/grpc"
	"jobconnect/wallet/internal/application"
	"jobconnect/wallet/internal/config"
	"jobconnect/wallet/internal/infrastructure/clock"
	"jobconnect/wallet/internal/infrastructure/db"
	"jobconnect/wallet/internal/infrastructure/tokens"

	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := loadDotEnv(".env"); err != nil && !os.IsNotExist(err) {
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

	repo := db.NewWalletRepo(pool)
	clockImpl := clock.NewRealClock()
	jwtParser := tokens.NewJWTParser(cfg.JWTSecret)

	createWalletUC := &application.CreateWallet{Wallets: repo}
	getWalletUC := &application.GetWallet{Wallets: repo}
	getBalanceUC := &application.GetBalance{Wallets: repo}
	creditWalletInternalUC := &application.CreditWalletInternal{Wallets: repo}
	debitWalletInternalUC := &application.DebitWalletInternal{Wallets: repo}
	placeHoldUC := &application.PlaceHold{Wallets: repo, Clock: clockImpl}
	releaseHoldUC := &application.ReleaseHold{Wallets: repo}
	captureHoldUC := &application.CaptureHold{Wallets: repo}
	listTransactionsUC := &application.ListTransactions{Wallets: repo}

	walletServer := grpcadapter.NewWalletServer(
		createWalletUC,
		getWalletUC,
		getBalanceUC,
		creditWalletInternalUC,
		debitWalletInternalUC,
		placeHoldUC,
		releaseHoldUC,
		captureHoldUC,
		listTransactionsUC,
		jwtParser,
	)

	lis, err := net.Listen("tcp", cfg.GRPCListenAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	grpcadapter.NewServer(walletServer).Register(grpcServer)

	go func() {
		log.Printf("wallet gRPC listening on %s", cfg.GRPCListenAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("serve: %v", err)
			cancel()
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigCh:
		log.Printf("shutdown requested")
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

func loadDotEnv(path string) error {
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
		val = strings.TrimSpace(val)
		val = strings.Trim(val, "\"'")
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}

	return scanner.Err()
}
