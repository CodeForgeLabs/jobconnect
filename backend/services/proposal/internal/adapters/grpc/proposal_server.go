package grpcadapter

import (
	"context"
	"strings"

	proposalv1 "jobconnect/proposal/gen/proposal/v1"
	"jobconnect/proposal/internal/application"
	"jobconnect/proposal/internal/domain"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProposalServer struct {
	proposalv1.UnimplementedProposalServiceServer

	SubmitUC    *application.SubmitProposal
	ModifyUC    *application.ModifyProposal
	WithdrawUC  *application.WithdrawProposal
	GetUC       *application.GetProposal
	ListByJobUC *application.ListProposalsByJob
	ListMineUC  *application.ListMyProposals
	SetStatusUC *application.SetProposalStatus

	TokenParser TokenParser
}

func NewProposalServer(
	submit *application.SubmitProposal,
	modify *application.ModifyProposal,
	withdraw *application.WithdrawProposal,
	get *application.GetProposal,
	listByJob *application.ListProposalsByJob,
	listMine *application.ListMyProposals,
	setStatus *application.SetProposalStatus,
	tokenParser TokenParser,
) *ProposalServer {
	return &ProposalServer{
		SubmitUC:    submit,
		ModifyUC:    modify,
		WithdrawUC:  withdraw,
		GetUC:       get,
		ListByJobUC: listByJob,
		ListMineUC:  listMine,
		SetStatusUC: setStatus,
		TokenParser: tokenParser,
	}
}

func (s *ProposalServer) SubmitProposal(ctx context.Context, req *proposalv1.SubmitProposalRequest) (*proposalv1.SubmitProposalResponse, error) {
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

	out, err := s.SubmitUC.Execute(ctx, application.SubmitProposalInput{
		FreelancerID:  callerID,
		JobID:         req.JobId,
		CoverLetter:   req.CoverLetter,
		BidType:       req.BidType,
		BidAmount:     req.BidAmount,
		EstimatedDays: req.EstimatedDays,
		Attachments:   fromProtoAttachments(req.Attachments),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &proposalv1.SubmitProposalResponse{Proposal: toProtoProposal(out.Proposal)}, nil
}

func (s *ProposalServer) ModifyProposal(ctx context.Context, req *proposalv1.ModifyProposalRequest) (*proposalv1.ModifyProposalResponse, error) {
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

	out, err := s.ModifyUC.Execute(ctx, application.ModifyProposalInput{
		ProposalID:    req.ProposalId,
		FreelancerID:  callerID,
		CoverLetter:   req.CoverLetter,
		BidAmount:     req.BidAmount,
		EstimatedDays: req.EstimatedDays,
		Attachments:   fromProtoAttachments(req.Attachments),
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &proposalv1.ModifyProposalResponse{Proposal: toProtoProposal(out.Proposal)}, nil
}

func (s *ProposalServer) WithdrawProposal(ctx context.Context, req *proposalv1.WithdrawProposalRequest) (*proposalv1.WithdrawProposalResponse, error) {
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

	out, err := s.WithdrawUC.Execute(ctx, application.WithdrawProposalInput{
		ProposalID:   req.ProposalId,
		FreelancerID: callerID,
		Reason:       req.Reason,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &proposalv1.WithdrawProposalResponse{Withdrawn: out.Withdrawn}, nil
}

func (s *ProposalServer) GetProposal(ctx context.Context, req *proposalv1.GetProposalRequest) (*proposalv1.GetProposalResponse, error) {
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

	out, err := s.GetUC.Execute(ctx, application.GetProposalInput{
		ProposalID: req.ProposalId,
		ActorID:    callerID,
		ActorRole:  role,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &proposalv1.GetProposalResponse{Proposal: toProtoProposal(out.Proposal)}, nil
}

func (s *ProposalServer) ListProposalsByJob(ctx context.Context, req *proposalv1.ListProposalsByJobRequest) (*proposalv1.ListProposalsByJobResponse, error) {
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

	statuses := make([]string, 0, len(req.StatusFilter))
	for _, s := range req.StatusFilter {
		if mapped, ok := fromProtoStatus(s); ok {
			statuses = append(statuses, mapped)
		}
	}

	var freelancerID *uuid.UUID
	if req.FreelancerIdFilter != nil {
		parsed, err := uuid.Parse(req.GetFreelancerIdFilter())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid freelancer_id_filter")
		}
		freelancerID = &parsed
	}

	out, err := s.ListByJobUC.Execute(ctx, application.ListProposalsByJobInput{
		ClientID:     callerID,
		JobID:        req.JobId,
		StatusFilter: statuses,
		FreelancerID: freelancerID,
		SortBy:       fromProtoSort(req.SortBy),
		PageSize:     req.PageSize,
		PageToken:    req.PageToken,
	})
	if err != nil {
		return nil, toStatus(err)
	}

	items := make([]*proposalv1.Proposal, 0, len(out.Proposals))
	for _, p := range out.Proposals {
		items = append(items, toProtoProposal(p))
	}
	return &proposalv1.ListProposalsByJobResponse{Proposals: items, NextPageToken: out.NextPageToken}, nil
}

func (s *ProposalServer) ListMyProposals(ctx context.Context, req *proposalv1.ListMyProposalsRequest) (*proposalv1.ListMyProposalsResponse, error) {
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

	statuses := make([]string, 0, len(req.StatusFilter))
	for _, s := range req.StatusFilter {
		if mapped, ok := fromProtoStatus(s); ok {
			statuses = append(statuses, mapped)
		}
	}

	var jobID *int64
	if req.JobIdFilter != nil {
		v := req.GetJobIdFilter()
		jobID = &v
	}

	out, err := s.ListMineUC.Execute(ctx, application.ListMyProposalsInput{
		FreelancerID: callerID,
		StatusFilter: statuses,
		JobIDFilter:  jobID,
		SortBy:       fromProtoSort(req.SortBy),
		PageSize:     req.PageSize,
		PageToken:    req.PageToken,
	})
	if err != nil {
		return nil, toStatus(err)
	}

	items := make([]*proposalv1.Proposal, 0, len(out.Proposals))
	for _, p := range out.Proposals {
		items = append(items, toProtoProposal(p))
	}
	return &proposalv1.ListMyProposalsResponse{Proposals: items, NextPageToken: out.NextPageToken}, nil
}

func (s *ProposalServer) SetProposalStatus(ctx context.Context, req *proposalv1.SetProposalStatusRequest) (*proposalv1.SetProposalStatusResponse, error) {
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

	next, ok := fromProtoStatus(req.Status)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "invalid status")
	}

	out, err := s.SetStatusUC.Execute(ctx, application.SetProposalStatusInput{
		ProposalID: req.ProposalId,
		ClientID:   callerID,
		Status:     next,
		Reason:     req.Reason,
	})
	if err != nil {
		return nil, toStatus(err)
	}
	return &proposalv1.SetProposalStatusResponse{Proposal: toProtoProposal(out.Proposal)}, nil
}

func fromProtoAttachments(in []*proposalv1.ProposalAttachment) []domain.Attachment {
	if len(in) == 0 {
		return nil
	}
	out := make([]domain.Attachment, 0, len(in))
	for _, a := range in {
		if a == nil {
			continue
		}
		out = append(out, domain.Attachment{FileName: a.FileName, ContentType: a.ContentType, URL: a.Url, SizeBytes: a.SizeBytes})
	}
	return out
}

func toProtoAttachments(in []domain.Attachment) []*proposalv1.ProposalAttachment {
	out := make([]*proposalv1.ProposalAttachment, 0, len(in))
	for _, a := range in {
		out = append(out, &proposalv1.ProposalAttachment{Id: a.ID, FileName: a.FileName, ContentType: a.ContentType, Url: a.URL, SizeBytes: a.SizeBytes})
	}
	return out
}

func toProtoStatus(v string) proposalv1.ProposalStatus {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case domain.StatusSent:
		return proposalv1.ProposalStatus_PROPOSAL_STATUS_SENT
	case domain.StatusShortlisted:
		return proposalv1.ProposalStatus_PROPOSAL_STATUS_SHORTLISTED
	case domain.StatusRejected:
		return proposalv1.ProposalStatus_PROPOSAL_STATUS_REJECTED
	case domain.StatusHired:
		return proposalv1.ProposalStatus_PROPOSAL_STATUS_HIRED
	case domain.StatusWithdrawn:
		return proposalv1.ProposalStatus_PROPOSAL_STATUS_WITHDRAWN
	default:
		return proposalv1.ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED
	}
}

func fromProtoStatus(v proposalv1.ProposalStatus) (string, bool) {
	switch v {
	case proposalv1.ProposalStatus_PROPOSAL_STATUS_SENT:
		return domain.StatusSent, true
	case proposalv1.ProposalStatus_PROPOSAL_STATUS_SHORTLISTED:
		return domain.StatusShortlisted, true
	case proposalv1.ProposalStatus_PROPOSAL_STATUS_REJECTED:
		return domain.StatusRejected, true
	case proposalv1.ProposalStatus_PROPOSAL_STATUS_HIRED:
		return domain.StatusHired, true
	case proposalv1.ProposalStatus_PROPOSAL_STATUS_WITHDRAWN:
		return domain.StatusWithdrawn, true
	default:
		return "", false
	}
}

func fromProtoSort(v proposalv1.SortBy) string {
	switch v {
	case proposalv1.SortBy_SORT_BY_OLDEST:
		return domain.SortOldest
	case proposalv1.SortBy_SORT_BY_BID_HIGH:
		return domain.SortBidHigh
	case proposalv1.SortBy_SORT_BY_BID_LOW:
		return domain.SortBidLow
	case proposalv1.SortBy_SORT_BY_NEWEST, proposalv1.SortBy_SORT_BY_UNSPECIFIED:
		fallthrough
	default:
		return domain.SortNewest
	}
}

func toProtoProposal(in domain.Proposal) *proposalv1.Proposal {
	out := &proposalv1.Proposal{
		Id:                   in.ID,
		JobId:                in.JobID,
		ClientId:             in.ClientID.String(),
		FreelancerId:         in.FreelancerID.String(),
		CoverLetter:          in.CoverLetter,
		BidType:              in.BidType,
		BidAmount:            in.BidAmount,
		EstimatedDays:        in.EstimatedDays,
		Attachments:          toProtoAttachments(in.Attachments),
		Status:               toProtoStatus(in.Status),
		StatusReason:         in.StatusReason,
		CreatedAtUnixSeconds: in.CreatedAt.Unix(),
		UpdatedAtUnixSeconds: in.UpdatedAt.Unix(),
	}
	if in.ShortlistedAt != nil {
		out.ShortlistedAtUnixSeconds = in.ShortlistedAt.Unix()
	}
	if in.RejectedAt != nil {
		out.RejectedAtUnixSeconds = in.RejectedAt.Unix()
	}
	if in.HiredAt != nil {
		out.HiredAtUnixSeconds = in.HiredAt.Unix()
	}
	if in.WithdrawnAt != nil {
		out.WithdrawnAtUnixSeconds = in.WithdrawnAt.Unix()
	}
	return out
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
	case strings.Contains(msg, "role") || strings.Contains(msg, "owner"):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
