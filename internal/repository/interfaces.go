package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/anonymous-support/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRepository defines the interface for user data persistence
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateLastActive(ctx context.Context, userID uuid.UUID) error
	UpdateStrengthPoints(ctx context.Context, userID uuid.UUID, points int) error
	UpdateProfile(ctx context.Context, userID uuid.UUID, username *string, avatarID *int) error
	UsernameExists(ctx context.Context, username string) (bool, error)
}

// PostRepository defines the interface for post data persistence
type PostRepository interface {
	Create(ctx context.Context, post *domain.Post) error
	GetByID(ctx context.Context, id string) (*domain.Post, error)
	GetFeed(ctx context.Context, categories []string, circleID *string, postType *domain.PostType, limit, offset int) ([]*domain.Post, error)
	Delete(ctx context.Context, id string) error
	UpdateUrgency(ctx context.Context, id string, urgencyLevel int32) error
	IncrementResponseCount(ctx context.Context, id string) error
	IncrementSupportCount(ctx context.Context, id string) error
}

// SupportRepository defines the interface for support response persistence
type SupportRepository interface {
	Create(ctx context.Context, response *domain.SupportResponse) error
	GetByPostID(ctx context.Context, postID primitive.ObjectID, limit, offset int) ([]*domain.SupportResponse, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.SupportResponse, error)
	CountByPostID(ctx context.Context, postID primitive.ObjectID) (int64, error)
}

// CircleRepository defines the interface for circle data persistence
type CircleRepository interface {
	Create(ctx context.Context, circle *domain.Circle) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Circle, error)
	List(ctx context.Context, category *string, limit, offset int) ([]*domain.Circle, error)
	JoinCircle(ctx context.Context, circleID, userID uuid.UUID) error
	LeaveCircle(ctx context.Context, circleID, userID uuid.UUID) error
	GetMembers(ctx context.Context, circleID uuid.UUID, limit, offset int) ([]uuid.UUID, error)
	IsMember(ctx context.Context, circleID, userID uuid.UUID) (bool, error)
	GetMemberCount(ctx context.Context, circleID uuid.UUID) (int, error)
}

// ModerationRepository defines the interface for moderation data persistence
type ModerationRepository interface {
	CreateReport(ctx context.Context, report *domain.Report) error
	GetReportByID(ctx context.Context, id uuid.UUID) (*domain.Report, error)
	ListReports(ctx context.Context, status *string, limit, offset int) ([]*domain.Report, error)
	UpdateReportStatus(ctx context.Context, id uuid.UUID, status string, reviewedBy uuid.UUID, notes string) error
	CreateBlock(ctx context.Context, blockerID, blockedID uuid.UUID) error
	RemoveBlock(ctx context.Context, blockerID, blockedID uuid.UUID) error
	IsBlocked(ctx context.Context, blockerID, blockedID uuid.UUID) (bool, error)
}

// SessionRepository defines the interface for session management
type SessionRepository interface {
	StoreRefreshToken(ctx context.Context, userID, token string, expiry time.Duration) error
	GetRefreshToken(ctx context.Context, userID string) (string, error)
	DeleteRefreshToken(ctx context.Context, userID string) error
	SetUserOnline(ctx context.Context, userID string, ttl time.Duration) error
	IsUserOnline(ctx context.Context, userID string) (bool, error)
}

// RealtimeRepository defines the interface for real-time data management
type RealtimeRepository interface {
	IncrementViewCount(ctx context.Context, postID string) error
	GetViewCount(ctx context.Context, postID string) (int64, error)
	AddSupporter(ctx context.Context, postID, userID string) error
	GetSupporters(ctx context.Context, postID string) ([]string, error)
	AddToFeed(ctx context.Context, userID, postID string, score float64) error
	GetFeed(ctx context.Context, userID string, limit int) ([]string, error)
	PublishNotification(ctx context.Context, channel string, message interface{}) error
	SubscribeToChannel(ctx context.Context, channel string) error
}

// CacheRepository defines the interface for caching
type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// AnalyticsRepository defines the interface for analytics and user tracking
type AnalyticsRepository interface {
	CreateUserTracker(ctx context.Context, userID uuid.UUID) error
	GetUserTracker(ctx context.Context, userID uuid.UUID) (*domain.UserTracker, error)
	UpdateStreak(ctx context.Context, userID uuid.UUID, hasRelapsed bool) error
	IncrementCravings(ctx context.Context, userID uuid.UUID, resisted bool) error
	AddMilestone(ctx context.Context, userID uuid.UUID, milestone string) error
}
