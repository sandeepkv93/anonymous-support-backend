package websocket

import (
	"encoding/json"
	"time"
)

type WSMessageType string

const (
	WSMessageTypeNewPost         WSMessageType = "new_post"
	WSMessageTypeNewResponse     WSMessageType = "new_response"
	WSMessageTypeSupporterCount  WSMessageType = "supporter_count"
	WSMessageTypeNotification    WSMessageType = "notification"
	WSMessageTypeUserOnline      WSMessageType = "user_online"
	WSMessageTypeUserOffline     WSMessageType = "user_offline"
	WSMessageTypeTypingIndicator WSMessageType = "typing"
	WSMessageTypePing            WSMessageType = "ping"
	WSMessageTypePong            WSMessageType = "pong"
)

type WSMessage struct {
	Type      WSMessageType   `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
}


type SupporterCountEvent struct {
	PostID    string `json:"post_id"`
	Count     int    `json:"count"`
	OnlineNow int    `json:"online_now"`
}

type NotificationEvent struct {
	Title   string `json:"title"`
	Body    string `json:"body"`
	Action  string `json:"action"`
	Payload string `json:"payload"`
}
