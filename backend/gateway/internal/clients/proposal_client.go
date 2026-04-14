package clients

import (
	proposalv1 "jobconnect/proposal/gen/proposal/v1"

	"google.golang.org/grpc"
)

func NewProposalClient(conn *grpc.ClientConn) proposalv1.ProposalServiceClient {
	return proposalv1.NewProposalServiceClient(conn)
}
