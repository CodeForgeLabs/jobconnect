package grpc

import (
	"context"

	"jobconnect/recommendation/internal/application"
	pb "jobconnect/recommendation/gen/recommendation/v1"
)

type Server struct {
	pb.UnimplementedRecommendationServiceServer
	app *application.RecommendationService
}

func NewServer(app *application.RecommendationService) *Server {
	return &Server{app: app}
}

func (s *Server) GetRecommendedJobs(ctx context.Context, req *pb.GetRecommendedJobsRequest) (*pb.GetRecommendedJobsResponse, error) {
	recs, err := s.app.GetRecommendedJobs(ctx, req.UserId, req.Limit)
	if err != nil {
		return nil, err
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
	recs, err := s.app.GetRecommendedFreelancers(ctx, req.JobId, req.Limit)
	if err != nil {
		return nil, err
	}
	
	var pbRecs []*pb.FreelancerRecommendation
	for _, r := range recs {
		pbRecs = append(pbRecs, &pb.FreelancerRecommendation{
			UserId:      r.UserID,
			MatchScore:  r.MatchScore,
			MatchReason: r.MatchReason,
		})
	}

	return &pb.GetRecommendedFreelancersResponse{Recommendations: pbRecs}, nil
}
