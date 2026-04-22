package grpcadapter

import (
	"context"
	"strings"
	"time"

	contractv1 "jobconnect/contract/gen/contract/v1"
	"jobconnect/contract/internal/application"
	"jobconnect/contract/internal/domain"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ContractServer struct {
	contractv1.UnimplementedContractServiceServer

	CreateUC                  *application.CreateContract
	GetUC                     *application.GetContract
	ListUC                    *application.ListMyContracts
	GetJobOfferStateUC        *application.GetJobOfferState
	AcceptUC                  *application.AcceptContract
	DeclineUC                 *application.DeclineContract
	RevokeUC                  *application.RevokeContractOffer
	SubmitMilestoneWorkUC     *application.SubmitMilestoneWork
	RequestMilestoneChangesUC *application.RequestMilestoneChanges
	ApproveMilestoneUC        *application.ApproveMilestoneSubmission
	UpdateMilestoneStatusUC   *application.UpdateMilestoneStatus
	LogHourlyWorkUC           *application.LogHourlyWork
	ListHourlyLogsUC          *application.ListHourlyLogs
	ReviewHourlyLogUC         *application.ReviewHourlyLog
	ProposeAmendmentUC        *application.ProposeAmendment
	RespondAmendmentUC        *application.RespondAmendment
	ListAmendmentsUC          *application.ListAmendments
	PauseUC                   *application.PauseContract
	ResumeUC                  *application.ResumeContract
	EndUC                     *application.EndContract
	GetStatusHistoryUC        *application.GetStatusHistory

	TokenParser TokenParser
}

func NewContractServer(
	create *application.CreateContract,
	get *application.GetContract,
	list *application.ListMyContracts,
	getJobOfferState *application.GetJobOfferState,
	accept *application.AcceptContract,
	decline *application.DeclineContract,
	revoke *application.RevokeContractOffer,
	submitMilestoneWork *application.SubmitMilestoneWork,
	requestMilestoneChanges *application.RequestMilestoneChanges,
	approveMilestone *application.ApproveMilestoneSubmission,
	updateMilestoneStatus *application.UpdateMilestoneStatus,
	logHourlyWork *application.LogHourlyWork,
	listHourlyLogs *application.ListHourlyLogs,
	reviewHourlyLog *application.ReviewHourlyLog,
	proposeAmendment *application.ProposeAmendment,
	respondAmendment *application.RespondAmendment,
	listAmendments *application.ListAmendments,
	pause *application.PauseContract,
	resume *application.ResumeContract,
	end *application.EndContract,
	getStatusHistory *application.GetStatusHistory,
	tokenParser TokenParser,
) *ContractServer {
	return &ContractServer{
		CreateUC:                  create,
		GetUC:                     get,
		ListUC:                    list,
		GetJobOfferStateUC:        getJobOfferState,
		AcceptUC:                  accept,
		DeclineUC:                 decline,
		RevokeUC:                  revoke,
		SubmitMilestoneWorkUC:     submitMilestoneWork,
		RequestMilestoneChangesUC: requestMilestoneChanges,
		ApproveMilestoneUC:        approveMilestone,
		UpdateMilestoneStatusUC:   updateMilestoneStatus,
		LogHourlyWorkUC:           logHourlyWork,
		ListHourlyLogsUC:          listHourlyLogs,
		ReviewHourlyLogUC:         reviewHourlyLog,
		ProposeAmendmentUC:        proposeAmendment,
		RespondAmendmentUC:        respondAmendment,
		ListAmendmentsUC:          listAmendments,
		PauseUC:                   pause,
		ResumeUC:                  resume,
		EndUC:                     end,
		GetStatusHistoryUC:        getStatusHistory,
		TokenParser:               tokenParser,
	}
}

func (s *ContractServer) CreateContract(ctx context.Context, req *contractv1.CreateContractRequest) (*contractv1.CreateContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	freelancerID, err := uuid.Parse(req.GetFreelancerId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid freelancer_id")
	}

	out, err := s.CreateUC.Execute(ctx, application.CreateContractInput{
		ClientID:        callerID,
		FreelancerID:    freelancerID,
		JobID:           req.GetJobId(),
		ProposalID:      req.GetProposalId(),
		ContractType:    fromProtoType(req.GetContractType()),
		Title:           req.GetTitle(),
		Description:     req.GetDescription(),
		HourlyRate:      req.GetHourlyRate(),
		FixedTotal:      req.GetFixedTotal(),
		WeeklyHourLimit: req.GetWeeklyHourLimit(),
		Milestones:      fromProtoMilestones(req.GetMilestones()),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &contractv1.CreateContractResponse{Contract: toProtoContract(out.Contract)}, nil
}

func (s *ContractServer) GetContract(ctx context.Context, req *contractv1.GetContractRequest) (*contractv1.GetContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, _, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	out, err := s.GetUC.Execute(ctx, application.GetContractInput{ContractID: req.GetContractId(), ActorID: callerID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &contractv1.GetContractResponse{Contract: toProtoContract(out.Contract)}, nil
}

func (s *ContractServer) ListMyContracts(ctx context.Context, req *contractv1.ListMyContractsRequest) (*contractv1.ListMyContractsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, _, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	out, err := s.ListUC.Execute(ctx, application.ListMyContractsInput{
		ActorID:   callerID,
		Status:    fromProtoStatus(req.GetStatus()),
		PageSize:  req.GetPageSize(),
		PageToken: req.GetPageToken(),
	})
	if err != nil {
		return nil, toStatus(err)
	}

	items := make([]*contractv1.Contract, 0, len(out.Contracts))
	for _, c := range out.Contracts {
		items = append(items, toProtoContract(c))
	}
	return &contractv1.ListMyContractsResponse{Contracts: items, NextPageToken: out.NextPageToken}, nil
}

func (s *ContractServer) GetJobOfferState(ctx context.Context, req *contractv1.GetJobOfferStateRequest) (*contractv1.GetJobOfferStateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	if err := requireInternalCaller(ctx, "job-service"); err != nil {
		return nil, err
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.GetJobOfferStateUC.Execute(ctx, application.GetJobOfferStateInput{
		JobID:     req.GetJobId(),
		ClientID:  callerID,
		ActorRole: role,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &contractv1.GetJobOfferStateResponse{
		JobId:             out.State.JobID,
		HasPendingOffer:   out.State.HasPendingOffer,
		PendingContractId: out.State.PendingContractID,
		HasActiveContract: out.State.HasActiveContract,
		ActiveContractId:  out.State.ActiveContractID,
	}, nil
}

func (s *ContractServer) AcceptContract(ctx context.Context, req *contractv1.AcceptContractRequest) (*contractv1.AcceptContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireFreelancerRole(role); err != nil {
		return nil, err
	}
	out, err := s.AcceptUC.Execute(ctx, application.AcceptContractInput{ContractID: req.GetContractId(), FreelancerID: callerID})
	if err != nil {
		return nil, toStatus(err)
	}
	return &contractv1.AcceptContractResponse{Contract: toProtoContract(out.Contract)}, nil
}

func (s *ContractServer) DeclineContract(ctx context.Context, req *contractv1.DeclineContractRequest) (*contractv1.DeclineContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireFreelancerRole(role); err != nil {
		return nil, err
	}
	out, err := s.DeclineUC.Execute(ctx, application.DeclineContractInput{ContractID: req.GetContractId(), FreelancerID: callerID, Reason: req.GetReason()})
	if err != nil {
		return nil, toStatus(err)
	}
	return &contractv1.DeclineContractResponse{Contract: toProtoContract(out.Contract)}, nil
}

func (s *ContractServer) RevokeContractOffer(ctx context.Context, req *contractv1.RevokeContractOfferRequest) (*contractv1.RevokeContractOfferResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.RevokeUC.Execute(ctx, application.RevokeContractOfferInput{
		ContractID: req.GetContractId(),
		ClientID:   callerID,
		Reason:     req.GetReason(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &contractv1.RevokeContractOfferResponse{Contract: toProtoContract(out.Contract)}, nil
}

func (s *ContractServer) SubmitMilestoneWork(ctx context.Context, req *contractv1.SubmitMilestoneWorkRequest) (*contractv1.SubmitMilestoneWorkResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	out, err := s.SubmitMilestoneWorkUC.Execute(ctx, application.SubmitMilestoneWorkInput{
		ContractID:  req.GetContractId(),
		MilestoneID: req.GetMilestoneId(),
		ActorID:     callerID,
		ActorRole:   role,
		Note:        req.GetNote(),
		Attachments: req.GetAttachments(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &contractv1.SubmitMilestoneWorkResponse{Contract: toProtoContract(out.Contract)}, nil
}

func (s *ContractServer) RequestMilestoneChanges(ctx context.Context, req *contractv1.RequestMilestoneChangesRequest) (*contractv1.RequestMilestoneChangesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	out, err := s.RequestMilestoneChangesUC.Execute(ctx, application.RequestMilestoneChangesInput{
		ContractID:  req.GetContractId(),
		MilestoneID: req.GetMilestoneId(),
		ActorID:     callerID,
		ActorRole:   role,
		Note:        req.GetNote(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &contractv1.RequestMilestoneChangesResponse{Contract: toProtoContract(out.Contract)}, nil
}

func (s *ContractServer) ApproveMilestoneSubmission(ctx context.Context, req *contractv1.ApproveMilestoneSubmissionRequest) (*contractv1.ApproveMilestoneSubmissionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	out, err := s.ApproveMilestoneUC.Execute(ctx, application.ApproveMilestoneSubmissionInput{
		ContractID:  req.GetContractId(),
		MilestoneID: req.GetMilestoneId(),
		ActorID:     callerID,
		ActorRole:   role,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &contractv1.ApproveMilestoneSubmissionResponse{Contract: toProtoContract(out.Contract)}, nil
}

func (s *ContractServer) UpdateMilestoneStatus(ctx context.Context, req *contractv1.UpdateMilestoneStatusRequest) (*contractv1.UpdateMilestoneStatusResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	out, err := s.UpdateMilestoneStatusUC.Execute(ctx, application.UpdateMilestoneStatusInput{
		ContractID:     req.GetContractId(),
		MilestoneID:    req.GetMilestoneId(),
		ActorID:        callerID,
		ActorRole:      role,
		Status:         fromProtoMilestoneStatus(req.GetStatus()),
		SubmissionNote: req.GetSubmissionNote(),
		SubmissionURLs: req.GetSubmissionUrls(),
		ReviewNote:     req.GetReviewNote(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &contractv1.UpdateMilestoneStatusResponse{Contract: toProtoContract(out.Contract)}, nil
}

func (s *ContractServer) LogHourlyWork(ctx context.Context, req *contractv1.LogHourlyWorkRequest) (*contractv1.LogHourlyWorkResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireFreelancerRole(role); err != nil {
		return nil, err
	}
	out, err := s.LogHourlyWorkUC.Execute(ctx, application.LogHourlyWorkInput{
		ContractID:   req.GetContractId(),
		FreelancerID: callerID,
		StartAt:      unixToTime(req.GetStartAtUnixSeconds()),
		EndAt:        unixToTime(req.GetEndAtUnixSeconds()),
		Note:         req.GetNote(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &contractv1.LogHourlyWorkResponse{HourlyLog: toProtoHourlyLog(out.HourlyLog)}, nil
}

func (s *ContractServer) ListHourlyLogs(ctx context.Context, req *contractv1.ListHourlyLogsRequest) (*contractv1.ListHourlyLogsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, _, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	out, err := s.ListHourlyLogsUC.Execute(ctx, application.ListHourlyLogsInput{
		ContractID: req.GetContractId(),
		ActorID:    callerID,
		PageSize:   req.GetPageSize(),
		PageToken:  req.GetPageToken(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	items := make([]*contractv1.HourlyLog, 0, len(out.HourlyLogs))
	for _, item := range out.HourlyLogs {
		items = append(items, toProtoHourlyLog(item))
	}
	return &contractv1.ListHourlyLogsResponse{HourlyLogs: items, NextPageToken: out.NextPageToken}, nil
}

func (s *ContractServer) ReviewHourlyLog(ctx context.Context, req *contractv1.ReviewHourlyLogRequest) (*contractv1.ReviewHourlyLogResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.ReviewHourlyLogUC.Execute(ctx, application.ReviewHourlyLogInput{
		HourlyLogID: req.GetHourlyLogId(),
		ClientID:    callerID,
		Status:      fromProtoHourlyLogStatus(req.GetStatus()),
		ReviewNote:  req.GetReviewNote(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &contractv1.ReviewHourlyLogResponse{HourlyLog: toProtoHourlyLog(out.HourlyLog)}, nil
}

func (s *ContractServer) ProposeAmendment(ctx context.Context, req *contractv1.ProposeAmendmentRequest) (*contractv1.ProposeAmendmentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	role = strings.ToLower(strings.TrimSpace(role))
	if role != "client" && role != "freelancer" {
		return nil, status.Error(codes.PermissionDenied, "client or freelancer role required")
	}
	var expiresAt *time.Time
	if req.GetExpiresAtUnixSeconds() > 0 {
		t := unixToTime(req.GetExpiresAtUnixSeconds())
		expiresAt = &t
	}
	out, err := s.ProposeAmendmentUC.Execute(ctx, application.ProposeAmendmentInput{
		ContractID: req.GetContractId(),
		ActorID:    callerID,
		Summary:    req.GetSummary(),
		Payload:    fromProtoAmendmentPayload(req.GetPayload()),
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &contractv1.ProposeAmendmentResponse{Amendment: toProtoAmendment(out.Amendment)}, nil
}

func (s *ContractServer) RespondAmendment(ctx context.Context, req *contractv1.RespondAmendmentRequest) (*contractv1.RespondAmendmentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	role = strings.ToLower(strings.TrimSpace(role))
	if role != "client" && role != "freelancer" {
		return nil, status.Error(codes.PermissionDenied, "client or freelancer role required")
	}
	out, err := s.RespondAmendmentUC.Execute(ctx, application.RespondAmendmentInput{
		AmendmentID:  req.GetAmendmentId(),
		ActorID:      callerID,
		Status:       fromProtoAmendmentStatus(req.GetStatus()),
		ResponseNote: req.GetResponseNote(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &contractv1.RespondAmendmentResponse{Amendment: toProtoAmendment(out.Amendment)}, nil
}

func (s *ContractServer) ListAmendments(ctx context.Context, req *contractv1.ListAmendmentsRequest) (*contractv1.ListAmendmentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, _, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	out, err := s.ListAmendmentsUC.Execute(ctx, application.ListAmendmentsInput{
		ContractID: req.GetContractId(),
		ActorID:    callerID,
		PageSize:   req.GetPageSize(),
		PageToken:  req.GetPageToken(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	items := make([]*contractv1.Amendment, 0, len(out.Amendments))
	for _, item := range out.Amendments {
		items = append(items, toProtoAmendment(item))
	}
	return &contractv1.ListAmendmentsResponse{Amendments: items, NextPageToken: out.NextPageToken}, nil
}

func (s *ContractServer) PauseContract(ctx context.Context, req *contractv1.PauseContractRequest) (*contractv1.PauseContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.PauseUC.Execute(ctx, application.PauseContractInput{ContractID: req.GetContractId(), ClientID: callerID, Reason: req.GetReason()})
	if err != nil {
		return nil, toStatus(err)
	}
	return &contractv1.PauseContractResponse{Contract: toProtoContract(out.Contract)}, nil
}

func (s *ContractServer) ResumeContract(ctx context.Context, req *contractv1.ResumeContractRequest) (*contractv1.ResumeContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.ResumeUC.Execute(ctx, application.ResumeContractInput{ContractID: req.GetContractId(), ClientID: callerID, Reason: req.GetReason()})
	if err != nil {
		return nil, toStatus(err)
	}
	return &contractv1.ResumeContractResponse{Contract: toProtoContract(out.Contract)}, nil
}

func (s *ContractServer) EndContract(ctx context.Context, req *contractv1.EndContractRequest) (*contractv1.EndContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	if err := requireClientRole(role); err != nil {
		return nil, err
	}
	out, err := s.EndUC.Execute(ctx, application.EndContractInput{ContractID: req.GetContractId(), ClientID: callerID, Reason: req.GetReason()})
	if err != nil {
		return nil, toStatus(err)
	}
	return &contractv1.EndContractResponse{Contract: toProtoContract(out.Contract)}, nil
}

func (s *ContractServer) GetStatusHistory(ctx context.Context, req *contractv1.GetStatusHistoryRequest) (*contractv1.GetStatusHistoryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, _, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	out, err := s.GetStatusHistoryUC.Execute(ctx, application.GetStatusHistoryInput{
		ContractID: req.GetContractId(),
		ActorID:    callerID,
		PageSize:   req.GetPageSize(),
		PageToken:  req.GetPageToken(),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	entries := make([]*contractv1.StatusHistoryEntry, 0, len(out.Entries))
	for _, e := range out.Entries {
		entries = append(entries, toProtoStatusHistory(e))
	}
	return &contractv1.GetStatusHistoryResponse{Entries: entries, NextPageToken: out.NextPageToken}, nil
}

func fromProtoType(v contractv1.ContractType) string {
	switch v {
	case contractv1.ContractType_CONTRACT_TYPE_FIXED:
		return domain.TypeFixed
	case contractv1.ContractType_CONTRACT_TYPE_HOURLY:
		return domain.TypeHourly
	default:
		return ""
	}
}

func toProtoType(v string) contractv1.ContractType {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case domain.TypeFixed:
		return contractv1.ContractType_CONTRACT_TYPE_FIXED
	case domain.TypeHourly:
		return contractv1.ContractType_CONTRACT_TYPE_HOURLY
	default:
		return contractv1.ContractType_CONTRACT_TYPE_UNSPECIFIED
	}
}

func fromProtoStatus(v contractv1.ContractStatus) string {
	switch v {
	case contractv1.ContractStatus_CONTRACT_STATUS_PENDING_ACCEPTANCE:
		return domain.StatusPendingAcceptance
	case contractv1.ContractStatus_CONTRACT_STATUS_ACTIVE:
		return domain.StatusActive
	case contractv1.ContractStatus_CONTRACT_STATUS_DECLINED:
		return domain.StatusDeclined
	case contractv1.ContractStatus_CONTRACT_STATUS_REVOKED:
		return domain.StatusRevoked
	case contractv1.ContractStatus_CONTRACT_STATUS_PAUSED:
		return domain.StatusPaused
	case contractv1.ContractStatus_CONTRACT_STATUS_ENDED:
		return domain.StatusEnded
	default:
		return ""
	}
}

func toProtoStatus(v string) contractv1.ContractStatus {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case domain.StatusPendingAcceptance:
		return contractv1.ContractStatus_CONTRACT_STATUS_PENDING_ACCEPTANCE
	case domain.StatusActive:
		return contractv1.ContractStatus_CONTRACT_STATUS_ACTIVE
	case domain.StatusDeclined:
		return contractv1.ContractStatus_CONTRACT_STATUS_DECLINED
	case domain.StatusRevoked:
		return contractv1.ContractStatus_CONTRACT_STATUS_REVOKED
	case domain.StatusPaused:
		return contractv1.ContractStatus_CONTRACT_STATUS_PAUSED
	case domain.StatusEnded:
		return contractv1.ContractStatus_CONTRACT_STATUS_ENDED
	default:
		return contractv1.ContractStatus_CONTRACT_STATUS_UNSPECIFIED
	}
}

func fromProtoMilestoneStatus(v contractv1.MilestoneStatus) string {
	switch v {
	case contractv1.MilestoneStatus_MILESTONE_STATUS_PENDING:
		return domain.MilestoneStatusPending
	case contractv1.MilestoneStatus_MILESTONE_STATUS_SUBMITTED:
		return domain.MilestoneStatusSubmitted
	case contractv1.MilestoneStatus_MILESTONE_STATUS_CHANGES_REQUESTED:
		return domain.MilestoneStatusChangesRequested
	case contractv1.MilestoneStatus_MILESTONE_STATUS_APPROVED:
		return domain.MilestoneStatusApproved
	case contractv1.MilestoneStatus_MILESTONE_STATUS_FUNDED:
		return domain.MilestoneStatusFunded
	case contractv1.MilestoneStatus_MILESTONE_STATUS_APPROVED_PENDING_SETTLEMENT:
		return domain.MilestoneStatusApprovedPendingSettlement
	default:
		return ""
	}
}

func toProtoMilestoneStatus(v string) contractv1.MilestoneStatus {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case domain.MilestoneStatusPending:
		return contractv1.MilestoneStatus_MILESTONE_STATUS_PENDING
	case domain.MilestoneStatusSubmitted:
		return contractv1.MilestoneStatus_MILESTONE_STATUS_SUBMITTED
	case domain.MilestoneStatusChangesRequested:
		return contractv1.MilestoneStatus_MILESTONE_STATUS_CHANGES_REQUESTED
	case domain.MilestoneStatusApproved:
		return contractv1.MilestoneStatus_MILESTONE_STATUS_APPROVED
	case domain.MilestoneStatusFunded:
		return contractv1.MilestoneStatus_MILESTONE_STATUS_FUNDED
	case domain.MilestoneStatusApprovedPendingSettlement:
		return contractv1.MilestoneStatus_MILESTONE_STATUS_APPROVED_PENDING_SETTLEMENT
	default:
		return contractv1.MilestoneStatus_MILESTONE_STATUS_UNSPECIFIED
	}
}

func fromProtoHourlyLogStatus(v contractv1.HourlyLogStatus) string {
	switch v {
	case contractv1.HourlyLogStatus_HOURLY_LOG_STATUS_PENDING:
		return domain.HourlyLogStatusPending
	case contractv1.HourlyLogStatus_HOURLY_LOG_STATUS_APPROVED:
		return domain.HourlyLogStatusApproved
	case contractv1.HourlyLogStatus_HOURLY_LOG_STATUS_REJECTED:
		return domain.HourlyLogStatusRejected
	default:
		return ""
	}
}

func toProtoHourlyLogStatus(v string) contractv1.HourlyLogStatus {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case domain.HourlyLogStatusPending:
		return contractv1.HourlyLogStatus_HOURLY_LOG_STATUS_PENDING
	case domain.HourlyLogStatusApproved:
		return contractv1.HourlyLogStatus_HOURLY_LOG_STATUS_APPROVED
	case domain.HourlyLogStatusRejected:
		return contractv1.HourlyLogStatus_HOURLY_LOG_STATUS_REJECTED
	default:
		return contractv1.HourlyLogStatus_HOURLY_LOG_STATUS_UNSPECIFIED
	}
}

func fromProtoAmendmentStatus(v contractv1.AmendmentStatus) string {
	switch v {
	case contractv1.AmendmentStatus_AMENDMENT_STATUS_PENDING:
		return domain.AmendmentStatusPending
	case contractv1.AmendmentStatus_AMENDMENT_STATUS_ACCEPTED:
		return domain.AmendmentStatusAccepted
	case contractv1.AmendmentStatus_AMENDMENT_STATUS_REJECTED:
		return domain.AmendmentStatusRejected
	case contractv1.AmendmentStatus_AMENDMENT_STATUS_EXPIRED:
		return domain.AmendmentStatusExpired
	default:
		return ""
	}
}

func toProtoAmendmentStatus(v string) contractv1.AmendmentStatus {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case domain.AmendmentStatusPending:
		return contractv1.AmendmentStatus_AMENDMENT_STATUS_PENDING
	case domain.AmendmentStatusAccepted:
		return contractv1.AmendmentStatus_AMENDMENT_STATUS_ACCEPTED
	case domain.AmendmentStatusRejected:
		return contractv1.AmendmentStatus_AMENDMENT_STATUS_REJECTED
	case domain.AmendmentStatusExpired:
		return contractv1.AmendmentStatus_AMENDMENT_STATUS_EXPIRED
	default:
		return contractv1.AmendmentStatus_AMENDMENT_STATUS_UNSPECIFIED
	}
}

func fromProtoMilestones(in []*contractv1.Milestone) []domain.Milestone {
	if len(in) == 0 {
		return nil
	}
	out := make([]domain.Milestone, 0, len(in))
	for _, m := range in {
		if m == nil {
			continue
		}
		item := domain.Milestone{
			ID:             m.GetId(),
			Title:          m.GetTitle(),
			Description:    m.GetDescription(),
			Amount:         m.GetAmount(),
			Status:         fromProtoMilestoneStatus(m.GetStatus()),
			SubmissionNote: strings.TrimSpace(m.GetSubmissionNote()),
			SubmissionURLs: m.GetSubmissionUrls(),
			ReviewNote:     strings.TrimSpace(m.GetReviewNote()),
			RevisionCount:  m.GetRevisionCount(),
		}
		if m.GetDueAtUnixSeconds() > 0 {
			due := unixToTime(m.GetDueAtUnixSeconds())
			item.DueAt = &due
		}
		if m.GetSubmittedAtUnixSeconds() > 0 {
			submittedAt := unixToTime(m.GetSubmittedAtUnixSeconds())
			item.SubmittedAt = &submittedAt
		}
		if m.GetReviewedAtUnixSeconds() > 0 {
			reviewedAt := unixToTime(m.GetReviewedAtUnixSeconds())
			item.ReviewedAt = &reviewedAt
		}
		out = append(out, item)
	}
	return out
}

func toProtoMilestones(in []domain.Milestone) []*contractv1.Milestone {
	out := make([]*contractv1.Milestone, 0, len(in))
	for _, m := range in {
		item := &contractv1.Milestone{
			Id:             m.ID,
			Title:          m.Title,
			Description:    m.Description,
			Amount:         m.Amount,
			Status:         toProtoMilestoneStatus(m.Status),
			SubmissionNote: m.SubmissionNote,
			SubmissionUrls: m.SubmissionURLs,
			ReviewNote:     m.ReviewNote,
			RevisionCount:  m.RevisionCount,
		}
		if m.DueAt != nil {
			item.DueAtUnixSeconds = m.DueAt.Unix()
		}
		if m.SubmittedAt != nil {
			item.SubmittedAtUnixSeconds = m.SubmittedAt.Unix()
		}
		if m.ReviewedAt != nil {
			item.ReviewedAtUnixSeconds = m.ReviewedAt.Unix()
		}
		out = append(out, item)
	}
	return out
}

func toProtoContract(in domain.Contract) *contractv1.Contract {
	out := &contractv1.Contract{
		Id:                   in.ID,
		ClientId:             in.ClientID.String(),
		FreelancerId:         in.FreelancerID.String(),
		JobId:                in.JobID,
		ProposalId:           in.ProposalID,
		ContractType:         toProtoType(in.ContractType),
		Status:               toProtoStatus(in.Status),
		Title:                in.Title,
		Description:          in.Description,
		HourlyRate:           in.HourlyRate,
		FixedTotal:           in.FixedTotal,
		WeeklyHourLimit:      in.WeeklyHourLimit,
		Milestones:           toProtoMilestones(in.Milestones),
		CreatedAtUnixSeconds: in.CreatedAt.Unix(),
		UpdatedAtUnixSeconds: in.UpdatedAt.Unix(),
	}
	if in.ActivatedAt != nil {
		out.ActivatedAtUnixSeconds = in.ActivatedAt.Unix()
	}
	if in.DeclinedAt != nil {
		out.DeclinedAtUnixSeconds = in.DeclinedAt.Unix()
	}
	if in.RevokedAt != nil {
		out.RevokedAtUnixSeconds = in.RevokedAt.Unix()
	}
	if in.PausedAt != nil {
		out.PausedAtUnixSeconds = in.PausedAt.Unix()
	}
	if in.EndedAt != nil {
		out.EndedAtUnixSeconds = in.EndedAt.Unix()
	}
	return out
}

func toProtoHourlyLog(in domain.HourlyLog) *contractv1.HourlyLog {
	out := &contractv1.HourlyLog{
		Id:                   in.ID,
		ContractId:           in.ContractID,
		FreelancerId:         in.FreelancerID.String(),
		WorkDateUnixSeconds:  in.WorkDate.Unix(),
		StartAtUnixSeconds:   in.StartAt.Unix(),
		EndAtUnixSeconds:     in.EndAt.Unix(),
		DurationMinutes:      in.DurationMin,
		Note:                 in.Note,
		Status:               toProtoHourlyLogStatus(in.Status),
		ReviewNote:           in.ReviewNote,
		CreatedAtUnixSeconds: in.CreatedAt.Unix(),
	}
	if in.ClientReviewAt != nil {
		out.ReviewedAtUnixSeconds = in.ClientReviewAt.Unix()
	}
	return out
}

func toProtoAmendment(in domain.Amendment) *contractv1.Amendment {
	out := &contractv1.Amendment{
		Id:                   in.ID,
		ContractId:           in.ContractID,
		ProposedBy:           in.ProposedBy.String(),
		Summary:              in.Summary,
		Payload:              toProtoAmendmentPayload(in.Payload),
		Status:               toProtoAmendmentStatus(in.Status),
		CreatedAtUnixSeconds: in.CreatedAt.Unix(),
		ResponseNote:         in.ResponseNote,
	}
	if in.ExpiresAt != nil {
		out.ExpiresAtUnixSeconds = in.ExpiresAt.Unix()
	}
	if in.RespondedAt != nil {
		out.RespondedAtUnixSeconds = in.RespondedAt.Unix()
	}
	if in.RespondedBy != nil {
		out.RespondedBy = in.RespondedBy.String()
	}
	return out
}

func fromProtoAmendmentPayload(in *contractv1.AmendmentPayload) domain.AmendmentPayload {
	if in == nil {
		return domain.AmendmentPayload{}
	}
	out := domain.AmendmentPayload{}
	if in.GetCompensationChange() != nil {
		out.CompensationChange = &domain.CompensationChange{
			NewHourlyRate: in.GetCompensationChange().GetNewHourlyRate(),
			NewFixedTotal: in.GetCompensationChange().GetNewFixedTotal(),
		}
	}
	if in.GetMilestonesChange() != nil {
		out.MilestonesChange = &domain.MilestonesChange{
			Milestones: fromProtoMilestones(in.GetMilestonesChange().GetMilestones()),
		}
	}
	if in.GetWeeklyLimitChange() != nil {
		out.WeeklyLimitChange = &domain.WeeklyLimitChange{
			NewWeeklyHourLimit: in.GetWeeklyLimitChange().GetNewWeeklyHourLimit(),
		}
	}
	if in.GetScopeChange() != nil {
		out.ScopeChange = &domain.ScopeChange{
			NewTitle:       in.GetScopeChange().GetNewTitle(),
			NewDescription: in.GetScopeChange().GetNewDescription(),
		}
	}
	return out
}

func toProtoAmendmentPayload(in domain.AmendmentPayload) *contractv1.AmendmentPayload {
	out := &contractv1.AmendmentPayload{}
	if in.CompensationChange != nil {
		out.CompensationChange = &contractv1.CompensationChange{
			NewHourlyRate: in.CompensationChange.NewHourlyRate,
			NewFixedTotal: in.CompensationChange.NewFixedTotal,
		}
	}
	if in.MilestonesChange != nil {
		out.MilestonesChange = &contractv1.MilestonesChange{
			Milestones: toProtoMilestones(in.MilestonesChange.Milestones),
		}
	}
	if in.WeeklyLimitChange != nil {
		out.WeeklyLimitChange = &contractv1.WeeklyLimitChange{
			NewWeeklyHourLimit: in.WeeklyLimitChange.NewWeeklyHourLimit,
		}
	}
	if in.ScopeChange != nil {
		out.ScopeChange = &contractv1.ScopeChange{
			NewTitle:       in.ScopeChange.NewTitle,
			NewDescription: in.ScopeChange.NewDescription,
		}
	}
	return out
}

func toProtoStatusHistory(in domain.StatusHistoryEntry) *contractv1.StatusHistoryEntry {
	return &contractv1.StatusHistoryEntry{
		Id:                   in.ID,
		ContractId:           in.ContractID,
		Status:               toProtoStatus(in.Status),
		Reason:               in.Reason,
		ActorId:              in.ActorID.String(),
		CreatedAtUnixSeconds: in.CreatedAt.Unix(),
	}
}

func toStatus(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := status.FromError(err); ok {
		return err
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "not found"):
		return status.Error(codes.NotFound, err.Error())
	case strings.Contains(msg, "required"), strings.Contains(msg, "invalid"), strings.Contains(msg, "too long"), strings.Contains(msg, "must"):
		return status.Error(codes.InvalidArgument, err.Error())
	case strings.Contains(msg, "role") || strings.Contains(msg, "owner") || strings.Contains(msg, "eligible") || strings.Contains(msg, "cannot"):
		return status.Error(codes.PermissionDenied, err.Error())
	case strings.Contains(msg, "already exists") ||
		strings.Contains(msg, "already has") ||
		strings.Contains(msg, "can only") ||
		strings.Contains(msg, "expired") ||
		strings.Contains(msg, "open dispute exists") ||
		strings.Contains(msg, "acceptable state") ||
		strings.Contains(msg, "revoke-able state") ||
		strings.Contains(msg, "does not belong to"):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func unixToTime(sec int64) time.Time {
	return time.Unix(sec, 0).UTC()
}
