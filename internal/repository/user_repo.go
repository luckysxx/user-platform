package repository

import (
	"context"
	"fmt"

	"github.com/luckysxx/common/rpc"
	"github.com/luckysxx/user-platform/internal/dberr"
	"github.com/luckysxx/user-platform/internal/ent"
	"github.com/luckysxx/user-platform/internal/ent/user"
)

type UserRepository interface {
	Create(ctx context.Context, email string, username string, passwordhash string) (*ent.User, error)
	GetByUsername(ctx context.Context, username string) (*ent.User, error)
	GetByID(ctx context.Context, id int64) (*ent.User, error)
}

type userRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) UserRepository {
	return &userRepository{client: client}
}

func (r *userRepository) Create(ctx context.Context, email string, username string, passwordhash string) (*ent.User, error) {
	// 1. 从远端发号器获取全局唯一的雪花 ID
	newID, err := rpc.GenerateID(ctx)
	if err != nil {
		return nil, fmt.Errorf("生成 Snowflake ID 失败: %w", err)
	}

	// 检查是否在事务中
	tx := ent.TxFromContext(ctx)
	// 如果在事务中, 使用事务
	if tx != nil {
		u, err := tx.User.Create().
			SetID(newID).
			SetEmail(email).
			SetUsername(username).
			SetPassword(passwordhash).
			Save(ctx)
		if err != nil {
			return nil, dberr.ParseDBError(err)
		}
		return u, nil
	}
	// 没有事务, 使用普通单表落库
	u, err := r.client.User.Create().
		SetID(newID).
		SetEmail(email).
		SetUsername(username).
		SetPassword(passwordhash).
		Save(ctx)
	if err != nil {
		return nil, dberr.ParseDBError(err)
	}
	return u, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*ent.User, error) {
	u, err := r.client.User.
		Query().
		Where(user.UsernameEQ(username)).
		Where(user.StatusEQ(user.StatusActive)).
		Only(ctx)
	if err != nil {
		return nil, dberr.ParseDBError(err)
	}
	return u, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*ent.User, error) {
	u, err := r.client.User.
		Query().
		Where(user.IDEQ(id)).
		Where(user.StatusEQ(user.StatusActive)).
		Only(ctx)
	if err != nil {
		return nil, dberr.ParseDBError(err)
	}
	return u, nil
}
