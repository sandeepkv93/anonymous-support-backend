package rpc

import (
	"context"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	authv1 "github.com/yourorg/anonymous-support/gen/auth/v1"
	"github.com/yourorg/anonymous-support/internal/dto"
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
	authResp, err := h.authService.RegisterAnonymous(ctx, req.Msg.Username)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	res := connect.NewResponse(&authv1.RegisterAnonymousResponse{
		UserId:       authResp.User.ID,
		Username:     authResp.User.Username,
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
	})

	return res, nil
}

func (h *AuthHandler) RegisterWithEmail(
	ctx context.Context,
	req *connect.Request[authv1.RegisterWithEmailRequest],
) (*connect.Response[authv1.RegisterWithEmailResponse], error) {
	// Validate request
	registerReq := &dto.RegisterWithEmailRequest{
		Username: req.Msg.Username,
		Email:    req.Msg.Email,
		Password: req.Msg.Password,
	}
	if err := registerReq.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	authResp, err := h.authService.RegisterWithEmail(ctx, registerReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	res := connect.NewResponse(&authv1.RegisterWithEmailResponse{
		UserId:       authResp.User.ID,
		Username:     authResp.User.Username,
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
	})

	return res, nil
}

func (h *AuthHandler) Login(
	ctx context.Context,
	req *connect.Request[authv1.LoginRequest],
) (*connect.Response[authv1.LoginResponse], error) {
	// Validate request
	loginReq := &dto.LoginRequest{
		Email:    req.Msg.Username, // Username field used for email/username
		Password: req.Msg.Password,
	}
	if err := loginReq.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	authResp, err := h.authService.Login(ctx, loginReq)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	res := connect.NewResponse(&authv1.LoginResponse{
		UserId:       authResp.User.ID,
		Username:     authResp.User.Username,
		AccessToken:  authResp.AccessToken,
		RefreshToken: authResp.RefreshToken,
	})

	return res, nil
}

func (h *AuthHandler) RefreshToken(
	ctx context.Context,
	req *connect.Request[authv1.RefreshTokenRequest],
) (*connect.Response[authv1.RefreshTokenResponse], error) {
	authResp, err := h.authService.RefreshToken(ctx, req.Msg.RefreshToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	res := connect.NewResponse(&authv1.RefreshTokenResponse{
		AccessToken: authResp.AccessToken,
	})

	return res, nil
}

func (h *AuthHandler) Logout(
	ctx context.Context,
	req *connect.Request[authv1.LogoutRequest],
) (*connect.Response[authv1.LogoutResponse], error) {
	userID, err := uuid.Parse(req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	err = h.authService.Logout(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&authv1.LogoutResponse{
		Success: true,
	})

	return res, nil
}
