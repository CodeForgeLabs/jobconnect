package grpcadapter

import (
	chatv1 "jobconnect/reviews/gen/reviews/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	chat chatv1.ReviewServiceServer
}

func NewServer(chat chatv1.ReviewServiceServer) *Server {
	return &Server{chat: chat}
}

func (s *Server) Register(grpcServer *grpc.Server) {
	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, hs)

	reflection.Register(grpcServer)

	if s.chat != nil {
		chatv1.RegisterReviewServiceServer(grpcServer, s.chat)
	}
}
