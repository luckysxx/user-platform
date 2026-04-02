package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/luckysxx/common/crypto"
	mqevents "github.com/luckysxx/common/mq/events"
	mqoutbox "github.com/luckysxx/common/mq/outbox"
	mqtopics "github.com/luckysxx/common/mq/topics"
	"github.com/luckysxx/common/trace"
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
	outbox              mqoutbox.Writer
	logger              *zap.Logger
	topicUserRegistered string
}

func NewUserService(
	tm repository.TransactionManager,
	userRepo repository.UserRepository,
	profileRepo repository.ProfileRepository,
	outbox mqoutbox.Writer,
	logger *zap.Logger,
	topicUserRegistered string,
) UserService {
	return &userService{
		tm:                  tm,
		userRepo:            userRepo,
		profileRepo:         profileRepo,
		outbox:              outbox,
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

		// 先把"用户已注册"这个领域事件写入 Outbox，而不是在事务里直接发 Kafka。
		// 这样用户数据和事件记录能一起提交，避免双写不一致。

		// 构建 outbox headers，透传链路追踪 ID 到下游消费者
		outboxHeaders := map[string]string{}
		if traceID := trace.FromContext(txCtx); traceID != "" {
			outboxHeaders[trace.HeaderTraceID] = traceID
		}

		record, err := mqoutbox.NewJSONRecord(
			uuid.NewString(),
			"user",
			fmt.Sprintf("%d", user.ID),
			s.topicUserRegistered,
			mqevents.UserRegistered{
				Version:   mqevents.UserRegisteredVersion,
				EventType: mqtopics.UserRegistered,
				UserID:    user.ID,
				Email:     user.Email,
				Username:  user.Username,
				Timestamp: time.Now().Unix(),
			},
			outboxHeaders,
		)
		if err != nil {
			return fmt.Errorf("序列化用户注册事件失败: %w", err)
		}

		// Append 只负责把事件安全落库。
		// 当前版本真正的 Kafka 投递仍由旧 Worker 完成，后面会逐步切到 CDC。
		err = s.outbox.Append(txCtx, record)
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
