package grpcadapter

import (
	proposalv1 "jobconnect/proposal/gen/proposal/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	proposal proposalv1.ProposalServiceServer
}

func NewServer(proposal proposalv1.ProposalServiceServer) *Server {
	return &Server{proposal: proposal}
}

func (s *Server) Register(grpcServer *grpc.Server) {
	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, hs)

	reflection.Register(grpcServer)

	if s.proposal != nil {
		proposalv1.RegisterProposalServiceServer(grpcServer, s.proposal)
	}
}
