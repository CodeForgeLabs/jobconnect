package main

import (
	"bufio"
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	grpcadapter "jobconnect/wallet/internal/adapters/grpc"
	"jobconnect/wallet/internal/application"
	"jobconnect/wallet/internal/config"
	"jobconnect/wallet/internal/infrastructure/chapa"
	"jobconnect/wallet/internal/infrastructure/db"
	"jobconnect/wallet/internal/infrastructure/tokens"

	"google.golang.org/grpc"
)

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
	chapaClient := &chapa.Client{
		SecretKey: cfg.ChapaSecretKey,
		BaseURL:   cfg.ChapaBaseURL,
	}

	pool, err := db.NewPool(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	repo := db.NewWalletRepo(pool)
	jwtParser := tokens.NewJWTParser(cfg.JWTSecret)

	// ================== USE CASES (NEW SIMPLIFIED MODEL) ==================

	createWalletUC := &application.CreateWallet{Wallets: repo}
	getWalletUC := &application.GetWallet{Wallets: repo}

	completeDepositUC := &application.CompleteDeposit{Wallets: repo}
	createDepositUC := &application.CreateDeposit{
		Wallets: repo,
		Chapa:   chapaClient, // The key and URL are now inside this client
	}
	listTransactionsUC := &application.GetTransaction{Wallets: repo}
	fetchWalletTransactionUc := &application.FetchWalletTransactions{Wallets: repo}
	// ================== GRPC SERVER ==================

	walletServer := grpcadapter.NewWalletServer(
		createWalletUC,
		getWalletUC,
		createDepositUC,
		completeDepositUC,
		listTransactionsUC,
		fetchWalletTransactionUc,
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
	go func() {
		http.HandleFunc("/webhook/chapa", walletServer.ChapaWebhook)

		log.Println("Webhook running on :8080")
		http.ListenAndServe(":8080", nil)
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

// ==================== ENV LOADER ====================

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
