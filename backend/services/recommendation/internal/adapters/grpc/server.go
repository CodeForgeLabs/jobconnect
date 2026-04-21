package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "jobconnect/recommendation/gen/recommendation/v1"
	"jobconnect/recommendation/internal/application"
)

type Server struct {
	pb.UnimplementedRecommendationServiceServer
	app *application.RecommendationService
}

func NewServer(app *application.RecommendationService) *Server {
	return &Server{app: app}
}

func (s *Server) GetRecommendedJobs(ctx context.Context, req *pb.GetRecommendedJobsRequest) (*pb.GetRecommendedJobsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	recs, err := s.app.GetRecommendedJobs(ctx, req.UserId, req.Limit)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var pbRecs []*pb.JobRecommendation
	for _, r := range recs {
		pbRecs = append(pbRecs, &pb.JobRecommendation{
			JobId:       r.JobID,
			MatchScore:  r.MatchScore,
			MatchReason: r.MatchReason,
		})
	}

	return &pb.GetRecommendedJobsResponse{Recommendations: pbRecs}, nil
}

func (s *Server) GetRecommendedFreelancers(ctx context.Context, req *pb.GetRecommendedFreelancersRequest) (*pb.GetRecommendedFreelancersResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetRecommendedFreelancers is not implemented in phase 1")
}
