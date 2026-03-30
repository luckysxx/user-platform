package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/luckysxx/common/crypto"
	mqevents "github.com/luckysxx/common/mq/events"
	mqtopics "github.com/luckysxx/common/mq/topics"
	"github.com/luckysxx/user-platform/internal/event"
	"github.com/luckysxx/user-platform/internal/repository"
	servicecontract "github.com/luckysxx/user-platform/internal/service/contract"

	"go.uber.org/zap"
)

type UserService interface {
	Register(ctx context.Context, req *servicecontract.RegisterCommand) (*servicecontract.RegisterResult, error)
}

type userService struct {
	tm                  repository.TransactionManager
	userRepo            repository.UserRepository
	profileRepo         repository.ProfileRepository
	outboxRepo          repository.EventOutboxRepository
	Publisher           event.Publisher
	logger              *zap.Logger
	topicUserRegistered string
}

func NewUserService(
	tm repository.TransactionManager,
	userRepo repository.UserRepository,
	profileRepo repository.ProfileRepository,
	outboxRepo repository.EventOutboxRepository,
	publisher event.Publisher,
	logger *zap.Logger,
	topicUserRegistered string,
) UserService {
	return &userService{
		tm:                  tm,
		userRepo:            userRepo,
		profileRepo:         profileRepo,
		outboxRepo:          outboxRepo,
		Publisher:           publisher,
		logger:              logger,
		topicUserRegistered: topicUserRegistered,
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

		// 为该用户创建空的 Profile
		_, err = s.profileRepo.CreateEmpty(txCtx, user.ID)
		if err != nil {
			return fmt.Errorf("创建用户资料失败: %w", err)
		}

		resp = &servicecontract.RegisterResult{
			Email:    user.Email,
			UserID:   user.ID,
			Username: user.Username,
		}

		eventPayload, err := json.Marshal(mqevents.UserRegistered{
			Version:   mqevents.UserRegisteredVersion,
			EventType: mqtopics.UserRegistered,
			UserID:    user.ID,
			Email:     user.Email,
			Username:  user.Username,
			Timestamp: time.Now().Unix(),
		})
		if err != nil {
			return fmt.Errorf("序列化用户注册事件失败: %w", err)
		}

		err = s.outboxRepo.SaveEvent(txCtx, s.topicUserRegistered, eventPayload)
		if err != nil {
			return fmt.Errorf("创建用户事件失败: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 事务提交成功后，向 Redis 发送唤醒信号
	_ = s.outboxRepo.Notify(context.WithoutCancel(ctx))

	return resp, nil
}
