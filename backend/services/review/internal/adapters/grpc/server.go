package grpcadapter

import (
	reviewv1 "jobconnect/review/gen/review/v1"

	"google.golang.org/grpc"
)

type Server struct {
	review *ReviewServer
}

func NewServer(review *ReviewServer) *Server {
	return &Server{review: review}
}

func (s *Server) Register(srv *grpc.Server) {
	reviewv1.RegisterReviewServiceServer(srv, s.review)
}
