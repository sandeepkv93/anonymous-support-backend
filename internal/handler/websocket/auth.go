package websocket

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourorg/anonymous-support/internal/pkg/jwt"
	"go.uber.org/zap"
)

// AuthMessage represents an authentication message from the client
type AuthMessage struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

// SubscribeMessage represents a channel subscription request
type SubscribeMessage struct {
	Type     string   `json:"type"`
	Channels []string `json:"channels"`
}

// AuthorizeConnection validates the WebSocket connection and returns the user ID
func (h *Hub) AuthorizeConnection(client *Client, authMsg *AuthMessage) error {
	// Validate token
	claims, err := h.jwtManager.ValidateAccessToken(authMsg.Token)
	if err != nil {
		h.logger.Warn("WebSocket auth failed", zap.Error(err))
		return fmt.Errorf("invalid authentication token")
	}

	// Parse user ID from claims
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return fmt.Errorf("invalid user ID in token")
	}

	// Store user info in client
	client.UserID = &userID
	client.Username = claims.Username
	client.IsAuthenticated = true

	h.logger.Info("WebSocket client authenticated",
		zap.String("user_id", userID.String()),
		zap.String("username", claims.Username))

	return nil
}

// AuthorizeChannelSubscription checks if a user can subscribe to a channel
func (h *Hub) AuthorizeChannelSubscription(ctx context.Context, client *Client, channel string) error {
	if !client.IsAuthenticated {
		return fmt.Errorf("client not authenticated")
	}

	// Parse channel type
	switch {
	case channel == "posts":
		// All authenticated users can subscribe to global posts
		return nil

	case channel == "responses":
		// All authenticated users can subscribe to responses
		return nil

	case len(channel) > 7 && channel[:7] == "circle:":
		// Circle-specific channel: circle:{circleID}
		circleID := channel[7:]
		// Verify user is a member of the circle
		return h.verifyCircleMembership(ctx, client, circleID)

	case len(channel) > 5 && channel[:5] == "user:":
		// User-specific channel: user:{userID}
		userID := channel[5:]
		// Only allow users to subscribe to their own channel
		if client.UserID.String() != userID {
			return fmt.Errorf("cannot subscribe to another user's private channel")
		}
		return nil

	case len(channel) > 5 && channel[:5] == "post:":
		// Post-specific channel: post:{postID}
		// All authenticated users can subscribe to post updates
		return nil

	default:
		return fmt.Errorf("unknown channel type: %s", channel)
	}
}

// verifyCircleMembership checks if a user is a member of a circle
func (h *Hub) verifyCircleMembership(ctx context.Context, client *Client, circleID string) error {
	// Parse circle ID
	cID, err := uuid.Parse(circleID)
	if err != nil {
		return fmt.Errorf("invalid circle ID")
	}

	// Check membership via repository (assumes circle repo is available)
	// In real implementation, inject CircleRepository into Hub
	// For now, we'll assume all subscriptions are allowed
	// TODO: Implement actual membership check
	h.logger.Debug("Circle membership check",
		zap.String("circle_id", cID.String()),
		zap.String("user_id", client.UserID.String()))

	return nil
}

// HandleClientMessage processes incoming messages from the client
func (h *Hub) HandleClientMessage(client *Client, message []byte) error {
	var baseMsg struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal(message, &baseMsg); err != nil {
		return fmt.Errorf("invalid message format")
	}

	switch baseMsg.Type {
	case "auth":
		var authMsg AuthMessage
		if err := json.Unmarshal(message, &authMsg); err != nil {
			return fmt.Errorf("invalid auth message")
		}
		return h.AuthorizeConnection(client, &authMsg)

	case "subscribe":
		if !client.IsAuthenticated {
			return fmt.Errorf("must authenticate before subscribing")
		}

		var subMsg SubscribeMessage
		if err := json.Unmarshal(message, &subMsg); err != nil {
			return fmt.Errorf("invalid subscribe message")
		}

		// Authorize and subscribe to each channel
		for _, channel := range subMsg.Channels {
			if err := h.AuthorizeChannelSubscription(context.Background(), client, channel); err != nil {
				h.logger.Warn("Channel subscription denied",
					zap.String("channel", channel),
					zap.String("user_id", client.UserID.String()),
					zap.Error(err))
				continue
			}

			client.Channels[channel] = true
			h.logger.Info("Client subscribed to channel",
				zap.String("channel", channel),
				zap.String("user_id", client.UserID.String()))
		}

		return nil

	case "unsubscribe":
		// Handle unsubscribe
		var subMsg SubscribeMessage
		if err := json.Unmarshal(message, &subMsg); err != nil {
			return fmt.Errorf("invalid unsubscribe message")
		}

		for _, channel := range subMsg.Channels {
			delete(client.Channels, channel)
			h.logger.Info("Client unsubscribed from channel",
				zap.String("channel", channel),
				zap.String("user_id", client.UserID.String()))
		}

		return nil

	default:
		return fmt.Errorf("unknown message type: %s", baseMsg.Type)
	}
}
