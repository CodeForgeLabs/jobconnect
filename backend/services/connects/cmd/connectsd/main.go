package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "jobconnect/api/proto/connects/v1"
	grpcAdapter "jobconnect/services/connects/internal/adapters/grpc"
	"jobconnect/services/connects/internal/application"
	"jobconnect/services/connects/internal/config"
	"jobconnect/services/connects/internal/infrastructure/db"
	eventsinfra "jobconnect/services/connects/internal/infrastructure/events"
	sharedevents "jobconnect/events"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.LoadConfig()

	// 1. Setup Postgres Connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to parse postgres url: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping db: %v", err)
	}
	log.Println("Connected to Postgres")

	// 2. Setup Dependency Injection
	repo := db.NewPostgresConnectsRepository(pool)
	app := application.NewUseCases(repo)
	connectsServer := grpcAdapter.NewConnectsServer(app)
	authConsumer := eventsinfra.StartAuthConsumer(context.Background(), sharedevents.ParseBrokers(os.Getenv("KAFKA_BROKERS")), getEnv("KAFKA_TOPIC_AUTH", "auth.events"), app)
	defer authConsumer.Close()

	// 3. Setup gRPC Server
	lis, err := net.Listen("tcp", cfg.GrpcListenAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", cfg.GrpcListenAddr, err)
	}

	grpcServer := grpc.NewServer()
	
	pb.RegisterConnectsServiceServer(grpcServer, connectsServer)

	// Health check & reflection
	healthServer := health.NewServer()
	healthServer.SetServingStatus("connects.v1.ConnectsService", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	reflection.Register(grpcServer)

	// 4. Graceful Shutdown
	go func() {
		log.Printf("Connects service listening on %s", cfg.GrpcListenAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down gRPC server...")

	grpcServer.GracefulStop()
	log.Println("Server stopped")
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
