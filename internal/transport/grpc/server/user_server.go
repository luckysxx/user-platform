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
	svc        service.UserService
	profileSvc service.ProfileService
	logger     *zap.Logger
}

func NewUserServer(svc service.UserService, profileSvc service.ProfileService, logger *zap.Logger) *UserServer {
	return &UserServer{svc: svc, profileSvc: profileSvc, logger: logger}
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
		UpdatedAt: resp.UpdatedAt,
	}, nil
}

func (s *UserServer) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.ProfileResponse, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	resp, err := s.profileSvc.UpdateProfile(ctx, &servicecontract.UpdateProfileCommand{
		UserID:    req.GetUserId(),
		Nickname:  req.GetNickname(),
		AvatarURL: req.GetAvatarUrl(),
		Bio:       req.GetBio(),
	})
	if err != nil {
		return nil, grpcerrs.ToGRPCError(err)
	}

	return &pb.ProfileResponse{
		UserId:    resp.UserID,
		Nickname:  resp.Nickname,
		AvatarUrl: resp.AvatarURL,
		Bio:       resp.Bio,
		UpdatedAt: resp.UpdatedAt,
	}, nil
}
