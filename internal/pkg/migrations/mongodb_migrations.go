package migrations

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetMongoMigrations returns all MongoDB migrations
func GetMongoMigrations() []Migration {
	return []Migration{
		{
			Version:     1,
			Description: "Create posts collection with indexes",
			Up:          createPostsCollection,
			Down:        dropPostsCollection,
		},
		{
			Version:     2,
			Description: "Create support_responses collection with indexes",
			Up:          createSupportResponsesCollection,
			Down:        dropSupportResponsesCollection,
		},
		{
			Version:     3,
			Description: "Create user_trackers collection with indexes",
			Up:          createUserTrackersCollection,
			Down:        dropUserTrackersCollection,
		},
		{
			Version:     4,
			Description: "Add TTL index on posts.expires_at",
			Up:          addPostsTTLIndex,
			Down:        removePostsTTLIndex,
		},
	}
}

// Migration 1: Create posts collection with indexes
func createPostsCollection(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("posts")

	// Create indexes
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("idx_user_id"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_created_at"),
		},
		{
			Keys:    bson.D{{Key: "type", Value: 1}},
			Options: options.Index().SetName("idx_type"),
		},
		{
			Keys:    bson.D{{Key: "categories", Value: 1}},
			Options: options.Index().SetName("idx_categories"),
		},
		{
			Keys:    bson.D{{Key: "circle_id", Value: 1}},
			Options: options.Index().SetName("idx_circle_id"),
		},
		{
			Keys:    bson.D{{Key: "is_moderated", Value: 1}},
			Options: options.Index().SetName("idx_is_moderated"),
		},
		{
			Keys: bson.D{
				{Key: "created_at", Value: -1},
				{Key: "urgency_level", Value: -1},
			},
			Options: options.Index().SetName("idx_feed_sorting"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func dropPostsCollection(ctx context.Context, db *mongo.Database) error {
	return db.Collection("posts").Drop(ctx)
}

// Migration 2: Create support_responses collection with indexes
func createSupportResponsesCollection(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("support_responses")

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "post_id", Value: 1}},
			Options: options.Index().SetName("idx_post_id"),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("idx_user_id"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: -1}},
			Options: options.Index().SetName("idx_created_at"),
		},
		{
			Keys: bson.D{
				{Key: "post_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("idx_post_responses"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func dropSupportResponsesCollection(ctx context.Context, db *mongo.Database) error {
	return db.Collection("support_responses").Drop(ctx)
}

// Migration 3: Create user_trackers collection with indexes
func createUserTrackersCollection(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("user_trackers")

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("idx_user_id_unique"),
		},
		{
			Keys:    bson.D{{Key: "last_updated", Value: -1}},
			Options: options.Index().SetName("idx_last_updated"),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	return err
}

func dropUserTrackersCollection(ctx context.Context, db *mongo.Database) error {
	return db.Collection("user_trackers").Drop(ctx)
}

// Migration 4: Add TTL index on posts.expires_at
func addPostsTTLIndex(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("posts")

	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().
			SetName("idx_expires_at_ttl").
			SetExpireAfterSeconds(0), // Delete documents when expires_at is reached
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	return err
}

func removePostsTTLIndex(ctx context.Context, db *mongo.Database) error {
	collection := db.Collection("posts")
	_, err := collection.Indexes().DropOne(ctx, "idx_expires_at_ttl")
	return err
}
