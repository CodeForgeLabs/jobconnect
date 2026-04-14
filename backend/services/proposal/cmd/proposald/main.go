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

	grpcadapter "jobconnect/proposal/internal/adapters/grpc"
	grpcclients "jobconnect/proposal/internal/adapters/grpc/clients"
	"jobconnect/proposal/internal/application"
	"jobconnect/proposal/internal/config"
	"jobconnect/proposal/internal/infrastructure/clock"
	"jobconnect/proposal/internal/infrastructure/contractgrpc"
	"jobconnect/proposal/internal/infrastructure/db"
	"jobconnect/proposal/internal/infrastructure/jobgrpc"
	"jobconnect/proposal/internal/infrastructure/storage"
	"jobconnect/proposal/internal/infrastructure/tokens"

	contractv1 "jobconnect/contract/gen/contract/v1"
	jobv1 "jobconnect/job/gen/job/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	proposalRepo := db.NewProposalRepo(pool)
	clockImpl := clock.NewRealClock()
	jwtParser := tokens.NewJWTParser(cfg.JWTSecret)

	attachmentStore, err := storage.NewAttachmentStore(ctx, cfg.AttachmentStorage)
	if err != nil {
		log.Fatalf("proposal attachment store: %v", err)
	}
	putTTL, err := time.ParseDuration(cfg.AttachmentStorage.PresignPutTTL)
	if err != nil {
		log.Fatalf("invalid proposal attachment put ttl: %v", err)
	}
	getTTL, err := time.ParseDuration(cfg.AttachmentStorage.PresignGetTTL)
	if err != nil {
		log.Fatalf("invalid proposal attachment get ttl: %v", err)
	}

	jobConn, err := grpc.NewClient(cfg.JobServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("job service dial: %v", err)
	}
	defer jobConn.Close()

	jobs := jobgrpc.NewJobClient(jobv1.NewJobServiceClient(jobConn))

	contractConn, err := grpc.NewClient(cfg.ContractServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("contract service dial: %v", err)
	}
	defer contractConn.Close()
	contracts := contractgrpc.NewContractClient(contractv1.NewContractServiceClient(contractConn))

	// Connects Client
	connectsAddr := os.Getenv("CONNECTS_SERVICE_ADDR")
	if connectsAddr == "" {
		connectsAddr = "localhost:50058"
	}
	connectsCli, err := grpcclients.NewConnectsClient(connectsAddr)
	if err != nil {
		log.Fatalf("connects service dial: %v", err)
	}

	submitUC := &application.SubmitProposal{Proposals: proposalRepo, Jobs: jobs, Connects: connectsCli, Clock: clockImpl}
	modifyUC := &application.ModifyProposal{Proposals: proposalRepo, Clock: clockImpl}
	withdrawUC := &application.WithdrawProposal{Proposals: proposalRepo, Clock: clockImpl}
	getUC := &application.GetProposal{Proposals: proposalRepo}
	getMineByJobUC := &application.GetMyProposalForJob{Proposals: proposalRepo}
	hasAppliedUC := &application.HasAppliedToJob{Proposals: proposalRepo}
	hireUC := &application.HireProposal{Proposals: proposalRepo, Jobs: jobs, JobLifecycle: jobs, Contracts: contracts, Clock: clockImpl}
	attachmentUploadURLUC := &application.GetProposalAttachmentUploadURL{Proposals: proposalRepo, Store: attachmentStore, PutTTL: putTTL}
	attachmentDownloadURLUC := &application.GetProposalAttachmentDownloadURL{Proposals: proposalRepo, Store: attachmentStore, GetTTL: getTTL}
	listByJobUC := &application.ListProposalsByJob{Proposals: proposalRepo}
	listMineUC := &application.ListMyProposals{Proposals: proposalRepo}
	listClientUC := &application.ListClientProposals{Proposals: proposalRepo}
	countByJobUC := &application.CountProposalsByJob{Proposals: proposalRepo}
	countInboxUC := &application.CountClientProposalInbox{Proposals: proposalRepo}
	setStatusUC := &application.SetProposalStatus{Proposals: proposalRepo, Clock: clockImpl}
	internalHireUC := &application.InternalHireProposal{Proposals: proposalRepo, Clock: clockImpl}

	proposalServer := grpcadapter.NewProposalServer(
		submitUC,
		modifyUC,
		withdrawUC,
		getUC,
		getMineByJobUC,
		hasAppliedUC,
		hireUC,
		attachmentUploadURLUC,
		attachmentDownloadURLUC,
		listByJobUC,
		listMineUC,
		listClientUC,
		countByJobUC,
		countInboxUC,
		setStatusUC,
		internalHireUC,
		jwtParser,
	)

	lis, err := net.Listen("tcp", cfg.GRPCListenAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	grpcadapter.NewServer(proposalServer).Register(grpcServer)

	go func() {
		log.Printf("proposal gRPC listening on %s", cfg.GRPCListenAddr)
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
