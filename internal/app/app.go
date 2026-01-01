package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	authv1connect "github.com/yourorg/anonymous-support/gen/auth/v1/authv1connect"
	circlev1connect "github.com/yourorg/anonymous-support/gen/circle/v1/circlev1connect"
	moderationv1connect "github.com/yourorg/anonymous-support/gen/moderation/v1/moderationv1connect"
	postv1connect "github.com/yourorg/anonymous-support/gen/post/v1/postv1connect"
	supportv1connect "github.com/yourorg/anonymous-support/gen/support/v1/supportv1connect"
	userv1connect "github.com/yourorg/anonymous-support/gen/user/v1/userv1connect"
	"github.com/yourorg/anonymous-support/internal/config"
	"github.com/yourorg/anonymous-support/internal/handler"
	"github.com/yourorg/anonymous-support/internal/handler/rpc"
	wsHandler "github.com/yourorg/anonymous-support/internal/handler/websocket"
	"github.com/yourorg/anonymous-support/internal/middleware"
	"github.com/yourorg/anonymous-support/internal/pkg/cache"
	"github.com/yourorg/anonymous-support/internal/pkg/encryption"
	"github.com/yourorg/anonymous-support/internal/pkg/jwt"
	"github.com/yourorg/anonymous-support/internal/pkg/migrations"
	"github.com/yourorg/anonymous-support/internal/pkg/moderator"
	"github.com/yourorg/anonymous-support/internal/pkg/tracing"
	"github.com/yourorg/anonymous-support/internal/pkg/transaction"
	"github.com/yourorg/anonymous-support/internal/repository"
	"github.com/yourorg/anonymous-support/internal/repository/mongodb"
	"github.com/yourorg/anonymous-support/internal/repository/postgres"
	redisrepo "github.com/yourorg/anonymous-support/internal/repository/redis"
	"github.com/yourorg/anonymous-support/internal/service"
)

const version = "1.0.0"

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

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
	AuditRepo      repository.AuditRepository

	// Services
	AuthService       service.AuthServiceInterface
	UserService       service.UserServiceInterface
	PostService       service.PostServiceInterface
	SupportService    service.SupportServiceInterface
	CircleService     service.CircleServiceInterface
	ModerationService service.ModerationServiceInterface
	AnalyticsService  service.AnalyticsServiceInterface

	// Infrastructure
	JWTManager        *jwt.JWTManager
	EncryptionManager *encryption.Manager
	TxManager         *transaction.Manager
	Cache             *cache.Cache
	WSHub             *wsHandler.Hub
	TracerProvider    *tracing.TracerProvider

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

	// Initialize transaction manager
	app.TxManager = transaction.NewManager(postgresDB, logger)

	// Initialize cache
	app.Cache = cache.NewCache(redisClient, logger, cache.Config{
		Prefix:     "app",
		DefaultTTL: 5 * time.Minute,
	})

	// Initialize WebSocket hub
	app.WSHub = wsHandler.NewHub(app.JWTManager, logger)

	// Initialize tracing
	tracerProvider, err := tracing.NewTracerProvider(context.Background(), tracing.Config{
		Enabled:     cfg.Server.Env == "production" || cfg.Server.Env == "staging",
		Endpoint:    "localhost:4317", // Configure via env var
		Environment: cfg.Server.Env,
		SampleRate:  1.0, // 100% sampling for now
	})
	if err != nil {
		logger.Warn("Failed to initialize tracing", zap.Error(err))
	}
	app.TracerProvider = tracerProvider

	// Run MongoDB migrations
	if err := migrations.RunMongoDBMigrations(context.Background(), mongoDB); err != nil {
		logger.Warn("Failed to run MongoDB migrations", zap.Error(err))
	}

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
	a.AuditRepo = postgres.NewAuditRepository(a.PostgresDB)

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
		a.UserRepo,
		a.SessionRepo,
		a.JWTManager,
		a.EncryptionManager,
		a.AuditRepo,
	)

	// User service
	a.UserService = service.NewUserService(a.UserRepo, a.AnalyticsRepo)

	// Post service
	contentFilter := moderator.NewContentFilter(a.Config.Moderation.ProfanityFilterLevel)
	a.PostService = service.NewPostService(a.PostRepo, a.RealtimeRepo, contentFilter, a.Cache)

	// Support service
	a.SupportService = service.NewSupportService(a.SupportRepo, a.PostRepo, a.UserRepo, a.RealtimeRepo)

	// Circle service
	a.CircleService = service.NewCircleService(a.CircleRepo, a.PostRepo, a.TxManager)

	// Moderation service
	a.ModerationService = service.NewModerationService(a.ModerationRepo)

	// Analytics service
	a.AnalyticsService = service.NewAnalyticsService(a.AnalyticsRepo)

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

	// Shutdown tracing
	if a.TracerProvider != nil {
		a.Logger.Info("Shutting down tracing")
		if err := a.TracerProvider.Shutdown(shutdownCtx); err != nil {
			a.Logger.Error("Error shutting down tracing", zap.Error(err))
		}
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
	authHandler := rpc.NewAuthHandler(a.AuthService)
	userHandler := rpc.NewUserHandler(a.UserService)
	postHandler := rpc.NewPostHandler(a.PostService)
	supportHandler := rpc.NewSupportHandler(a.SupportService)
	circleHandler := rpc.NewCircleHandler(a.CircleService)
	moderationHandler := rpc.NewModerationHandler(a.ModerationService)

	// Register Connect RPC routes
	authPath, authHTTPHandler := authv1connect.NewAuthServiceHandler(authHandler)
	userPath, userHTTPHandler := userv1connect.NewUserServiceHandler(userHandler)
	postPath, postHTTPHandler := postv1connect.NewPostServiceHandler(postHandler)
	supportPath, supportHTTPHandler := supportv1connect.NewSupportServiceHandler(supportHandler)
	circlePath, circleHTTPHandler := circlev1connect.NewCircleServiceHandler(circleHandler)
	moderationPath, moderationHTTPHandler := moderationv1connect.NewModerationServiceHandler(moderationHandler)

	mux.Handle(authPath, authHTTPHandler)
	mux.Handle(userPath, userHTTPHandler)
	mux.Handle(postPath, postHTTPHandler)
	mux.Handle(supportPath, supportHTTPHandler)
	mux.Handle(circlePath, circleHTTPHandler)
	mux.Handle(moderationPath, moderationHTTPHandler)

	// WebSocket endpoint with auth middleware
	mux.Handle("/ws", middleware.AuthMiddleware(a.JWTManager)(http.HandlerFunc(a.handleWebSocket)))

	// Health check endpoints
	healthHandler := handler.NewHealthHandler(a.Logger, a.PostgresDB, a.MongoDB, a.RedisClient, version, a.Config.Server.Env)
	mux.HandleFunc("/health", healthHandler.Check)
	mux.HandleFunc("/health/ready", healthHandler.Ready)
	mux.HandleFunc("/health/live", healthHandler.Live)

	// Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Setup middleware chain
	httpHandler := middleware.Chain(
		mux,
		middleware.RecoveryMiddleware(a.Logger),
		middleware.SecurityMiddleware(),
		middleware.RequestIDMiddleware(),
		middleware.TracingMiddleware(),
		middleware.MetricsMiddleware(),
		middleware.CORSMiddleware(),
		middleware.LoggingMiddleware(a.Logger),
	)

	// Wrap with h2c for HTTP/2 support without TLS (Connect-RPC)
	h2cHandler := h2c.NewHandler(httpHandler, &http2.Server{})

	// Create HTTP server
	a.HTTPServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", a.Config.Server.Port),
		Handler:      h2cHandler,
		ReadTimeout:  a.Config.Server.ReadTimeout,
		WriteTimeout: a.Config.Server.WriteTimeout,
		IdleTimeout:  a.Config.Server.IdleTimeout,
	}

	a.Logger.Info("HTTP server configured", zap.Int("port", a.Config.Server.Port))
	return nil
}

// handleWebSocket handles WebSocket upgrade and client management
func (a *Application) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserIDFromContext(r.Context())
	username := middleware.GetUsernameFromContext(r.Context())

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		a.Logger.Error("Failed to upgrade WebSocket connection", zap.Error(err))
		return
	}

	client := wsHandler.NewClient(a.WSHub, conn, userID, username)
	a.WSHub.Register <- client

	go client.WritePump()
	go client.ReadPump()
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
