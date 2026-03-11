package repository

import (
	"context"
	"github.com/luckysxx/user-platform/common/dberr"
	"github.com/luckysxx/user-platform/db"
)

type UserRepository interface {
	Create(ctx context.Context, parms *db.CreateUserParams) (*db.User, error)
	GetByUsername(ctx context.Context, username string) (*db.User, error)
}

type userRepository struct {
	q *db.Queries
}

func NewUserRepository(q *db.Queries) UserRepository {
	return &userRepository{q: q}
}

func (r *userRepository) Create(ctx context.Context, parms *db.CreateUserParams) (*db.User, error) {
	User, err := r.q.CreateUser(ctx, *parms)
	if err != nil {
		return nil, dberr.ParseDBError(err)
	}
	return &User, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*db.User, error) {
	user, err := r.q.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, dberr.ParseDBError(err)
	}
	return &user, nil
}
