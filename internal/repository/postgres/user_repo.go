package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/repository"
)

// Compile-time check to ensure UserRepository implements repository.UserRepository
var _ repository.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, avatar_id, is_anonymous, strength_points)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, last_active_at
	`
	return r.db.QueryRowContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.AvatarID, user.IsAnonymous, user.StrengthPoints,
	).Scan(&user.CreatedAt, &user.LastActiveAt)
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	query := `SELECT * FROM users WHERE id = $1 AND is_banned = false`
	err := r.db.GetContext(ctx, &user, query, id)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return &user, err
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	query := `SELECT * FROM users WHERE username = $1 AND is_banned = false`
	err := r.db.GetContext(ctx, &user, query, username)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return &user, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	query := `SELECT * FROM users WHERE email = $1 AND is_banned = false`
	err := r.db.GetContext(ctx, &user, query, email)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return &user, err
}

func (r *UserRepository) UpdateLastActive(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET last_active_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *UserRepository) UpdateStrengthPoints(ctx context.Context, userID uuid.UUID, points int) error {
	query := `UPDATE users SET strength_points = strength_points + $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, points, userID)
	return err
}

func (r *UserRepository) UpdateProfile(ctx context.Context, userID uuid.UUID, username *string, avatarID *int) error {
	if username != nil {
		query := `UPDATE users SET username = $1 WHERE id = $2`
		if _, err := r.db.ExecContext(ctx, query, *username, userID); err != nil {
			return err
		}
	}
	if avatarID != nil {
		query := `UPDATE users SET avatar_id = $1 WHERE id = $2`
		if _, err := r.db.ExecContext(ctx, query, *avatarID, userID); err != nil {
			return err
		}
	}
	return nil
}

func (r *UserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	err := r.db.GetContext(ctx, &exists, query, username)
	return exists, err
}
