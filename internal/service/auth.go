package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/luckysxx/user-platform/internal/dberr"
	"github.com/luckysxx/user-platform/internal/repository"
	servicecontract "github.com/luckysxx/user-platform/internal/service/contract"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// 将错误定义移回所属的域
var (
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrTokenGeneration    = errors.New("生成 Token 失败")
	ErrAccountAbnormal    = errors.New("账号异常或已被封禁")
	ErrAppNotFound        = errors.New("应用不存在")
)

type AuthService interface {
	Login(ctx context.Context, req *servicecontract.LoginCommand) (*servicecontract.LoginResult, error)
	VerifyToken(ctx context.Context, req *servicecontract.VerifyTokenCommand) (*servicecontract.VerifyTokenResult, error)
	RefreshToken(ctx context.Context, req *servicecontract.RefreshTokenCommand) (*servicecontract.RefreshTokenResult, error)
}

type authService struct {
	repo       repository.UserRepository
	redisCli   *redis.Client
	jwtManager *auth.JWTManager
	logger     *zap.Logger
}

func NewAuthService(repo repository.UserRepository, redisCli *redis.Client, jwtManager *auth.JWTManager, logger *zap.Logger) AuthService {
	return &authService{
		repo:       repo,
		redisCli:   redisCli,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

func (s *authService) Login(ctx context.Context, req *servicecontract.LoginCommand) (*servicecontract.LoginResult, error) {
	// 1. 获取用户
	user, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		if dberr.IsNotFoundError(err) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 2. 比较密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 3. 首次登录某应用时自动建立授权关系（幂等）。
	err = s.repo.EnsureAppAuthorized(ctx, user.ID, req.AppCode)
	if err != nil {
		if errors.Is(err, dberr.ErrNoRows) {
			return nil, ErrAppNotFound
		}
		return nil, fmt.Errorf("应用授权处理失败 (uid:%d, app_code:%s): %w", user.ID, req.AppCode, err)
	}

	// 4. 签发双 Token (提取为一个内部方法，复用逻辑)
	return s.issueTokens(ctx, user.ID, user.Username)
}

func (s *authService) VerifyToken(ctx context.Context, req *servicecontract.VerifyTokenCommand) (*servicecontract.VerifyTokenResult, error) {
	claims, err := s.jwtManager.VerifyToken(req.Token)
	if err != nil {
		// 验签失败属于常规情况（如过期），包装返回由最外层决定记录级别即可
		return nil, fmt.Errorf("Token验证失败: %w", err)
	}
	return &servicecontract.VerifyTokenResult{
		UserID:   claims.UserID,
		Username: claims.Username,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, req *servicecontract.RefreshTokenCommand) (*servicecontract.RefreshTokenResult, error) {
	// 1. 从 Redis 中验证 Refresh Token 是否有效 (注意增加前缀)
	redisKey := fmt.Sprintf("refresh_token:%s", req.Token)
	userID, err := s.redisCli.Get(ctx, redisKey).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, auth.ErrInvalidOrExpiredToken
		}
		return nil, fmt.Errorf("查询Refresh Token失败: %w", err)
	}

	// 2. 轮换自毁，从 Redis 删掉这个用过的 RT
	s.redisCli.Del(ctx, redisKey)

	// 3. 去数据库查一下该用户的最新状态（获取最新的 username，并拦截被封禁的用户）
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("试图刷新Token但用户(uid:%d)不存在: %w", userID, ErrAccountAbnormal)
	}

	// 4. 重新签发一套全新的 Access Token 和 Refresh Token
	result, err := s.issueTokens(ctx, user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	// 5. 映射为 RefreshToken 的专属返回结构
	return &servicecontract.RefreshTokenResult{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

// 内部方法：统一处理双 Token 的签发与 Redis 存储
func (s *authService) issueTokens(ctx context.Context, userID int64, username string) (*servicecontract.LoginResult, error) {
	// 1. 生成 Access Token
	accessToken, err := s.jwtManager.GenerateAccessToken(userID, username)
	if err != nil {
		return nil, fmt.Errorf("生成 Access Token 失败 (uid:%d, cause:%v): %w", userID, err, ErrTokenGeneration)
	}

	// 2. 生成纯 UUID 的 Refresh Token
	refreshToken := uuid.New().String()
	redisKey := fmt.Sprintf("refresh_token:%s", refreshToken)

	// 3. 存储到 Redis
	if err := s.redisCli.Set(ctx, redisKey, userID, auth.RefreshTokenDuration).Err(); err != nil {
		return nil, fmt.Errorf("存储 Refresh Token 失败 (uid:%d, cause:%v): %w", userID, err, ErrTokenGeneration)
	}

	return &servicecontract.LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       userID,
		Username:     username,
	}, nil
}
