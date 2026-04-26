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

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "jobconnect/recommendation/gen/recommendation/v1"
	adaptergrpc "jobconnect/recommendation/internal/adapters/grpc"
	"jobconnect/recommendation/internal/application"
	"jobconnect/recommendation/internal/config"
	"jobconnect/recommendation/internal/infrastructure/cache"
	"jobconnect/recommendation/internal/infrastructure/jobgrpc"
	"jobconnect/recommendation/internal/infrastructure/reviewgrpc"
	"jobconnect/recommendation/internal/infrastructure/usergrpc"
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

	jobConn, err := grpc.NewClient(cfg.JobServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("job service dial: %v", err)
	}
	defer jobConn.Close()

	userConn, err := grpc.NewClient(cfg.UserServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("user service dial: %v", err)
	}
	defer userConn.Close()

	reviewConn, err := grpc.NewClient(cfg.ReviewServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("review service dial: %v", err)
	}
	defer reviewConn.Close()

	recommendationCache, closeCache := buildRecommendationCache(cfg)
	defer closeCache()

	app := application.NewRecommendationService(
		jobgrpc.NewClient(jobConn),
		usergrpc.NewClient(userConn),
		reviewgrpc.NewClient(reviewConn),
		recommendationCache,
		application.ServiceConfig{
			DefaultLimit:      cfg.DefaultRecommendationLimit,
			MaxLimit:          cfg.MaxRecommendationLimit,
			CandidatePageSize: cfg.CandidatePageSize,
			PerSkillPageSize:  cfg.PerSkillPageSize,
			MaxSkillQueries:   cfg.MaxSkillQueries,
		},
	)

	server := adaptergrpc.NewServer(app)
	grpcServer := grpc.NewServer()
	pb.RegisterRecommendationServiceServer(grpcServer, server)

	lis, err := net.Listen("tcp", cfg.GRPCListenAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	go func() {
		log.Printf("recommendation gRPC listening on %s", cfg.GRPCListenAddr)
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

func buildRecommendationCache(cfg config.Config) (application.RecommendationCache, func()) {
	switch cfg.RecommendationCacheBackend {
	case "redis":
		redisCache := cache.NewRedisCache(cache.RedisConfig{
			Addr:     cfg.RecommendationRedisAddr,
			Password: cfg.RecommendationRedisPassword,
			DB:       cfg.RecommendationRedisDB,
			TTL:      cfg.RecommendationCacheTTL,
		})
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := redisCache.Ping(ctx); err != nil {
			log.Fatalf("redis cache ping: %v", err)
		}
		log.Printf("recommendation cache backend: redis addr=%s db=%d ttl=%s", cfg.RecommendationRedisAddr, cfg.RecommendationRedisDB, cfg.RecommendationCacheTTL)
		return redisCache, func() {
			if err := redisCache.Close(); err != nil {
				log.Printf("redis cache close: %v", err)
			}
		}
	default:
		log.Printf("recommendation cache backend: memory ttl=%s", cfg.RecommendationCacheTTL)
		return cache.NewMemoryCache(cfg.RecommendationCacheTTL), func() {}
	}
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
