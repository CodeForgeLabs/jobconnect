package main

import (
	"bufio"
	"context"
	"errors"
	"log"
	"net"
	"net/http"
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
	embedderpython "jobconnect/recommendation/internal/infrastructure/embedder/python"
	embeddingstorememory "jobconnect/recommendation/internal/infrastructure/embeddingstore/memory"
	embeddingstorepgvector "jobconnect/recommendation/internal/infrastructure/embeddingstore/pgvector"
	"jobconnect/recommendation/internal/infrastructure/jobgrpc"
	"jobconnect/recommendation/internal/infrastructure/metrics"
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

	recorder := metrics.NewPrometheusRecorder()

	recommendationCache, closeCache := buildRecommendationCache(cfg, recorder)
	defer closeCache()

	embedder, closeEmbedder := buildEmbedder(ctx, cfg)
	defer closeEmbedder()

	embeddingStore, closeEmbeddingStore := buildEmbeddingStore(ctx, cfg)
	defer closeEmbeddingStore()

	app := application.NewRecommendationService(
		jobgrpc.NewClient(jobConn),
		usergrpc.NewClient(userConn),
		reviewgrpc.NewClient(reviewConn),
		recommendationCache,
		recorder,
		embedder,
		embeddingStore,
		application.ServiceConfig{
			DefaultLimit:      cfg.DefaultRecommendationLimit,
			MaxLimit:          cfg.MaxRecommendationLimit,
			CandidatePageSize: cfg.CandidatePageSize,
			PerSkillPageSize:  cfg.PerSkillPageSize,
			MaxSkillQueries:   cfg.MaxSkillQueries,
		},
	)

	metricsServer := startMetricsServer(cfg.MetricsListenAddr, recorder)
	defer shutdownMetricsServer(metricsServer)

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

func buildRecommendationCache(cfg config.Config, recorder cache.MetricsRecorder) (application.RecommendationCache, func()) {
	switch cfg.RecommendationCacheBackend {
	case "redis":
		redisCache := cache.NewRedisCache(cache.RedisConfig{
			Addr:     cfg.RecommendationRedisAddr,
			Password: cfg.RecommendationRedisPassword,
			DB:       cfg.RecommendationRedisDB,
			TTL:      cfg.RecommendationCacheTTL,
			Metrics:  recorder,
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

func buildEmbeddingStore(ctx context.Context, cfg config.Config) (application.EmbeddingStore, func()) {
	switch cfg.EmbeddingStoreBackend {
	case "memory":
		log.Printf("recommendation embedding store backend: memory")
		return embeddingstorememory.New(), func() {}
	case "pgvector":
		poolCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		pool, err := embeddingstorepgvector.NewPool(poolCtx, cfg.PostgresURL)
		if err != nil {
			log.Fatalf("recommendation embedding store: pgvector pool init: %v", err)
		}
		log.Printf("recommendation embedding store backend: pgvector")
		return embeddingstorepgvector.New(pool), func() { pool.Close() }
	default:
		log.Printf("recommendation embedding store backend: noop (lazy embedding disabled)")
		return nil, func() {}
	}
}

func buildEmbedder(ctx context.Context, cfg config.Config) (application.Embedder, func()) {
	switch cfg.EmbedderBackend {
	case "python":
		client := embedderpython.New(embedderpython.Config{
			PythonPath:       cfg.EmbedderPythonPath,
			WorkerScript:     cfg.EmbedderWorkerScript,
			ModelName:        cfg.EmbedderModel,
			SocketPath:       cfg.EmbedderSocketPath,
			BatchSize:        cfg.EmbedderBatchSize,
			OperationTimeout: cfg.EmbedderOperationTimeout,
			StartupTimeout:   cfg.EmbedderStartupTimeout,
		}, nil)
		if err := client.Start(ctx); err != nil {
			log.Printf("recommendation embedder: start failed, ranking will use token cosine fallback: %v", err)
			return nil, func() { _ = client.Close() }
		}
		log.Printf("recommendation embedder backend: python model=%s socket=%s", cfg.EmbedderModel, cfg.EmbedderSocketPath)
		return client, func() { _ = client.Close() }
	default:
		log.Printf("recommendation embedder backend: noop (semantic ranking disabled)")
		return nil, func() {}
	}
}

func startMetricsServer(addr string, recorder *metrics.PrometheusRecorder) *http.Server {
	if strings.TrimSpace(addr) == "" {
		log.Printf("recommendation metrics: disabled")
		return nil
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", recorder.Handler())
	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("recommendation metrics listening on %s/metrics", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("metrics server: %v", err)
		}
	}()
	return srv
}

func shutdownMetricsServer(srv *http.Server) {
	if srv == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("metrics server shutdown: %v", err)
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
