package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/pkg/cache"
	"github.com/yourorg/anonymous-support/internal/pkg/moderator"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) Create(ctx context.Context, post *domain.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostRepository) GetByID(ctx context.Context, id string) (*domain.Post, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Post), args.Error(1)
}

func (m *MockPostRepository) GetFeed(ctx context.Context, categories []string, circleID *string, postType *domain.PostType, limit, offset int) ([]*domain.Post, error) {
	args := m.Called(ctx, categories, circleID, postType, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Post), args.Error(1)
}

func (m *MockPostRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostRepository) UpdateUrgency(ctx context.Context, id string, urgencyLevel int32) error {
	args := m.Called(ctx, id, urgencyLevel)
	return args.Error(0)
}

func (m *MockPostRepository) IncrementResponseCount(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPostRepository) IncrementSupportCount(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockRealtimeRepository struct {
	mock.Mock
}

func (m *MockRealtimeRepository) PublishNewPost(ctx context.Context, postID, postType string, categories []string) error {
	args := m.Called(ctx, postID, postType, categories)
	return args.Error(0)
}

func (m *MockRealtimeRepository) IncrementViewCount(ctx context.Context, postID string) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockRealtimeRepository) AddToFeed(ctx context.Context, feedKey, postID string, score float64) error {
	args := m.Called(ctx, feedKey, postID, score)
	return args.Error(0)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	args := m.Called(ctx, key, dest)
	return args.Bool(0), args.Error(1)
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl int) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func TestPostService_CreatePost(t *testing.T) {
	mockPostRepo := new(MockPostRepository)
	mockRealtimeRepo := new(MockRealtimeRepository)
	mockCache := &cache.Cache{} // Use real cache or mock if needed
	contentFilter := moderator.NewContentFilter("low")

	service := NewPostService(mockPostRepo, mockRealtimeRepo, contentFilter, mockCache)

	ctx := context.Background()
	userID := "user123"
	username := "testuser"
	content := "Test post content"

	mockPostRepo.On("Create", ctx, mock.AnythingOfType("*domain.Post")).Return(nil)
	mockRealtimeRepo.On("PublishNewPost", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockRealtimeRepo.On("AddToFeed", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	post, err := service.CreatePost(ctx, userID, username, domain.PostTypeSOS, content, []string{"test"}, 5, "evening", 10, []string{}, "public", nil)

	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, userID, post.UserID)
	assert.Equal(t, content, post.Content)
	mockPostRepo.AssertExpectations(t)
}

func TestPostService_GetPost(t *testing.T) {
	mockPostRepo := new(MockPostRepository)
	mockRealtimeRepo := new(MockRealtimeRepository)
	mockCache := &cache.Cache{}
	contentFilter := moderator.NewContentFilter("low")

	service := NewPostService(mockPostRepo, mockRealtimeRepo, contentFilter, mockCache)

	ctx := context.Background()
	postID := primitive.NewObjectID().Hex()
	expectedPost := &domain.Post{
		ID:      primitive.NewObjectID(),
		UserID:  "user123",
		Content: "Test content",
	}

	mockRealtimeRepo.On("IncrementViewCount", ctx, postID).Return(nil)
	mockPostRepo.On("GetByID", ctx, postID).Return(expectedPost, nil)

	post, err := service.GetPost(ctx, postID)

	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, expectedPost.UserID, post.UserID)
	mockPostRepo.AssertExpectations(t)
	mockRealtimeRepo.AssertExpectations(t)
}

func TestPostService_GetFeed(t *testing.T) {
	mockPostRepo := new(MockPostRepository)
	mockRealtimeRepo := new(MockRealtimeRepository)
	mockCache := &cache.Cache{}
	contentFilter := moderator.NewContentFilter("low")

	service := NewPostService(mockPostRepo, mockRealtimeRepo, contentFilter, mockCache)

	ctx := context.Background()
	categories := []string{"mental-health"}
	postType := domain.PostTypeSOS
	limit := 10
	offset := 0

	expectedPosts := []*domain.Post{
		{
			ID:      primitive.NewObjectID(),
			UserID:  "user123",
			Content: "Test post 1",
			Type:    domain.PostTypeSOS,
		},
		{
			ID:      primitive.NewObjectID(),
			UserID:  "user456",
			Content: "Test post 2",
			Type:    domain.PostTypeSOS,
		},
	}

	mockPostRepo.On("GetFeed", ctx, categories, (*string)(nil), &postType, limit, offset).Return(expectedPosts, nil)

	posts, err := service.GetFeed(ctx, categories, nil, &postType, limit, offset)

	assert.NoError(t, err)
	assert.NotNil(t, posts)
	assert.Len(t, posts, 2)
	mockPostRepo.AssertExpectations(t)
}

func TestPostService_DeletePost(t *testing.T) {
	mockPostRepo := new(MockPostRepository)
	mockRealtimeRepo := new(MockRealtimeRepository)
	mockCache := &cache.Cache{}
	contentFilter := moderator.NewContentFilter("low")

	service := NewPostService(mockPostRepo, mockRealtimeRepo, contentFilter, mockCache)

	ctx := context.Background()
	postID := primitive.NewObjectID().Hex()
	userID := "user123"

	post := &domain.Post{
		ID:      primitive.NewObjectID(),
		UserID:  userID,
		Content: "Test content",
	}

	mockPostRepo.On("GetByID", ctx, postID).Return(post, nil)
	mockPostRepo.On("Delete", ctx, postID).Return(nil)

	err := service.DeletePost(ctx, postID, userID)

	assert.NoError(t, err)
	mockPostRepo.AssertExpectations(t)
}

func TestPostService_DeletePost_Unauthorized(t *testing.T) {
	mockPostRepo := new(MockPostRepository)
	mockRealtimeRepo := new(MockRealtimeRepository)
	mockCache := &cache.Cache{}
	contentFilter := moderator.NewContentFilter("low")

	service := NewPostService(mockPostRepo, mockRealtimeRepo, contentFilter, mockCache)

	ctx := context.Background()
	postID := primitive.NewObjectID().Hex()
	userID := "user123"
	differentUserID := "user456"

	post := &domain.Post{
		ID:      primitive.NewObjectID(),
		UserID:  differentUserID,
		Content: "Test content",
	}

	mockPostRepo.On("GetByID", ctx, postID).Return(post, nil)

	err := service.DeletePost(ctx, postID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
	mockPostRepo.AssertExpectations(t)
}
