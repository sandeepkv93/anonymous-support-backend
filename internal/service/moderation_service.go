package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/repository/postgres"
)

type ModerationService struct {
	modRepo *postgres.ModerationRepository
}

func NewModerationService(modRepo *postgres.ModerationRepository) *ModerationService {
	return &ModerationService{modRepo: modRepo}
}

func (s *ModerationService) ReportContent(ctx context.Context, reporterID, contentType, contentID, reason, description string) (string, error) {
	uid, err := uuid.Parse(reporterID)
	if err != nil {
		return "", err
	}

	report := &domain.ContentReport{
		ID:          uuid.New(),
		ReporterID:  uid,
		ContentType: contentType,
		ContentID:   contentID,
		Reason:      reason,
		Description: description,
		Status:      "pending",
	}

	if err := s.modRepo.CreateReport(ctx, report); err != nil {
		return "", err
	}

	return report.ID.String(), nil
}

func (s *ModerationService) GetReports(ctx context.Context, status *string, limit, offset int) ([]*domain.ContentReport, error) {
	return s.modRepo.GetReports(ctx, status, limit, offset)
}

func (s *ModerationService) ModerateContent(ctx context.Context, reportID, reviewerID, action string) error {
	rid, err := uuid.Parse(reportID)
	if err != nil {
		return err
	}

	uid, err := uuid.Parse(reviewerID)
	if err != nil {
		return err
	}

	var status string
	switch action {
	case "approve":
		status = "dismissed"
	case "remove":
		status = "actioned"
	default:
		status = "reviewed"
	}

	return s.modRepo.UpdateReportStatus(ctx, rid, status, uid, "")
}
