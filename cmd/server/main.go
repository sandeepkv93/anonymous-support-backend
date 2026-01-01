package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"

	"github.com/yourorg/anonymous-support/internal/app"
	"github.com/yourorg/anonymous-support/internal/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize logger
	logger, err := initLogger(cfg.Server.Env)
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}
	defer func() { _ = logger.Sync() }()

	// Initialize database connections
	postgresDB, err := initPostgres(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize PostgreSQL", zap.Error(err))
	}
	defer postgresDB.Close()

	mongoDB, mongoDisconnect, err := initMongoDB(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize MongoDB", zap.Error(err))
	}
	defer mongoDisconnect()

	redisClient, err := initRedis(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Create application with all wired dependencies
	application, err := app.New(cfg, logger, postgresDB, mongoDB, redisClient)
	if err != nil {
		logger.Fatal("Failed to create application", zap.Error(err))
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start application in a goroutine
	appErrors := make(chan error, 1)
	go func() {
		appErrors <- application.Run(ctx)
	}()

	// Wait for shutdown signal or application error
	select {
	case err := <-appErrors:
		if err != nil {
			logger.Fatal("Application error", zap.Error(err))
		}
	case sig := <-quit:
		logger.Info("Shutdown signal received", zap.String("signal", sig.String()))
		cancel()

		// Give application time to gracefully shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := application.Stop(shutdownCtx); err != nil {
			logger.Error("Graceful shutdown failed", zap.Error(err))
			os.Exit(1)
		}
	}

	logger.Info("Server exited successfully")
}

// initLogger creates a logger based on environment
func initLogger(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

// initPostgres initializes PostgreSQL connection with proper pooling
func initPostgres(cfg *config.Config, logger *zap.Logger) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.Postgres.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(time.Minute)

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping: %w", err)
	}

	logger.Info("PostgreSQL connected successfully")
	return db, nil
}

// initMongoDB initializes MongoDB connection with proper configuration
func initMongoDB(cfg *config.Config, logger *zap.Logger) (*mongo.Database, func(), error) {
	opts := options.Client().
		ApplyURI(cfg.MongoDB.URI).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(5 * time.Minute).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(5 * time.Second)

	client, err := mongo.Connect(context.Background(), opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Verify connection
	if err := client.Ping(context.Background(), nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, nil, fmt.Errorf("failed to ping: %w", err)
	}

	db := client.Database(cfg.MongoDB.Database)
	disconnect := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			logger.Error("Failed to disconnect from MongoDB", zap.Error(err))
		}
	}

	logger.Info("MongoDB connected successfully")
	return db, disconnect, nil
}

// initRedis initializes Redis connection with proper configuration
func initRedis(cfg *config.Config, logger *zap.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
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

	// Verify connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping: %w", err)
	}

	logger.Info("Redis connected successfully")
	return client, nil
}
