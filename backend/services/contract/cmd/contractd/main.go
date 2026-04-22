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

	userv1 "jobconnect/contract/gen/user"
	grpcadapter "jobconnect/contract/internal/adapters/grpc"
	"jobconnect/contract/internal/application"
	"jobconnect/contract/internal/config"
	"jobconnect/contract/internal/infrastructure/clock"
	"jobconnect/contract/internal/infrastructure/db"
	"jobconnect/contract/internal/infrastructure/jobgrpc"
	"jobconnect/contract/internal/infrastructure/proposalgrpc"
	"jobconnect/contract/internal/infrastructure/tokens"
	"jobconnect/contract/internal/infrastructure/usergrpc"

	jobv1 "jobconnect/job/gen/job/v1"
	proposalv1 "jobconnect/proposal/gen/proposal/v1"

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

	proposalConn, err := grpc.NewClient(cfg.ProposalServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("proposal service dial: %v", err)
	}
	defer proposalConn.Close()

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

	repo := db.NewContractRepo(pool)
	clockImpl := clock.NewRealClock()
	jwtParser := tokens.NewJWTParser(cfg.JWTSecret)
	jwtIssuer := tokens.NewJWTIssuer(cfg.JWTSecret)
	proposalClient := proposalgrpc.NewProposalClient(proposalv1.NewProposalServiceClient(proposalConn), jwtIssuer)
	jobClient := jobgrpc.NewJobClient(jobv1.NewJobServiceClient(jobConn), jwtIssuer)
	userPolicy := usergrpc.NewClient(userv1.NewUserServiceClient(userConn))

	createUC := &application.CreateContract{Contracts: repo, Proposals: proposalClient, Jobs: jobClient, Actors: userPolicy, Clock: clockImpl}
	getUC := &application.GetContract{Contracts: repo}
	listUC := &application.ListMyContracts{Contracts: repo}
	getJobOfferStateUC := &application.GetJobOfferState{Contracts: repo}
	acceptUC := &application.AcceptContract{Contracts: repo, Proposals: proposalClient, Jobs: jobClient, Actors: userPolicy, Clock: clockImpl}
	declineUC := &application.DeclineContract{Contracts: repo, Proposals: proposalClient, Clock: clockImpl}
	revokeUC := &application.RevokeContractOffer{Contracts: repo, Proposals: proposalClient, Actors: userPolicy, Clock: clockImpl}
	updateMilestoneStatusUC := &application.UpdateMilestoneStatus{Contracts: repo, Clock: clockImpl}
	submitMilestoneWorkUC := &application.SubmitMilestoneWork{UpdateMilestoneStatus: updateMilestoneStatusUC}
	requestMilestoneChangesUC := &application.RequestMilestoneChanges{UpdateMilestoneStatus: updateMilestoneStatusUC}
	approveMilestoneSubmissionUC := &application.ApproveMilestoneSubmission{UpdateMilestoneStatus: updateMilestoneStatusUC}
	logHourlyWorkUC := &application.LogHourlyWork{Contracts: repo, Clock: clockImpl}
	listHourlyLogsUC := &application.ListHourlyLogs{Contracts: repo}
	reviewHourlyLogUC := &application.ReviewHourlyLog{Contracts: repo, Clock: clockImpl}
	proposeAmendmentUC := &application.ProposeAmendment{Contracts: repo, Clock: clockImpl}
	respondAmendmentUC := &application.RespondAmendment{Contracts: repo, Clock: clockImpl}
	listAmendmentsUC := &application.ListAmendments{Contracts: repo, Clock: clockImpl}
	pauseUC := &application.PauseContract{Contracts: repo, Clock: clockImpl}
	resumeUC := &application.ResumeContract{Contracts: repo, Clock: clockImpl}
	endUC := &application.EndContract{Contracts: repo, Clock: clockImpl}
	getStatusHistoryUC := &application.GetStatusHistory{Contracts: repo}

	contractServer := grpcadapter.NewContractServer(
		createUC,
		getUC,
		listUC,
		getJobOfferStateUC,
		acceptUC,
		declineUC,
		revokeUC,
		submitMilestoneWorkUC,
		requestMilestoneChangesUC,
		approveMilestoneSubmissionUC,
		updateMilestoneStatusUC,
		logHourlyWorkUC,
		listHourlyLogsUC,
		reviewHourlyLogUC,
		proposeAmendmentUC,
		respondAmendmentUC,
		listAmendmentsUC,
		pauseUC,
		resumeUC,
		endUC,
		getStatusHistoryUC,
		jwtParser,
	)

	lis, err := net.Listen("tcp", cfg.GRPCListenAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	grpcadapter.NewServer(contractServer).Register(grpcServer)

	go func() {
		log.Printf("contract gRPC listening on %s", cfg.GRPCListenAddr)
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
