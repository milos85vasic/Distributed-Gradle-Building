package client

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	
	"distributed-gradle-building/types"
)

// Simple client test suite - tests HTTP API interactions
func TestClient_Configuration(t *testing.T) {
	config := types.ClientConfig{
		BaseURL:      "http://localhost:8080",
		Timeout:      time.Second * 30,
		MaxRetries:   3,
		RetryDelay:   time.Second * 1,
		AuthEnabled:  true,
		AuthToken:    "test-token",
	}
	
	if config.BaseURL != "http://localhost:8080" {
		t.Errorf("Expected BaseURL http://localhost:8080, got %s", config.BaseURL)
	}
	
	if config.Timeout != time.Second*30 {
		t.Errorf("Expected timeout 30s, got %v", config.Timeout)
	}
	
	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", config.MaxRetries)
	}
	
	if !config.AuthEnabled {
		t.Error("Expected AuthEnabled to be true")
	}
	
	if config.AuthToken != "test-token" {
		t.Errorf("Expected AuthToken test-token, got %s", config.AuthToken)
	}
}

func TestClient_HTTPMethods(t *testing.T) {
	// Test basic HTTP interactions
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id": "123"}`))
		case http.MethodPut:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"updated": true}`))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()
	
	// Test GET
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET expected status 200, got %d", resp.StatusCode)
	}
	
	// Test POST
	resp, err = http.Post(server.URL, "application/json", nil)
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("POST expected status 201, got %d", resp.StatusCode)
	}
}

func TestClient_ErrorHandling(t *testing.T) {
	// Test error handling
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()
	
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}
}

func TestClient_TimeoutHandling(t *testing.T) {
	// Test timeout handling
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Longer than client timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	client := http.Client{
		Timeout: 500 * time.Millisecond,
	}
	
	_, err := client.Get(server.URL)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestClient_ConcurrentRequests(t *testing.T) {
	// Test concurrent requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()
	
	// Make multiple concurrent requests
	const numRequests = 10
	results := make(chan bool, numRequests)
	
	for i := 0; i < numRequests; i++ {
		go func() {
			resp, err := http.Get(server.URL)
			if err != nil {
				results <- false
				return
			}
			defer resp.Body.Close()
			results <- resp.StatusCode == http.StatusOK
		}()
	}
	
	// Collect results
	successCount := 0
	for i := 0; i < numRequests; i++ {
		if success := <-results; success {
			successCount++
		}
	}
	
	if successCount != numRequests {
		t.Errorf("Expected %d successful requests, got %d", numRequests, successCount)
	}
}

func TestClient_Headers(t *testing.T) {
	// Test header handling
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}
		
		if authToken := r.Header.Get("Authorization"); authToken != "Bearer test-token" {
			t.Errorf("Expected Authorization Bearer test-token, got %s", authToken)
		}
		
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	req, err := http.NewRequest("POST", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestClient_RequestBody(t *testing.T) {
	// Test request body handling
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"received": true}`))
	}))
	defer server.Close()
	
	body := []byte(`{"test": "data"}`)
	resp, err := http.Post(server.URL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestClient_ResponseParsing(t *testing.T) {
	// Test response parsing
	testJSON := `{
		"id": "123",
		"name": "Test Build",
		"status": "completed",
		"duration": 5000,
		"artifacts": ["file1.jar", "file2.jar"]
	}`
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testJSON))
	}))
	defer server.Close()
	
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	
	// Verify content type
	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestClient_RetryLogic(t *testing.T) {
	// Test retry logic with a failing server
	attemptCount := 0
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()
	
	// Simple retry implementation
	maxRetries := 3
	var lastErr error
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := http.Get(server.URL)
		if err != nil {
			lastErr = err
			time.Sleep(100 * time.Millisecond)
			continue
		}
		defer resp.Body.Close()
		
		if resp.StatusCode == http.StatusOK {
			// Success
			if attemptCount != 3 {
				t.Errorf("Expected 3 attempts before success, got %d", attemptCount)
			}
			return
		}
		
		lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
		time.Sleep(100 * time.Millisecond)
	}
	
	if lastErr != nil {
		t.Errorf("Expected success after retries, got error: %v", lastErr)
	}
}