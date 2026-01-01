package rpc

import (
	"context"

	"connectrpc.com/connect"
	circlev1 "github.com/yourorg/anonymous-support/gen/circle/v1"
	postv1 "github.com/yourorg/anonymous-support/gen/post/v1"
	"github.com/yourorg/anonymous-support/internal/middleware"
	"github.com/yourorg/anonymous-support/internal/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CircleHandler struct {
	circleService service.CircleServiceInterface
}

func NewCircleHandler(circleService service.CircleServiceInterface) *CircleHandler {
	return &CircleHandler{
		circleService: circleService,
	}
}

func (h *CircleHandler) CreateCircle(
	ctx context.Context,
	req *connect.Request[circlev1.CreateCircleRequest],
) (*connect.Response[circlev1.CreateCircleResponse], error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	circleID, err := h.circleService.CreateCircle(
		ctx,
		userID,
		req.Msg.Name,
		req.Msg.Description,
		req.Msg.Category,
		int(req.Msg.MaxMembers),
		req.Msg.IsPrivate,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	res := connect.NewResponse(&circlev1.CreateCircleResponse{
		CircleId: circleID,
	})

	return res, nil
}

func (h *CircleHandler) JoinCircle(
	ctx context.Context,
	req *connect.Request[circlev1.JoinCircleRequest],
) (*connect.Response[circlev1.JoinCircleResponse], error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	err := h.circleService.JoinCircle(ctx, userID, req.Msg.CircleId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	res := connect.NewResponse(&circlev1.JoinCircleResponse{
		Success: true,
	})

	return res, nil
}

func (h *CircleHandler) LeaveCircle(
	ctx context.Context,
	req *connect.Request[circlev1.LeaveCircleRequest],
) (*connect.Response[circlev1.LeaveCircleResponse], error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	err := h.circleService.LeaveCircle(ctx, userID, req.Msg.CircleId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	res := connect.NewResponse(&circlev1.LeaveCircleResponse{
		Success: true,
	})

	return res, nil
}

func (h *CircleHandler) GetCircleMembers(
	ctx context.Context,
	req *connect.Request[circlev1.GetCircleMembersRequest],
) (*connect.Response[circlev1.GetCircleMembersResponse], error) {
	members, err := h.circleService.GetCircleMembers(
		ctx,
		req.Msg.CircleId,
		int(req.Msg.Limit),
		int(req.Msg.Offset),
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoMembers := make([]*circlev1.CircleMember, len(members))
	for i, member := range members {
		protoMembers[i] = &circlev1.CircleMember{
			UserId:   member.UserID.String(),
			Username: "", // TODO: fetch username
			AvatarId: 0,  // TODO: fetch avatar
			JoinedAt: timestamppb.New(member.JoinedAt),
			Role:     member.Role,
		}
	}

	res := connect.NewResponse(&circlev1.GetCircleMembersResponse{
		Members:    protoMembers,
		TotalCount: int32(len(protoMembers)), //nolint:gosec // Member count won't overflow int32
	})

	return res, nil
}

func (h *CircleHandler) GetCircleFeed(
	ctx context.Context,
	req *connect.Request[circlev1.GetCircleFeedRequest],
) (*connect.Response[circlev1.GetCircleFeedResponse], error) {
	posts, err := h.circleService.GetCircleFeed(
		ctx,
		req.Msg.CircleId,
		int(req.Msg.Limit),
		int(req.Msg.Offset),
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoPosts := make([]*postv1.Post, len(posts))
	for i, post := range posts {
		protoPosts[i] = mapDomainPostToProto(post)
	}

	res := connect.NewResponse(&circlev1.GetCircleFeedResponse{
		Posts:      protoPosts,
		TotalCount: int32(len(protoPosts)), //nolint:gosec // Post count won't overflow int32
	})

	return res, nil
}

func (h *CircleHandler) GetCircles(
	ctx context.Context,
	req *connect.Request[circlev1.GetCirclesRequest],
) (*connect.Response[circlev1.GetCirclesResponse], error) {
	var category *string
	if req.Msg.Category != nil {
		category = req.Msg.Category
	}

	circles, err := h.circleService.GetCircles(
		ctx,
		category,
		int(req.Msg.Limit),
		int(req.Msg.Offset),
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	protoCircles := make([]*circlev1.Circle, len(circles))
	for i, circle := range circles {
		protoCircles[i] = &circlev1.Circle{
			Id:          circle.ID.String(),
			Name:        circle.Name,
			Description: circle.Description,
			Category:    circle.Category,
			MaxMembers:  int32(circle.MaxMembers),  //nolint:gosec // Member limits won't overflow int32
			MemberCount: int32(circle.MemberCount), //nolint:gosec // Member count won't overflow int32
			IsPrivate:   circle.IsPrivate,
			CreatedAt:   timestamppb.New(circle.CreatedAt),
		}
	}

	res := connect.NewResponse(&circlev1.GetCirclesResponse{
		Circles:    protoCircles,
		TotalCount: int32(len(protoCircles)), //nolint:gosec // Circle count won't overflow int32
	})

	return res, nil
}
