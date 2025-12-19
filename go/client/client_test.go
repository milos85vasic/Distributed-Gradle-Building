package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:8080")
	if client.BaseURL != "http://localhost:8080" {
		t.Errorf("Expected BaseURL 'http://localhost:8080', got %s", client.BaseURL)
	}
	if client.HTTPClient.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", client.HTTPClient.Timeout)
	}
	if client.AuthToken != "" {
		t.Errorf("Expected empty AuthToken, got %s", client.AuthToken)
	}
}

func TestNewAuthenticatedClient(t *testing.T) {
	client := NewAuthenticatedClient("http://localhost:8080", "test-token")
	if client.BaseURL != "http://localhost:8080" {
		t.Errorf("Expected BaseURL 'http://localhost:8080', got %s", client.BaseURL)
	}
	if client.AuthToken != "test-token" {
		t.Errorf("Expected AuthToken 'test-token', got %s", client.AuthToken)
	}
}

func TestSubmitBuild_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/api/build" {
			t.Errorf("Expected path /api/build, got %s", r.URL.Path)
		}
		if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
		}

		var req BuildRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}
		if req.ProjectPath != "/test/project" {
			t.Errorf("Expected ProjectPath /test/project, got %s", req.ProjectPath)
		}
		if req.TaskName != "build" {
			t.Errorf("Expected TaskName build, got %s", req.TaskName)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(BuildResponse{
			BuildID: "test-build-123",
			Status:  "queued",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req := BuildRequest{
		ProjectPath:  "/test/project",
		TaskName:     "build",
		CacheEnabled: true,
		BuildOptions: map[string]string{"clean": "true"},
	}

	resp, err := client.SubmitBuild(req)
	if err != nil {
		t.Fatalf("SubmitBuild failed: %v", err)
	}
	if resp.BuildID != "test-build-123" {
		t.Errorf("Expected BuildID test-build-123, got %s", resp.BuildID)
	}
	if resp.Status != "queued" {
		t.Errorf("Expected Status queued, got %s", resp.Status)
	}
}

func TestSubmitBuild_WithAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authToken := r.Header.Get("X-Auth-Token"); authToken != "test-token" {
			t.Errorf("Expected X-Auth-Token test-token, got %s", authToken)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(BuildResponse{
			BuildID: "auth-build-456",
			Status:  "queued",
		})
	}))
	defer server.Close()

	client := NewAuthenticatedClient(server.URL, "test-token")
	req := BuildRequest{
		ProjectPath: "/test/project",
		TaskName:    "build",
	}

	resp, err := client.SubmitBuild(req)
	if err != nil {
		t.Fatalf("SubmitBuild failed: %v", err)
	}
	if resp.BuildID != "auth-build-456" {
		t.Errorf("Expected BuildID auth-build-456, got %s", resp.BuildID)
	}
}

func TestSubmitBuild_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req := BuildRequest{
		ProjectPath: "/test/project",
		TaskName:    "build",
	}

	_, err := client.SubmitBuild(req)
	if err == nil {
		t.Error("Expected error for bad request, got nil")
	}
	expectedErr := "unexpected status code: 400"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%v'", expectedErr, err)
	}
}

func TestSubmitBuild_NetworkError(t *testing.T) {
	client := NewClient("http://nonexistent-server:9999")
	req := BuildRequest{
		ProjectPath: "/test/project",
		TaskName:    "build",
	}

	_, err := client.SubmitBuild(req)
	if err == nil {
		t.Error("Expected network error, got nil")
	}
}

func TestGetBuildStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/api/build/test-build-123" {
			t.Errorf("Expected path /api/build/test-build-123, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(BuildStatus{
			BuildID:      "test-build-123",
			WorkerID:     "worker-1",
			ProjectPath:  "/test/project",
			TaskName:     "build",
			Status:       "completed",
			StartTime:    time.Now().Add(-5 * time.Minute),
			EndTime:      time.Now(),
			Duration:     5 * time.Minute,
			Success:      true,
			CacheHitRate: 0.75,
			Artifacts:    []string{"app.jar", "test-results.xml"},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	status, err := client.GetBuildStatus("test-build-123")
	if err != nil {
		t.Fatalf("GetBuildStatus failed: %v", err)
	}
	if status.BuildID != "test-build-123" {
		t.Errorf("Expected BuildID test-build-123, got %s", status.BuildID)
	}
	if status.Status != "completed" {
		t.Errorf("Expected Status completed, got %s", status.Status)
	}
	if !status.Success {
		t.Error("Expected Success true")
	}
	if len(status.Artifacts) != 2 {
		t.Errorf("Expected 2 artifacts, got %d", len(status.Artifacts))
	}
}

func TestGetBuildStatus_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.GetBuildStatus("nonexistent-build")
	if err == nil {
		t.Error("Expected error for not found, got nil")
	}
	expectedErr := "unexpected status code: 404"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%v'", expectedErr, err)
	}
}

func TestGetWorkers_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/api/workers" {
			t.Errorf("Expected path /api/workers, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		workers := map[string]*WorkerInfo{
			"worker-1": {
				ID:               "worker-1",
				Host:             "localhost:8081",
				Status:           "active",
				CPU:              0.25,
				Memory:           1024 * 1024 * 100,
				DiskUsage:        1024 * 1024 * 500,
				BuildCount:       10,
				LastBuildTime:    time.Now().Add(-5 * time.Minute),
				AverageBuildTime: 2 * time.Minute,
				LastCheckin:      time.Now(),
				ResponseTime:     100 * time.Millisecond,
			},
			"worker-2": {
				ID:               "worker-2",
				Host:             "localhost:8082",
				Status:           "idle",
				CPU:              0.1,
				Memory:           1024 * 1024 * 50,
				DiskUsage:        1024 * 1024 * 200,
				BuildCount:       5,
				LastBuildTime:    time.Now().Add(-10 * time.Minute),
				AverageBuildTime: 3 * time.Minute,
				LastCheckin:      time.Now().Add(-30 * time.Second),
				ResponseTime:     150 * time.Millisecond,
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
	if worker1, ok := workers["worker-1"]; !ok {
		t.Error("Expected worker-1 in response")
	} else if worker1.Status != "active" {
		t.Errorf("Expected worker-1 status active, got %s", worker1.Status)
	}
	if worker2, ok := workers["worker-2"]; !ok {
		t.Error("Expected worker-2 in response")
	} else if worker2.Status != "idle" {
		t.Errorf("Expected worker-2 status idle, got %s", worker2.Status)
	}
}

func TestGetSystemStatus_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/api/status" {
			t.Errorf("Expected path /api/status, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SystemStatus{
			Timestamp:    time.Now(),
			WorkerCount:  3,
			QueueLength:  2,
			ActiveBuilds: 1,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	status, err := client.GetSystemStatus()
	if err != nil {
		t.Fatalf("GetSystemStatus failed: %v", err)
	}
	if status.WorkerCount != 3 {
		t.Errorf("Expected WorkerCount 3, got %d", status.WorkerCount)
	}
	if status.QueueLength != 2 {
		t.Errorf("Expected QueueLength 2, got %d", status.QueueLength)
	}
	if status.ActiveBuilds != 1 {
		t.Errorf("Expected ActiveBuilds 1, got %d", status.ActiveBuilds)
	}
}

func TestWaitForBuild_Success(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		var status BuildStatus
		if attempts == 1 {
			status = BuildStatus{
				BuildID: "test-build-123",
				Status:  "running",
			}
		} else {
			status = BuildStatus{
				BuildID: "test-build-123",
				Status:  "completed",
				Success: true,
			}
		}
		json.NewEncoder(w).Encode(status)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	status, err := client.WaitForBuild("test-build-123", 10*time.Second)
	if err != nil {
		t.Fatalf("WaitForBuild failed: %v", err)
	}
	if status.Status != "completed" {
		t.Errorf("Expected Status completed, got %s", status.Status)
	}
	if !status.Success {
		t.Error("Expected Success true")
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestWaitForBuild_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(BuildStatus{
			BuildID: "test-build-123",
			Status:  "running",
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.WaitForBuild("test-build-123", 1*time.Second)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
	expectedErr := "timeout waiting for build test-build-123 to complete"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%v'", expectedErr, err)
	}
}

func TestHealthCheck_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		if r.URL.Path != "/health" {
			t.Errorf("Expected path /health, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.HealthCheck()
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}
}

func TestHealthCheck_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.HealthCheck()
	if err == nil {
		t.Error("Expected health check error, got nil")
	}
	expectedErr := "health check failed with status code: 500"
	if err.Error() != expectedErr {
		t.Errorf("Expected error '%s', got '%v'", expectedErr, err)
	}
}

func TestBuildRequest_JSONSerialization(t *testing.T) {
	req := BuildRequest{
		ProjectPath:  "/test/project",
		TaskName:     "build",
		WorkerID:     "worker-1",
		CacheEnabled: true,
		BuildOptions: map[string]string{"clean": "true", "parallel": "4"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal BuildRequest: %v", err)
	}

	var decodedReq BuildRequest
	if err := json.Unmarshal(data, &decodedReq); err != nil {
		t.Fatalf("Failed to unmarshal BuildRequest: %v", err)
	}

	if decodedReq.ProjectPath != req.ProjectPath {
		t.Errorf("Expected ProjectPath %s, got %s", req.ProjectPath, decodedReq.ProjectPath)
	}
	if decodedReq.TaskName != req.TaskName {
		t.Errorf("Expected TaskName %s, got %s", req.TaskName, decodedReq.TaskName)
	}
	if decodedReq.WorkerID != req.WorkerID {
		t.Errorf("Expected WorkerID %s, got %s", req.WorkerID, decodedReq.WorkerID)
	}
	if decodedReq.CacheEnabled != req.CacheEnabled {
		t.Errorf("Expected CacheEnabled %v, got %v", req.CacheEnabled, decodedReq.CacheEnabled)
	}
	if len(decodedReq.BuildOptions) != 2 {
		t.Errorf("Expected 2 BuildOptions, got %d", len(decodedReq.BuildOptions))
	}
}

func TestBuildStatus_JSONSerialization(t *testing.T) {
	now := time.Now()
	status := BuildStatus{
		BuildID:      "test-build-123",
		WorkerID:     "worker-1",
		ProjectPath:  "/test/project",
		TaskName:     "build",
		Status:       "completed",
		StartTime:    now.Add(-5 * time.Minute),
		EndTime:      now,
		Duration:     5 * time.Minute,
		Success:      true,
		CacheHitRate: 0.75,
		Artifacts:    []string{"app.jar", "test-results.xml"},
		ErrorMessage: "",
	}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("Failed to marshal BuildStatus: %v", err)
	}

	var decodedStatus BuildStatus
	if err := json.Unmarshal(data, &decodedStatus); err != nil {
		t.Fatalf("Failed to unmarshal BuildStatus: %v", err)
	}

	if decodedStatus.BuildID != status.BuildID {
		t.Errorf("Expected BuildID %s, got %s", status.BuildID, decodedStatus.BuildID)
	}
	if decodedStatus.Status != status.Status {
		t.Errorf("Expected Status %s, got %s", status.Status, decodedStatus.Status)
	}
	if decodedStatus.Success != status.Success {
		t.Errorf("Expected Success %v, got %v", status.Success, decodedStatus.Success)
	}
	if decodedStatus.CacheHitRate != status.CacheHitRate {
		t.Errorf("Expected CacheHitRate %f, got %f", status.CacheHitRate, decodedStatus.CacheHitRate)
	}
	if len(decodedStatus.Artifacts) != 2 {
		t.Errorf("Expected 2 Artifacts, got %d", len(decodedStatus.Artifacts))
	}
}

func TestWorkerInfo_JSONSerialization(t *testing.T) {
	now := time.Now()
	worker := WorkerInfo{
		ID:               "worker-1",
		Host:             "localhost:8081",
		Status:           "active",
		CPU:              0.25,
		Memory:           1024 * 1024 * 100,
		DiskUsage:        1024 * 1024 * 500,
		BuildCount:       10,
		LastBuildTime:    now.Add(-5 * time.Minute),
		AverageBuildTime: 2 * time.Minute,
		LastCheckin:      now,
		ResponseTime:     100 * time.Millisecond,
	}

	data, err := json.Marshal(worker)
	if err != nil {
		t.Fatalf("Failed to marshal WorkerInfo: %v", err)
	}

	var decodedWorker WorkerInfo
	if err := json.Unmarshal(data, &decodedWorker); err != nil {
		t.Fatalf("Failed to unmarshal WorkerInfo: %v", err)
	}

	if decodedWorker.ID != worker.ID {
		t.Errorf("Expected ID %s, got %s", worker.ID, decodedWorker.ID)
	}
	if decodedWorker.Status != worker.Status {
		t.Errorf("Expected Status %s, got %s", worker.Status, decodedWorker.Status)
	}
	if decodedWorker.CPU != worker.CPU {
		t.Errorf("Expected CPU %f, got %f", worker.CPU, decodedWorker.CPU)
	}
	if decodedWorker.Memory != worker.Memory {
		t.Errorf("Expected Memory %d, got %d", worker.Memory, decodedWorker.Memory)
	}
	if decodedWorker.BuildCount != worker.BuildCount {
		t.Errorf("Expected BuildCount %d, got %d", worker.BuildCount, decodedWorker.BuildCount)
	}
}

func TestSystemStatus_JSONSerialization(t *testing.T) {
	now := time.Now()
	status := SystemStatus{
		Timestamp:    now,
		WorkerCount:  3,
		QueueLength:  2,
		ActiveBuilds: 1,
	}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("Failed to marshal SystemStatus: %v", err)
	}

	var decodedStatus SystemStatus
	if err := json.Unmarshal(data, &decodedStatus); err != nil {
		t.Fatalf("Failed to unmarshal SystemStatus: %v", err)
	}

	if decodedStatus.WorkerCount != status.WorkerCount {
		t.Errorf("Expected WorkerCount %d, got %d", status.WorkerCount, decodedStatus.WorkerCount)
	}
	if decodedStatus.QueueLength != status.QueueLength {
		t.Errorf("Expected QueueLength %d, got %d", status.QueueLength, decodedStatus.QueueLength)
	}
	if decodedStatus.ActiveBuilds != status.ActiveBuilds {
		t.Errorf("Expected ActiveBuilds %d, got %d", status.ActiveBuilds, decodedStatus.ActiveBuilds)
	}
}
