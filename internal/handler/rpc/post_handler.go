package rpc

import (
	"context"

	"connectrpc.com/connect"
	postv1 "github.com/yourorg/anonymous-support/gen/post/v1"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/middleware"
	"github.com/yourorg/anonymous-support/internal/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PostHandler struct {
	postService *service.PostService
}

func NewPostHandler(postService *service.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

func (h *PostHandler) CreatePost(
	ctx context.Context,
	req *connect.Request[postv1.CreatePostRequest],
) (*connect.Response[postv1.CreatePostResponse], error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	username, ok := middleware.GetUsername(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	postType := mapProtoPostTypeToDomain(req.Msg.Type)

	var circleID *string
	if req.Msg.CircleId != nil {
		circleID = req.Msg.CircleId
	}

	post, err := h.postService.CreatePost(
		ctx,
		userID,
		username,
		postType,
		req.Msg.Content,
		req.Msg.Categories,
		int(req.Msg.UrgencyLevel),
		req.Msg.TimeContext,
		int(req.Msg.DaysSinceRelapse),
		req.Msg.Tags,
		req.Msg.Visibility,
		circleID,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	res := connect.NewResponse(&postv1.CreatePostResponse{
		PostId:    post.ID.Hex(),
		CreatedAt: timestamppb.New(post.CreatedAt),
	})

	return res, nil
}

func (h *PostHandler) GetPost(
	ctx context.Context,
	req *connect.Request[postv1.GetPostRequest],
) (*connect.Response[postv1.GetPostResponse], error) {
	post, err := h.postService.GetPost(ctx, req.Msg.PostId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	res := connect.NewResponse(&postv1.GetPostResponse{
		Post: mapDomainPostToProto(post),
	})

	return res, nil
}

func (h *PostHandler) GetFeed(
	ctx context.Context,
	req *connect.Request[postv1.GetFeedRequest],
) (*connect.Response[postv1.GetFeedResponse], error) {
	var circleID *string
	if req.Msg.CircleId != nil {
		circleID = req.Msg.CircleId
	}

	var postType *domain.PostType
	if req.Msg.TypeFilter != nil {
		pt := mapProtoPostTypeToDomain(*req.Msg.TypeFilter)
		postType = &pt
	}

	posts, err := h.postService.GetFeed(
		ctx,
		req.Msg.Categories,
		circleID,
		postType,
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

	res := connect.NewResponse(&postv1.GetFeedResponse{
		Posts:      protoPosts,
		TotalCount: int32(len(protoPosts)),
	})

	return res, nil
}

func (h *PostHandler) DeletePost(
	ctx context.Context,
	req *connect.Request[postv1.DeletePostRequest],
) (*connect.Response[postv1.DeletePostResponse], error) {
	// Auth check
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	// Ownership verification is done in service layer
	err := h.postService.DeletePost(ctx, req.Msg.PostId, userID)
	if err != nil {
		// Service already checks ownership - return appropriate error
		return nil, connect.NewError(connect.CodePermissionDenied, err)
	}

	res := connect.NewResponse(&postv1.DeletePostResponse{
		Success: true,
	})

	return res, nil
}

func (h *PostHandler) UpdatePostUrgency(
	ctx context.Context,
	req *connect.Request[postv1.UpdatePostUrgencyRequest],
) (*connect.Response[postv1.UpdatePostUrgencyResponse], error) {
	_, ok := middleware.GetUserID(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, nil)
	}

	err := h.postService.UpdatePostUrgency(ctx, req.Msg.PostId, int(req.Msg.UrgencyLevel))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&postv1.UpdatePostUrgencyResponse{
		Success: true,
	})

	return res, nil
}

func mapProtoPostTypeToDomain(pt postv1.PostType) domain.PostType {
	switch pt {
	case postv1.PostType_POST_TYPE_SOS:
		return domain.PostTypeSOS
	case postv1.PostType_POST_TYPE_CHECK_IN:
		return domain.PostTypeCheckIn
	case postv1.PostType_POST_TYPE_VICTORY:
		return domain.PostTypeVictory
	case postv1.PostType_POST_TYPE_QUESTION:
		return domain.PostTypeQuestion
	default:
		return domain.PostTypeCheckIn
	}
}

func mapDomainPostTypeToProto(pt domain.PostType) postv1.PostType {
	switch pt {
	case domain.PostTypeSOS:
		return postv1.PostType_POST_TYPE_SOS
	case domain.PostTypeCheckIn:
		return postv1.PostType_POST_TYPE_CHECK_IN
	case domain.PostTypeVictory:
		return postv1.PostType_POST_TYPE_VICTORY
	case domain.PostTypeQuestion:
		return postv1.PostType_POST_TYPE_QUESTION
	default:
		return postv1.PostType_POST_TYPE_UNSPECIFIED
	}
}

func mapDomainPostToProto(post *domain.Post) *postv1.Post {
	return &postv1.Post{
		Id:            post.ID.Hex(),
		UserId:        post.UserID,
		Username:      post.Username,
		Type:          mapDomainPostTypeToProto(post.Type),
		Content:       post.Content,
		Categories:    post.Categories,
		UrgencyLevel:  int32(post.UrgencyLevel),
		ResponseCount: int32(post.ResponseCount),
		SupportCount:  int32(post.SupportCount),
		CreatedAt:     timestamppb.New(post.CreatedAt),
		Context: &postv1.PostContext{
			DaysSinceRelapse: int32(post.Context.DaysSinceRelapse),
			TimeContext:      post.Context.TimeContext,
			Tags:             post.Context.Tags,
		},
	}
}
