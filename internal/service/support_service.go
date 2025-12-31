package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/pkg/validator"
	"github.com/yourorg/anonymous-support/internal/repository/mongodb"
	"github.com/yourorg/anonymous-support/internal/repository/postgres"
	"github.com/yourorg/anonymous-support/internal/repository/redis"
)

type SupportService struct {
	supportRepo  *mongodb.SupportRepository
	postRepo     *mongodb.PostRepository
	userRepo     *postgres.UserRepository
	realtimeRepo *redis.RealtimeRepository
}

func NewSupportService(
	supportRepo *mongodb.SupportRepository,
	postRepo *mongodb.PostRepository,
	userRepo *postgres.UserRepository,
	realtimeRepo *redis.RealtimeRepository,
) *SupportService {
	return &SupportService{
		supportRepo:  supportRepo,
		postRepo:     postRepo,
		userRepo:     userRepo,
		realtimeRepo: realtimeRepo,
	}
}

func (s *SupportService) CreateResponse(ctx context.Context, userID, username, postID string, responseType domain.ResponseType, content string, voiceNoteURL *string) (string, int, error) {
	if responseType == domain.ResponseTypeText {
		if err := validator.ValidateResponseContent(content); err != nil {
			return "", 0, err
		}
	}

	strengthPoints := s.calculateStrengthPoints(responseType, content)

	response := &domain.SupportResponse{
		PostID:         postID,
		UserID:         userID,
		Username:       username,
		Type:           responseType,
		Content:        content,
		VoiceNoteURL:   voiceNoteURL,
		StrengthPoints: strengthPoints,
	}

	if err := s.supportRepo.CreateResponse(ctx, response); err != nil {
		return "", 0, err
	}

	s.postRepo.IncrementResponseCount(ctx, postID)

	uid, _ := uuid.Parse(userID)
	s.userRepo.UpdateStrengthPoints(ctx, uid, strengthPoints)

	s.realtimeRepo.PublishNewResponse(ctx, postID, response.ID.Hex())

	return response.ID.Hex(), strengthPoints, nil
}

func (s *SupportService) GetResponses(ctx context.Context, postID string, limit, offset int) ([]*domain.SupportResponse, error) {
	return s.supportRepo.GetResponses(ctx, postID, limit, offset)
}

func (s *SupportService) QuickSupport(ctx context.Context, userID, postID, messageType string) (int, error) {
	s.realtimeRepo.AddSupporterToPost(ctx, postID, userID)
	s.postRepo.IncrementSupportCount(ctx, postID)

	count, err := s.realtimeRepo.GetSupporterCount(ctx, postID)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (s *SupportService) GetSupportStats(ctx context.Context, userID string) (given, received int64, strengthPoints, peopleHelped int, error error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	user, err := s.userRepo.GetByID(ctx, uid)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	given, received, err = s.supportRepo.GetUserStats(ctx, userID)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return given, received, user.StrengthPoints, int(given), nil
}

func (s *SupportService) calculateStrengthPoints(responseType domain.ResponseType, content string) int {
	switch responseType {
	case domain.ResponseTypeQuick:
		return 1
	case domain.ResponseTypeText:
		if len(content) > 100 {
			return 5
		}
		return 3
	case domain.ResponseTypeVoice:
		return 5
	default:
		return 1
	}
}
