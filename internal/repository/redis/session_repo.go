package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	client *redis.Client
}

func NewSessionRepository(client *redis.Client) *SessionRepository {
	return &SessionRepository{client: client}
}

// hashToken creates a SHA-256 hash of the token for secure storage
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// StoreRefreshToken stores a refresh token with rotation support
// Each user can have multiple active tokens (token family)
func (r *SessionRepository) StoreRefreshToken(ctx context.Context, userID, token string, expiry time.Duration) error {
	tokenHash := hashToken(token)

	// Store token hash in a set for this user
	tokenKey := fmt.Sprintf("user:session:%s:tokens", userID)
	pipe := r.client.Pipeline()

	// Add token to user's active tokens set
	pipe.SAdd(ctx, tokenKey, tokenHash)
	pipe.Expire(ctx, tokenKey, expiry)

	// Store token metadata (issued time, expiry)
	metaKey := fmt.Sprintf("user:session:%s:token:%s", userID, tokenHash)
	pipe.HSet(ctx, metaKey, map[string]interface{}{
		"issued_at": time.Now().Unix(),
		"expires_at": time.Now().Add(expiry).Unix(),
	})
	pipe.Expire(ctx, metaKey, expiry)

	_, err := pipe.Exec(ctx)
	return err
}

// ValidateRefreshToken checks if a token is valid and not revoked
func (r *SessionRepository) ValidateRefreshToken(ctx context.Context, userID, token string) (bool, error) {
	tokenHash := hashToken(token)
	tokenKey := fmt.Sprintf("user:session:%s:tokens", userID)

	// Check if token exists in user's active tokens
	exists, err := r.client.SIsMember(ctx, tokenKey, tokenHash).Result()
	if err != nil {
		return false, err
	}

	if !exists {
		return false, nil
	}

	// Check token metadata
	metaKey := fmt.Sprintf("user:session:%s:token:%s", userID, tokenHash)
	meta, err := r.client.HGetAll(ctx, metaKey).Result()
	if err != nil {
		return false, err
	}

	if len(meta) == 0 {
		return false, nil
	}

	return true, nil
}

// RevokeRefreshToken revokes a specific token
func (r *SessionRepository) RevokeRefreshToken(ctx context.Context, userID, token string) error {
	tokenHash := hashToken(token)
	tokenKey := fmt.Sprintf("user:session:%s:tokens", userID)
	metaKey := fmt.Sprintf("user:session:%s:token:%s", userID, tokenHash)

	pipe := r.client.Pipeline()
	pipe.SRem(ctx, tokenKey, tokenHash)
	pipe.Del(ctx, metaKey)

	_, err := pipe.Exec(ctx)
	return err
}

// RevokeAllRefreshTokens revokes all tokens for a user (e.g., on logout or security breach)
func (r *SessionRepository) RevokeAllRefreshTokens(ctx context.Context, userID string) error {
	tokenKey := fmt.Sprintf("user:session:%s:tokens", userID)

	// Get all token hashes
	tokenHashes, err := r.client.SMembers(ctx, tokenKey).Result()
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()

	// Delete all token metadata
	for _, tokenHash := range tokenHashes {
		metaKey := fmt.Sprintf("user:session:%s:token:%s", userID, tokenHash)
		pipe.Del(ctx, metaKey)
	}

	// Delete the tokens set
	pipe.Del(ctx, tokenKey)

	_, err = pipe.Exec(ctx)
	return err
}

// GetRefreshToken returns the active refresh token for a user (backward compatibility)
// This is deprecated in favor of token rotation
func (r *SessionRepository) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("user:session:%s", userID)
	return r.client.Get(ctx, key).Result()
}

// DeleteRefreshToken deletes the session (backward compatibility)
func (r *SessionRepository) DeleteRefreshToken(ctx context.Context, userID string) error {
	return r.RevokeAllRefreshTokens(ctx, userID)
}

func (r *SessionRepository) SetUserOnline(ctx context.Context, userID string, ttl time.Duration) error {
	key := fmt.Sprintf("user:online:%s", userID)
	return r.client.Set(ctx, key, "1", ttl).Err()
}

func (r *SessionRepository) IsUserOnline(ctx context.Context, userID string) (bool, error) {
	key := fmt.Sprintf("user:online:%s", userID)
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}
