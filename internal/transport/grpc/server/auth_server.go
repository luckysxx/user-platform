package server

import (
	"context"
	"strings"

	pb "github.com/luckysxx/common/proto/auth"
	"github.com/luckysxx/user-platform/internal/service"
	servicecontract "github.com/luckysxx/user-platform/internal/service/contract"
	grpcerrs "github.com/luckysxx/user-platform/internal/transport/grpc/errs"
	"github.com/luckysxx/user-platform/internal/transport/grpc/interceptor"
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
		DeviceID: req.GetDeviceId(),
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
	userID, err := interceptor.UserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.avc.Logout(ctx, &servicecontract.LogoutCommand{
		UserID:   userID,
		DeviceID: req.GetDeviceId(),
	})
	if err != nil {
		return nil, grpcerrs.ToGRPCError(err)
	}

	return &pb.LogoutResponse{}, nil
}

func (s *AuthServer) VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.VerifyTokenResponse, error) {
	if strings.TrimSpace(req.GetToken()) == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	resp, err := s.avc.VerifyToken(ctx, &servicecontract.VerifyTokenCommand{Token: req.GetToken()})
	if err != nil {
		return nil, grpcerrs.ToGRPCError(err)
	}

	return &pb.VerifyTokenResponse{
		UserId:   resp.UserID,
		Username: resp.Username,
	}, nil
}
