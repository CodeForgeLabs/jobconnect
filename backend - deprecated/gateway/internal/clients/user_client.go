package clients

import (
	userv1 "jobconnect/user/gen/user"

	"google.golang.org/grpc"
)

func NewUserClient(conn *grpc.ClientConn) userv1.UserServiceClient {
	return userv1.NewUserServiceClient(conn)
}
