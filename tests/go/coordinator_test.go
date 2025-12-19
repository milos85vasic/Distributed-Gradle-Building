package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
	
	"distributed-gradle-building/coordinator"
	"distributed-gradle-building/types"
)

// Test coordinator functionality
func TestBuildCoordinator_Creation(t *testing.T) {
	coordinator := coordinator.NewBuildCoordinator(5)
	
	if coordinator == nil {
		t.Fatal("Failed to create coordinator")
	}
}

func TestBuildCoordinator_WorkerRegistration(t *testing.T) {
	coord := coordinator.NewBuildCoordinator(5)
	
	worker := &coordinator.Worker{
		ID:   "test-worker-1",
		Host: "localhost",
		Port: 8082,
	}
	
	err := coord.RegisterWorker(worker)
	if err != nil {
		t.Fatalf("Failed to register worker: %v", err)
	}
	
	// Check worker was registered
	workers := coord.GetWorkers()
	if len(workers) != 1 {
		t.Errorf("Expected 1 worker, got %d", len(workers))
	}
	
	if workers[0].ID != "test-worker-1" {
		t.Errorf("Expected worker ID test-worker-1, got %s", workers[0].ID)
	}
}

func TestBuildCoordinator_BuildSubmission(t *testing.T) {
	coord := coordinator.NewBuildCoordinator(5)
	
	request := types.BuildRequest{
		ProjectPath: "/tmp/test-project",
		TaskName:    "build",
		CacheEnabled: true,
		BuildOptions: map[string]string{"clean": "true"},
	}
	
	buildID, err := coord.SubmitBuild(request)
	if err != nil {
		t.Fatalf("Failed to submit build: %v", err)
	}
	
	if buildID == "" {
		t.Error("Expected build ID, got empty string")
	}
	
	// Check build status
	response, err := coord.GetBuildStatus(buildID)
	if err != nil {
		t.Fatalf("Failed to get build status: %v", err)
	}
	
	if response.RequestID != buildID {
		t.Errorf("Expected build ID %s, got %s", buildID, response.RequestID)
	}
}

func TestBuildCoordinator_BuildStatusNotFound(t *testing.T) {
	coord := coordinator.NewBuildCoordinator(5)
	
	_, err := coord.GetBuildStatus("non-existent-build")
	if err == nil {
		t.Error("Expected error for non-existent build")
	}
}

func TestBuildCoordinator_GetWorkersEmpty(t *testing.T) {
	coord := coordinator.NewBuildCoordinator(5)
	
	workers := coord.GetWorkers()
	if len(workers) != 0 {
		t.Errorf("Expected 0 workers, got %d", len(workers))
	}
}

func TestBuildCoordinator_WorkerUnregistration(t *testing.T) {
	coord := coordinator.NewBuildCoordinator(5)
	
	worker := &coordinator.Worker{
		ID:   "test-worker-1",
		Host: "localhost",
		Port: 8082,
	}
	
	coord.RegisterWorker(worker)
	coord.UnregisterWorker("test-worker-1")
	
	workers := coord.GetWorkers()
	if len(workers) != 0 {
		t.Errorf("Expected 0 workers after unregister, got %d", len(workers))
	}
}

func TestBuildCoordinator_HTTPHandlers(t *testing.T) {
	coord := coordinator.NewBuildCoordinator(5)
	
	// Test status handler
	req, err := http.NewRequest("GET", "/api/status", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(coord.HandleStatus)
	handler.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	
	// Check that the response contains expected fields
	if _, ok := response["workers"]; !ok {
		t.Error("Response missing workers field")
	}
	if _, ok := response["queue"]; !ok {
		t.Error("Response missing queue field")
	}
	if _, ok := response["builds"]; !ok {
		t.Error("Response missing builds field")
	}
}

func TestBuildCoordinator_SubmitBuildHTTP(t *testing.T) {
	coord := coordinator.NewBuildCoordinator(5)
	
	request := types.BuildRequest{
		ProjectPath: "/tmp/test-project",
		TaskName:    "build",
		CacheEnabled: true,
	}
	
	jsonBody, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	
	req, err := http.NewRequest("POST", "/api/builds", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(coord.HandleBuilds)
	handler.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	
	if response["build_id"] == "" {
		t.Error("Expected build_id in response")
	}
}

func TestBuildCoordinator_GetBuildHTTP(t *testing.T) {
	coord := coordinator.NewBuildCoordinator(5)
	
	// First submit a build
	request := types.BuildRequest{
		ProjectPath: "/tmp/test-project",
		TaskName:    "test",
	}
	
	buildID, err := coord.SubmitBuild(request)
	if err != nil {
		t.Fatal(err)
	}
	
	// Then get its status
	req, err := http.NewRequest("GET", "/api/builds?id="+buildID, nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(coord.HandleBuilds)
	handler.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var response types.BuildResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	
	if response.RequestID != buildID {
		t.Errorf("Expected build ID %s, got %s", buildID, response.RequestID)
	}
}

func TestBuildCoordinator_GetWorkersHTTP(t *testing.T) {
	coord := coordinator.NewBuildCoordinator(5)
	
	// Register a test worker
	worker := &coordinator.Worker{
		ID:   "test-worker-1",
		Host: "localhost",
		Port: 8082,
	}
	coord.RegisterWorker(worker)
	
	req, err := http.NewRequest("GET", "/api/workers", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(coord.HandleWorkers)
	handler.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	var workers []coordinator.Worker
	err = json.Unmarshal(rr.Body.Bytes(), &workers)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}
	
	if len(workers) != 1 {
		t.Errorf("Expected 1 worker, got %d", len(workers))
	}
	
	if workers[0].ID != "test-worker-1" {
		t.Errorf("Expected worker ID test-worker-1, got %s", workers[0].ID)
	}
}

func TestBuildCoordinator_ConcurrentBuildSubmission(t *testing.T) {
	coord := coordinator.NewBuildCoordinator(5)
	
	var wg sync.WaitGroup
	buildIDs := make(chan string, 10)
	
	// Submit 10 builds concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			request := types.BuildRequest{
				ProjectPath: fmt.Sprintf("/tmp/test-project-%d", id),
				TaskName:    "build",
			}
			
			buildID, err := coord.SubmitBuild(request)
			if err != nil {
				t.Errorf("Failed to submit build %d: %v", id, err)
				return
			}
			
			buildIDs <- buildID
		}(i)
	}
	
	wg.Wait()
	close(buildIDs)
	
	// Verify all builds were submitted
	count := 0
	for buildID := range buildIDs {
		if buildID != "" {
			count++
		}
	}
	
	if count != 10 {
		t.Errorf("Expected 10 builds submitted, got %d", count)
	}
}

func TestBuildCoordinator_Shutdown(t *testing.T) {
	coord := coordinator.NewBuildCoordinator(5)
	
	err := coord.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}