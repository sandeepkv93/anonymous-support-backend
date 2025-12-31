package domain

import (
	"time"

	"github.com/google/uuid"
)

type Circle struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Category    string    `db:"category" json:"category"`
	MaxMembers  int       `db:"max_members" json:"max_members"`
	MemberCount int       `db:"member_count" json:"member_count"`
	IsPrivate   bool      `db:"is_private" json:"is_private"`
	CreatedBy   uuid.UUID `db:"created_by" json:"created_by"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type CircleMembership struct {
	ID       uuid.UUID `db:"id" json:"id"`
	CircleID uuid.UUID `db:"circle_id" json:"circle_id"`
	UserID   uuid.UUID `db:"user_id" json:"user_id"`
	JoinedAt time.Time `db:"joined_at" json:"joined_at"`
	Role     string    `db:"role" json:"role"`
}

// Invite represents a circle invitation
type Invite struct {
	ID        uuid.UUID `db:"id" json:"id"`
	CircleID  uuid.UUID `db:"circle_id" json:"circle_id"`
	Code      string    `db:"code" json:"code"`
	CreatedBy uuid.UUID `db:"created_by" json:"created_by"`
	MaxUses   int       `db:"max_uses" json:"max_uses"`       // 0 = unlimited
	UsedCount int       `db:"used_count" json:"used_count"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	IsActive  bool      `db:"is_active" json:"is_active"`
}
