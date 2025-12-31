//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/yourorg/anonymous-support/internal/domain"
	"github.com/yourorg/anonymous-support/internal/repository"
	"github.com/yourorg/anonymous-support/internal/service"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPostServiceIntegration(t *testing.T) {
	ctx := context.Background()

	// Start MongoDB container
	mongoContainer, err := mongodb.Run(ctx,
		"mongo:7",
		mongodb.WithUsername("test"),
		mongodb.WithPassword("test"),
	)
	require.NoError(t, err)
	defer func() {
		if err := testcontainers.TerminateContainer(mongoContainer); err != nil {
			t.Logf("failed to terminate mongodb container: %s", err)
		}
	}()

	mongoURI, err := mongoContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	require.NoError(t, err)
	defer mongoClient.Disconnect(ctx)

	mongoDB := mongoClient.Database("test_db")

	// Start PostgreSQL container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Minute)),
	)
	require.NoError(t, err)
	defer func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			t.Logf("failed to terminate postgres container: %s", err)
		}
	}()

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Connect to PostgreSQL
	pgDB, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate tables
	err = pgDB.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	// Start Redis container
	redisContainer, err := redis.Run(ctx, "redis:7-alpine")
	require.NoError(t, err)
	defer func() {
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			t.Logf("failed to terminate redis container: %s", err)
		}
	}()

	redisURI, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisURI,
	})
	defer redisClient.Close()

	// Setup repositories and services
	logger, _ := zap.NewDevelopment()
	postRepo := repository.NewMongoPostRepository(mongoDB, logger)
	analyticsRepo := repository.NewMongoAnalyticsRepository(mongoDB, logger)

	postService := service.NewPostService(postRepo, analyticsRepo)

	// Test: Create a post
	t.Run("CreatePost", func(t *testing.T) {
		post, err := postService.CreatePost(
			ctx,
			"user123",
			"testuser",
			domain.PostTypeSOS,
			"Need help with my recovery",
			[]string{"addiction"},
			5,
			"evening",
			10,
			[]string{"relapse-prevention"},
			"public",
			nil,
		)

		require.NoError(t, err)
		assert.NotNil(t, post)
		assert.Equal(t, "user123", post.UserID)
		assert.Equal(t, "testuser", post.Username)
		assert.Equal(t, domain.PostTypeSOS, post.Type)
		assert.Equal(t, "Need help with my recovery", post.Content)
		assert.Equal(t, 5, post.UrgencyLevel)
	})

	// Test: Get post
	t.Run("GetPost", func(t *testing.T) {
		// Create a post first
		created, err := postService.CreatePost(
			ctx,
			"user456",
			"anotheruser",
			domain.PostTypeCheckIn,
			"Day 30 clean!",
			[]string{"recovery"},
			1,
			"morning",
			30,
			[]string{"milestone"},
			"public",
			nil,
		)
		require.NoError(t, err)

		// Retrieve it
		retrieved, err := postService.GetPost(ctx, created.ID.Hex())
		require.NoError(t, err)
		assert.Equal(t, created.ID, retrieved.ID)
		assert.Equal(t, "Day 30 clean!", retrieved.Content)
	})

	// Test: Get feed
	t.Run("GetFeed", func(t *testing.T) {
		// Create multiple posts
		for i := 0; i < 5; i++ {
			_, err := postService.CreatePost(
				ctx,
				"user789",
				"feeduser",
				domain.PostTypeVictory,
				"Victory post",
				[]string{"recovery"},
				1,
				"morning",
				i,
				[]string{"victory"},
				"public",
				nil,
			)
			require.NoError(t, err)
		}

		// Get feed
		posts, err := postService.GetFeed(ctx, []string{"recovery"}, nil, nil, 10, 0)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(posts), 5)
	})

	// Test: Delete post
	t.Run("DeletePost", func(t *testing.T) {
		// Create a post
		created, err := postService.CreatePost(
			ctx,
			"user999",
			"deleteuser",
			domain.PostTypeQuestion,
			"To be deleted",
			[]string{"test"},
			1,
			"evening",
			0,
			[]string{},
			"public",
			nil,
		)
		require.NoError(t, err)

		// Delete it
		err = postService.DeletePost(ctx, created.ID.Hex(), "user999")
		require.NoError(t, err)

		// Try to get it (should fail or return soft-deleted)
		_, err = postService.GetPost(ctx, created.ID.Hex())
		assert.Error(t, err)
	})
}
