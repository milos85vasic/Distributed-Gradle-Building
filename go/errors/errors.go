package errors

import (
	"fmt"
	"net/http"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Worker errors
	ErrCodeWorkerNotFound    ErrorCode = "WORKER_NOT_FOUND"
	ErrCodeWorkerUnavailable ErrorCode = "WORKER_UNAVAILABLE"
	ErrCodeWorkerTimeout     ErrorCode = "WORKER_TIMEOUT"
	ErrCodeWorkerCapacity    ErrorCode = "WORKER_CAPACITY_EXCEEDED"

	// Build errors
	ErrCodeBuildNotFound  ErrorCode = "BUILD_NOT_FOUND"
	ErrCodeBuildFailed    ErrorCode = "BUILD_FAILED"
	ErrCodeBuildTimeout   ErrorCode = "BUILD_TIMEOUT"
	ErrCodeBuildCancelled ErrorCode = "BUILD_CANCELLED"

	// Cache errors
	ErrCodeCacheMiss    ErrorCode = "CACHE_MISS"
	ErrCodeCacheError   ErrorCode = "CACHE_ERROR"
	ErrCodeStorageError ErrorCode = "STORAGE_ERROR"

	// Configuration errors
	ErrCodeInvalidConfig    ErrorCode = "INVALID_CONFIG"
	ErrCodeMissingConfig    ErrorCode = "MISSING_CONFIG"
	ErrCodeConfigValidation ErrorCode = "CONFIG_VALIDATION"

	// Network/HTTP errors
	ErrCodeBadRequest         ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden          ErrorCode = "FORBIDDEN"
	ErrCodeNotFound           ErrorCode = "NOT_FOUND"
	ErrCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

// APIError represents a structured API error
type APIError struct {
	Code       ErrorCode      `json:"code"`
	Message    string         `json:"message"`
	Details    map[string]any `json:"details,omitempty"`
	RequestID  string         `json:"request_id,omitempty"`
	Timestamp  int64          `json:"timestamp"`
	HTTPStatus int            `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("[%s] %s: %s", e.RequestID, e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// WithRequest adds request ID to error
func (e *APIError) WithRequest(requestID string) *APIError {
	e.RequestID = requestID
	return e
}

// WithDetail adds detail to error
func (e *APIError) WithDetail(key string, value any) *APIError {
	if e.Details == nil {
		e.Details = make(map[string]any)
	}
	e.Details[key] = value
	return e
}

// NewAPIError creates a new API error
func NewAPIError(code ErrorCode, message string) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		Timestamp:  getCurrentTimestamp(),
		HTTPStatus: getHTTPStatusForCode(code),
	}
}

// WorkerError creates a worker-related error
func WorkerError(code ErrorCode, workerID, message string) *APIError {
	return NewAPIError(code, message).
		WithDetail("worker_id", workerID)
}

// BuildError creates a build-related error
func BuildError(code ErrorCode, buildID, message string) *APIError {
	return NewAPIError(code, message).
		WithDetail("build_id", buildID)
}

// ConfigError creates a configuration-related error
func ConfigError(code ErrorCode, field, message string) *APIError {
	return NewAPIError(code, message).
		WithDetail("config_field", field)
}

// getHTTPStatusForCode maps error codes to HTTP status codes
func getHTTPStatusForCode(code ErrorCode) int {
	switch code {
	case ErrCodeBadRequest, ErrCodeInvalidConfig, ErrCodeConfigValidation:
		return http.StatusBadRequest
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeWorkerNotFound, ErrCodeBuildNotFound, ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeWorkerUnavailable, ErrCodeWorkerCapacity, ErrCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrCodeBuildFailed, ErrCodeCacheError, ErrCodeStorageError, ErrCodeInternalError:
		return http.StatusInternalServerError
	case ErrCodeWorkerTimeout, ErrCodeBuildTimeout:
		return http.StatusRequestTimeout
	default:
		return http.StatusInternalServerError
	}
}

// getCurrentTimestamp returns current Unix timestamp
func getCurrentTimestamp() int64 {
	// In real implementation, use time.Now().Unix()
	return 1640995200 // Placeholder timestamp
}
