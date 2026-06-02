package grpcadapter

import (
	"context"
	"strconv"
	"strings"

	disputev1 "jobconnect/dispute/gen/dispute/v1"
	"jobconnect/dispute/internal/application"
	"jobconnect/dispute/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	disputev1.UnimplementedDisputeServiceServer
	App         *application.Service
	TokenParser TokenParser
}

func NewServer(app *application.Service, tokenParser TokenParser) *Server {
	return &Server{App: app, TokenParser: tokenParser}
}

func (s *Server) OpenDispute(ctx context.Context, req *disputev1.OpenDisputeRequest) (*disputev1.OpenDisputeResponse, error) {
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
	item, err := s.App.OpenDispute(ctx, req.GetReferenceType(), req.GetReferenceId(), callerID, req.GetReason())
	if err != nil {
		return nil, toStatus(err)
	}
	return &disputev1.OpenDisputeResponse{Dispute: toProto(item)}, nil
}

func (s *Server) GetDispute(ctx context.Context, req *disputev1.GetDisputeRequest) (*disputev1.GetDisputeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	_, _, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	item, err := s.App.GetDispute(ctx, req.GetDisputeId())
	if err != nil {
		return nil, toStatus(err)
	}
	return &disputev1.GetDisputeResponse{Dispute: toProto(item)}, nil
}

func (s *Server) ListDisputes(ctx context.Context, req *disputev1.ListDisputesRequest) (*disputev1.ListDisputesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	_, _, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	items, next, err := s.App.ListDisputes(ctx, req.GetReferenceType(), req.GetReferenceId(), fromProtoStatus(req.GetStatus()), req.GetPageSize(), req.GetPageToken())
	if err != nil {
		return nil, toStatus(err)
	}
	out := make([]*disputev1.Dispute, 0, len(items))
	for _, item := range items {
		out = append(out, toProto(item))
	}
	return &disputev1.ListDisputesResponse{Disputes: out, NextPageToken: next}, nil
}

func (s *Server) ResolveDispute(ctx context.Context, req *disputev1.ResolveDisputeRequest) (*disputev1.ResolveDisputeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	callerID, role, err := callerFromContext(ctx, s.TokenParser)
	if err != nil {
		return nil, err
	}
	item, err := s.App.ResolveDispute(ctx, req.GetDisputeId(), fromProtoDecision(req.GetDecision()), req.GetNote(), callerID, role)
	if err != nil {
		return nil, toStatus(err)
	}
	return &disputev1.ResolveDisputeResponse{Dispute: toProto(item)}, nil
}

func toProto(in domain.Dispute) *disputev1.Dispute {
	out := &disputev1.Dispute{
		Id:                   in.ID,
		ReferenceType:        in.ReferenceType,
		ReferenceId:          in.ReferenceID,
		OpenedBy:             in.OpenedBy.String(),
		Reason:               in.Reason,
		Status:               toProtoStatus(in.Status),
		Decision:             toProtoDecision(in.Decision),
		ResolutionNote:       in.ResolutionNote,
		CreatedAtUnixSeconds: in.CreatedAt.Unix(),
	}
	if in.ResolvedBy != nil {
		out.ResolvedBy = in.ResolvedBy.String()
	}
	if in.ResolvedAt != nil {
		out.ResolvedAtUnixSeconds = in.ResolvedAt.Unix()
	}
	return out
}

func fromProtoStatus(v disputev1.DisputeStatus) string {
	switch v {
	case disputev1.DisputeStatus_DISPUTE_STATUS_OPEN:
		return domain.StatusOpen
	case disputev1.DisputeStatus_DISPUTE_STATUS_RESOLVED:
		return domain.StatusResolved
	default:
		return ""
	}
}

func toProtoStatus(v string) disputev1.DisputeStatus {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case domain.StatusOpen:
		return disputev1.DisputeStatus_DISPUTE_STATUS_OPEN
	case domain.StatusResolved:
		return disputev1.DisputeStatus_DISPUTE_STATUS_RESOLVED
	default:
		return disputev1.DisputeStatus_DISPUTE_STATUS_UNSPECIFIED
	}
}

func fromProtoDecision(v disputev1.DisputeDecision) string {
	switch v {
	case disputev1.DisputeDecision_DISPUTE_DECISION_RELEASE:
		return domain.DecisionRelease
	case disputev1.DisputeDecision_DISPUTE_DECISION_REFUND:
		return domain.DecisionRefund
	default:
		return ""
	}
}

func toProtoDecision(v string) disputev1.DisputeDecision {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case domain.DecisionRelease:
		return disputev1.DisputeDecision_DISPUTE_DECISION_RELEASE
	case domain.DecisionRefund:
		return disputev1.DisputeDecision_DISPUTE_DECISION_REFUND
	default:
		return disputev1.DisputeDecision_DISPUTE_DECISION_UNSPECIFIED
	}
}

func toStatus(err error) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	switch {
	case strings.Contains(msg, "required"), strings.Contains(msg, "invalid"):
		return status.Error(codes.InvalidArgument, err.Error())
	case strings.Contains(msg, "not found"):
		return status.Error(codes.NotFound, err.Error())
	case strings.Contains(msg, "role required"):
		return status.Error(codes.PermissionDenied, err.Error())
	case strings.Contains(msg, "not open"):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

func encodePageToken(offset int) string {
	return strconv.Itoa(offset)
}
