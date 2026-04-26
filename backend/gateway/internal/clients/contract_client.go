package clients

import (
	contractv1 "jobconnect/contract/gen/contract/v1"

	"google.golang.org/grpc"
)

func NewContractClient(conn *grpc.ClientConn) contractv1.ContractServiceClient {
	return contractv1.NewContractServiceClient(conn)
}
