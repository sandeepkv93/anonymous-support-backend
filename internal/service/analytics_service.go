package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/repository"
)

type AnalyticsService struct {
	analyticsRepo repository.AnalyticsRepository
}

func NewAnalyticsService(analyticsRepo repository.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{analyticsRepo: analyticsRepo}
}

func (s *AnalyticsService) GetTracker(ctx context.Context, userID string) (*domain.UserTracker, error) {
	return s.analyticsRepo.GetTracker(ctx, userID)
}

func (s *AnalyticsService) UpdateStreak(ctx context.Context, userID string, hadRelapse bool) (int, error) {
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

func (s *AnalyticsService) RecordCraving(ctx context.Context, userID string, resisted bool) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return s.analyticsRepo.IncrementCravings(ctx, uid, resisted)
}
