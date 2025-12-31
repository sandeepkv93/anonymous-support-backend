package rpc

import (
	"context"

	"connectrpc.com/connect"
	moderationv1 "github.com/yourorg/anonymous-support/gen/moderation/v1"
	"github.com/yourorg/anonymous-support/internal/middleware"
	"github.com/yourorg/anonymous-support/internal/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ModerationHandler struct {
	moderationService *service.ModerationService
}

func NewModerationHandler(moderationService *service.ModerationService) *ModerationHandler {
	return &ModerationHandler{
		moderationService: moderationService,
	}
}

func (h *ModerationHandler) ReportContent(
	ctx context.Context,
	req *connect.Request[moderationv1.ReportContentRequest],
) (*connect.Response[moderationv1.ReportContentResponse], error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	reportID, err := h.moderationService.ReportContent(
		ctx,
		userID,
		req.Msg.ContentType,
		req.Msg.ContentId,
		req.Msg.Reason,
		req.Msg.Description,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	res := connect.NewResponse(&moderationv1.ReportContentResponse{
		ReportId: reportID,
	})

	return res, nil
}

func (h *ModerationHandler) GetReports(
	ctx context.Context,
	req *connect.Request[moderationv1.GetReportsRequest],
) (*connect.Response[moderationv1.GetReportsResponse], error) {
	var status *string
	if req.Msg.Status != nil {
		status = req.Msg.Status
	}

	reports, err := h.moderationService.GetReports(
		ctx,
		status,
		int(req.Msg.Limit),
		int(req.Msg.Offset),
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoReports := make([]*moderationv1.Report, len(reports))
	for i, report := range reports {
		protoReports[i] = &moderationv1.Report{
			Id:          report.ID.String(),
			ReporterId:  report.ReporterID.String(),
			ContentType: report.ContentType,
			ContentId:   report.ContentID,
			Reason:      report.Reason,
			Description: report.Description,
			Status:      report.Status,
			CreatedAt:   timestamppb.New(report.CreatedAt),
		}
	}

	res := connect.NewResponse(&moderationv1.GetReportsResponse{
		Reports:    protoReports,
		TotalCount: int32(len(protoReports)),
	})

	return res, nil
}

func (h *ModerationHandler) ModerateContent(
	ctx context.Context,
	req *connect.Request[moderationv1.ModerateContentRequest],
) (*connect.Response[moderationv1.ModerateContentResponse], error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	err := h.moderationService.ModerateContent(
		ctx,
		req.Msg.ReportId,
		userID,
		req.Msg.Action,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	res := connect.NewResponse(&moderationv1.ModerateContentResponse{
		Success: true,
	})

	return res, nil
}
