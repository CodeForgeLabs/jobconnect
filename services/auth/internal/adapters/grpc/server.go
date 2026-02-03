package grpcadapter

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct{}

func NewServer() *Server { return &Server{} }

// Register wires handlers into the given gRPC server.
func (s *Server) Register(grpcServer *grpc.Server) {
	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, hs)

	// Local dev convenience.
	reflection.Register(grpcServer)

	// TODO: Register auth.v1.AuthServiceServer once proto generation is set up.
}
