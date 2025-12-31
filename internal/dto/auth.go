package dto

import (
	"fmt"

	"github.com/yourorg/anonymous-support/internal/domain"
	apperrors "github.com/yourorg/anonymous-support/internal/errors"
	"github.com/yourorg/anonymous-support/internal/pkg/validator"
)

// RegisterAnonymousRequest represents a request to register an anonymous user
type RegisterAnonymousRequest struct {
	Username string
}

// Validate validates the request
func (r *RegisterAnonymousRequest) Validate() error {
	if err := validator.ValidateUsername(r.Username); err != nil {
		return apperrors.NewValidationError("Invalid username", err)
	}
	return nil
}

// RegisterWithEmailRequest represents a request to register with email
type RegisterWithEmailRequest struct {
	Username string
	Email    string
	Password string
}

// Validate validates the request
func (r *RegisterWithEmailRequest) Validate() error {
	if err := validator.ValidateUsername(r.Username); err != nil {
		return apperrors.NewValidationError("Invalid username", err)
	}
	if err := validator.ValidateEmail(r.Email); err != nil {
		return apperrors.NewValidationError("Invalid email", err)
	}
	if err := validator.ValidatePassword(r.Password); err != nil {
		return apperrors.NewValidationError("Invalid password", err)
	}
	return nil
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string
	Password string
}

// Validate validates the request
func (r *LoginRequest) Validate() error {
	if err := validator.ValidateEmail(r.Email); err != nil {
		return apperrors.NewValidationError("Invalid email", err)
	}
	if r.Password == "" {
		return apperrors.NewValidationError("Password is required", nil)
	}
	return nil
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string
}

// Validate validates the request
func (r *RefreshTokenRequest) Validate() error {
	if r.RefreshToken == "" {
		return apperrors.NewValidationError("Refresh token is required", nil)
	}
	return nil
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	AccessToken  string
	RefreshToken string
	User         *UserDTO
	ExpiresIn    int64 // Token expiration in seconds
}

// UserDTO represents user data for responses
type UserDTO struct {
	ID             string
	Username       string
	Email          string
	AvatarID       string
	Role           domain.Role
	IsAnonymous    bool
	IsPremium      bool
	StrengthPoints int
	CreatedAt      string
	LastActiveAt   string
}

// NewUserDTO creates a UserDTO from a domain.User
func NewUserDTO(user *domain.User) *UserDTO {
	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	return &UserDTO{
		ID:             user.ID.String(),
		Username:       user.Username,
		Email:          email,
		AvatarID:       fmt.Sprintf("%d", user.AvatarID),
		Role:           user.Role,
		IsAnonymous:    user.IsAnonymous,
		IsPremium:      user.IsPremium,
		StrengthPoints: user.StrengthPoints,
		CreatedAt:      user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		LastActiveAt:   user.LastActiveAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
