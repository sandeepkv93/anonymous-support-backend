package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"
)

type User struct {
	ID             uuid.UUID `db:"id" json:"id"`
	Username       string    `db:"username" json:"username"`
	Email          *string   `db:"email" json:"email,omitempty"`
	PasswordHash   string    `db:"password_hash" json:"-"`
	AvatarID       int       `db:"avatar_id" json:"avatar_id"`
	Role           Role      `db:"role" json:"role"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	LastActiveAt   time.Time `db:"last_active_at" json:"last_active_at"`
	IsAnonymous    bool      `db:"is_anonymous" json:"is_anonymous"`
	IsBanned       bool      `db:"is_banned" json:"is_banned"`
	IsPremium      bool      `db:"is_premium" json:"is_premium"`
	StrengthPoints int       `db:"strength_points" json:"strength_points"`
}

type UserClaims struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	IsAnonymous bool   `json:"is_anonymous"`
}
