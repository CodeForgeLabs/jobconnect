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
	"jobconnect/auth/internal/application"
	"jobconnect/auth/internal/config"
	"jobconnect/auth/internal/infrastructure/clock"
	"jobconnect/auth/internal/infrastructure/db"
	"jobconnect/auth/internal/infrastructure/email"
	"jobconnect/auth/internal/infrastructure/hasher"
	"jobconnect/auth/internal/infrastructure/tokens"

	"google.golang.org/grpc"
)

const tosVersion, privacyVersion = "1.0", "1.0"

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

	// Repositories
	userRepo := db.NewUserRepo(pool)
	credRepo := db.NewCredentialRepo(pool)
	otpRepo := db.NewOTPRepo(pool)
	sessionRepo := db.NewSessionRepo(pool)
	tosRepo := db.NewTOSRepo(pool)

	// Infrastructure
	hasherImpl := hasher.NewArgon2Hasher()
	clockImpl := clock.NewRealClock()
	tokenIssuer := tokens.NewJWTIssuer(cfg.JWTSecret)
	emailSender := email.NewNoopSender()

	// Use-cases
	registerUC := &application.RegisterUser{
		Users:          userRepo,
		Creds:          credRepo,
		OTPs:           otpRepo,
		TOS:            tosRepo,
		Hasher:         hasherImpl,
		Clock:          clockImpl,
		EmailSend:      emailSender,
		OTPTTL:         cfg.OTPTTL,
		TOSVersion:     tosVersion,
		PrivacyVersion: privacyVersion,
	}
	verifyOTPUC := &application.VerifyEmailOTP{
		Users:  userRepo,
		OTPs:   otpRepo,
		Hasher: hasherImpl,
		Clock:  clockImpl,
	}
	loginUC := &application.Login{
		Users:      userRepo,
		Creds:      credRepo,
		Sessions:   sessionRepo,
		Hasher:     hasherImpl,
		Tokens:     tokenIssuer,
		Clock:      clockImpl,
		AccessTTL:  cfg.AccessTokenTTL,
		RefreshTTL: cfg.RefreshTokenTTL,
	}
	refreshUC := &application.Refresh{
		Users:      userRepo,
		Sessions:   sessionRepo,
		Tokens:     tokenIssuer,
		Clock:      clockImpl,
		AccessTTL:  cfg.AccessTokenTTL,
		RefreshTTL: cfg.RefreshTokenTTL,
	}
	logoutUC := &application.LogoutEverywhere{
		Sessions: sessionRepo,
	}

	authServer := grpcadapter.NewAuthServer(registerUC, verifyOTPUC, loginUC, refreshUC, logoutUC)

	lis, err := net.Listen("tcp", cfg.GRPCListenAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	grpcadapter.NewServer(authServer).Register(grpcServer)

	go func() {
		log.Printf("auth gRPC listening on %s", cfg.GRPCListenAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("serve: %v", err)
			cancel()
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	select {
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
