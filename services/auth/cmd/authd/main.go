package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcadapter "jobconnect/auth/internal/adapters/grpc"
	"jobconnect/auth/internal/config"
	"jobconnect/auth/internal/infrastructure/db"

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

	_ = pool // TODO: pass into repositories/use-cases

	lis, err := net.Listen("tcp", cfg.GRPCListenAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	grpcadapter.NewServer().Register(grpcServer)

	go func() {
		log.Printf("auth gRPC listening on %s", cfg.GRPCListenAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("serve: %v", err)
			cancel()
		}
	}()

	// Graceful shutdown.
	sigCh := make(chan os.Signal, 1)                    // channel which receives os.Signals as datatype and has buffer size of 1 which means it can only hold 1 signal at a time
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM) // signal.Notify is a function that registers the given channel to receive notifications of the specified signals
	select {                                            // select is used to wait for a signal to be received
	case <-sigCh:
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
