package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	rpcHandler "github.com/yourorg/anonymous-support/internal/handler/rpc"
	wsHandler "github.com/yourorg/anonymous-support/internal/handler/websocket"
	"github.com/yourorg/anonymous-support/internal/middleware"
	"github.com/yourorg/anonymous-support/internal/pkg/encryption"
	"github.com/yourorg/anonymous-support/internal/pkg/jwt"
	"github.com/yourorg/anonymous-support/internal/pkg/moderator"
	"github.com/yourorg/anonymous-support/internal/repository/mongodb"
	"github.com/yourorg/anonymous-support/internal/repository/postgres"
	redisRepo "github.com/yourorg/anonymous-support/internal/repository/redis"
	"github.com/yourorg/anonymous-support/internal/service"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}
	defer logger.Sync()

	postgresDB, err := sqlx.Connect("postgres", cfg.Postgres.DSN())
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	defer postgresDB.Close()

	// Configure connection pool
	postgresDB.SetMaxOpenConns(25)
	postgresDB.SetMaxIdleConns(5)
	postgresDB.SetConnMaxLifetime(5 * time.Minute)
	postgresDB.SetConnMaxIdleTime(time.Minute)

	mongoOpts := options.Client().
		ApplyURI(cfg.MongoDB.URI).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(5 * time.Minute).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(5 * time.Second)

	mongoClient, err := mongo.Connect(context.Background(), mongoOpts)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB", zap.Error(err))
	}
	defer mongoClient.Disconnect(context.Background())
	mongoDB := mongoClient.Database(cfg.MongoDB.Database)

	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     50,
		MinIdleConns: 10,
		MaxRetries:   3,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	encManager, err := encryption.NewManager(cfg.Encryption.Key)
	if err != nil {
		logger.Fatal("Failed to create encryption manager", zap.Error(err))
	}
	contentFilter := moderator.NewContentFilter(cfg.Moderation.ProfanityFilterLevel)

	userRepo := postgres.NewUserRepository(postgresDB)
	circleRepo := postgres.NewCircleRepository(postgresDB)
	modRepo := postgres.NewModerationRepository(postgresDB)
	postRepo := mongodb.NewPostRepository(mongoDB)
	supportRepo := mongodb.NewSupportRepository(mongoDB)
	analyticsRepo := mongodb.NewAnalyticsRepository(mongoDB)
	sessionRepo := redisRepo.NewSessionRepository(redisClient)
	realtimeRepo := redisRepo.NewRealtimeRepository(redisClient)
	cacheRepo := redisRepo.NewCacheRepository(redisClient)

	authService := service.NewAuthService(userRepo, sessionRepo, jwtManager, encManager)
	userService := service.NewUserService(userRepo, analyticsRepo)
	postService := service.NewPostService(postRepo, realtimeRepo, contentFilter)
	supportService := service.NewSupportService(supportRepo, postRepo, userRepo, realtimeRepo)
	circleService := service.NewCircleService(circleRepo, postRepo)
	modService := service.NewModerationService(modRepo)
	analyticsService := service.NewAnalyticsService(analyticsRepo)
	notificationService := service.NewNotificationService()

	_ = analyticsService
	_ = notificationService
	_ = cacheRepo

	// Create RPC handlers
	authRPC := rpcHandler.NewAuthHandler(authService)
	userRPC := rpcHandler.NewUserHandler(userService)
	postRPC := rpcHandler.NewPostHandler(postService)
	supportRPC := rpcHandler.NewSupportHandler(supportService)
	circleRPC := rpcHandler.NewCircleHandler(circleService)
	moderationRPC := rpcHandler.NewModerationHandler(modService)

	hub := wsHandler.NewHub()
	go hub.Run()

	// Create health handler
	healthHandler := handler.NewHealthHandler(logger, postgresDB, mongoDB, redisClient, "1.0.0", cfg.Server.Env)

	mux := http.NewServeMux()

	// Register Connect-RPC services
	authPath, authHandler := authv1connect.NewAuthServiceHandler(authRPC)
	userPath, userHandler := userv1connect.NewUserServiceHandler(userRPC)
	postPath, postHandler := postv1connect.NewPostServiceHandler(postRPC)
	supportPath, supportHandler := supportv1connect.NewSupportServiceHandler(supportRPC)
	circlePath, circleHandler := circlev1connect.NewCircleServiceHandler(circleRPC)
	moderationPath, moderationHandler := moderationv1connect.NewModerationServiceHandler(moderationRPC)

	// Mount Connect-RPC handlers
	mux.Handle(authPath, authHandler)
	mux.Handle(userPath, userHandler)
	mux.Handle(postPath, postHandler)
	mux.Handle(supportPath, supportHandler)
	mux.Handle(circlePath, circleHandler)
	mux.Handle(moderationPath, moderationHandler)

	mux.Handle("/ws", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.GetUserIDFromContext(r.Context())
		username := middleware.GetUsernameFromContext(r.Context())

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Error("Failed to upgrade connection", zap.Error(err))
			return
		}

		client := wsHandler.NewClient(hub, conn, userID, username)
		hub.Register <- client

		go client.WritePump()
		go client.ReadPump()
	})))

	// Health check endpoints
	mux.HandleFunc("/health", healthHandler.Check)
	mux.HandleFunc("/health/ready", healthHandler.Ready)
	mux.HandleFunc("/health/live", healthHandler.Live)

	// Metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Apply middleware in order: Recovery -> RequestID -> Metrics -> CORS -> Logging
	handler := middleware.RecoveryMiddleware(logger)(
		middleware.RequestIDMiddleware()(
			middleware.MetricsMiddleware()(
				middleware.CORSMiddleware()(
					middleware.LoggingMiddleware(logger)(mux),
				),
			),
		),
	)

	// Wrap with h2c for HTTP/2 support without TLS (Connect-RPC)
	h2cHandler := h2c.NewHandler(handler, &http2.Server{})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      h2cHandler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.Info("Starting server", zap.Int("port", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
