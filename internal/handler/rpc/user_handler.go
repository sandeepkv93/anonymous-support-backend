package rpc

import (
	"context"

	"connectrpc.com/connect"
	userv1 "github.com/yourorg/anonymous-support/gen/user/v1"
	"github.com/yourorg/anonymous-support/internal/service"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHandler struct {
	userService service.UserServiceInterface
}

func NewUserHandler(userService service.UserServiceInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetProfile(
	ctx context.Context,
	req *connect.Request[userv1.GetProfileRequest],
) (*connect.Response[userv1.GetProfileResponse], error) {
	user, err := h.userService.GetProfile(ctx, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	res := connect.NewResponse(&userv1.GetProfileResponse{
		Profile: &userv1.UserProfile{
			Id:             user.ID.String(),
			Username:       user.Username,
			AvatarId:       int32(user.AvatarID),
			CreatedAt:      timestamppb.New(user.CreatedAt),
			LastActiveAt:   timestamppb.New(user.LastActiveAt),
			IsAnonymous:    user.IsAnonymous,
			IsPremium:      user.IsPremium,
			StrengthPoints: int32(user.StrengthPoints),
		},
	})

	return res, nil
}

func (h *UserHandler) UpdateProfile(
	ctx context.Context,
	req *connect.Request[userv1.UpdateProfileRequest],
) (*connect.Response[userv1.UpdateProfileResponse], error) {
	var username *string
	var avatarID *int

	if req.Msg.Username != nil {
		username = req.Msg.Username
	}

	if req.Msg.AvatarId != nil {
		aid := int(*req.Msg.AvatarId)
		avatarID = &aid
	}

	err := h.userService.UpdateProfile(ctx, req.Msg.UserId, username, avatarID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	res := connect.NewResponse(&userv1.UpdateProfileResponse{
		Success: true,
	})

	return res, nil
}

func (h *UserHandler) GetStreak(
	ctx context.Context,
	req *connect.Request[userv1.GetStreakRequest],
) (*connect.Response[userv1.GetStreakResponse], error) {
	tracker, err := h.userService.GetStreak(ctx, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	response := &userv1.GetStreakResponse{
		StreakDays:       int32(tracker.StreakDays),
		TotalCravings:    int32(tracker.TotalCravings),
		CravingsResisted: int32(tracker.CravingsResisted),
	}

	if tracker.LastRelapseDate != nil {
		response.LastRelapseDate = timestamppb.New(*tracker.LastRelapseDate)
	}

	res := connect.NewResponse(response)
	return res, nil
}

func (h *UserHandler) UpdateStreak(
	ctx context.Context,
	req *connect.Request[userv1.UpdateStreakRequest],
) (*connect.Response[userv1.UpdateStreakResponse], error) {
	newStreak, err := h.userService.UpdateStreak(ctx, req.Msg.UserId, req.Msg.HadRelapse)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&userv1.UpdateStreakResponse{
		Success:   true,
		NewStreak: int32(newStreak),
	})

	return res, nil
}
