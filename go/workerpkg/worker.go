package workerpkg

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"distributed-gradle-building/types"
)

// WorkerService represents the build worker service
type WorkerService struct {
	ID           string
	Config       types.WorkerConfig
	Coordinator  string
	BuildQueue   chan types.BuildRequest
	ActiveBuilds map[string]*types.BuildResponse
	LastPing     time.Time
	Mutex        sync.RWMutex
	BuildDir     string
}

// NewWorkerService creates a new worker service
func NewWorkerService(id, coordinator string, config types.WorkerConfig) *WorkerService {
	return &WorkerService{
		ID:           id,
		Config:       config,
		Coordinator:  coordinator,
		BuildQueue:   make(chan types.BuildRequest, 10),
		ActiveBuilds: make(map[string]*types.BuildResponse),
		LastPing:     time.Now(),
		BuildDir:     config.BuildDir,
	}
}

// RegisterRPCMethods registers the RPC methods for the worker
func (ws *WorkerService) RegisterRPCMethods() {
	rpc.Register(ws)
}

// Ping responds to coordinator ping
func (ws *WorkerService) Ping(args *struct{}, reply *struct{}) error {
	ws.Mutex.Lock()
	ws.LastPing = time.Now()
	ws.Mutex.Unlock()
	return nil
}

// ExecuteBuild executes a build request
func (ws *WorkerService) ExecuteBuild(request types.BuildRequest, response *types.BuildResponse) error {
	log.Printf("Worker %s executing build %s for project %s", ws.ID, request.RequestID, request.ProjectPath)

	// Initialize response
	*response = types.BuildResponse{
		RequestID: request.RequestID,
		WorkerID:  ws.ID,
		Timestamp: time.Now(),
		Success:   false,
	}

	// Create build directory
	buildDir := filepath.Join(ws.BuildDir, request.RequestID)
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		response.ErrorMessage = fmt.Sprintf("Failed to create build directory: %v", err)
		return nil
	}

	// Execute build
	startTime := time.Now()
	err := ws.runGradleBuild(request, response)
	response.BuildDuration = time.Since(startTime)

	if err != nil {
		response.ErrorMessage = fmt.Sprintf("Build failed: %v", err)
	} else {
		response.Success = true
		response.Artifacts = ws.collectArtifacts(buildDir)
	}

	// Clean up build directory
	os.RemoveAll(buildDir)

	log.Printf("Worker %s completed build %s in %v", ws.ID, request.RequestID, response.BuildDuration)
	return nil
}

// GetStatus returns the current worker status
func (ws *WorkerService) GetStatus(args *struct{}, status *WorkerStatus) error {
	ws.Mutex.RLock()
	defer ws.Mutex.RUnlock()

	*status = WorkerStatus{
		ID:           ws.ID,
		ActiveBuilds: len(ws.ActiveBuilds),
		QueueLength:  len(ws.BuildQueue),
		LastPing:     ws.LastPing,
		BuildDir:     ws.BuildDir,
		IsHealthy:    time.Since(ws.LastPing) < 5*time.Minute,
		WorkerType:   ws.Config.WorkerType,
	}

	return nil
}

// WorkerStatus represents the current status of a worker
type WorkerStatus struct {
	ID           string    `json:"id"`
	ActiveBuilds int       `json:"active_builds"`
	QueueLength  int       `json:"queue_length"`
	LastPing     time.Time `json:"last_ping"`
	BuildDir     string    `json:"build_dir"`
	IsHealthy    bool      `json:"is_healthy"`
	WorkerType   string    `json:"worker_type"`
}

// StartServer starts the worker's RPC server
func (ws *WorkerService) StartServer() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", ws.Config.RPCPort))
	if err != nil {
		return fmt.Errorf("failed to start RPC listener: %v", err)
	}

	ws.RegisterRPCMethods()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				break
			}
			go rpc.ServeConn(conn)
		}
	}()

	log.Printf("Worker %s RPC server listening on port %d", ws.ID, ws.Config.RPCPort)
	return nil
}

// runGradleBuild executes the actual Gradle build
func (ws *WorkerService) runGradleBuild(request types.BuildRequest, response *types.BuildResponse) error {
	// Prepare Gradle command
	args := []string{request.TaskName}

	// Add build options
	for key, value := range request.BuildOptions {
		args = append(args, fmt.Sprintf("--%s=%s", key, value))
	}

	// Create command
	cmd := exec.Command("gradle", args...)
	cmd.Dir = request.ProjectPath

	// Capture output
	output, err := cmd.CombinedOutput()

	// Update response metrics
	response.Metrics = types.BuildMetrics{
		BuildSteps: []types.BuildStep{
			{
				Name:     request.TaskName,
				Duration: response.BuildDuration,
				CacheHit: request.CacheEnabled,
			},
		},
		TestResults: types.TestResults{
			TotalTests:  0,
			PassedTests: 0,
			FailedTests: 0,
		},
	}

	if err != nil {
		response.ErrorMessage = string(output)
		return err
	}

	return nil
}

// collectArtifacts collects build artifacts from the build directory
func (ws *WorkerService) collectArtifacts(buildDir string) []string {
	var artifacts []string

	filepath.Walk(buildDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() {
			relPath, _ := filepath.Rel(buildDir, path)
			artifacts = append(artifacts, relPath)
		}

		return nil
	})

	return artifacts
}

// Shutdown gracefully shuts down the worker
func (ws *WorkerService) Shutdown() error {
	close(ws.BuildQueue)
	return nil
}
