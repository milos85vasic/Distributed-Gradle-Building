package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// BuildRequest represents a distributed build request
type BuildRequest struct {
	ProjectPath   string
	TaskName      string
	WorkerID      string
	CacheEnabled  bool
	BuildOptions  map[string]string
	Timestamp     time.Time
	RequestID     string
}

// BuildResponse represents the response from a build worker
type BuildResponse struct {
	Success       bool
	WorkerID      string
	BuildDuration time.Duration
	Artifacts     []string
	ErrorMessage  string
	Metrics       BuildMetrics
	RequestID     string
	Timestamp     time.Time
}

// BuildMetrics contains detailed build performance metrics
type BuildMetrics struct {
	BuildSteps     []BuildStep
	CacheHitRate   float64
	CompiledFiles  int
	TestResults    TestResults
	ResourceUsage  ResourceMetrics
}

// BuildStep represents individual build step metrics
type BuildStep struct {
	Name        string
	Duration    time.Duration
	Success     bool
	Output      string
	StartTime   time.Time
}

// TestResults contains test execution metrics
type TestResults struct {
	TotalTests    int
	PassedTests   int
	FailedTests   int
	SkippedTests  int
	Duration      time.Duration
	Coverage      float64
}

// ResourceMetrics tracks resource usage during build
type ResourceMetrics struct {
	CpuUsage     float64
	MemoryUsage  uint64
	DiskIO       uint64
	NetworkIO    uint64
}

// WorkerPool manages distributed build workers
type WorkerPool struct {
	Workers    map[string]*Worker
	mutex      sync.RWMutex
	maxWorkers int
}

// Worker represents a distributed build worker
type Worker struct {
	ID           string
	Host         string
	Port         int
	Status       string
	LastCheckin  time.Time
	BuildCount   int
	Capabilities []string
	Resources    ResourceMetrics
	mutex        sync.RWMutex
}

// BuildCoordinator coordinates distributed builds
type BuildCoordinator struct {
	WorkerPool    *WorkerPool
	BuildQueue    chan BuildRequest
	ActiveBuilds  map[string]chan BuildResponse
	mutex         sync.RWMutex
	httpServer    *http.Server
	rpcServer     *rpc.Server
	listener      net.Listener
}

// NewBuildCoordinator creates a new build coordinator
func NewBuildCoordinator(maxWorkers int) *BuildCoordinator {
	coordinator := &BuildCoordinator{
		WorkerPool:   NewWorkerPool(maxWorkers),
		BuildQueue:   make(chan BuildRequest, 100),
		ActiveBuilds: make(map[string]chan BuildResponse),
	}
	
	// Initialize RPC server
	coordinator.rpcServer = rpc.NewServer()
	coordinator.rpcServer.Register(coordinator)
	
	return coordinator
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(maxWorkers int) *WorkerPool {
	return &WorkerPool{
		Workers:    make(map[string]*Worker),
		maxWorkers: maxWorkers,
	}
}

// AddWorker adds a new worker to the pool
func (wp *WorkerPool) AddWorker(worker *Worker) error {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()
	
	if len(wp.Workers) >= wp.maxWorkers {
		return fmt.Errorf("worker pool at maximum capacity")
	}
	
	worker.Status = "idle"
	worker.LastCheckin = time.Now()
	wp.Workers[worker.ID] = worker
	
	log.Printf("Added worker %s (%s:%d) to pool", worker.ID, worker.Host, worker.Port)
	return nil
}

// RemoveWorker removes a worker from the pool
func (wp *WorkerPool) RemoveWorker(workerID string) error {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()
	
	if _, exists := wp.Workers[workerID]; !exists {
		return fmt.Errorf("worker %s not found", workerID)
	}
	
	delete(wp.Workers, workerID)
	log.Printf("Removed worker %s from pool", workerID)
	return nil
}

// GetAvailableWorker finds an available worker for the given task
func (wp *WorkerPool) GetAvailableWorker(taskType string) (*Worker, error) {
	wp.mutex.RLock()
	defer wp.mutex.RUnlock()
	
	for _, worker := range wp.Workers {
		worker.mutex.RLock()
		isAvailable := worker.Status == "idle"
		hasCapability := false
		
		for _, capability := range worker.Capabilities {
			if capability == taskType || capability == "all" {
				hasCapability = true
				break
			}
		}
		
		worker.mutex.RUnlock()
		
		if isAvailable && hasCapability {
			return worker, nil
		}
	}
	
	return nil, fmt.Errorf("no available workers for task type: %s", taskType)
}

// UpdateWorkerStatus updates a worker's status
func (wp *WorkerPool) UpdateWorkerStatus(workerID string, status string) error {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()
	
	worker, exists := wp.Workers[workerID]
	if !exists {
		return fmt.Errorf("worker %s not found", workerID)
	}
	
	worker.mutex.Lock()
	worker.Status = status
	worker.LastCheckin = time.Now()
	worker.mutex.Unlock()
	
	return nil
}

// SubmitBuild submits a build request to the coordinator
func (bc *BuildCoordinator) SubmitBuild(request BuildRequest) (string, error) {
	responseChan := make(chan BuildResponse)
	
	bc.mutex.Lock()
	bc.ActiveBuilds[request.RequestID] = responseChan
	bc.mutex.Unlock()
	
	bc.BuildQueue <- request
	
	return request.RequestID, nil
}

// ProcessBuild processes a build request
func (bc *BuildCoordinator) ProcessBuild(request BuildRequest) BuildResponse {
	worker, err := bc.WorkerPool.GetAvailableWorker("gradle")
	if err != nil {
		return BuildResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("No available workers: %v", err),
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
		return BuildResponse{
			Success:       false,
			WorkerID:      worker.ID,
			BuildDuration: duration,
			ErrorMessage:  fmt.Sprintf("Build failed: %v\nOutput: %s", err, string(output)),
			RequestID:     request.RequestID,
			Timestamp:     time.Now(),
		}
	}
	
	worker.BuildCount++
	
	return BuildResponse{
		Success:       true,
		WorkerID:      worker.ID,
		BuildDuration: duration,
		Artifacts:     bc.findArtifacts(request.ProjectPath),
		RequestID:     request.RequestID,
		Timestamp:     time.Now(),
		Metrics: BuildMetrics{
			CacheHitRate: 0.85, // Placeholder - should be calculated
			CompiledFiles: 42,  // Placeholder - should be calculated
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
		go func(req BuildRequest) {
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
	
	var request BuildRequest
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
	bc.WorkerPool.mutex.RLock()
	workers := bc.WorkerPool.Workers
	bc.WorkerPool.mutex.RUnlock()
	
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

func main() {
	coordinator := NewBuildCoordinator(10)
	
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
	select {}
}