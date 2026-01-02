package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"

	authv1 "github.com/yourorg/anonymous-support/gen/auth/v1"
	"github.com/yourorg/anonymous-support/gen/auth/v1/authv1connect"
	postv1 "github.com/yourorg/anonymous-support/gen/post/v1"
	"github.com/yourorg/anonymous-support/gen/post/v1/postv1connect"
	supportv1 "github.com/yourorg/anonymous-support/gen/support/v1"
	"github.com/yourorg/anonymous-support/gen/support/v1/supportv1connect"
)

const defaultBaseURL = "http://localhost:8080"

func TestIntegrationSOSCreateResponseFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	baseURL := integrationBaseURL()
	waitForServer(t, baseURL)

	httpClient := &http.Client{Timeout: 10 * time.Second}

	authClient := authv1connect.NewAuthServiceClient(httpClient, baseURL)
	postClient := postv1connect.NewPostServiceClient(httpClient, baseURL)
	supportClient := supportv1connect.NewSupportServiceClient(httpClient, baseURL)

	ctx := context.Background()
	username := fmt.Sprintf("integration_sos_%d", time.Now().UnixNano())

	registerResp, err := authClient.RegisterAnonymous(ctx, connect.NewRequest(&authv1.RegisterAnonymousRequest{
		Username: username,
	}))
	require.NoError(t, err)
	require.NotEmpty(t, registerResp.Msg.AccessToken)

	authHeader := fmt.Sprintf("Bearer %s", registerResp.Msg.AccessToken)

	createPostReq := connect.NewRequest(&postv1.CreatePostRequest{
		Type:             postv1.PostType_POST_TYPE_SOS,
		Content:          "Need help staying on track today.",
		Categories:       []string{"recovery"},
		UrgencyLevel:     3,
		TimeContext:      "today",
		DaysSinceRelapse: 2,
		Tags:             []string{"support", "check-in"},
		Visibility:       "public",
	})
	createPostReq.Header().Set("Authorization", authHeader)

	createPostResp, err := postClient.CreatePost(ctx, createPostReq)
	require.NoError(t, err)
	require.NotEmpty(t, createPostResp.Msg.PostId)

	createResponseReq := connect.NewRequest(&supportv1.CreateResponseRequest{
		PostId:  createPostResp.Msg.PostId,
		Type:    supportv1.ResponseType_RESPONSE_TYPE_TEXT,
		Content: "You are not alone. Take it one step at a time.",
	})
	createResponseReq.Header().Set("Authorization", authHeader)

	createResponseResp, err := supportClient.CreateResponse(ctx, createResponseReq)
	require.NoError(t, err)
	require.NotEmpty(t, createResponseResp.Msg.ResponseId)
	require.Greater(t, createResponseResp.Msg.StrengthPointsEarned, int32(0))

	getResponsesResp, err := supportClient.GetResponses(ctx, connect.NewRequest(&supportv1.GetResponsesRequest{
		PostId: createPostResp.Msg.PostId,
		Limit:  10,
		Offset: 0,
	}))
	require.NoError(t, err)
	require.NotEmpty(t, getResponsesResp.Msg.Responses)

	found := false
	for _, response := range getResponsesResp.Msg.Responses {
		if response.Id == createResponseResp.Msg.ResponseId {
			found = true
			require.Equal(t, username, response.Username)
			require.Equal(t, supportv1.ResponseType_RESPONSE_TYPE_TEXT, response.Type)
			require.Equal(t, createResponseReq.Msg.Content, response.Content)
			break
		}
	}

	require.True(t, found, "expected created response to be returned by GetResponses")
}

func integrationBaseURL() string {
	if baseURL := os.Getenv("INTEGRATION_BASE_URL"); baseURL != "" {
		return baseURL
	}
	return defaultBaseURL
}

func waitForServer(t *testing.T, baseURL string) {
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(20 * time.Second)

	for time.Now().Before(deadline) {
		resp, err := client.Get(baseURL + "/ready")
		if err == nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		if err == nil && resp.StatusCode == http.StatusOK {
			return
		}
		time.Sleep(500 * time.Millisecond)
	}

	t.Skipf("Server not reachable at %s (try docker-compose up -d)", baseURL)
}
