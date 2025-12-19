package errors

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNewAPIError(t *testing.T) {
	err := NewAPIError(ErrCodeBadRequest, "Invalid request")
	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Code != ErrCodeBadRequest {
		t.Errorf("Expected code %s, got %s", ErrCodeBadRequest, err.Code)
	}
	if err.Message != "Invalid request" {
		t.Errorf("Expected message 'Invalid request', got %s", err.Message)
	}
	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusBadRequest, err.HTTPStatus)
	}
	if err.Timestamp == 0 {
		t.Error("Expected timestamp to be set")
	}
}

func TestAPIError_Error(t *testing.T) {
	err := NewAPIError(ErrCodeInternalError, "Something went wrong")
	errorStr := err.Error()
	expected := fmt.Sprintf("%s: %s", ErrCodeInternalError, "Something went wrong")
	if errorStr != expected {
		t.Errorf("Expected error string '%s', got '%s'", expected, errorStr)
	}

	// Test with request ID
	err.WithRequest("req-123")
	errorStr = err.Error()
	expected = fmt.Sprintf("[%s] %s: %s", "req-123", ErrCodeInternalError, "Something went wrong")
	if errorStr != expected {
		t.Errorf("Expected error string with request ID '%s', got '%s'", expected, errorStr)
	}
}

func TestAPIError_WithRequest(t *testing.T) {
	err := NewAPIError(ErrCodeBadRequest, "Invalid request")
	err = err.WithRequest("test-request-123")

	if err.RequestID != "test-request-123" {
		t.Errorf("Expected RequestID 'test-request-123', got %s", err.RequestID)
	}
}

func TestAPIError_WithDetail(t *testing.T) {
	err := NewAPIError(ErrCodeBadRequest, "Invalid request")
	err = err.WithDetail("field", "username").WithDetail("reason", "too short")

	if len(err.Details) != 2 {
		t.Errorf("Expected 2 details, got %d", len(err.Details))
	}
	if err.Details["field"] != "username" {
		t.Errorf("Expected detail 'field' to be 'username', got %v", err.Details["field"])
	}
	if err.Details["reason"] != "too short" {
		t.Errorf("Expected detail 'reason' to be 'too short', got %v", err.Details["reason"])
	}
}

func TestWorkerError(t *testing.T) {
	err := WorkerError(ErrCodeWorkerNotFound, "worker-123", "Worker not found")
	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Code != ErrCodeWorkerNotFound {
		t.Errorf("Expected code %s, got %s", ErrCodeWorkerNotFound, err.Code)
	}
	if err.Message != "Worker not found" {
		t.Errorf("Expected message 'Worker not found', got %s", err.Message)
	}
	if err.Details["worker_id"] != "worker-123" {
		t.Errorf("Expected detail 'worker_id' to be 'worker-123', got %v", err.Details["worker_id"])
	}
	if err.HTTPStatus != http.StatusNotFound {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusNotFound, err.HTTPStatus)
	}
}

func TestBuildError(t *testing.T) {
	err := BuildError(ErrCodeBuildFailed, "build-456", "Build failed with errors")
	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Code != ErrCodeBuildFailed {
		t.Errorf("Expected code %s, got %s", ErrCodeBuildFailed, err.Code)
	}
	if err.Message != "Build failed with errors" {
		t.Errorf("Expected message 'Build failed with errors', got %s", err.Message)
	}
	if err.Details["build_id"] != "build-456" {
		t.Errorf("Expected detail 'build_id' to be 'build-456', got %v", err.Details["build_id"])
	}
	if err.HTTPStatus != http.StatusInternalServerError {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusInternalServerError, err.HTTPStatus)
	}
}

func TestConfigError(t *testing.T) {
	err := ConfigError(ErrCodeInvalidConfig, "max_workers", "Must be positive integer")
	if err == nil {
		t.Fatal("Expected error to be created")
	}

	if err.Code != ErrCodeInvalidConfig {
		t.Errorf("Expected code %s, got %s", ErrCodeInvalidConfig, err.Code)
	}
	if err.Message != "Must be positive integer" {
		t.Errorf("Expected message 'Must be positive integer', got %s", err.Message)
	}
	if err.Details["config_field"] != "max_workers" {
		t.Errorf("Expected detail 'config_field' to be 'max_workers', got %v", err.Details["config_field"])
	}
	if err.HTTPStatus != http.StatusBadRequest {
		t.Errorf("Expected HTTP status %d, got %d", http.StatusBadRequest, err.HTTPStatus)
	}
}

func TestGetHTTPStatusForCode(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected int
	}{
		{ErrCodeBadRequest, http.StatusBadRequest},
		{ErrCodeInvalidConfig, http.StatusBadRequest},
		{ErrCodeConfigValidation, http.StatusBadRequest},
		{ErrCodeUnauthorized, http.StatusUnauthorized},
		{ErrCodeForbidden, http.StatusForbidden},
		{ErrCodeWorkerNotFound, http.StatusNotFound},
		{ErrCodeBuildNotFound, http.StatusNotFound},
		{ErrCodeNotFound, http.StatusNotFound},
		{ErrCodeWorkerUnavailable, http.StatusServiceUnavailable},
		{ErrCodeWorkerCapacity, http.StatusServiceUnavailable},
		{ErrCodeServiceUnavailable, http.StatusServiceUnavailable},
		{ErrCodeBuildFailed, http.StatusInternalServerError},
		{ErrCodeCacheError, http.StatusInternalServerError},
		{ErrCodeStorageError, http.StatusInternalServerError},
		{ErrCodeInternalError, http.StatusInternalServerError},
		{ErrCodeWorkerTimeout, http.StatusRequestTimeout},
		{ErrCodeBuildTimeout, http.StatusRequestTimeout},
		{"UNKNOWN_CODE", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			result := getHTTPStatusForCode(tt.code)
			if result != tt.expected {
				t.Errorf("Expected HTTP status %d for code %s, got %d", tt.expected, tt.code, result)
			}
		})
	}
}

func TestErrorConstants(t *testing.T) {
	// Test that all error codes are defined
	errorCodes := []ErrorCode{
		ErrCodeWorkerNotFound,
		ErrCodeWorkerUnavailable,
		ErrCodeWorkerTimeout,
		ErrCodeWorkerCapacity,
		ErrCodeBuildNotFound,
		ErrCodeBuildFailed,
		ErrCodeBuildTimeout,
		ErrCodeBuildCancelled,
		ErrCodeCacheMiss,
		ErrCodeCacheError,
		ErrCodeStorageError,
		ErrCodeInvalidConfig,
		ErrCodeMissingConfig,
		ErrCodeConfigValidation,
		ErrCodeBadRequest,
		ErrCodeUnauthorized,
		ErrCodeForbidden,
		ErrCodeNotFound,
		ErrCodeInternalError,
		ErrCodeServiceUnavailable,
	}

	for _, code := range errorCodes {
		if code == "" {
			t.Error("Found empty error code")
		}
	}
}

func TestErrorChaining(t *testing.T) {
	// Test that we can chain methods
	err := NewAPIError(ErrCodeBadRequest, "Invalid request").
		WithRequest("req-789").
		WithDetail("field", "email").
		WithDetail("validation", "invalid format")

	if err.RequestID != "req-789" {
		t.Errorf("Expected RequestID 'req-789', got %s", err.RequestID)
	}
	if len(err.Details) != 2 {
		t.Errorf("Expected 2 details, got %d", len(err.Details))
	}
	if err.Details["field"] != "email" {
		t.Errorf("Expected detail 'field' to be 'email', got %v", err.Details["field"])
	}
}

func TestErrorImplementsErrorInterface(t *testing.T) {
	// This test ensures APIError implements the error interface
	var err error = NewAPIError(ErrCodeInternalError, "Test error")
	if err == nil {
		t.Fatal("Expected error to implement error interface")
	}

	// Test type assertion
	if apiErr, ok := err.(*APIError); !ok {
		t.Fatal("Expected error to be of type *APIError")
	} else if apiErr.Code != ErrCodeInternalError {
		t.Errorf("Expected code %s, got %s", ErrCodeInternalError, apiErr.Code)
	}
}
