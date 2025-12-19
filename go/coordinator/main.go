package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
	
	"distributed-gradle-building/types"
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
	Name        string        `json:"name"`
	Duration    time.Duration `json:"duration"`
	Status      string        `json:"status"`
	Artifacts   []string      `json:"artifacts"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
}

// TestResults represents test execution results
type TestResults struct {
	Total    int `json:"total"`
	Passed   int `json:"passed"`
	Failed   int `json:"failed"`
	Skipped  int `json:"skipped"`
	Duration time.Duration `json:"duration"`
}

// ResourceMetrics tracks resource usage during builds
type ResourceMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage int64   `json:"memory_usage"`
	DiskIO      int64   `json:"disk_io"`
	NetworkIO   int64   `json:"network_io"`
}

// Worker represents a build worker node
type Worker struct {
	ID           string            `json:"id"`
	Host         string            `json:"host"`
	Port         int               `json:"port"`
	Status       string            `json:"status"`
	Capabilities []string          `json:"capabilities"`
	LastPing     time.Time         `json:"last_ping"`
	Builds       []BuildRequest    `json:"builds"`
	Metrics      WorkerMetrics     `json:"metrics"`
}

// WorkerMetrics tracks worker performance
type WorkerMetrics struct {
	BuildCount         int           `json:"build_count"`
	TotalBuildTime     time.Duration `json:"total_build_time"`
	AverageBuildTime   time.Duration `json:"average_build_time"`
	SuccessRate        float64       `json:"success_rate"`
	LastBuildTime      time.Time     `json:"last_build_time"`
}

// BuildCoordinator manages the distributed build system
type BuildCoordinator struct {
	workers      map[string]*Worker
	buildQueue   chan BuildRequest
	builds       map[string]*BuildResponse
	mutex        sync.RWMutex
	httpServer   *http.Server
	rpcServer    *rpc.Server
	shutdown     chan struct{}
	maxWorkers   int
}

// NewBuildCoordinator creates a new build coordinator
func NewBuildCoordinator(maxWorkers int) *BuildCoordinator {
	return &BuildCoordinator{
		workers:    make(map[string]*Worker),
		buildQueue: make(chan types.BuildRequest, 100),
		builds:     make(map[string]*types.BuildResponse),
		shutdown:   make(chan struct{}),
		maxWorkers: maxWorkers,
	}
}

// RegisterWorker adds a new worker to the pool
func (bc *BuildCoordinator) RegisterWorker(worker *Worker) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	if len(bc.workers) >= bc.maxWorkers {
		return fmt.Errorf("maximum workers (%d) reached", bc.maxWorkers)
	}
	
	worker.Status = "idle"
	worker.LastPing = time.Now()
	bc.workers[worker.ID] = worker
	
	log.Printf("Worker %s registered from %s:%d", worker.ID, worker.Host, worker.Port)
	return nil
}

// UnregisterWorker removes a worker from the pool
func (bc *BuildCoordinator) UnregisterWorker(workerID string) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	if _, exists := bc.workers[workerID]; exists {
		delete(bc.workers, workerID)
		log.Printf("Worker %s unregistered", workerID)
	}
}

// SubmitBuild adds a build request to the queue
func (bc *BuildCoordinator) SubmitBuild(request types.BuildRequest) (string, error) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	if request.RequestID == "" {
		request.RequestID = generateBuildID()
	}
	
	request.Timestamp = time.Now()
	
	// Store initial build response
	response := &types.BuildResponse{
		RequestID: request.RequestID,
		Timestamp: time.Now(),
		Success:   false,
	}
	bc.builds[request.RequestID] = response
	
	// Add to queue
	select {
	case bc.buildQueue <- request:
		log.Printf("Build %s queued for project %s", request.RequestID, request.ProjectPath)
		return request.RequestID, nil
	default:
		return "", fmt.Errorf("build queue is full")
	}
}

// GetBuildStatus returns the status of a build
func (bc *BuildCoordinator) GetBuildStatus(buildID string) (*types.BuildResponse, error) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	response, exists := bc.builds[buildID]
	if !exists {
		return nil, fmt.Errorf("build %s not found", buildID)
	}
	
	return response, nil
}

// GetWorkers returns the list of registered workers
func (bc *BuildCoordinator) GetWorkers() []Worker {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()
	
	workers := make([]Worker, 0, len(bc.workers))
	for _, worker := range bc.workers {
		workers = append(workers, *worker)
	}
	
	return workers
}

// StartHTTPServer starts the HTTP API server
func (bc *BuildCoordinator) StartHTTPServer(port int) error {
	mux := http.NewServeMux()
	
	// API endpoints
	mux.HandleFunc("/api/build", bc.handleBuildRequest)
	mux.HandleFunc("/api/builds/", bc.handleGetBuild)
	mux.HandleFunc("/api/workers", bc.handleGetWorkers)
	mux.HandleFunc("/api/health", bc.handleHealthCheck)
	
	bc.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	
	log.Printf("HTTP server listening on port %d", port)
	return bc.httpServer.ListenAndServe()
}

// StartRPCServer starts the RPC server
func (bc *BuildCoordinator) StartRPCServer(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	
	bc.rpcServer = rpc.NewServer()
	rpc.Register(bc)
	
	log.Printf("RPC server listening on port %d", port)
	go bc.rpcServer.Accept(listener)
	
	return nil
}

// BuildQueueProcessor handles the build queue
func (bc *BuildCoordinator) BuildQueueProcessor() {
	for {
		select {
		case request := <-bc.buildQueue:
			go bc.processBuild(request)
		case <-bc.shutdown:
			return
		}
	}
}

// processBuild assigns a build to an available worker
func (bc *BuildCoordinator) processBuild(request BuildRequest) {
	bc.mutex.RLock()
	availableWorkers := bc.getAvailableWorkers()
	bc.mutex.RUnlock()
	
	if len(availableWorkers) == 0 {
		// Re-queue if no workers available
		time.Sleep(5 * time.Second)
		select {
		case bc.buildQueue <- request:
			return
		case <-bc.shutdown:
			return
		default:
			// Mark as failed if queue is full
			bc.markBuildFailed(request.RequestID, "no workers available")
			return
		}
	}
	
	// Assign to first available worker
	worker := availableWorkers[0]
	bc.assignBuildToWorker(worker, request)
}

// getAvailableWorkers returns idle workers
func (bc *BuildCoordinator) getAvailableWorkers() []*Worker {
	var available []*Worker
	for _, worker := range bc.workers {
		if worker.Status == "idle" && time.Since(worker.LastPing) < 30*time.Second {
			available = append(available, worker)
		}
	}
	return available
}

// assignBuildToWorker assigns a build to a specific worker
func (bc *BuildCoordinator) assignBuildToWorker(worker *Worker, request BuildRequest) {
	worker.Status = "busy"
	worker.Builds = append(worker.Builds, request)
	
	// Execute build via RPC
	go bc.executeBuildOnWorker(worker, request)
}

// executeBuildOnWorker executes a build on a remote worker
func (bc *BuildCoordinator) executeBuildOnWorker(worker *Worker, request BuildRequest) {
	defer func() {
		worker.Status = "idle"
		worker.LastPing = time.Now()
	}()
	
	// Connect to worker RPC server
	client, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", worker.Host, worker.Port))
	if err != nil {
		bc.markBuildFailed(request.RequestID, fmt.Sprintf("failed to connect to worker: %v", err))
		return
	}
	defer client.Close()
	
	// Execute build
	var response string
	err = client.Call("WorkerService.Build", request, &response)
	if err != nil {
		bc.markBuildFailed(request.RequestID, fmt.Sprintf("build failed: %v", err))
		return
	}
	
	// Mark as successful
	bc.markBuildCompleted(request.RequestID, worker.ID)
}

// markBuildCompleted marks a build as completed successfully
func (bc *BuildCoordinator) markBuildCompleted(buildID, workerID string) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	if response, exists := bc.builds[buildID]; exists {
		response.Success = true
		response.WorkerID = workerID
		response.Timestamp = time.Now()
	}
}

// markBuildFailed marks a build as failed
func (bc *BuildCoordinator) markBuildFailed(buildID, errorMsg string) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	
	if response, exists := bc.builds[buildID]; exists {
		response.Success = false
		response.ErrorMessage = errorMsg
		response.Timestamp = time.Now()
	}
}

// Shutdown gracefully shuts down the coordinator
func (bc *BuildCoordinator) Shutdown() {
	close(bc.shutdown)
	
	if bc.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		bc.httpServer.Shutdown(ctx)
	}
	
	if bc.rpcServer != nil {
		// RPC server doesn't have explicit close method in standard library
		// The listener will be closed when the process exits
		log.Println("RPC server shutdown requested")
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
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	buildID, err := bc.SubmitBuild(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(map[string]string{"build_id": buildID})
}

func (bc *BuildCoordinator) handleGetBuild(w http.ResponseWriter, r *http.Request) {
	buildID := r.URL.Path[len("/api/builds/"):]
	
	response, err := bc.GetBuildStatus(buildID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	
	json.NewEncoder(w).Encode(response)
}

func (bc *BuildCoordinator) handleGetWorkers(w http.ResponseWriter, r *http.Request) {
	workers := bc.GetWorkers()
	json.NewEncoder(w).Encode(workers)
}

func (bc *BuildCoordinator) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// Utility functions
func generateBuildID() string {
	return fmt.Sprintf("build-%d", time.Now().UnixNano())
}

// Main coordinator application entry point
func coordinatorMain() {
	coordinator := NewBuildCoordinator(10)
	
	// Start build queue processor
	go coordinator.BuildQueueProcessor()
	
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

// Main function for coordinator
func main() {
	coordinatorMain()
}