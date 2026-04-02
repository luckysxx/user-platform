package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/google/uuid"
	"github.com/luckysxx/common/crypto"
	pkgerrs "github.com/luckysxx/common/errs"
	commonlogger "github.com/luckysxx/common/logger"
	"github.com/luckysxx/common/ratelimiter"
	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/luckysxx/user-platform/internal/dberr"
	"github.com/luckysxx/user-platform/internal/repository"
	servicecontract "github.com/luckysxx/user-platform/internal/service/contract"
	"go.uber.org/zap"
)

var (
	ErrInvalidCredentials   = pkgerrs.NewParamErr("用户名或密码错误", nil)
	ErrTokenGeneration      = pkgerrs.NewServerErr(errors.New("生成 Token 失败"))
	ErrAccountAbnormal      = pkgerrs.New(pkgerrs.Forbidden, "账号异常或已被封禁", nil)
	ErrAppNotFound          = pkgerrs.NewParamErr("应用不存在", nil)
	ErrTooManyLoginAttempts = pkgerrs.NewParamErr("尝试登录次数过多，请15分钟后再试", nil)
)

type AuthService interface {
	Login(ctx context.Context, req *servicecontract.LoginCommand) (*servicecontract.LoginResult, error)
	VerifyToken(ctx context.Context, req *servicecontract.VerifyTokenCommand) (*servicecontract.VerifyTokenResult, error)
	RefreshToken(ctx context.Context, req *servicecontract.RefreshTokenCommand) (*servicecontract.RefreshTokenResult, error)
	Logout(ctx context.Context, req *servicecontract.LogoutCommand) error
}

type authService struct {
	repo         repository.UserRepository
	appRepo      repository.AppRepository
	session      repository.SessionRepository
	jwtManager   *auth.JWTManager
	limiter      ratelimiter.Limiter
	logger       *zap.Logger
	requestGroup *singleflight.Group
}

func NewAuthService(
	repo repository.UserRepository,
	appRepo repository.AppRepository,
	session repository.SessionRepository,
	jwtManager *auth.JWTManager,
	limiter ratelimiter.Limiter,
	logger *zap.Logger,
) AuthService {
	return &authService{
		repo:         repo,
		appRepo:      appRepo,
		session:      session,
		jwtManager:   jwtManager,
		limiter:      limiter,
		logger:       logger,
		requestGroup: &singleflight.Group{},
	}
}

func (s *authService) Login(ctx context.Context, req *servicecontract.LoginCommand) (*servicecontract.LoginResult, error) {
	// 限制同一个 Username 的高频尝试
	limiterKey := fmt.Sprintf("rl:login:user:%s", req.Username)
	if err := s.limiter.Allow(ctx, limiterKey, 5, 15*60*1000000000); err != nil { // 15分钟(纳秒)
		if errors.Is(err, ratelimiter.ErrRateLimitExceeded) {
			commonlogger.Ctx(ctx, s.logger).Warn("防止暴力破解, 账号登录被限流", zap.String("username", req.Username))
			return nil, ErrTooManyLoginAttempts
		}
		// 理论上 Fail-Open 机制不会走到这里，但为了严谨
		return nil, fmt.Errorf("限流器验证失败: %w", err)
	}

	// 1. 获取用户
	user, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		if dberr.IsNotFoundError(err) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 2. 比较密码
	if !crypto.CheckPasswordHash(req.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	// 3. 首次登录某应用时自动建立授权关系（幂等）。
	err = s.appRepo.EnsureAppAuthorized(ctx, user.ID, req.AppCode)
	if err != nil {
		if errors.Is(err, dberr.ErrNoRows) {
			return nil, ErrAppNotFound
		}
		return nil, fmt.Errorf("应用授权处理失败 (uid:%d, app_code:%s): %w", user.ID, req.AppCode, err)
	}

	// 4. 签发双 Token (提取为一个内部方法，复用逻辑)
	return s.issueTokens(ctx, user.ID, user.Username, req.DeviceID)
}

func (s *authService) VerifyToken(ctx context.Context, req *servicecontract.VerifyTokenCommand) (*servicecontract.VerifyTokenResult, error) {
	claims, err := s.jwtManager.VerifyToken(req.Token)
	if err != nil {
		// 验签失败属于常规情况（如过期），返回 Unauthorized 错误
		return nil, pkgerrs.New(pkgerrs.Unauthorized, "无效的访问凭证或已过期", err)
	}
	return &servicecontract.VerifyTokenResult{
		UserID:   claims.UserID,
		Username: claims.Username,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, req *servicecontract.RefreshTokenCommand) (*servicecontract.RefreshTokenResult, error) {
	// 先查 Redis 极速通道（是否有 grace period 保护）
	if result, found := s.session.CheckGracePeriod(ctx, req.Token); found {
		return result, nil
	}

	// 使用 Singleflight 避免同一台机器上的重复击穿
	v, err, _ := s.requestGroup.Do(req.Token, func() (interface{}, error) {
		// key 使用旧 Token，因为我们要查的就是旧 Token 对应的用户
		lockKey := fmt.Sprintf("lock:refresh:%s", req.Token)

		//尝试获取锁
		locked, err := s.session.TryLock(ctx, lockKey, 5*time.Second)
		if err == nil && locked {
			defer s.session.UnLock(ctx, lockKey)

			// 先获取旧 Token 的关联信息（userID, deviceID）
			userID, deviceID, sessionErr := s.session.GetSessionByToken(ctx, req.Token)
			if sessionErr != nil {
				// 极端情况：刚好别人在我抢锁前0.01秒把它干掉了并存入了 Grace 表里
				if graceRes, ok := s.session.CheckGracePeriod(ctx, req.Token); ok {
					return graceRes, nil
				}
				return nil, sessionErr // 真找不到了，直接抛出，这是伪造的 Token
			}

			// 验证设备 Token
			if err := s.session.ValidateDeviceToken(ctx, userID, deviceID, req.Token); err != nil {
				return nil, err
			}

			// 轮换自毁：删掉旧的逆向索引
			s.session.DeleteTokenIndex(ctx, req.Token)

			// 去数据库查一下该用户的最新状态（获取最新的 username，并拦截被封禁的用户）
			user, err := s.repo.GetByID(ctx, userID)
			if err != nil {
				return nil, fmt.Errorf("试图刷新Token但用户(uid:%d)不存在: %w", userID, ErrAccountAbnormal)
			}

			// 重新签发一套全新的 Access Token 和 Refresh Token 给相同的 deviceID
			result, err := s.issueTokens(ctx, user.ID, user.Username, deviceID)
			if err != nil {
				return nil, err
			}

			// 映射为 RefreshToken 的专属返回结构
			res := &servicecontract.RefreshTokenResult{
				AccessToken:  result.AccessToken,
				RefreshToken: result.RefreshToken,
			}

			// 将新颁发的 Token 存入 Grace Period，这样并发的请求能够拿到新 Token
			s.session.SaveGracePeriod(ctx, req.Token, *res, 15*time.Second)
			return res, nil
		}
		// 锁被占用
		time.Sleep(200 * time.Millisecond)
		if graceRes, ok := s.session.CheckGracePeriod(ctx, req.Token); ok {
			return graceRes, nil
		}
		// 超过 200ms 还没轮到，直接判定为无效
		return nil, ErrInvalidCredentials
	})
	if err != nil {
		return nil, err
	}
	return v.(*servicecontract.RefreshTokenResult), nil
}

// 内部方法：统一处理双 Token 的签发与 Redis 设备级存储
func (s *authService) issueTokens(ctx context.Context, userID int64, username string, deviceID string) (*servicecontract.LoginResult, error) {
	// 1. 生成 Access Token
	accessToken, err := s.jwtManager.GenerateAccessToken(userID, username)
	if err != nil {
		return nil, fmt.Errorf("生成 Access Token 失败 (uid:%d, cause:%v): %w", userID, err, ErrTokenGeneration)
	}

	// 2. 生成纯 UUID 的 Refresh Token
	refreshToken := uuid.New().String()

	// 3. 存储到持久化层
	err = s.session.SaveDeviceSession(ctx, userID, deviceID, refreshToken, auth.RefreshTokenDuration)
	if err != nil {
		return nil, fmt.Errorf("存储 Session 失败 (uid:%d): %w", userID, err)
	}

	return &servicecontract.LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       userID,
		Username:     username,
	}, nil
}

// Logout 登出功能：只删除特定设备的 Session 信息，不影响当前用户的其他设备。
func (s *authService) Logout(ctx context.Context, req *servicecontract.LogoutCommand) error {
	// 交给底层：先删除哈希中的记录，如果有旧Token也一并清理
	oldToken, err := s.session.DeleteDeviceSession(ctx, req.UserID, req.DeviceID)
	if err != nil {
		return fmt.Errorf("注销设备失败: %w", err)
	}
	if oldToken != "" {
		s.session.DeleteTokenIndex(ctx, oldToken)
	}

	commonlogger.Ctx(ctx, s.logger).Info("用户退出设备", zap.Int64("user_id", req.UserID), zap.String("device_id", req.DeviceID))
	return nil
}
