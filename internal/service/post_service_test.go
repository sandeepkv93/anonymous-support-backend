package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/pkg/validator"
)

// TestPostValidation tests post content validation
func TestPostValidation(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid content",
			content: "This is a valid post about my recovery journey",
			wantErr: false,
		},
		{
			name:    "empty content",
			content: "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			content: "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePostContent(tt.content)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestPostTypeValidation tests post type values
func TestPostTypeValidation(t *testing.T) {
	validTypes := []domain.PostType{
		domain.PostTypeSOS,
		domain.PostTypeVictory,
		domain.PostTypeCheckIn,
		domain.PostTypeQuestion,
	}

	for _, pt := range validTypes {
		assert.NotEmpty(t, string(pt), "PostType should not be empty")
	}
}

// Note: Service-level tests with mocked repositories are difficult because
// the service constructors take concrete types (*mongodb.PostRepository, *redis.RealtimeRepository)
// instead of interfaces.
//
// For proper unit testing of service logic, consider:
// 1. Refactoring services to accept interfaces instead of concrete types
// 2. Using integration tests that spin up real test databases
// 3. Testing at the repository level with real database connections
//
// The integration tests in tests/integration/ provide end-to-end coverage.
