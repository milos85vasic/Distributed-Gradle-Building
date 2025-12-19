package coordinator

import (
	"testing"
	
	"distributed-gradle-building/coordinatorpkg"
	"distributed-gradle-building/types"
)

// Test coordinator functionality
func TestBuildCoordinator_Creation(t *testing.T) {
	coord := coordinatorpkg.NewBuildCoordinator(5)
	
	if coord == nil {
		t.Fatal("Failed to create coordinator")
	}
}

func TestBuildCoordinator_WorkerRegistration(t *testing.T) {
	coord := coordinatorpkg.NewBuildCoordinator(5)
	
	worker := &coordinatorpkg.Worker{
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
	coord := coordinatorpkg.NewBuildCoordinator(5)
	
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