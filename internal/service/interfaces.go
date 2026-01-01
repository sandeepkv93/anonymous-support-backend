package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/dto"
	"github.com/yourorg/anonymous-support/internal/pkg/feed"
)

// AuthServiceInterface defines the authentication service interface
type AuthServiceInterface interface {
	RegisterAnonymous(ctx context.Context, username string) (*dto.AuthResponse, error)
	RegisterWithEmail(ctx context.Context, req *dto.RegisterWithEmailRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)
	Logout(ctx context.Context, userID uuid.UUID) error
}

// UserServiceInterface defines the user service interface
type UserServiceInterface interface {
	GetProfile(ctx context.Context, userID string) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID string, username *string, avatarID *int) error
	GetStreak(ctx context.Context, userID string) (*domain.UserTracker, error)
	UpdateStreak(ctx context.Context, userID string, hadRelapse bool) (int, error)
}

// PostServiceInterface defines the post service interface
type PostServiceInterface interface {
	CreatePost(ctx context.Context, userID, username string, postType domain.PostType, content string, categories []string, urgencyLevel int, timeContext string, daysSinceRelapse int, tags []string, visibility string, circleID *string) (*domain.Post, error)
	GetPost(ctx context.Context, postID string) (*domain.Post, error)
	GetFeed(ctx context.Context, categories []string, circleID *string, postType *domain.PostType, limit, offset int) ([]*domain.Post, error)
	DeletePost(ctx context.Context, postID, userID string) error
	UpdatePostUrgency(ctx context.Context, postID string, urgencyLevel int) error
	GetPersonalizedFeed(ctx context.Context, userPrefs *feed.UserPreferences, limit, offset int) ([]*domain.Post, error)
}

// SupportServiceInterface defines the support service interface
type SupportServiceInterface interface {
	CreateResponse(ctx context.Context, userID, username, postID string, responseType domain.ResponseType, content string, voiceNoteURL *string) (string, int, error)
	GetResponses(ctx context.Context, postID string, limit, offset int) ([]*domain.SupportResponse, error)
	QuickSupport(ctx context.Context, userID, postID, messageType string) (int, error)
	GetSupportStats(ctx context.Context, userID string) (given, received int64, strengthPoints, peopleHelped int, error error)
}

// CircleServiceInterface defines the circle service interface
type CircleServiceInterface interface {
	CreateCircle(ctx context.Context, userID, name, description, category string, maxMembers int, isPrivate bool) (string, error)
	JoinCircle(ctx context.Context, userID, circleID string) error
	LeaveCircle(ctx context.Context, userID, circleID string) error
	GetCircleMembers(ctx context.Context, circleID string, limit, offset int) ([]*domain.CircleMembership, error)
	GetCircleFeed(ctx context.Context, circleID string, limit, offset int) ([]*domain.Post, error)
	GetCircles(ctx context.Context, category *string, limit, offset int) ([]*domain.Circle, error)
}

// ModerationServiceInterface defines the moderation service interface
type ModerationServiceInterface interface {
	ReportContent(ctx context.Context, reporterID, contentType, contentID, reason, description string) (string, error)
	GetReports(ctx context.Context, status *string, limit, offset int) ([]*domain.ContentReport, error)
	ModerateContent(ctx context.Context, reportID, reviewerID, action string) error
}

// AnalyticsServiceInterface defines the analytics service interface
type AnalyticsServiceInterface interface {
	GetTracker(ctx context.Context, userID string) (*domain.UserTracker, error)
	UpdateStreak(ctx context.Context, userID string, hadRelapse bool) (int, error)
	RecordCraving(ctx context.Context, userID string, resisted bool) error
}
