package service

import (
	"context"

	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/repository/mongodb"
)

type AnalyticsService struct {
	analyticsRepo *mongodb.AnalyticsRepository
}

func NewAnalyticsService(analyticsRepo *mongodb.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{analyticsRepo: analyticsRepo}
}

func (s *AnalyticsService) GetTracker(ctx context.Context, userID string) (*domain.UserTracker, error) {
	return s.analyticsRepo.GetTracker(ctx, userID)
}

func (s *AnalyticsService) UpdateStreak(ctx context.Context, userID string, hadRelapse bool) (int, error) {
	return s.analyticsRepo.UpdateStreak(ctx, userID, hadRelapse)
}

func (s *AnalyticsService) RecordCraving(ctx context.Context, userID string, resisted bool) error {
	return s.analyticsRepo.IncrementCravings(ctx, userID, resisted)
}
