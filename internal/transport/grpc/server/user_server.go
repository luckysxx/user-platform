package server

import (
	"context"
	"errors"
	"strings"

	servicecontract "github.com/luckysxx/user-platform/internal/service/contract"
	grpcerrs "github.com/luckysxx/user-platform/internal/transport/grpc/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/luckysxx/user-platform/internal/service"
	pb "github.com/luckysxx/user-platform/proto/user"

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
	if strings.TrimSpace(req.GetUsername()) == "" || strings.TrimSpace(req.GetPassword()) == "" || strings.TrimSpace(req.GetEmail()) == "" {
		return nil, status.Error(codes.InvalidArgument, "username/password/email are required")
	}

	resp, err := s.svc.Register(ctx, &servicecontract.RegisterCommand{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		Email:    req.GetEmail(),
	})
	if err != nil {
		s.logger.Error("grpc register failed", zap.Error(err))
		return nil, grpcerrs.ToGRPCError(err)
	}

	return &pb.RegisterResponse{
		UserId:   resp.UserID,
		Username: resp.Username,
		Email:    resp.Email,
	}, nil
}

func (s *UserServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if strings.TrimSpace(req.GetUsername()) == "" || strings.TrimSpace(req.GetPassword()) == "" {
		return nil, status.Error(codes.InvalidArgument, "username/password are required")
	}

	resp, err := s.svc.Login(ctx, &servicecontract.LoginCommand{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
	})
	if err != nil {
		s.logger.Error("grpc login failed", zap.Error(err))

		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			return nil, status.Error(codes.Unauthenticated, "用户名或密码错误")
		case errors.Is(err, service.ErrTokenGeneration):
			return nil, status.Error(codes.Internal, "internal error")
		default:
			return nil, grpcerrs.ToGRPCError(err)
		}
	}

	return &pb.LoginResponse{
		Token:    resp.Token,
		UserId:   resp.UserID,
		Username: resp.Username,
		Email:    resp.Email,
	}, nil
}
