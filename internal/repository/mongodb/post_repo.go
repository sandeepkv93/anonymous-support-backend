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

type PostRepository struct {
	collection *mongo.Collection
}

func NewPostRepository(db *mongo.Database) *PostRepository {
	return &PostRepository{
		collection: db.Collection("posts"),
	}
}

func (r *PostRepository) Create(ctx context.Context, post *domain.Post) error {
	post.ID = primitive.NewObjectID()
	post.CreatedAt = time.Now()
	post.ResponseCount = 0
	post.SupportCount = 0

	if post.ExpiresAt == nil {
		expiresAt := time.Now().Add(30 * 24 * time.Hour)
		post.ExpiresAt = &expiresAt
	}

	_, err := r.collection.InsertOne(ctx, post)
	return err
}

func (r *PostRepository) GetByID(ctx context.Context, id string) (*domain.Post, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid post ID")
	}

	var post domain.Post
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&post)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("post not found")
	}
	return &post, err
}

func (r *PostRepository) GetFeed(ctx context.Context, categories []string, circleID *string, postType *domain.PostType, limit, offset int) ([]*domain.Post, error) {
	filter := bson.M{"is_moderated": false}

	if len(categories) > 0 {
		filter["categories"] = bson.M{"$in": categories}
	}

	if circleID != nil {
		filter["circle_id"] = *circleID
	} else {
		filter["visibility"] = "public"
	}

	if postType != nil {
		filter["type"] = *postType
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	posts := []*domain.Post{}
	if err := cursor.All(ctx, &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *PostRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid post ID")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("post not found")
	}
	return nil
}

func (r *PostRepository) UpdateUrgency(ctx context.Context, id string, urgencyLevel int32) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid post ID")
	}

	update := bson.M{"$set": bson.M{"urgency_level": urgencyLevel}}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("post not found")
	}
	return nil
}

func (r *PostRepository) IncrementResponseCount(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid post ID")
	}

	update := bson.M{"$inc": bson.M{"response_count": 1}}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *PostRepository) IncrementSupportCount(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid post ID")
	}

	update := bson.M{"$inc": bson.M{"support_count": 1}}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *PostRepository) FlagForModeration(ctx context.Context, id string, flags []string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid post ID")
	}

	update := bson.M{
		"$set": bson.M{
			"is_moderated":     true,
			"moderation_flags": flags,
		},
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}
