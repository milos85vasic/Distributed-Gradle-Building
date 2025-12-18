package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

// TestClientCreation tests client initialization
func TestClientCreation(t *testing.T) {
	// Test client creation with valid URL
	baseURL := "http://localhost:8080"
	client, err := NewClient(baseURL)
	
	if err != nil {
		t.Errorf("Expected successful client creation, got error: %v", err)
	}
	
	if client == nil {
		t.Error("Client should not be nil")
	}
	
	if client.BaseURL != baseURL {
		t.Errorf("Expected base URL '%s', got '%s'", baseURL, client.BaseURL)
	}
	
	if client.HTTPClient == nil {
		t.Error("HTTP client should not be nil")
	}
	
	if client.AuthToken != "" {
		t.Errorf("Expected empty auth token, got '%s'", client.AuthToken)
	}
}

// TestClientAuthentication tests client authentication methods
func TestClientAuthentication(t *testing.T) {
	baseURL := "http://localhost:8080"
	client, err := NewClient(baseURL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Test token authentication
	testToken := "test-auth-token-12345"
	client.SetAuthToken(testToken)
	
	if client.AuthToken != testToken {
		t.Errorf("Expected auth token '%s', got '%s'", testToken, client.AuthToken)
	}
	
	// Test API key authentication
	testAPIKey := "test-api-key-67890"
	client.SetAPIKey(testAPIKey)
	
	if client.APIKey != testAPIKey {
		t.Errorf("Expected API key '%s', got '%s'", testAPIKey, client.APIKey)
	}
	
	// Test clearing authentication
	client.ClearAuth()
	
	if client.AuthToken != "" {
		t.Error("Auth token should be cleared")
	}
	
	if client.APIKey != "" {
		t.Error("API key should be cleared")
	}
}

// TestBuildSubmission tests build submission via client
func TestBuildSubmission(t *testing.T) {
	// Mock coordinator server
	var receivedRequest BuildRequest
	expectedBuildID := "test-build-123"
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/api/build" {
			// Parse request
			err := json.NewDecoder(r.Body).Decode(&receivedRequest)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			
			// Mock successful response
			response := BuildSubmissionResponse{
				BuildID: expectedBuildID,
				Status:  "queued",
				Message: "Build submitted successfully",
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()
	
	// Create client with mock server URL
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Create build request
	buildReq := BuildRequest{
		ProjectPath: "/tmp/test-project",
		TaskName:    "build",
		CacheEnabled: true,
		BuildOptions: map[string]string{
			"parallel":   "true",
			"maxWorkers": "4",
		},
		Priority: 1,
	}
	
	// Submit build
	response, err := client.SubmitBuild(buildReq)
	if err != nil {
		t.Errorf("Expected successful build submission, got error: %v", err)
	}
	
	// Verify response
	if response.BuildID != expectedBuildID {
		t.Errorf("Expected build ID '%s', got '%s'", expectedBuildID, response.BuildID)
	}
	
	if response.Status != "queued" {
		t.Errorf("Expected status 'queued', got '%s'", response.Status)
	}
	
	// Verify request was sent correctly
	if receivedRequest.ProjectPath != buildReq.ProjectPath {
		t.Errorf("Expected project path '%s', got '%s'", buildReq.ProjectPath, receivedRequest.ProjectPath)
	}
	
	if receivedRequest.TaskName != buildReq.TaskName {
		t.Errorf("Expected task name '%s', got '%s'", buildReq.TaskName, receivedRequest.TaskName)
	}
	
	if receivedRequest.CacheEnabled != buildReq.CacheEnabled {
		t.Errorf("Expected cache enabled %t, got %t", buildReq.CacheEnabled, receivedRequest.CacheEnabled)
	}
}

// TestBuildStatusRetrieval tests build status retrieval via client
func TestBuildStatusRetrieval(t *testing.T) {
	// Mock coordinator server
	testBuildID := "test-build-456"
	expectedStatus := "running"
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/api/build/"+testBuildID {
			// Mock successful response
			response := BuildStatusResponse{
				ID:          testBuildID,
				Status:      expectedStatus,
				Progress:    0.65,
				StartTime:   time.Now().Add(-5 * time.Minute),
				ProjectPath: "/tmp/test-project",
				TaskName:    "build",
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()
	
	// Create client with mock server URL
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Get build status
	status, err := client.GetBuildStatus(testBuildID)
	if err != nil {
		t.Errorf("Expected successful status retrieval, got error: %v", err)
	}
	
	// Verify response
	if status.ID != testBuildID {
		t.Errorf("Expected build ID '%s', got '%s'", testBuildID, status.ID)
	}
	
	if status.Status != expectedStatus {
		t.Errorf("Expected status '%s', got '%s'", expectedStatus, status.Status)
	}
	
	if status.Progress != 0.65 {
		t.Errorf("Expected progress 0.65, got %f", status.Progress)
	}
}

// TestBuildCancellation tests build cancellation via client
func TestBuildCancellation(t *testing.T) {
	// Mock coordinator server
	testBuildID := "test-build-789"
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" && r.URL.Path == "/api/build/"+testBuildID {
			// Mock successful response
			response := CancelBuildResponse{
				BuildID: testBuildID,
				Success: true,
				Message: "Build cancelled successfully",
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()
	
	// Create client with mock server URL
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Cancel build
	response, err := client.CancelBuild(testBuildID)
	if err != nil {
		t.Errorf("Expected successful build cancellation, got error: %v", err)
	}
	
	// Verify response
	if response.BuildID != testBuildID {
		t.Errorf("Expected build ID '%s', got '%s'", testBuildID, response.BuildID)
	}
	
	if !response.Success {
		t.Error("Expected successful cancellation")
	}
}

// TestWorkerStatusRetrieval tests worker status retrieval via client
func TestWorkerStatusRetrieval(t *testing.T) {
	// Mock coordinator server
	expectedWorkers := []Worker{
		{
			ID:       "worker-1",
			Host:     "localhost",
			Port:     8083,
			Status:   "available",
			Capacity: 4,
			Load:     0.25,
		},
		{
			ID:       "worker-2",
			Host:     "localhost",
			Port:     8084,
			Status:   "busy",
			Capacity: 4,
			Load:     0.75,
			ActiveTask: "compileJava",
		},
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/api/workers" {
			// Mock successful response
			response := WorkersStatusResponse{
				Workers: expectedWorkers,
				Total:   len(expectedWorkers),
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()
	
	// Create client with mock server URL
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Get workers status
	workers, err := client.GetWorkers()
	if err != nil {
		t.Errorf("Expected successful workers retrieval, got error: %v", err)
	}
	
	// Verify response
	if len(workers) != len(expectedWorkers) {
		t.Errorf("Expected %d workers, got %d", len(expectedWorkers), len(workers))
	}
	
	for i, expected := range expectedWorkers {
		if i >= len(workers) {
			t.Errorf("Missing worker %d in response", i)
			continue
		}
		
		actual := workers[i]
		if actual.ID != expected.ID {
			t.Errorf("Expected worker ID '%s', got '%s'", expected.ID, actual.ID)
		}
		
		if actual.Status != expected.Status {
			t.Errorf("Expected worker status '%s', got '%s'", expected.Status, actual.Status)
		}
	}
}

// TestMetricsRetrieval tests system metrics retrieval via client
func TestMetricsRetrieval(t *testing.T) {
	// Mock coordinator server
	expectedMetrics := MetricsResponse{
		TotalBuilds:      100,
		ActiveBuilds:     5,
		CompletedBuilds:   95,
		FailedBuilds:      10,
		AverageBuildTime:  180,
		TotalWorkers:      3,
		AvailableWorkers:  2,
		BusyWorkers:      1,
		CacheHitRate:     0.75,
	}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/api/metrics" {
			// Mock successful response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedMetrics)
		}
	}))
	defer server.Close()
	
	// Create client with mock server URL
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Get metrics
	metrics, err := client.GetMetrics()
	if err != nil {
		t.Errorf("Expected successful metrics retrieval, got error: %v", err)
	}
	
	// Verify response
	if metrics.TotalBuilds != expectedMetrics.TotalBuilds {
		t.Errorf("Expected total builds %d, got %d", expectedMetrics.TotalBuilds, metrics.TotalBuilds)
	}
	
	if metrics.ActiveBuilds != expectedMetrics.ActiveBuilds {
		t.Errorf("Expected active builds %d, got %d", expectedMetrics.ActiveBuilds, metrics.ActiveBuilds)
	}
	
	if metrics.CacheHitRate != expectedMetrics.CacheHitRate {
		t.Errorf("Expected cache hit rate %f, got %f", expectedMetrics.CacheHitRate, metrics.CacheHitRate)
	}
}

// TestBuildWaitForCompletion tests waiting for build completion
func TestBuildWaitForCompletion(t *testing.T) {
	// Mock coordinator server
	testBuildID := "test-build-complete"
	var pollCount int
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/api/build/"+testBuildID {
			pollCount++
			
			// Return different statuses based on poll count
			var response BuildStatusResponse
			if pollCount < 3 {
				// First two polls: running
				response = BuildStatusResponse{
					ID:       testBuildID,
					Status:   "running",
					Progress: float64(pollCount) / 4.0,
				}
			} else {
				// Third poll: completed
				response = BuildStatusResponse{
					ID:        testBuildID,
					Status:    "completed",
					Progress:  1.0,
					StartTime: time.Now().Add(-5 * time.Minute),
					EndTime:   time.Now(),
					Artifacts: []string{
						"build/libs/test-app.jar",
						"build/reports/tests/test/index.html",
					},
				}
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()
	
	// Create client with mock server URL
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Set short poll interval for testing
	client.PollInterval = 10 * time.Millisecond
	
	// Wait for build completion
	status, err := client.WaitForBuild(testBuildID, 5*time.Second)
	if err != nil {
		t.Errorf("Expected successful wait for completion, got error: %v", err)
	}
	
	// Verify final status
	if status.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", status.Status)
	}
	
	if status.Progress != 1.0 {
		t.Errorf("Expected progress 1.0, got %f", status.Progress)
	}
	
	if len(status.Artifacts) != 2 {
		t.Errorf("Expected 2 artifacts, got %d", len(status.Artifacts))
	}
	
	// Verify polling occurred
	if pollCount < 3 {
		t.Errorf("Expected at least 3 polls, got %d", pollCount)
	}
}

// TestBuildWaitForTimeout tests timeout when waiting for build completion
func TestBuildWaitForTimeout(t *testing.T) {
	// Mock coordinator server that always returns "running"
	testBuildID := "test-build-timeout"
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.URL.Path == "/api/build/"+testBuildID {
			// Always return running
			response := BuildStatusResponse{
				ID:       testBuildID,
				Status:   "running",
				Progress: 0.5,
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()
	
	// Create client with mock server URL
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Set short poll interval for testing
	client.PollInterval = 10 * time.Millisecond
	
	// Wait for build completion with short timeout
	startTime := time.Now()
	status, err := client.WaitForBuild(testBuildID, 200*time.Millisecond)
	elapsedTime := time.Since(startTime)
	
	// Verify timeout
	if err == nil {
		t.Error("Expected timeout error")
	}
	
	if elapsedTime < 200*time.Millisecond {
		t.Errorf("Expected to wait at least 200ms, got %v", elapsedTime)
	}
	
	// Status should still be the last received status
	if status.Status != "running" {
		t.Errorf("Expected status 'running', got '%s'", status.Status)
	}
}

// TestErrorHandling tests client error handling
func TestErrorHandling(t *testing.T) {
	// Mock coordinator server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/build/nonexistent":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Build not found",
			})
			
		case "/api/invalid-endpoint":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid endpoint",
			})
			
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()
	
	// Create client with mock server URL
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Test nonexistent build
	_, err = client.GetBuildStatus("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent build")
	}
	
	// Test custom request with invalid endpoint
	req, _ := http.NewRequest("GET", server.URL+"/api/invalid-endpoint", nil)
	_, err = client.DoRequest(req)
	if err == nil {
		t.Error("Expected error for invalid endpoint")
	}
}

// TestClientConfiguration tests client configuration options
func TestClientConfiguration(t *testing.T) {
	baseURL := "http://localhost:8080"
	
	// Test custom HTTP client configuration
	customHTTPClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	config := ClientConfig{
		BaseURL:     baseURL,
		HTTPClient:   customHTTPClient,
		AuthToken:    "test-token",
		APIKey:      "test-key",
		PollInterval: 5 * time.Second,
		MaxRetries:   5,
		UserAgent:    "test-client/1.0",
	}
	
	client, err := NewClientWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create client with config: %v", err)
	}
	
	// Verify configuration
	if client.BaseURL != baseURL {
		t.Errorf("Expected base URL '%s', got '%s'", baseURL, client.BaseURL)
	}
	
	if client.HTTPClient.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", client.HTTPClient.Timeout)
	}
	
	if client.AuthToken != "test-token" {
		t.Errorf("Expected auth token 'test-token', got '%s'", client.AuthToken)
	}
	
	if client.APIKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got '%s'", client.APIKey)
	}
	
	if client.PollInterval != 5*time.Second {
		t.Errorf("Expected poll interval 5s, got %v", client.PollInterval)
	}
	
	if client.MaxRetries != 5 {
		t.Errorf("Expected max retries 5, got %d", client.MaxRetries)
	}
	
	if client.UserAgent != "test-client/1.0" {
		t.Errorf("Expected user agent 'test-client/1.0', got '%s'", client.UserAgent)
	}
}

// TestClientRetryLogic tests client retry logic for failed requests
func TestClientRetryLogic(t *testing.T) {
	var requestCount int
	
	// Mock coordinator server that fails first 2 requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		
		if requestCount <= 2 {
			// Fail first 2 requests
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		
		// Succeed on 3rd request
		response := BuildStatusResponse{
			ID:     "test-build-retry",
			Status: "completed",
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	// Create client with retry configuration
	config := ClientConfig{
		BaseURL:     server.URL,
		MaxRetries:   3,
		PollInterval: 10 * time.Millisecond,
	}
	
	client, err := NewClientWithConfig(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Make request that should trigger retries
	status, err := client.GetBuildStatus("test-build-retry")
	if err != nil {
		t.Errorf("Expected successful request after retries, got error: %v", err)
	}
	
	// Verify retries occurred
	if requestCount != 3 {
		t.Errorf("Expected 3 requests (1 initial + 2 retries), got %d", requestCount)
	}
	
	// Verify final response
	if status.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", status.Status)
	}
}

// TestClientHeaders tests custom headers in client requests
func TestClientHeaders(t *testing.T) {
	var receivedHeaders http.Header
	
	// Mock coordinator server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		response := BuildStatusResponse{
			ID:     "test-build-headers",
			Status: "completed",
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	// Create client
	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	
	// Set custom headers
	client.SetHeader("X-Custom-Header", "custom-value")
	client.SetHeader("X-Client-Version", "1.0.0")
	
	// Make request
	_, err = client.GetBuildStatus("test-build-headers")
	if err != nil {
		t.Errorf("Expected successful request, got error: %v", err)
	}
	
	// Verify headers
	if receivedHeaders.Get("X-Custom-Header") != "custom-value" {
		t.Errorf("Expected custom header 'custom-value', got '%s'", receivedHeaders.Get("X-Custom-Header"))
	}
	
	if receivedHeaders.Get("X-Client-Version") != "1.0.0" {
		t.Errorf("Expected version header '1.0.0', got '%s'", receivedHeaders.Get("X-Client-Version"))
	}
	
	// Verify default User-Agent header
	if receivedHeaders.Get("User-Agent") == "" {
		t.Error("User-Agent header should be set")
	}
}

// TestClientValidation tests client configuration validation
func TestClientValidation(t *testing.T) {
	testCases := []struct {
		name        string
		config      ClientConfig
		expectError bool
	}{
		{
			name: "Valid configuration",
			config: ClientConfig{
				BaseURL:     "http://localhost:8080",
				MaxRetries:   3,
				PollInterval: 1 * time.Second,
			},
			expectError: false,
		},
		{
			name: "Empty base URL",
			config: ClientConfig{
				BaseURL:     "",
				MaxRetries:   3,
				PollInterval: 1 * time.Second,
			},
			expectError: true,
		},
		{
			name: "Invalid URL",
			config: ClientConfig{
				BaseURL:     "not-a-url",
				MaxRetries:   3,
				PollInterval: 1 * time.Second,
			},
			expectError: true,
		},
		{
			name: "Negative max retries",
			config: ClientConfig{
				BaseURL:     "http://localhost:8080",
				MaxRetries:   -1,
				PollInterval: 1 * time.Second,
			},
			expectError: true,
		},
		{
			name: "Zero poll interval",
			config: ClientConfig{
				BaseURL:     "http://localhost:8080",
				MaxRetries:   3,
				PollInterval: 0,
			},
			expectError: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClientWithConfig(tc.config)
			
			if tc.expectError {
				if err == nil {
					t.Error("Expected error but client creation succeeded")
				}
				if client != nil {
					t.Error("Expected nil client on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
				}
				if client == nil {
					t.Error("Expected non-nil client on success")
				}
			}
		})
	}
}