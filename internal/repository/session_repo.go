package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/redis/go-redis/v9"
)

// SessionRepository 屏蔽了到底是用 Redis 还是用其他缓存/数据库存 Session 的底层细节。
// 业务层（AuthService）只需要面向这个接口编程。
type SessionRepository interface {
	SaveDeviceSession(ctx context.Context, userID int64, deviceID string, refreshToken string, duration time.Duration) error
	GetSessionByToken(ctx context.Context, refreshToken string) (userID int64, deviceID string, err error)
	ValidateDeviceToken(ctx context.Context, userID int64, deviceID string, candidateToken string) error
	DeleteTokenIndex(ctx context.Context, refreshToken string) error
	DeleteDeviceSession(ctx context.Context, userID int64, deviceID string) (oldToken string, err error)
}

type redisSessionRepo struct {
	cli *redis.Client
}

func NewRedisSessionRepo(cli *redis.Client) SessionRepository {
	return &redisSessionRepo{cli: cli}
}

func (r *redisSessionRepo) SaveDeviceSession(ctx context.Context, userID int64, deviceID string, refreshToken string, duration time.Duration) error {
	pipe := r.cli.TxPipeline()
	
	// a. 逆向索引
	redisKey := fmt.Sprintf("refresh_token:%s", refreshToken)
	val := fmt.Sprintf("%d:%s", userID, deviceID)
	pipe.Set(ctx, redisKey, val, duration)

	// b. 正向哈希索引
	hashKey := fmt.Sprintf("user_sessions:%d", userID)
	pipe.HSet(ctx, hashKey, deviceID, refreshToken)
	pipe.Expire(ctx, hashKey, duration)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis transaction failed: %w", err)
	}
	return nil
}

func (r *redisSessionRepo) GetSessionByToken(ctx context.Context, refreshToken string) (int64, string, error) {
	redisKey := fmt.Sprintf("refresh_token:%s", refreshToken)
	val, err := r.cli.Get(ctx, redisKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, "", auth.ErrInvalidOrExpiredToken
		}
		return 0, "", fmt.Errorf("redis get failed: %w", err)
	}

	parts := strings.SplitN(val, ":", 2)
	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid format in redis: %s", val)
	}
	
	userID, _ := strconv.ParseInt(parts[0], 10, 64)
	return userID, parts[1], nil
}

func (r *redisSessionRepo) ValidateDeviceToken(ctx context.Context, userID int64, deviceID string, candidateToken string) error {
	hashKey := fmt.Sprintf("user_sessions:%d", userID)
	savedToken, err := r.cli.HGet(ctx, hashKey, deviceID).Result()
	if err != nil || savedToken != candidateToken {
		return auth.ErrInvalidOrExpiredToken
	}
	return nil
}

func (r *redisSessionRepo) DeleteTokenIndex(ctx context.Context, refreshToken string) error {
	return r.cli.Del(ctx, fmt.Sprintf("refresh_token:%s", refreshToken)).Err()
}

func (r *redisSessionRepo) DeleteDeviceSession(ctx context.Context, userID int64, deviceID string) (string, error) {
	hashKey := fmt.Sprintf("user_sessions:%d", userID)
	// 先获取这台设备关联的旧 token
	oldToken, err := r.cli.HGet(ctx, hashKey, deviceID).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("redis hget failed: %w", err)
	}
	
	// 从用户的设备集合中彻底剔除该设备
	if err := r.cli.HDel(ctx, hashKey, deviceID).Err(); err != nil {
		return "", fmt.Errorf("redis hdel failed: %w", err)
	}
	
	return oldToken, nil
}
