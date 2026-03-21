package service

import (
	"context"
	"fmt"

	"github.com/luckysxx/common/crypto"
	"github.com/luckysxx/user-platform/internal/event"
	"github.com/luckysxx/user-platform/internal/repository"
	servicecontract "github.com/luckysxx/user-platform/internal/service/contract"

	"go.uber.org/zap"
)

type UserService interface {
	Register(ctx context.Context, req *servicecontract.RegisterCommand) (*servicecontract.RegisterResult, error)
}

type userService struct {
	tm         repository.TransactionManager
	userRepo   repository.UserRepository
	outboxRepo repository.EventOutboxRepository
	Publisher  event.Publisher
	logger     *zap.Logger
}

func NewUserService(
	tm repository.TransactionManager,
	userRepo repository.UserRepository,
	outboxRepo repository.EventOutboxRepository,
	publisher event.Publisher,
	logger *zap.Logger,
) UserService {
	return &userService{
		tm:         tm,
		userRepo:   userRepo,
		outboxRepo: outboxRepo,
		Publisher:  publisher,
		logger:     logger,
	}
}

func (s *userService) Register(ctx context.Context, req *servicecontract.RegisterCommand) (*servicecontract.RegisterResult, error) {
	// 加密密码
	hashedPwd, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	var resp *servicecontract.RegisterResult

	// 开启事务
	err = s.tm.WithTx(ctx, func(txCtx context.Context) error {
		// 创建用户
		user, err := s.userRepo.Create(txCtx, req.Email, req.Username, hashedPwd)
		if err != nil {
			return fmt.Errorf("创建用户失败: %w", err)
		}

		resp = &servicecontract.RegisterResult{
			Email:    user.Email,
			UserID:   user.ID,
			Username: user.Username,
		}

		eventPayload := []byte(fmt.Sprintf(
			`{"user_id": %d, "email": "%s", "username": "%s"}`,
			user.ID,
			user.Email,
			user.Username,
		))

		err = s.outboxRepo.SaveEvent(txCtx, "UserRegistered", eventPayload)
		if err != nil {
			return fmt.Errorf("创建用户事件失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}
