package dto

import (
	apperrors "github.com/yourorg/anonymous-support/internal/errors"
	"github.com/yourorg/anonymous-support/internal/pkg/validator"
)

// UpdateProfileRequest represents a request to update user profile
type UpdateProfileRequest struct {
	Username string
	AvatarID string
}

// Validate validates the request
func (r *UpdateProfileRequest) Validate() error {
	if r.Username != "" {
		if err := validator.ValidateUsername(r.Username); err != nil {
			return apperrors.NewValidationError("Invalid username", err)
		}
	}
	return nil
}

// UpdateStreakRequest represents a request to update recovery streak
type UpdateStreakRequest struct {
	HasRelapsed bool
}

// Validate validates the request
func (r *UpdateStreakRequest) Validate() error {
	// No validation needed for boolean
	return nil
}

// StreakDTO represents recovery streak data
type StreakDTO struct {
	StreakDays           int32
	LastRelapseDate      string
	TotalCravings        int32
	CravingsResisted     int32
	VulnerabilityPattern string
	Goals                []string
	Milestones           []string
}

// ProfileDTO represents complete user profile data
type ProfileDTO struct {
	User   *UserDTO
	Streak *StreakDTO
}
