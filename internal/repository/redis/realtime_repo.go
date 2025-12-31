package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RealtimeRepository struct {
	client *redis.Client
}

func NewRealtimeRepository(client *redis.Client) *RealtimeRepository {
	return &RealtimeRepository{client: client}
}

func (r *RealtimeRepository) PublishNewPost(ctx context.Context, postID, postType string, categories []string) error {
	data := map[string]interface{}{
		"post_id":    postID,
		"type":       postType,
		"categories": categories,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return r.client.Publish(ctx, "channel:post:new", payload).Err()
}

func (r *RealtimeRepository) PublishNewResponse(ctx context.Context, postID, responseID string) error {
	data := map[string]interface{}{
		"post_id":     postID,
		"response_id": responseID,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	channel := fmt.Sprintf("channel:post:%s:response", postID)
	return r.client.Publish(ctx, channel, payload).Err()
}

func (r *RealtimeRepository) AddSupporterToPost(ctx context.Context, postID, userID string) error {
	key := fmt.Sprintf("post:supporters:%s", postID)
	return r.client.SAdd(ctx, key, userID).Err()
}

func (r *RealtimeRepository) GetSupporterCount(ctx context.Context, postID string) (int64, error) {
	key := fmt.Sprintf("post:supporters:%s", postID)
	return r.client.SCard(ctx, key).Result()
}

func (r *RealtimeRepository) IncrementViewCount(ctx context.Context, postID string) error {
	key := fmt.Sprintf("post:view_count:%s", postID)
	return r.client.Incr(ctx, key).Err()
}

func (r *RealtimeRepository) GetViewCount(ctx context.Context, postID string) (int64, error) {
	key := fmt.Sprintf("post:view_count:%s", postID)
	count, err := r.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

func (r *RealtimeRepository) AddToFeed(ctx context.Context, feedKey, postID string, score float64) error {
	return r.client.ZAdd(ctx, feedKey, redis.Z{
		Score:  score,
		Member: postID,
	}).Err()
}

func (r *RealtimeRepository) GetFeed(ctx context.Context, userID string, limit int) ([]string, error) {
	feedKey := fmt.Sprintf("feed:%s", userID)
	return r.client.ZRevRange(ctx, feedKey, 0, int64(limit-1)).Result()
}

func (r *RealtimeRepository) CheckRateLimit(ctx context.Context, userID, action string, limit int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("ratelimit:%s:%s", action, userID)

	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		r.client.Expire(ctx, key, window)
	}

	return count <= int64(limit), nil
}

func (r *RealtimeRepository) AddSupporter(ctx context.Context, postID, userID string) error {
	return r.AddSupporterToPost(ctx, postID, userID)
}
func (r *RealtimeRepository) GetSupporters(ctx context.Context, postID string) ([]string, error) {
	return []string{}, nil
}
func (r *RealtimeRepository) PublishNotification(ctx context.Context, channel string, message interface{}) error {
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return r.client.Publish(ctx, channel, payload).Err()
}
func (r *RealtimeRepository) SubscribeToChannel(ctx context.Context, channel string) error {
	return nil
}
