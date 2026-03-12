package service

import (
	"context"
	"errors"

	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/luckysxx/user-platform/internal/db"
	"github.com/luckysxx/user-platform/internal/dberr"
	"github.com/luckysxx/user-platform/internal/repository"
	servicecontract "github.com/luckysxx/user-platform/internal/service/contract"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("用户名或密码错误")
	ErrTokenGeneration    = errors.New("生成 Token 失败")
)

type UserService interface {
	Register(ctx context.Context, req *servicecontract.RegisterCommand) (*servicecontract.RegisterResult, error)
	Login(ctx context.Context, req *servicecontract.LoginCommand) (*servicecontract.LoginResult, error)
}

type userService struct {
	repo       repository.UserRepository
	logger     *zap.Logger
	jwtManager *auth.JWTManager
}

func NewUserService(repo repository.UserRepository, jwtManager *auth.JWTManager, logger *zap.Logger) UserService {
	return &userService{repo: repo, jwtManager: jwtManager, logger: logger}
}

func (s *userService) Register(ctx context.Context, req *servicecontract.RegisterCommand) (*servicecontract.RegisterResult, error) {
	// 加密密码
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("密码加密失败", zap.Error(err))
		return nil, err
	}

	// 构造参数
	params := &db.CreateUserParams{
		Username: req.Username,
		Password: string(hashedPwd),
		Email:    req.Email,
	}

	// 调用数据库创建用户
	user, err := s.repo.Create(ctx, params)
	if err != nil {
		s.logger.Error("创建用户失败", zap.Error(err))
		return nil, err
	}

	resp := &servicecontract.RegisterResult{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	return resp, nil
}

func (s *userService) Login(ctx context.Context, req *servicecontract.LoginCommand) (*servicecontract.LoginResult, error) {
	// 根据用户名获取用户
	user, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		// 检查是否是 NotFound 错误
		if dberr.IsNotFoundError(err) {
			// 返回领域错误，不泄露具体是用户名错误
			return nil, ErrInvalidCredentials
		}
		// 其他数据库错误
		s.logger.Error("查询用户失败", zap.Error(err))
		return nil, err
	}

	// 比较密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 生成 JWT Token
	token, err := s.jwtManager.GenerateToken(user.ID)
	if err != nil {
		s.logger.Error("生成 Token 失败", zap.Error(err))
		return nil, ErrTokenGeneration
	}

	// 生成登录响应
	resp := &servicecontract.LoginResult{
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	return resp, nil
}
