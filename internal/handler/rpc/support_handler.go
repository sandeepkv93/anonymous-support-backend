package rpc

import (
	"context"

	"connectrpc.com/connect"
	supportv1 "github.com/yourorg/anonymous-support/gen/support/v1"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/middleware"
	"github.com/yourorg/anonymous-support/internal/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SupportHandler struct {
	supportService *service.SupportService
}

func NewSupportHandler(supportService *service.SupportService) *SupportHandler {
	return &SupportHandler{
		supportService: supportService,
	}
}

func (h *SupportHandler) CreateResponse(
	ctx context.Context,
	req *connect.Request[supportv1.CreateResponseRequest],
) (*connect.Response[supportv1.CreateResponseResponse], error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	username, ok := middleware.GetUsername(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	responseType := mapProtoResponseTypeToDomain(req.Msg.Type)

	var voiceNoteURL *string
	if req.Msg.VoiceNoteUrl != nil {
		voiceNoteURL = req.Msg.VoiceNoteUrl
	}

	responseID, strengthPoints, err := h.supportService.CreateResponse(
		ctx,
		userID,
		username,
		req.Msg.PostId,
		responseType,
		req.Msg.Content,
		voiceNoteURL,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	res := connect.NewResponse(&supportv1.CreateResponseResponse{
		ResponseId:           responseID,
		StrengthPointsEarned: int32(strengthPoints),
	})

	return res, nil
}

func (h *SupportHandler) GetResponses(
	ctx context.Context,
	req *connect.Request[supportv1.GetResponsesRequest],
) (*connect.Response[supportv1.GetResponsesResponse], error) {
	responses, err := h.supportService.GetResponses(
		ctx,
		req.Msg.PostId,
		int(req.Msg.Limit),
		int(req.Msg.Offset),
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoResponses := make([]*supportv1.SupportResponse, len(responses))
	for i, resp := range responses {
		protoResponses[i] = &supportv1.SupportResponse{
			Id:        resp.ID.Hex(),
			PostId:    resp.PostID,
			UserId:    resp.UserID,
			Username:  resp.Username,
			Type:      mapDomainResponseTypeToProto(resp.Type),
			Content:   resp.Content,
			CreatedAt: timestamppb.New(resp.CreatedAt),
		}
	}

	res := connect.NewResponse(&supportv1.GetResponsesResponse{
		Responses:  protoResponses,
		TotalCount: int32(len(protoResponses)),
	})

	return res, nil
}

func (h *SupportHandler) QuickSupport(
	ctx context.Context,
	req *connect.Request[supportv1.QuickSupportRequest],
) (*connect.Response[supportv1.QuickSupportResponse], error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	totalSupporters, err := h.supportService.QuickSupport(
		ctx,
		userID,
		req.Msg.PostId,
		req.Msg.QuickMessageType,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&supportv1.QuickSupportResponse{
		Success:         true,
		TotalSupporters: int32(totalSupporters),
	})

	return res, nil
}

func (h *SupportHandler) GetSupportStats(
	ctx context.Context,
	req *connect.Request[supportv1.GetSupportStatsRequest],
) (*connect.Response[supportv1.GetSupportStatsResponse], error) {
	res := connect.NewResponse(&supportv1.GetSupportStatsResponse{
		TotalResponsesGiven:    0,
		TotalResponsesReceived: 0,
		StrengthPoints:         0,
		PeopleHelped:           0,
	})

	return res, nil
}

func mapProtoResponseTypeToDomain(rt supportv1.ResponseType) domain.ResponseType {
	switch rt {
	case supportv1.ResponseType_RESPONSE_TYPE_QUICK:
		return domain.ResponseTypeQuick
	case supportv1.ResponseType_RESPONSE_TYPE_TEXT:
		return domain.ResponseTypeText
	case supportv1.ResponseType_RESPONSE_TYPE_VOICE:
		return domain.ResponseTypeVoice
	default:
		return domain.ResponseTypeText
	}
}

func mapDomainResponseTypeToProto(rt domain.ResponseType) supportv1.ResponseType {
	switch rt {
	case domain.ResponseTypeQuick:
		return supportv1.ResponseType_RESPONSE_TYPE_QUICK
	case domain.ResponseTypeText:
		return supportv1.ResponseType_RESPONSE_TYPE_TEXT
	case domain.ResponseTypeVoice:
		return supportv1.ResponseType_RESPONSE_TYPE_VOICE
	default:
		return supportv1.ResponseType_RESPONSE_TYPE_UNSPECIFIED
	}
}
