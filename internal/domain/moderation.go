package domain

import (
	"time"

	"github.com/google/uuid"
)

type ContentReport struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	ReporterID  uuid.UUID  `db:"reporter_id" json:"reporter_id"`
	ContentType string     `db:"content_type" json:"content_type"`
	ContentID   string     `db:"content_id" json:"content_id"`
	Reason      string     `db:"reason" json:"reason"`
	Description string     `db:"description" json:"description"`
	Status      string     `db:"status" json:"status"`
	ReviewedBy  *uuid.UUID `db:"reviewed_by" json:"reviewed_by,omitempty"`
	ReviewedAt  *time.Time `db:"reviewed_at" json:"reviewed_at,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}

type UserBlock struct {
	ID        uuid.UUID `db:"id" json:"id"`
	BlockerID uuid.UUID `db:"blocker_id" json:"blocker_id"`
	BlockedID uuid.UUID `db:"blocked_id" json:"blocked_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
