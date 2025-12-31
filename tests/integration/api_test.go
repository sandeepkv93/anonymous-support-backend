// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/yourorg/anonymous-support/internal/app"
	"github.com/yourorg/anonymous-support/internal/domain"
)

type APITestSuite struct {
	suite.Suite
	app *app.App
	ctx context.Context
}

func (suite *APITestSuite) SetupSuite() {
	// Initialize test app with test configuration
	suite.ctx = context.Background()

	cfg := &app.Config{
		Environment: "test",
		PostgresDSN: "postgres://localhost:5432/anonymous_support_test?sslmode=disable",
		RedisDSN:    "localhost:6379",
		MongoDSN:    "mongodb://localhost:27017/anonymous_support_test",
	}

	var err error
	suite.app, err = app.NewApp(cfg)
	assert.NoError(suite.T(), err)
}

func (suite *APITestSuite) TearDownSuite() {
	// Clean up test app
	suite.app.Close()
}

func (suite *APITestSuite) TestUserRegistrationFlow() {
	t := suite.T()

	// Test anonymous registration
	authResp, err := suite.app.AuthService.RegisterAnonymous(suite.ctx, "integration_test_user")
	assert.NoError(t, err)
	assert.NotEmpty(t, authResp.AccessToken)
	assert.NotEmpty(t, authResp.RefreshToken)
	assert.Equal(t, "integration_test_user", authResp.User.Username)
	assert.True(t, authResp.User.IsAnonymous)

	userID := authResp.User.ID

	// Test profile retrieval
	profile, err := suite.app.UserService.GetProfile(suite.ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, "integration_test_user", profile.Username)
	assert.True(t, profile.IsAnonymous)

	// Test profile update
	newUsername := "updated_test_user"
	err = suite.app.UserService.UpdateProfile(suite.ctx, userID, &newUsername, nil)
	assert.NoError(t, err)

	// Verify update
	updatedProfile, err := suite.app.UserService.GetProfile(suite.ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, "updated_test_user", updatedProfile.Username)
}

func (suite *APITestSuite) TestPostCreationAndRetrieval() {
	t := suite.T()

	// Create test user
	authResp, err := suite.app.AuthService.RegisterAnonymous(suite.ctx, "post_test_user")
	assert.NoError(t, err)
	userID := authResp.User.ID

	// Create a post
	post, err := suite.app.PostService.CreatePost(
		suite.ctx,
		userID,
		"post_test_user",
		domain.PostTypeSOS,
		"Need help with addiction recovery",
		[]string{"addiction", "support"},
		7,
		"morning",
		8,
		[]string{},
		"public",
		nil,
	)
	assert.NoError(t, err)
	assert.NotNil(t, post)
	assert.Equal(t, userID, post.UserID)
	assert.Equal(t, "Need help with addiction recovery", post.Content)
	assert.Equal(t, domain.PostTypeSOS, post.Type)

	postID := post.ID.Hex()

	// Retrieve the post
	retrievedPost, err := suite.app.PostService.GetPost(suite.ctx, postID)
	assert.NoError(t, err)
	assert.Equal(t, post.ID, retrievedPost.ID)
	assert.Equal(t, post.Content, retrievedPost.Content)

	// Test feed retrieval
	feed, err := suite.app.PostService.GetFeed(suite.ctx, []string{"addiction"}, nil, nil, 10, 0)
	assert.NoError(t, err)
	assert.NotEmpty(t, feed)
}

func (suite *APITestSuite) TestCircleCreationAndMembership() {
	t := suite.T()

	// Create test user
	authResp, err := suite.app.AuthService.RegisterAnonymous(suite.ctx, "circle_test_user")
	assert.NoError(t, err)
	userID := authResp.User.ID

	// Create a circle
	circle, err := suite.app.CircleService.CreateCircle(
		suite.ctx,
		userID,
		"Test Support Circle",
		"A circle for testing",
		"addiction",
		false,
		50,
	)
	assert.NoError(t, err)
	assert.NotNil(t, circle)
	assert.Equal(t, "Test Support Circle", circle.Name)
	assert.Equal(t, userID, circle.OwnerID.String())

	circleID := circle.ID.String()

	// Create another user
	authResp2, err := suite.app.AuthService.RegisterAnonymous(suite.ctx, "circle_member_user")
	assert.NoError(t, err)
	memberUserID := authResp2.User.ID

	// Join circle
	err = suite.app.CircleService.JoinCircle(suite.ctx, memberUserID, circleID)
	assert.NoError(t, err)

	// Verify membership
	circles, err := suite.app.CircleService.GetUserCircles(suite.ctx, memberUserID)
	assert.NoError(t, err)
	assert.NotEmpty(t, circles)

	found := false
	for _, c := range circles {
		if c.ID.String() == circleID {
			found = true
			break
		}
	}
	assert.True(t, found, "User should be member of the circle")
}

func (suite *APITestSuite) TestStreakTracking() {
	t := suite.T()

	// Create test user
	authResp, err := suite.app.AuthService.RegisterAnonymous(suite.ctx, "streak_test_user")
	assert.NoError(t, err)
	userID := authResp.User.ID

	// Get initial streak (should be 0)
	tracker, err := suite.app.UserService.GetStreak(suite.ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, 0, tracker.StreakDays)

	// Update streak (check-in)
	newStreak, err := suite.app.UserService.UpdateStreak(suite.ctx, userID, false)
	assert.NoError(t, err)
	assert.Equal(t, 1, newStreak)

	// Verify updated streak
	updatedTracker, err := suite.app.UserService.GetStreak(suite.ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, 1, updatedTracker.StreakDays)
}

func (suite *APITestSuite) TestTokenRefresh() {
	t := suite.T()

	// Register user
	authResp, err := suite.app.AuthService.RegisterAnonymous(suite.ctx, "token_test_user")
	assert.NoError(t, err)

	originalAccessToken := authResp.AccessToken
	refreshToken := authResp.RefreshToken

	// Wait a bit to ensure new token will be different
	time.Sleep(1 * time.Second)

	// Refresh token
	newAuthResp, err := suite.app.AuthService.RefreshToken(suite.ctx, refreshToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, newAuthResp.AccessToken)
	assert.NotEmpty(t, newAuthResp.RefreshToken)

	// New tokens should be different from original
	assert.NotEqual(t, originalAccessToken, newAuthResp.AccessToken)
	assert.NotEqual(t, refreshToken, newAuthResp.RefreshToken)

	// Old refresh token should be invalid now (token rotation)
	_, err = suite.app.AuthService.RefreshToken(suite.ctx, refreshToken)
	assert.Error(t, err, "Old refresh token should be revoked")
}

func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
