package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ResponseType string

const (
	ResponseTypeQuick ResponseType = "quick"
	ResponseTypeText  ResponseType = "text"
	ResponseTypeVoice ResponseType = "voice"
)

type SupportResponse struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PostID         string             `bson:"post_id" json:"post_id"`
	UserID         string             `bson:"user_id" json:"user_id"`
	Username       string             `bson:"username" json:"username"`
	Type           ResponseType       `bson:"type" json:"type"`
	Content        string             `bson:"content" json:"content"`
	VoiceNoteURL   *string            `bson:"voice_note_url,omitempty" json:"voice_note_url,omitempty"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	StrengthPoints int                `bson:"strength_points" json:"strength_points"`
}

type UserTracker struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID               string             `bson:"user_id" json:"user_id"`
	StreakDays           int                `bson:"streak_days" json:"streak_days"`
	LongestStreak        int                `bson:"longest_streak" json:"longest_streak"`
	TotalDaysClean       int                `bson:"total_days_clean" json:"total_days_clean"`
	TotalRelapses        int                `bson:"total_relapses" json:"total_relapses"`
	LastRelapseDate      *time.Time         `bson:"last_relapse_date,omitempty" json:"last_relapse_date,omitempty"`
	TotalCravings        int                `bson:"total_cravings" json:"total_cravings"`
	CravingsResisted     int                `bson:"cravings_resisted" json:"cravings_resisted"`
	SupportGiven         int                `bson:"support_given" json:"support_given"`
	SupportReceived      int                `bson:"support_received" json:"support_received"`
	VulnerabilityPattern map[string]int     `bson:"vulnerability_pattern" json:"vulnerability_pattern"`
	Categories           []string           `bson:"categories" json:"categories"`
	Goals                []Goal             `bson:"goals" json:"goals"`
	Milestones           []Milestone        `bson:"milestones" json:"milestones"`
	UpdatedAt            time.Time          `bson:"updated_at" json:"updated_at"`
}

type Goal struct {
	Description string    `bson:"description" json:"description"`
	TargetDays  int       `bson:"target_days" json:"target_days"`
	StartDate   time.Time `bson:"start_date" json:"start_date"`
	IsAchieved  bool      `bson:"is_achieved" json:"is_achieved"`
}

type Milestone struct {
	Name        string    `bson:"name" json:"name"`
	Days        int       `bson:"days" json:"days"`
	AchievedAt  time.Time `bson:"achieved_at" json:"achieved_at"`
	Description string    `bson:"description" json:"description"`
}
