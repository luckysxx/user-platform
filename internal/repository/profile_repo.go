package repository

import (
	"context"

	"github.com/luckysxx/common/rpc"
	"github.com/luckysxx/user-platform/internal/dberr"
	"github.com/luckysxx/user-platform/internal/ent"
	"github.com/luckysxx/user-platform/internal/ent/profile"
	"github.com/luckysxx/user-platform/internal/ent/user"
)

type ProfileRepository interface {
	CreateEmpty(ctx context.Context, userID int64) (*ent.Profile, error)
	GetByUserID(ctx context.Context, userID int64) (*ent.Profile, error)
	Update(ctx context.Context, userID int64, nickname, avatarUrl, bio string) (*ent.Profile, error)
}

type profileRepository struct {
	client *ent.Client
}

func NewProfileRepository(client *ent.Client) ProfileRepository {
	return &profileRepository{client: client}
}

func (r *profileRepository) CreateEmpty(ctx context.Context, userID int64) (*ent.Profile, error) {
	newID, err := rpc.GenerateID(ctx)
	if err != nil {
		return nil, err
	}

	tx := ent.TxFromContext(ctx)
	if tx != nil {
		p, err := tx.Profile.Create().
			SetID(newID).
			SetUserID(userID).
			Save(ctx)
		if err != nil {
			return nil, dberr.ParseDBError(err)
		}
		return p, err
	}

	p, err := r.client.Profile.Create().
		SetID(newID).
		SetUserID(userID).
		Save(ctx)
	if err != nil {
		return nil, dberr.ParseDBError(err)
	}
	return p, nil
}

func (r *profileRepository) GetByUserID(ctx context.Context, userID int64) (*ent.Profile, error) {
	p, err := r.client.Profile.Query().
		Where(profile.HasUserWith(user.IDEQ(userID))).
		Only(ctx)
	if err != nil {
		return nil, dberr.ParseDBError(err)
	}
	return p, nil
}

func (r *profileRepository) Update(ctx context.Context, userID int64, nickname, avatarUrl, bio string) (*ent.Profile, error) {
	// 先根据 userID 查出 profile 的 ID，因为 ent 的 update 会需要通过 edge 查询或直接通过自身实体 ID
	p, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	updated, err := p.Update().
		SetNickname(nickname).
		SetAvatarURL(avatarUrl).
		SetBio(bio).
		Save(ctx)
	
	if err != nil {
		return nil, dberr.ParseDBError(err)
	}
	return updated, nil
}
