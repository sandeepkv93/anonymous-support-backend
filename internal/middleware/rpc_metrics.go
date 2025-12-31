package middleware

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/yourorg/anonymous-support/internal/pkg/metrics"
)

// RPCMetricsInterceptor captures metrics for all RPC calls
type RPCMetricsInterceptor struct{}

// NewRPCMetricsInterceptor creates a new RPC metrics interceptor
func NewRPCMetricsInterceptor() *RPCMetricsInterceptor {
	return &RPCMetricsInterceptor{}
}

// WrapUnary wraps a unary RPC handler with metrics collection
func (i *RPCMetricsInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		start := time.Now()

		// Extract service and method from the request
		procedure := req.Spec().Procedure
		service := extractServiceName(procedure)
		method := extractMethodName(procedure)

		// Call the handler
		resp, err := next(ctx, req)

		// Record metrics
		duration := time.Since(start).Seconds()
		code := "OK"
		if err != nil {
			if connectErr, ok := err.(*connect.Error); ok {
				code = connectErr.Code().String()
			} else {
				code = "Unknown"
			}
			metrics.RPCErrorsTotal.WithLabelValues(service, method, code).Inc()
		}

		metrics.RPCRequestsTotal.WithLabelValues(service, method, code).Inc()
		metrics.RPCRequestDuration.WithLabelValues(service, method).Observe(duration)

		return resp, err
	}
}

// WrapStreamingClient wraps a streaming client RPC handler
func (i *RPCMetricsInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		// For streaming, we just track connection establishment
		service := extractServiceName(spec.Procedure)
		method := extractMethodName(spec.Procedure)
		metrics.RPCRequestsTotal.WithLabelValues(service, method, "STREAMING").Inc()
		return next(ctx, spec)
	}
}

// WrapStreamingHandler wraps a streaming server RPC handler
func (i *RPCMetricsInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		start := time.Now()

		procedure := conn.Spec().Procedure
		service := extractServiceName(procedure)
		method := extractMethodName(procedure)

		err := next(ctx, conn)

		duration := time.Since(start).Seconds()
		code := "OK"
		if err != nil {
			if connectErr, ok := err.(*connect.Error); ok {
				code = connectErr.Code().String()
			} else {
				code = "Unknown"
			}
			metrics.RPCErrorsTotal.WithLabelValues(service, method, code).Inc()
		}

		metrics.RPCRequestsTotal.WithLabelValues(service, method, code).Inc()
		metrics.RPCRequestDuration.WithLabelValues(service, method).Observe(duration)

		return err
	}
}

// extractServiceName extracts the service name from the procedure path
// e.g., "/auth.v1.AuthService/Login" -> "auth.v1.AuthService"
func extractServiceName(procedure string) string {
	// Remove leading slash
	if len(procedure) > 0 && procedure[0] == '/' {
		procedure = procedure[1:]
	}

	// Find the last slash to separate service from method
	for i := len(procedure) - 1; i >= 0; i-- {
		if procedure[i] == '/' {
			return procedure[:i]
		}
	}

	return procedure
}

// extractMethodName extracts the method name from the procedure path
// e.g., "/auth.v1.AuthService/Login" -> "Login"
func extractMethodName(procedure string) string {
	// Find the last slash
	for i := len(procedure) - 1; i >= 0; i-- {
		if procedure[i] == '/' {
			return procedure[i+1:]
		}
	}

	return procedure
}
