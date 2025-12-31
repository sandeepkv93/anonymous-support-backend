package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/pkg/transaction"
	"github.com/yourorg/anonymous-support/internal/repository/mongodb"
	"github.com/yourorg/anonymous-support/internal/repository/postgres"
)

type CircleService struct {
	circleRepo *postgres.CircleRepository
	postRepo   *mongodb.PostRepository
	txManager  *transaction.Manager
}

func NewCircleService(
	circleRepo *postgres.CircleRepository,
	postRepo *mongodb.PostRepository,
	txManager *transaction.Manager,
) *CircleService {
	return &CircleService{
		circleRepo: circleRepo,
		postRepo:   postRepo,
		txManager:  txManager,
	}
}

func (s *CircleService) CreateCircle(ctx context.Context, userID, name, description, category string, maxMembers int, isPrivate bool) (string, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return "", err
	}

	circleID := uuid.New()

	// Use transaction to ensure atomicity of circle creation and auto-join
	err = s.txManager.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		// Create circle
		circleQuery := `
			INSERT INTO circles (id, name, description, category, max_members, member_count, is_private, created_by, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, 1, $6, $7, NOW(), NOW())
		`
		if _, err := tx.ExecContext(ctx, circleQuery, circleID, name, description, category, maxMembers, isPrivate, uid); err != nil {
			return fmt.Errorf("failed to create circle: %w", err)
		}

		// Auto-join creator to circle
		membershipQuery := `
			INSERT INTO circle_memberships (circle_id, user_id, joined_at)
			VALUES ($1, $2, NOW())
		`
		if _, err := tx.ExecContext(ctx, membershipQuery, circleID, uid); err != nil {
			return fmt.Errorf("failed to join creator to circle: %w", err)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	return circleID.String(), nil
}

func (s *CircleService) JoinCircle(ctx context.Context, userID, circleID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	cid, err := uuid.Parse(circleID)
	if err != nil {
		return err
	}

	// Use transaction with row locking to prevent race conditions
	return s.txManager.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		// Lock the circle row for update and check capacity
		var memberCount, maxMembers int
		lockQuery := `SELECT member_count, max_members FROM circles WHERE id = $1 FOR UPDATE`
		if err := tx.QueryRowContext(ctx, lockQuery, cid).Scan(&memberCount, &maxMembers); err != nil {
			return fmt.Errorf("circle not found: %w", err)
		}

		// Check if circle is full
		if memberCount >= maxMembers {
			return fmt.Errorf("circle is full")
		}

		// Check if already a member
		var existingCount int
		checkQuery := `SELECT COUNT(*) FROM circle_memberships WHERE circle_id = $1 AND user_id = $2`
		if err := tx.QueryRowContext(ctx, checkQuery, cid, uid).Scan(&existingCount); err != nil {
			return fmt.Errorf("failed to check existing membership: %w", err)
		}

		if existingCount > 0 {
			return fmt.Errorf("already a member of this circle")
		}

		// Create membership
		insertQuery := `INSERT INTO circle_memberships (circle_id, user_id, joined_at) VALUES ($1, $2, NOW())`
		if _, err := tx.ExecContext(ctx, insertQuery, cid, uid); err != nil {
			return fmt.Errorf("failed to create membership: %w", err)
		}

		// Increment member count
		updateQuery := `UPDATE circles SET member_count = member_count + 1, updated_at = NOW() WHERE id = $1`
		if _, err := tx.ExecContext(ctx, updateQuery, cid); err != nil {
			return fmt.Errorf("failed to update member count: %w", err)
		}

		return nil
	})
}

func (s *CircleService) LeaveCircle(ctx context.Context, userID, circleID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	cid, err := uuid.Parse(circleID)
	if err != nil {
		return err
	}

	// Use transaction to ensure atomicity of membership removal and count update
	return s.txManager.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		// Delete membership
		deleteQuery := `DELETE FROM circle_memberships WHERE circle_id = $1 AND user_id = $2`
		result, err := tx.ExecContext(ctx, deleteQuery, cid, uid)
		if err != nil {
			return fmt.Errorf("failed to delete membership: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("not a member of this circle")
		}

		// Decrement member count
		updateQuery := `UPDATE circles SET member_count = member_count - 1, updated_at = NOW() WHERE id = $1`
		if _, err := tx.ExecContext(ctx, updateQuery, cid); err != nil {
			return fmt.Errorf("failed to update member count: %w", err)
		}

		return nil
	})
}

func (s *CircleService) GetCircleMembers(ctx context.Context, circleID string, limit, offset int) ([]*domain.CircleMembership, error) {
	cid, err := uuid.Parse(circleID)
	if err != nil {
		return nil, err
	}

	memberIDs, err := s.circleRepo.GetMembers(ctx, cid, limit, offset)
	if err != nil {
		return nil, err
	}
	memberships := make([]*domain.CircleMembership, len(memberIDs))
	for i, uid := range memberIDs {
		memberships[i] = &domain.CircleMembership{UserID: uid, CircleID: cid}
	}
	return memberships, nil
}

func (s *CircleService) GetCircleFeed(ctx context.Context, circleID string, limit, offset int) ([]*domain.Post, error) {
	return s.postRepo.GetFeed(ctx, nil, &circleID, nil, limit, offset)
}

func (s *CircleService) GetCircles(ctx context.Context, category *string, limit, offset int) ([]*domain.Circle, error) {
	return s.circleRepo.List(ctx, category, limit, offset)
}
