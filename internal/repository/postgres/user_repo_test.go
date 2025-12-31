package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/repository/postgres"
)

// Integration test example using real database
// Run with: go test -tags=integration ./internal/repository/postgres/...

func setupTestDB(t *testing.T) *sqlx.DB {
	// This would typically use testcontainers or a test database
	// For now, this is a template
	t.Skip("Integration tests require database setup")
	return nil
}

func TestUserRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewUserRepository(db)

	ctx := context.Background()
	email := "test@example.com"
	user := &domain.User{
		ID:             uuid.New(),
		Username:       "testuser",
		Email:          &email,
		PasswordHash:   "hashed_password",
		AvatarID:       1,
		IsAnonymous:    false,
		StrengthPoints: 0,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)
	assert.NotZero(t, user.CreatedAt)
	assert.NotZero(t, user.LastActiveAt)
}

func TestUserRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewUserRepository(db)
	ctx := context.Background()

	// Create test user
	email := "test@example.com"
	user := &domain.User{
		ID:             uuid.New(),
		Username:       "testuser",
		Email:          &email,
		PasswordHash:   "hashed_password",
		AvatarID:       1,
		IsAnonymous:    false,
		StrengthPoints: 0,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Retrieve user
	retrieved, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Username, retrieved.Username)
	assert.Equal(t, user.Email, retrieved.Email)
}

func TestUserRepository_UsernameExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewUserRepository(db)
	ctx := context.Background()

	username := "existinguser"

	// Initially should not exist
	exists, err := repo.UsernameExists(ctx, username)
	require.NoError(t, err)
	assert.False(t, exists)

	// Create user
	email2 := "test@example.com"
	user := &domain.User{
		ID:             uuid.New(),
		Username:       username,
		Email:          &email2,
		PasswordHash:   "hashed_password",
		AvatarID:       1,
		IsAnonymous:    false,
		StrengthPoints: 0,
	}

	err = repo.Create(ctx, user)
	require.NoError(t, err)

	// Now should exist
	exists, err = repo.UsernameExists(ctx, username)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepository_UpdateStrengthPoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewUserRepository(db)
	ctx := context.Background()

	// Create user
	email4 := "test@example.com"
	user := &domain.User{
		ID:             uuid.New(),
		Username:       "testuser",
		Email:          &email4,
		PasswordHash:   "hashed_password",
		AvatarID:       1,
		IsAnonymous:    false,
		StrengthPoints: 100,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Update strength points
	err = repo.UpdateStrengthPoints(ctx, user.ID, 50)
	require.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, 150, updated.StrengthPoints)
}
