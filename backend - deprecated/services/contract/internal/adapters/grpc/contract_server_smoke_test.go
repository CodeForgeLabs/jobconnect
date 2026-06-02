package grpcadapter

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	contractv1 "jobconnect/contract/gen/contract/v1"
	"jobconnect/contract/internal/application"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type smokeTokenParser struct{}

func (smokeTokenParser) ParseAccessToken(token string) (uuid.UUID, string, error) {
	switch token {
	case "client-token":
		return uuid.MustParse("11111111-1111-1111-1111-111111111111"), "client", nil
	case "freelancer-token":
		return uuid.MustParse("22222222-2222-2222-2222-222222222222"), "freelancer", nil
	case "admin-token":
		return uuid.MustParse("33333333-3333-3333-3333-333333333333"), "admin", nil
	case "service-token":
		return uuid.MustParse("00000000-0000-0000-0000-00000000c0de"), "service", nil
	default:
		return uuid.Nil, "", fmt.Errorf("invalid access token")
	}
}

func newSmokeServer() *ContractServer {
	return NewContractServer(
		&application.CreateContract{},
		&application.GetContract{},
		&application.ListMyContracts{},
		&application.GetJobOfferState{},
		&application.AcceptContract{},
		&application.DeclineContract{},
		&application.RevokeContractOffer{},
		&application.SubmitMilestoneWork{},
		&application.RequestMilestoneChanges{},
		&application.ApproveMilestoneSubmission{},
		&application.UpdateMilestoneStatus{},
		&application.LogHourlyWork{},
		&application.GetHourlyLogEvidenceUploadURL{},
		&application.ListHourlyLogs{},
		&application.GetHourlyWorkSummary{},
		&application.UpdateHourlyLog{},
		&application.DeleteHourlyLog{},
		&application.ReviewHourlyLog{},
		&application.GetHourlyInvoice{},
		&application.ListHourlyInvoices{},
		&application.InternalCloseHourlyWeek{},
		&application.InternalSettleHourlyInvoice{},
		&application.CreateContractBonus{},
		&application.ListContractBonuses{},
		&application.InternalMarkContractBonusPaid{},
		&application.ProposeAmendment{},
		&application.RespondAmendment{},
		&application.ListAmendments{},
		&application.PauseContract{},
		&application.ResumeContract{},
		&application.EndContract{},
		&application.GetStatusHistory{},
		smokeTokenParser{},
	)
}

func bearerCtx(token string) context.Context {
	return metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))
}

func internalCtx(caller string) context.Context {
	return metadata.NewOutgoingContext(context.Background(), metadata.Pairs(
		"x-jobconnect-internal", caller,
		"x-jobconnect-internal-secret", "smoke-secret",
	))
}

func TestContractRPCSmoke_AllRPCsReachHandler(t *testing.T) {
	t.Setenv("JOBCONNECT_INTERNAL_CALLER_SECRET", "smoke-secret")

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer lis.Close()

	grpcServer := grpc.NewServer()
	contractv1.RegisterContractServiceServer(grpcServer, newSmokeServer())
	go func() {
		_ = grpcServer.Serve(lis)
	}()
	defer grpcServer.Stop()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	client := contractv1.NewContractServiceClient(conn)

	cases := []struct {
		name string
		call func() error
	}{
		{"CreateContract", func() error {
			_, err := client.CreateContract(bearerCtx("client-token"), &contractv1.CreateContractRequest{
				FreelancerId: "22222222-2222-2222-2222-222222222222",
				JobId:        1,
				ProposalId:   1,
			})
			return err
		}},
		{"GetContract", func() error {
			_, err := client.GetContract(bearerCtx("client-token"), &contractv1.GetContractRequest{ContractId: 1})
			return err
		}},
		{"ListMyContracts", func() error {
			_, err := client.ListMyContracts(bearerCtx("client-token"), &contractv1.ListMyContractsRequest{})
			return err
		}},
		{"InternalGetJobOfferState", func() error {
			_, err := client.InternalGetJobOfferState(metadata.AppendToOutgoingContext(
				internalCtx("job-service"), "authorization", "Bearer client-token",
			), &contractv1.GetJobOfferStateRequest{JobId: 1})
			return err
		}},
		{"AcceptContract", func() error {
			_, err := client.AcceptContract(bearerCtx("freelancer-token"), &contractv1.AcceptContractRequest{ContractId: 1})
			return err
		}},
		{"DeclineContract", func() error {
			_, err := client.DeclineContract(bearerCtx("freelancer-token"), &contractv1.DeclineContractRequest{ContractId: 1})
			return err
		}},
		{"RevokeContractOffer", func() error {
			_, err := client.RevokeContractOffer(bearerCtx("client-token"), &contractv1.RevokeContractOfferRequest{ContractId: 1})
			return err
		}},
		{"SubmitMilestoneWork", func() error {
			_, err := client.SubmitMilestoneWork(bearerCtx("freelancer-token"), &contractv1.SubmitMilestoneWorkRequest{ContractId: 1, MilestoneId: 1})
			return err
		}},
		{"RequestMilestoneChanges", func() error {
			_, err := client.RequestMilestoneChanges(bearerCtx("client-token"), &contractv1.RequestMilestoneChangesRequest{ContractId: 1, MilestoneId: 1})
			return err
		}},
		{"ApproveMilestoneSubmission", func() error {
			_, err := client.ApproveMilestoneSubmission(bearerCtx("client-token"), &contractv1.ApproveMilestoneSubmissionRequest{ContractId: 1, MilestoneId: 1})
			return err
		}},
		{"InternalMarkMilestoneFunded", func() error {
			_, err := client.InternalMarkMilestoneFunded(internalCtx("payment-service"), &contractv1.InternalMarkMilestoneFundedRequest{ContractId: 1, MilestoneId: 1})
			return err
		}},
		{"LogHourlyWork", func() error {
			_, err := client.LogHourlyWork(bearerCtx("freelancer-token"), &contractv1.LogHourlyWorkRequest{ContractId: 1, StartAtUnixSeconds: 1, EndAtUnixSeconds: 2})
			return err
		}},
		{"GetHourlyLogEvidenceUploadUrl", func() error {
			_, err := client.GetHourlyLogEvidenceUploadUrl(bearerCtx("freelancer-token"), &contractv1.GetHourlyLogEvidenceUploadUrlRequest{
				ContractId:  1,
				FileName:    "proof.png",
				ContentType: "image/png",
			})
			return err
		}},
		{"ListHourlyLogs", func() error {
			_, err := client.ListHourlyLogs(bearerCtx("client-token"), &contractv1.ListHourlyLogsRequest{ContractId: 1})
			return err
		}},
		{"GetHourlyWorkSummary", func() error {
			_, err := client.GetHourlyWorkSummary(bearerCtx("client-token"), &contractv1.GetHourlyWorkSummaryRequest{ContractId: 1})
			return err
		}},
		{"UpdateHourlyLog", func() error {
			_, err := client.UpdateHourlyLog(bearerCtx("freelancer-token"), &contractv1.UpdateHourlyLogRequest{HourlyLogId: 1, StartAtUnixSeconds: 1, EndAtUnixSeconds: 2})
			return err
		}},
		{"DeleteHourlyLog", func() error {
			_, err := client.DeleteHourlyLog(bearerCtx("freelancer-token"), &contractv1.DeleteHourlyLogRequest{HourlyLogId: 1})
			return err
		}},
		{"ReviewHourlyLog", func() error {
			_, err := client.ReviewHourlyLog(bearerCtx("client-token"), &contractv1.ReviewHourlyLogRequest{HourlyLogId: 1, Status: contractv1.HourlyLogStatus_HOURLY_LOG_STATUS_APPROVED})
			return err
		}},
		{"GetHourlyInvoice", func() error {
			_, err := client.GetHourlyInvoice(bearerCtx("client-token"), &contractv1.GetHourlyInvoiceRequest{InvoiceId: 1})
			return err
		}},
		{"ListHourlyInvoices", func() error {
			_, err := client.ListHourlyInvoices(bearerCtx("client-token"), &contractv1.ListHourlyInvoicesRequest{ContractId: 1})
			return err
		}},
		{"InternalCloseHourlyWeek", func() error {
			_, err := client.InternalCloseHourlyWeek(internalCtx("scheduler-service"), &contractv1.InternalCloseHourlyWeekRequest{ContractId: 1})
			return err
		}},
		{"InternalSettleHourlyInvoice", func() error {
			_, err := client.InternalSettleHourlyInvoice(internalCtx("scheduler-service"), &contractv1.InternalSettleHourlyInvoiceRequest{InvoiceId: 1})
			return err
		}},
		{"CreateContractBonus", func() error {
			_, err := client.CreateContractBonus(bearerCtx("client-token"), &contractv1.CreateContractBonusRequest{ContractId: 1, AmountMinor: 100})
			return err
		}},
		{"ListContractBonuses", func() error {
			_, err := client.ListContractBonuses(bearerCtx("client-token"), &contractv1.ListContractBonusesRequest{ContractId: 1})
			return err
		}},
		{"InternalMarkContractBonusPaid", func() error {
			ctx := metadata.AppendToOutgoingContext(internalCtx("payment-service"), "x-payment-reference-id", "smoke-ref")
			_, err := client.InternalMarkContractBonusPaid(ctx, &contractv1.InternalMarkContractBonusPaidRequest{BonusId: 1})
			return err
		}},
		{"ProposeAmendment", func() error {
			_, err := client.ProposeAmendment(bearerCtx("client-token"), &contractv1.ProposeAmendmentRequest{ContractId: 1, Summary: "scope update", Payload: &contractv1.AmendmentPayload{}})
			return err
		}},
		{"RespondAmendment", func() error {
			_, err := client.RespondAmendment(bearerCtx("freelancer-token"), &contractv1.RespondAmendmentRequest{AmendmentId: 1, Status: contractv1.AmendmentStatus_AMENDMENT_STATUS_REJECTED, ResponseNote: "not now"})
			return err
		}},
		{"ListAmendments", func() error {
			_, err := client.ListAmendments(bearerCtx("client-token"), &contractv1.ListAmendmentsRequest{ContractId: 1})
			return err
		}},
		{"PauseContract", func() error {
			_, err := client.PauseContract(bearerCtx("client-token"), &contractv1.PauseContractRequest{ContractId: 1})
			return err
		}},
		{"ResumeContract", func() error {
			_, err := client.ResumeContract(bearerCtx("client-token"), &contractv1.ResumeContractRequest{ContractId: 1})
			return err
		}},
		{"EndContract", func() error {
			_, err := client.EndContract(bearerCtx("client-token"), &contractv1.EndContractRequest{ContractId: 1})
			return err
		}},
		{"GetStatusHistory", func() error {
			_, err := client.GetStatusHistory(bearerCtx("client-token"), &contractv1.GetStatusHistoryRequest{ContractId: 1})
			return err
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call()
			if err == nil {
				return
			}
			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("expected grpc status, got: %v", err)
			}
			if st.Code() == codes.Unimplemented {
				t.Fatalf("rpc returned unimplemented: %v", err)
			}
			if st.Code() == codes.Unknown {
				t.Fatalf("rpc returned unknown (possible panic/path issue): %v", err)
			}
		})
	}
}

func TestMain(m *testing.M) {
	os.Setenv("JOBCONNECT_INTERNAL_CALLER_SECRET", "smoke-secret")
	os.Exit(m.Run())
}

