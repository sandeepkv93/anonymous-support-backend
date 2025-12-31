package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/repository/mongodb"
	"github.com/yourorg/anonymous-support/internal/repository/postgres"
)

type CircleService struct {
	circleRepo *postgres.CircleRepository
	postRepo   *mongodb.PostRepository
}

func NewCircleService(
	circleRepo *postgres.CircleRepository,
	postRepo *mongodb.PostRepository,
) *CircleService {
	return &CircleService{
		circleRepo: circleRepo,
		postRepo:   postRepo,
	}
}

func (s *CircleService) CreateCircle(ctx context.Context, userID, name, description, category string, maxMembers int, isPrivate bool) (string, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return "", err
	}

	circle := &domain.Circle{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Category:    category,
		MaxMembers:  maxMembers,
		MemberCount: 0,
		IsPrivate:   isPrivate,
		CreatedBy:   uid,
	}

	if err := s.circleRepo.Create(ctx, circle); err != nil {
		return "", err
	}

	membership := &domain.CircleMembership{
		ID:       uuid.New(),
		CircleID: circle.ID,
		UserID:   uid,
		Role:     "moderator",
	}

	if err := s.circleRepo.JoinCircle(ctx, membership); err != nil {
		return "", err
	}

	return circle.ID.String(), nil
}

func (s *CircleService) JoinCircle(ctx context.Context, userID, circleID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	cid, err := uuid.Parse(circleID)
	if err != nil {
		return err
	}

	circle, err := s.circleRepo.GetByID(ctx, cid)
	if err != nil {
		return err
	}

	if circle.MemberCount >= circle.MaxMembers {
		return fmt.Errorf("circle is full")
	}

	membership := &domain.CircleMembership{
		ID:       uuid.New(),
		CircleID: cid,
		UserID:   uid,
		Role:     "member",
	}

	return s.circleRepo.JoinCircle(ctx, membership)
}

func (s *CircleService) LeaveCircle(ctx context.Context, userID, circleID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	cid, err := uuid.Parse(circleID)
	if err != nil {
		return err
	}

	return s.circleRepo.LeaveCircle(ctx, cid, uid)
}

func (s *CircleService) GetCircleMembers(ctx context.Context, circleID string, limit, offset int) ([]*domain.CircleMembership, error) {
	cid, err := uuid.Parse(circleID)
	if err != nil {
		return nil, err
	}

	return s.circleRepo.GetMembers(ctx, cid, limit, offset)
}

func (s *CircleService) GetCircleFeed(ctx context.Context, circleID string, limit, offset int) ([]*domain.Post, error) {
	return s.postRepo.GetFeed(ctx, nil, &circleID, nil, limit, offset)
}

func (s *CircleService) GetCircles(ctx context.Context, category *string, limit, offset int) ([]*domain.Circle, error) {
	return s.circleRepo.List(ctx, category, limit, offset)
}
