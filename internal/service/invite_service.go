package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/repository"
)

type InviteService struct {
	inviteRepo repository.InviteRepository
	circleRepo repository.CircleRepository
}

func NewInviteService(inviteRepo repository.InviteRepository, circleRepo repository.CircleRepository) *InviteService {
	return &InviteService{
		inviteRepo: inviteRepo,
		circleRepo: circleRepo,
	}
}

// CreateInvite generates a new invite for a circle
func (s *InviteService) CreateInvite(ctx context.Context, circleID, createdBy string, maxUses int, expiresIn time.Duration) (*domain.Invite, error) {
	// Verify user has permission to create invites (is circle owner/admin)
	circleUUID, err := uuid.Parse(circleID)
	if err != nil {
		return nil, fmt.Errorf("invalid circle ID")
	}

	circle, err := s.circleRepo.GetByID(ctx, circleUUID)
	if err != nil {
		return nil, fmt.Errorf("circle not found")
	}

	createdByUUID, _ := uuid.Parse(createdBy)
	if circle.CreatedBy != createdByUUID {
		return nil, fmt.Errorf("only circle owner can create invites")
	}

	// Generate invite code
	code, err := generateInviteCode()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(expiresIn)

	invite := &domain.Invite{
		ID:        uuid.New(),
		CircleID:  circleUUID,
		Code:      code,
		CreatedBy: createdByUUID,
		MaxUses:   maxUses,
		UsedCount: 0,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	if err := s.inviteRepo.Create(ctx, invite); err != nil {
		return nil, err
	}

	return invite, nil
}

// AcceptInvite joins a circle using an invite code
func (s *InviteService) AcceptInvite(ctx context.Context, code, userID string) (*domain.Circle, error) {
	// Get invite
	invite, err := s.inviteRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("invalid invite code")
	}

	// Validate invite
	if !invite.IsActive {
		return nil, fmt.Errorf("invite is inactive")
	}

	if time.Now().After(invite.ExpiresAt) {
		return nil, fmt.Errorf("invite has expired")
	}

	if invite.MaxUses > 0 && invite.UsedCount >= invite.MaxUses {
		return nil, fmt.Errorf("invite has reached max uses")
	}

	// Get circle
	circle, err := s.circleRepo.GetByID(ctx, invite.CircleID)
	if err != nil {
		return nil, err
	}

	// Join circle (this should be delegated to CircleService)
	// For now, just increment the used count
	if err := s.inviteRepo.IncrementUsedCount(ctx, invite.ID); err != nil {
		return nil, err
	}

	return circle, nil
}

// RevokeInvite deactivates an invite
func (s *InviteService) RevokeInvite(ctx context.Context, inviteID, userID string) error {
	inviteUUID, err := uuid.Parse(inviteID)
	if err != nil {
		return fmt.Errorf("invalid invite ID")
	}

	invite, err := s.inviteRepo.GetByID(ctx, inviteUUID)
	if err != nil {
		return err
	}

	// Verify user has permission
	circle, err := s.circleRepo.GetByID(ctx, invite.CircleID)
	if err != nil {
		return err
	}

	userUUID, _ := uuid.Parse(userID)
	if circle.CreatedBy != userUUID {
		return fmt.Errorf("only circle owner can revoke invites")
	}

	return s.inviteRepo.Deactivate(ctx, inviteUUID)
}

// GetCircleInvites returns all active invites for a circle
func (s *InviteService) GetCircleInvites(ctx context.Context, circleID, userID string) ([]*domain.Invite, error) {
	circleUUID, err := uuid.Parse(circleID)
	if err != nil {
		return nil, fmt.Errorf("invalid circle ID")
	}

	// Verify permission
	circle, err := s.circleRepo.GetByID(ctx, circleUUID)
	if err != nil {
		return nil, err
	}

	userUUID, _ := uuid.Parse(userID)
	if circle.CreatedBy != userUUID {
		return nil, fmt.Errorf("only circle owner can view invites")
	}

	return s.inviteRepo.GetByCircleID(ctx, circleUUID)
}

// generateInviteCode creates a random invite code
func generateInviteCode() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
