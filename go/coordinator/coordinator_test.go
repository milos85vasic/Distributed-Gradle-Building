package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewBuildCoordinator(t *testing.T) {
	coordinator := NewBuildCoordinator(5)
	if coordinator == nil {
		t.Fatal("Expected coordinator to be created")
	}

	if len(coordinator.workers) != 0 {
		t.Errorf("Expected 0 workers initially, got %d", len(coordinator.workers))
	}

	if cap(coordinator.buildQueue) != 100 {
		t.Errorf("Expected build queue capacity of 100, got %d", cap(coordinator.buildQueue))
	}
}

func TestSubmitBuild(t *testing.T) {
	coordinator := NewBuildCoordinator(5)

	request := BuildRequest{
		ProjectPath:  "/test/project",
		TaskName:     "build",
		RequestID:    "test-request-1",
		CacheEnabled: true,
		Timestamp:    time.Now(),
	}

	// Submit build without workers (should queue it)
	buildID, err := coordinator.SubmitBuild(request)
	if err != nil {
		t.Fatalf("Failed to submit build: %v", err)
	}

	if buildID != request.RequestID {
		t.Errorf("Expected build ID %s, got %s", request.RequestID, buildID)
	}

	// Check if build is in queue
	select {
	case queuedRequest := <-coordinator.buildQueue:
		if queuedRequest.RequestID != request.RequestID {
			t.Errorf("Expected request ID %s, got %s", request.RequestID, queuedRequest.RequestID)
		}
	default:
		t.Error("Expected build to be in queue")
	}
}

func TestGetBuildStatus(t *testing.T) {
	coordinator := NewBuildCoordinator(5)

	// Test non-existent build
	status, err := coordinator.GetBuildStatus("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent build")
	}
	if status != nil {
		t.Error("Expected nil status for non-existent build")
	}

	// Add a build response
	response := &BuildResponse{
		RequestID: "test-request-1",
		Success:   true,
	}
	coordinator.builds["test-request-1"] = response

	// Get existing build status
	status, err = coordinator.GetBuildStatus("test-request-1")
	if err != nil {
		t.Fatalf("Failed to get build status: %v", err)
	}
	if status.RequestID != "test-request-1" {
		t.Errorf("Expected request ID %s, got %s", "test-request-1", status.RequestID)
	}
}

func TestGetWorkers(t *testing.T) {
	coordinator := NewBuildCoordinator(5)

	// Add some workers directly to the map
	coordinator.workers["worker-1"] = &Worker{
		ID:     "worker-1",
		Host:   "localhost",
		Port:   8080,
		Status: "idle",
	}
	coordinator.workers["worker-2"] = &Worker{
		ID:     "worker-2",
		Host:   "localhost",
		Port:   8081,
		Status: "idle",
	}

	retrievedWorkers := coordinator.GetWorkers()
	if len(retrievedWorkers) != 2 {
		t.Errorf("Expected 2 workers, got %d", len(retrievedWorkers))
	}
}

func TestHandleBuildRequestHTTP(t *testing.T) {
	coordinator := NewBuildCoordinator(5)

	// Create test HTTP request
	req := httptest.NewRequest("POST", "/api/build", nil)
	w := httptest.NewRecorder()

	coordinator.handleBuildRequest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for empty request, got %d", w.Code)
	}
}

func TestHandleGetBuildHTTP(t *testing.T) {
	coordinator := NewBuildCoordinator(5)

	// Test non-existent build
	req := httptest.NewRequest("GET", "/api/builds/non-existent", nil)
	w := httptest.NewRecorder()

	coordinator.handleGetBuild(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent build, got %d", w.Code)
	}

	// Add a build response
	response := &BuildResponse{
		RequestID: "test-request-1",
		Success:   true,
	}
	coordinator.builds["test-request-1"] = response

	// Test existing build
	req = httptest.NewRequest("GET", "/api/builds/test-request-1", nil)
	w = httptest.NewRecorder()

	coordinator.handleGetBuild(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for existing build, got %d", w.Code)
	}
}

func TestProcessBuild(t *testing.T) {
	coordinator := NewBuildCoordinator(5)

	// Add a worker directly
	worker := &Worker{
		ID:           "worker-1",
		Host:         "localhost",
		Port:         8080,
		Status:       "idle",
		Capabilities: []string{"gradle"},
		LastPing:     time.Now(),
	}
	coordinator.workers["worker-1"] = worker

	request := BuildRequest{
		ProjectPath:  "/test/project",
		TaskName:     "build",
		RequestID:    "test-request-1",
		CacheEnabled: true,
		Timestamp:    time.Now(),
	}

	// Process build directly (bypass queue)
	coordinator.assignBuildToWorker(worker, request)

	// Check if worker status changed to busy (build was assigned)
	if worker.Status != "busy" {
		t.Errorf("Expected worker status 'busy', got '%s'", worker.Status)
	}

	if len(worker.Builds) != 1 {
		t.Errorf("Expected 1 build assigned to worker, got %d", len(worker.Builds))
	}

	if worker.Builds[0].RequestID != request.RequestID {
		t.Errorf("Expected request ID %s, got %s", request.RequestID, worker.Builds[0].RequestID)
	}
}

func TestGetAvailableWorkers(t *testing.T) {
	coordinator := NewBuildCoordinator(5)

	// Add workers with different statuses
	coordinator.workers["worker-1"] = &Worker{
		ID:       "worker-1",
		Host:     "localhost",
		Port:     8080,
		Status:   "idle",
		LastPing: time.Now(),
	}
	coordinator.workers["worker-2"] = &Worker{
		ID:       "worker-2",
		Host:     "localhost",
		Port:     8081,
		Status:   "busy",
		LastPing: time.Now(),
	}
	coordinator.workers["worker-3"] = &Worker{
		ID:       "worker-3",
		Host:     "localhost",
		Port:     8082,
		Status:   "idle",
		LastPing: time.Now(),
	}

	availableWorkers := coordinator.getAvailableWorkers()
	if len(availableWorkers) != 2 {
		t.Errorf("Expected 2 available workers, got %d", len(availableWorkers))
	}

	// Check that only idle workers are returned
	for _, worker := range availableWorkers {
		if worker.Status != "idle" {
			t.Errorf("Expected worker status 'idle', got '%s'", worker.Status)
		}
	}
}

func TestAssignBuildToWorker(t *testing.T) {
	coordinator := NewBuildCoordinator(5)

	worker := &Worker{
		ID:     "worker-1",
		Host:   "localhost",
		Port:   8080,
		Status: "idle",
	}
	coordinator.workers["worker-1"] = worker

	request := BuildRequest{
		ProjectPath:  "/test/project",
		TaskName:     "build",
		RequestID:    "test-request-1",
		CacheEnabled: true,
		Timestamp:    time.Now(),
	}

	coordinator.assignBuildToWorker(worker, request)

	if worker.Status != "busy" {
		t.Errorf("Expected worker status 'busy', got '%s'", worker.Status)
	}

	if len(worker.Builds) != 1 {
		t.Errorf("Expected 1 build assigned to worker, got %d", len(worker.Builds))
	}

	if worker.Builds[0].RequestID != request.RequestID {
		t.Errorf("Expected request ID %s, got %s", request.RequestID, worker.Builds[0].RequestID)
	}
}

func TestMarkBuildCompleted(t *testing.T) {
	coordinator := NewBuildCoordinator(5)

	// Add a build response
	response := &BuildResponse{
		RequestID: "test-request-1",
		Success:   false,
	}
	coordinator.builds["test-request-1"] = response

	// Mark build as completed
	coordinator.markBuildCompleted("test-request-1", "worker-1")

	// Check if build was marked as successful
	updatedResponse, exists := coordinator.builds["test-request-1"]
	if !exists {
		t.Fatal("Build response should exist")
	}

	if !updatedResponse.Success {
		t.Error("Expected build to be marked as successful")
	}

	if updatedResponse.WorkerID != "worker-1" {
		t.Errorf("Expected worker ID 'worker-1', got '%s'", updatedResponse.WorkerID)
	}
}

func TestMarkBuildFailed(t *testing.T) {
	coordinator := NewBuildCoordinator(5)

	// Add a build response
	response := &BuildResponse{
		RequestID: "test-request-1",
		Success:   false,
	}
	coordinator.builds["test-request-1"] = response

	// Mark build as failed
	coordinator.markBuildFailed("test-request-1", "build failed with error")

	// Check if build was marked as failed with error message
	updatedResponse, exists := coordinator.builds["test-request-1"]
	if !exists {
		t.Fatal("Build response should exist")
	}

	if updatedResponse.Success {
		t.Error("Expected build to be marked as failed")
	}

	if updatedResponse.ErrorMessage != "build failed with error" {
		t.Errorf("Expected error message 'build failed with error', got '%s'", updatedResponse.ErrorMessage)
	}
}

func TestTestRPCMethod(t *testing.T) {
	coordinator := NewBuildCoordinator(5)

	testInput := "test input"
	var reply string

	err := coordinator.Test(&testInput, &reply)
	if err != nil {
		t.Fatalf("Test RPC method failed: %v", err)
	}

	expectedReply := "RPC test successful: " + testInput
	if reply != expectedReply {
		t.Errorf("Expected reply %s, got %s", expectedReply, reply)
	}
}
