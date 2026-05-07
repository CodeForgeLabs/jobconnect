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

	sharedevents "jobconnect/events"
	grpcadapter "jobconnect/review/internal/adapters/grpc"
	"jobconnect/review/internal/application"
	"jobconnect/review/internal/config"
	"jobconnect/review/internal/infrastructure/clock"
	"jobconnect/review/internal/infrastructure/db"
	eventsinfra "jobconnect/review/internal/infrastructure/events"
	"jobconnect/review/internal/infrastructure/tokens"

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

	reviewRepo := db.NewReviewRepo(pool)
	clockImpl := clock.NewRealClock()
	jwtParser := tokens.NewJWTParser(cfg.JWTSecret)
	kafkaPublisher := sharedevents.NewPublisher(sharedevents.ParseBrokers(os.Getenv("KAFKA_BROKERS")), getEnv("KAFKA_TOPIC_REVIEW", "review.events"))
	defer kafkaPublisher.Close()
	reviewEvents := eventsinfra.NewReviewPublisher(kafkaPublisher)

	createReviewUC := &application.CreateReview{Reviews: reviewRepo, Clock: clockImpl, Events: reviewEvents}
	getReviewUC := &application.GetReview{Reviews: reviewRepo}
	listByUserUC := &application.ListReviewsByUser{Reviews: reviewRepo}
	listByContractUC := &application.ListReviewsByContract{Reviews: reviewRepo}
	ratingSummaryUC := &application.GetUserRatingSummary{Reviews: reviewRepo}
	updateReviewUC := &application.UpdateReview{Reviews: reviewRepo, Clock: clockImpl, Events: reviewEvents}
	deleteReviewUC := &application.DeleteReview{Reviews: reviewRepo, Clock: clockImpl, Events: reviewEvents}
	replyToReviewUC := &application.ReplyToReview{Reviews: reviewRepo, Clock: clockImpl}

	reviewServer := grpcadapter.NewReviewServer(
		createReviewUC,
		getReviewUC,
		listByUserUC,
		listByContractUC,
		ratingSummaryUC,
		updateReviewUC,
		deleteReviewUC,
		replyToReviewUC,
		jwtParser,
	)

	lis, err := net.Listen("tcp", cfg.GRPCListenAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	grpcadapter.NewServer(reviewServer).Register(grpcServer)

	go func() {
		log.Printf("review gRPC listening on %s", cfg.GRPCListenAddr)
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

	gracefulStop(grpcServer)
}

func gracefulStop(srv *grpc.Server) {
	stopped := make(chan struct{})
	go func() {
		srv.GracefulStop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		srv.Stop()
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

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
