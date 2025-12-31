package websocket

import (
	"encoding/json"
	"fmt"
	"time"
)

// MessageVersion represents the WebSocket message schema version
type MessageVersion string

const (
	MessageV1 MessageVersion = "v1"
	MessageV2 MessageVersion = "v2" // Future version
)

// CurrentMessageVersion is the current schema version
const CurrentMessageVersion = MessageV1

// BaseMessage is the base structure for all WebSocket messages
type BaseMessage struct {
	Version   MessageVersion `json:"version"`
	Type      string         `json:"type"`
	Timestamp time.Time      `json:"timestamp"`
}

// EventMessage represents an event notification
type EventMessage struct {
	BaseMessage
	EventType string                 `json:"eventType"`
	Data      map[string]interface{} `json:"data"`
}

// NewPostEvent represents a new post creation event
type NewPostEvent struct {
	BaseMessage
	EventType string `json:"eventType"` // "new_post"
	PostID    string `json:"postId"`
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	Type      string `json:"postType"`
	Content   string `json:"content"`
	Category  string `json:"category"`
	Urgency   int32  `json:"urgency"`
}

// NewResponseEvent represents a new response event
type NewResponseEvent struct {
	BaseMessage
	EventType  string `json:"eventType"` // "new_response"
	ResponseID string `json:"responseId"`
	PostID     string `json:"postId"`
	UserID     string `json:"userId"`
	Username   string `json:"username"`
	Type       string `json:"responseType"`
}

// SupporterJoinedEvent represents a quick support event
type SupporterJoinedEvent struct {
	BaseMessage
	EventType string `json:"eventType"` // "supporter_joined"
	PostID    string `json:"postId"`
	UserID    string `json:"userId"`
	Username  string `json:"username"`
}

// PostDeletedEvent represents a post deletion event
type PostDeletedEvent struct {
	BaseMessage
	EventType string `json:"eventType"` // "post_deleted"
	PostID    string `json:"postId"`
}

// MessageEncoder handles encoding messages with proper versioning
type MessageEncoder struct {
	version MessageVersion
}

// NewMessageEncoder creates a new message encoder
func NewMessageEncoder(version MessageVersion) *MessageEncoder {
	return &MessageEncoder{version: version}
}

// Encode encodes a message with the appropriate version
func (e *MessageEncoder) Encode(message interface{}) ([]byte, error) {
	// Set version and timestamp on base message
	switch msg := message.(type) {
	case *NewPostEvent:
		msg.Version = e.version
		msg.Timestamp = time.Now()
	case *NewResponseEvent:
		msg.Version = e.version
		msg.Timestamp = time.Now()
	case *SupporterJoinedEvent:
		msg.Version = e.version
		msg.Timestamp = time.Now()
	case *PostDeletedEvent:
		msg.Version = e.version
		msg.Timestamp = time.Now()
	case *EventMessage:
		msg.Version = e.version
		msg.Timestamp = time.Now()
	}

	return json.Marshal(message)
}

// MessageDecoder handles decoding messages and version negotiation
type MessageDecoder struct {
	supportedVersions []MessageVersion
}

// NewMessageDecoder creates a new message decoder
func NewMessageDecoder() *MessageDecoder {
	return &MessageDecoder{
		supportedVersions: []MessageVersion{MessageV1},
	}
}

// Decode decodes a message and validates version
func (d *MessageDecoder) Decode(data []byte, target interface{}) error {
	// First, decode base message to check version
	var base BaseMessage
	if err := json.Unmarshal(data, &base); err != nil {
		return fmt.Errorf("failed to decode base message: %w", err)
	}

	// Validate version is supported
	if !d.isVersionSupported(base.Version) {
		return fmt.Errorf("unsupported message version: %s", base.Version)
	}

	// Decode full message
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to decode message: %w", err)
	}

	return nil
}

// isVersionSupported checks if a version is supported
func (d *MessageDecoder) isVersionSupported(version MessageVersion) bool {
	for _, v := range d.supportedVersions {
		if v == version {
			return true
		}
	}
	return false
}

// MigrateMessage migrates a message from one version to another
func MigrateMessage(from MessageVersion, to MessageVersion, data []byte) ([]byte, error) {
	if from == to {
		return data, nil
	}

	// Future: Implement version-specific migration logic
	// For now, only v1 exists
	switch {
	case from == MessageV1 && to == MessageV2:
		// Migrate from v1 to v2
		return nil, fmt.Errorf("migration from v1 to v2 not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported migration: %s -> %s", from, to)
	}
}

// CreateNewPostEvent creates a properly versioned new post event
func CreateNewPostEvent(postID, userID, username, postType, content, category string, urgency int32) *NewPostEvent {
	return &NewPostEvent{
		BaseMessage: BaseMessage{
			Version:   CurrentMessageVersion,
			Type:      "event",
			Timestamp: time.Now(),
		},
		EventType: "new_post",
		PostID:    postID,
		UserID:    userID,
		Username:  username,
		Type:      postType,
		Content:   content,
		Category:  category,
		Urgency:   urgency,
	}
}

// CreateNewResponseEvent creates a properly versioned new response event
func CreateNewResponseEvent(responseID, postID, userID, username, responseType string) *NewResponseEvent {
	return &NewResponseEvent{
		BaseMessage: BaseMessage{
			Version:   CurrentMessageVersion,
			Type:      "event",
			Timestamp: time.Now(),
		},
		EventType:  "new_response",
		ResponseID: responseID,
		PostID:     postID,
		UserID:     userID,
		Username:   username,
		Type:       responseType,
	}
}
