package errors

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
)

// ErrorReporter handles error reporting to external services
type ErrorReporter struct {
	logger      *zap.Logger
	environment string
	enabled     bool
}

// NewErrorReporter creates a new error reporter with Sentry integration
func NewErrorReporter(logger *zap.Logger, environment string, sentryDSN string) *ErrorReporter {
	enabled := sentryDSN != "" && (environment == "production" || environment == "staging")

	// Initialize Sentry SDK when DSN is provided
	if enabled {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              sentryDSN,
			Environment:      environment,
			TracesSampleRate: 0.1, // Sample 10% of transactions
			EnableTracing:    true,
			Debug:            environment != "production",
			AttachStacktrace: true,
			BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
				// Filter out sensitive data
				if event.Request != nil {
					event.Request.Cookies = ""
					event.Request.Headers = filterSensitiveHeaders(event.Request.Headers)
				}
				return event
			},
		})

		if err != nil {
			logger.Error("Failed to initialize Sentry", zap.Error(err))
			enabled = false
		} else {
			logger.Info("Sentry error reporting enabled", zap.String("environment", environment))
		}
	}

	return &ErrorReporter{
		logger:      logger,
		environment: environment,
		enabled:     enabled,
	}
}

// ReportError reports an error to the external error tracking service
func (r *ErrorReporter) ReportError(ctx context.Context, err error, tags map[string]string) {
	if !r.enabled {
		r.logger.Debug("Error reporting disabled", zap.Error(err))
		return
	}

	// Log locally
	r.logger.Error("Reported error",
		zap.Error(err),
		zap.String("environment", r.environment),
		zap.Any("tags", tags),
	)

	// Send to Sentry with context
	sentry.WithScope(func(scope *sentry.Scope) {
		// Add tags
		for k, v := range tags {
			scope.SetTag(k, v)
		}

		// Add context data if available
		if ctx != nil {
			if userID := ctx.Value("user_id"); userID != nil {
				scope.SetUser(sentry.User{
					ID: userID.(string),
				})
			}
			if requestID := ctx.Value("request_id"); requestID != nil {
				scope.SetTag("request_id", requestID.(string))
			}
		}

		// Capture the exception
		sentry.CaptureException(err)
	})
}

// ReportMessage reports a message to Sentry (for non-error events)
func (r *ErrorReporter) ReportMessage(ctx context.Context, message string, level sentry.Level, tags map[string]string) {
	if !r.enabled {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		for k, v := range tags {
			scope.SetTag(k, v)
		}
		scope.SetLevel(level)

		if ctx != nil {
			if userID := ctx.Value("user_id"); userID != nil {
				scope.SetUser(sentry.User{
					ID: userID.(string),
				})
			}
		}

		sentry.CaptureMessage(message)
	})
}

// Flush waits for error reports to be sent
func (r *ErrorReporter) Flush(timeout time.Duration) {
	if !r.enabled {
		return
	}
	sentry.Flush(timeout)
}

// filterSensitiveHeaders removes sensitive headers before sending to Sentry
func filterSensitiveHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return nil
	}

	filtered := make(map[string]string)
	sensitiveKeys := map[string]bool{
		"authorization": true,
		"cookie":        true,
		"x-api-key":     true,
		"x-auth-token":  true,
	}

	for k, v := range headers {
		lowerKey := k
		if sensitiveKeys[lowerKey] {
			filtered[k] = "[REDACTED]"
		} else {
			filtered[k] = v
		}
	}

	return filtered
}
