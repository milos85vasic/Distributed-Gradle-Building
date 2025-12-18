package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

// TestWorkerStartup tests worker node startup and initialization
func TestWorkerStartup(t *testing.T) {
	// Create temporary directory for worker
	tempDir, err := os.MkdirTemp("", "worker-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create worker configuration
	config := WorkerConfig{
		ID:              "test-worker-1",
		Host:            "localhost",
		Port:            8083,
		CoordinatorURL:  "http://localhost:8080",
		WorkDir:         tempDir,
		Capacity:        4,
		HeartbeatInterval: 30 * time.Second,
		MaxRetries:      3,
	}
	
	// Create worker instance
	worker, err := NewWorker(config)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	
	// Verify worker initialization
	if worker.ID != config.ID {
		t.Errorf("Expected worker ID '%s', got '%s'", config.ID, worker.ID)
	}
	
	if worker.Status != "initializing" {
		t.Errorf("Expected status 'initializing', got '%s'", worker.Status)
	}
	
	if worker.Capacity != config.Capacity {
		t.Errorf("Expected capacity %d, got %d", config.Capacity, worker.Capacity)
	}
}

// TestWorkerRegistration tests worker registration with coordinator
func TestWorkerRegistration(t *testing.T) {
	// Create temporary directory for worker
	tempDir, err := os.MkdirTemp("", "worker-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Mock coordinator endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/api/workers" {
			// Mock successful registration response
			response := WorkerRegistrationResponse{
				WorkerID: "test-worker-1",
				Success:  true,
				Message:  "Worker registered successfully",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()
	
	// Create worker configuration with mock coordinator URL
	config := WorkerConfig{
		ID:              "test-worker-1",
		Host:            "localhost",
		Port:            8083,
		CoordinatorURL:  server.URL,
		WorkDir:         tempDir,
		Capacity:        4,
		HeartbeatInterval: 30 * time.Second,
		MaxRetries:      3,
	}
	
	// Create worker instance
	worker, err := NewWorker(config)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	
	// Test registration
	err = worker.Register()
	if err != nil {
		t.Errorf("Worker registration should succeed, got error: %v", err)
	}
	
	// Verify worker status after registration
	if worker.Status != "registered" {
		t.Errorf("Expected status 'registered', got '%s'", worker.Status)
	}
}

// TestTaskExecution tests build task execution
func TestTaskExecution(t *testing.T) {
	// Create temporary directory for worker
	tempDir, err := os.MkdirTemp("", "worker-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create test project structure
	projectDir := filepath.Join(tempDir, "project")
	err = os.MkdirAll(filepath.Join(projectDir, "src/main/java/com/example"), 0755)
	if err != nil {
		t.Fatalf("Failed to create project structure: %v", err)
	}
	
	// Create build.gradle
	buildGradle := filepath.Join(projectDir, "build.gradle")
	buildContent := `plugins {
    id 'java'
}
repositories {
    mavenCentral()
}
dependencies {
    implementation 'org.apache.commons:commons-lang3:3.12.0'
}
`
	err = os.WriteFile(buildGradle, []byte(buildContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write build.gradle: %v", err)
	}
	
	// Create test Java source
	javaFile := filepath.Join(projectDir, "src/main/java/com/example/TestApp.java")
	javaContent := `package com.example;

public class TestApp {
    public static void main(String[] args) {
        System.out.println("Test Application");
    }
}
`
	err = os.WriteFile(javaFile, []byte(javaContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write Java source: %v", err)
	}
	
	// Create worker instance
	config := WorkerConfig{
		ID:              "test-worker-1",
		Host:            "localhost",
		Port:            8083,
		CoordinatorURL:  "http://localhost:8080",
		WorkDir:         tempDir,
		Capacity:        4,
		HeartbeatInterval: 30 * time.Second,
		MaxRetries:      3,
	}
	
	worker, err := NewWorker(config)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	
	// Create test task
	task := &BuildTask{
		ID:          "task-123",
		BuildID:     "build-456",
		Type:        "compileJava",
		ProjectPath: projectDir,
		Options: map[string]string{
			"javaHome": "/usr/lib/jvm/java-11",
			"debug":    "true",
		},
		Timeout: 5 * time.Minute,
	}
	
	// Execute task
	result, err := worker.ExecuteTask(task)
	if err != nil {
		t.Errorf("Task execution should succeed, got error: %v", err)
	}
	
	// Verify result
	if result.TaskID != task.ID {
		t.Errorf("Expected task ID '%s', got '%s'", task.ID, result.TaskID)
	}
	
	if result.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", result.Status)
	}
	
	if result.StartTime.IsZero() {
		t.Error("StartTime should not be zero")
	}
	
	if result.EndTime.IsZero() {
		t.Error("EndTime should not be zero")
	}
	
	if !result.EndTime.After(result.StartTime) {
		t.Error("EndTime should be after StartTime")
	}
}

// TestTaskTimeout tests task timeout handling
func TestTaskTimeout(t *testing.T) {
	// Create temporary directory for worker
	tempDir, err := os.MkdirTemp("", "worker-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create worker instance
	config := WorkerConfig{
		ID:              "test-worker-1",
		Host:            "localhost",
		Port:            8083,
		CoordinatorURL:  "http://localhost:8080",
		WorkDir:         tempDir,
		Capacity:        4,
		HeartbeatInterval: 30 * time.Second,
		MaxRetries:      3,
	}
	
	worker, err := NewWorker(config)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	
	// Create test task with very short timeout
	task := &BuildTask{
		ID:          "task-timeout",
		BuildID:     "build-timeout",
		Type:        "compileJava",
		ProjectPath: "/nonexistent/project",
		Options:     map[string]string{},
		Timeout:     1 * time.Second,
	}
	
	// Execute task with timeout
	startTime := time.Now()
	result, err := worker.ExecuteTask(task)
	executionTime := time.Since(startTime)
	
	// Verify timeout occurred
	if err == nil {
		t.Error("Task execution should fail due to timeout")
	}
	
	if result.Status != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", result.Status)
	}
	
	if executionTime > 30*time.Second {
		t.Errorf("Task should timeout quickly, but took %v", executionTime)
	}
	
	if !contains(result.Error, "timeout") && !contains(result.Error, "deadline") {
		t.Errorf("Error should mention timeout, got: %s", result.Error)
	}
}

// TestHeartbeatMechanism tests worker heartbeat to coordinator
func TestHeartbeatMechanism(t *testing.T) {
	// Create temporary directory for worker
	tempDir, err := os.MkdirTemp("", "worker-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Track heartbeat requests
	var heartbeatRequests []WorkerHeartbeat
	
	// Mock coordinator endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/api/workers/heartbeat" {
			var heartbeat WorkerHeartbeat
			err := json.NewDecoder(r.Body).Decode(&heartbeat)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			
			heartbeatRequests = append(heartbeatRequests, heartbeat)
			
			response := HeartbeatResponse{
				Success: true,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()
	
	// Create worker configuration with mock coordinator URL
	config := WorkerConfig{
		ID:              "test-worker-1",
		Host:            "localhost",
		Port:            8083,
		CoordinatorURL:  server.URL,
		WorkDir:         tempDir,
		Capacity:        4,
		HeartbeatInterval: 100 * time.Millisecond, // Fast heartbeat for testing
		MaxRetries:      3,
	}
	
	// Create worker instance
	worker, err := NewWorker(config)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	
	// Start heartbeat goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	
	go worker.StartHeartbeat(ctx)
	
	// Wait for heartbeats
	time.Sleep(400 * time.Millisecond)
	
	// Verify heartbeats were sent
	if len(heartbeatRequests) < 2 {
		t.Errorf("Expected at least 2 heartbeats, got %d", len(heartbeatRequests))
	}
	
	// Verify heartbeat content
	for _, hb := range heartbeatRequests {
		if hb.WorkerID != config.ID {
			t.Errorf("Expected worker ID '%s', got '%s'", config.ID, hb.WorkerID)
		}
		
		if hb.Status != "running" {
			t.Errorf("Expected status 'running', got '%s'", hb.Status)
		}
		
		if hb.Timestamp.IsZero() {
			t.Error("Timestamp should not be zero")
		}
		
		if hb.CPUUsage < 0 || hb.CPUUsage > 1 {
			t.Errorf("CPU usage should be between 0 and 1, got %f", hb.CPUUsage)
		}
		
		if hb.MemoryUsage < 0 || hb.MemoryUsage > 1 {
			t.Errorf("Memory usage should be between 0 and 1, got %f", hb.MemoryUsage)
		}
	}
}

// TestResourceMonitoring tests worker resource monitoring
func TestResourceMonitoring(t *testing.T) {
	// Create temporary directory for worker
	tempDir, err := os.MkdirTemp("", "worker-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create worker instance
	config := WorkerConfig{
		ID:              "test-worker-1",
		Host:            "localhost",
		Port:            8083,
		CoordinatorURL:  "http://localhost:8080",
		WorkDir:         tempDir,
		Capacity:        4,
		HeartbeatInterval: 30 * time.Second,
		MaxRetries:      3,
	}
	
	worker, err := NewWorker(config)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	
	// Test resource collection
	resources := worker.CollectResources()
	
	// Verify resource data
	if resources.Timestamp.IsZero() {
		t.Error("Resource timestamp should not be zero")
	}
	
	if resources.CPUUsage < 0 || resources.CPUUsage > 1 {
		t.Errorf("CPU usage should be between 0 and 1, got %f", resources.CPUUsage)
	}
	
	if resources.MemoryUsage < 0 || resources.MemoryUsage > 1 {
		t.Errorf("Memory usage should be between 0 and 1, got %f", resources.MemoryUsage)
	}
	
	if resources.DiskUsage < 0 || resources.DiskUsage > 1 {
		t.Errorf("Disk usage should be between 0 and 1, got %f", resources.DiskUsage)
	}
	
	if resources.NumCPU <= 0 {
		t.Errorf("NumCPU should be positive, got %d", resources.NumCPU)
	}
	
	if resources.TotalMemory <= 0 {
		t.Errorf("TotalMemory should be positive, got %d", resources.TotalMemory)
	}
	
	if resources.AvailableMemory <= 0 {
		t.Errorf("AvailableMemory should be positive, got %d", resources.AvailableMemory)
	}
	
	if resources.AvailableMemory > resources.TotalMemory {
		t.Error("AvailableMemory should not exceed TotalMemory")
	}
}

// TestTaskQueue tests task queue management
func TestTaskQueue(t *testing.T) {
	// Create temporary directory for worker
	tempDir, err := os.MkdirTemp("", "worker-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create worker instance
	config := WorkerConfig{
		ID:              "test-worker-1",
		Host:            "localhost",
		Port:            8083,
		CoordinatorURL:  "http://localhost:8080",
		WorkDir:         tempDir,
		Capacity:        4,
		HeartbeatInterval: 30 * time.Second,
		MaxRetries:      3,
	}
	
	worker, err := NewWorker(config)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	
	// Test adding tasks to queue
	tasks := []*BuildTask{
		{
			ID:          "task-1",
			BuildID:     "build-1",
			Type:        "compileJava",
			ProjectPath: "/tmp/project1",
			Options:     map[string]string{},
			Timeout:     5 * time.Minute,
		},
		{
			ID:          "task-2",
			BuildID:     "build-2",
			Type:        "compileTestJava",
			ProjectPath: "/tmp/project2",
			Options:     map[string]string{},
			Timeout:     5 * time.Minute,
		},
		{
			ID:          "task-3",
			BuildID:     "build-3",
			Type:        "test",
			ProjectPath: "/tmp/project3",
			Options:     map[string]string{},
			Timeout:     5 * time.Minute,
		},
	}
	
	// Add tasks to queue
	for _, task := range tasks {
		worker.AddTask(task)
	}
	
	// Verify queue size
	if len(worker.TaskQueue) != len(tasks) {
		t.Errorf("Expected queue size %d, got %d", len(tasks), len(worker.TaskQueue))
	}
	
	// Verify task order (FIFO)
	for i, task := range worker.TaskQueue {
		if task.ID != tasks[i].ID {
			t.Errorf("Expected task ID '%s' at position %d, got '%s'", tasks[i].ID, i, task.ID)
		}
	}
	
	// Test task retrieval
	retrievedTask := worker.GetNextTask()
	if retrievedTask == nil {
		t.Error("Should retrieve next task from queue")
		return
	}
	
	if retrievedTask.ID != "task-1" {
		t.Errorf("Expected task ID 'task-1', got '%s'", retrievedTask.ID)
	}
	
	// Verify queue size after retrieval
	if len(worker.TaskQueue) != 2 {
		t.Errorf("Expected queue size 2 after retrieval, got %d", len(worker.TaskQueue))
	}
}

// TestErrorHandling tests error handling in worker operations
func TestErrorHandling(t *testing.T) {
	// Create temporary directory for worker
	tempDir, err := os.MkdirTemp("", "worker-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create worker instance
	config := WorkerConfig{
		ID:              "test-worker-1",
		Host:            "localhost",
		Port:            8083,
		CoordinatorURL:  "http://localhost:8080",
		WorkDir:         tempDir,
		Capacity:        4,
		HeartbeatInterval: 30 * time.Second,
		MaxRetries:      3,
	}
	
	worker, err := NewWorker(config)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	
	testCases := []struct {
		name        string
		task        *BuildTask
		expectError bool
	}{
		{
			name: "Non-existent project",
			task: &BuildTask{
				ID:          "task-1",
				BuildID:     "build-1",
				Type:        "compileJava",
				ProjectPath: "/nonexistent/path",
				Options:     map[string]string{},
				Timeout:     5 * time.Minute,
			},
			expectError: true,
		},
		{
			name: "Invalid task type",
			task: &BuildTask{
				ID:          "task-2",
				BuildID:     "build-2",
				Type:        "invalidTaskType",
				ProjectPath: tempDir,
				Options:     map[string]string{},
				Timeout:     5 * time.Minute,
			},
			expectError: true,
		},
		{
			name: "Missing required fields",
			task: &BuildTask{
				ID:          "", // Missing ID
				BuildID:     "build-3",
				Type:        "compileJava",
				ProjectPath: tempDir,
				Options:     map[string]string{},
				Timeout:     5 * time.Minute,
			},
			expectError: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := worker.ExecuteTask(tc.task)
			
			if tc.expectError {
				if err == nil {
					t.Error("Expected error but task succeeded")
				}
				if result.Status != "failed" {
					t.Errorf("Expected status 'failed', got '%s'", result.Status)
				}
			} else {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
				}
			}
		})
	}
}

// TestWorkerLifecycle tests worker startup, operation, and shutdown
func TestWorkerLifecycle(t *testing.T) {
	// Create temporary directory for worker
	tempDir, err := os.MkdirTemp("", "worker-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Mock coordinator endpoints
	var registeredWorker bool
	var heartbeatCount int
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/api/workers":
			var reg Worker
			err := json.NewDecoder(r.Body).Decode(&reg)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			registeredWorker = true
			
			response := WorkerRegistrationResponse{
				WorkerID: reg.ID,
				Success:  true,
				Message:  "Worker registered successfully",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(response)
			
		case r.Method == "POST" && r.URL.Path == "/api/workers/heartbeat":
			heartbeatCount++
			response := HeartbeatResponse{
				Success: true,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()
	
	// Create worker configuration
	config := WorkerConfig{
		ID:              "test-worker-1",
		Host:            "localhost",
		Port:            8083,
		CoordinatorURL:  server.URL,
		WorkDir:         tempDir,
		Capacity:        4,
		HeartbeatInterval: 100 * time.Millisecond,
		MaxRetries:      3,
	}
	
	// Create worker instance
	worker, err := NewWorker(config)
	if err != nil {
		t.Fatalf("Failed to create worker: %v", err)
	}
	
	// Test startup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	// Start worker
	worker.Start(ctx)
	
	// Wait for registration and heartbeats
	time.Sleep(500 * time.Millisecond)
	
	// Verify registration
	if !registeredWorker {
		t.Error("Worker should register with coordinator")
	}
	
	// Verify heartbeats
	if heartbeatCount < 2 {
		t.Errorf("Expected at least 2 heartbeats, got %d", heartbeatCount)
	}
	
	// Verify worker status
	if worker.Status != "running" {
		t.Errorf("Expected status 'running', got '%s'", worker.Status)
	}
	
	// Test graceful shutdown
	cancel()
	time.Sleep(200 * time.Millisecond)
	
	if worker.Status != "stopped" {
		t.Errorf("Expected status 'stopped', got '%s'", worker.Status)
	}
}

// TestWorkerConfiguration tests worker configuration validation
func TestWorkerConfiguration(t *testing.T) {
	testCases := []struct {
		name        string
		config      WorkerConfig
		expectError bool
	}{
		{
			name: "Valid configuration",
			config: WorkerConfig{
				ID:              "worker-1",
				Host:            "localhost",
				Port:            8083,
				CoordinatorURL:  "http://localhost:8080",
				WorkDir:         "/tmp/worker",
				Capacity:        4,
				HeartbeatInterval: 30 * time.Second,
				MaxRetries:      3,
			},
			expectError: false,
		},
		{
			name: "Missing ID",
			config: WorkerConfig{
				Host:            "localhost",
				Port:            8083,
				CoordinatorURL:  "http://localhost:8080",
				WorkDir:         "/tmp/worker",
				Capacity:        4,
				HeartbeatInterval: 30 * time.Second,
				MaxRetries:      3,
			},
			expectError: true,
		},
		{
			name: "Invalid port",
			config: WorkerConfig{
				ID:              "worker-1",
				Host:            "localhost",
				Port:            0,
				CoordinatorURL:  "http://localhost:8080",
				WorkDir:         "/tmp/worker",
				Capacity:        4,
				HeartbeatInterval: 30 * time.Second,
				MaxRetries:      3,
			},
			expectError: true,
		},
		{
			name: "Invalid coordinator URL",
			config: WorkerConfig{
				ID:              "worker-1",
				Host:            "localhost",
				Port:            8083,
				CoordinatorURL:  "not-a-url",
				WorkDir:         "/tmp/worker",
				Capacity:        4,
				HeartbeatInterval: 30 * time.Second,
				MaxRetries:      3,
			},
			expectError: true,
		},
		{
			name: "Zero capacity",
			config: WorkerConfig{
				ID:              "worker-1",
				Host:            "localhost",
				Port:            8083,
				CoordinatorURL:  "http://localhost:8080",
				WorkDir:         "/tmp/worker",
				Capacity:        0,
				HeartbeatInterval: 30 * time.Second,
				MaxRetries:      3,
			},
			expectError: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			worker, err := NewWorker(tc.config)
			
			if tc.expectError {
				if err == nil {
					t.Error("Expected error but worker creation succeeded")
				}
				if worker != nil {
					t.Error("Expected nil worker on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
				}
				if worker == nil {
					t.Error("Expected non-nil worker on success")
				}
			}
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 (len(s) > len(substr) && 
		  (s[:len(substr)] == substr || 
		   s[len(s)-len(substr):] == substr || 
		   indexOfSubstring(s, substr) >= 0)))
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}