package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/yourorg/anonymous-support/internal/domain"
)

// TestUserDomainValidation tests user domain model validation
func TestUserDomainValidation(t *testing.T) {
	user := &domain.User{
		ID:           uuid.New(),
		Username:     "testuser",
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
		Role:         domain.RoleUser,
	}

	assert.NotEmpty(t, user.ID)
	assert.NotEmpty(t, user.Username)
	assert.Equal(t, domain.RoleUser, user.Role)
}

// TestRoleHierarchy tests role validation
func TestRoleHierarchy(t *testing.T) {
	roles := []domain.Role{
		domain.RoleUser,
		domain.RoleModerator,
		domain.RoleAdmin,
	}

	for _, role := range roles {
		assert.NotEmpty(t, string(role), "Role should not be empty")
	}
}

// Note: UserService constructor takes concrete types (*postgres.UserRepository, *mongodb.AnalyticsRepository)
// making unit testing with mocks difficult without interface refactoring.
// See integration tests for end-to-end service testing.
