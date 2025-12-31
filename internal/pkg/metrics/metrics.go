package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Database metrics
	DBQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"database", "operation"},
	)

	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"database", "operation"},
	)

	// WebSocket metrics
	WSConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_connections_active",
			Help: "Number of active WebSocket connections",
		},
	)

	WSMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "websocket_messages_total",
			Help: "Total number of WebSocket messages",
		},
		[]string{"type"},
	)

	// Cache metrics
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache"},
	)

	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache"},
	)

	CacheOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_operation_duration_seconds",
			Help:    "Cache operation duration in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"cache", "operation"},
	)

	// RPC/gRPC metrics
	RPCRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rpc_requests_total",
			Help: "Total number of RPC requests",
		},
		[]string{"service", "method", "code"},
	)

	RPCRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rpc_request_duration_seconds",
			Help:    "RPC request duration in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method"},
	)

	RPCErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rpc_errors_total",
			Help: "Total number of RPC errors",
		},
		[]string{"service", "method", "code"},
	)

	// Auth metrics
	AuthAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"type", "result"},
	)

	TokenRefreshesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "token_refreshes_total",
			Help: "Total number of token refresh attempts",
		},
		[]string{"result"},
	)

	ActiveSessionsGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_sessions",
			Help: "Number of active user sessions",
		},
	)

	// Business metrics
	PostsCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "posts_created_total",
			Help: "Total number of posts created",
		},
		[]string{"type"},
	)

	SupportResponsesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "support_responses_total",
			Help: "Total number of support responses created",
		},
	)

	UsersRegisteredTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "users_registered_total",
			Help: "Total number of users registered",
		},
		[]string{"type"},
	)

	CirclesCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "circles_created_total",
			Help: "Total number of circles created",
		},
	)

	CircleMembersGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circle_members",
			Help: "Number of members in each circle",
		},
		[]string{"circle_id"},
	)

	// Content moderation metrics
	ContentReportsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "content_reports_total",
			Help: "Total number of content reports",
		},
		[]string{"content_type"},
	)

	ModerationActionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "moderation_actions_total",
			Help: "Total number of moderation actions taken",
		},
		[]string{"action"},
	)

	// System health metrics
	ConnectionPoolSizeGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "connection_pool_size",
			Help: "Current connection pool size",
		},
		[]string{"database"},
	)

	ConnectionPoolIdleGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "connection_pool_idle",
			Help: "Number of idle connections in pool",
		},
		[]string{"database"},
	)

	ConnectionPoolWaitDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "connection_pool_wait_duration_seconds",
			Help:    "Time spent waiting for a connection",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
		[]string{"database"},
	)
)
