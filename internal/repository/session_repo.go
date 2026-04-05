package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/luckysxx/user-platform/internal/service/contract"
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
	TryLock(ctx context.Context, key string, expiration time.Duration) (bool, error)
	UnLock(ctx context.Context, key string) error
	SaveGracePeriod(ctx context.Context, oldToken string, newToken contract.RefreshTokenResult, duration time.Duration) error
	CheckGracePeriod(ctx context.Context, oldToken string) (*contract.RefreshTokenResult, bool)
}

type redisSessionRepo struct {
	cli *redis.Client
}

func NewRedisSessionRepo(cli *redis.Client) SessionRepository {
	return &redisSessionRepo{cli: cli}
}

func (r *redisSessionRepo) SaveDeviceSession(ctx context.Context, userID int64, deviceID string, refreshToken string, duration time.Duration) error {
	hashKey := fmt.Sprintf("user_sessions:%d", userID)

	// 0. 清理旧 session 的逆向索引（防止重新登录后产生 orphan key）
	if oldToken, err := r.cli.HGet(ctx, hashKey, deviceID).Result(); err == nil && oldToken != "" {
		r.cli.Del(ctx, fmt.Sprintf("refresh_token:%s", oldToken))
	}

	pipe := r.cli.TxPipeline()

	// a. 逆向索引
	redisKey := fmt.Sprintf("refresh_token:%s", refreshToken)
	val := fmt.Sprintf("%d:%s", userID, deviceID)
	pipe.Set(ctx, redisKey, val, duration)

	// b. 正向哈希索引
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

func (r *redisSessionRepo) TryLock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return r.cli.SetNX(ctx, key, "locked", expiration).Result()
}

func (r *redisSessionRepo) UnLock(ctx context.Context, key string) error {
	return r.cli.Del(ctx, key).Err()
}

func (r *redisSessionRepo) SaveGracePeriod(ctx context.Context, oldToken string, newToken contract.RefreshTokenResult, duration time.Duration) error {
	// Redis 原生不支持直接存 Go Struct，需要拼成字符串或者 JSON。最快的方法是用管道符拼接：
	val := fmt.Sprintf("%s|%s", newToken.AccessToken, newToken.RefreshToken)
	return r.cli.Set(ctx, fmt.Sprintf("grace_period:%s", oldToken), val, duration).Err()
}

func (r *redisSessionRepo) CheckGracePeriod(ctx context.Context, oldToken string) (*contract.RefreshTokenResult, bool) {
	val, err := r.cli.Get(ctx, fmt.Sprintf("grace_period:%s", oldToken)).Result()
	if err != nil {
		return nil, false
	}
	
	// 把之前存进去的字符串重新拆成结构体
	parts := strings.Split(val, "|")
	if len(parts) == 2 {
		return &contract.RefreshTokenResult{
			AccessToken:  parts[0],
			RefreshToken: parts[1],
		}, true
	}
	return nil, false
}
