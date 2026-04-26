package clients

import (
	disputev1 "jobconnect/dispute/gen/dispute/v1"

	"google.golang.org/grpc"
)

func NewDisputeClient(conn *grpc.ClientConn) disputev1.DisputeServiceClient {
	return disputev1.NewDisputeServiceClient(conn)
}
