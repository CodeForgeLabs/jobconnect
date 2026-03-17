package main

import (
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	
	"jobconnect/recommendation/internal/application"
	adaptergrpc "jobconnect/recommendation/internal/adapters/grpc"
	pb "jobconnect/recommendation/gen/recommendation/v1"
	
	"jobconnect/recommendation/internal/infrastructure/jobgrpc"
	"jobconnect/recommendation/internal/infrastructure/usergrpc"
)

func main() {
	jobServiceAddr := os.Getenv("JOB_SERVICE_ADDR")
	if jobServiceAddr == "" {
		jobServiceAddr = "localhost:50053" // Default job service port
	}

	userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
	if userServiceAddr == "" {
		userServiceAddr = "localhost:50052" // Default user service port
	}

	// Connect to Job Service
	jobConn, err := grpc.NewClient(jobServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to job service: %v", err)
	}
	defer jobConn.Close()
	jobClient := jobgrpc.NewClient(jobConn)

	// Connect to User Service
	userConn, err := grpc.NewClient(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to user service: %v", err)
	}
	defer userConn.Close()
	userClient := usergrpc.NewClient(userConn)

	// Setup application service
	app := application.NewRecommendationService(jobClient, userClient)

	// Setup gRPC server
	server := adaptergrpc.NewServer(app)
	grpcServer := grpc.NewServer()
	pb.RegisterRecommendationServiceServer(grpcServer, server)

	// Listen and serve
	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("Recommendation service listening on :50055")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
