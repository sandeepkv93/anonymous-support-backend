package service

import (
	"context"
	"fmt"
)

type NotificationService struct {
}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

func (s *NotificationService) SendNotification(ctx context.Context, userID, title, body string) error {
	fmt.Printf("Sending notification to %s: %s - %s\n", userID, title, body)
	return nil
}

func (s *NotificationService) NotifyNewResponse(ctx context.Context, postAuthorID, responderUsername string) error {
	return s.SendNotification(ctx, postAuthorID, "New Response", fmt.Sprintf("%s responded to your post", responderUsername))
}

func (s *NotificationService) NotifyNewSupport(ctx context.Context, postAuthorID string, supportCount int) error {
	return s.SendNotification(ctx, postAuthorID, "New Support", fmt.Sprintf("%d people are supporting you", supportCount))
}
