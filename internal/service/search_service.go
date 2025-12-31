package service

import (
	"context"
	"strings"

	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/repository"
)

type SearchService struct {
	postRepo   repository.PostRepository
	circleRepo repository.CircleRepository
	userRepo   repository.UserRepository
}

func NewSearchService(
	postRepo repository.PostRepository,
	circleRepo repository.CircleRepository,
	userRepo repository.UserRepository,
) *SearchService {
	return &SearchService{
		postRepo:   postRepo,
		circleRepo: circleRepo,
		userRepo:   userRepo,
	}
}

// SearchResults represents combined search results
type SearchResults struct {
	Posts   []*domain.Post   `json:"posts"`
	Circles []*domain.Circle `json:"circles"`
	Query   string           `json:"query"`
	Total   int              `json:"total"`
}

// Search performs a full-text search across posts and circles
func (s *SearchService) Search(ctx context.Context, query string, filters *SearchFilters) (*SearchResults, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return &SearchResults{
			Posts:   []*domain.Post{},
			Circles: []*domain.Circle{},
			Query:   query,
			Total:   0,
		}, nil
	}

	results := &SearchResults{
		Query: query,
	}

	// Search posts
	if filters == nil || filters.IncludePosts {
		posts, err := s.searchPosts(ctx, query, filters)
		if err == nil {
			results.Posts = posts
		}
	}

	// Search circles
	if filters == nil || filters.IncludeCircles {
		circles, err := s.searchCircles(ctx, query, filters)
		if err == nil {
			results.Circles = circles
		}
	}

	results.Total = len(results.Posts) + len(results.Circles)

	return results, nil
}

// searchPosts searches posts by content and categories
func (s *SearchService) searchPosts(ctx context.Context, query string, filters *SearchFilters) ([]*domain.Post, error) {
	// TODO: Implement full-text search using MongoDB text index or Elasticsearch
	// For now, use basic category filtering

	limit := 20
	if filters != nil && filters.Limit > 0 {
		limit = filters.Limit
	}

	categories := []string{}
	if filters != nil {
		categories = filters.Categories
	}

	// Fallback to basic feed retrieval with category filter
	posts, err := s.postRepo.GetFeed(ctx, categories, nil, nil, limit, 0)
	if err != nil {
		return nil, err
	}

	// Filter by query in content (basic text matching)
	filtered := []*domain.Post{}
	queryLower := strings.ToLower(query)
	for _, post := range posts {
		if strings.Contains(strings.ToLower(post.Content), queryLower) {
			filtered = append(filtered, post)
		}
	}

	return filtered, nil
}

// searchCircles searches circles by name and category
func (s *SearchService) searchCircles(ctx context.Context, query string, filters *SearchFilters) ([]*domain.Circle, error) {
	// TODO: Implement proper circle search
	// For now, return empty results
	return []*domain.Circle{}, nil
}

// SearchFilters defines search filtering options
type SearchFilters struct {
	IncludePosts   bool
	IncludeCircles bool
	Categories     []string
	PostType       *domain.PostType
	Limit          int
	Offset         int
}
