package server

import (
	"context"
	"strings"

	pb "github.com/luckysxx/common/proto/auth"
	"github.com/luckysxx/user-platform/internal/service"
	servicecontract "github.com/luckysxx/user-platform/internal/service/contract"
	grpcerrs "github.com/luckysxx/user-platform/internal/transport/grpc/errs"
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
	if strings.TrimSpace(req.GetUsername()) == "" || strings.TrimSpace(req.GetPassword()) == "" || strings.TrimSpace(req.GetAppCode()) == "" {
		return nil, status.Error(codes.InvalidArgument, "username/password/app_code are required")
	}

	resp, err := s.avc.Login(ctx, &servicecontract.LoginCommand{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		AppCode:  req.GetAppCode(),
	})
	if err != nil {
		return nil, grpcerrs.ToGRPCError(err)
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
		return nil, grpcerrs.ToGRPCError(err)
	}

	return &pb.RefreshTokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	// TODO: 目前尚未实现全局 gRPC Interceptor 以解析 JWT 获取 UserID。
	// 这里暂时屏蔽 gRPC 的登出实现，前端如果走 WebSocket+gRPC 需要在网关层或引入 Interceptor 处理鉴权。
	return nil, status.Error(codes.Unimplemented, "Logout via gRPC requires Auth Interceptor which is pending")
}
