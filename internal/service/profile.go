package service

import (
	"context"

	"github.com/luckysxx/user-platform/internal/repository"
	servicecontract "github.com/luckysxx/user-platform/internal/service/contract"
	"go.uber.org/zap"
)

type ProfileService interface {
	GetProfile(ctx context.Context, query *servicecontract.GetProfileQuery) (*servicecontract.GetProfileResult, error)
	UpdateProfile(ctx context.Context, cmd *servicecontract.UpdateProfileCommand) (*servicecontract.UpdateProfileResult, error)
}

type profileService struct {
	profileRepo repository.ProfileRepository
	logger      *zap.Logger
}

func NewProfileService(profileRepo repository.ProfileRepository, logger *zap.Logger) ProfileService {
	return &profileService{
		profileRepo: profileRepo,
		logger:      logger,
	}
}

func (s *profileService) GetProfile(ctx context.Context, query *servicecontract.GetProfileQuery) (*servicecontract.GetProfileResult, error) {
	p, err := s.profileRepo.GetByUserID(ctx, query.UserID)
	if err != nil {
		return nil, err
	}

	return &servicecontract.GetProfileResult{
		UserID:    query.UserID,
		Nickname:  p.Nickname,
		AvatarURL: p.AvatarURL,
		Bio:       p.Bio,
		UpdatedAt: p.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *profileService) UpdateProfile(ctx context.Context, cmd *servicecontract.UpdateProfileCommand) (*servicecontract.UpdateProfileResult, error) {
	updated, err := s.profileRepo.Update(ctx, cmd.UserID, cmd.Nickname, cmd.AvatarURL, cmd.Bio)
	if err != nil {
		return nil, err
	}

	return &servicecontract.UpdateProfileResult{
		UserID:    cmd.UserID,
		Nickname:  updated.Nickname,
		AvatarURL: updated.AvatarURL,
		Bio:       updated.Bio,
		UpdatedAt: updated.UpdatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
