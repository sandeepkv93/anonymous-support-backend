package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yourorg/anonymous-support/internal/domain"
)

type ModerationRepository struct {
	db *sqlx.DB
}

func NewModerationRepository(db *sqlx.DB) *ModerationRepository {
	return &ModerationRepository{db: db}
}

func (r *ModerationRepository) CreateReport(ctx context.Context, report *domain.ContentReport) error {
	query := `
		INSERT INTO content_reports (id, reporter_id, content_type, content_id, reason, description, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at
	`
	return r.db.QueryRowContext(ctx, query,
		report.ID, report.ReporterID, report.ContentType, report.ContentID,
		report.Reason, report.Description, report.Status,
	).Scan(&report.CreatedAt)
}

func (r *ModerationRepository) GetReports(ctx context.Context, status *string, limit, offset int) ([]*domain.ContentReport, error) {
	reports := []*domain.ContentReport{}
	var query string
	var args []interface{}

	if status != nil {
		query = `SELECT * FROM content_reports WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = []interface{}{*status, limit, offset}
	} else {
		query = `SELECT * FROM content_reports ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	}

	err := r.db.SelectContext(ctx, &reports, query, args...)
	return reports, err
}

func (r *ModerationRepository) GetReportByID(ctx context.Context, id uuid.UUID) (*domain.ContentReport, error) {
	var report domain.ContentReport
	query := `SELECT * FROM content_reports WHERE id = $1`
	err := r.db.GetContext(ctx, &report, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("report not found")
	}
	return &report, err
}

func (r *ModerationRepository) UpdateReportStatus(ctx context.Context, reportID uuid.UUID, status string, reviewerID uuid.UUID) error {
	query := `
		UPDATE content_reports
		SET status = $1, reviewed_by = $2, reviewed_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, status, reviewerID, reportID)
	return err
}

func (r *ModerationRepository) CreateBlock(ctx context.Context, block *domain.UserBlock) error {
	query := `
		INSERT INTO user_blocks (id, blocker_id, blocked_id)
		VALUES ($1, $2, $3)
		RETURNING created_at
	`
	return r.db.QueryRowContext(ctx, query,
		block.ID, block.BlockerID, block.BlockedID,
	).Scan(&block.CreatedAt)
}

func (r *ModerationRepository) IsBlocked(ctx context.Context, blockerID, blockedID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM user_blocks WHERE blocker_id = $1 AND blocked_id = $2)`
	err := r.db.GetContext(ctx, &exists, query, blockerID, blockedID)
	return exists, err
}
