package dto

import (
	"github.com/yourorg/anonymous-support/internal/domain"
	apperrors "github.com/yourorg/anonymous-support/internal/errors"
	"github.com/yourorg/anonymous-support/internal/pkg/validator"
)

// CreateResponseRequest represents a request to create a support response
type CreateResponseRequest struct {
	PostID       string
	Type         domain.ResponseType
	Content      string
	VoiceNoteURL string
}

// Validate validates the request
func (r *CreateResponseRequest) Validate() error {
	if r.PostID == "" {
		return apperrors.NewValidationError("Post ID is required", nil)
	}

	// Validate response type
	switch r.Type {
	case domain.ResponseTypeQuick, domain.ResponseTypeText, domain.ResponseTypeVoice:
		// valid types
	default:
		return apperrors.NewValidationError("Invalid response type", nil)
	}

	// For text responses, validate content
	if r.Type == domain.ResponseTypeText {
		if err := validator.ValidateResponseContent(r.Content); err != nil {
			return apperrors.NewValidationError("Invalid response content", err)
		}
	}

	// For voice responses, voice note URL is required
	if r.Type == domain.ResponseTypeVoice && r.VoiceNoteURL == "" {
		return apperrors.NewValidationError("Voice note URL is required for voice responses", nil)
	}

	return nil
}

// GetResponsesRequest represents a request to get responses
type GetResponsesRequest struct {
	PostID string
	Limit  int32
	Offset int32
}

// Validate validates the request
func (r *GetResponsesRequest) Validate() error {
	if r.PostID == "" {
		return apperrors.NewValidationError("Post ID is required", nil)
	}
	if r.Limit <= 0 {
		r.Limit = 20 // default
	}
	if r.Limit > 100 {
		return apperrors.NewValidationError("Limit cannot exceed 100", nil)
	}
	if r.Offset < 0 {
		return apperrors.NewValidationError("Offset cannot be negative", nil)
	}
	return nil
}

// SupportResponseDTO represents support response data for responses
type SupportResponseDTO struct {
	ID             string
	PostID         string
	UserID         string
	Username       string
	Type           string
	Content        string
	VoiceNoteURL   string
	StrengthPoints int32
	CreatedAt      string
}

// NewSupportResponseDTO creates a SupportResponseDTO from a domain.SupportResponse
func NewSupportResponseDTO(response *domain.SupportResponse) *SupportResponseDTO {
	voiceNoteURL := ""
	if response.VoiceNoteURL != nil {
		voiceNoteURL = *response.VoiceNoteURL
	}
	return &SupportResponseDTO{
		ID:             response.ID.Hex(),
		PostID:         response.PostID,
		UserID:         response.UserID,
		Username:       response.Username,
		Type:           string(response.Type),
		Content:        response.Content,
		VoiceNoteURL:   voiceNoteURL,
		StrengthPoints: int32(response.StrengthPoints),
		CreatedAt:      response.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ResponsesListResponse represents a list of responses
type ResponsesListResponse struct {
	Responses  []*SupportResponseDTO
	TotalCount int32
	HasMore    bool
}

// SupportStatsDTO represents support statistics
type SupportStatsDTO struct {
	TotalGiven    int32
	TotalReceived int32
	StreakDays    int32
	ImpactScore   int32
}
