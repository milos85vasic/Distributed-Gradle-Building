package coordinatorpkg

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

// Worker represents a build worker node
type Worker struct {
	ID       string
	Host     string
	Port     int
	Status   string
	LastPing time.Time
	Metadata map[string]interface{}
}

// BuildCoordinator manages distributed builds across workers
type BuildCoordinator struct {
	workers    map[string]*Worker
	buildQueue chan types.BuildRequest
	builds     map[string]*types.BuildResponse
	mutex      sync.RWMutex
	httpServer *http.Server
	rpcServer  *rpc.Server
	shutdown   chan struct{}
	maxWorkers int
}

// RPC argument and reply types for worker registration
type RegisterWorkerArgs struct {
	ID           string   `json:"id"`
	Host         string   `json:"host"`
	Port         int      `json:"port"`
	Capabilities []string `json:"capabilities"`
	Status       string   `json:"status"`
}

type RegisterWorkerReply struct {
	Message string `json:"message"`
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

// RegisterWorker adds a new worker to the pool (internal method)
func (bc *BuildCoordinator) RegisterWorker(worker *Worker) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if len(bc.workers) >= bc.maxWorkers {
		return fmt.Errorf("maximum workers (%d) reached", bc.maxWorkers)
	}

	bc.workers[worker.ID] = worker
	log.Printf("Worker %s registered from %s:%d", worker.ID, worker.Host, worker.Port)
	return nil
}

// RegisterWorkerRPC is the RPC method for worker registration
func (bc *BuildCoordinator) RegisterWorkerRPC(args *RegisterWorkerArgs, reply *RegisterWorkerReply) error {
	log.Printf("RegisterWorkerRPC called with args: %+v", args)
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if len(bc.workers) >= bc.maxWorkers {
		return fmt.Errorf("maximum workers (%d) reached", bc.maxWorkers)
	}

	worker := &Worker{
		ID:       args.ID,
		Host:     args.Host,
		Port:     args.Port,
		Status:   "idle",
		LastPing: time.Now(),
		Metadata: map[string]interface{}{
			"capabilities": args.Capabilities,
		},
	}

	bc.workers[worker.ID] = worker

	log.Printf("Worker %s registered from %s:%d", worker.ID, worker.Host, worker.Port)
	reply.Message = fmt.Sprintf("Worker %s registered successfully", worker.ID)
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

// StartServer starts the HTTP server
func (bc *BuildCoordinator) StartServer(port int) error {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/builds", bc.handleBuilds)
	mux.HandleFunc("/api/workers", bc.handleWorkers)
	mux.HandleFunc("/api/health", bc.handleHealth)

	bc.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Printf("HTTP server listening on port %d", port)
	return bc.httpServer.ListenAndServe()
}

// StartRPCServer starts the RPC server for worker registration
func (bc *BuildCoordinator) StartRPCServer(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to start RPC listener: %v", err)
	}

	bc.rpcServer = rpc.NewServer()
	err = bc.rpcServer.RegisterName("BuildCoordinator", bc)
	if err != nil {
		return fmt.Errorf("RPC registration failed: %v", err)
	}

	log.Printf("RPC server listening on port %d", port)
	go bc.rpcServer.Accept(listener)

	return nil
}

// handleHealth handles health check requests
func (bc *BuildCoordinator) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// Shutdown gracefully shuts down the coordinator
func (bc *BuildCoordinator) Shutdown() error {
	close(bc.shutdown)

	if bc.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return bc.httpServer.Shutdown(ctx)
	}

	return nil
}

// generateBuildID generates a unique build ID
func generateBuildID() string {
	return fmt.Sprintf("build_%d", time.Now().UnixNano())
}

// handleBuilds handles build-related HTTP requests
func (bc *BuildCoordinator) handleBuilds(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var request types.BuildRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		buildID, err := bc.SubmitBuild(request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"build_id": buildID})

	case http.MethodGet:
		buildID := r.URL.Query().Get("id")
		if buildID == "" {
			http.Error(w, "Missing build_id parameter", http.StatusBadRequest)
			return
		}

		response, err := bc.GetBuildStatus(buildID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleWorkers handles worker-related HTTP requests
func (bc *BuildCoordinator) handleWorkers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	workers := bc.GetWorkers()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workers)
}

// handleStatus handles status requests
func (bc *BuildCoordinator) HandleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"workers": len(bc.workers),
		"queue":   len(bc.buildQueue),
		"builds":  len(bc.builds),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// HandleBuilds handles build-related HTTP requests (exported for testing)
func (bc *BuildCoordinator) HandleBuilds(w http.ResponseWriter, r *http.Request) {
	bc.handleBuilds(w, r)
}

// HandleWorkers handles worker-related HTTP requests (exported for testing)
func (bc *BuildCoordinator) HandleWorkers(w http.ResponseWriter, r *http.Request) {
	bc.handleWorkers(w, r)
}
