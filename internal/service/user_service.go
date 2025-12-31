package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/repository/mongodb"
	"github.com/yourorg/anonymous-support/internal/repository/postgres"
)

type UserService struct {
	userRepo      *postgres.UserRepository
	analyticsRepo *mongodb.AnalyticsRepository
}

func NewUserService(
	userRepo *postgres.UserRepository,
	analyticsRepo *mongodb.AnalyticsRepository,
) *UserService {
	return &UserService{
		userRepo:      userRepo,
		analyticsRepo: analyticsRepo,
	}
}

func (s *UserService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	return s.userRepo.GetByID(ctx, uid)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID string, username *string, avatarID *int) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return s.userRepo.UpdateProfile(ctx, uid, username, avatarID)
}

func (s *UserService) GetStreak(ctx context.Context, userID string) (*domain.UserTracker, error) {
	return s.analyticsRepo.GetTracker(ctx, userID)
}

func (s *UserService) UpdateStreak(ctx context.Context, userID string, hadRelapse bool) (int, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return 0, err
	}
	if err := s.analyticsRepo.UpdateStreak(ctx, uid, hadRelapse); err != nil {
		return 0, err
	}
	tracker, err := s.analyticsRepo.GetTracker(ctx, userID)
	if err != nil {
		return 0, err
	}
	return tracker.StreakDays, nil
}
