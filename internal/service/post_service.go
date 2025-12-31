package service

import (
	"context"
	"time"

	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/pkg/moderator"
	"github.com/yourorg/anonymous-support/internal/pkg/validator"
	"github.com/yourorg/anonymous-support/internal/repository/mongodb"
	"github.com/yourorg/anonymous-support/internal/repository/redis"
)

type PostService struct {
	postRepo      *mongodb.PostRepository
	realtimeRepo  *redis.RealtimeRepository
	contentFilter *moderator.ContentFilter
}

func NewPostService(
	postRepo *mongodb.PostRepository,
	realtimeRepo *redis.RealtimeRepository,
	contentFilter *moderator.ContentFilter,
) *PostService {
	return &PostService{
		postRepo:      postRepo,
		realtimeRepo:  realtimeRepo,
		contentFilter: contentFilter,
	}
}

func (s *PostService) CreatePost(ctx context.Context, userID, username string, postType domain.PostType, content string, categories []string, urgencyLevel int, timeContext string, daysSinceRelapse int, tags []string, visibility string, circleID *string) (*domain.Post, error) {
	if err := validator.ValidatePostContent(content); err != nil {
		return nil, err
	}

	post := &domain.Post{
		UserID:        userID,
		Username:      username,
		Type:          postType,
		Content:       content,
		Categories:    categories,
		UrgencyLevel:  urgencyLevel,
		Visibility:    visibility,
		CircleID:      circleID,
		Context: domain.PostContext{
			DaysSinceRelapse: daysSinceRelapse,
			TimeContext:      timeContext,
			Tags:             tags,
		},
		IsModerated: false,
	}

	flags := s.contentFilter.CheckContent(content)
	if len(flags) > 0 {
		post.IsModerated = true
		post.ModerationFlags = flags
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, err
	}

	if !post.IsModerated {
		s.realtimeRepo.PublishNewPost(ctx, post.ID.Hex(), string(postType), categories)
		feedScore := float64(time.Now().Unix())
		s.realtimeRepo.AddToFeed(ctx, "feed:global:latest", post.ID.Hex(), feedScore)
	}

	return post, nil
}

func (s *PostService) GetPost(ctx context.Context, postID string) (*domain.Post, error) {
	s.realtimeRepo.IncrementViewCount(ctx, postID)
	return s.postRepo.GetByID(ctx, postID)
}

func (s *PostService) GetFeed(ctx context.Context, categories []string, circleID *string, postType *domain.PostType, limit, offset int) ([]*domain.Post, error) {
	return s.postRepo.GetFeed(ctx, categories, circleID, postType, limit, offset)
}

func (s *PostService) DeletePost(ctx context.Context, postID, userID string) error {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}

	if post.UserID != userID {
		return nil
	}

	return s.postRepo.Delete(ctx, postID)
}

func (s *PostService) UpdatePostUrgency(ctx context.Context, postID string, urgencyLevel int) error {
	return s.postRepo.UpdateUrgency(ctx, postID, urgencyLevel)
}
