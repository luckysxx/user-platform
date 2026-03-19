package repository

import (
	"context"

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
	u, err := r.client.User.Create().
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
