package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// Test client functionality
func TestClient_Creation(t *testing.T) {
	baseURL := "http://localhost:8080"
	client := NewClient(baseURL)
	
	if client == nil {
		t.Fatal("Failed to create client")
	}
	
	if client.BaseURL != baseURL {
		t.Errorf("Expected base URL %s, got %s", baseURL, client.BaseURL)
	}
	
	if client.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", client.Timeout)
	}
}

func TestClient_SubmitBuild(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		
		if r.URL.Path != "/api/build" {
			t.Errorf("Expected /api/build path, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"build_id": "test-build-123"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	
	request := BuildRequest{
		ProjectPath:  "/tmp/test-project",
		TaskName:     "build",
		CacheEnabled: true,
		BuildOptions: map[string]string{"parallel": "true"},
	}
	
	response, err := client.SubmitBuild(request)
	if err != nil {
		t.Fatalf("SubmitBuild failed: %v", err)
	}
	
	if response.BuildID != "test-build-123" {
		t.Errorf("Expected build ID test-build-123, got %s", response.BuildID)
	}
	
	if response.Status != "queued" {
		t.Errorf("Expected queued status, got %s", response.Status)
	}
}

func TestClient_GetBuildStatus(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		
		if r.URL.Path != "/api/builds/test-build-123" {
			t.Errorf("Expected /api/builds/test-build-123 path, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		response := BuildResponse{
			Success:       true,
			WorkerID:      "worker-1",
			BuildDuration: 5 * time.Minute,
			Artifacts:     []string{"app.jar", "test-results.xml"},
			RequestID:     "test-build-123",
			Timestamp:     time.Now(),
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	
	status, err := client.GetBuildStatus("test-build-123")
	if err != nil {
		t.Fatalf("GetBuildStatus failed: %v", err)
	}
	
	if !status.Success {
		t.Error("Expected success to be true")
	}
	
	if status.WorkerID != "worker-1" {
		t.Errorf("Expected worker ID worker-1, got %s", status.WorkerID)
	}
	
	if len(status.Artifacts) != 2 {
		t.Errorf("Expected 2 artifacts, got %d", len(status.Artifacts))
	}
}

func TestClient_GetWorkers(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		
		if r.URL.Path != "/api/workers" {
			t.Errorf("Expected /api/workers path, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		workers := []Worker{
			{
				ID:           "worker-1",
				Host:         "localhost",
				Port:         8082,
				Status:       "idle",
				Capabilities: []string{"java", "gradle"},
			},
			{
				ID:           "worker-2",
				Host:         "localhost",
				Port:         8083,
				Status:       "busy",
				Capabilities: []string{"java", "kotlin"},
			},
		}
		json.NewEncoder(w).Encode(workers)
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	
	workers, err := client.GetWorkers()
	if err != nil {
		t.Fatalf("GetWorkers failed: %v", err)
	}
	
	if len(workers) != 2 {
		t.Errorf("Expected 2 workers, got %d", len(workers))
	}
	
	if workers[0].ID != "worker-1" {
		t.Errorf("Expected worker-1, got %s", workers[0].ID)
	}
	
	if workers[1].Status != "busy" {
		t.Errorf("Expected busy status, got %s", workers[1].Status)
	}
}

func TestClient_WaitForBuild(t *testing.T) {
	completed := false
	
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if !completed {
			// First call - still running
			response := BuildResponse{
				Success:   false,
				RequestID: "test-build-123",
				Timestamp: time.Now(),
			}
			json.NewEncoder(w).Encode(response)
		} else {
			// Second call - completed
			response := BuildResponse{
				Success:       true,
				WorkerID:      "worker-1",
				BuildDuration: 3 * time.Minute,
				RequestID:     "test-build-123",
				Timestamp:     time.Now(),
			}
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	
	// Simulate completion after 1 second
	go func() {
		time.Sleep(1 * time.Second)
		completed = true
	}()
	
	status, err := client.WaitForBuild("test-build-123", 5*time.Second)
	if err != nil {
		t.Fatalf("WaitForBuild failed: %v", err)
	}
	
	if !status.Success {
		t.Error("Expected final status to be success")
	}
}

func TestClient_WaitForBuild_Timeout(t *testing.T) {
	// Create mock server that always returns incomplete
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := BuildResponse{
			Success:   false,
			RequestID: "test-build-123",
			Timestamp: time.Now(),
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	
	_, err := client.WaitForBuild("test-build-123", 1*time.Second)
	if err == nil {
		t.Error("Expected timeout error")
	}
	
	if err.Error() != "timeout waiting for build completion" {
		t.Errorf("Expected timeout error, got %v", err)
	}
}

func TestClient_GetMetrics(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		
		if r.URL.Path != "/api/metrics" {
			t.Errorf("Expected /api/metrics path, got %s", r.URL.Path)
		}
		
		w.Header().Set("Content-Type", "application/json")
		metrics := map[string]interface{}{
			"total_workers":   5,
			"active_builds":   3,
			"completed_builds": 150,
			"cache_hit_rate":  0.85,
			"average_build_time": "3m45s",
		}
		json.NewEncoder(w).Encode(metrics)
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	
	metrics, err := client.GetMetrics()
	if err != nil {
		t.Fatalf("GetMetrics failed: %v", err)
	}
	
	totalWorkers, exists := metrics["total_workers"]
	if !exists {
		t.Error("Expected total_workers in metrics")
	}
	
	if totalWorkers != 5.0 {
		t.Errorf("Expected total_workers 5, got %v", totalWorkers)
	}
	
	cacheHitRate, exists := metrics["cache_hit_rate"]
	if !exists {
		t.Error("Expected cache_hit_rate in metrics")
	}
	
	if cacheHitRate != 0.85 {
		t.Errorf("Expected cache_hit_rate 0.85, got %v", cacheHitRate)
	}
}

func TestClient_SetTimeout(t *testing.T) {
	client := NewClient("http://localhost:8080")
	
	client.SetTimeout(60 * time.Second)
	
	if client.Timeout != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", client.Timeout)
	}
}

func TestClient_SetAuthToken(t *testing.T) {
	client := NewClient("http://localhost:8080")
	
	token := "test-auth-token"
	client.SetAuthToken(token)
	
	if client.AuthToken != token {
		t.Errorf("Expected auth token %s, got %s", token, client.AuthToken)
	}
}

func TestClient_HTTPRequestHeaders(t *testing.T) {
	var receivedAuth string
	var receivedContentType string
	
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		receivedContentType = r.Header.Get("Content-Type")
		
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"build_id": "test-build-123"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	client.SetAuthToken("Bearer test-token")
	
	request := BuildRequest{
		ProjectPath: "/tmp/test-project",
		TaskName:    "build",
	}
	
	_, err := client.SubmitBuild(request)
	if err != nil {
		t.Fatalf("SubmitBuild failed: %v", err)
	}
	
	if receivedAuth != "Bearer test-token" {
		t.Errorf("Expected Authorization header Bearer test-token, got %s", receivedAuth)
	}
	
	if receivedContentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", receivedContentType)
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"error": "Internal server error"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	
	request := BuildRequest{
		ProjectPath: "/tmp/test-project",
		TaskName:    "build",
	}
	
	_, err := client.SubmitBuild(request)
	if err == nil {
		t.Error("Expected error for server error response")
	}
}

func TestClient_InvalidJSONResponse(t *testing.T) {
	// Create mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json response"))
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	
	_, err := client.GetBuildStatus("test-build-123")
	if err == nil {
		t.Error("Expected error for invalid JSON response")
	}
}

func TestClient_ConcurrentRequests(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"status": "ok"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	
	// Test concurrent requests
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			workers, err := client.GetWorkers()
			if err != nil {
				t.Errorf("Concurrent request %d failed: %v", id, err)
			}
			
			if workers == nil {
				t.Errorf("Concurrent request %d returned nil", id)
			}
		}(i)
	}
	
	wg.Wait()
}

func TestClient_RateLimiting(t *testing.T) {
	requestCount := 0
	
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		
		if requestCount > 5 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"build_id": "test-build-123"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	client := NewClient(server.URL)
	
	request := BuildRequest{
		ProjectPath: "/tmp/test-project",
		TaskName:    "build",
	}
	
	// First 5 requests should succeed
	for i := 0; i < 5; i++ {
		_, err := client.SubmitBuild(request)
		if err != nil {
			t.Errorf("Request %d should succeed: %v", i, err)
		}
	}
	
	// 6th request should be rate limited
	_, err := client.SubmitBuild(request)
	if err == nil {
		t.Error("Expected rate limit error")
	}
}

func TestClient_ConnectionPool(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{"build_id": "test-build-123"}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()
	
	// Create client with custom transport
	client := NewClient(server.URL)
	client.SetTimeout(10 * time.Second)
	
	// Test multiple requests to ensure connection pooling works
	for i := 0; i < 20; i++ {
		request := BuildRequest{
			ProjectPath: "/tmp/test-project",
			TaskName:    "build",
		}
		
		_, err := client.SubmitBuild(request)
		if err != nil {
			t.Errorf("Request %d failed: %v", i, err)
		}
	}
}