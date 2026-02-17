package grpcadapter

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	authv1 "jobconnect/auth/gen/auth/v1"
)

type Server struct {
	auth authv1.AuthServiceServer
}

func NewServer(auth authv1.AuthServiceServer) *Server {
	return &Server{auth: auth}
}

// Register wires handlers into the given gRPC server.
func (s *Server) Register(grpcServer *grpc.Server) {
	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, hs)

	reflection.Register(grpcServer)

	if s.auth != nil {
		authv1.RegisterAuthServiceServer(grpcServer, s.auth)
	}
}
