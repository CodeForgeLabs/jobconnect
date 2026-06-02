package grpcadapter

import (
	verificationv1 "jobconnect/verification/gen/verification/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	verification verificationv1.VerificationServiceServer
}

func NewServer(verification verificationv1.VerificationServiceServer) *Server {
	return &Server{verification: verification}
}

func (s *Server) Register(grpcServer *grpc.Server) {
	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, hs)

	reflection.Register(grpcServer)

	if s.verification != nil {
		verificationv1.RegisterVerificationServiceServer(grpcServer, s.verification)
	}
}
