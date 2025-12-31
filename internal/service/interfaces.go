package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/dto"
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
	GetProfile(ctx context.Context, userID uuid.UUID) (*dto.ProfileDTO, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req *dto.UpdateProfileRequest) (*dto.UserDTO, error)
	GetStreak(ctx context.Context, userID uuid.UUID) (*dto.StreakDTO, error)
	UpdateStreak(ctx context.Context, userID uuid.UUID, hasRelapsed bool) (*dto.StreakDTO, error)
}

// PostServiceInterface defines the post service interface
type PostServiceInterface interface {
	CreatePost(ctx context.Context, userID uuid.UUID, username string, req *dto.CreatePostRequest) (*dto.PostDTO, error)
	GetPost(ctx context.Context, postID string, viewerID *uuid.UUID) (*dto.PostDTO, error)
	GetFeed(ctx context.Context, req *dto.GetFeedRequest) (*dto.FeedResponse, error)
	DeletePost(ctx context.Context, postID string, userID uuid.UUID) error
	UpdatePostUrgency(ctx context.Context, postID string, userID uuid.UUID, urgencyLevel int32) (*dto.PostDTO, error)
}

// SupportServiceInterface defines the support service interface
type SupportServiceInterface interface {
	CreateResponse(ctx context.Context, userID uuid.UUID, username string, req *dto.CreateResponseRequest) (*dto.SupportResponseDTO, error)
	GetResponses(ctx context.Context, req *dto.GetResponsesRequest) (*dto.ResponsesListResponse, error)
	QuickSupport(ctx context.Context, userID uuid.UUID, postID string) error
	GetSupportStats(ctx context.Context, userID uuid.UUID) (*dto.SupportStatsDTO, error)
}

// CircleServiceInterface defines the circle service interface
type CircleServiceInterface interface {
	CreateCircle(ctx context.Context, userID uuid.UUID, req *dto.CreateCircleRequest) (*dto.CircleDTO, error)
	JoinCircle(ctx context.Context, userID uuid.UUID, circleID uuid.UUID) error
	LeaveCircle(ctx context.Context, userID uuid.UUID, circleID uuid.UUID) error
	GetCircleMembers(ctx context.Context, req *dto.GetCircleMembersRequest) (*dto.CircleMembersResponse, error)
	GetCircleFeed(ctx context.Context, circleID uuid.UUID, limit, offset int32) (*dto.FeedResponse, error)
	GetCircles(ctx context.Context, req *dto.GetCirclesRequest) (*dto.CircleListResponse, error)
}

// ModerationServiceInterface defines the moderation service interface
type ModerationServiceInterface interface {
	ReportContent(ctx context.Context, userID uuid.UUID, req *dto.ReportContentRequest) error
	GetReports(ctx context.Context, req *dto.GetReportsRequest) (*dto.ReportsListResponse, error)
	ModerateContent(ctx context.Context, moderatorID uuid.UUID, req *dto.ModerateContentRequest) error
}

// AnalyticsServiceInterface defines the analytics service interface
type AnalyticsServiceInterface interface {
	TrackUserActivity(ctx context.Context, userID uuid.UUID, activityType string, metadata map[string]interface{}) error
	GetUserMetrics(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
}
