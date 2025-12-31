package dto

import (
	apperrors "github.com/yourorg/anonymous-support/internal/errors"
)

// ReportContentRequest represents a request to report content
type ReportContentRequest struct {
	ContentType string
	ContentID   string
	Reason      string
	Description string
}

// Validate validates the request
func (r *ReportContentRequest) Validate() error {
	if r.ContentType == "" {
		return apperrors.NewValidationError("Content type is required", nil)
	}
	if r.ContentType != "post" && r.ContentType != "response" && r.ContentType != "user" {
		return apperrors.NewValidationError("Invalid content type", nil)
	}
	if r.ContentID == "" {
		return apperrors.NewValidationError("Content ID is required", nil)
	}
	if r.Reason == "" {
		return apperrors.NewValidationError("Reason is required", nil)
	}
	if len(r.Description) > 1000 {
		return apperrors.NewValidationError("Description cannot exceed 1000 characters", nil)
	}
	return nil
}

// GetReportsRequest represents a request to get reports
type GetReportsRequest struct {
	Status string
	Limit  int32
	Offset int32
}

// Validate validates the request
func (r *GetReportsRequest) Validate() error {
	if r.Status != "" && r.Status != "pending" && r.Status != "reviewed" && r.Status != "dismissed" {
		return apperrors.NewValidationError("Invalid status", nil)
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

// ModerateContentRequest represents a request to moderate content
type ModerateContentRequest struct {
	ReportID string
	Action   string
	Notes    string
}

// Validate validates the request
func (r *ModerateContentRequest) Validate() error {
	if r.ReportID == "" {
		return apperrors.NewValidationError("Report ID is required", nil)
	}
	if r.Action == "" {
		return apperrors.NewValidationError("Action is required", nil)
	}
	if r.Action != "approve" && r.Action != "remove" && r.Action != "warn" && r.Action != "ban" {
		return apperrors.NewValidationError("Invalid action", nil)
	}
	if len(r.Notes) > 1000 {
		return apperrors.NewValidationError("Notes cannot exceed 1000 characters", nil)
	}
	return nil
}

// ReportDTO represents a content report
type ReportDTO struct {
	ID          string
	ContentType string
	ContentID   string
	ReporterID  string
	Reason      string
	Description string
	Status      string
	CreatedAt   string
	ReviewedAt  string
	ReviewedBy  string
	Notes       string
}

// ReportsListResponse represents a list of reports
type ReportsListResponse struct {
	Reports    []*ReportDTO
	TotalCount int32
	HasMore    bool
}
