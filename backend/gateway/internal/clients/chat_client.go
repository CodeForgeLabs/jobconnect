package clients

import (
	chatv1 "jobconnect/chat/gen/chat/v1"

	"google.golang.org/grpc"
)

func NewChatClient(conn *grpc.ClientConn) chatv1.ChatServiceClient {
	return chatv1.NewChatServiceClient(conn)
}
