package rpc

import (
	"context"

	"connectrpc.com/connect"
	authv1 "github.com/yourorg/anonymous-support/gen/auth/v1"
	"github.com/yourorg/anonymous-support/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) RegisterAnonymous(
	ctx context.Context,
	req *connect.Request[authv1.RegisterAnonymousRequest],
) (*connect.Response[authv1.RegisterAnonymousResponse], error) {
	user, accessToken, refreshToken, err := h.authService.RegisterAnonymous(
		ctx,
		req.Msg.Username,
		int(req.Msg.AvatarId),
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	res := connect.NewResponse(&authv1.RegisterAnonymousResponse{
		UserId:       user.ID.String(),
		Username:     user.Username,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})

	return res, nil
}

func (h *AuthHandler) RegisterWithEmail(
	ctx context.Context,
	req *connect.Request[authv1.RegisterWithEmailRequest],
) (*connect.Response[authv1.RegisterWithEmailResponse], error) {
	user, accessToken, refreshToken, err := h.authService.RegisterWithEmail(
		ctx,
		req.Msg.Username,
		req.Msg.Email,
		req.Msg.Password,
		int(req.Msg.AvatarId),
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	res := connect.NewResponse(&authv1.RegisterWithEmailResponse{
		UserId:       user.ID.String(),
		Username:     user.Username,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})

	return res, nil
}

func (h *AuthHandler) Login(
	ctx context.Context,
	req *connect.Request[authv1.LoginRequest],
) (*connect.Response[authv1.LoginResponse], error) {
	user, accessToken, refreshToken, err := h.authService.Login(
		ctx,
		req.Msg.Username,
		req.Msg.Password,
	)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	res := connect.NewResponse(&authv1.LoginResponse{
		UserId:       user.ID.String(),
		Username:     user.Username,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})

	return res, nil
}

func (h *AuthHandler) RefreshToken(
	ctx context.Context,
	req *connect.Request[authv1.RefreshTokenRequest],
) (*connect.Response[authv1.RefreshTokenResponse], error) {
	accessToken, err := h.authService.RefreshToken(ctx, req.Msg.RefreshToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	res := connect.NewResponse(&authv1.RefreshTokenResponse{
		AccessToken: accessToken,
	})

	return res, nil
}

func (h *AuthHandler) Logout(
	ctx context.Context,
	req *connect.Request[authv1.LogoutRequest],
) (*connect.Response[authv1.LogoutResponse], error) {
	err := h.authService.Logout(ctx, req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&authv1.LogoutResponse{
		Success: true,
	})

	return res, nil
}
