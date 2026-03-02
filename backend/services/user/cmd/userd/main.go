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

	grpcadapter "jobconnect/user/internal/adapter/grpc"
	"jobconnect/user/internal/application"
	"jobconnect/user/internal/config"
	"jobconnect/user/internal/infrastructure/clock"
	"jobconnect/user/internal/infrastructure/db"
	"jobconnect/user/internal/infrastructure/media"

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

	profileRepo := db.NewProfileRepo(pool)
	clockImpl := clock.NewRealClock()

	createProfileUC := &application.CreateProfile{
		Profiles: profileRepo,
		Clock:    clockImpl,
	}
	getProfileUC := &application.GetProfile{Profiles: profileRepo}
	updateProfileUC := &application.UpdateProfile{Profiles: profileRepo, Clock: clockImpl}
	deleteProfileUC := &application.DeleteProfile{Profiles: profileRepo, Clock: clockImpl}
	getOnboardingStatusUC := &application.GetOnboardingStatus{Profiles: profileRepo}
	uploadAvatarUC := &application.UploadAvatar{
		Profiles:  profileRepo,
		Processor: media.NewAvatarProcessor(),
		Moderator: media.NewBasicAvatarModerator(),
		Clock:     clockImpl,
	}
	getAvatarUC := &application.GetAvatar{Profiles: profileRepo}
	removeAvatarUC := &application.RemoveAvatar{Profiles: profileRepo}

	userServer := grpcadapter.NewUserServer(
		createProfileUC,
		getProfileUC,
		updateProfileUC,
		deleteProfileUC,
		getOnboardingStatusUC,
		uploadAvatarUC,
		getAvatarUC,
		removeAvatarUC,
	)

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
