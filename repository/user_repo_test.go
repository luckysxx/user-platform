package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/luckysxx/user-platform/common/dberr"
	"github.com/luckysxx/user-platform/db"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

func newMockUserRepository(t *testing.T) (UserRepository, sqlmock.Sqlmock, func()) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("创建 sqlmock 失败: %v", err)
	}

	repo := NewUserRepository(db.New(sqlDB))
	cleanup := func() {
		sqlDB.Close()
	}

	return repo, mock, cleanup
}

func TestUserRepository_Create_Success(t *testing.T) {
	repo, mock, cleanup := newMockUserRepository(t)
	defer cleanup()

	createdAt := time.Now()
	params := &db.CreateUserParams{
		Username: "alice",
		Password: "hashed_pwd",
		Email:    "alice@example.com",
	}

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(params.Username, params.Password, params.Email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "email", "created_at"}).
			AddRow(int64(1), params.Username, params.Password, params.Email, createdAt))

	user, err := repo.Create(context.Background(), params)
	if err != nil {
		t.Fatalf("Create() 返回错误: %v", err)
	}

	if user.ID != 1 || user.Username != params.Username || user.Email != params.Email {
		t.Fatalf("Create() 返回值不符合预期: %+v", user)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL 期望未满足: %v", err)
	}
}

func TestUserRepository_Create_MapsDBError(t *testing.T) {
	repo, mock, cleanup := newMockUserRepository(t)
	defer cleanup()

	params := &db.CreateUserParams{
		Username: "alice",
		Password: "hashed_pwd",
		Email:    "alice@example.com",
	}

	mock.ExpectQuery("INSERT INTO users").
		WithArgs(params.Username, params.Password, params.Email).
		WillReturnError(&pq.Error{Code: dberr.PgErrUniqueViolation, Constraint: "users_username_key"})

	user, err := repo.Create(context.Background(), params)
	if user != nil {
		t.Fatal("Create() 出错时应返回 nil user")
	}
	if !errors.Is(err, dberr.ErrUsernameDuplicate) {
		t.Fatalf("期望 ErrUsernameDuplicate, 实际: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL 期望未满足: %v", err)
	}
}

func TestUserRepository_GetByUsername_Success(t *testing.T) {
	repo, mock, cleanup := newMockUserRepository(t)
	defer cleanup()

	createdAt := time.Now()
	username := "alice"

	mock.ExpectQuery("SELECT id, username, password, email, created_at FROM users").
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "email", "created_at"}).
			AddRow(int64(1), username, "hashed_pwd", "alice@example.com", createdAt))

	user, err := repo.GetByUsername(context.Background(), username)
	if err != nil {
		t.Fatalf("GetByUsername() 返回错误: %v", err)
	}

	if user.Username != username || user.Email != "alice@example.com" {
		t.Fatalf("GetByUsername() 返回值不符合预期: %+v", user)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL 期望未满足: %v", err)
	}
}

func TestUserRepository_GetByUsername_NotFound(t *testing.T) {
	repo, mock, cleanup := newMockUserRepository(t)
	defer cleanup()

	username := "missing"
	mock.ExpectQuery("SELECT id, username, password, email, created_at FROM users").
		WithArgs(username).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetByUsername(context.Background(), username)
	if user != nil {
		t.Fatal("GetByUsername() 未找到时应返回 nil user")
	}
	if !errors.Is(err, dberr.ErrNoRows) {
		t.Fatalf("期望 ErrNoRows, 实际: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL 期望未满足: %v", err)
	}
}
