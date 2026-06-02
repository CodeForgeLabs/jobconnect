package clients

import (
	recommendationv1 "jobconnect/recommendation/gen/recommendation/v1"

	"google.golang.org/grpc"
)

func NewRecommendationClient(conn *grpc.ClientConn) recommendationv1.RecommendationServiceClient {
	return recommendationv1.NewRecommendationServiceClient(conn)
}
