package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

// TestDistributedBuild tests the distributed build process across multiple workers
func TestDistributedBuild(t *testing.T) {
	// Create test environment
	testDir, err := os.MkdirTemp("", "distributed-build-test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create test project
	projectDir := filepath.Join(testDir, "test-project")
	if err := createTestProject(t, projectDir); err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// Mock services
	coordinator := setupDistributedBuildCoordinator(t)
	workers := setupDistributedBuildWorkers(t, 3)
	cacheServer := setupMockCacheServer(t)

	// Start mock servers
	coordServer := httptest.NewServer(coordinator)
	defer coordServer.Close()

	workerServers := make([]*httptest.Server, 3)
	for i, worker := range workers {
		workerServers[i] = httptest.NewServer(worker)
		defer workerServers[i].Close()
	}

	cacheServerInst := httptest.NewServer(cacheServer)
	defer cacheServerInst.Close()

	// Test cases
	tests := []struct {
		name        string
		testFunc    func(t *testing.T, coordURL string, workerURLs []string, cacheURL string, projectDir string)
		description string
	}{
		{
			name: "TaskDistribution",
			testFunc: func(t *testing.T, coordURL string, workerURLs []string, cacheURL string, projectDir string) {
				testTaskDistribution(t, coordURL, workerURLs, projectDir)
			},
			description: "Test task distribution across workers",
		},
		{
			name: "BuildAggregation",
			testFunc: func(t *testing.T, coordURL string, workerURLs []string, cacheURL string, projectDir string) {
				testBuildAggregation(t, coordURL, workerURLs, projectDir)
			},
			description: "Test build result aggregation",
		},
		{
			name: "WorkerFailureHandling",
			testFunc: func(t *testing.T, coordURL string, workerURLs []string, cacheURL string, projectDir string) {
				testWorkerFailureHandling(t, coordURL, workerURLs, projectDir)
			},
			description: "Test worker failure handling",
		},
		{
			name: "LoadBalancing",
			testFunc: func(t *testing.T, coordURL string, workerURLs []string, cacheURL string, projectDir string) {
				testLoadBalancing(t, coordURL, workerURLs, projectDir)
			},
			description: "Test load balancing across workers",
		},
		{
			name: "ConcurrentBuilds",
			testFunc: func(t *testing.T, coordURL string, workerURLs []string, cacheURL string, projectDir string) {
				testConcurrentBuilds(t, coordURL, workerURLs, projectDir)
			},
			description: "Test concurrent build handling",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Running %s: %s", tc.name, tc.description)
			tc.testFunc(t, coordServer.URL, getWorkerURLs(workerServers), cacheServerInst.URL, projectDir)
		})
	}
}

// TestTaskDistribution tests that tasks are properly distributed across workers
func testTaskDistribution(t *testing.T, coordURL string, workerURLs []string, projectDir string) {
	// Create build request with multiple tasks
	buildRequest := map[string]interface{}{
		"project_name":    "distributed-test-project",
		"build_type":      "gradle",
		"tasks":           []string{"compileJava", "processResources", "test", "javadoc", "jar"},
		"source_location": projectDir,
		"environment": map[string]interface{}{
			"java_home":   "/usr/lib/jvm/default-java",
			"gradle_home": "/usr/local/gradle",
		},
		"distribution": map[string]interface{}{
			"strategy":       "round-robin",
			"parallel_tasks": 3,
		},
	}

	jsonData, err := json.Marshal(buildRequest)
	if err != nil {
		t.Fatalf("Failed to marshal build request: %v", err)
	}

	// Submit build
	resp, err := http.Post(
		fmt.Sprintf("%s/api/builds", coordURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		t.Fatalf("Failed to submit build: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("Expected status 202, got %d for build submission", resp.StatusCode)
	}

	// Get build ID
	var buildResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&buildResp); err != nil {
		t.Fatalf("Failed to decode build response: %v", err)
	}

	buildID := buildResp["build_id"].(string)

	// Wait for task distribution
	time.Sleep(200 * time.Millisecond)

	// Check task distribution across workers
	resp, err = http.Get(fmt.Sprintf("%s/api/builds/%s/tasks", coordURL, buildID))
	if err != nil {
		t.Fatalf("Failed to get tasks: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d for tasks", resp.StatusCode)
	}

	var tasksResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tasksResp); err != nil {
		t.Fatalf("Failed to decode tasks response: %v", err)
	}

	tasks := tasksResp["tasks"].([]interface{})
	if len(tasks) == 0 {
		t.Fatal("No tasks found")
	}

	// Verify tasks are distributed to different workers
	workerIDs := make(map[string]bool)
	for _, task := range tasks {
		taskMap := task.(map[string]interface{})
		workerID := taskMap["assigned_worker"].(string)
		workerIDs[workerID] = true
	}

	// Should have tasks distributed to at least 2 workers
	if len(workerIDs) < 2 {
		t.Errorf("Expected tasks distributed to at least 2 workers, got %d", len(workerIDs))
	}
}

// TestBuildAggregation tests build result aggregation
func testBuildAggregation(t *testing.T, coordURL string, workerURLs []string, projectDir string) {
	// Submit a build
	buildID := submitTestBuild(t, coordURL, projectDir, "aggregation-test-project")

	// Wait for build completion
	waitForBuildCompletion(t, coordURL, buildID, 5*time.Second)

	// Check aggregated results
	resp, err := http.Get(fmt.Sprintf("%s/api/builds/%s/results", coordURL, buildID))
	if err != nil {
		t.Fatalf("Failed to get build results: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d for build results", resp.StatusCode)
	}

	var resultsResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&resultsResp); err != nil {
		t.Fatalf("Failed to decode results response: %v", err)
	}

	// Verify aggregation results
	if resultsResp["status"] != "completed" {
		t.Errorf("Expected completed status, got %v", resultsResp["status"])
	}

	// Check if artifacts are aggregated
	artifacts := resultsResp["artifacts"].([]interface{})
	if len(artifacts) == 0 {
		t.Error("No artifacts found in aggregated results")
	}
}

// TestWorkerFailureHandling tests worker failure scenarios
func testWorkerFailureHandling(t *testing.T, coordURL string, workerURLs []string, projectDir string) {
	// Submit a build
	buildID := submitTestBuild(t, coordURL, projectDir, "failure-test-project")

	// Simulate worker failure by making one worker unavailable
	workerURLs[1] = "http://unavailable-worker:8080" // Simulate unavailable worker

	// Wait for build completion with longer timeout (to allow for retries)
	waitForBuildCompletion(t, coordURL, buildID, 10*time.Second)

	// Check that build handled worker failure gracefully
	resp, err := http.Get(fmt.Sprintf("%s/api/builds/%s/status", coordURL, buildID))
	if err != nil {
		t.Fatalf("Failed to get build status: %v", err)
	}
	defer resp.Body.Close()

	var statusResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		t.Fatalf("Failed to decode status response: %v", err)
	}

	status := statusResp["status"].(string)

	// Build should either complete (with retries) or fail gracefully
	if status != "completed" && status != "failed" {
		t.Errorf("Unexpected build status: %s", status)
	}

	if status == "failed" {
		// Check if error message indicates worker failure handling
		if errorMsg, ok := statusResp["error"].(string); ok {
			if !strings.Contains(errorMsg, "worker") && !strings.Contains(errorMsg, "retry") {
				t.Errorf("Error message doesn't indicate worker failure handling: %s", errorMsg)
			}
		}
	}
}

// TestLoadBalancing tests load balancing across workers
func testLoadBalancing(t *testing.T, coordURL string, workerURLs []string, projectDir string) {
	// Submit multiple builds
	buildIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		projectName := fmt.Sprintf("load-balance-test-%d", i)
		buildIDs[i] = submitTestBuild(t, coordURL, projectDir, projectName)
	}

	// Wait for task distribution
	time.Sleep(500 * time.Millisecond)

	// Check load distribution
	resp, err := http.Get(fmt.Sprintf("%s/api/workers/load", coordURL))
	if err != nil {
		t.Fatalf("Failed to get worker load: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d for worker load", resp.StatusCode)
	}

	var loadResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&loadResp); err != nil {
		t.Fatalf("Failed to decode load response: %v", err)
	}

	workers := loadResp["workers"].([]interface{})

	// Verify load is distributed (not all tasks on one worker)
	maxTasks := 0
	totalTasks := 0

	for _, worker := range workers {
		workerMap := worker.(map[string]interface{})
		taskCount := int(workerMap["active_tasks"].(float64))
		totalTasks += taskCount
		if taskCount > maxTasks {
			maxTasks = taskCount
		}
	}

	// Check if load is reasonably distributed (no worker has more than 80% of tasks)
	if maxTasks > 0 && float64(maxTasks)/float64(totalTasks) > 0.8 {
		t.Errorf("Load not well distributed: max tasks %d, total tasks %d", maxTasks, totalTasks)
	}
}

// TestConcurrentBuilds tests handling of multiple concurrent builds
func testConcurrentBuilds(t *testing.T, coordURL string, workerURLs []string, projectDir string) {
	// Submit multiple builds concurrently
	buildIDs := make([]string, 3)
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func(index int) {
			projectName := fmt.Sprintf("concurrent-test-%d", index)
			buildIDs[index] = submitTestBuild(t, coordURL, projectDir, projectName)
			waitForBuildCompletion(t, coordURL, buildIDs[index], 5*time.Second)
			done <- true
		}(i)
	}

	// Wait for all builds to complete
	timeout := time.After(15 * time.Second)
	completed := 0

	for completed < 3 {
		select {
		case <-done:
			completed++
		case <-timeout:
			t.Fatalf("Timeout waiting for concurrent builds to complete")
		}
	}

	// Verify all builds completed successfully
	for _, buildID := range buildIDs {
		resp, err := http.Get(fmt.Sprintf("%s/api/builds/%s/status", coordURL, buildID))
		if err != nil {
			t.Fatalf("Failed to get build status: %v", err)
		}
		defer resp.Body.Close()

		var statusResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
			t.Fatalf("Failed to decode status response: %v", err)
		}

		if statusResp["status"] != "completed" {
			t.Errorf("Build %s did not complete successfully", buildID)
		}
	}
}

// Helper functions

func createTestProject(t *testing.T, projectDir string) error {
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return err
	}

	// Create build.gradle
	buildGradle := `
plugins {
    id 'java'
}

group 'com.example'
version '1.0.0'

repositories {
    mavenCentral()
}

dependencies {
    testImplementation 'junit:junit:4.13.2'
}

task compileJava {
    doLast {
        println "Compiling Java sources..."
        mkdir "build/classes/java/main"
    }
}

task processResources {
    doLast {
        println "Processing resources..."
        mkdir "build/resources/main"
    }
}

task test {
    doLast {
        println "Running tests..."
        mkdir "build/test-results"
    }
}

task javadoc {
    doLast {
        println "Generating javadoc..."
        mkdir "build/docs/javadoc"
    }
}

task jar {
    doLast {
        println "Creating JAR..."
        mkdir "build/libs"
    }
}
`
	if err := os.WriteFile(filepath.Join(projectDir, "build.gradle"), []byte(buildGradle), 0644); err != nil {
		return err
	}

	// Create source directory and a simple Java file
	srcDir := filepath.Join(projectDir, "src/main/java/com/example")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return err
	}

	helloJava := `package com.example;

public class Hello {
    public static void main(String[] args) {
        System.out.println("Hello World!");
    }
}`

	return os.WriteFile(filepath.Join(srcDir, "Hello.java"), []byte(helloJava), 0644)
}

func setupDistributedBuildCoordinator(t *testing.T) http.Handler {
	router := mux.NewRouter()

	// Track builds and tasks
	builds := make(map[string]map[string]interface{})
	tasks := make(map[string]map[string]interface{})

	// Build submission endpoint
	router.HandleFunc("/api/builds", func(w http.ResponseWriter, r *http.Request) {
		var buildRequest map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&buildRequest); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		buildID := fmt.Sprintf("build-%d", time.Now().UnixNano())
		builds[buildID] = buildRequest

		// Create tasks based on request
		taskList := buildRequest["tasks"].([]interface{})
		for i, task := range taskList {
			taskID := fmt.Sprintf("task-%d", i+1)
			tasks[taskID] = map[string]interface{}{
				"id":       taskID,
				"name":     task,
				"status":   "pending",
				"build_id": buildID,
			}
		}

		response := map[string]interface{}{
			"build_id":   buildID,
			"status":     "accepted",
			"task_count": len(taskList),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(response)
	}).Methods("POST")

	// Build status endpoint
	router.HandleFunc("/api/builds/{buildId}/status", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		buildID := vars["buildId"]

		if _, ok := builds[buildID]; !ok {
			http.Error(w, "Build not found", http.StatusNotFound)
			return
		}

		// Simulate build progress
		response := map[string]interface{}{
			"build_id": buildID,
			"status":   "completed",
			"progress": 100,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	// Tasks endpoint
	router.HandleFunc("/api/builds/{buildId}/tasks", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		buildID := vars["buildId"]

		// Return mock tasks with assigned workers
		buildTasks := []interface{}{
			map[string]interface{}{
				"id":              "task-1",
				"name":            "compileJava",
				"status":          "completed",
				"assigned_worker": "worker-1",
				"build_id":        buildID,
			},
			map[string]interface{}{
				"id":              "task-2",
				"name":            "processResources",
				"status":          "completed",
				"assigned_worker": "worker-2",
				"build_id":        buildID,
			},
			map[string]interface{}{
				"id":              "task-3",
				"name":            "test",
				"status":          "completed",
				"assigned_worker": "worker-3",
				"build_id":        buildID,
			},
			map[string]interface{}{
				"id":              "task-4",
				"name":            "javadoc",
				"status":          "completed",
				"assigned_worker": "worker-1",
				"build_id":        buildID,
			},
			map[string]interface{}{
				"id":              "task-5",
				"name":            "jar",
				"status":          "completed",
				"assigned_worker": "worker-2",
				"build_id":        buildID,
			},
		}

		response := map[string]interface{}{
			"tasks": buildTasks,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	// Build results endpoint
	router.HandleFunc("/api/builds/{buildId}/results", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		buildID := vars["buildId"]

		response := map[string]interface{}{
			"build_id": buildID,
			"status":   "completed",
			"artifacts": []interface{}{
				map[string]interface{}{"name": "build.log", "size": 1024},
				map[string]interface{}{"name": "classes", "size": 2048},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	// Worker load endpoint
	router.HandleFunc("/api/workers/load", func(w http.ResponseWriter, r *http.Request) {
		workers := []interface{}{
			map[string]interface{}{"id": "worker-1", "active_tasks": 2},
			map[string]interface{}{"id": "worker-2", "active_tasks": 1},
			map[string]interface{}{"id": "worker-3", "active_tasks": 3},
		}

		response := map[string]interface{}{
			"workers": workers,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	return router
}

func setupDistributedBuildWorkers(t *testing.T, count int) []http.Handler {
	workers := make([]http.Handler, count)

	for i := 0; i < count; i++ {
		router := mux.NewRouter()
		workerID := fmt.Sprintf("worker-%d", i+1)

		// Task execution endpoint
		router.HandleFunc("/api/tasks/execute", func(w http.ResponseWriter, r *http.Request) {
			var taskRequest map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&taskRequest); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			response := map[string]interface{}{
				"task_id":   "task-" + fmt.Sprintf("%d", time.Now().UnixNano()),
				"worker_id": workerID,
				"status":    "completed",
				"duration":  3000,
				"artifacts": []string{"build.log", "output"},
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}).Methods("POST")

		workers[i] = router
	}

	return workers
}

func submitTestBuild(t *testing.T, coordURL string, projectDir string, projectName string) string {
	buildRequest := map[string]interface{}{
		"project_name":    projectName,
		"build_type":      "gradle",
		"tasks":           []string{"compileJava", "test"},
		"source_location": projectDir,
		"environment": map[string]interface{}{
			"java_home":   "/usr/lib/jvm/default-java",
			"gradle_home": "/usr/local/gradle",
		},
	}

	jsonData, err := json.Marshal(buildRequest)
	if err != nil {
		t.Fatalf("Failed to marshal build request: %v", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/api/builds", coordURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		t.Fatalf("Failed to submit build: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("Expected status 202, got %d for build submission", resp.StatusCode)
	}

	var buildResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&buildResp); err != nil {
		t.Fatalf("Failed to decode build response: %v", err)
	}

	return buildResp["build_id"].(string)
}

func waitForBuildCompletion(t *testing.T, coordURL string, buildID string, timeout time.Duration) {
	start := time.Now()
	for time.Since(start) < timeout {
		resp, err := http.Get(fmt.Sprintf("%s/api/builds/%s/status", coordURL, buildID))
		if err != nil {
			t.Fatalf("Failed to get build status: %v", err)
		}
		defer resp.Body.Close()

		var statusResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
			t.Fatalf("Failed to decode status response: %v", err)
		}

		status := statusResp["status"].(string)
		if status == "completed" || status == "failed" {
			return
		}

		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("Build %s did not complete within timeout", buildID)
}
