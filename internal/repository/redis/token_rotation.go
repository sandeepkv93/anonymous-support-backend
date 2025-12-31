package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenRotation handles refresh token rotation and reuse detection
type TokenRotation struct {
	client *redis.Client
}

// NewTokenRotation creates a new token rotation manager
func NewTokenRotation(client *redis.Client) *TokenRotation {
	return &TokenRotation{client: client}
}

// StoreRefreshTokenWithRotation stores a refresh token with rotation tracking
func (r *TokenRotation) StoreRefreshTokenWithRotation(ctx context.Context, userID, tokenID, token string, expiry time.Duration) error {
	// Store the current token
	tokenKey := fmt.Sprintf("refresh_token:%s:%s", userID, tokenID)
	if err := r.client.Set(ctx, tokenKey, token, expiry).Err(); err != nil {
		return err
	}

	// Track the latest token ID for this user
	latestKey := fmt.Sprintf("refresh_token:latest:%s", userID)
	if err := r.client.Set(ctx, latestKey, tokenID, expiry).Err(); err != nil {
		return err
	}

	// Add to the token family (for revocation on reuse detection)
	familyKey := fmt.Sprintf("refresh_token:family:%s", userID)
	if err := r.client.SAdd(ctx, familyKey, tokenID).Err(); err != nil {
		return err
	}
	// Set expiry on the family set
	if err := r.client.Expire(ctx, familyKey, expiry).Err(); err != nil {
		return err
	}

	return nil
}

// ValidateAndRotateToken validates a refresh token and generates a new one
// Returns: newTokenID, isValid, error
func (r *TokenRotation) ValidateAndRotateToken(ctx context.Context, userID, tokenID string) (string, bool, error) {
	// Get the latest token ID for this user
	latestKey := fmt.Sprintf("refresh_token:latest:%s", userID)
	latestTokenID, err := r.client.Get(ctx, latestKey).Result()
	if err == redis.Nil {
		// No refresh token found - user needs to login again
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}

	// Check if the provided token is the latest one
	if tokenID != latestTokenID {
		// Token reuse detected - revoke all tokens for this user
		if err := r.RevokeAllTokens(ctx, userID); err != nil {
			return "", false, fmt.Errorf("failed to revoke tokens after reuse detection: %w", err)
		}
		return "", false, fmt.Errorf("token reuse detected - all tokens revoked")
	}

	// Token is valid - generate new token ID for rotation
	newTokenID := fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())
	return newTokenID, true, nil
}

// RevokeAllTokens revokes all refresh tokens for a user (used on reuse detection or logout)
func (r *TokenRotation) RevokeAllTokens(ctx context.Context, userID string) error {
	familyKey := fmt.Sprintf("refresh_token:family:%s", userID)
	latestKey := fmt.Sprintf("refresh_token:latest:%s", userID)

	// Get all token IDs in the family
	tokenIDs, err := r.client.SMembers(ctx, familyKey).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	// Delete each token
	for _, tokenID := range tokenIDs {
		tokenKey := fmt.Sprintf("refresh_token:%s:%s", userID, tokenID)
		if err := r.client.Del(ctx, tokenKey).Err(); err != nil {
			return err
		}
	}

	// Delete the family set and latest key
	if err := r.client.Del(ctx, familyKey, latestKey).Err(); err != nil {
		return err
	}

	return nil
}

// GetRefreshToken retrieves a refresh token by user ID and token ID
func (r *TokenRotation) GetRefreshToken(ctx context.Context, userID, tokenID string) (string, error) {
	tokenKey := fmt.Sprintf("refresh_token:%s:%s", userID, tokenID)
	return r.client.Get(ctx, tokenKey).Result()
}

// MarkTokenAsUsed marks a token as used (for one-time use enforcement)
func (r *TokenRotation) MarkTokenAsUsed(ctx context.Context, userID, tokenID string) error {
	usedKey := fmt.Sprintf("refresh_token:used:%s:%s", userID, tokenID)
	// Mark as used with a short TTL (enough time to complete the rotation)
	return r.client.Set(ctx, usedKey, "1", 60*time.Second).Err()
}

// IsTokenUsed checks if a token has already been used
func (r *TokenRotation) IsTokenUsed(ctx context.Context, userID, tokenID string) (bool, error) {
	usedKey := fmt.Sprintf("refresh_token:used:%s:%s", userID, tokenID)
	result, err := r.client.Exists(ctx, usedKey).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}
