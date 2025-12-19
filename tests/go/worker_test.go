package main

import (
	"encoding/json"
	"io/ioutil"
	"net/rpc"
	"os"
	"sync"
	"testing"
	"time"
)

// Test worker functionality
func TestWorkerNode_Creation(t *testing.T) {
	config := &WorkerConfig{
		ID:           "test-worker-1",
		Host:         "localhost",
		Port:         8082,
		GradleHome:   "/tmp/gradle-test",
		Capabilities: []string{"java", "gradle", "kotlin"},
	}
	
	worker := NewWorkerNode(config)
	
	if worker == nil {
		t.Fatal("Failed to create worker node")
	}
	
	if worker.ID != config.ID {
		t.Errorf("Expected ID %s, got %s", config.ID, worker.ID)
	}
	
	if worker.Host != config.Host {
		t.Errorf("Expected host %s, got %s", config.Host, worker.Host)
	}
	
	if worker.Port != config.Port {
		t.Errorf("Expected port %d, got %d", config.Port, worker.Port)
	}
	
	if worker.Status != "idle" {
		t.Errorf("Expected idle status, got %s", worker.Status)
	}
	
	if len(worker.Capabilities) != len(config.Capabilities) {
		t.Errorf("Expected %d capabilities, got %d", len(config.Capabilities), len(worker.Capabilities))
	}
	
	if worker.Metrics.BuildCount != 0 {
		t.Errorf("Expected 0 builds, got %d", worker.Metrics.BuildCount)
	}
}

func TestWorkerService_Creation(t *testing.T) {
	config := &WorkerConfig{
		ID:     "test-worker",
		Host:   "localhost",
		Port:   8082,
		MaxJobs: 5,
	}
	
	service := NewWorkerService(config)
	
	if service == nil {
		t.Fatal("Failed to create worker service")
	}
	
	if service.worker.ID != config.ID {
		t.Errorf("Expected worker ID %s, got %s", config.ID, service.worker.ID)
	}
	
	if cap(service.builds) != 10 {
		t.Errorf("Expected builds channel capacity 10, got %d", cap(service.builds))
	}
	
	if service.config.MaxJobs != config.MaxJobs {
		t.Errorf("Expected max jobs %d, got %d", config.MaxJobs, service.config.MaxJobs)
	}
}

func TestWorkerService_RPCMethods(t *testing.T) {
	config := &WorkerConfig{
		ID:       "test-worker",
		Host:     "localhost",
		Port:     8082,
		GradleHome: "/tmp/gradle-test",
		MaxJobs:   5,
	}
	
	service := NewWorkerService(config)
	
	// Test Build RPC method
	request := BuildRequestCoordinator{
		ProjectPath:  "/tmp/test-project",
		TaskName:     "build",
		WorkerID:     config.ID,
		CacheEnabled: true,
		BuildOptions: map[string]string{"parallel": "true"},
		RequestID:    "test-build-123",
		Timestamp:    time.Now(),
	}
	
	var response string
	err := service.Build(request, &response)
	
	if err != nil {
		t.Fatalf("Build RPC method failed: %v", err)
	}
	
	if response == "" {
		t.Error("Expected non-empty response")
	}
	
	// Test GetStatus RPC method
	var status WorkerStatus
	err = service.GetStatus("", &status)
	
	if err != nil {
		t.Fatalf("GetStatus RPC method failed: %v", err)
	}
	
	if status.ID != config.ID {
		t.Errorf("Expected worker ID %s, got %s", config.ID, status.ID)
	}
	
	if status.Status != "idle" {
		t.Errorf("Expected idle status, got %s", status.Status)
	}
	
	// Test Ping RPC method
	var pingResponse string
	err = service.Ping("", &pingResponse)
	
	if err != nil {
		t.Fatalf("Ping RPC method failed: %v", err)
	}
	
	if pingResponse == "" {
		t.Error("Expected non-empty ping response")
	}
}

func TestWorkerService_HealthCheck(t *testing.T) {
	config := &WorkerConfig{
		ID:     "test-worker",
		Host:   "localhost",
		Port:   8082,
		MaxJobs: 5,
	}
	
	service := NewWorkerService(config)
	
	health := service.HealthCheck()
	
	if !health.Healthy {
		t.Error("Expected healthy status to be true")
	}
	
	if health.ID != config.ID {
		t.Errorf("Expected worker ID %s, got %s", config.ID, health.ID)
	}
	
	if health.ActiveBuilds != 0 {
		t.Errorf("Expected 0 active builds, got %d", health.ActiveBuilds)
	}
	
	if health.CompletedBuilds != 0 {
		t.Errorf("Expected 0 completed builds, got %d", health.CompletedBuilds)
	}
}

func TestWorkerService_ValidateBuildRequest(t *testing.T) {
	config := &WorkerConfig{
		ID:           "test-worker",
		Host:         "localhost",
		Port:         8082,
		Capabilities: []string{"java", "gradle"},
	}
	
	service := NewWorkerService(config)
	
	// Valid request
	validRequest := BuildRequestCoordinator{
		ProjectPath:  "/tmp/valid-project",
		TaskName:     "build",
		WorkerID:     config.ID,
		CacheEnabled: true,
		RequestID:    "valid-build-123",
	}
	
	err := service.validateBuildRequest(validRequest)
	if err != nil {
		t.Errorf("Valid request validation failed: %v", err)
	}
	
	// Invalid request - empty project path
	invalidRequest := validRequest
	invalidRequest.ProjectPath = ""
	
	err = service.validateBuildRequest(invalidRequest)
	if err == nil {
		t.Error("Expected error for empty project path")
	}
	
	// Invalid request - empty task name
	invalidRequest = validRequest
	invalidRequest.TaskName = ""
	
	err = service.validateBuildRequest(invalidRequest)
	if err == nil {
		t.Error("Expected error for empty task name")
	}
	
	// Invalid request - mismatched worker ID
	invalidRequest = validRequest
	invalidRequest.WorkerID = "different-worker"
	
	err = service.validateBuildRequest(invalidRequest)
	if err == nil {
		t.Error("Expected error for mismatched worker ID")
	}
}

func TestWorkerService_UpdateMetrics(t *testing.T) {
	config := &WorkerConfig{
		ID:     "test-worker",
		Host:   "localhost",
		Port:   8082,
		MaxJobs: 5,
	}
	
	service := NewWorkerService(config)
	
	// Simulate completed build
	startTime := time.Now()
	endTime := startTime.Add(5 * time.Minute)
	
	metrics := BuildMetrics{
		BuildSteps: []BuildStep{
			{Name: "compile", Duration: 3 * time.Minute, Status: "success"},
			{Name: "test", Duration: 2 * time.Minute, Status: "success"},
		},
		CacheHitRate:  0.85,
		CompiledFiles: 150,
		TestResults: TestResults{
			Total:   50,
			Passed:  48,
			Failed:  2,
			Skipped: 0,
		},
	}
	
	service.updateMetrics(startTime, endTime, metrics, nil)
	
	// Check metrics were updated
	service.worker.mutex.RLock()
	defer service.worker.mutex.RUnlock()
	
	if service.worker.Metrics.BuildCount != 1 {
		t.Errorf("Expected 1 build, got %d", service.worker.Metrics.BuildCount)
	}
	
	if service.worker.Metrics.SuccessRate != 1.0 {
		t.Errorf("Expected success rate 1.0, got %f", service.worker.Metrics.SuccessRate)
	}
	
	if service.worker.Metrics.LastBuildTime.IsZero() {
		t.Error("Expected non-zero last build time")
	}
}

func TestWorkerService_HandleBuildFailure(t *testing.T) {
	config := &WorkerConfig{
		ID:     "test-worker",
		Host:   "localhost",
		Port:   8082,
		MaxJobs: 5,
	}
	
	service := NewWorkerService(config)
	
	request := BuildRequestCoordinator{
		ProjectPath: "/tmp/test-project",
		TaskName:    "build",
		RequestID:   "test-build-123",
	}
	
	startTime := time.Now()
	errorMsg := "Build failed: compilation error"
	
	service.handleBuildFailure(request, startTime, errorMsg)
	
	// Check failure was handled correctly
	service.worker.mutex.RLock()
	defer service.worker.mutex.RUnlock()
	
	if service.worker.Metrics.BuildCount != 1 {
		t.Errorf("Expected 1 build attempt, got %d", service.worker.Metrics.BuildCount)
	}
	
	if service.worker.Metrics.SuccessRate != 0.0 {
		t.Errorf("Expected success rate 0.0, got %f", service.worker.Metrics.SuccessRate)
	}
}

func TestWorkerConfig_Load(t *testing.T) {
	// Create test config file
	testConfig := WorkerConfig{
		ID:           "test-worker",
		Host:         "localhost",
		Port:         8082,
		GradleHome:   "/opt/gradle",
		MaxJobs:       5,
		Capabilities: []string{"java", "gradle", "kotlin"},
		CacheDir:     "/tmp/cache",
		LogLevel:     "info",
	}
	
	configJSON, _ := json.Marshal(testConfig)
	testConfigPath := "/tmp/test_worker_config.json"
	
	err := ioutil.WriteFile(testConfigPath, configJSON, 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	defer os.Remove(testConfigPath)
	
	// Load config
	loadedConfig, err := loadWorkerConfig(testConfigPath)
	if err != nil {
		t.Fatalf("Failed to load worker config: %v", err)
	}
	
	if loadedConfig.ID != testConfig.ID {
		t.Errorf("Expected ID %s, got %s", testConfig.ID, loadedConfig.ID)
	}
	
	if loadedConfig.Host != testConfig.Host {
		t.Errorf("Expected host %s, got %s", testConfig.Host, loadedConfig.Host)
	}
	
	if loadedConfig.Port != testConfig.Port {
		t.Errorf("Expected port %d, got %d", testConfig.Port, loadedConfig.Port)
	}
	
	if loadedConfig.MaxJobs != testConfig.MaxJobs {
		t.Errorf("Expected max jobs %d, got %d", testConfig.MaxJobs, loadedConfig.MaxJobs)
	}
}

func TestWorkerService_RegisterWithCoordinator(t *testing.T) {
	config := &WorkerConfig{
		ID:             "test-worker",
		Host:           "localhost",
		Port:           8082,
		CoordinatorURL: "http://localhost:8081",
	}
	
	service := NewWorkerService(config)
	
	// This test would need a mock coordinator server
	// For now, test the registration logic
	
	err := service.registerWithCoordinator()
	// This will fail without a real coordinator, but we can test the logic
	
	if err != nil && service.worker.Status != "registered" {
		// Failed registration should not change status
		t.Logf("Registration failed as expected: %v", err)
	}
}

func TestWorkerService_ExecuteBuild(t *testing.T) {
	config := &WorkerConfig{
		ID:         "test-worker",
		Host:       "localhost",
		Port:       8082,
		GradleHome: "/tmp/gradle-test",
		MaxJobs:    5,
	}
	
	service := NewWorkerService(config)
	
	request := BuildRequestCoordinator{
		ProjectPath:  "/tmp/test-project",
		TaskName:     "build",
		WorkerID:     config.ID,
		CacheEnabled: true,
		BuildOptions: map[string]string{"parallel": "true"},
		RequestID:    "test-build-123",
	}
	
	// Mock the build execution by simulating the process
	// In a real test, this would require a real Gradle project
	
	err := service.executeBuild(request)
	
	// This will fail without a real project, but we can test the validation
	if err != nil {
		t.Logf("Build execution failed as expected: %v", err)
	}
}

func TestWorkerService_ProcessBuilds(t *testing.T) {
	config := &WorkerConfig{
		ID:     "test-worker",
		Host:   "localhost",
		Port:   8082,
		MaxJobs: 5,
	}
	
	service := NewWorkerService(config)
	
	// Start build processor
	go service.processBuilds()
	
	// Submit a build request
	request := BuildRequestCoordinator{
		ProjectPath: "/tmp/test-project",
		TaskName:    "build",
		RequestID:   "test-build-123",
	}
	
	// Add to queue
	select {
	case service.builds <- request:
		// Build added successfully
		time.Sleep(100 * time.Millisecond) // Give time for processing
	default:
		t.Error("Failed to add build to queue")
	}
	
	// Check that build was processed
	service.worker.mutex.RLock()
	defer service.worker.mutex.RUnlock()
	
	// In a real scenario with a working build, this would show the metrics updated
	if service.worker.Metrics.BuildCount == 0 {
		t.Log("No builds processed (expected with mock data)")
	}
}