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

// UserServer 是用户服务的 gRPC 接口实现。
type UserServer struct {
	pb.UnimplementedUserServiceServer
	svc        service.UserService
	profileSvc service.ProfileService
	logger     *zap.Logger
}

// NewUserServer 创建一个用户服务 gRPC Server。
func NewUserServer(svc service.UserService, profileSvc service.ProfileService, logger *zap.Logger) *UserServer {
	return &UserServer{svc: svc, profileSvc: profileSvc, logger: logger}
}

// Register 处理用户注册的 gRPC 请求。
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

// GetProfile 处理获取用户资料的 gRPC 请求。
func (s *UserServer) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.ProfileResponse, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	resp, err := s.profileSvc.GetProfile(ctx, &servicecontract.GetProfileQuery{
		UserID: req.GetUserId(),
	})
	if err != nil {
		return nil, grpcerrs.ToGRPCError(err)
	}

	return &pb.ProfileResponse{
		UserId:    resp.UserID,
		Nickname:  resp.Nickname,
		AvatarUrl: resp.AvatarURL,
		Bio:       resp.Bio,
		Birthday:  resp.Birthday,
		UpdatedAt: resp.UpdatedAt,
	}, nil
}

// UpdateProfile 处理更新用户资料的 gRPC 请求。
func (s *UserServer) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.ProfileResponse, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	resp, err := s.profileSvc.UpdateProfile(ctx, &servicecontract.UpdateProfileCommand{
		UserID:    req.GetUserId(),
		Nickname:  req.GetNickname(),
		AvatarURL: req.GetAvatarUrl(),
		Bio:       req.GetBio(),
		Birthday:  req.GetBirthday(),
	})
	if err != nil {
		return nil, grpcerrs.ToGRPCError(err)
	}

	return &pb.ProfileResponse{
		UserId:    resp.UserID,
		Nickname:  resp.Nickname,
		AvatarUrl: resp.AvatarURL,
		Bio:       resp.Bio,
		Birthday:  resp.Birthday,
		UpdatedAt: resp.UpdatedAt,
	}, nil
}
