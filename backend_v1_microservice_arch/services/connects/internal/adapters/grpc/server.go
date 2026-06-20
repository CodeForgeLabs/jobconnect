package grpc

import (
	"context"
	"errors"

	"jobconnect/services/connects/internal/application"
	"jobconnect/services/connects/internal/domain"
	pb "jobconnect/api/proto/connects/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ConnectsServer struct {
	pb.UnimplementedConnectsServiceServer
	app *application.UseCases
}

func NewConnectsServer(app *application.UseCases) *ConnectsServer {
	return &ConnectsServer{app: app}
}

func (s *ConnectsServer) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	bal, err := s.app.GetBalance(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get balance: %v", err)
	}

	return &pb.GetBalanceResponse{
		Balance: bal.Balance,
	}, nil
}

func (s *ConnectsServer) DeductConnects(ctx context.Context, req *pb.DeductConnectsRequest) (*pb.DeductConnectsResponse, error) {
	if req.UserId == "" || req.Amount <= 0 || req.ReferenceId == "" || req.ReferenceType == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request fields")
	}

	bal, err := s.app.DeductConnects(ctx, req.UserId, req.Amount, req.ReferenceId, req.ReferenceType)
	if err != nil {
		if errors.Is(err, domain.ErrInsufficientBalance) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		if errors.Is(err, domain.ErrInvalidConnectsMinimum) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		// Likely a unique constraint violation from idempotency (already deducted)
		return nil, status.Errorf(codes.Internal, "failed to deduct connects: %v", err)
	}

	return &pb.DeductConnectsResponse{
		Success:    true,
		NewBalance: bal.Balance,
	}, nil
}

func (s *ConnectsServer) RefundConnects(ctx context.Context, req *pb.RefundConnectsRequest) (*pb.RefundConnectsResponse, error) {
	if req.UserId == "" || req.Amount <= 0 || req.ReferenceId == "" || req.ReferenceType == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request fields")
	}

	bal, err := s.app.RefundConnects(ctx, req.UserId, req.Amount, req.ReferenceId, req.ReferenceType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to refund connects: %v", err)
	}

	return &pb.RefundConnectsResponse{
		Success:    true,
		NewBalance: bal.Balance,
	}, nil
}

func (s *ConnectsServer) GrantInitialConnects(ctx context.Context, req *pb.GrantInitialConnectsRequest) (*pb.GrantInitialConnectsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	bal, err := s.app.GrantInitialConnects(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to grant initial connects: %v", err)
	}

	return &pb.GrantInitialConnectsResponse{
		Success:    true,
		NewBalance: bal.Balance,
	}, nil
}

func (s *ConnectsServer) PurchaseConnects(ctx context.Context, req *pb.PurchaseConnectsRequest) (*pb.PurchaseConnectsResponse, error) {
	if req.UserId == "" || req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid request fields")
	}

	// NOTE: In the future, this should call the wallet service to check/deduct fiat before granting connects.
	// For phase 1, we just grant the connects directly assuming external validation.
	bal, err := s.app.PurchaseConnects(ctx, req.UserId, req.Amount)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to purchase connects: %v", err)
	}

	return &pb.PurchaseConnectsResponse{
		Success:    true,
		NewBalance: bal.Balance,
	}, nil
}
