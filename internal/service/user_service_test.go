package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourorg/anonymous-support/internal/domain"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	args := m.Called(ctx, username)
	return args.Bool(0), args.Error(1)
}

// MockAnalyticsRepository is a mock for analytics
type MockAnalyticsRepository struct {
	mock.Mock
}

func (m *MockAnalyticsRepository) IncrementUserActivity(ctx context.Context, userID uuid.UUID, activityType string) error {
	args := m.Called(ctx, userID, activityType)
	return args.Error(0)
}

func TestUserService_GetProfile(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockAnalyticsRepo := new(MockAnalyticsRepository)
	service := NewUserService(mockUserRepo, mockAnalyticsRepo)

	userID := uuid.New()
	expectedUser := &domain.User{
		ID:        userID,
		Username:  "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockUserRepo.On("GetByID", mock.Anything, userID).Return(expectedUser, nil)

	// Act
	profile, err := service.GetProfile(context.Background(), userID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, expectedUser.ID.String(), profile.ID)
	assert.Equal(t, expectedUser.Username, profile.Username)
	mockUserRepo.AssertExpectations(t)
}
