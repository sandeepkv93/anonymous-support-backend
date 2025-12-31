package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yourorg/anonymous-support/internal/domain"
)

type CircleRepository struct {
	db *sqlx.DB
}

func NewCircleRepository(db *sqlx.DB) *CircleRepository {
	return &CircleRepository{db: db}
}

func (r *CircleRepository) Create(ctx context.Context, circle *domain.Circle) error {
	query := `
		INSERT INTO circles (id, name, description, category, max_members, is_private, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at
	`
	return r.db.QueryRowContext(ctx, query,
		circle.ID, circle.Name, circle.Description, circle.Category,
		circle.MaxMembers, circle.IsPrivate, circle.CreatedBy,
	).Scan(&circle.CreatedAt)
}

func (r *CircleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Circle, error) {
	var circle domain.Circle
	query := `SELECT * FROM circles WHERE id = $1`
	err := r.db.GetContext(ctx, &circle, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("circle not found")
	}
	return &circle, err
}

func (r *CircleRepository) List(ctx context.Context, category *string, limit, offset int) ([]*domain.Circle, error) {
	circles := []*domain.Circle{}
	var query string
	var args []interface{}

	if category != nil {
		query = `SELECT * FROM circles WHERE category = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = []interface{}{*category, limit, offset}
	} else {
		query = `SELECT * FROM circles ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	}

	err := r.db.SelectContext(ctx, &circles, query, args...)
	return circles, err
}

func (r *CircleRepository) JoinCircle(ctx context.Context, circleID, userID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		INSERT INTO circle_memberships (id, circle_id, user_id, role)
		VALUES ($1, $2, $3, $4)
	`
	membershipID := uuid.New()
	if _, err := tx.ExecContext(ctx, query,
		membershipID, circleID, userID, "member",
	); err != nil {
		return err
	}

	updateQuery := `UPDATE circles SET member_count = member_count + 1 WHERE id = $1`
	if _, err := tx.ExecContext(ctx, updateQuery, circleID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *CircleRepository) LeaveCircle(ctx context.Context, circleID, userID uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	deleteQuery := `DELETE FROM circle_memberships WHERE circle_id = $1 AND user_id = $2`
	if _, err := tx.ExecContext(ctx, deleteQuery, circleID, userID); err != nil {
		return err
	}

	updateQuery := `UPDATE circles SET member_count = member_count - 1 WHERE id = $1`
	if _, err := tx.ExecContext(ctx, updateQuery, circleID); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *CircleRepository) GetMembers(ctx context.Context, circleID uuid.UUID, limit, offset int) ([]uuid.UUID, error) {
	members := []uuid.UUID{}
	query := `
		SELECT user_id FROM circle_memberships
		WHERE circle_id = $1
		ORDER BY joined_at DESC
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &members, query, circleID, limit, offset)
	return members, err
}

func (r *CircleRepository) IsMember(ctx context.Context, circleID, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM circle_memberships WHERE circle_id = $1 AND user_id = $2)`
	err := r.db.GetContext(ctx, &exists, query, circleID, userID)
	return exists, err
}

func (r *CircleRepository) GetMemberCount(ctx context.Context, circleID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM circle_memberships WHERE circle_id = $1`
	err := r.db.GetContext(ctx, &count, query, circleID)
	return count, err
}
