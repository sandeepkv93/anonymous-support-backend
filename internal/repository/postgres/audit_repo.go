package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yourorg/anonymous-support/internal/domain"
)

type AuditRepository struct {
	db *sqlx.DB
}

func NewAuditRepository(db *sqlx.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// CreateAuditLog implements repository.AuditRepository interface
func (r *AuditRepository) CreateAuditLog(ctx context.Context, log *domain.AuditLog) error {
	return r.Log(ctx, log)
}

// GetAuditLogs implements repository.AuditRepository interface
func (r *AuditRepository) GetAuditLogs(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*domain.AuditLog, error) {
	query := `SELECT * FROM audit_logs ORDER BY timestamp DESC LIMIT $1 OFFSET $2`
	var logs []*domain.AuditLog
	err := r.db.SelectContext(ctx, &logs, query, limit, offset)
	return logs, err
}

// Log creates a new audit log entry
func (r *AuditRepository) Log(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			id, event_type, actor_id, actor_ip, target_id, target_type,
			action, metadata, success, error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	log.ID = uuid.New()

	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.EventType,
		log.ActorID,
		log.ActorIP,
		log.TargetID,
		log.TargetType,
		log.Action,
		log.Metadata,
		log.Success,
		log.ErrorMessage,
		log.CreatedAt,
	)

	return err
}

// GetByID retrieves an audit log by ID
func (r *AuditRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.AuditLog, error) {
	var log domain.AuditLog
	query := `SELECT * FROM audit_logs WHERE id = $1`
	err := r.db.GetContext(ctx, &log, query, id)
	return &log, err
}

// GetByActor retrieves audit logs for a specific actor
func (r *AuditRepository) GetByActor(ctx context.Context, actorID uuid.UUID, limit, offset int) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	query := `
		SELECT * FROM audit_logs
		WHERE actor_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &logs, query, actorID, limit, offset)
	return logs, err
}

// GetByEventType retrieves audit logs by event type
func (r *AuditRepository) GetByEventType(ctx context.Context, eventType domain.AuditEventType, limit, offset int) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	query := `
		SELECT * FROM audit_logs
		WHERE event_type = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &logs, query, eventType, limit, offset)
	return logs, err
}

// GetFailedEvents retrieves failed audit events
func (r *AuditRepository) GetFailedEvents(ctx context.Context, limit, offset int) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	query := `
		SELECT * FROM audit_logs
		WHERE success = false
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	err := r.db.SelectContext(ctx, &logs, query, limit, offset)
	return logs, err
}

// GetByTarget retrieves audit logs for a specific target resource
func (r *AuditRepository) GetByTarget(ctx context.Context, targetID uuid.UUID, limit, offset int) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	query := `
		SELECT * FROM audit_logs
		WHERE target_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &logs, query, targetID, limit, offset)
	return logs, err
}

// Helper function to create audit log with metadata
func (r *AuditRepository) LogWithMetadata(ctx context.Context, eventType domain.AuditEventType, actorID *uuid.UUID, actorIP string, targetID *uuid.UUID, targetType string, action string, metadata *domain.AuditLogMetadata, success bool, errorMessage *string) error {
	var metadataJSON []byte
	var err error

	if metadata != nil {
		metadataJSON, err = json.Marshal(metadata)
		if err != nil {
			return err
		}
	}

	log := &domain.AuditLog{
		EventType:    eventType,
		ActorID:      actorID,
		ActorIP:      actorIP,
		TargetID:     targetID,
		TargetType:   targetType,
		Action:       action,
		Metadata:     string(metadataJSON),
		Success:      success,
		ErrorMessage: errorMessage,
	}

	return r.Log(ctx, log)
}
