package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"jobconnect/payment/internal/adapters/grpc"
	webhook "jobconnect/payment/internal/adapters/http"
	"jobconnect/payment/internal/application"
	"jobconnect/payment/internal/config"
	"jobconnect/payment/internal/infrastructure/clients"
	"jobconnect/payment/internal/infrastructure/db"
	"jobconnect/payment/internal/infrastructure/gateway"
	"jobconnect/payment/internal/infrastructure/storage"
	paymentv1 "jobconnect/payment/gen/payment/v1"

	grpc_server "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type RealClock struct{}
func (c RealClock) Now() time.Time { return time.Now() }

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize DB
	pool, err := db.NewPool(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer pool.Close()

	repo := db.NewPaymentRepo(pool)

	// 2. Initialize MinIO
	receiptStore, err := storage.NewReceiptStore(ctx, storage.MinioConfig{
		Endpoint:     cfg.MinioEndpoint,
		AccessKey:    cfg.MinioAccessKey,
		SecretKey:    cfg.MinioSecretKey,
		UseSSL:       cfg.MinioUseSSL,
		Bucket:       "payment-receipts",
		CreateBucket: true,
	})
	if err != nil {
		log.Fatalf("failed to init receipt store: %v", err)
	}

	// 3. Initialize Gateways
	chapaGW := gateway.NewChapaGateway(cfg.ChapaSecretKey)
	telebirrGW := gateway.NewTelebirrGateway(cfg.TelebirrAppKey, cfg.TelebirrAppID)

	gatewayFactory := func(provider string) application.PaymentGateway {
		if provider == "TELEBIRR" {
			return telebirrGW
		}
		return chapaGW
	}

	// 4. Initialize gRPC Clients
	walletClient, err := clients.NewWalletClient(cfg.WalletSvcAddr)
	if err != nil {
		log.Fatalf("failed to connect to wallet svc: %v", err)
	}
	defer walletClient.Close()

	contractClient, err := clients.NewContractClient(cfg.ContractSvcAddr)
	if err != nil {
		log.Fatalf("failed to connect to contract svc: %v", err)
	}
	defer contractClient.Close()

	verificationClient, err := clients.NewVerificationClient(cfg.VerificationSvcAddr)
	if err != nil {
		log.Fatalf("failed to connect to verification svc: %v", err)
	}
	defer verificationClient.Close()

	clock := RealClock{}

	// 5. Initialize Use Cases
	initiateDepositUseCase := &application.InitiateDeposit{
		Sessions: repo,
		Gateway:  gatewayFactory,
		Clock:    clock,
	}
	verifyDepositUseCase := &application.VerifyDeposit{
		Sessions: repo,
		Gateway:  gatewayFactory,
		Wallet:   walletClient,
		Contract: contractClient,
		Clock:    clock,
	}
	requestWithdrawalUseCase := &application.RequestWithdrawal{
		Sessions: repo,
		Gateway:  gatewayFactory,
		Wallet:   walletClient,
		Verification: verificationClient,
		Clock:    clock,
	}
	getSessionUseCase := &application.GetSession{
		Sessions: repo,
	}
	listSessionsUseCase := &application.ListSessions{
		Sessions: repo,
	}
	uploadReceiptUseCase := &application.UploadReceipt{
		Sessions: repo,
		Receipts: receiptStore,
		Clock:    clock,
	}

	// 6. Initialize gRPC Server
	grpcSrv := grpc.NewServer(
		initiateDepositUseCase,
		verifyDepositUseCase,
		requestWithdrawalUseCase,
		getSessionUseCase,
		listSessionsUseCase,
		uploadReceiptUseCase,
	)

	s := grpc_server.NewServer()
	paymentv1.RegisterPaymentServiceServer(s, grpcSrv)
	reflection.Register(s)

	go func() {
		lis, err := net.Listen("tcp", cfg.GRPCListenAddr)
		if err != nil {
			log.Fatalf("failed to listen on grpc loop: %v", err)
		}
		log.Printf("gRPC server listening at %v", lis.Addr())
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve grpc: %v", err)
		}
	}()

	// 7. Initialize HTTP Webhook Server
	webhookHandler := webhook.NewWebhookHandler(verifyDepositUseCase, chapaGW, telebirrGW)
	mux := http.NewServeMux()
	webhookHandler.RegisterRoutes(mux)

	httpSrv := &http.Server{
		Addr:    cfg.HTTPListenAddr,
		Handler: mux,
	}

	go func() {
		log.Printf("HTTP webhook server listening at %v", cfg.HTTPListenAddr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to serve http: %v", err)
		}
	}()

	// 8. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down servers...")

	s.GracefulStop()
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server shutdown failed: %v", err)
	}

	log.Println("Servers stopped cleanly")
}
