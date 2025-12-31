package retry

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Config holds retry configuration
type Config struct {
	MaxAttempts     int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	Multiplier      float64
	RetryableErrors []error
}

// DefaultConfig returns a default retry configuration
func DefaultConfig() Config {
	return Config{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
}

// Operation represents a retryable operation
type Operation func(ctx context.Context) error

// Retrier handles retry logic with exponential backoff
type Retrier struct {
	config Config
	logger *zap.Logger
}

// NewRetrier creates a new retrier with the given configuration
func NewRetrier(config Config, logger *zap.Logger) *Retrier {
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = 3
	}
	if config.InitialDelay <= 0 {
		config.InitialDelay = 100 * time.Millisecond
	}
	if config.MaxDelay <= 0 {
		config.MaxDelay = 5 * time.Second
	}
	if config.Multiplier <= 0 {
		config.Multiplier = 2.0
	}

	return &Retrier{
		config: config,
		logger: logger,
	}
}

// Do executes the operation with retry logic
func (r *Retrier) Do(ctx context.Context, operation Operation) error {
	var lastErr error
	delay := r.config.InitialDelay

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		// Execute the operation
		err := operation(ctx)
		if err == nil {
			// Success
			if attempt > 1 {
				r.logger.Info("Operation succeeded after retry",
					zap.Int("attempt", attempt))
			}
			return nil
		}

		lastErr = err

		// Check if we should retry this error
		if !r.shouldRetry(err) {
			r.logger.Warn("Operation failed with non-retryable error",
				zap.Error(err),
				zap.Int("attempt", attempt))
			return err
		}

		// Check if we have attempts left
		if attempt >= r.config.MaxAttempts {
			r.logger.Error("Operation failed after max attempts",
				zap.Error(err),
				zap.Int("max_attempts", r.config.MaxAttempts))
			return fmt.Errorf("operation failed after %d attempts: %w", r.config.MaxAttempts, err)
		}

		// Log retry attempt
		r.logger.Warn("Operation failed, retrying",
			zap.Error(err),
			zap.Int("attempt", attempt),
			zap.Duration("delay", delay))

		// Wait before retrying (with context cancellation support)
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return fmt.Errorf("operation cancelled: %w", ctx.Err())
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * r.config.Multiplier)
		if delay > r.config.MaxDelay {
			delay = r.config.MaxDelay
		}
	}

	return lastErr
}

// shouldRetry determines if an error is retryable
func (r *Retrier) shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	// If no specific retryable errors are configured, retry all errors
	if len(r.config.RetryableErrors) == 0 {
		return true
	}

	// Check if the error matches any of the configured retryable errors
	for _, retryableErr := range r.config.RetryableErrors {
		if err == retryableErr {
			return true
		}
	}

	return false
}

// WithRetry is a convenience function that creates a retrier and executes the operation
func WithRetry(ctx context.Context, logger *zap.Logger, operation Operation) error {
	retrier := NewRetrier(DefaultConfig(), logger)
	return retrier.Do(ctx, operation)
}

// WithCustomRetry is a convenience function with custom configuration
func WithCustomRetry(ctx context.Context, logger *zap.Logger, config Config, operation Operation) error {
	retrier := NewRetrier(config, logger)
	return retrier.Do(ctx, operation)
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	maxFailures     int
	resetTimeout    time.Duration
	failureCount    int
	lastFailureTime time.Time
	state           CircuitState
	logger          *zap.Logger
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration, logger *zap.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        CircuitClosed,
		logger:       logger,
	}
}

// Execute executes an operation through the circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, operation Operation) error {
	// Check circuit state
	if cb.state == CircuitOpen {
		// Check if we should transition to half-open
		if time.Since(cb.lastFailureTime) >= cb.resetTimeout {
			cb.logger.Info("Circuit breaker transitioning to half-open state")
			cb.state = CircuitHalfOpen
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	}

	// Execute operation
	err := operation(ctx)

	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// recordFailure records a failed operation
func (cb *CircuitBreaker) recordFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.failureCount >= cb.maxFailures {
		cb.logger.Warn("Circuit breaker opening",
			zap.Int("failure_count", cb.failureCount))
		cb.state = CircuitOpen
	}
}

// recordSuccess records a successful operation
func (cb *CircuitBreaker) recordSuccess() {
	if cb.state == CircuitHalfOpen {
		cb.logger.Info("Circuit breaker closing after successful operation")
		cb.state = CircuitClosed
		cb.failureCount = 0
	}
}

// GetState returns the current circuit state
func (cb *CircuitBreaker) GetState() CircuitState {
	return cb.state
}
