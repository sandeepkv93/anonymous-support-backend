package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PostType string

const (
	PostTypeSOS      PostType = "sos"
	PostTypeCheckIn  PostType = "check_in"
	PostTypeVictory  PostType = "victory"
	PostTypeQuestion PostType = "question"
)

type Post struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          string             `bson:"user_id" json:"user_id"`
	Username        string             `bson:"username" json:"username"`
	Type            PostType           `bson:"type" json:"type"`
	Content         string             `bson:"content" json:"content"`
	Categories      []string           `bson:"categories" json:"categories"`
	UrgencyLevel    int                `bson:"urgency_level" json:"urgency_level"`
	Context         PostContext        `bson:"context" json:"context"`
	Visibility      string             `bson:"visibility" json:"visibility"`
	CircleID        *string            `bson:"circle_id,omitempty" json:"circle_id,omitempty"`
	ResponseCount   int                `bson:"response_count" json:"response_count"`
	SupportCount    int                `bson:"support_count" json:"support_count"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	ExpiresAt       *time.Time         `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	IsModerated     bool               `bson:"is_moderated" json:"is_moderated"`
	ModerationFlags []string           `bson:"moderation_flags,omitempty" json:"moderation_flags,omitempty"`
}

type PostContext struct {
	DaysSinceRelapse int      `bson:"days_since_relapse" json:"days_since_relapse"`
	TimeContext      string   `bson:"time_context" json:"time_context"`
	Tags             []string `bson:"tags" json:"tags"`
}
