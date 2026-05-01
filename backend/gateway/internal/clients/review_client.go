package clients

import (
	reviewsv1 "jobconnect/reviews/gen/reviews/v1"

	"google.golang.org/grpc"
)

func NewReviewClient(conn *grpc.ClientConn) reviewsv1.ReviewServiceClient {
	return reviewsv1.NewReviewServiceClient(conn)
}
