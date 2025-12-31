package dto

import (
	"github.com/anonymous-support/internal/domain"
	apperrors "github.com/anonymous-support/internal/errors"
	"github.com/anonymous-support/internal/pkg/validator"
)

// CreatePostRequest represents a request to create a post
type CreatePostRequest struct {
	Type            domain.PostType
	Content         string
	Categories      []string
	UrgencyLevel    int32
	DaysSinceRelapse int32
	TimeContext     string
	Tags            []string
	CircleID        string
}

// Validate validates the request
func (r *CreatePostRequest) Validate() error {
	if err := validator.ValidatePostContent(r.Content); err != nil {
		return apperrors.NewValidationError("Invalid post content", err)
	}

	// Validate post type
	switch r.Type {
	case domain.PostTypeSOS, domain.PostTypeCheckIn, domain.PostTypeVictory, domain.PostTypeQuestion:
		// valid types
	default:
		return apperrors.NewValidationError("Invalid post type", nil)
	}

	// Validate urgency level
	if r.UrgencyLevel < 1 || r.UrgencyLevel > 5 {
		return apperrors.NewValidationError("Urgency level must be between 1 and 5", nil)
	}

	// Validate categories
	if len(r.Categories) == 0 {
		return apperrors.NewValidationError("At least one category is required", nil)
	}

	return nil
}

// GetFeedRequest represents a request to get a feed
type GetFeedRequest struct {
	Category string
	CircleID string
	Limit    int32
	Offset   int32
}

// Validate validates the request
func (r *GetFeedRequest) Validate() error {
	if r.Limit <= 0 {
		r.Limit = 20 // default
	}
	if r.Limit > 100 {
		return apperrors.NewValidationError("Limit cannot exceed 100", nil)
	}
	if r.Offset < 0 {
		return apperrors.NewValidationError("Offset cannot be negative", nil)
	}
	return nil
}

// UpdatePostUrgencyRequest represents a request to update post urgency
type UpdatePostUrgencyRequest struct {
	PostID       string
	UrgencyLevel int32
}

// Validate validates the request
func (r *UpdatePostUrgencyRequest) Validate() error {
	if r.PostID == "" {
		return apperrors.NewValidationError("Post ID is required", nil)
	}
	if r.UrgencyLevel < 1 || r.UrgencyLevel > 5 {
		return apperrors.NewValidationError("Urgency level must be between 1 and 5", nil)
	}
	return nil
}

// PostDTO represents post data for responses
type PostDTO struct {
	ID               string
	UserID           string
	Username         string
	Type             string
	Content          string
	Categories       []string
	UrgencyLevel     int32
	DaysSinceRelapse int32
	TimeContext      string
	Tags             []string
	Visibility       string
	CircleID         string
	ResponseCount    int32
	SupportCount     int32
	CreatedAt        string
	ExpiresAt        string
	IsModerated      bool
}

// NewPostDTO creates a PostDTO from a domain.Post
func NewPostDTO(post *domain.Post) *PostDTO {
	return &PostDTO{
		ID:               post.ID.Hex(),
		UserID:           post.UserID.String(),
		Username:         post.Username,
		Type:             string(post.Type),
		Content:          post.Content,
		Categories:       post.Categories,
		UrgencyLevel:     post.UrgencyLevel,
		DaysSinceRelapse: post.Context.DaysSinceRelapse,
		TimeContext:      post.Context.TimeContext,
		Tags:             post.Context.Tags,
		Visibility:       string(post.Visibility),
		CircleID:         post.CircleID,
		ResponseCount:    post.ResponseCount,
		SupportCount:     post.SupportCount,
		CreatedAt:        post.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ExpiresAt:        post.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		IsModerated:      post.IsModerated,
	}
}

// FeedResponse represents a paginated feed response
type FeedResponse struct {
	Posts      []*PostDTO
	TotalCount int32
	HasMore    bool
}
