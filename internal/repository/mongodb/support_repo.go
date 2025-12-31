package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/yourorg/anonymous-support/internal/domain"
)

type SupportRepository struct {
	responses *mongo.Collection
}

func NewSupportRepository(db *mongo.Database) *SupportRepository {
	return &SupportRepository{
		responses: db.Collection("support_responses"),
	}
}

func (r *SupportRepository) CreateResponse(ctx context.Context, response *domain.SupportResponse) error {
	response.ID = primitive.NewObjectID()
	response.CreatedAt = time.Now()

	_, err := r.responses.InsertOne(ctx, response)
	return err
}

func (r *SupportRepository) GetResponses(ctx context.Context, postID string, limit, offset int) ([]*domain.SupportResponse, error) {
	filter := bson.M{"post_id": postID}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := r.responses.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	responses := []*domain.SupportResponse{}
	if err := cursor.All(ctx, &responses); err != nil {
		return nil, err
	}

	return responses, nil
}

func (r *SupportRepository) GetResponseCount(ctx context.Context, postID string) (int64, error) {
	count, err := r.responses.CountDocuments(ctx, bson.M{"post_id": postID})
	return count, err
}

func (r *SupportRepository) GetUserStats(ctx context.Context, userID string) (given, received int64, err error) {
	given, err = r.responses.CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, 0, err
	}

	return given, 0, nil
}
