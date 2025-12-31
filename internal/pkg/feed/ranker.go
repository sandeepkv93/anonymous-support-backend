package feed

import (
	"context"
	"math"
	"time"

	"github.com/yourorg/anonymous-support/internal/domain"
)

// FeedRanker implements personalized feed ranking algorithm
type FeedRanker struct {
	weights RankingWeights
}

// RankingWeights defines the weights for different ranking factors
type RankingWeights struct {
	RecencyWeight    float64 // How recent the post is
	UrgencyWeight    float64 // Post urgency level
	EngagementWeight float64 // Responses and support count
	CategoryWeight   float64 // User's preferred categories
	CircleWeight     float64 // Posts from user's circles
	DiversityPenalty float64 // Penalty for similar consecutive posts
}

// DefaultWeights returns the default ranking weights
func DefaultWeights() RankingWeights {
	return RankingWeights{
		RecencyWeight:    0.3,
		UrgencyWeight:    0.25,
		EngagementWeight: 0.2,
		CategoryWeight:   0.15,
		CircleWeight:     0.1,
		DiversityPenalty: 0.1,
	}
}

// NewFeedRanker creates a new feed ranker with default weights
func NewFeedRanker() *FeedRanker {
	return &FeedRanker{
		weights: DefaultWeights(),
	}
}

// RankedPost represents a post with its calculated score
type RankedPost struct {
	Post  *domain.Post
	Score float64
}

// RankPosts scores and ranks posts based on user preferences and engagement
func (r *FeedRanker) RankPosts(ctx context.Context, posts []*domain.Post, userPrefs *UserPreferences) []*RankedPost {
	if len(posts) == 0 {
		return []*RankedPost{}
	}

	ranked := make([]*RankedPost, len(posts))
	now := time.Now()

	for i, post := range posts {
		score := r.calculateScore(post, userPrefs, now)
		ranked[i] = &RankedPost{
			Post:  post,
			Score: score,
		}
	}

	// Sort by score (bubble sort for simplicity, can be optimized)
	for i := 0; i < len(ranked); i++ {
		for j := i + 1; j < len(ranked); j++ {
			if ranked[j].Score > ranked[i].Score {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}

	// Apply diversity penalty to consecutive similar posts
	r.applyDiversityPenalty(ranked)

	// Re-sort after diversity penalty
	for i := 0; i < len(ranked); i++ {
		for j := i + 1; j < len(ranked); j++ {
			if ranked[j].Score > ranked[i].Score {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}

	return ranked
}

// calculateScore computes the ranking score for a single post
func (r *FeedRanker) calculateScore(post *domain.Post, prefs *UserPreferences, now time.Time) float64 {
	score := 0.0

	// 1. Recency score (exponential decay)
	recencyScore := r.calculateRecencyScore(post.CreatedAt, now)
	score += recencyScore * r.weights.RecencyWeight

	// 2. Urgency score (normalized urgency level)
	urgencyScore := float64(post.UrgencyLevel) / 10.0
	score += urgencyScore * r.weights.UrgencyWeight

	// 3. Engagement score (responses + support reactions)
	engagementScore := r.calculateEngagementScore(post.ResponseCount, post.SupportCount)
	score += engagementScore * r.weights.EngagementWeight

	// 4. Category match score
	if prefs != nil {
		categoryScore := r.calculateCategoryScore(post.Categories, prefs.PreferredCategories)
		score += categoryScore * r.weights.CategoryWeight

		// 5. Circle affinity score
		circleScore := r.calculateCircleScore(post.CircleID, prefs.UserCircles)
		score += circleScore * r.weights.CircleWeight
	}

	return score
}

// calculateRecencyScore uses exponential decay for post age
func (r *FeedRanker) calculateRecencyScore(createdAt, now time.Time) float64 {
	ageHours := now.Sub(createdAt).Hours()

	// Exponential decay with half-life of 24 hours
	// score = e^(-ln(2) * age / half_life)
	halfLife := 24.0
	score := math.Exp(-math.Ln2 * ageHours / halfLife)

	return score
}

// calculateEngagementScore combines responses and support reactions
func (r *FeedRanker) calculateEngagementScore(responses, support int) float64 {
	// Logarithmic scaling to prevent viral posts from dominating
	engagementCount := float64(responses*2 + support) // Weight responses higher
	if engagementCount == 0 {
		return 0
	}

	// log(1 + x) normalized to [0, 1] range
	// Assume max engagement of 100 for normalization
	maxEngagement := 100.0
	score := math.Log1p(engagementCount) / math.Log1p(maxEngagement)

	return math.Min(score, 1.0)
}

// calculateCategoryScore measures overlap with user preferences
func (r *FeedRanker) calculateCategoryScore(postCategories, preferredCategories []string) float64 {
	if len(preferredCategories) == 0 {
		return 0.5 // Neutral score if no preferences
	}

	matches := 0
	for _, postCat := range postCategories {
		for _, prefCat := range preferredCategories {
			if postCat == prefCat {
				matches++
				break
			}
		}
	}

	// Jaccard similarity
	totalCategories := len(postCategories) + len(preferredCategories) - matches
	if totalCategories == 0 {
		return 0
	}

	return float64(matches) / float64(totalCategories)
}

// calculateCircleScore boosts posts from user's circles
func (r *FeedRanker) calculateCircleScore(postCircleID *string, userCircles []string) float64 {
	if postCircleID == nil || len(userCircles) == 0 {
		return 0
	}

	for _, circleID := range userCircles {
		if circleID == *postCircleID {
			return 1.0 // Full score for circle match
		}
	}

	return 0
}

// applyDiversityPenalty penalizes consecutive posts from same category/user
func (r *FeedRanker) applyDiversityPenalty(ranked []*RankedPost) {
	if len(ranked) < 2 {
		return
	}

	for i := 1; i < len(ranked); i++ {
		current := ranked[i]
		previous := ranked[i-1]

		// Penalize if same user
		if current.Post.UserID == previous.Post.UserID {
			current.Score *= (1.0 - r.weights.DiversityPenalty)
		}

		// Penalize if same primary category
		if len(current.Post.Categories) > 0 && len(previous.Post.Categories) > 0 {
			if current.Post.Categories[0] == previous.Post.Categories[0] {
				current.Score *= (1.0 - r.weights.DiversityPenalty*0.5)
			}
		}
	}
}

// UserPreferences represents user's feed preferences
type UserPreferences struct {
	PreferredCategories []string
	UserCircles         []string
	BlockedUsers        []string
	PreferredTimeOfDay  string
}

// FilterPosts removes blocked users and applies basic filters
func FilterPosts(posts []*domain.Post, prefs *UserPreferences) []*domain.Post {
	if prefs == nil {
		return posts
	}

	filtered := make([]*domain.Post, 0, len(posts))
	for _, post := range posts {
		// Skip blocked users
		blocked := false
		for _, blockedID := range prefs.BlockedUsers {
			if post.UserID == blockedID {
				blocked = true
				break
			}
		}

		if !blocked {
			filtered = append(filtered, post)
		}
	}

	return filtered
}
