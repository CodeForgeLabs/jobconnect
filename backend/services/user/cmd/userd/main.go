package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcadapter "jobconnect/user/internal/adapter/grpc"
	"jobconnect/user/internal/application"
	"jobconnect/user/internal/config"
	"jobconnect/user/internal/infrastructure/clock"
	"jobconnect/user/internal/infrastructure/db"

	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	pool, err := db.NewPool(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	profileRepo := db.NewProfileRepo(pool)
	clockImpl := clock.NewRealClock()

	createProfileUC := &application.CreateProfile{
		Profiles: profileRepo,
		Clock:    clockImpl,
	}

	userServer := grpcadapter.NewUserServer(createProfileUC)

	lis, err := net.Listen("tcp", cfg.GRPCListenAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	grpcadapter.NewServer(userServer).Register(grpcServer)

	go func() {
		log.Printf("user gRPC listening on %s", cfg.GRPCListenAddr)
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
