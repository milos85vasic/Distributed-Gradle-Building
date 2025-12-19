package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"distributed-gradle-building/errors"
)

func TestNewAuthService(t *testing.T) {
	service := NewAuthService("test-secret", time.Hour)
	if service == nil {
		t.Fatal("Expected AuthService to be created")
	}
}

func TestGenerateToken(t *testing.T) {
	service := NewAuthService("test-secret", time.Hour)
	token, err := service.GenerateToken("user123", "admin", []string{"read", "write"})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	if token == "" {
		t.Fatal("Expected non-empty token")
	}
}

func TestValidateToken(t *testing.T) {
	service := NewAuthService("test-secret", time.Hour)

	// Generate a token
	token, err := service.GenerateToken("user123", "admin", []string{"read", "write"})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate the token
	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got %s", claims.UserID)
	}
	if claims.Role != "admin" {
		t.Errorf("Expected Role 'admin', got %s", claims.Role)
	}
	if len(claims.Permissions) != 2 {
		t.Errorf("Expected 2 permissions, got %d", len(claims.Permissions))
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	service := NewAuthService("test-secret", time.Hour)

	// Test with invalid token
	_, err := service.ValidateToken("invalid.token.here")
	if err == nil {
		t.Fatal("Expected error for invalid token")
	}

	// Check if it's an API error
	if apiErr, ok := err.(*errors.APIError); !ok || apiErr.Code != errors.ErrCodeUnauthorized {
		t.Errorf("Expected unauthorized error, got %v", err)
	}
}

func TestTokenManagement(t *testing.T) {
	service := NewAuthService("test-secret", time.Hour)

	// Test allowed tokens
	service.AddAllowedToken("service-token-123")
	if !service.IsAllowed("service-token-123") {
		t.Error("Expected token to be allowed")
	}
	if service.IsAllowed("non-existent-token") {
		t.Error("Expected non-existent token to not be allowed")
	}

	// Test admin tokens
	service.AddAdminToken("admin-token-456")
	if !service.IsAdmin("admin-token-456") {
		t.Error("Expected token to be admin")
	}
	if service.IsAdmin("non-admin-token") {
		t.Error("Expected non-admin token to not be admin")
	}
}

func TestAuthMiddleware(t *testing.T) {
	service := NewAuthService("test-secret", time.Hour)
	service.AddAllowedToken("static-token")

	// Create test handler
	handler := service.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "No auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid format",
			authHeader:     "InvalidFormat",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Valid static token",
			authHeader:     "Bearer static-token",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid JWT token",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestAdminMiddleware(t *testing.T) {
	service := NewAuthService("test-secret", time.Hour)
	service.AddAdminToken("admin-token")

	handler := service.AdminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "No auth header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Non-admin token",
			authHeader:     "Bearer non-admin-token",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Admin token",
			authHeader:     "Bearer admin-token",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/admin", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	handler := CORSMiddleware(
		[]string{"http://example.com"},
		[]string{"GET", "POST"},
		[]string{"Content-Type", "Authorization"},
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test preflight request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 for OPTIONS, got %d", rr.Code)
	}

	// Check CORS headers
	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "http://example.com",
		"Access-Control-Allow-Methods": "GET, POST",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
	}

	for header, expectedValue := range expectedHeaders {
		if value := rr.Header().Get(header); value != expectedValue {
			t.Errorf("Expected header %s: %s, got %s", header, expectedValue, value)
		}
	}
}

func TestHasPermission(t *testing.T) {
	claims := &Claims{
		UserID:      "user123",
		Role:        "user",
		Permissions: []string{"read", "write"},
	}

	if !HasPermission(claims, "read") {
		t.Error("Expected user to have 'read' permission")
	}
	if !HasPermission(claims, "write") {
		t.Error("Expected user to have 'write' permission")
	}
	if HasPermission(claims, "delete") {
		t.Error("Expected user to not have 'delete' permission")
	}

	// Test wildcard permission
	wildcardClaims := &Claims{
		Permissions: []string{"*"},
	}
	if !HasPermission(wildcardClaims, "any-permission") {
		t.Error("Expected wildcard permission to grant access")
	}
}

func TestGetClaimsFromContext(t *testing.T) {
	claims := &Claims{
		UserID: "test-user",
		Role:   "admin",
	}

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := req.Context()
	ctx = context.WithValue(ctx, "claims", claims)
	req = req.WithContext(ctx)

	retrievedClaims, ok := GetClaimsFromContext(req)
	if !ok {
		t.Fatal("Expected to retrieve claims from context")
	}
	if retrievedClaims.UserID != claims.UserID {
		t.Errorf("Expected UserID %s, got %s", claims.UserID, retrievedClaims.UserID)
	}
}

func TestRequirePermissionMiddleware(t *testing.T) {
	service := NewAuthService("test-secret", time.Hour)

	// Create a handler that requires 'write' permission
	handler := service.AuthMiddleware(
		service.RequirePermission("write")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	// Generate token with only 'read' permission
	token, err := service.GenerateToken("user123", "user", []string{"read"})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Test without required permission
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for missing permission, got %d", rr.Code)
	}
}
