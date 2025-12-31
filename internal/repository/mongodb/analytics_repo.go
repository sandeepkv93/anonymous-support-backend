package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/yourorg/anonymous-support/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AnalyticsRepository struct {
	trackers *mongo.Collection
}

func NewAnalyticsRepository(db *mongo.Database) *AnalyticsRepository {
	return &AnalyticsRepository{
		trackers: db.Collection("user_trackers"),
	}
}

func (r *AnalyticsRepository) GetTracker(ctx context.Context, userID string) (*domain.UserTracker, error) {
	var tracker domain.UserTracker
	err := r.trackers.FindOne(ctx, bson.M{"user_id": userID}).Decode(&tracker)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("tracker not found")
	}
	return &tracker, err
}

func (r *AnalyticsRepository) UpsertTracker(ctx context.Context, tracker *domain.UserTracker) error {
	if tracker.ID.IsZero() {
		tracker.ID = primitive.NewObjectID()
	}
	tracker.UpdatedAt = time.Now()

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"user_id": tracker.UserID}
	update := bson.M{"$set": tracker}

	_, err := r.trackers.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *AnalyticsRepository) UpdateStreak(ctx context.Context, userID string, hadRelapse bool) (int, error) {
	tracker, err := r.GetTracker(ctx, userID)
	if err != nil && err.Error() != "tracker not found" {
		return 0, err
	}

	if tracker == nil {
		tracker = &domain.UserTracker{
			UserID:               userID,
			StreakDays:           0,
			TotalCravings:        0,
			CravingsResisted:     0,
			VulnerabilityPattern: make(map[string]int),
			Categories:           []string{},
			Goals:                []domain.Goal{},
			Milestones:           []domain.Milestone{},
		}
	}

	if hadRelapse {
		now := time.Now()
		tracker.LastRelapseDate = &now
		tracker.StreakDays = 0
	} else {
		tracker.StreakDays++
	}

	if err := r.UpsertTracker(ctx, tracker); err != nil {
		return 0, err
	}

	return tracker.StreakDays, nil
}

func (r *AnalyticsRepository) IncrementCravings(ctx context.Context, userID string, resisted bool) error {
	filter := bson.M{"user_id": userID}
	update := bson.M{
		"$inc": bson.M{"total_cravings": 1},
	}

	if resisted {
		update["$inc"].(bson.M)["cravings_resisted"] = 1
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.trackers.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *AnalyticsRepository) AddMilestone(ctx context.Context, userID string, milestone domain.Milestone) error {
	return nil
}
