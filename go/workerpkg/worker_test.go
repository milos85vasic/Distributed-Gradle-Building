package workerpkg

import (
	"encoding/json"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"testing"
	"time"

	"distributed-gradle-building/types"
)

func TestNewWorkerService(t *testing.T) {
	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9091,
		BuildDir:            "/tmp/builds",
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)

	if worker.ID != "test-worker" {
		t.Errorf("Expected ID 'test-worker', got %s", worker.ID)
	}

	if worker.Coordinator != "localhost:8080" {
		t.Errorf("Expected coordinator 'localhost:8080', got %s", worker.Coordinator)
	}

	if worker.BuildDir != "/tmp/builds" {
		t.Errorf("Expected build dir '/tmp/builds', got %s", worker.BuildDir)
	}

	if cap(worker.BuildQueue) != 10 {
		t.Errorf("Expected build queue capacity 10, got %d", cap(worker.BuildQueue))
	}

	if len(worker.ActiveBuilds) != 0 {
		t.Errorf("Expected empty active builds, got %d", len(worker.ActiveBuilds))
	}
}

func TestPing(t *testing.T) {
	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9091,
		BuildDir:            "/tmp/builds",
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)
	initialPing := worker.LastPing

	time.Sleep(10 * time.Millisecond)

	var reply struct{}
	err := worker.Ping(&struct{}{}, &reply)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}

	if worker.LastPing.Equal(initialPing) {
		t.Error("LastPing should have been updated")
	}

	if worker.LastPing.Before(initialPing) {
		t.Error("LastPing should be after initial ping")
	}
}

func TestGetStatus(t *testing.T) {
	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9091,
		BuildDir:            "/tmp/builds",
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)

	var status WorkerStatus
	err := worker.GetStatus(&struct{}{}, &status)
	if err != nil {
		t.Errorf("GetStatus failed: %v", err)
	}

	if status.ID != "test-worker" {
		t.Errorf("Expected ID 'test-worker', got %s", status.ID)
	}

	if status.ActiveBuilds != 0 {
		t.Errorf("Expected 0 active builds, got %d", status.ActiveBuilds)
	}

	if status.QueueLength != 0 {
		t.Errorf("Expected queue length 0, got %d", status.QueueLength)
	}

	if status.BuildDir != "/tmp/builds" {
		t.Errorf("Expected build dir '/tmp/builds', got %s", status.BuildDir)
	}

	if !status.IsHealthy {
		t.Error("Worker should be healthy")
	}

	if status.WorkerType != "standard" {
		t.Errorf("Expected worker type 'standard', got %s", status.WorkerType)
	}
}

func TestGetStatus_WithActiveBuilds(t *testing.T) {
	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9091,
		BuildDir:            "/tmp/builds",
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)

	// Add an active build
	worker.Mutex.Lock()
	worker.ActiveBuilds["build-123"] = &types.BuildResponse{
		RequestID: "build-123",
		WorkerID:  "test-worker",
		Timestamp: time.Now(),
		Success:   true,
	}
	worker.Mutex.Unlock()

	var status WorkerStatus
	err := worker.GetStatus(&struct{}{}, &status)
	if err != nil {
		t.Errorf("GetStatus failed: %v", err)
	}

	if status.ActiveBuilds != 1 {
		t.Errorf("Expected 1 active build, got %d", status.ActiveBuilds)
	}
}

func TestGetStatus_WithQueue(t *testing.T) {
	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9091,
		BuildDir:            "/tmp/builds",
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)

	// Add items to queue
	go func() {
		worker.BuildQueue <- types.BuildRequest{
			RequestID:   "build-1",
			ProjectPath: "/tmp/project1",
			TaskName:    "build",
		}
		worker.BuildQueue <- types.BuildRequest{
			RequestID:   "build-2",
			ProjectPath: "/tmp/project2",
			TaskName:    "test",
		}
	}()

	time.Sleep(50 * time.Millisecond)

	var status WorkerStatus
	err := worker.GetStatus(&struct{}{}, &status)
	if err != nil {
		t.Errorf("GetStatus failed: %v", err)
	}

	if status.QueueLength != 2 {
		t.Errorf("Expected queue length 2, got %d", status.QueueLength)
	}
}

func TestGetStatus_Unhealthy(t *testing.T) {
	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9091,
		BuildDir:            "/tmp/builds",
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)

	// Set last ping to be more than 5 minutes ago
	worker.Mutex.Lock()
	worker.LastPing = time.Now().Add(-6 * time.Minute)
	worker.Mutex.Unlock()

	var status WorkerStatus
	err := worker.GetStatus(&struct{}{}, &status)
	if err != nil {
		t.Errorf("GetStatus failed: %v", err)
	}

	if status.IsHealthy {
		t.Error("Worker should be unhealthy")
	}
}

func TestExecuteBuild_Success(t *testing.T) {
	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9091,
		BuildDir:            t.TempDir(),
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)

	// Create a simple project directory with a build.gradle file
	projectDir := t.TempDir()
	buildGradle := filepath.Join(projectDir, "build.gradle")
	os.WriteFile(buildGradle, []byte("task hello { doLast { println 'Hello, World!' } }"), 0644)

	request := types.BuildRequest{
		RequestID:    "test-build-123",
		ProjectPath:  projectDir,
		TaskName:     "hello",
		CacheEnabled: true,
		BuildOptions: map[string]string{
			"quiet": "true",
		},
		Timestamp: time.Now(),
	}

	var response types.BuildResponse
	err := worker.ExecuteBuild(request, &response)
	if err != nil {
		t.Errorf("ExecuteBuild failed: %v", err)
	}

	if response.RequestID != "test-build-123" {
		t.Errorf("Expected request ID 'test-build-123', got %s", response.RequestID)
	}

	if response.WorkerID != "test-worker" {
		t.Errorf("Expected worker ID 'test-worker', got %s", response.WorkerID)
	}

	// Note: The build might fail if Gradle is not installed or the task fails
	// We'll check for either success or a meaningful error message
	if !response.Success && response.ErrorMessage == "" {
		t.Error("Build should either succeed or have an error message")
	}

	if response.BuildDuration <= 0 {
		t.Error("Build duration should be positive")
	}
}

func TestExecuteBuild_Failure(t *testing.T) {
	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9091,
		BuildDir:            t.TempDir(),
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)

	// Create a non-existent project path
	request := types.BuildRequest{
		RequestID:    "test-build-456",
		ProjectPath:  "/non/existent/path",
		TaskName:     "build",
		CacheEnabled: true,
		Timestamp:    time.Now(),
	}

	var response types.BuildResponse
	err := worker.ExecuteBuild(request, &response)
	if err != nil {
		t.Errorf("ExecuteBuild failed: %v", err)
	}

	if response.Success {
		t.Error("Build should have failed")
	}

	if response.ErrorMessage == "" {
		t.Error("Error message should not be empty")
	}
}

func TestExecuteBuild_InvalidProjectPath(t *testing.T) {
	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9091,
		BuildDir:            t.TempDir(),
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)

	// Use a file instead of a directory
	tmpFile := filepath.Join(t.TempDir(), "not-a-dir.txt")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	request := types.BuildRequest{
		RequestID:    "test-build-789",
		ProjectPath:  tmpFile,
		TaskName:     "build",
		CacheEnabled: true,
		Timestamp:    time.Now(),
	}

	var response types.BuildResponse
	err := worker.ExecuteBuild(request, &response)
	if err != nil {
		t.Errorf("ExecuteBuild failed: %v", err)
	}

	if response.Success {
		t.Error("Build should have failed")
	}
}

func TestCollectArtifacts(t *testing.T) {
	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9091,
		BuildDir:            t.TempDir(),
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)

	// Create test build directory with artifacts
	buildDir := filepath.Join(t.TempDir(), "test-build")
	os.MkdirAll(buildDir, 0755)

	// Create some artifact files
	artifacts := []string{
		"build/libs/app.jar",
		"build/reports/tests/test/index.html",
		"build/distributions/app.zip",
	}

	for _, artifact := range artifacts {
		artifactPath := filepath.Join(buildDir, artifact)
		os.MkdirAll(filepath.Dir(artifactPath), 0755)
		os.WriteFile(artifactPath, []byte("test content"), 0644)
	}

	// Test collectArtifacts
	collected := worker.collectArtifacts(buildDir)

	if len(collected) != len(artifacts) {
		t.Errorf("Expected %d artifacts, got %d", len(artifacts), len(collected))
	}

	// Check that all artifacts were collected
	for _, artifact := range artifacts {
		found := false
		for _, collectedArtifact := range collected {
			if collectedArtifact == artifact {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Artifact %s not collected", artifact)
		}
	}
}

func TestCollectArtifacts_EmptyDirectory(t *testing.T) {
	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9091,
		BuildDir:            t.TempDir(),
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)

	// Create empty build directory
	buildDir := filepath.Join(t.TempDir(), "empty-build")
	os.MkdirAll(buildDir, 0755)

	collected := worker.collectArtifacts(buildDir)

	if len(collected) != 0 {
		t.Errorf("Expected 0 artifacts, got %d", len(collected))
	}
}

func TestCollectArtifacts_NonExistentDirectory(t *testing.T) {
	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9091,
		BuildDir:            t.TempDir(),
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)

	// Non-existent directory
	buildDir := filepath.Join(t.TempDir(), "non-existent")

	collected := worker.collectArtifacts(buildDir)

	if len(collected) != 0 {
		t.Errorf("Expected 0 artifacts, got %d", len(collected))
	}
}

func TestShutdown(t *testing.T) {
	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9091,
		BuildDir:            "/tmp/builds",
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)

	err := worker.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Verify queue is closed
	select {
	case _, ok := <-worker.BuildQueue:
		if ok {
			t.Error("Build queue should be closed")
		}
	default:
		t.Error("Build queue should be closed")
	}
}

func TestRegisterRPCMethods(t *testing.T) {
	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9091,
		BuildDir:            "/tmp/builds",
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)

	// This should not panic
	worker.RegisterRPCMethods()
}

func TestStartServer(t *testing.T) {
	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9092, // Use different port to avoid conflicts
		BuildDir:            "/tmp/builds",
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)

	// Start server in goroutine
	go func() {
		err := worker.StartServer()
		if err != nil {
			t.Errorf("StartServer failed: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Try to connect to the RPC server
	client, err := rpc.Dial("tcp", "localhost:9092")
	if err != nil {
		t.Errorf("Failed to connect to RPC server: %v", err)
		return
	}
	defer client.Close()

	// Test Ping
	var pingReply struct{}
	err = client.Call("WorkerService.Ping", &struct{}{}, &pingReply)
	if err != nil {
		t.Errorf("Ping RPC call failed: %v", err)
	}

	// Test GetStatus
	var status WorkerStatus
	err = client.Call("WorkerService.GetStatus", &struct{}{}, &status)
	if err != nil {
		t.Errorf("GetStatus RPC call failed: %v", err)
	}

	if status.ID != "test-worker" {
		t.Errorf("Expected ID 'test-worker', got %s", status.ID)
	}
}

func TestStartServer_PortInUse(t *testing.T) {
	// First, occupy the port
	listener, err := net.Listen("tcp", ":9093")
	if err != nil {
		t.Skipf("Cannot test port in use: %v", err)
	}
	defer listener.Close()

	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9093, // This port is already in use
		BuildDir:            "/tmp/builds",
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)

	err = worker.StartServer()
	if err == nil {
		t.Error("Expected error when port is in use")
	}
}

func TestWorkerStatus_JSONSerialization(t *testing.T) {
	status := WorkerStatus{
		ID:           "worker-123",
		ActiveBuilds: 2,
		QueueLength:  5,
		LastPing:     time.Now().Truncate(time.Second),
		BuildDir:     "/tmp/builds",
		IsHealthy:    true,
		WorkerType:   "standard",
	}

	// Test marshaling
	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("Failed to marshal WorkerStatus: %v", err)
	}

	// Test unmarshaling
	var decoded WorkerStatus
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal WorkerStatus: %v", err)
	}

	if decoded.ID != status.ID {
		t.Errorf("Expected ID %s, got %s", status.ID, decoded.ID)
	}

	if decoded.ActiveBuilds != status.ActiveBuilds {
		t.Errorf("Expected ActiveBuilds %d, got %d", status.ActiveBuilds, decoded.ActiveBuilds)
	}

	if decoded.QueueLength != status.QueueLength {
		t.Errorf("Expected QueueLength %d, got %d", status.QueueLength, decoded.QueueLength)
	}

	if decoded.BuildDir != status.BuildDir {
		t.Errorf("Expected BuildDir %s, got %s", status.BuildDir, decoded.BuildDir)
	}

	if decoded.IsHealthy != status.IsHealthy {
		t.Errorf("Expected IsHealthy %v, got %v", status.IsHealthy, decoded.IsHealthy)
	}

	if decoded.WorkerType != status.WorkerType {
		t.Errorf("Expected WorkerType %s, got %s", status.WorkerType, decoded.WorkerType)
	}
}

func TestWorkerService_ConcurrentAccess(t *testing.T) {
	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9091,
		BuildDir:            t.TempDir(),
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)

	// Run concurrent operations
	done := make(chan bool)
	errors := make(chan error, 10)

	for i := 0; i < 5; i++ {
		go func(id int) {
			// Ping
			var pingReply struct{}
			err := worker.Ping(&struct{}{}, &pingReply)
			if err != nil {
				errors <- err
			}

			// GetStatus
			var status WorkerStatus
			err = worker.GetStatus(&struct{}{}, &status)
			if err != nil {
				errors <- err
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrent operation failed: %v", err)
	}
}

func TestWorkerService_ExecuteBuild_Concurrent(t *testing.T) {
	config := types.WorkerConfig{
		ID:                  "test-worker",
		CoordinatorURL:      "localhost:8080",
		HTTPPort:            8081,
		RPCPort:             9091,
		BuildDir:            t.TempDir(),
		CacheEnabled:        true,
		MaxConcurrentBuilds: 5,
		WorkerType:          "standard",
	}

	worker := NewWorkerService("test-worker", "localhost:8080", config)

	// Run concurrent builds
	done := make(chan bool)
	errors := make(chan error, 10)

	for i := 0; i < 3; i++ {
		go func(id int) {
			request := types.BuildRequest{
				RequestID:    "concurrent-build",
				ProjectPath:  t.TempDir(),
				TaskName:     "help",
				CacheEnabled: true,
				Timestamp:    time.Now(),
			}

			var response types.BuildResponse
			err := worker.ExecuteBuild(request, &response)
			if err != nil {
				errors <- err
			}

			if !response.Success && response.ErrorMessage == "" {
				errors <- err
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrent build failed: %v", err)
	}
}
