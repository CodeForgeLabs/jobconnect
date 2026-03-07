package grpcadapter

import (
	jobv1 "jobconnect/job/gen/job/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	job jobv1.JobServiceServer
}

func NewServer(job jobv1.JobServiceServer) *Server {
	return &Server{job: job}
}

func (s *Server) Register(grpcServer *grpc.Server) {
	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, hs)

	reflection.Register(grpcServer)

	if s.job != nil {
		jobv1.RegisterJobServiceServer(grpcServer, s.job)
	}
}
