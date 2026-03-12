package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/luckysxx/user-platform/internal/db"
	servicecontract "github.com/luckysxx/user-platform/internal/service/contract"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	createFn   func(ctx context.Context, params *db.CreateUserParams) (*db.User, error)
	getFn      func(ctx context.Context, username string) (*db.User, error)
	createCall int
	getCall    int
}

func (m *mockUserRepo) Create(ctx context.Context, params *db.CreateUserParams) (*db.User, error) {
	m.createCall++
	if m.createFn != nil {
		return m.createFn(ctx, params)
	}
	return nil, errors.New("createFn not implemented")
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*db.User, error) {
	m.getCall++
	if m.getFn != nil {
		return m.getFn(ctx, username)
	}
	return nil, errors.New("getFn not implemented")
}

func TestUserService_Register_Success(t *testing.T) {
	repo := &mockUserRepo{
		createFn: func(ctx context.Context, params *db.CreateUserParams) (*db.User, error) {
			if params.Username != "alice" || params.Email != "alice@example.com" {
				t.Fatalf("Create 参数不正确: %+v", params)
			}

			if params.Password == "plain_password" {
				t.Fatal("密码应当被加密后存储")
			}

			if err := bcrypt.CompareHashAndPassword([]byte(params.Password), []byte("plain_password")); err != nil {
				t.Fatalf("密码哈希校验失败: %v", err)
			}

			return &db.User{
				ID:        1,
				Username:  params.Username,
				Password:  params.Password,
				Email:     params.Email,
				CreatedAt: time.Now(),
			}, nil
		},
	}

	svc := NewUserService(repo, auth.NewJWTManager("test-secret"), zap.NewNop())

	resp, err := svc.Register(context.Background(), &servicecontract.RegisterCommand{
		Username: "alice",
		Password: "plain_password",
		Email:    "alice@example.com",
	})
	if err != nil {
		t.Fatalf("Register() 返回错误: %v", err)
	}

	if resp.UserID != 1 || resp.Username != "alice" || resp.Email != "alice@example.com" {
		t.Fatalf("Register() 返回值不符合预期: %+v", resp)
	}
	if repo.createCall != 1 {
		t.Fatalf("repo.Create 应调用 1 次, 实际: %d", repo.createCall)
	}
}

func TestUserService_Register_RepoError(t *testing.T) {
	expectedErr := errors.New("db failed")
	repo := &mockUserRepo{
		createFn: func(ctx context.Context, params *db.CreateUserParams) (*db.User, error) {
			return nil, expectedErr
		},
	}

	svc := NewUserService(repo, auth.NewJWTManager("test-secret"), zap.NewNop())

	resp, err := svc.Register(context.Background(), &servicecontract.RegisterCommand{
		Username: "alice",
		Password: "plain_password",
		Email:    "alice@example.com",
	})
	if resp != nil {
		t.Fatal("Register() 出错时应返回 nil response")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("期望错误 %v, 实际: %v", expectedErr, err)
	}
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{
		getFn: func(ctx context.Context, username string) (*db.User, error) {
			return nil, sql.ErrNoRows
		},
	}

	svc := NewUserService(repo, auth.NewJWTManager("test-secret"), zap.NewNop())

	resp, err := svc.Login(context.Background(), &servicecontract.LoginCommand{
		Username: "missing",
		Password: "password",
	})
	if resp != nil {
		t.Fatal("Login() 失败时应返回 nil response")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("期望 ErrInvalidCredentials, 实际: %v", err)
	}
}

func TestUserService_Login_RepoError(t *testing.T) {
	expectedErr := errors.New("db down")
	repo := &mockUserRepo{
		getFn: func(ctx context.Context, username string) (*db.User, error) {
			return nil, expectedErr
		},
	}

	svc := NewUserService(repo, auth.NewJWTManager("test-secret"), zap.NewNop())

	resp, err := svc.Login(context.Background(), &servicecontract.LoginCommand{
		Username: "alice",
		Password: "password",
	})
	if resp != nil {
		t.Fatal("Login() 出错时应返回 nil response")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("期望错误 %v, 实际: %v", expectedErr, err)
	}
}

func TestUserService_Login_InvalidPassword(t *testing.T) {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("准备哈希密码失败: %v", err)
	}

	repo := &mockUserRepo{
		getFn: func(ctx context.Context, username string) (*db.User, error) {
			return &db.User{
				ID:        1,
				Username:  username,
				Password:  string(hashedPwd),
				Email:     "alice@example.com",
				CreatedAt: time.Now(),
			}, nil
		},
	}

	svc := NewUserService(repo, auth.NewJWTManager("test-secret"), zap.NewNop())

	resp, err := svc.Login(context.Background(), &servicecontract.LoginCommand{
		Username: "alice",
		Password: "wrong_password",
	})
	if resp != nil {
		t.Fatal("Login() 密码错误时应返回 nil response")
	}
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("期望 ErrInvalidCredentials, 实际: %v", err)
	}
}

func TestUserService_Login_Success(t *testing.T) {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("准备哈希密码失败: %v", err)
	}

	user := &db.User{
		ID:        99,
		Username:  "alice",
		Password:  string(hashedPwd),
		Email:     "alice@example.com",
		CreatedAt: time.Now(),
	}

	repo := &mockUserRepo{
		getFn: func(ctx context.Context, username string) (*db.User, error) {
			if username != user.Username {
				t.Fatalf("GetByUsername 参数不正确, got=%s", username)
			}
			return user, nil
		},
	}

	jwtManager := auth.NewJWTManager("test-secret")
	svc := NewUserService(repo, jwtManager, zap.NewNop())

	resp, err := svc.Login(context.Background(), &servicecontract.LoginCommand{
		Username: "alice",
		Password: "correct_password",
	})
	if err != nil {
		t.Fatalf("Login() 返回错误: %v", err)
	}

	if resp.Token == "" {
		t.Fatal("登录成功后应返回非空 token")
	}
	if resp.UserID != user.ID || resp.Username != user.Username || resp.Email != user.Email {
		t.Fatalf("Login() 返回值不符合预期: %+v", resp)
	}

	claims, err := jwtManager.ParseToken(resp.Token)
	if err != nil {
		t.Fatalf("返回 token 解析失败: %v", err)
	}
	if claims.UserID != user.ID {
		t.Fatalf("token UserID 不正确, 期望=%d 实际=%d", user.ID, claims.UserID)
	}
}
