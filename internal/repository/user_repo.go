package repository

import (
	"context"

	"github.com/luckysxx/user-platform/internal/dberr"
	"github.com/luckysxx/user-platform/internal/ent"
	"github.com/luckysxx/user-platform/internal/ent/app"
	"github.com/luckysxx/user-platform/internal/ent/user"
	"github.com/luckysxx/user-platform/internal/ent/userappprofile"
)

type UserRepository interface {
	Create(ctx context.Context, email string, username string, passwordhash string) (*ent.User, error)
	EnsureAppAuthorized(ctx context.Context, userID int64, appCode string) error
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

func (r *userRepository) EnsureAppAuthorized(ctx context.Context, userID int64, appCode string) error {
	appNode, err := r.client.App.Query().Where(app.AppCodeEQ(appCode)).Only(ctx)
	if err != nil {
		return dberr.ParseDBError(err)
	}

	_, err = r.client.UserAppProfile.Query().
		Where(userappprofile.HasUserWith(user.IDEQ(userID))).
		Where(userappprofile.HasAppWith(app.IDEQ(appNode.ID))).
		Only(ctx)
	if err == nil {
		return nil
	}
	if !ent.IsNotFound(err) {
		return dberr.ParseDBError(err)
	}

	_, err = r.client.UserAppProfile.Create().
		SetUserID(userID).
		SetAppID(appNode.ID).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil
		}
		return dberr.ParseDBError(err)
	}
	return nil
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
