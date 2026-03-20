package server

import (
	"context"
	"strings"

	servicecontract "github.com/luckysxx/user-platform/internal/service/contract"
	grpcerrs "github.com/luckysxx/user-platform/internal/transport/grpc/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/luckysxx/common/proto/user"
	"github.com/luckysxx/user-platform/internal/service"

	"go.uber.org/zap"
)

type UserServer struct {
	pb.UnimplementedUserServiceServer
	svc    service.UserService
	logger *zap.Logger
}

func NewUserServer(svc service.UserService, logger *zap.Logger) *UserServer {
	return &UserServer{svc: svc, logger: logger}
}

func (s *UserServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if strings.TrimSpace(req.GetEmail()) == "" || strings.TrimSpace(req.GetUsername()) == "" || strings.TrimSpace(req.GetPassword()) == "" {
		return nil, status.Error(codes.InvalidArgument, "email/username/password are required")
	}

	resp, err := s.svc.Register(ctx, &servicecontract.RegisterCommand{
		Email:    req.GetEmail(),
		Username: req.GetUsername(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, grpcerrs.ToGRPCError(err)
	}

	return &pb.RegisterResponse{
		UserId:   resp.UserID,
		Username: resp.Username,
	}, nil
}
