package clients

import (
	jobv1 "jobconnect/job/gen/job/v1"

	"google.golang.org/grpc"
)

func NewJobClient(conn *grpc.ClientConn) jobv1.JobServiceClient {
	return jobv1.NewJobServiceClient(conn)
}
