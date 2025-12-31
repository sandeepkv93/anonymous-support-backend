package errors

import (
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"go.uber.org/zap"
)

// AppError represents a structured application error with client and server concerns separated
type AppError struct {
	// Code is the error code (e.g., "INVALID_INPUT", "NOT_FOUND")
	Code string
	// Message is the client-safe error message
	Message string
	// HTTPStatus is the HTTP status code to return
	HTTPStatus int
	// ConnectCode is the Connect RPC error code
	ConnectCode connect.Code
	// Internal is the internal error details (not exposed to client)
	Internal error
	// Fields contains additional context fields for logging
	Fields map[string]interface{}
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Internal)
	}
	return e.Message
}

// Unwrap returns the internal error for error unwrapping
func (e *AppError) Unwrap() error {
	return e.Internal
}

// ToConnectError converts AppError to a Connect RPC error
func (e *AppError) ToConnectError() *connect.Error {
	err := connect.NewError(e.ConnectCode, errors.New(e.Message))
	if e.Code != "" {
		err.Meta().Set("code", e.Code)
	}
	return err
}

// LogFields returns structured log fields for this error
func (e *AppError) LogFields() []zap.Field {
	fields := []zap.Field{
		zap.String("error_code", e.Code),
		zap.String("error_message", e.Message),
		zap.Int("http_status", e.HTTPStatus),
	}

	if e.Internal != nil {
		fields = append(fields, zap.Error(e.Internal))
	}

	for k, v := range e.Fields {
		fields = append(fields, zap.Any(k, v))
	}

	return fields
}

// Common error constructors

// NewValidationError creates a validation error
func NewValidationError(message string, internal error) *AppError {
	return &AppError{
		Code:        "VALIDATION_ERROR",
		Message:     message,
		HTTPStatus:  http.StatusBadRequest,
		ConnectCode: connect.CodeInvalidArgument,
		Internal:    internal,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:        "NOT_FOUND",
		Message:     fmt.Sprintf("%s not found", resource),
		HTTPStatus:  http.StatusNotFound,
		ConnectCode: connect.CodeNotFound,
		Fields:      map[string]interface{}{"resource": resource},
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "Unauthorized"
	}
	return &AppError{
		Code:        "UNAUTHORIZED",
		Message:     message,
		HTTPStatus:  http.StatusUnauthorized,
		ConnectCode: connect.CodeUnauthenticated,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
	if message == "" {
		message = "Forbidden"
	}
	return &AppError{
		Code:        "FORBIDDEN",
		Message:     message,
		HTTPStatus:  http.StatusForbidden,
		ConnectCode: connect.CodePermissionDenied,
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string, internal error) *AppError {
	return &AppError{
		Code:        "CONFLICT",
		Message:     message,
		HTTPStatus:  http.StatusConflict,
		ConnectCode: connect.CodeAlreadyExists,
		Internal:    internal,
	}
}

// NewInternalError creates an internal server error
func NewInternalError(message string, internal error) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return &AppError{
		Code:        "INTERNAL_ERROR",
		Message:     message,
		HTTPStatus:  http.StatusInternalServerError,
		ConnectCode: connect.CodeInternal,
		Internal:    internal,
	}
}

// NewRateLimitError creates a rate limit error
func NewRateLimitError(message string) *AppError {
	if message == "" {
		message = "Rate limit exceeded"
	}
	return &AppError{
		Code:        "RATE_LIMIT_EXCEEDED",
		Message:     message,
		HTTPStatus:  http.StatusTooManyRequests,
		ConnectCode: connect.CodeResourceExhausted,
	}
}

// NewUnavailableError creates a service unavailable error
func NewUnavailableError(service string, internal error) *AppError {
	return &AppError{
		Code:        "SERVICE_UNAVAILABLE",
		Message:     fmt.Sprintf("%s is temporarily unavailable", service),
		HTTPStatus:  http.StatusServiceUnavailable,
		ConnectCode: connect.CodeUnavailable,
		Internal:    internal,
		Fields:      map[string]interface{}{"service": service},
	}
}

// NewDeadlineExceededError creates a deadline exceeded error
func NewDeadlineExceededError(operation string) *AppError {
	return &AppError{
		Code:        "DEADLINE_EXCEEDED",
		Message:     fmt.Sprintf("Operation '%s' took too long to complete", operation),
		HTTPStatus:  http.StatusRequestTimeout,
		ConnectCode: connect.CodeDeadlineExceeded,
		Fields:      map[string]interface{}{"operation": operation},
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError attempts to convert an error to an AppError
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// WrapError wraps a generic error as an internal error if it's not already an AppError
func WrapError(err error) *AppError {
	if err == nil {
		return nil
	}

	if appErr, ok := AsAppError(err); ok {
		return appErr
	}

	return NewInternalError("An unexpected error occurred", err)
}
