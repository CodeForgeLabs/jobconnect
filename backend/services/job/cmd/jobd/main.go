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

	grpcadapter "jobconnect/job/internal/adapters/grpc"
	"jobconnect/job/internal/adapters/grpc/clients"
	"jobconnect/job/internal/application"
	"jobconnect/job/internal/config"
	"jobconnect/job/internal/infrastructure/clock"
	"jobconnect/job/internal/infrastructure/db"
	"jobconnect/job/internal/infrastructure/storage"
	"jobconnect/job/internal/infrastructure/tokens"

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

	jobRepo := db.NewJobRepo(pool)
	attachmentStore, err := storage.NewAttachmentStore(ctx, cfg.AttachmentStorage)
	if err != nil {
		log.Fatalf("attachment storage: %v", err)
	}
	clockImpl := clock.NewRealClock()
	jwtParser := tokens.NewJWTParser(cfg.JWTSecret)
	// Connects Client
	connectsAddr := os.Getenv("CONNECTS_SERVICE_ADDR")
	if connectsAddr == "" {
		connectsAddr = "localhost:50058"
	}
	connectsCli, err := clients.NewConnectsClient(connectsAddr)
	if err != nil {
		log.Fatalf("connects service dial: %v", err)
	}

	// Proposal Client
	proposalAddr := os.Getenv("PROPOSAL_SERVICE_ADDR")
	if proposalAddr == "" {
		proposalAddr = "localhost:50054"
	}
	proposalCli, err := clients.NewProposalClient(proposalAddr)
	if err != nil {
		log.Fatalf("proposal service dial: %v", err)
	}

	createJobUC := &application.CreateJob{Jobs: jobRepo, Clock: clockImpl}
	getJobUC := &application.GetJob{Jobs: jobRepo}
	updateJobUC := &application.UpdateJob{Jobs: jobRepo, Clock: clockImpl}
	listMyJobsUC := &application.ListMyJobs{Jobs: jobRepo}
	listOpenJobsUC := &application.ListOpenJobs{Jobs: jobRepo}
	closeJobUC := &application.CloseJob{Jobs: jobRepo, Proposals: proposalCli, Connects: connectsCli, Clock: clockImpl}
	uploadAttachmentUC := &application.UploadJobAttachment{Jobs: jobRepo, Storage: attachmentStore}
	deleteAttachmentUC := &application.DeleteJobAttachment{Jobs: jobRepo, Storage: attachmentStore}
	inviteFreelancerUC := &application.InviteFreelancerToJob{Jobs: jobRepo, Clock: clockImpl}
	listApplicantsUC := &application.ListJobApplicants{Jobs: jobRepo, Proposals: proposalCli}
	setApplicantUC := &application.SetApplicantStage{Proposals: proposalCli}
	setVisibilityUC := &application.SetJobVisibility{Jobs: jobRepo, Clock: clockImpl}
	setBudgetRangeUC := &application.SetJobBudgetRange{Jobs: jobRepo, Clock: clockImpl}
	setExperienceUC := &application.SetJobExperienceLevel{Jobs: jobRepo, Clock: clockImpl}
	pauseJobUC := &application.PauseJob{Jobs: jobRepo, Clock: clockImpl}
	reopenJobUC := &application.ReopenJob{Jobs: jobRepo, Clock: clockImpl}
	markFilledUC := &application.MarkJobFilled{Jobs: jobRepo, Clock: clockImpl}
	searchJobsUC := &application.SearchJobs{Jobs: jobRepo}
	listFacetsUC := &application.ListJobFacets{Jobs: jobRepo}
	listAttachmentsUC := &application.ListJobAttachments{Jobs: jobRepo}
	getAttachmentURLUC := &application.GetJobAttachmentDownloadURL{Jobs: jobRepo}

	jobServer := grpcadapter.NewJobServer(
		createJobUC,
		getJobUC,
		updateJobUC,
		listMyJobsUC,
		listOpenJobsUC,
		closeJobUC,
		uploadAttachmentUC,
		deleteAttachmentUC,
		inviteFreelancerUC,
		listApplicantsUC,
		setApplicantUC,
		setVisibilityUC,
		setBudgetRangeUC,
		setExperienceUC,
		pauseJobUC,
		reopenJobUC,
		markFilledUC,
		searchJobsUC,
		listFacetsUC,
		listAttachmentsUC,
		getAttachmentURLUC,
		jwtParser,
	)

	lis, err := net.Listen("tcp", cfg.GRPCListenAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	grpcadapter.NewServer(jobServer).Register(grpcServer)

	go func() {
		log.Printf("job gRPC listening on %s", cfg.GRPCListenAddr)
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
