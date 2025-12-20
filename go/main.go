package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
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

	"distributed-gradle-building/ml/service"
	"distributed-gradle-building/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "net/http/pprof"
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
	MLService    *service.MLService
	mutex        sync.RWMutex
	httpServer   *http.Server
	rpcServer    *rpc.Server
	listener     net.Listener
	shutdown     chan struct{}
	startTime    time.Time
}

// Prometheus metrics for coordinator
var (
	activeBuilds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_builds",
			Help: "Number of currently active builds",
		},
	)
	buildRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "build_requests_total",
			Help: "Total number of build requests",
		},
		[]string{"status"},
	)
	buildDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "build_duration_seconds",
			Help:    "Build duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)
	coordinatorHTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
)

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
		MLService:    service.NewMLService(),
		shutdown:     make(chan struct{}),
		startTime:    time.Now(),
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
	activeBuilds.Set(float64(len(bc.ActiveBuilds)))
	bc.mutex.Unlock()

	bc.BuildQueue <- request
	buildRequestsTotal.WithLabelValues("submitted").Inc()

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
	// Get ML predictions for failure risk assessment
	predictions := bc.MLService.GetBuildInsights(request.ProjectPath, request.TaskName, request.BuildOptions)

	// Handle high-risk builds with mitigation strategies
	if predictions.FailureRisk > 0.7 {
		log.Printf("High failure risk detected for build %s (%.2f), applying mitigation strategies",
			request.RequestID, predictions.FailureRisk)
		return bc.processHighRiskBuild(request, predictions)
	}

	// Try to get the best available worker using ML predictions
	var worker *types.Worker
	var err error

	for retry := 0; retry < 3; retry++ {
		worker, err = bc.selectBestWorkerForBuild(request)
		if err == nil {
			break // Found suitable worker
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
		buildRequestsTotal.WithLabelValues("failed").Inc()
		buildDuration.WithLabelValues("failed").Observe(duration.Seconds())
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

	// Update Prometheus metrics
	buildRequestsTotal.WithLabelValues("successful").Inc()
	buildDuration.WithLabelValues("successful").Observe(duration.Seconds())

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

// processHighRiskBuild handles builds with high failure risk using mitigation strategies
func (bc *BuildCoordinator) processHighRiskBuild(request types.BuildRequest, predictions service.PredictionResult) types.BuildResponse {
	log.Printf("Applying failure mitigation for high-risk build %s", request.RequestID)

	// Strategy 1: Select most reliable worker (with best historical success rate)
	worker, err := bc.selectMostReliableWorkerForBuild(request)
	if err != nil {
		return types.BuildResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("No reliable workers available for high-risk build: %v", err),
			RequestID:    request.RequestID,
			Timestamp:    time.Now(),
		}
	}

	// Mark worker as busy
	bc.WorkerPool.UpdateWorkerStatus(worker.ID, "busy")
	defer bc.WorkerPool.UpdateWorkerStatus(worker.ID, "idle")

	startTime := time.Now()

	// Strategy 2: Enhanced monitoring and logging
	log.Printf("Starting high-risk build %s on reliable worker %s", request.RequestID, worker.ID)

	// Execute build with enhanced error handling
	cmd := exec.Command("gradle", request.TaskName)
	cmd.Dir = request.ProjectPath
	cmd.Env = append(os.Environ(),
		"GRADLE_OPTS=-Xmx2g -XX:+HeapDumpOnOutOfMemoryError",            // More memory for stability
		fmt.Sprintf("BUILD_FAILURE_RISK=%.2f", predictions.FailureRisk), // Pass risk to build
	)

	if request.CacheEnabled {
		cmd.Env = append(cmd.Env, "GRADLE_USER_HOME=/tmp/gradle-cache")
	}

	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	if err != nil {
		buildRequestsTotal.WithLabelValues("failed").Inc()
		buildDuration.WithLabelValues("failed").Observe(duration.Seconds())

		// Strategy 3: Detailed failure analysis and reporting
		failureAnalysis := bc.analyzeBuildFailure(string(output), predictions)
		log.Printf("High-risk build %s failed: %s", request.RequestID, failureAnalysis)

		return types.BuildResponse{
			Success:       false,
			WorkerID:      worker.ID,
			BuildDuration: duration,
			ErrorMessage:  fmt.Sprintf("High-risk build failed: %v\nFailure Analysis: %s\nOutput: %s", err, failureAnalysis, string(output)),
			RequestID:     request.RequestID,
			Timestamp:     time.Now(),
		}
	}

	worker.BuildCount++

	// Calculate metrics
	cacheHits, totalRequests := bc.calculateCacheMetrics(request.ProjectPath)
	compiledFiles := bc.countCompiledFiles(request.ProjectPath)

	buildRequestsTotal.WithLabelValues("successful").Inc()
	buildDuration.WithLabelValues("successful").Observe(duration.Seconds())

	log.Printf("High-risk build %s completed successfully on reliable worker %s", request.RequestID, worker.ID)

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

// selectMostReliableWorkerForBuild selects the most reliable worker for high-risk builds
func (bc *BuildCoordinator) selectMostReliableWorkerForBuild(request types.BuildRequest) (*types.Worker, error) {
	bc.WorkerPool.WorkerPool.Mutex.RLock()
	defer bc.WorkerPool.WorkerPool.Mutex.RUnlock()

	var bestWorker *types.Worker
	var bestReliability float64 = -1

	for _, worker := range bc.WorkerPool.WorkerPool.Workers {
		if worker.Status != "idle" {
			continue
		}

		// Check capabilities
		hasCapability := false
		for _, capability := range worker.Capabilities {
			if capability == request.TaskName || capability == "gradle" || capability == "all" {
				hasCapability = true
				break
			}
		}
		if !hasCapability {
			continue
		}

		// Calculate reliability score based on build history
		reliability := bc.calculateWorkerReliability(worker)
		if reliability > bestReliability {
			bestReliability = reliability
			bestWorker = worker
		}
	}

	if bestWorker == nil {
		return nil, fmt.Errorf("no reliable workers available")
	}

	log.Printf("Selected most reliable worker %s (reliability: %.2f) for high-risk build", bestWorker.ID, bestReliability)
	return bestWorker, nil
}

// calculateWorkerReliability calculates a worker's reliability score
func (bc *BuildCoordinator) calculateWorkerReliability(worker *types.Worker) float64 {
	if worker.BuildCount == 0 {
		return 0.5 // Neutral score for new workers
	}

	// This is a simplified reliability calculation
	// In production, you'd track success rates per worker
	successRate := 0.8 // Placeholder - would be calculated from actual metrics

	// Factor in recent performance and system health
	return successRate
}

// analyzeBuildFailure provides detailed analysis of build failures
func (bc *BuildCoordinator) analyzeBuildFailure(output string, predictions service.PredictionResult) string {
	analysis := "Build failure analysis:\n"

	// Check for common failure patterns
	if strings.Contains(output, "OutOfMemoryError") {
		analysis += "- Memory exhaustion detected. Consider increasing heap size.\n"
	}
	if strings.Contains(output, "Compilation failed") {
		analysis += "- Compilation errors detected. Check source code for syntax issues.\n"
	}
	if strings.Contains(output, "Test failed") {
		analysis += "- Test failures detected. Review test cases and assertions.\n"
	}
	if strings.Contains(output, "Dependency resolution") {
		analysis += "- Dependency issues detected. Check artifact repositories and network connectivity.\n"
	}

	analysis += fmt.Sprintf("- Predicted failure risk was: %.2f\n", predictions.FailureRisk)
	analysis += fmt.Sprintf("- Suggested resource allocation: CPU=%.1f, Memory=%.1f, Disk=%.1f\n",
		predictions.ResourceNeeds.CPU, predictions.ResourceNeeds.Memory, predictions.ResourceNeeds.Disk)

	return analysis
}

// selectBestWorkerForBuild selects the most suitable worker for a build using ML predictions
func (bc *BuildCoordinator) selectBestWorkerForBuild(request types.BuildRequest) (*types.Worker, error) {
	bc.WorkerPool.WorkerPool.Mutex.RLock()
	defer bc.WorkerPool.WorkerPool.Mutex.RUnlock()

	// Get ML predictions for this build
	predictions := bc.MLService.GetBuildInsights(request.ProjectPath, request.TaskName, request.BuildOptions)

	var bestWorker *types.Worker
	var bestScore float64 = -1

	// Score each available worker
	for _, worker := range bc.WorkerPool.WorkerPool.Workers {
		if worker.Status != "idle" {
			continue // Skip busy workers
		}

		// Check if worker has required capabilities
		hasCapability := false
		for _, capability := range worker.Capabilities {
			if capability == request.TaskName || capability == "gradle" || capability == "all" {
				hasCapability = true
				break
			}
		}
		if !hasCapability {
			continue
		}

		// Calculate worker score based on ML predictions
		score := bc.calculateWorkerScore(worker, predictions)

		if score > bestScore {
			bestScore = score
			bestWorker = worker
		}
	}

	if bestWorker == nil {
		return nil, fmt.Errorf("no suitable workers available for build %s", request.RequestID)
	}

	return bestWorker, nil
}

// calculateWorkerScore calculates how suitable a worker is for a build based on ML predictions
func (bc *BuildCoordinator) calculateWorkerScore(worker *types.Worker, predictions service.PredictionResult) float64 {
	score := 0.0

	// Base score for being available
	score += 10.0

	// Factor in predicted resource needs vs worker capabilities
	// This is a simplified scoring - in production you'd have more sophisticated logic
	resourceFit := 1.0 - (math.Abs(predictions.ResourceNeeds.CPU-0.5)+math.Abs(predictions.ResourceNeeds.Memory-0.5))/2.0
	score += resourceFit * 5.0

	// Factor in cache hit rate prediction (higher cache hit rate = better score)
	score += predictions.CacheHitRate * 3.0

	// Factor in failure risk (lower risk = better score)
	score += (1.0 - predictions.FailureRisk) * 2.0

	// Factor in worker performance history
	if worker.BuildCount > 0 {
		// Prefer workers with good track record (this would be expanded with real metrics)
		score += 1.0
	}

	return score
}

// startAutoScaling starts the auto-scaling monitoring goroutine
func (bc *BuildCoordinator) startAutoScaling() {
	go func() {
		ticker := time.NewTicker(30 * time.Second) // Check scaling every 30 seconds
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				bc.checkAndPerformScaling()
			case <-bc.shutdown:
				return
			}
		}
	}()
}

// checkAndPerformScaling checks if scaling is needed and performs the action
func (bc *BuildCoordinator) checkAndPerformScaling() {
	bc.mutex.RLock()
	queueLength := len(bc.BuildQueue)
	activeBuilds := len(bc.ActiveBuilds)
	bc.mutex.RUnlock()

	bc.WorkerPool.WorkerPool.Mutex.RLock()
	currentWorkers := len(bc.WorkerPool.WorkerPool.Workers)
	busyWorkers := 0
	avgCPULoad := 0.0
	workerCount := 0

	for _, worker := range bc.WorkerPool.WorkerPool.Workers {
		if worker.Status == "busy" {
			busyWorkers++
		}
		// In a real implementation, you'd collect actual CPU metrics
		// For now, estimate based on busy/idle status
		if worker.Status == "busy" {
			avgCPULoad += 0.8 // Assume busy workers are at 80% CPU
		} else {
			avgCPULoad += 0.2 // Assume idle workers are at 20% CPU
		}
		workerCount++
	}

	if workerCount > 0 {
		avgCPULoad /= float64(workerCount)
	}
	bc.WorkerPool.WorkerPool.Mutex.RUnlock()

	// Get scaling recommendation from ML service
	scalingAdvice := bc.MLService.PredictScalingNeeds(queueLength, avgCPULoad, currentWorkers)

	log.Printf("Scaling check: queue=%d, active=%d, workers=%d/%d, avg_cpu=%.2f, recommendation=%s (%d workers)",
		queueLength, activeBuilds, busyWorkers, currentWorkers, avgCPULoad,
		scalingAdvice.Action, scalingAdvice.WorkersNeeded)

	// Perform scaling action
	switch scalingAdvice.Action {
	case "scale_up":
		bc.performScaleUp(scalingAdvice.WorkersNeeded - currentWorkers)
	case "scale_down":
		bc.performScaleDown(currentWorkers - scalingAdvice.WorkersNeeded)
	case "maintain":
		// Do nothing
	}
}

// performScaleUp adds new workers to the pool
func (bc *BuildCoordinator) performScaleUp(workersToAdd int) {
	if workersToAdd <= 0 {
		return
	}

	log.Printf("Scaling up: adding %d workers", workersToAdd)

	// In a real implementation, this would provision new worker instances
	// For now, we'll just log the scaling action
	for i := 0; i < workersToAdd; i++ {
		workerID := fmt.Sprintf("auto-scaled-worker-%d-%d", time.Now().Unix(), i)
		log.Printf("Would provision new worker: %s", workerID)
		// TODO: Implement actual worker provisioning logic
	}
}

// performScaleDown removes workers from the pool
func (bc *BuildCoordinator) performScaleDown(workersToRemove int) {
	if workersToRemove <= 0 {
		return
	}

	log.Printf("Scaling down: removing %d workers", workersToRemove)

	// In a real implementation, this would gracefully shut down workers
	// For now, we'll just log the scaling action
	bc.WorkerPool.WorkerPool.Mutex.Lock()
	defer bc.WorkerPool.WorkerPool.Mutex.Unlock()

	removed := 0
	for workerID, worker := range bc.WorkerPool.WorkerPool.Workers {
		if worker.Status == "idle" && removed < workersToRemove {
			log.Printf("Would shut down idle worker: %s", workerID)
			// TODO: Implement actual worker shutdown logic
			removed++
		}
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
	mux.Handle("/metrics", promhttp.Handler())

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
	go bc.startAutoScaling() // Start auto-scaling monitoring
	go bc.rpcServer.Accept(bc.listener)

	return nil
}

// processBuildQueue processes queued build requests with ML-based prioritization
func (bc *BuildCoordinator) processBuildQueue() {
	// Use a priority queue to process builds based on ML predictions
	buildQueue := make(chan types.BuildRequest, 100)

	go func() {
		for request := range bc.BuildQueue {
			// Calculate priority score for the build
			priority := bc.calculateBuildPriority(request)

			// For now, we'll process immediately if priority is high, otherwise queue
			// In a full implementation, you'd use a proper priority queue
			if priority > 7.0 { // High priority builds
				go bc.processBuildWithPriority(request, priority)
			} else {
				select {
				case buildQueue <- request:
				default:
					// Queue full, process immediately
					go bc.processBuildWithPriority(request, priority)
				}
			}
		}
	}()

	// Process lower priority builds from the secondary queue
	for request := range buildQueue {
		go bc.processBuildWithPriority(request, bc.calculateBuildPriority(request))
	}
}

// processBuildWithPriority processes a build with priority logging
func (bc *BuildCoordinator) processBuildWithPriority(request types.BuildRequest, priority float64) {
	log.Printf("Processing build %s with priority %.2f", request.RequestID, priority)
	response := bc.ProcessBuild(request)

	bc.mutex.Lock()
	if responseChan, exists := bc.ActiveBuilds[request.RequestID]; exists {
		responseChan <- response
		delete(bc.ActiveBuilds, request.RequestID)
		activeBuilds.Set(float64(len(bc.ActiveBuilds)))
	}
	bc.mutex.Unlock()
}

// calculateBuildPriority calculates a priority score for a build based on ML predictions
func (bc *BuildCoordinator) calculateBuildPriority(request types.BuildRequest) float64 {
	predictions := bc.MLService.GetBuildInsights(request.ProjectPath, request.TaskName, request.BuildOptions)

	score := 5.0 // Base priority

	// Higher priority for builds with high failure risk (need to fail fast)
	score += predictions.FailureRisk * 3.0

	// Higher priority for builds predicted to be short (quick wins)
	if predictions.PredictedTime < 2*time.Minute {
		score += 2.0
	}

	// Lower priority for builds predicted to be very long (save resources for shorter builds)
	if predictions.PredictedTime > 10*time.Minute {
		score -= 1.0
	}

	// Factor in cache hit rate (higher hit rate = higher priority)
	score += predictions.CacheHitRate * 1.0

	return math.Max(0.0, math.Min(10.0, score)) // Clamp between 0-10
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
	status := map[string]any{
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

	bc.mutex.RLock()
	workerCount := len(bc.WorkerPool.Workers)
	buildQueueLength := len(bc.BuildQueue)
	bc.mutex.RUnlock()

	health := map[string]any{
		"status":      "healthy",
		"service":     "coordinator",
		"version":     "1.0.0",
		"timestamp":   time.Now().Format(time.RFC3339),
		"workers":     workerCount,
		"build_queue": buildQueueLength,
		"uptime":      time.Since(bc.startTime).String(),
	}

	json.NewEncoder(w).Encode(health)
}

// Main coordinator application entry point
func coordinatorMain() {
	// Register Prometheus metrics
	prometheus.MustRegister(
		activeBuilds,
		buildRequestsTotal,
		buildDuration,
		coordinatorHTTPRequestsTotal,
	)

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

// Main function removed - use coordinator/main.go instead
