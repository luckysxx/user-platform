package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/luckysxx/user-platform/internal/auth"
	pkgerrs "github.com/luckysxx/user-platform/pkg/errs"
	"github.com/luckysxx/user-platform/internal/dberr"
	"github.com/luckysxx/user-platform/internal/repository"
	servicecontract "github.com/luckysxx/user-platform/internal/service/contract"
	"github.com/luckysxx/user-platform/pkg/crypto"
	"github.com/luckysxx/user-platform/pkg/ratelimiter"
	"go.uber.org/zap"
)

// 将错误定义移回所属的域，并统一升级为领域模型错误
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
	repo       repository.UserRepository
	appRepo    repository.AppRepository
	session    repository.SessionRepository
	jwtManager *auth.JWTManager
	limiter    ratelimiter.Limiter
	logger     *zap.Logger
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
		repo:       repo,
		appRepo:    appRepo,
		session:    session,
		jwtManager: jwtManager,
		limiter:    limiter,
		logger:     logger,
	}
}

func (s *authService) Login(ctx context.Context, req *servicecontract.LoginCommand) (*servicecontract.LoginResult, error) {
	// 限制同一个 Username 的高频尝试
	limiterKey := fmt.Sprintf("rl:login:user:%s", req.Username)
	if err := s.limiter.Allow(ctx, limiterKey, 5, 15*60*1000000000); err != nil { // 15分钟(纳秒)
		if errors.Is(err, ratelimiter.ErrRateLimitExceeded) {
			s.logger.Warn("防止暴力破解, 账号登录被限流", zap.String("username", req.Username))
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
		// 验签失败属于常规情况（如过期），包装返回由最外层决定记录级别即可
		return nil, fmt.Errorf("Token验证失败: %w", err)
	}
	return &servicecontract.VerifyTokenResult{
		UserID:   claims.UserID,
		Username: claims.Username,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, req *servicecontract.RefreshTokenCommand) (*servicecontract.RefreshTokenResult, error) {
	// 1. 底层解析验证 Hash mapping 完全解耦：查验并获取 UserID
	userID, deviceID, err := s.session.GetSessionByToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	// 2. 验证 Hash 表中该设备的 Token 也对得上
	if err := s.session.ValidateDeviceToken(ctx, userID, deviceID, req.Token); err != nil {
		return nil, err
	}

	// 3. 轮换自毁：删掉旧的逆向索引
	s.session.DeleteTokenIndex(ctx, req.Token)

	// 4. 去数据库查一下该用户的最新状态（获取最新的 username，并拦截被封禁的用户）
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("试图刷新Token但用户(uid:%d)不存在: %w", userID, ErrAccountAbnormal)
	}

	// 5. 重新签发一套全新的 Access Token 和 Refresh Token 给相同的 deviceID
	result, err := s.issueTokens(ctx, user.ID, user.Username, deviceID)
	if err != nil {
		return nil, err
	}

	// 5. 映射为 RefreshToken 的专属返回结构
	return &servicecontract.RefreshTokenResult{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
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

	s.logger.Info("用户退出设备", zap.Int64("user_id", req.UserID), zap.String("device_id", req.DeviceID))
	return nil
}
