package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type HealthHandler struct {
	logger      *zap.Logger
	postgres    *sqlx.DB
	mongodb     *mongo.Database
	redis       *redis.Client
	version     string
	environment string
}

type HealthResponse struct {
	Status       string                      `json:"status"`
	Version      string                      `json:"version"`
	Environment  string                      `json:"environment"`
	Timestamp    string                      `json:"timestamp"`
	Dependencies map[string]DependencyHealth `json:"dependencies"`
}

type DependencyHealth struct {
	Status       string `json:"status"`
	ResponseTime string `json:"response_time,omitempty"`
	Error        string `json:"error,omitempty"`
}

func NewHealthHandler(logger *zap.Logger, pg *sqlx.DB, mongo *mongo.Database, redis *redis.Client, version, env string) *HealthHandler {
	return &HealthHandler{
		logger:      logger,
		postgres:    pg,
		mongodb:     mongo,
		redis:       redis,
		version:     version,
		environment: env,
	}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	deps := make(map[string]DependencyHealth)

	// Check PostgreSQL
	pgStatus := h.checkPostgres(ctx)
	deps["postgres"] = pgStatus

	// Check MongoDB
	mongoStatus := h.checkMongo(ctx)
	deps["mongodb"] = mongoStatus

	// Check Redis
	redisStatus := h.checkRedis(ctx)
	deps["redis"] = redisStatus

	// Determine overall status
	overallStatus := "healthy"
	httpStatus := http.StatusOK
	for _, dep := range deps {
		if dep.Status != "healthy" {
			overallStatus = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
			break
		}
	}

	response := HealthResponse{
		Status:       overallStatus,
		Version:      h.version,
		Environment:  h.environment,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
		Dependencies: deps,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(response)
}

func (h *HealthHandler) checkPostgres(ctx context.Context) DependencyHealth {
	start := time.Now()
	err := h.postgres.PingContext(ctx)
	duration := time.Since(start)

	if err != nil {
		h.logger.Error("Postgres health check failed", zap.Error(err))
		return DependencyHealth{
			Status:       "unhealthy",
			ResponseTime: duration.String(),
			Error:        err.Error(),
		}
	}

	return DependencyHealth{
		Status:       "healthy",
		ResponseTime: duration.String(),
	}
}

func (h *HealthHandler) checkMongo(ctx context.Context) DependencyHealth {
	start := time.Now()
	err := h.mongodb.Client().Ping(ctx, nil)
	duration := time.Since(start)

	if err != nil {
		h.logger.Error("MongoDB health check failed", zap.Error(err))
		return DependencyHealth{
			Status:       "unhealthy",
			ResponseTime: duration.String(),
			Error:        err.Error(),
		}
	}

	return DependencyHealth{
		Status:       "healthy",
		ResponseTime: duration.String(),
	}
}

func (h *HealthHandler) checkRedis(ctx context.Context) DependencyHealth {
	start := time.Now()
	err := h.redis.Ping(ctx).Err()
	duration := time.Since(start)

	if err != nil {
		h.logger.Error("Redis health check failed", zap.Error(err))
		return DependencyHealth{
			Status:       "unhealthy",
			ResponseTime: duration.String(),
			Error:        err.Error(),
		}
	}

	return DependencyHealth{
		Status:       "healthy",
		ResponseTime: duration.String(),
	}
}

// Ready returns a simple readiness check (lighter than full health check)
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// Live returns a simple liveness check
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
