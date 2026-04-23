package grpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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
	recs, err := s.app.GetRecommendedJobs(forwardAuthMetadata(ctx), req.UserId, req.Limit)
	if err != nil {
		return nil, toStatusError(err)
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
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	if req.JobId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "job_id is required")
	}

	recs, err := s.app.GetRecommendedFreelancers(forwardAuthMetadata(ctx), req.JobId, req.Limit, authCacheScope(ctx))
	if err != nil {
		return nil, toStatusError(err)
	}

	pbRecs := make([]*pb.FreelancerRecommendation, 0, len(recs))
	for _, r := range recs {
		pbRecs = append(pbRecs, &pb.FreelancerRecommendation{
			UserId:      r.UserID,
			MatchScore:  r.MatchScore,
			MatchReason: r.MatchReason,
		})
	}

	return &pb.GetRecommendedFreelancersResponse{Recommendations: pbRecs}, nil
}

// forwardAuthMetadata copies the caller's incoming authorization header onto
// the outgoing context so downstream gRPC calls (user/job/review) carry the
// same credentials.
func forwardAuthMetadata(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	authz := md.Get("authorization")
	if len(authz) == 0 {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "authorization", authz[0])
}

func authCacheScope(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	authz := md.Get("authorization")
	if len(authz) == 0 {
		return ""
	}
	sum := sha256.Sum256([]byte(authz[0]))
	return hex.EncodeToString(sum[:8])
}

func toStatusError(err error) error {
	if st, ok := status.FromError(err); ok {
		return status.Error(st.Code(), err.Error())
	}
	return status.Error(codes.Internal, err.Error())
}
