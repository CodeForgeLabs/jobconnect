package grpcadapter

import (
	walletv1 "jobconnect/wallet/gen/wallet/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	wallet walletv1.WalletServiceServer
}

func NewServer(wallet walletv1.WalletServiceServer) *Server {
	return &Server{wallet: wallet}
}

func (s *Server) Register(grpcServer *grpc.Server) {
	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, hs)

	reflection.Register(grpcServer)

	if s.wallet != nil {
		walletv1.RegisterWalletServiceServer(grpcServer, s.wallet)
	}
}
