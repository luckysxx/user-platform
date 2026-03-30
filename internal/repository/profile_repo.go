package repository

import (
	"context"
	"errors"

	"github.com/luckysxx/common/rpc"
	"github.com/luckysxx/user-platform/internal/dberr"
	"github.com/luckysxx/user-platform/internal/ent"
	"github.com/luckysxx/user-platform/internal/ent/profile"
	"github.com/luckysxx/user-platform/internal/ent/user"
)

// ProfileRepository 定义了资料实体的持久化接口。
type ProfileRepository interface {
	CreateEmpty(ctx context.Context, userID int64) (*ent.Profile, error)
	GetByUserID(ctx context.Context, userID int64) (*ent.Profile, error)
	EnsureByUserID(ctx context.Context, userID int64) (*ent.Profile, error)
	Update(ctx context.Context, userID int64, nickname, avatarURL, bio, birthday string) (*ent.Profile, error)
}

// profileRepository 是 ProfileRepository 的 Ent 实现。
type profileRepository struct {
	client *ent.Client
}

// NewProfileRepository 创建一个资料仓储实例。
func NewProfileRepository(client *ent.Client) ProfileRepository {
	return &profileRepository{client: client}
}

// CreateEmpty 为新用户创建一条空资料记录。
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

// GetByUserID 按用户 ID 查询对应的资料记录。
func (r *profileRepository) GetByUserID(ctx context.Context, userID int64) (*ent.Profile, error) {
	p, err := r.client.Profile.Query().
		Where(profile.HasUserWith(user.IDEQ(userID))).
		Only(ctx)
	if err != nil {
		return nil, dberr.ParseDBError(err)
	}
	return p, nil
}

// EnsureByUserID 确保指定用户一定拥有一条资料记录。
func (r *profileRepository) EnsureByUserID(ctx context.Context, userID int64) (*ent.Profile, error) {
	p, err := r.GetByUserID(ctx, userID)
	if err == nil {
		return p, nil
	}
	if !errors.Is(err, dberr.ErrNoRows) {
		return nil, err
	}

	p, err = r.CreateEmpty(ctx, userID)
	if err == nil {
		return p, nil
	}

	// 如果并发请求已经补齐了资料，再查一次即可。
	if dberr.IsDuplicateKeyError(err) {
		return r.GetByUserID(ctx, userID)
	}
	return nil, err
}

// Update 按用户 ID 更新一条资料记录。
func (r *profileRepository) Update(ctx context.Context, userID int64, nickname, avatarURL, bio, birthday string) (*ent.Profile, error) {
	// 先根据 userID 查出 profile 的 ID，因为 ent 的 update 会需要通过 edge 查询或直接通过自身实体 ID
	p, err := r.EnsureByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	updated, err := p.Update().
		SetNickname(nickname).
		SetAvatarURL(avatarURL).
		SetBio(bio).
		SetBirthday(birthday).
		Save(ctx)

	if err != nil {
		return nil, dberr.ParseDBError(err)
	}
	return updated, nil
}
