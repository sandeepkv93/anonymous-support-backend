package redis

import (
	"context"
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

func (r *SessionRepository) StoreRefreshToken(ctx context.Context, userID, token string, expiry time.Duration) error {
	key := fmt.Sprintf("user:session:%s", userID)
	return r.client.Set(ctx, key, token, expiry).Err()
}

func (r *SessionRepository) GetRefreshToken(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("user:session:%s", userID)
	return r.client.Get(ctx, key).Result()
}

func (r *SessionRepository) DeleteRefreshToken(ctx context.Context, userID string) error {
	key := fmt.Sprintf("user:session:%s", userID)
	return r.client.Del(ctx, key).Err()
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
