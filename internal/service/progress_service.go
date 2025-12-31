package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/repository"
)

type ProgressService struct {
	analyticsRepo repository.AnalyticsRepository
	postRepo      repository.PostRepository
}

func NewProgressService(analyticsRepo repository.AnalyticsRepository, postRepo repository.PostRepository) *ProgressService {
	return &ProgressService{
		analyticsRepo: analyticsRepo,
		postRepo:      postRepo,
	}
}

// ProgressDashboard represents a user's progress dashboard
type ProgressDashboard struct {
	UserID           string                   `json:"user_id"`
	CurrentStreak    int                      `json:"current_streak"`
	LongestStreak    int                      `json:"longest_streak"`
	TotalDaysClean   int                      `json:"total_days_clean"`
	Milestones       []string                 `json:"milestones"`
	CravingsResisted int                      `json:"cravings_resisted"`
	TotalCravings    int                      `json:"total_cravings"`
	SupportGiven     int                      `json:"support_given"`
	SupportReceived  int                      `json:"support_received"`
	RelapsePattern   *RelapsePattern          `json:"relapse_pattern"`
	WeeklyProgress   []DayProgress            `json:"weekly_progress"`
	Achievements     []Achievement            `json:"achievements"`
}

// RelapsePattern analyzes user's relapse patterns
type RelapsePattern struct {
	TotalRelapses      int                  `json:"total_relapses"`
	AverageTimeClean   float64              `json:"average_time_clean"` // in days
	HighRiskTimeOfDay  string               `json:"high_risk_time_of_day"`
	HighRiskDayOfWeek  string               `json:"high_risk_day_of_week"`
	CommonTriggers     []string             `json:"common_triggers"`
	RecentRelapses     []RelapseEvent       `json:"recent_relapses"`
}

// RelapseEvent represents a single relapse
type RelapseEvent struct {
	Date        time.Time `json:"date"`
	DaysClean   int       `json:"days_clean"`
	Trigger     string    `json:"trigger,omitempty"`
	TimeOfDay   string    `json:"time_of_day"`
}

// DayProgress represents progress for a single day
type DayProgress struct {
	Date            time.Time `json:"date"`
	CheckedIn       bool      `json:"checked_in"`
	CravingsCount   int       `json:"cravings_count"`
	SupportGiven    int       `json:"support_given"`
	MoodScore       int       `json:"mood_score"` // 1-10
}

// Achievement represents a milestone achievement
type Achievement struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	UnlockedAt  time.Time `json:"unlocked_at"`
	Icon        string    `json:"icon"`
	Rarity      string    `json:"rarity"` // common, rare, epic, legendary
}

// GetDashboard retrieves comprehensive progress dashboard for a user
func (s *ProgressService) GetDashboard(ctx context.Context, userID string) (*ProgressDashboard, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	// Get user tracker
	tracker, err := s.analyticsRepo.GetUserTracker(ctx, uid)
	if err != nil {
		return nil, err
	}

	// Calculate milestones
	milestones := s.calculateMilestones(tracker)

	// Get relapse pattern
	relapsePattern := s.analyzeRelapsePattern(tracker)

	// Get weekly progress
	weeklyProgress := s.getWeeklyProgress(ctx, userID)

	// Calculate achievements
	achievements := s.calculateAchievements(tracker)

	dashboard := &ProgressDashboard{
		UserID:           userID,
		CurrentStreak:    tracker.StreakDays,
		LongestStreak:    tracker.LongestStreak,
		TotalDaysClean:   tracker.TotalDaysClean,
		Milestones:       milestones,
		CravingsResisted: tracker.CravingsResisted,
		TotalCravings:    tracker.TotalCravings,
		SupportGiven:     tracker.SupportGiven,
		SupportReceived:  tracker.SupportReceived,
		RelapsePattern:   relapsePattern,
		WeeklyProgress:   weeklyProgress,
		Achievements:     achievements,
	}

	return dashboard, nil
}

// calculateMilestones generates milestone badges based on tracker data
func (s *ProgressService) calculateMilestones(tracker *domain.UserTracker) []string {
	milestones := []string{}

	dayMilestones := []int{1, 7, 14, 30, 60, 90, 180, 365}
	for _, days := range dayMilestones {
		if tracker.StreakDays >= days {
			milestones = append(milestones, formatDayMilestone(days))
		}
	}

	if tracker.SupportGiven >= 10 {
		milestones = append(milestones, "Helpful Friend - 10 supports given")
	}
	if tracker.SupportGiven >= 50 {
		milestones = append(milestones, "Support Champion - 50 supports given")
	}

	if tracker.CravingsResisted >= 20 {
		milestones = append(milestones, "Craving Warrior - 20 cravings resisted")
	}

	return milestones
}

func formatDayMilestone(days int) string {
	switch days {
	case 1:
		return "First Day Clean"
	case 7:
		return "One Week Strong"
	case 14:
		return "Two Weeks Clean"
	case 30:
		return "One Month Milestone"
	case 60:
		return "Two Months Clean"
	case 90:
		return "Three Months Strong"
	case 180:
		return "Six Months Clean"
	case 365:
		return "One Year Anniversary"
	default:
		return "Milestone Achieved"
	}
}

// analyzeRelapsePattern analyzes relapse patterns
func (s *ProgressService) analyzeRelapsePattern(tracker *domain.UserTracker) *RelapsePattern {
	avgTimeClean := float64(0)
	if tracker.TotalRelapses > 0 {
		avgTimeClean = float64(tracker.TotalDaysClean) / float64(tracker.TotalRelapses+1)
	}

	return &RelapsePattern{
		TotalRelapses:      tracker.TotalRelapses,
		AverageTimeClean:   avgTimeClean,
		HighRiskTimeOfDay:  "evening", // TODO: Calculate from actual data
		HighRiskDayOfWeek:  "weekend", // TODO: Calculate from actual data
		CommonTriggers:     []string{}, // TODO: Extract from posts/check-ins
		RecentRelapses:     []RelapseEvent{}, // TODO: Load from relapse history
	}
}

// getWeeklyProgress fetches progress data for the last 7 days
func (s *ProgressService) getWeeklyProgress(ctx context.Context, userID string) []DayProgress {
	// TODO: Implement actual data fetching
	// For now, return mock data structure
	progress := []DayProgress{}
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		progress = append(progress, DayProgress{
			Date:          date,
			CheckedIn:     true,
			CravingsCount: 0,
			SupportGiven:  0,
			MoodScore:     7,
		})
	}
	return progress
}

// calculateAchievements generates achievement list
func (s *ProgressService) calculateAchievements(tracker *domain.UserTracker) []Achievement {
	achievements := []Achievement{}

	if tracker.StreakDays >= 7 {
		achievements = append(achievements, Achievement{
			ID:          "first_week",
			Title:       "First Week Strong",
			Description: "Maintained a 7-day streak",
			UnlockedAt:  time.Now(),
			Icon:        "🏆",
			Rarity:      "common",
		})
	}

	if tracker.StreakDays >= 30 {
		achievements = append(achievements, Achievement{
			ID:          "first_month",
			Title:       "One Month Milestone",
			Description: "Completed 30 days clean",
			UnlockedAt:  time.Now(),
			Icon:        "🎖️",
			Rarity:      "rare",
		})
	}

	if tracker.SupportGiven >= 50 {
		achievements = append(achievements, Achievement{
			ID:          "support_champion",
			Title:       "Support Champion",
			Description: "Helped 50 community members",
			UnlockedAt:  time.Now(),
			Icon:        "🤝",
			Rarity:      "epic",
		})
	}

	return achievements
}

// RecordCheckIn records a daily check-in
func (s *ProgressService) RecordCheckIn(ctx context.Context, userID string, hadRelapse bool, moodScore int) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	return s.analyticsRepo.UpdateStreak(ctx, uid, hadRelapse)
}

// RecordCraving records a craving event
func (s *ProgressService) RecordCraving(ctx context.Context, userID string, resisted bool) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	return s.analyticsRepo.IncrementCravings(ctx, uid, resisted)
}
