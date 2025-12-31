package dto

import (
	"github.com/yourorg/anonymous-support/internal/domain"
	apperrors "github.com/yourorg/anonymous-support/internal/errors"
)

// CreateCircleRequest represents a request to create a circle
type CreateCircleRequest struct {
	Name        string
	Description string
	Category    string
	MaxMembers  int32
	IsPrivate   bool
}

// Validate validates the request
func (r *CreateCircleRequest) Validate() error {
	if len(r.Name) < 3 || len(r.Name) > 100 {
		return apperrors.NewValidationError("Circle name must be between 3 and 100 characters", nil)
	}
	if len(r.Description) > 500 {
		return apperrors.NewValidationError("Description cannot exceed 500 characters", nil)
	}
	if r.Category == "" {
		return apperrors.NewValidationError("Category is required", nil)
	}
	if r.MaxMembers < 2 || r.MaxMembers > 10000 {
		return apperrors.NewValidationError("Max members must be between 2 and 10000", nil)
	}
	return nil
}

// JoinCircleRequest represents a request to join a circle
type JoinCircleRequest struct {
	CircleID string
}

// Validate validates the request
func (r *JoinCircleRequest) Validate() error {
	if r.CircleID == "" {
		return apperrors.NewValidationError("Circle ID is required", nil)
	}
	return nil
}

// GetCircleMembersRequest represents a request to get circle members
type GetCircleMembersRequest struct {
	CircleID string
	Limit    int32
	Offset   int32
}

// Validate validates the request
func (r *GetCircleMembersRequest) Validate() error {
	if r.CircleID == "" {
		return apperrors.NewValidationError("Circle ID is required", nil)
	}
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

// GetCirclesRequest represents a request to get circles
type GetCirclesRequest struct {
	Category string
	Limit    int32
	Offset   int32
}

// Validate validates the request
func (r *GetCirclesRequest) Validate() error {
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

// CircleDTO represents circle data for responses
type CircleDTO struct {
	ID          string
	Name        string
	Description string
	Category    string
	MaxMembers  int32
	MemberCount int32
	IsPrivate   bool
	CreatedBy   string
	CreatedAt   string
	UpdatedAt   string
}

// NewCircleDTO creates a CircleDTO from a domain.Circle
func NewCircleDTO(circle *domain.Circle) *CircleDTO {
	return &CircleDTO{
		ID:          circle.ID.String(),
		Name:        circle.Name,
		Description: circle.Description,
		Category:    circle.Category,
		MaxMembers:  int32(circle.MaxMembers),
		MemberCount: int32(circle.MemberCount),
		IsPrivate:   circle.IsPrivate,
		CreatedBy:   circle.CreatedBy.String(),
		CreatedAt:   circle.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   circle.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// CircleListResponse represents a list of circles
type CircleListResponse struct {
	Circles    []*CircleDTO
	TotalCount int32
	HasMore    bool
}

// CircleMemberDTO represents a circle member
type CircleMemberDTO struct {
	UserID      string
	Username    string
	AvatarID    string
	JoinedAt    string
	IsAnonymous bool
}

// CircleMembersResponse represents a list of circle members
type CircleMembersResponse struct {
	Members    []*CircleMemberDTO
	TotalCount int32
	HasMore    bool
}
