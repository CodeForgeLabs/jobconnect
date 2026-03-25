package main

import (
	"bufio"
	"context"
	chatv1 "jobconnect/chat/gen/chat/v1"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	grpcadapter "jobconnect/chat/internal/adapters/grpc"
	"jobconnect/chat/internal/adapters/ws"
	applications "jobconnect/chat/internal/application"
	"jobconnect/chat/internal/config"
	"jobconnect/chat/internal/infrastructure/clock"
	"jobconnect/chat/internal/infrastructure/db"
	"jobconnect/chat/internal/infrastructure/tokens"

	"google.golang.org/grpc"
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

	pool, err := db.NewPool(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	chatRepo := db.NewChatRepo(pool)
	clockImpl := clock.NewRealClock()
	jwtParser := tokens.NewJWTParser(cfg.JWTSecret)

	createMessageUC := &applications.CreateMessage{Chats: chatRepo, Clock: clockImpl}
	getMessagesUC := &applications.GetMessages{Chats: chatRepo}
	markAsSeenUC := &applications.MarkAsSeen{Chats: chatRepo, Clock: clockImpl}
	editMessageUC := &applications.EditMessage{Chats: chatRepo, Clock: clockImpl}
	deleteMessageUC := &applications.DeleteMessage{Chats: chatRepo, Clock: clockImpl}
	getConversationUC := &applications.GetConversation{Chats: chatRepo}
	deleteConversationUC := &applications.DeleteConversation{Chats: chatRepo}

	// WEB SOCKET HUB
	// Create the Hub
	hub := ws.NewHub()

	// 1. Start WebSocket Server (HTTP)
	go func() {
		http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
			ws.ServeWS(hub, w, r)
		})
		log.Println("WebSocket server listening on :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
	chatServer := grpcadapter.NewChatServer(
		createMessageUC,
		getMessagesUC,
		markAsSeenUC,
		editMessageUC,
		deleteMessageUC,
		getConversationUC,
		deleteConversationUC,
		hub,
		jwtParser,
	)

	lis, err := net.Listen("tcp", cfg.GRPCListenAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	chatv1.RegisterChatServiceServer(grpcServer, chatServer)

	log.Printf("Chat gRPC listening on %s", cfg.GRPCListenAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}

	log.Printf("Chat gRPC listening on %s", cfg.GRPCListenAddr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}

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
