package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

// TestBuildSubmission tests the build submission endpoint
func TestBuildSubmission(t *testing.T) {
	// Create coordinator instance
	coordinator := &Coordinator{
		BuildQueue:    make(chan *BuildRequest, 10),
		WorkerPool:    NewWorkerPool(3),
		ActiveBuilds:  make(map[string]*Build),
		CompletedBuilds: make([]*Build, 0),
	}
	
	// Create test build request
	buildReq := BuildRequest{
		ProjectPath: "/tmp/test-project",
		TaskName:    "build",
		CacheEnabled: true,
		BuildOptions: map[string]string{
			"parallel":   "true",
			"maxWorkers": "4",
		},
	}
	
	// Convert to JSON
	reqBody, err := json.Marshal(buildReq)
	if err != nil {
		t.Fatalf("Failed to marshal build request: %v", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequest("POST", "/api/build", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Create router and handler
	router := mux.NewRouter()
	router.HandleFunc("/api/build", coordinator.SubmitBuildHandler).Methods("POST")
	
	// Serve the request
	router.ServeHTTP(rr, req)
	
	// Check status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, status)
	}
	
	// Check response body
	var response BuildResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	// Verify response
	if response.BuildID == "" {
		t.Error("BuildID should not be empty")
	}
	
	if response.Status != "queued" {
		t.Errorf("Expected status 'queued', got '%s'", response.Status)
	}
}

// TestBuildStatus tests the build status endpoint
func TestBuildStatus(t *testing.T) {
	// Create coordinator instance with a test build
	coordinator := &Coordinator{
		BuildQueue:    make(chan *BuildRequest, 10),
		WorkerPool:    NewWorkerPool(3),
		ActiveBuilds:  make(map[string]*Build),
		CompletedBuilds: make([]*Build, 0),
	}
	
	// Add a test build
	testBuildID := "test-build-123"
	coordinator.ActiveBuilds[testBuildID] = &Build{
		ID:         testBuildID,
		Status:     "running",
		Progress:   0.5,
		StartTime:  time.Now(),
		ProjectPath: "/tmp/test-project",
		TaskName:   "build",
	}
	
	// Create HTTP request
	req, err := http.NewRequest("GET", "/api/build/"+testBuildID, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Create router and handler
	router := mux.NewRouter()
	router.HandleFunc("/api/build/{buildId}", coordinator.GetBuildStatusHandler).Methods("GET")
	
	// Serve the request
	router.ServeHTTP(rr, req)
	
	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}
	
	// Check response body
	var response BuildStatusResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	// Verify response
	if response.ID != testBuildID {
		t.Errorf("Expected build ID '%s', got '%s'", testBuildID, response.ID)
	}
	
	if response.Status != "running" {
		t.Errorf("Expected status 'running', got '%s'", response.Status)
	}
	
	if response.Progress != 0.5 {
		t.Errorf("Expected progress 0.5, got %f", response.Progress)
	}
}

// TestWorkerRegistration tests worker registration
func TestWorkerRegistration(t *testing.T) {
	// Create coordinator instance
	coordinator := &Coordinator{
		BuildQueue:    make(chan *BuildRequest, 10),
		WorkerPool:    NewWorkerPool(10),
		ActiveBuilds:  make(map[string]*Build),
		CompletedBuilds: make([]*Build, 0),
	}
	
	// Create test worker
	worker := &Worker{
		ID:       "test-worker",
		Host:     "localhost",
		Port:     8082,
		Status:   "available",
		Capacity: 4,
	}
	
	// Convert to JSON
	reqBody, err := json.Marshal(worker)
	if err != nil {
		t.Fatalf("Failed to marshal worker: %v", err)
	}
	
	// Create HTTP request
	req, err := http.NewRequest("POST", "/api/workers", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Create router and handler
	router := mux.NewRouter()
	router.HandleFunc("/api/workers", coordinator.RegisterWorkerHandler).Methods("POST")
	
	// Serve the request
	router.ServeHTTP(rr, req)
	
	// Check status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, status)
	}
	
	// Check response body
	var response WorkerRegistrationResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	// Verify response
	if response.WorkerID == "" {
		t.Error("WorkerID should not be empty")
	}
	
	if !response.Success {
		t.Error("Registration should be successful")
	}
	
	// Verify worker was added to pool
	if len(coordinator.WorkerPool.Workers) != 1 {
		t.Errorf("Expected 1 worker in pool, got %d", len(coordinator.WorkerPool.Workers))
	}
}

// TestWorkerStatus tests worker status endpoint
func TestWorkerStatus(t *testing.T) {
	// Create coordinator instance with test workers
	coordinator := &Coordinator{
		BuildQueue:    make(chan *BuildRequest, 10),
		WorkerPool:    NewWorkerPool(3),
		ActiveBuilds:  make(map[string]*Build),
		CompletedBuilds: make([]*Build, 0),
	}
	
	// Add test workers
	workers := []*Worker{
		{
			ID:         "worker-1",
			Host:       "localhost",
			Port:       8082,
			Status:     "busy",
			ActiveTask: "compileJava",
		},
		{
			ID:     "worker-2",
			Host:   "localhost",
			Port:   8083,
			Status: "available",
		},
		{
			ID:     "worker-3",
			Host:   "localhost",
			Port:   8084,
			Status: "offline",
		},
	}
	
	for _, w := range workers {
		coordinator.WorkerPool.AddWorker(w)
	}
	
	// Create HTTP request
	req, err := http.NewRequest("GET", "/api/workers", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Create router and handler
	router := mux.NewRouter()
	router.HandleFunc("/api/workers", coordinator.GetWorkersHandler).Methods("GET")
	
	// Serve the request
	router.ServeHTTP(rr, req)
	
	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}
	
	// Check response body
	var response WorkersStatusResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	// Verify response
	if len(response.Workers) != 3 {
		t.Errorf("Expected 3 workers, got %d", len(response.Workers))
	}
	
	// Check specific worker statuses
	workerStatuses := make(map[string]Worker)
	for _, w := range response.Workers {
		workerStatuses[w.ID] = w
	}
	
	if workerStatuses["worker-1"].Status != "busy" {
		t.Errorf("Expected worker-1 status 'busy', got '%s'", workerStatuses["worker-1"].Status)
	}
	
	if workerStatuses["worker-2"].Status != "available" {
		t.Errorf("Expected worker-2 status 'available', got '%s'", workerStatuses["worker-2"].Status)
	}
	
	if workerStatuses["worker-3"].Status != "offline" {
		t.Errorf("Expected worker-3 status 'offline', got '%s'", workerStatuses["worker-3"].Status)
	}
}

// TestBuildCancellation tests build cancellation
func TestBuildCancellation(t *testing.T) {
	// Create coordinator instance with a test build
	coordinator := &Coordinator{
		BuildQueue:    make(chan *BuildRequest, 10),
		WorkerPool:    NewWorkerPool(3),
		ActiveBuilds:  make(map[string]*Build),
		CompletedBuilds: make([]*Build, 0),
	}
	
	// Add a test build
	testBuildID := "test-build-123"
	coordinator.ActiveBuilds[testBuildID] = &Build{
		ID:         testBuildID,
		Status:     "running",
		Progress:   0.5,
		StartTime:  time.Now(),
		ProjectPath: "/tmp/test-project",
		TaskName:   "build",
	}
	
	// Create HTTP request
	req, err := http.NewRequest("DELETE", "/api/build/"+testBuildID, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Create router and handler
	router := mux.NewRouter()
	router.HandleFunc("/api/build/{buildId}", coordinator.CancelBuildHandler).Methods("DELETE")
	
	// Serve the request
	router.ServeHTTP(rr, req)
	
	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}
	
	// Check response body
	var response CancelBuildResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	// Verify response
	if !response.Success {
		t.Error("Cancellation should be successful")
	}
	
	// Verify build was cancelled
	build, exists := coordinator.ActiveBuilds[testBuildID]
	if exists {
		if build.Status != "cancelled" {
			t.Errorf("Expected build status 'cancelled', got '%s'", build.Status)
		}
	} else {
		t.Error("Build should still exist in ActiveBuilds with cancelled status")
	}
}

// TestMetricsEndpoint tests system metrics endpoint
func TestMetricsEndpoint(t *testing.T) {
	// Create coordinator instance with metrics
	coordinator := &Coordinator{
		BuildQueue:    make(chan *BuildRequest, 10),
		WorkerPool:    NewWorkerPool(3),
		ActiveBuilds:  make(map[string]*Build),
		CompletedBuilds: make([]*Build, 0),
		Metrics: &CoordinatorMetrics{
			TotalBuilds:       100,
			ActiveBuilds:      5,
			CompletedBuilds:   95,
			FailedBuilds:      10,
			AverageBuildTime:   180 * time.Second,
			TotalWorkers:      3,
			AvailableWorkers:  2,
			BusyWorkers:       1,
		},
	}
	
	// Create HTTP request
	req, err := http.NewRequest("GET", "/api/metrics", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	// Create response recorder
	rr := httptest.NewRecorder()
	
	// Create router and handler
	router := mux.NewRouter()
	router.HandleFunc("/api/metrics", coordinator.GetMetricsHandler).Methods("GET")
	
	// Serve the request
	router.ServeHTTP(rr, req)
	
	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}
	
	// Check response body
	var response MetricsResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	// Verify metrics
	if response.TotalBuilds != 100 {
		t.Errorf("Expected TotalBuilds 100, got %d", response.TotalBuilds)
	}
	
	if response.ActiveBuilds != 5 {
		t.Errorf("Expected ActiveBuilds 5, got %d", response.ActiveBuilds)
	}
	
	if response.CompletedBuilds != 95 {
		t.Errorf("Expected CompletedBuilds 95, got %d", response.CompletedBuilds)
	}
	
	if response.TotalWorkers != 3 {
		t.Errorf("Expected TotalWorkers 3, got %d", response.TotalWorkers)
	}
}

// TestBuildPriorityQueue tests build priority queue functionality
func TestBuildPriorityQueue(t *testing.T) {
	// Create coordinator instance
	coordinator := &Coordinator{
		BuildQueue:    make(chan *BuildRequest, 10),
		WorkerPool:    NewWorkerPool(3),
		ActiveBuilds:  make(map[string]*Build),
		CompletedBuilds: make([]*Build, 0),
		PriorityQueue:  make([]*BuildRequest, 0),
	}
	
	// Add builds with different priorities
	builds := []*BuildRequest{
		{
			ID:       "build-1",
			Priority: 1, // Low priority
			TaskName: "test",
		},
		{
			ID:       "build-2",
			Priority: 3, // High priority
			TaskName: "build",
		},
		{
			ID:       "build-3",
			Priority: 2, // Medium priority
			TaskName: "assemble",
		},
	}
	
	// Add builds to queue
	for _, build := range builds {
		coordinator.EnqueueBuild(build)
	}
	
	// Verify queue ordering
	if len(coordinator.PriorityQueue) != 3 {
		t.Errorf("Expected queue length 3, got %d", len(coordinator.PriorityQueue))
	}
	
	// Check that high priority build is first
	if coordinator.PriorityQueue[0].ID != "build-2" {
		t.Errorf("Expected first build to be 'build-2', got '%s'", coordinator.PriorityQueue[0].ID)
	}
	
	// Check medium priority
	if coordinator.PriorityQueue[1].ID != "build-3" {
		t.Errorf("Expected second build to be 'build-3', got '%s'", coordinator.PriorityQueue[1].ID)
	}
	
	// Check low priority
	if coordinator.PriorityQueue[2].ID != "build-1" {
		t.Errorf("Expected third build to be 'build-1', got '%s'", coordinator.PriorityQueue[2].ID)
	}
}

// TestWorkerSelection tests worker selection logic
func TestWorkerSelection(t *testing.T) {
	// Create coordinator instance
	coordinator := &Coordinator{
		BuildQueue:    make(chan *BuildRequest, 10),
		WorkerPool:    NewWorkerPool(3),
		ActiveBuilds:  make(map[string]*Build),
		CompletedBuilds: make([]*Build, 0),
	}
	
	// Add test workers with different loads
	workers := []*Worker{
		{
			ID:         "worker-1",
			Status:     "available",
			Capacity:   4,
			ActiveTask: "",
			Load:       0.2,
		},
		{
			ID:         "worker-2",
			Status:     "available",
			Capacity:   4,
			ActiveTask: "",
			Load:       0.8,
		},
		{
			ID:         "worker-3",
			Status:     "busy",
			Capacity:   4,
			ActiveTask: "compileJava",
			Load:       0.5,
		},
	}
	
	for _, w := range workers {
		coordinator.WorkerPool.AddWorker(w)
	}
	
	// Test worker selection for compile task
	selectedWorker := coordinator.SelectWorker("compileJava")
	
	// Should select the least loaded available worker
	if selectedWorker == nil {
		t.Error("Should select a worker")
	}
	
	if selectedWorker.ID != "worker-1" {
		t.Errorf("Expected to select 'worker-1' (least loaded), got '%s'", selectedWorker.ID)
	}
	
	// Test worker selection when all workers are busy
	for _, w := range workers {
		w.Status = "busy"
	}
	
	selectedWorker = coordinator.SelectWorker("compileJava")
	if selectedWorker != nil {
		t.Error("Should return nil when all workers are busy")
	}
}

// TestConfigurationValidation tests configuration validation
func TestConfigurationValidation(t *testing.T) {
	// Test valid configuration
	validConfig := CoordinatorConfig{
		Port:             8080,
		MaxWorkers:       10,
		BuildTimeout:      30 * time.Minute,
		WorkerHeartbeat:   30 * time.Second,
		MetricsInterval:   60 * time.Second,
		CleanupInterval:  1 * time.Hour,
	}
	
	err := validConfig.Validate()
	if err != nil {
		t.Errorf("Valid config should not return error: %v", err)
	}
	
	// Test invalid configuration
	invalidConfigs := []struct {
		name string
		config CoordinatorConfig
	}{
		{
			name: "Invalid port",
			config: CoordinatorConfig{
				Port: 0,
			},
		},
		{
			name: "Negative max workers",
			config: CoordinatorConfig{
				Port:       8080,
				MaxWorkers: -1,
			},
		},
		{
			name: "Zero build timeout",
			config: CoordinatorConfig{
				Port:         8080,
				MaxWorkers:   10,
				BuildTimeout: 0,
			},
		},
	}
	
	for _, tc := range invalidConfigs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if err == nil {
				t.Errorf("Invalid config should return error for %s", tc.name)
			}
		})
	}
}

// TestErrorHandling tests various error conditions
func TestErrorHandling(t *testing.T) {
	// Create coordinator instance
	coordinator := &Coordinator{
		BuildQueue:    make(chan *BuildRequest, 10),
		WorkerPool:    NewWorkerPool(3),
		ActiveBuilds:  make(map[string]*Build),
		CompletedBuilds: make([]*Build, 0),
	}
	
	testCases := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
	}{
		{
			name:           "Invalid build request",
			method:         "POST",
			path:           "/api/build",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing project path",
			method:         "POST",
			path:           "/api/build",
			body:           BuildRequest{TaskName: "build"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent build",
			method:         "GET",
			path:           "/api/build/non-existent",
			body:           nil,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Invalid worker registration",
			method:         "POST",
			path:           "/api/workers",
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			var err error
			
			if tc.body != nil {
				var reqBody []byte
				if str, ok := tc.body.(string); ok {
					reqBody = []byte(str)
				} else {
					reqBody, err = json.Marshal(tc.body)
					if err != nil {
						t.Fatalf("Failed to marshal request body: %v", err)
					}
				}
				
				req, err = http.NewRequest(tc.method, tc.path, bytes.NewBuffer(reqBody))
				if err != nil {
					t.Fatalf("Failed to create request: %v", err)
				}
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(tc.method, tc.path, nil)
				if err != nil {
					t.Fatalf("Failed to create request: %v", err)
				}
			}
			
			rr := httptest.NewRecorder()
			
			router := mux.NewRouter()
			router.HandleFunc("/api/build", coordinator.SubmitBuildHandler).Methods("POST")
			router.HandleFunc("/api/build/{buildId}", coordinator.GetBuildStatusHandler).Methods("GET")
			router.HandleFunc("/api/workers", coordinator.RegisterWorkerHandler).Methods("POST")
			
			router.ServeHTTP(rr, req)
			
			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, status)
			}
		})
	}
}

// TestConcurrentOperations tests concurrent build operations
func TestConcurrentOperations(t *testing.T) {
	// Create coordinator instance
	coordinator := &Coordinator{
		BuildQueue:    make(chan *BuildRequest, 100),
		WorkerPool:    NewWorkerPool(10),
		ActiveBuilds:  make(map[string]*Build),
		CompletedBuilds: make([]*Build, 0),
	}
	
	// Start coordinator in background
	go coordinator.Start()
	defer coordinator.Stop()
	
	// Create multiple concurrent build requests
	numBuilds := 10
	buildRequests := make(chan *BuildRequest, numBuilds)
	
	for i := 0; i < numBuilds; i++ {
		buildID := fmt.Sprintf("test-build-%d", i)
		build := &BuildRequest{
			ID:          buildID,
			ProjectPath:  "/tmp/test-project",
			TaskName:     "build",
			CacheEnabled: true,
		}
		buildRequests <- build
	}
	
	close(buildRequests)
	
	// Submit builds concurrently
	var wg sync.WaitGroup
	results := make(chan string, numBuilds)
	
	for build := range buildRequests {
		wg.Add(1)
		go func(b *BuildRequest) {
			defer wg.Done()
			resp, err := coordinator.SubmitBuild(b)
			if err != nil {
				t.Errorf("Failed to submit build %s: %v", b.ID, err)
				return
			}
			results <- resp.BuildID
		}(build)
	}
	
	wg.Wait()
	close(results)
	
	// Verify all builds were submitted
	submittedBuilds := make([]string, 0, numBuilds)
	for buildID := range results {
		submittedBuilds = append(submittedBuilds, buildID)
	}
	
	if len(submittedBuilds) != numBuilds {
		t.Errorf("Expected %d builds to be submitted, got %d", numBuilds, len(submittedBuilds))
	}
	
	// Verify all builds have unique IDs
	uniqueIDs := make(map[string]bool)
	for _, id := range submittedBuilds {
		if uniqueIDs[id] {
			t.Errorf("Duplicate build ID found: %s", id)
		}
		uniqueIDs[id] = true
	}
}

// TestPersistence tests build state persistence
func TestPersistence(t *testing.T) {
	// Create temporary directory for test data
	tempDir, err := os.MkdirTemp("", "coordinator-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create coordinator with persistence enabled
	config := CoordinatorConfig{
		PersistenceEnabled: true,
		PersistencePath:   filepath.Join(tempDir, "state.json"),
	}
	
	coordinator := &Coordinator{
		Config:           config,
		BuildQueue:       make(chan *BuildRequest, 10),
		WorkerPool:       NewWorkerPool(3),
		ActiveBuilds:     make(map[string]*Build),
		CompletedBuilds:  make([]*Build, 0),
	}
	
	// Add some test data
	testBuildID := "test-build-123"
	coordinator.ActiveBuilds[testBuildID] = &Build{
		ID:         testBuildID,
		Status:     "running",
		Progress:   0.5,
		StartTime:  time.Now(),
		ProjectPath: "/tmp/test-project",
		TaskName:   "build",
	}
	
	// Save state
	err = coordinator.SaveState()
	if err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}
	
	// Verify state file exists
	stateFile := filepath.Join(tempDir, "state.json")
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Error("State file should exist after saving")
	}
	
	// Load state in new coordinator instance
	newCoordinator := &Coordinator{
		Config:          config,
		BuildQueue:       make(chan *BuildRequest, 10),
		WorkerPool:       NewWorkerPool(3),
		ActiveBuilds:     make(map[string]*Build),
		CompletedBuilds:  make([]*Build, 0),
	}
	
	err = newCoordinator.LoadState()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}
	
	// Verify loaded state
	build, exists := newCoordinator.ActiveBuilds[testBuildID]
	if !exists {
		t.Error("Build should exist in loaded state")
		return
	}
	
	if build.ID != testBuildID {
		t.Errorf("Expected build ID '%s', got '%s'", testBuildID, build.ID)
	}
	
	if build.Status != "running" {
		t.Errorf("Expected status 'running', got '%s'", build.Status)
	}
}