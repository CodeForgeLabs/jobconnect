package grpcadapter

import (
	disputev1 "jobconnect/dispute/gen/dispute/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func Register(grpcServer *grpc.Server, dispute disputev1.DisputeServiceServer) {
	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, hs)
	reflection.Register(grpcServer)
	if dispute != nil {
		disputev1.RegisterDisputeServiceServer(grpcServer, dispute)
	}
}
