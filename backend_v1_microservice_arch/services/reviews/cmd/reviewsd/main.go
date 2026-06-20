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

	reviewsv1 "jobconnect/reviews/gen/reviews/v1"

	grpcadapter "jobconnect/reviews/internal/adapters/grpc"
	"jobconnect/reviews/internal/applications"
	"jobconnect/reviews/internal/config"
	"jobconnect/reviews/internal/infrastructure/clock"
	"jobconnect/reviews/internal/infrastructure/db"

	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := loadDotEnv(
		"../../../../.env",
		"../../.env",
		".env",
	); err != nil {
		log.Fatalf("load .env: %v", err)
	}
	// Load env
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	contCfg, err := config.LoadFromEnvForContract()
	if err != nil {
		log.Fatalf("contract config: %v", err)
	}
	// DB
	pool, err := db.NewPool(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("db connection failed: %v", err)
	}
	defer pool.Close()

	contractPool, err := db.NewPool(ctx, contCfg.PostgresURL)
	if err != nil {
		log.Fatalf("contract db connection failed: %v", err)
	}
	defer contractPool.Close()

	reviewRepo := db.NewReviewRepo(pool, contractPool)
	clockImpl := clock.NewRealClock()

	// Use cases
	createUC := &applications.CreateReview{
		Reviews: reviewRepo,
		Clock:   clockImpl,
	}
	getUC := &applications.GetReview{
		Reviews: reviewRepo,
	}
	updateUC := &applications.UpdateReview{
		Reviews: reviewRepo,
		Clock:   clockImpl,
	}
	deleteUC := &applications.DeleteReview{
		Reviews: reviewRepo,
	}
	listUC := &applications.ListReviews{
		Reviews: reviewRepo,
	}
	getContractsUsersUC := &applications.GetContractUsers{
		Reviews: reviewRepo,
		Clock:   clockImpl,
	}
	getUserRatingSummaryUC := &applications.GetUserRatingSummary{
		Reviews: reviewRepo,
	}

	// gRPC server
	server := grpcadapter.NewReviewServer(
		createUC,
		deleteUC,
		getUC,
		listUC,
		updateUC,
		getContractsUsersUC,
		getUserRatingSummaryUC,
	)

	lis, err := net.Listen("tcp", cfg.GRPCListenAddr)
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}

	grpcServer := grpc.NewServer()
	reviewsv1.RegisterReviewServiceServer(grpcServer, server)

	go func() {
		log.Printf("reviews gRPC running on %s", cfg.GRPCListenAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("serve error: %v", err)
		}
	}()

	// graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh
	log.Println("shutdown requested...")

	gracefulStop(grpcServer)
}

func gracefulStop(srv *grpc.Server) {
	done := make(chan struct{})

	go func() {
		srv.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		log.Println("graceful shutdown complete")
	case <-time.After(5 * time.Second):
		log.Println("forced shutdown")
		srv.Stop()
	}
}

type Config struct {
	PostgresURL    string
	GRPCListenAddr string
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
