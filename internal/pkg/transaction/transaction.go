package transaction

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// TxFunc is a function that runs within a transaction
type TxFunc func(*sqlx.Tx) error

// Manager handles database transactions
type Manager struct {
	db     *sqlx.DB
	logger *zap.Logger
}

// NewManager creates a new transaction manager
func NewManager(db *sqlx.DB, logger *zap.Logger) *Manager {
	return &Manager{
		db:     db,
		logger: logger,
	}
}

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, it is committed
func (m *Manager) WithTransaction(ctx context.Context, fn TxFunc) error {
	// Begin transaction
	tx, err := m.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback on panic
	defer func() {
		if r := recover(); r != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				m.logger.Error("Failed to rollback transaction after panic",
					zap.Error(rbErr),
					zap.Any("panic", r))
			}
			panic(r) // Re-throw panic after rollback
		}
	}()

	// Execute function
	if err := fn(tx); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(); rbErr != nil {
			m.logger.Error("Failed to rollback transaction",
				zap.Error(rbErr),
				zap.NamedError("original_error", err))
			return fmt.Errorf("transaction failed and rollback failed: %w (original error: %v)", rbErr, err)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithTransactionIsolation executes a function within a transaction with custom isolation level
func (m *Manager) WithTransactionIsolation(ctx context.Context, isolation sql.IsolationLevel, fn TxFunc) error {
	tx, err := m.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: isolation,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				m.logger.Error("Failed to rollback transaction after panic",
					zap.Error(rbErr),
					zap.Any("panic", r))
			}
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			m.logger.Error("Failed to rollback transaction",
				zap.Error(rbErr),
				zap.NamedError("original_error", err))
			return fmt.Errorf("transaction failed and rollback failed: %w (original error: %v)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Example usage patterns:

// JoinCircleTransaction joins a user to a circle with transactional integrity
func JoinCircleTransaction(ctx context.Context, tm *Manager, circleID, userID string) error {
	return tm.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		// Check if circle is full
		var memberCount, maxMembers int
		err := tx.GetContext(ctx, &memberCount, "SELECT member_count FROM circles WHERE id = $1", circleID)
		if err != nil {
			return fmt.Errorf("failed to get member count: %w", err)
		}

		err = tx.GetContext(ctx, &maxMembers, "SELECT max_members FROM circles WHERE id = $1", circleID)
		if err != nil {
			return fmt.Errorf("failed to get max members: %w", err)
		}

		if memberCount >= maxMembers {
			return fmt.Errorf("circle is full")
		}

		// Add membership
		_, err = tx.ExecContext(ctx,
			"INSERT INTO circle_memberships (circle_id, user_id, joined_at) VALUES ($1, $2, NOW())",
			circleID, userID)
		if err != nil {
			return fmt.Errorf("failed to create membership: %w", err)
		}

		// Increment member count
		_, err = tx.ExecContext(ctx,
			"UPDATE circles SET member_count = member_count + 1, updated_at = NOW() WHERE id = $1",
			circleID)
		if err != nil {
			return fmt.Errorf("failed to update member count: %w", err)
		}

		return nil
	})
}

// LeaveCircleTransaction removes a user from a circle with transactional integrity
func LeaveCircleTransaction(ctx context.Context, tm *Manager, circleID, userID string) error {
	return tm.WithTransaction(ctx, func(tx *sqlx.Tx) error {
		// Remove membership
		result, err := tx.ExecContext(ctx,
			"DELETE FROM circle_memberships WHERE circle_id = $1 AND user_id = $2",
			circleID, userID)
		if err != nil {
			return fmt.Errorf("failed to delete membership: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("membership not found")
		}

		// Decrement member count
		_, err = tx.ExecContext(ctx,
			"UPDATE circles SET member_count = member_count - 1, updated_at = NOW() WHERE id = $1",
			circleID)
		if err != nil {
			return fmt.Errorf("failed to update member count: %w", err)
		}

		return nil
	})
}
