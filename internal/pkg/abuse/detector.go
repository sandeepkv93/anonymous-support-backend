package abuse

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yourorg/anonymous-support/internal/domain"
)

// AbuseDetector detects and prevents abusive behavior
type AbuseDetector struct {
	spamThresholds SpamThresholds
	blocklist      map[string]bool
}

// SpamThresholds defines limits for spam detection
type SpamThresholds struct {
	MaxPostsPerHour     int
	MaxPostsPerDay      int
	MaxIdenticalPosts   int
	MinPostInterval     time.Duration
	MaxReportsPerDay    int
	MaxFailedLogins     int
}

// DefaultThresholds returns default spam detection thresholds
func DefaultThresholds() SpamThresholds {
	return SpamThresholds{
		MaxPostsPerHour:     10,
		MaxPostsPerDay:      50,
		MaxIdenticalPosts:   3,
		MinPostInterval:     30 * time.Second,
		MaxReportsPerDay:    20,
		MaxFailedLogins:     5,
	}
}

// NewAbuseDetector creates a new abuse detector
func NewAbuseDetector() *AbuseDetector {
	return &AbuseDetector{
		spamThresholds: DefaultThresholds(),
		blocklist:      make(map[string]bool),
	}
}

// DetectionResult contains abuse detection results
type DetectionResult struct {
	IsAbuse     bool
	Reason      string
	Severity    string // low, medium, high, critical
	Action      string // warn, throttle, block, ban
	Confidence  float64
}

// CheckPost checks a post for abusive content
func (d *AbuseDetector) CheckPost(ctx context.Context, post *domain.Post, userHistory *UserHistory) *DetectionResult {
	// Check for spam patterns
	if d.isSpam(post, userHistory) {
		return &DetectionResult{
			IsAbuse:    true,
			Reason:     "Spam detected",
			Severity:   "high",
			Action:     "throttle",
			Confidence: 0.9,
		}
	}

	// Check for prohibited content
	if d.containsProhibitedContent(post.Content) {
		return &DetectionResult{
			IsAbuse:    true,
			Reason:     "Prohibited content detected",
			Severity:   "critical",
			Action:     "block",
			Confidence: 0.95,
		}
	}

	// Check for excessive posting
	if userHistory != nil && userHistory.PostsLastHour > d.spamThresholds.MaxPostsPerHour {
		return &DetectionResult{
			IsAbuse:    true,
			Reason:     "Posting too frequently",
			Severity:   "medium",
			Action:     "throttle",
			Confidence: 0.8,
		}
	}

	return &DetectionResult{
		IsAbuse:    false,
		Confidence: 0.0,
	}
}

// isSpam checks for spam patterns
func (d *AbuseDetector) isSpam(post *domain.Post, history *UserHistory) bool {
	if history == nil {
		return false
	}

	// Check for identical posts
	if history.IdenticalPostCount >= d.spamThresholds.MaxIdenticalPosts {
		return true
	}

	// Check for rapid posting
	if history.LastPostTime != nil {
		timeSinceLastPost := time.Since(*history.LastPostTime)
		if timeSinceLastPost < d.spamThresholds.MinPostInterval {
			return true
		}
	}

	return false
}

// containsProhibitedContent checks for prohibited keywords and patterns
func (d *AbuseDetector) containsProhibitedContent(content string) bool {
	contentLower := strings.ToLower(content)

	// Prohibited patterns (drugs for sale, self-harm encouragement, etc.)
	prohibitedPatterns := []string{
		"buy drugs",
		"sell drugs",
		"dealer contact",
		"suicide method",
		"how to overdose",
		"selling prescription",
	}

	for _, pattern := range prohibitedPatterns {
		if strings.Contains(contentLower, pattern) {
			return true
		}
	}

	return false
}

// CheckUser checks a user's overall behavior for abuse
func (d *AbuseDetector) CheckUser(ctx context.Context, userID string, history *UserHistory) *DetectionResult {
	if history == nil {
		return &DetectionResult{IsAbuse: false}
	}

	// Check for excessive reporting
	if history.ReportsLastDay > d.spamThresholds.MaxReportsPerDay {
		return &DetectionResult{
			IsAbuse:    true,
			Reason:     "Excessive reporting (possible harassment)",
			Severity:   "high",
			Action:     "warn",
			Confidence: 0.85,
		}
	}

	// Check for mass posting
	if history.PostsLastDay > d.spamThresholds.MaxPostsPerDay {
		return &DetectionResult{
			IsAbuse:    true,
			Reason:     "Excessive posting",
			Severity:   "medium",
			Action:     "throttle",
			Confidence: 0.8,
		}
	}

	// Check if user is in blocklist
	if d.blocklist[userID] {
		return &DetectionResult{
			IsAbuse:    true,
			Reason:     "User is blocked",
			Severity:   "critical",
			Action:     "ban",
			Confidence: 1.0,
		}
	}

	return &DetectionResult{IsAbuse: false}
}

// BlockUser adds a user to the blocklist
func (d *AbuseDetector) BlockUser(userID string) {
	d.blocklist[userID] = true
}

// UnblockUser removes a user from the blocklist
func (d *AbuseDetector) UnblockUser(userID string) {
	delete(d.blocklist, userID)
}

// UserHistory tracks user activity for abuse detection
type UserHistory struct {
	PostsLastHour      int
	PostsLastDay       int
	ReportsLastDay     int
	IdenticalPostCount int
	LastPostTime       *time.Time
	LastPostContent    string
	FailedLoginCount   int
	AccountAge         time.Duration
}

// RateLimiter manages rate limiting for abuse prevention
type RateLimiter struct {
	limits map[string]*RateLimit
}

// RateLimit tracks rate limit state for a resource
type RateLimit struct {
	Count       int
	ResetAt     time.Time
	MaxRequests int
	Window      time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limits: make(map[string]*RateLimit),
	}
}

// CheckLimit checks if a request is within rate limits
func (r *RateLimiter) CheckLimit(key string, maxRequests int, window time.Duration) (bool, error) {
	limit, exists := r.limits[key]
	now := time.Now()

	if !exists || now.After(limit.ResetAt) {
		// Initialize new limit window
		r.limits[key] = &RateLimit{
			Count:       1,
			ResetAt:     now.Add(window),
			MaxRequests: maxRequests,
			Window:      window,
		}
		return true, nil
	}

	if limit.Count >= limit.MaxRequests {
		return false, fmt.Errorf("rate limit exceeded, resets at %s", limit.ResetAt.Format(time.RFC3339))
	}

	limit.Count++
	return true, nil
}
