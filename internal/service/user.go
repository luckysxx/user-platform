package service

import (
	"context"
	"fmt"

	"github.com/luckysxx/common/crypto"
	"github.com/luckysxx/common/trace"
	"github.com/luckysxx/user-platform/internal/event"
	"github.com/luckysxx/user-platform/internal/repository"
	servicecontract "github.com/luckysxx/user-platform/internal/service/contract"

	"go.uber.org/zap"
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
	hashedPwd, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	// 调用数据库创建用户
	user, err := s.repo.Create(ctx, req.Email, req.Username, hashedPwd)
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	resp := &servicecontract.RegisterResult{
		Email:    user.Email,
		UserID:   user.ID,
		Username: user.Username,
	}

	// 从当前请求提取 TraceID 以传给后台任务
	traceID := trace.FromContext(ctx)

	// 发送 Kafka 消息
	go func() {
		// ⚠️ 极其经典的坑：不能在新的协程中复用来自于前台（比如 HTTP 请求）的 ctx
		// 解决办法：传入一个独立的上下文，但需要手动把 traceID 带过去
		bgCtx := trace.IntoContext(context.Background(), traceID)
		err := s.Publisher.PublishUserRegistered(bgCtx, user.ID, user.Email, user.Username)
		if err != nil {
			s.logger.Error("跨服务发送用户注册事件失败", zap.Error(err))
		}
	}()

	return resp, nil
}
