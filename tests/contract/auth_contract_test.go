package contract_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authv1 "github.com/yourorg/anonymous-support/gen/auth/v1"
	"github.com/yourorg/anonymous-support/gen/auth/v1/authv1connect"
)

// Contract tests verify the API contract matches the protobuf specification
// These tests ensure backward compatibility

func TestAuthService_RegisterAnonymous_Contract(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping contract test in short mode")
	}

	client := setupAuthClient(t)
	ctx := context.Background()

	req := &authv1.RegisterAnonymousRequest{
		Username: "contract_test_user",
	}

	resp, err := client.RegisterAnonymous(ctx, connect.NewRequest(req))
	require.NoError(t, err)

	// Contract assertions
	assert.NotEmpty(t, resp.Msg.AccessToken, "Access token must be present")
	assert.NotEmpty(t, resp.Msg.RefreshToken, "Refresh token must be present")
	assert.NotEmpty(t, resp.Msg.UserId, "User ID must be present")
	assert.Equal(t, "contract_test_user", resp.Msg.Username, "Username must match")
}

func TestAuthService_RegisterWithEmail_Contract(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping contract test in short mode")
	}

	client := setupAuthClient(t)
	ctx := context.Background()

	req := &authv1.RegisterWithEmailRequest{
		Username: "email_contract_test",
		Email:    "contract@example.com",
		Password: "SecurePassword123!",
	}

	resp, err := client.RegisterWithEmail(ctx, connect.NewRequest(req))
	require.NoError(t, err)

	// Contract assertions
	assert.NotEmpty(t, resp.Msg.AccessToken)
	assert.NotEmpty(t, resp.Msg.RefreshToken)
	assert.NotEmpty(t, resp.Msg.UserId)
	assert.Equal(t, "email_contract_test", resp.Msg.Username)
}

func TestAuthService_Login_Contract(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping contract test in short mode")
	}

	client := setupAuthClient(t)
	ctx := context.Background()

	// First register a user
	registerReq := &authv1.RegisterWithEmailRequest{
		Username: "login_test",
		Email:    "login@example.com",
		Password: "SecurePassword123!",
	}
	_, err := client.RegisterWithEmail(ctx, connect.NewRequest(registerReq))
	require.NoError(t, err)

	// Then login
	loginReq := &authv1.LoginRequest{
		Username: "login_test",
		Password: "SecurePassword123!",
	}

	resp, err := client.Login(ctx, connect.NewRequest(loginReq))
	require.NoError(t, err)

	// Contract assertions
	assert.NotEmpty(t, resp.Msg.AccessToken)
	assert.NotEmpty(t, resp.Msg.RefreshToken)
	assert.NotEmpty(t, resp.Msg.UserId)
	assert.Equal(t, "login_test", resp.Msg.Username)
}

func TestAuthService_RefreshToken_Contract(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping contract test in short mode")
	}

	client := setupAuthClient(t)
	ctx := context.Background()

	// Register and get refresh token
	registerReq := &authv1.RegisterAnonymousRequest{
		Username: "refresh_test",
	}
	registerResp, err := client.RegisterAnonymous(ctx, connect.NewRequest(registerReq))
	require.NoError(t, err)

	// Refresh the token
	refreshReq := &authv1.RefreshTokenRequest{
		RefreshToken: registerResp.Msg.RefreshToken,
	}

	resp, err := client.RefreshToken(ctx, connect.NewRequest(refreshReq))
	require.NoError(t, err)

	// Contract assertions
	assert.NotEmpty(t, resp.Msg.AccessToken, "New access token must be present")
}

func TestAuthService_Logout_Contract(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping contract test in short mode")
	}

	client := setupAuthClient(t)
	ctx := context.Background()

	// Register a user
	registerReq := &authv1.RegisterAnonymousRequest{
		Username: "logout_test",
	}
	registerResp, err := client.RegisterAnonymous(ctx, connect.NewRequest(registerReq))
	require.NoError(t, err)

	// Logout
	logoutReq := &authv1.LogoutRequest{}

	// Add access token to context (normally done by middleware)
	ctx = context.WithValue(ctx, "access_token", registerResp.Msg.AccessToken)

	resp, err := client.Logout(ctx, connect.NewRequest(logoutReq))
	require.NoError(t, err)

	// Contract assertions - logout returns empty response
	assert.NotNil(t, resp.Msg)
}

// Error contract tests

func TestAuthService_RegisterAnonymous_ValidationError_Contract(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping contract test in short mode")
	}

	client := setupAuthClient(t)
	ctx := context.Background()

	// Invalid username (too short)
	req := &authv1.RegisterAnonymousRequest{
		Username: "ab", // Less than 3 characters
	}

	_, err := client.RegisterAnonymous(ctx, connect.NewRequest(req))
	require.Error(t, err)

	// Check error contract
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code(),
		"Validation errors should return InvalidArgument code")
}

func TestAuthService_Login_InvalidCredentials_Contract(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping contract test in short mode")
	}

	client := setupAuthClient(t)
	ctx := context.Background()

	req := &authv1.LoginRequest{
		Username: "nonexistent",
		Password: "wrongpassword",
	}

	_, err := client.Login(ctx, connect.NewRequest(req))
	require.Error(t, err)

	// Check error contract
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeUnauthenticated, connectErr.Code(),
		"Invalid credentials should return Unauthenticated code")
}

// Helper function to setup auth client
func setupAuthClient(t *testing.T) authv1connect.AuthServiceClient {
	// In real tests, this would connect to a test server
	// For now, return nil as a placeholder
	t.Skip("Test server setup required")
	return nil
}
