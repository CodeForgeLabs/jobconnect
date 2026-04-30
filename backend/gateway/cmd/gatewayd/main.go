package main

import (
	"bufio"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "jobconnect/gateway/cmd/gatewayd/docs"
	"jobconnect/gateway/internal/clients"
	"jobconnect/gateway/internal/config"
	"jobconnect/gateway/internal/handlers"
	"jobconnect/gateway/internal/router"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @title JobConnect Gateway API
// @version 1.0
// @description API Gateway for JobConnect platform.
// @termsOfService http://example.com/terms/

// @contact.name API Support
// @contact.email support@jobconnect.com

// @license.name MIT
// @host localhost:8080

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := loadDotEnv(".env", "../.env", "../../.env"); err != nil {
		log.Fatalf("load .env: %v", err)
	}

	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	authConn, err := grpc.NewClient(cfg.AuthServiceGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("auth service dial: %v", err)
	}
	defer authConn.Close()

	userConn, err := grpc.NewClient(cfg.UserServiceGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("user service dial: %v", err)
	}
	defer userConn.Close()

	verificationConn, err := grpc.NewClient(cfg.VerificationServiceGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("verification service dial: %v", err)
	}
	defer verificationConn.Close()

	jobConn, err := grpc.NewClient(cfg.JobServiceGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("job service dial: %v", err)
	}
	defer jobConn.Close()

	proposalConn, err := grpc.NewClient(cfg.ProposalServiceGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("proposal service dial: %v", err)
	}
	defer proposalConn.Close()

	contractConn, err := grpc.NewClient(cfg.ContractServiceGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("contract service dial: %v", err)
	}
	defer contractConn.Close()
	disputeConn, err := grpc.NewClient(cfg.DisputeServiceGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dispute service dial: %v", err)
	}
	defer disputeConn.Close()

	recommendationConn, err := grpc.NewClient(cfg.RecommendationServiceGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("recommendation service dial: %v", err)
	}
	defer recommendationConn.Close()
	chatConn, err := grpc.NewClient(cfg.ChatServiceGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("chat service dial: %v", err)
	}
	defer chatConn.Close()

	authClient := clients.NewAuthClient(authConn)
	userClient := clients.NewUserClient(userConn)
	verificationClient := clients.NewVerificationClient(verificationConn)
	jobClient := clients.NewJobClient(jobConn)
	proposalClient := clients.NewProposalClient(proposalConn)
	contractClient := clients.NewContractClient(contractConn)
	disputeClient := clients.NewDisputeClient(disputeConn)
	recommendationClient := clients.NewRecommendationClient(recommendationConn)
	chatClient := clients.NewChatClient(chatConn)
	authHandler := handlers.NewAuthHandler(cfg, authClient)
	userHandler := handlers.NewUserHandler(userClient, verificationClient)
	verificationHandler := handlers.NewVerificationHandler(verificationClient)
	jobHandler := handlers.NewJobHandler(jobClient)
	proposalHandler := handlers.NewProposalHandler(proposalClient)
	contractHandler := handlers.NewContractHandler(contractClient, jobClient, proposalClient)
	disputeHandler := handlers.NewDisputeHandler(disputeClient)
	recommendationHandler := handlers.NewRecommendationHandler(recommendationClient)
	chatHandler := handlers.NewChatHandler(chatClient)
	engine := router.New(cfg, authHandler, verificationHandler, userHandler, jobHandler, proposalHandler, contractHandler, disputeHandler, recommendationHandler, chatHandler)

	httpServer := &http.Server{
		Addr:              cfg.HTTPListenAddr,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("gateway http listening on %s", cfg.HTTPListenAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("gateway serve: %v", err)
			cancel()
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigCh:
	case <-ctx.Done():
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("gateway graceful shutdown failed: %v", err)
		_ = httpServer.Close()
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
