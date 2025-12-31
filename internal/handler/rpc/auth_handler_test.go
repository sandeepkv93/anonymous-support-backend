package rpc

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	authv1 "github.com/yourorg/anonymous-support/gen/auth/v1"
	"github.com/yourorg/anonymous-support/internal/dto"
	"github.com/yourorg/anonymous-support/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) RegisterAnonymous(ctx context.Context, username string) (*dto.AuthResponse, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) RegisterWithEmail(ctx context.Context, req *dto.RegisterWithEmailRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func TestAuthHandler_RegisterAnonymous_Success(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	ctx := context.Background()
	req := connect.NewRequest(&authv1.RegisterAnonymousRequest{
		Username: "testuser",
	})

	expectedResponse := &dto.AuthResponse{
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		User: &dto.UserDTO{
			Username: "testuser",
		},
	}

	mockService.On("RegisterAnonymous", ctx, "testuser").Return(expectedResponse, nil)

	resp, err := handler.RegisterAnonymous(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "access_token_123", resp.Msg.AccessToken)
	assert.Equal(t, "testuser", resp.Msg.User.Username)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_RegisterAnonymous_InvalidUsername(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	ctx := context.Background()
	req := connect.NewRequest(&authv1.RegisterAnonymousRequest{
		Username: "", // Empty username should fail
	})

	resp, err := handler.RegisterAnonymous(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	mockService.AssertNotCalled(t, "RegisterAnonymous")
}

func TestAuthHandler_Login_Success(t *testing.T) {
	mockService := new(MockAuthService)
	handler := NewAuthHandler(mockService)

	ctx := context.Background()
	req := connect.NewRequest(&authv1.LoginRequest{
		Email:    "user@example.com",
		Password: "password123",
	})

	expectedResponse := &dto.AuthResponse{
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_123",
		User: &dto.UserDTO{
			Username: "testuser",
			Email:    "user@example.com",
		},
	}

	mockService.On("Login", ctx, mock.AnythingOfType("*dto.LoginRequest")).Return(expectedResponse, nil)

	resp, err := handler.Login(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "access_token_123", resp.Msg.AccessToken)
	mockService.AssertExpectations(t)
}
