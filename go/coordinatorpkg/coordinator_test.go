package coordinatorpkg

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"distributed-gradle-building/types"
)

func TestNewBuildCoordinator(t *testing.T) {
	coordinator := NewBuildCoordinator(10)
	if coordinator == nil {
		t.Fatal("Expected coordinator to be created")
	}
	if len(coordinator.workers) != 0 {
		t.Errorf("Expected 0 workers, got %d", len(coordinator.workers))
	}
	if cap(coordinator.buildQueue) != 100 {
		t.Errorf("Expected build queue capacity 100, got %d", cap(coordinator.buildQueue))
	}
	if coordinator.maxWorkers != 10 {
		t.Errorf("Expected maxWorkers 10, got %d", coordinator.maxWorkers)
	}
}

func TestRegisterWorker(t *testing.T) {
	coordinator := NewBuildCoordinator(5)
	worker := &Worker{
		ID:       "worker-1",
		Host:     "localhost",
		Port:     8080,
		Status:   "idle",
		LastPing: time.Now(),
		Metadata: map[string]interface{}{"capabilities": []string{"build", "test"}},
	}

	err := coordinator.RegisterWorker(worker)
	if err != nil {
		t.Fatalf("Failed to register worker: %v", err)
	}
	if len(coordinator.workers) != 1 {
		t.Errorf("Expected 1 worker, got %d", len(coordinator.workers))
	}
	if registeredWorker, ok := coordinator.workers["worker-1"]; !ok {
		t.Error("Worker not found in coordinator")
	} else if registeredWorker.Host != "localhost" {
		t.Errorf("Expected host localhost, got %s", registeredWorker.Host)
	}
}

func TestRegisterWorker_MaxWorkers(t *testing.T) {
	coordinator := NewBuildCoordinator(2)

	// Register first worker
	worker1 := &Worker{ID: "worker-1", Host: "localhost", Port: 8080}
	err := coordinator.RegisterWorker(worker1)
	if err != nil {
		t.Fatalf("Failed to register worker1: %v", err)
	}

	// Register second worker
	worker2 := &Worker{ID: "worker-2", Host: "localhost", Port: 8081}
	err = coordinator.RegisterWorker(worker2)
	if err != nil {
		t.Fatalf("Failed to register worker2: %v", err)
	}

	// Try to register third worker (should fail)
	worker3 := &Worker{ID: "worker-3", Host: "localhost", Port: 8082}
	err = coordinator.RegisterWorker(worker3)
	if err == nil {
		t.Error("Expected error when exceeding max workers")
	}
	expectedErr := "maximum workers (2) reached"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%v'", expectedErr, err)
	}
}

func TestRegisterWorkerRPC(t *testing.T) {
	coordinator := NewBuildCoordinator(5)
	args := &RegisterWorkerArgs{
		ID:           "worker-1",
		Host:         "localhost",
		Port:         8080,
		Capabilities: []string{"build", "test"},
		Status:       "ready",
	}
	reply := &RegisterWorkerReply{}

	err := coordinator.RegisterWorkerRPC(args, reply)
	if err != nil {
		t.Fatalf("RegisterWorkerRPC failed: %v", err)
	}
	if len(coordinator.workers) != 1 {
		t.Errorf("Expected 1 worker, got %d", len(coordinator.workers))
	}
	if reply.Message != "Worker worker-1 registered successfully" {
		t.Errorf("Expected success message, got %s", reply.Message)
	}
	if worker, ok := coordinator.workers["worker-1"]; !ok {
		t.Error("Worker not found")
	} else if worker.Status != "idle" {
		t.Errorf("Expected status idle, got %s", worker.Status)
	}
}

func TestUnregisterWorker(t *testing.T) {
	coordinator := NewBuildCoordinator(5)
	worker := &Worker{ID: "worker-1", Host: "localhost", Port: 8080}
	coordinator.RegisterWorker(worker)

	if len(coordinator.workers) != 1 {
		t.Errorf("Expected 1 worker before unregister, got %d", len(coordinator.workers))
	}

	coordinator.UnregisterWorker("worker-1")
	if len(coordinator.workers) != 0 {
		t.Errorf("Expected 0 workers after unregister, got %d", len(coordinator.workers))
	}

	// Unregister non-existent worker (should not panic)
	coordinator.UnregisterWorker("nonexistent")
}

func TestSubmitBuild(t *testing.T) {
	coordinator := NewBuildCoordinator(5)
	request := types.BuildRequest{
		ProjectPath:  "/test/project",
		TaskName:     "build",
		CacheEnabled: true,
		BuildOptions: map[string]string{"clean": "true"},
	}

	buildID, err := coordinator.SubmitBuild(request)
	if err != nil {
		t.Fatalf("SubmitBuild failed: %v", err)
	}
	if buildID == "" {
		t.Error("Expected non-empty build ID")
	}

	// Verify build was stored
	response, err := coordinator.GetBuildStatus(buildID)
	if err != nil {
		t.Fatalf("GetBuildStatus failed: %v", err)
	}
	if response.RequestID != buildID {
		t.Errorf("Expected RequestID %s, got %s", buildID, response.RequestID)
	}
	if response.Success {
		t.Error("Expected Success false for new build")
	}
}

func TestSubmitBuild_WithRequestID(t *testing.T) {
	coordinator := NewBuildCoordinator(5)
	request := types.BuildRequest{
		RequestID:    "custom-build-123",
		ProjectPath:  "/test/project",
		TaskName:     "build",
		CacheEnabled: true,
	}

	buildID, err := coordinator.SubmitBuild(request)
	if err != nil {
		t.Fatalf("SubmitBuild failed: %v", err)
	}
	if buildID != "custom-build-123" {
		t.Errorf("Expected build ID custom-build-123, got %s", buildID)
	}
}

func TestSubmitBuild_QueueFull(t *testing.T) {
	coordinator := NewBuildCoordinator(5)
	coordinator.buildQueue = make(chan types.BuildRequest, 1) // Small queue for testing

	// Fill the queue
	request1 := types.BuildRequest{ProjectPath: "/project1", TaskName: "build"}
	buildID1, err := coordinator.SubmitBuild(request1)
	if err != nil {
		t.Fatalf("First SubmitBuild failed: %v", err)
	}
	if buildID1 == "" {
		t.Error("Expected non-empty build ID for first request")
	}

	// Try to submit another build (should fail)
	request2 := types.BuildRequest{ProjectPath: "/project2", TaskName: "test"}
	_, err = coordinator.SubmitBuild(request2)
	if err == nil {
		t.Error("Expected error when queue is full")
	}
	expectedErr := "build queue is full"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%v'", expectedErr, err)
	}
}

func TestGetBuildStatus(t *testing.T) {
	coordinator := NewBuildCoordinator(5)
	request := types.BuildRequest{
		ProjectPath: "/test/project",
		TaskName:    "build",
	}

	buildID, err := coordinator.SubmitBuild(request)
	if err != nil {
		t.Fatalf("SubmitBuild failed: %v", err)
	}

	response, err := coordinator.GetBuildStatus(buildID)
	if err != nil {
		t.Fatalf("GetBuildStatus failed: %v", err)
	}
	if response.RequestID != buildID {
		t.Errorf("Expected RequestID %s, got %s", buildID, response.RequestID)
	}
	if response.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

func TestGetBuildStatus_NotFound(t *testing.T) {
	coordinator := NewBuildCoordinator(5)
	_, err := coordinator.GetBuildStatus("nonexistent-build")
	if err == nil {
		t.Error("Expected error for non-existent build")
	}
	expectedErr := "build nonexistent-build not found"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%v'", expectedErr, err)
	}
}

func TestGetWorkers(t *testing.T) {
	coordinator := NewBuildCoordinator(5)

	// Register some workers
	worker1 := &Worker{ID: "worker-1", Host: "localhost", Port: 8080, Status: "idle"}
	worker2 := &Worker{ID: "worker-2", Host: "localhost", Port: 8081, Status: "busy"}
	coordinator.RegisterWorker(worker1)
	coordinator.RegisterWorker(worker2)

	workers := coordinator.GetWorkers()
	if len(workers) != 2 {
		t.Errorf("Expected 2 workers, got %d", len(workers))
	}

	// Verify worker details
	foundWorker1 := false
	foundWorker2 := false
	for _, worker := range workers {
		if worker.ID == "worker-1" && worker.Host == "localhost" && worker.Port == 8080 {
			foundWorker1 = true
		}
		if worker.ID == "worker-2" && worker.Host == "localhost" && worker.Port == 8081 {
			foundWorker2 = true
		}
	}
	if !foundWorker1 {
		t.Error("Worker-1 not found in GetWorkers result")
	}
	if !foundWorker2 {
		t.Error("Worker-2 not found in GetWorkers result")
	}
}

func TestHandleHealth(t *testing.T) {
	coordinator := NewBuildCoordinator(5)
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	coordinator.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if status, ok := response["status"]; !ok || status != "healthy" {
		t.Errorf("Expected status healthy, got %v", response)
	}
}

func TestHandleBuilds_POST(t *testing.T) {
	coordinator := NewBuildCoordinator(5)
	requestBody := map[string]interface{}{
		"project_path":  "/test/project",
		"task_name":     "build",
		"cache_enabled": true,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/builds", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	coordinator.HandleBuilds(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if buildID, ok := response["build_id"]; !ok || buildID == "" {
		t.Errorf("Expected non-empty build_id, got %v", response)
	}
}

func TestHandleBuilds_POST_InvalidJSON(t *testing.T) {
	coordinator := NewBuildCoordinator(5)
	req := httptest.NewRequest("POST", "/api/builds", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	coordinator.HandleBuilds(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleBuilds_GET(t *testing.T) {
	coordinator := NewBuildCoordinator(5)

	// First submit a build
	request := types.BuildRequest{ProjectPath: "/test/project", TaskName: "build"}
	buildID, _ := coordinator.SubmitBuild(request)

	// Now get its status
	req := httptest.NewRequest("GET", "/api/builds?id="+buildID, nil)
	w := httptest.NewRecorder()

	coordinator.HandleBuilds(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response types.BuildResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response.RequestID != buildID {
		t.Errorf("Expected RequestID %s, got %s", buildID, response.RequestID)
	}
}

func TestHandleBuilds_GET_MissingID(t *testing.T) {
	coordinator := NewBuildCoordinator(5)
	req := httptest.NewRequest("GET", "/api/builds", nil)
	w := httptest.NewRecorder()

	coordinator.HandleBuilds(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleBuilds_GET_NotFound(t *testing.T) {
	coordinator := NewBuildCoordinator(5)
	req := httptest.NewRequest("GET", "/api/builds?id=nonexistent", nil)
	w := httptest.NewRecorder()

	coordinator.HandleBuilds(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestHandleBuilds_MethodNotAllowed(t *testing.T) {
	coordinator := NewBuildCoordinator(5)
	req := httptest.NewRequest("PUT", "/api/builds", nil)
	w := httptest.NewRecorder()

	coordinator.HandleBuilds(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleWorkers(t *testing.T) {
	coordinator := NewBuildCoordinator(5)

	// Register some workers
	worker1 := &Worker{ID: "worker-1", Host: "localhost", Port: 8080}
	worker2 := &Worker{ID: "worker-2", Host: "localhost", Port: 8081}
	coordinator.RegisterWorker(worker1)
	coordinator.RegisterWorker(worker2)

	req := httptest.NewRequest("GET", "/api/workers", nil)
	w := httptest.NewRecorder()

	coordinator.HandleWorkers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var workers []Worker
	if err := json.NewDecoder(w.Body).Decode(&workers); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if len(workers) != 2 {
		t.Errorf("Expected 2 workers, got %d", len(workers))
	}
}

func TestHandleWorkers_MethodNotAllowed(t *testing.T) {
	coordinator := NewBuildCoordinator(5)
	req := httptest.NewRequest("POST", "/api/workers", nil)
	w := httptest.NewRecorder()

	coordinator.HandleWorkers(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleStatus(t *testing.T) {
	coordinator := NewBuildCoordinator(5)

	// Register a worker
	worker := &Worker{ID: "worker-1", Host: "localhost", Port: 8080}
	coordinator.RegisterWorker(worker)

	// Submit a build
	request := types.BuildRequest{ProjectPath: "/test/project", TaskName: "build"}
	coordinator.SubmitBuild(request)

	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	coordinator.HandleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var status map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if workers, ok := status["workers"].(float64); !ok || int(workers) != 1 {
		t.Errorf("Expected 1 worker, got %v", status["workers"])
	}
	if queue, ok := status["queue"].(float64); !ok || int(queue) != 1 {
		t.Errorf("Expected queue length 1, got %v", status["queue"])
	}
	if builds, ok := status["builds"].(float64); !ok || int(builds) != 1 {
		t.Errorf("Expected 1 build, got %v", status["builds"])
	}
}

func TestGenerateBuildID(t *testing.T) {
	id1 := generateBuildID()
	time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	id2 := generateBuildID()

	if id1 == id2 {
		t.Error("Expected unique build IDs")
	}
	if len(id1) == 0 {
		t.Error("Expected non-empty build ID")
	}
	if len(id2) == 0 {
		t.Error("Expected non-empty build ID")
	}
}

func TestShutdown(t *testing.T) {
	coordinator := NewBuildCoordinator(5)

	// Shutdown should not panic even without server started
	err := coordinator.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}
