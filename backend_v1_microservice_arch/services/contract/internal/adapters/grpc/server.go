package grpcadapter

import (
	contractv1 "jobconnect/contract/gen/contract/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	contract contractv1.ContractServiceServer
}

func NewServer(contract contractv1.ContractServiceServer) *Server {
	return &Server{contract: contract}
}

func (s *Server) Register(grpcServer *grpc.Server) {
	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, hs)

	reflection.Register(grpcServer)

	if s.contract != nil {
		contractv1.RegisterContractServiceServer(grpcServer, s.contract)
	}
}
