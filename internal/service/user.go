package service

import (
	"context"

	"github.com/luckysxx/user-platform/internal/event"
	"github.com/luckysxx/user-platform/internal/repository"
	servicecontract "github.com/luckysxx/user-platform/internal/service/contract"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, req *servicecontract.RegisterCommand) (*servicecontract.RegisterResult, error)
}

type userService struct {
	repo      repository.UserRepository
	Publisher event.Publisher
	logger    *zap.Logger
}

func NewUserService(repo repository.UserRepository, publisher event.Publisher, logger *zap.Logger) UserService {
	return &userService{repo: repo, Publisher: publisher, logger: logger}
}

func (s *userService) Register(ctx context.Context, req *servicecontract.RegisterCommand) (*servicecontract.RegisterResult, error) {
	// 加密密码
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("密码加密失败", zap.Error(err))
		return nil, err
	}

	// 调用数据库创建用户
	user, err := s.repo.Create(ctx, req.Email, req.Username, string(hashedPwd))
	if err != nil {
		s.logger.Error("创建用户失败", zap.Error(err))
		return nil, err
	}

	resp := &servicecontract.RegisterResult{
		Email:    user.Email,
		UserID:   user.ID,
		Username: user.Username,
	}

	// 发送 Kafka 消息
	go func() {
		err := s.Publisher.PublishUserRegistered(ctx, user.ID, user.Email)
		if err != nil {
			s.logger.Error("发送用户注册事件失败", zap.Error(err))
		}
	}()

	return resp, nil
}
