package worker

import (
	"testing"

	"distributed-gradle-building/types"
	"distributed-gradle-building/workerpkg"
)

// Test worker service functionality
func TestWorkerService_Creation(t *testing.T) {
	config := types.WorkerConfig{
		ID:         "worker-1",
		HTTPPort:   8081,
		RPCPort:    8082,
		BuildDir:   "/tmp/builds",
		WorkerType: "java",
	}

	worker := workerpkg.NewWorkerService("worker-1", "localhost:8080", config)

	if worker == nil {
		t.Fatal("Failed to create worker service")
	}

	if worker.ID != "worker-1" {
		t.Errorf("Expected worker ID worker-1, got %s", worker.ID)
	}

	if worker.Coordinator != "localhost:8080" {
		t.Errorf("Expected coordinator localhost:8080, got %s", worker.Coordinator)
	}
}

func TestWorkerService_Ping(t *testing.T) {
	config := types.WorkerConfig{
		ID:         "worker-1",
		HTTPPort:   8081,
		RPCPort:    8082,
		BuildDir:   "/tmp/builds",
		WorkerType: "java",
	}

	worker := workerpkg.NewWorkerService("worker-1", "localhost:8080", config)

	var args struct{}
	var reply struct{}

	err := worker.Ping(&args, &reply)
	if err != nil {
		t.Fatalf("Ping failed: %v", err)
	}

	// Check that LastPing was updated
	if worker.LastPing.IsZero() {
		t.Error("Expected LastPing to be updated")
	}
}

func TestWorkerService_GetStatus(t *testing.T) {
	config := types.WorkerConfig{
		ID:         "worker-1",
		HTTPPort:   8081,
		RPCPort:    8082,
		BuildDir:   "/tmp/builds",
		WorkerType: "java",
	}

	worker := workerpkg.NewWorkerService("worker-1", "localhost:8080", config)

	var args struct{}
	var status workerpkg.WorkerStatus

	err := worker.GetStatus(&args, &status)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.ID != "worker-1" {
		t.Errorf("Expected worker ID worker-1, got %s", status.ID)
	}

	if status.QueueLength != 0 {
		t.Errorf("Expected queue length 0, got %d", status.QueueLength)
	}

	if !status.IsHealthy {
		t.Error("Expected worker to be healthy")
	}

	if status.WorkerType != "java" {
		t.Errorf("Expected worker type java, got %s", status.WorkerType)
	}
}

func TestWorkerService_Shutdown(t *testing.T) {
	config := types.WorkerConfig{
		ID:         "worker-1",
		HTTPPort:   8081,
		RPCPort:    8082,
		BuildDir:   "/tmp/builds",
		WorkerType: "java",
	}

	worker := workerpkg.NewWorkerService("worker-1", "localhost:8080", config)

	err := worker.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestWorkerService_RPCRegistration(t *testing.T) {
	config := types.WorkerConfig{
		ID:         "worker-1",
		HTTPPort:   8081,
		RPCPort:    8082,
		BuildDir:   "/tmp/builds",
		WorkerType: "java",
	}

	worker := workerpkg.NewWorkerService("worker-1", "localhost:8080", config)

	// This should not panic
	worker.RegisterRPCMethods()
}

func TestWorkerService_ExecuteBuild(t *testing.T) {
	config := types.WorkerConfig{
		ID:         "worker-1",
		HTTPPort:   8081,
		RPCPort:    8082,
		BuildDir:   "/tmp/builds",
		WorkerType: "java",
	}

	worker := workerpkg.NewWorkerService("worker-1", "localhost:8080", config)

	request := types.BuildRequest{
		RequestID:    "test-build-1",
		ProjectPath:  "/tmp",
		TaskName:     "help",
		CacheEnabled: true,
		BuildOptions: map[string]string{"info": "true"},
	}

	var response types.BuildResponse
	err := worker.ExecuteBuild(request, &response)

	if err != nil {
		t.Logf("Build execution failed (expected if gradle not installed): %v", err)
	}

	if response.RequestID != request.RequestID {
		t.Errorf("Expected request ID %s, got %s", request.RequestID, response.RequestID)
	}

	if response.WorkerID != "worker-1" {
		t.Errorf("Expected worker ID worker-1, got %s", response.WorkerID)
	}

	if response.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
}
