package rpc

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/yourorg/anonymous-support/internal/dto"
)

// MockAuthServiceInterface is a testable mock that implements the methods needed by AuthHandler
type MockAuthServiceInterface struct {
	mock.Mock
}

func (m *MockAuthServiceInterface) RegisterAnonymous(ctx context.Context, username string) (*dto.AuthResponse, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthServiceInterface) RegisterWithEmail(ctx context.Context, req *dto.RegisterWithEmailRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthServiceInterface) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthServiceInterface) RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func (m *MockAuthServiceInterface) HandleOAuthLogin(ctx context.Context, provider, providerUserID, email, name string) (*dto.AuthResponse, error) {
	args := m.Called(ctx, provider, providerUserID, email, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

// Note: These tests verify handler logic, not the full service integration
// The handler expects a concrete *service.AuthService, but we can't easily mock that
// In a real scenario, you'd want to test the service separately with repository mocks
// and test handlers with integration tests or by mocking at the repository layer
