package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	AuditEventLogin           AuditEventType = "auth.login"
	AuditEventLogout          AuditEventType = "auth.logout"
	AuditEventRefreshToken    AuditEventType = "auth.refresh_token"
	AuditEventLoginFailed     AuditEventType = "auth.login_failed"
	AuditEventTokenRevoked    AuditEventType = "auth.token_revoked"
	AuditEventPasswordChanged AuditEventType = "auth.password_changed"

	AuditEventUserCreated   AuditEventType = "user.created"
	AuditEventUserUpdated   AuditEventType = "user.updated"
	AuditEventUserBanned    AuditEventType = "user.banned"
	AuditEventUserUnbanned  AuditEventType = "user.unbanned"
	AuditEventUserDeleted   AuditEventType = "user.deleted"

	AuditEventPostCreated  AuditEventType = "post.created"
	AuditEventPostUpdated  AuditEventType = "post.updated"
	AuditEventPostDeleted  AuditEventType = "post.deleted"
	AuditEventPostModerated AuditEventType = "post.moderated"

	AuditEventReportCreated   AuditEventType = "moderation.report_created"
	AuditEventReportReviewed  AuditEventType = "moderation.report_reviewed"
	AuditEventContentRemoved  AuditEventType = "moderation.content_removed"
	AuditEventUserWarned      AuditEventType = "moderation.user_warned"

	AuditEventCircleCreated AuditEventType = "circle.created"
	AuditEventCircleJoined  AuditEventType = "circle.joined"
	AuditEventCircleLeft    AuditEventType = "circle.left"
	AuditEventCircleDeleted AuditEventType = "circle.deleted"

	AuditEventPermissionGranted AuditEventType = "admin.permission_granted"
	AuditEventPermissionRevoked AuditEventType = "admin.permission_revoked"
	AuditEventRoleChanged       AuditEventType = "admin.role_changed"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID          uuid.UUID      `db:"id"`
	EventType   AuditEventType `db:"event_type"`
	ActorID     *uuid.UUID     `db:"actor_id"`     // User who performed the action (nil for system events)
	ActorIP     string         `db:"actor_ip"`
	TargetID    *uuid.UUID     `db:"target_id"`    // ID of the affected resource
	TargetType  string         `db:"target_type"`  // Type of affected resource (user, post, circle, etc.)
	Action      string         `db:"action"`       // Human-readable action description
	Metadata    string         `db:"metadata"`     // JSON metadata with additional context
	Success     bool           `db:"success"`      // Whether the action succeeded
	ErrorMessage *string       `db:"error_message"` // Error message if action failed
	CreatedAt   time.Time      `db:"created_at"`
}

// AuditLogMetadata contains additional context for audit events
type AuditLogMetadata struct {
	UserAgent     string            `json:"user_agent,omitempty"`
	RequestID     string            `json:"request_id,omitempty"`
	SessionID     string            `json:"session_id,omitempty"`
	ResourceBefore map[string]interface{} `json:"resource_before,omitempty"`
	ResourceAfter  map[string]interface{} `json:"resource_after,omitempty"`
	Changes       map[string]interface{} `json:"changes,omitempty"`
	Reason        string            `json:"reason,omitempty"`
	Extra         map[string]interface{} `json:"extra,omitempty"`
}
