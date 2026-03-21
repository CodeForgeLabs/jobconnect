package clients

import (
	verificationv1 "jobconnect/verification/gen/verification/v1"

	"google.golang.org/grpc"
)

func NewVerificationClient(conn *grpc.ClientConn) verificationv1.VerificationServiceClient {
	return verificationv1.NewVerificationServiceClient(conn)
}
