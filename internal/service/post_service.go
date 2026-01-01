package service

import (
	"context"
	"fmt"
	"time"

	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/pkg/cache"
	"github.com/yourorg/anonymous-support/internal/pkg/feed"
	"github.com/yourorg/anonymous-support/internal/pkg/metrics"
	"github.com/yourorg/anonymous-support/internal/pkg/moderator"
	"github.com/yourorg/anonymous-support/internal/pkg/validator"
	"github.com/yourorg/anonymous-support/internal/repository"
)

type PostService struct {
	postRepo      repository.PostRepository
	realtimeRepo  repository.RealtimeRepository
	contentFilter *moderator.ContentFilter
	cache         *cache.Cache
	feedRanker    *feed.FeedRanker
}

func NewPostService(
	postRepo repository.PostRepository,
	realtimeRepo repository.RealtimeRepository,
	contentFilter *moderator.ContentFilter,
	cache *cache.Cache,
) *PostService {
	return &PostService{
		postRepo:      postRepo,
		realtimeRepo:  realtimeRepo,
		contentFilter: contentFilter,
		cache:         cache,
		feedRanker:    feed.NewFeedRanker(),
	}
}

func (s *PostService) CreatePost(ctx context.Context, userID, username string, postType domain.PostType, content string, categories []string, urgencyLevel int, timeContext string, daysSinceRelapse int, tags []string, visibility string, circleID *string) (*domain.Post, error) {
	if err := validator.ValidatePostContent(content); err != nil {
		return nil, err
	}

	post := &domain.Post{
		UserID:       userID,
		Username:     username,
		Type:         postType,
		Content:      content,
		Categories:   categories,
		UrgencyLevel: urgencyLevel,
		Visibility:   visibility,
		CircleID:     circleID,
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
		_ = s.realtimeRepo.PublishNewPost(ctx, post.ID.Hex(), string(postType), categories)
		feedScore := float64(time.Now().Unix())
		_ = s.realtimeRepo.AddToFeed(ctx, "feed:global:latest", post.ID.Hex(), feedScore)
	}

	// Emit metrics
	metrics.PostsCreatedTotal.WithLabelValues(string(postType)).Inc()

	return post, nil
}

func (s *PostService) GetPost(ctx context.Context, postID string) (*domain.Post, error) {
	_ = s.realtimeRepo.IncrementViewCount(ctx, postID)
	return s.postRepo.GetByID(ctx, postID)
}

func (s *PostService) GetFeed(ctx context.Context, categories []string, circleID *string, postType *domain.PostType, limit, offset int) ([]*domain.Post, error) {
	// Build cache key
	cacheKey := fmt.Sprintf("feed:%v:%v:%v:%d:%d", categories, circleID, postType, limit, offset)

	// Try cache first
	var cachedPosts []*domain.Post
	found, err := s.cache.Get(ctx, cacheKey, &cachedPosts)
	if err == nil && found {
		metrics.CacheHitsTotal.WithLabelValues("feed").Inc()
		return cachedPosts, nil
	}
	metrics.CacheMissesTotal.WithLabelValues("feed").Inc()

	// Cache miss - fetch from DB
	posts, err := s.postRepo.GetFeed(ctx, categories, circleID, postType, limit, offset)
	if err != nil {
		return nil, err
	}

	// Store in cache (5 min TTL)
	_ = s.cache.Set(ctx, cacheKey, posts, 5*time.Minute)

	return posts, nil
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
	return s.postRepo.UpdateUrgency(ctx, postID, int32(urgencyLevel)) //nolint:gosec // Urgency level 1-10
}

// GetPersonalizedFeed returns a feed ranked by relevance to the user
func (s *PostService) GetPersonalizedFeed(ctx context.Context, userPrefs *feed.UserPreferences, limit, offset int) ([]*domain.Post, error) {
	// Build cache key with user preferences hash
	cacheKey := fmt.Sprintf("feed:personalized:%v:%d:%d", userPrefs.PreferredCategories, limit, offset)

	// Try cache first
	var cachedPosts []*domain.Post
	found, err := s.cache.Get(ctx, cacheKey, &cachedPosts)
	if err == nil && found {
		metrics.CacheHitsTotal.WithLabelValues("personalized_feed").Inc()
		return cachedPosts, nil
	}
	metrics.CacheMissesTotal.WithLabelValues("personalized_feed").Inc()

	// Fetch larger set for ranking (2x limit for better personalization)
	fetchLimit := limit * 2
	posts, err := s.postRepo.GetFeed(ctx, userPrefs.PreferredCategories, nil, nil, fetchLimit, 0)
	if err != nil {
		return nil, err
	}

	// Filter out blocked users
	filtered := feed.FilterPosts(posts, userPrefs)

	// Rank posts
	ranked := s.feedRanker.RankPosts(ctx, filtered, userPrefs)

	// Extract posts from ranked results
	result := make([]*domain.Post, 0, limit)
	start := offset
	end := offset + limit
	if start >= len(ranked) {
		return result, nil
	}
	if end > len(ranked) {
		end = len(ranked)
	}

	for i := start; i < end; i++ {
		result = append(result, ranked[i].Post)
	}

	// Cache for shorter TTL since it's personalized (2 min)
	_ = s.cache.Set(ctx, cacheKey, result, 2*time.Minute)

	return result, nil
}
