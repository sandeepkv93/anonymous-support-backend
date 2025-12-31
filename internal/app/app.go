package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/yourorg/anonymous-support/internal/config"
	"github.com/yourorg/anonymous-support/internal/handler/rpc"
	"github.com/yourorg/anonymous-support/internal/handler/websocket"
	"github.com/yourorg/anonymous-support/internal/middleware"
	"github.com/yourorg/anonymous-support/internal/pkg/encryption"
	"github.com/yourorg/anonymous-support/internal/pkg/jwt"
	"github.com/yourorg/anonymous-support/internal/pkg/moderator"
	"github.com/yourorg/anonymous-support/internal/repository"
	"github.com/yourorg/anonymous-support/internal/repository/mongodb"
	"github.com/yourorg/anonymous-support/internal/repository/postgres"
	redisrepo "github.com/yourorg/anonymous-support/internal/repository/redis"
	"github.com/yourorg/anonymous-support/internal/service"
)

// Application represents the entire application with all its dependencies
type Application struct {
	Config *config.Config
	Logger *zap.Logger

	// Database clients
	PostgresDB  *sqlx.DB
	MongoDB     *mongo.Database
	RedisClient *redis.Client

	// Repositories
	UserRepo       repository.UserRepository
	PostRepo       repository.PostRepository
	SupportRepo    repository.SupportRepository
	CircleRepo     repository.CircleRepository
	ModerationRepo repository.ModerationRepository
	SessionRepo    repository.SessionRepository
	RealtimeRepo   repository.RealtimeRepository
	CacheRepo      repository.CacheRepository
	AnalyticsRepo  repository.AnalyticsRepository

	// Services
	AuthService       *service.AuthService
	UserService       *service.UserService
	PostService       *service.PostService
	SupportService    *service.SupportService
	CircleService     *service.CircleService
	ModerationService *service.ModerationService
	AnalyticsService  *service.AnalyticsService

	// Infrastructure
	JWTManager        *jwt.JWTManager
	EncryptionManager *encryption.Manager
	WSHub             *websocket.Hub

	// HTTP Server
	HTTPServer *http.Server
}

// New creates and wires up all application dependencies
func New(cfg *config.Config, logger *zap.Logger, postgresDB *sqlx.DB, mongoDB *mongo.Database, redisClient *redis.Client) (*Application, error) {
	app := &Application{
		Config:      cfg,
		Logger:      logger,
		PostgresDB:  postgresDB,
		MongoDB:     mongoDB,
		RedisClient: redisClient,
	}

	// Initialize repositories
	app.wireRepositories()

	// Initialize JWT manager
	app.JWTManager = jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)

	// Initialize encryption manager
	encManager, err := encryption.NewManager(cfg.Encryption.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryption manager: %w", err)
	}
	app.EncryptionManager = encManager

	// Initialize WebSocket hub
	app.WSHub = websocket.NewHub(app.JWTManager, logger)

	// Initialize services
	if err := app.wireServices(); err != nil {
		return nil, fmt.Errorf("failed to wire services: %w", err)
	}

	return app, nil
}

// wireRepositories initializes all repository implementations
func (a *Application) wireRepositories() {
	// Postgres repositories
	a.UserRepo = postgres.NewUserRepository(a.PostgresDB)
	a.CircleRepo = postgres.NewCircleRepository(a.PostgresDB)
	a.ModerationRepo = postgres.NewModerationRepository(a.PostgresDB)

	// MongoDB repositories
	a.PostRepo = mongodb.NewPostRepository(a.MongoDB)
	a.SupportRepo = mongodb.NewSupportRepository(a.MongoDB)
	a.AnalyticsRepo = mongodb.NewAnalyticsRepository(a.MongoDB)

	// Redis repositories
	a.SessionRepo = redisrepo.NewSessionRepository(a.RedisClient)
	a.RealtimeRepo = redisrepo.NewRealtimeRepository(a.RedisClient)
	a.CacheRepo = redisrepo.NewCacheRepository(a.RedisClient)
}

// wireServices initializes all service implementations
func (a *Application) wireServices() error {
	// Auth service
	a.AuthService = service.NewAuthService(
		a.UserRepo.(*postgres.UserRepository),
		a.SessionRepo.(*redisrepo.SessionRepository),
		a.JWTManager,
		a.EncryptionManager,
	)

	// User service
	a.UserService = service.NewUserService(
		a.UserRepo.(*postgres.UserRepository),
		a.AnalyticsRepo.(*mongodb.AnalyticsRepository),
	)

	// Post service
	contentFilter := moderator.NewContentFilter(a.Config.Moderation.ProfanityFilterLevel)
	a.PostService = service.NewPostService(
		a.PostRepo.(*mongodb.PostRepository),
		a.RealtimeRepo.(*redisrepo.RealtimeRepository),
		contentFilter,
	)

	// Support service
	a.SupportService = service.NewSupportService(
		a.SupportRepo.(*mongodb.SupportRepository),
		a.PostRepo.(*mongodb.PostRepository),
		a.UserRepo.(*postgres.UserRepository),
		a.RealtimeRepo.(*redisrepo.RealtimeRepository),
	)

	// Circle service
	a.CircleService = service.NewCircleService(
		a.CircleRepo.(*postgres.CircleRepository),
		a.PostRepo.(*mongodb.PostRepository),
	)

	// Moderation service
	a.ModerationService = service.NewModerationService(
		a.ModerationRepo,
		a.PostRepo,
		a.UserRepo,
		a.Logger,
	)

	// Analytics service
	a.AnalyticsService = service.NewAnalyticsService(
		a.AnalyticsRepo,
		a.Logger,
	)

	return nil
}

// Start starts all application components
func (a *Application) Start(ctx context.Context) error {
	a.Logger.Info("Starting application components")

	// Start WebSocket hub
	go a.WSHub.Run()

	a.Logger.Info("All application components started successfully")
	return nil
}

// Stop gracefully shuts down all application components
func (a *Application) Stop(ctx context.Context) error {
	a.Logger.Info("Stopping application components")

	// Create a context with timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Stop HTTP server
	if a.HTTPServer != nil {
		a.Logger.Info("Shutting down HTTP server")
		if err := a.HTTPServer.Shutdown(shutdownCtx); err != nil {
			a.Logger.Error("Error shutting down HTTP server", zap.Error(err))
		}
	}

	// Stop WebSocket hub
	if a.WSHub != nil {
		a.Logger.Info("Stopping WebSocket hub")
		a.WSHub.Stop()
	}

	// Close database connections
	if a.PostgresDB != nil {
		a.Logger.Info("Closing PostgreSQL connection")
		if err := a.PostgresDB.Close(); err != nil {
			a.Logger.Error("Error closing PostgreSQL connection", zap.Error(err))
		}
	}

	if a.MongoDB != nil {
		a.Logger.Info("Disconnecting from MongoDB")
		if err := a.MongoDB.Client().Disconnect(shutdownCtx); err != nil {
			a.Logger.Error("Error disconnecting from MongoDB", zap.Error(err))
		}
	}

	if a.RedisClient != nil {
		a.Logger.Info("Closing Redis connection")
		if err := a.RedisClient.Close(); err != nil {
			a.Logger.Error("Error closing Redis connection", zap.Error(err))
		}
	}

	a.Logger.Info("Application stopped successfully")
	return nil
}

// SetupHTTPServer creates and configures the HTTP server with all handlers and middleware
func (a *Application) SetupHTTPServer() error {
	mux := http.NewServeMux()

	// Setup RPC handlers
	authHandler := rpc.NewAuthHandler(a.AuthService, a.Logger)
	userHandler := rpc.NewUserHandler(a.UserService, a.Logger)
	postHandler := rpc.NewPostHandler(a.PostService, a.Logger)
	supportHandler := rpc.NewSupportHandler(a.SupportService, a.Logger)
	circleHandler := rpc.NewCircleHandler(a.CircleService, a.Logger)
	moderationHandler := rpc.NewModerationHandler(a.ModerationService, a.Logger)

	// Register RPC routes
	authHandler.RegisterRoutes(mux)
	userHandler.RegisterRoutes(mux)
	postHandler.RegisterRoutes(mux)
	supportHandler.RegisterRoutes(mux)
	circleHandler.RegisterRoutes(mux)
	moderationHandler.RegisterRoutes(mux)

	// Setup middleware chain
	handler := middleware.Chain(
		mux,
		middleware.NewRecoveryMiddleware(a.Logger),
		middleware.NewRequestIDMiddleware(),
		middleware.NewMetricsMiddleware(),
		middleware.NewCORSMiddleware(a.Config),
		middleware.NewLoggingMiddleware(a.Logger),
	)

	// Create HTTP server
	a.HTTPServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", a.Config.Server.Port),
		Handler:      handler,
		ReadTimeout:  a.Config.Server.ReadTimeout,
		WriteTimeout: a.Config.Server.WriteTimeout,
		IdleTimeout:  a.Config.Server.IdleTimeout,
	}

	a.Logger.Info("HTTP server configured", zap.Int("port", a.Config.Server.Port))
	return nil
}

// Run starts the HTTP server and blocks until shutdown
func (a *Application) Run(ctx context.Context) error {
	// Start background components
	if err := a.Start(ctx); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}

	// Setup HTTP server
	if err := a.SetupHTTPServer(); err != nil {
		return fmt.Errorf("failed to setup HTTP server: %w", err)
	}

	// Start HTTP server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		a.Logger.Info("Starting HTTP server", zap.String("address", a.HTTPServer.Addr))
		serverErrors <- a.HTTPServer.ListenAndServe()
	}()

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	case <-ctx.Done():
		a.Logger.Info("Shutdown signal received")
		if err := a.Stop(ctx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}
	}

	return nil
}
