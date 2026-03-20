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

	grpcadapter "jobconnect/auth/internal/adapters/grpc"
	"jobconnect/auth/internal/application"
	"jobconnect/auth/internal/config"
	"jobconnect/auth/internal/infrastructure/clock"
	"jobconnect/auth/internal/infrastructure/db"
	"jobconnect/auth/internal/infrastructure/email"
	"jobconnect/auth/internal/infrastructure/hasher"
	"jobconnect/auth/internal/infrastructure/tokens"
	"jobconnect/auth/internal/infrastructure/usergrpc"
	userv1 "jobconnect/user/gen/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const tosVersion, privacyVersion = "1.0", "1.0"

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
	emailChangeRepo := db.NewEmailChangeRepo(pool)
	oauthIdentityRepo := db.NewOAuthIdentityRepo(pool)

	// Infrastructure
	hasherImpl := hasher.NewArgon2Hasher()
	clockImpl := clock.NewRealClock()
	tokenIssuer := tokens.NewJWTIssuer(cfg.JWTSecret)
	var emailSender application.EmailSender = email.NewNoopSender()
	if cfg.SMTPHost != "" {
		emailSender = email.NewSMTPSender(
			cfg.SMTPHost,
			cfg.SMTPPort,
			cfg.SMTPTLSMode,
			cfg.SMTPUsername,
			cfg.SMTPPassword,
			cfg.SMTPFromAddress,
			cfg.SMTPFromName,
		)
		log.Printf("smtp email sender enabled host=%s port=%d tls_mode=%s from=%s", cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPTLSMode, cfg.SMTPFromAddress)
	} else {
		log.Printf("smtp email sender disabled; using noop sender")
	}

	userConn, err := grpc.NewClient(cfg.UserServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("user service dial: %v", err)
	}
	defer userConn.Close()

	userClient := userv1.NewUserServiceClient(userConn)
	userProfiles := usergrpc.NewProfileClient(userClient)

	// Use-cases
	registerUC := &application.RegisterUser{
		Users:          userRepo,
		Creds:          credRepo,
		OTPs:           otpRepo,
		TOS:            tosRepo,
		UserProfiles:   userProfiles,
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
	logoutUC := &application.Logout{
		Sessions: sessionRepo,
		Tokens:   tokenIssuer,
	}
	forgotPasswordUC := &application.ForgotPassword{
		Users:     userRepo,
		OTPs:      otpRepo,
		Hasher:    hasherImpl,
		Clock:     clockImpl,
		EmailSend: emailSender,
		OTPTTL:    cfg.OTPTTL,
	}
	resetPasswordUC := &application.ResetPassword{
		Users:    userRepo,
		Creds:    credRepo,
		OTPs:     otpRepo,
		Sessions: sessionRepo,
		Hasher:   hasherImpl,
	}
	requestEmailChangeUC := &application.RequestEmailChange{
		Users:          userRepo,
		EmailChanges:   emailChangeRepo,
		Hasher:         hasherImpl,
		Clock:          clockImpl,
		EmailSend:      emailSender,
		EmailOTPExpiry: cfg.OTPTTL,
	}
	confirmEmailChangeUC := &application.ConfirmEmailChange{
		Users:        userRepo,
		EmailChanges: emailChangeRepo,
		Hasher:       hasherImpl,
		Clock:        clockImpl,
	}
	oauthLoginUC := &application.OAuthLogin{
		Users:        userRepo,
		Identities:   oauthIdentityRepo,
		Sessions:     sessionRepo,
		Tokens:       tokenIssuer,
		Clock:        clockImpl,
		UserProfiles: userProfiles,
		AccessTTL:    cfg.AccessTokenTTL,
		RefreshTTL:   cfg.RefreshTokenTTL,
	}
	listSessionsUC := &application.ListSessions{Sessions: sessionRepo}
	revokeSessionUC := &application.RevokeSession{Sessions: sessionRepo, Clock: clockImpl}

	authServer := grpcadapter.NewAuthServer(
		registerUC,
		verifyOTPUC,
		loginUC,
		refreshUC,
		logoutUC,
		forgotPasswordUC,
		resetPasswordUC,
		requestEmailChangeUC,
		confirmEmailChangeUC,
		oauthLoginUC,
		listSessionsUC,
		revokeSessionUC,
	)

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
