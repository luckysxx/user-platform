package server

import (
	"context"
	"errors"
	"strings"

	"github.com/luckysxx/user-platform/internal/service"
	servicecontract "github.com/luckysxx/user-platform/internal/service/contract"
	grpcerrs "github.com/luckysxx/user-platform/internal/transport/grpc/errs"
	pb "github.com/luckysxx/user-platform/proto/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	avc    service.AuthService
	logger *zap.Logger
}

func NewAuthServer(avc service.AuthService, logger *zap.Logger) *AuthServer {
	return &AuthServer{avc: avc, logger: logger}
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if strings.TrimSpace(req.GetUsername()) == "" || strings.TrimSpace(req.GetPassword()) == "" {
		return nil, status.Error(codes.InvalidArgument, "username/password are required")
	}

	resp, err := s.avc.Login(ctx, &servicecontract.LoginCommand{
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
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		UserId:       resp.UserID,
		Username:     resp.Username,
	}, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	if strings.TrimSpace(req.GetToken()) == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	resp, err := s.avc.RefreshToken(ctx, &servicecontract.RefreshTokenCommand{Token: req.GetToken()})
	if err != nil {
		s.logger.Error("grpc refresh token failed", zap.Error(err))
		return nil, grpcerrs.ToGRPCError(err)
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}
