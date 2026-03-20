package clients

import (
	authv1 "jobconnect/auth/gen/auth/v1"

	"google.golang.org/grpc"
)

func NewAuthClient(conn *grpc.ClientConn) authv1.AuthServiceClient {
	return authv1.NewAuthServiceClient(conn)
}
