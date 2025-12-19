package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"distributed-gradle-building/types"
)

// Use types from types package instead of redefining them

// Local extension of WorkerPool for methods
type WorkerPool struct {
	*types.WorkerPool
}

// Local extension of Worker for methods
type Worker struct {
	*types.Worker
}

// BuildCoordinator coordinates distributed builds
type BuildCoordinator struct {
	WorkerPool   *WorkerPool
	BuildQueue   chan types.BuildRequest
	ActiveBuilds map[string]chan types.BuildResponse
	mutex        sync.RWMutex
	httpServer   *http.Server
	rpcServer    *rpc.Server
	listener     net.Listener
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(maxWorkers int) *WorkerPool {
	return &WorkerPool{
		WorkerPool: &types.WorkerPool{
			Workers:    make(map[string]*types.Worker),
			MaxWorkers: maxWorkers,
		},
	}
}

// NewBuildCoordinator creates a new build coordinator
func NewBuildCoordinator(maxWorkers int) *BuildCoordinator {
	coordinator := &BuildCoordinator{
		WorkerPool:   NewWorkerPool(maxWorkers),
		BuildQueue:   make(chan types.BuildRequest, 100),
		ActiveBuilds: make(map[string]chan types.BuildResponse),
	}

	// Initialize RPC server
	coordinator.rpcServer = rpc.NewServer()
	coordinator.rpcServer.Register(coordinator)

	return coordinator
}

// AddWorker adds a new worker to the pool
func (wp *WorkerPool) AddWorker(worker *types.Worker) error {
	wp.WorkerPool.Mutex.Lock()
	defer wp.WorkerPool.Mutex.Unlock()

	if len(wp.WorkerPool.Workers) >= wp.WorkerPool.MaxWorkers {
		return fmt.Errorf("worker pool at maximum capacity")
	}

	worker.Status = "idle"
	worker.LastCheckin = time.Now()
	wp.WorkerPool.Workers[worker.ID] = worker

	log.Printf("Added worker %s (%s:%d) to pool", worker.ID, worker.Host, worker.Port)
	return nil
}

// RemoveWorker removes a worker from the pool
func (wp *WorkerPool) RemoveWorker(workerID string) error {
	wp.WorkerPool.Mutex.Lock()
	defer wp.WorkerPool.Mutex.Unlock()

	if _, exists := wp.WorkerPool.Workers[workerID]; !exists {
		return fmt.Errorf("worker %s not found", workerID)
	}

	delete(wp.WorkerPool.Workers, workerID)
	log.Printf("Removed worker %s from pool", workerID)
	return nil
}

// GetAvailableWorker finds an available worker for the given task
func (wp *WorkerPool) GetAvailableWorker(taskType string) (*types.Worker, error) {
	wp.WorkerPool.Mutex.RLock()
	defer wp.WorkerPool.Mutex.RUnlock()

	for _, worker := range wp.WorkerPool.Workers {
		isAvailable := worker.Status == "idle"
		hasCapability := false

		for _, capability := range worker.Capabilities {
			if capability == taskType || capability == "all" {
				hasCapability = true
				break
			}
		}

		if isAvailable && hasCapability {
			return worker, nil
		}
	}

	return nil, fmt.Errorf("no available workers for task type: %s", taskType)
}

// UpdateWorkerStatus updates a worker's status
func (wp *WorkerPool) UpdateWorkerStatus(workerID string, status string) error {
	wp.WorkerPool.Mutex.Lock()
	defer wp.WorkerPool.Mutex.Unlock()

	worker, exists := wp.WorkerPool.Workers[workerID]
	if !exists {
		return fmt.Errorf("worker %s not found", workerID)
	}

	worker.Status = status
	worker.LastCheckin = time.Now()

	return nil
}

// SubmitBuild submits a build request to the coordinator
func (bc *BuildCoordinator) SubmitBuild(request types.BuildRequest) (string, error) {
	responseChan := make(chan types.BuildResponse)

	bc.mutex.Lock()
	bc.ActiveBuilds[request.RequestID] = responseChan
	bc.mutex.Unlock()

	bc.BuildQueue <- request

	return request.RequestID, nil
}

// calculateCacheMetrics calculates cache hit rate for a project
func (bc *BuildCoordinator) calculateCacheMetrics(projectPath string) (hits, total int) {
	// In a real implementation, this would analyze cache usage
	// For now, return reasonable defaults based on project structure
	total = 100

	// Count cache files to estimate hit rate
	cacheDir := filepath.Join(projectPath, ".gradle", "caches")
	if _, err := os.Stat(cacheDir); err == nil {
		// Count cache entries
		filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				hits++
			}
			return nil
		})
	}

	// Ensure we don't exceed total
	if hits > total {
		hits = total
	}

	return hits, total
}

// countCompiledFiles counts compiled files in a project
func (bc *BuildCoordinator) countCompiledFiles(projectPath string) int {
	count := 0

	// Look for common output directories
	outputDirs := []string{
		filepath.Join(projectPath, "build", "classes"),
		filepath.Join(projectPath, "build", "libs"),
		filepath.Join(projectPath, "target", "classes"),
	}

	for _, dir := range outputDirs {
		if _, err := os.Stat(dir); err == nil {
			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() {
					if strings.HasSuffix(path, ".class") ||
						strings.HasSuffix(path, ".jar") {
						count++
					}
				}
				return nil
			})
		}
	}

	return count
}

// ProcessBuild processes a build request
func (bc *BuildCoordinator) ProcessBuild(request types.BuildRequest) types.BuildResponse {
	// Try to get an available worker with retry mechanism
	var worker *types.Worker
	var err error

	for retry := 0; retry < 3; retry++ {
		worker, err = bc.WorkerPool.GetAvailableWorker("gradle")
		if err == nil {
			break // Found available worker
		}

		if retry < 2 {
			// Wait a bit and retry
			time.Sleep(time.Duration(retry+1) * time.Second)
		}
	}

	if err != nil {
		return types.BuildResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("No available workers after retries: %v", err),
			RequestID:    request.RequestID,
			Timestamp:    time.Now(),
		}
	}

	// Mark worker as busy
	bc.WorkerPool.UpdateWorkerStatus(worker.ID, "busy")
	defer bc.WorkerPool.UpdateWorkerStatus(worker.ID, "idle")

	startTime := time.Now()

	// Execute build command
	cmd := exec.Command("gradle", request.TaskName)
	cmd.Dir = request.ProjectPath

	if request.CacheEnabled {
		cmd.Env = append(os.Environ(), "GRADLE_USER_HOME=/tmp/gradle-cache")
	}

	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	if err != nil {
		return types.BuildResponse{
			Success:       false,
			WorkerID:      worker.ID,
			BuildDuration: duration,
			ErrorMessage:  fmt.Sprintf("Build failed: %v\nOutput: %s", err, string(output)),
			RequestID:     request.RequestID,
			Timestamp:     time.Now(),
		}
	}

	worker.BuildCount++

	// Calculate actual build metrics
	cacheHits, totalRequests := bc.calculateCacheMetrics(request.ProjectPath)
	compiledFiles := bc.countCompiledFiles(request.ProjectPath)

	return types.BuildResponse{
		Success:       true,
		WorkerID:      worker.ID,
		BuildDuration: duration,
		Artifacts:     bc.findArtifacts(request.ProjectPath),
		RequestID:     request.RequestID,
		Timestamp:     time.Now(),
		Metrics: types.BuildMetrics{
			CacheHitRate:  float64(cacheHits) / float64(totalRequests),
			CompiledFiles: compiledFiles,
		},
	}
}

// findArtifacts finds build artifacts in the project directory
func (bc *BuildCoordinator) findArtifacts(projectPath string) []string {
	var artifacts []string

	// Look for common Gradle output directories
	outputDirs := []string{
		filepath.Join(projectPath, "build/libs"),
		filepath.Join(projectPath, "build/distributions"),
	}

	for _, dir := range outputDirs {
		if _, err := os.Stat(dir); err == nil {
			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() {
					artifacts = append(artifacts, path)
				}
				return nil
			})
		}
	}

	return artifacts
}

// StartHTTPServer starts the HTTP API server
func (bc *BuildCoordinator) StartHTTPServer(port int) error {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/build", bc.handleBuildRequest)
	mux.HandleFunc("/api/workers", bc.handleWorkersRequest)
	mux.HandleFunc("/api/status", bc.handleStatusRequest)
	mux.HandleFunc("/health", bc.handleHealthCheck)

	bc.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Printf("Starting HTTP server on port %d", port)
	return bc.httpServer.ListenAndServe()
}

// StartRPCServer starts the RPC server
func (bc *BuildCoordinator) StartRPCServer(port int) error {
	var err error
	bc.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	log.Printf("Starting RPC server on port %d", port)
	go bc.processBuildQueue()
	go bc.rpcServer.Accept(bc.listener)

	return nil
}

// processBuildQueue processes queued build requests
func (bc *BuildCoordinator) processBuildQueue() {
	for request := range bc.BuildQueue {
		go func(req types.BuildRequest) {
			response := bc.ProcessBuild(req)

			bc.mutex.RLock()
			if responseChan, exists := bc.ActiveBuilds[req.RequestID]; exists {
				responseChan <- response
			}
			bc.mutex.RUnlock()
		}(request)
	}
}

// HTTP Handlers
func (bc *BuildCoordinator) handleBuildRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request types.BuildRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Generate request ID if not provided
	if request.RequestID == "" {
		request.RequestID = fmt.Sprintf("build-%d", time.Now().UnixNano())
	}

	buildID, err := bc.SubmitBuild(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"build_id": buildID})
}

func (bc *BuildCoordinator) handleWorkersRequest(w http.ResponseWriter, r *http.Request) {
	bc.WorkerPool.Mutex.RLock()
	workers := bc.WorkerPool.Workers
	bc.WorkerPool.Mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workers)
}

func (bc *BuildCoordinator) handleStatusRequest(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"timestamp":     time.Now(),
		"worker_count":  len(bc.WorkerPool.Workers),
		"queue_length":  len(bc.BuildQueue),
		"active_builds": len(bc.ActiveBuilds),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (bc *BuildCoordinator) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// Main coordinator application entry point
func coordinatorMain() {
	coordinator := NewBuildCoordinator(10)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start servers in goroutines
	go func() {
		if err := coordinator.StartHTTPServer(8080); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	go func() {
		if err := coordinator.StartRPCServer(8081); err != nil {
			log.Fatalf("RPC server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	log.Println("Shutting down coordinator...")

	// Graceful shutdown
	if coordinator.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		coordinator.httpServer.Shutdown(ctx)
	}

	if coordinator.listener != nil {
		coordinator.listener.Close()
	}
}

// Helper function to start coordinator
func runCoordinator() {
	coordinatorMain()
}

// Main function removed - use coordinator/main.go instead
