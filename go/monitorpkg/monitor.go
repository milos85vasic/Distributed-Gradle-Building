package monitorpkg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
	
	"distributed-gradle-building/types"
)

// Monitor represents the system monitoring service
type Monitor struct {
	Config        types.MonitorConfig
	Metrics       SystemMetrics
	Alerts        []Alert
	Workers       map[string]WorkerMetrics
	Builds        map[string]BuildMetrics
	mutex         sync.RWMutex
	httpServer    *http.Server
	shutdown      chan struct{}
}

// SystemMetrics contains overall system metrics
type SystemMetrics struct {
	TotalBuilds       int64     `json:"total_builds"`
	SuccessfulBuilds  int64     `json:"successful_builds"`
	FailedBuilds      int64     `json:"failed_builds"`
	ActiveWorkers     int       `json:"active_workers"`
	QueueLength       int       `json:"queue_length"`
	AverageBuildTime  float64   `json:"average_build_time"`
	CacheHitRate      float64   `json:"cache_hit_rate"`
	LastUpdate        time.Time `json:"last_update"`
	ResourceUsage     ResourceMetrics `json:"resource_usage"`
}

// WorkerMetrics contains metrics for a specific worker
type WorkerMetrics struct {
	ID                string    `json:"id"`
	Status            string    `json:"status"`
	ActiveBuilds      int       `json:"active_builds"`
	CompletedBuilds   int64     `json:"completed_builds"`
	FailedBuilds      int64     `json:"failed_builds"`
	LastSeen          time.Time `json:"last_seen"`
	AverageBuildTime  float64   `json:"average_build_time"`
	CPUUsage          float64   `json:"cpu_usage"`
	MemoryUsage       int64     `json:"memory_usage"`
}

// BuildMetrics contains metrics for a specific build
type BuildMetrics struct {
	BuildID       string        `json:"build_id"`
	WorkerID      string        `json:"worker_id"`
	Status        string        `json:"status"`
	StartTime     time.Time     `json:"start_time"`
	EndTime       time.Time     `json:"end_time"`
	Duration      time.Duration `json:"duration"`
	CacheHit      bool          `json:"cache_hit"`
	Artifacts     []string      `json:"artifacts"`
	ResourceUsage ResourceMetrics `json:"resource_usage"`
}

// ResourceMetrics contains resource usage information
type ResourceMetrics struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryUsage   int64   `json:"memory_usage"`
	DiskIO        int64   `json:"disk_io"`
	NetworkIO     int64   `json:"network_io"`
	MaxMemoryUsed int64   `json:"max_memory_used"`
}

// Alert represents a system alert
type Alert struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message"`
	Timestamp   time.Time              `json:"timestamp"`
	Resolved    bool                   `json:"resolved"`
	Data        map[string]interface{} `json:"data"`
}

// NewMonitor creates a new monitoring service
func NewMonitor(config types.MonitorConfig) *Monitor {
	return &Monitor{
		Config:  config,
		Workers: make(map[string]WorkerMetrics),
		Builds:  make(map[string]BuildMetrics),
		Alerts:  make([]Alert, 0),
		shutdown: make(chan struct{}),
		Metrics: SystemMetrics{
			LastUpdate: time.Now(),
		},
	}
}

// RecordBuildStart records the start of a build
func (m *Monitor) RecordBuildStart(buildID, workerID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.Builds[buildID] = BuildMetrics{
		BuildID:   buildID,
		WorkerID:  workerID,
		Status:    "running",
		StartTime: time.Now(),
	}
	
	m.Metrics.TotalBuilds++
	m.Metrics.QueueLength--
	m.Metrics.LastUpdate = time.Now()
}

// RecordBuildComplete records the completion of a build
func (m *Monitor) RecordBuildComplete(buildID string, success bool, duration time.Duration, cacheHit bool, artifacts []string, resourceUsage ResourceMetrics) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	buildMetrics, exists := m.Builds[buildID]
	if !exists {
		return
	}
	
	buildMetrics.Status = "completed"
	buildMetrics.EndTime = time.Now()
	buildMetrics.Duration = duration
	buildMetrics.CacheHit = cacheHit
	buildMetrics.Artifacts = artifacts
	buildMetrics.ResourceUsage = resourceUsage
	
	m.Builds[buildID] = buildMetrics
	
	if success {
		m.Metrics.SuccessfulBuilds++
	} else {
		m.Metrics.FailedBuilds++
	}
	
	// Update average build time
	totalBuilds := m.Metrics.SuccessfulBuilds + m.Metrics.FailedBuilds
	if totalBuilds > 0 {
		m.Metrics.AverageBuildTime = ((m.Metrics.AverageBuildTime * float64(totalBuilds-1)) + duration.Seconds()) / float64(totalBuilds)
	}
	
	m.Metrics.LastUpdate = time.Now()
}

// UpdateWorkerMetrics updates metrics for a specific worker
func (m *Monitor) UpdateWorkerMetrics(workerID, status string, activeBuilds int, completedBuilds, failedBuilds int64, lastSeen time.Time, cpuUsage float64, memoryUsage int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	workerMetrics := WorkerMetrics{
		ID:               workerID,
		Status:           status,
		ActiveBuilds:     activeBuilds,
		CompletedBuilds:  completedBuilds,
		FailedBuilds:     failedBuilds,
		LastSeen:         lastSeen,
		CPUUsage:         cpuUsage,
		MemoryUsage:      memoryUsage,
	}
	
	m.Workers[workerID] = workerMetrics
	
	// Update active workers count
	activeCount := 0
	for _, worker := range m.Workers {
		if worker.Status == "active" {
			activeCount++
		}
	}
	
	m.Metrics.ActiveWorkers = activeCount
	m.Metrics.LastUpdate = time.Now()
}

// TriggerAlert triggers a new alert
func (m *Monitor) TriggerAlert(alertType, severity, message string, data map[string]interface{}) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	alert := Alert{
		ID:        fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		Type:      alertType,
		Severity:  severity,
		Message:   message,
		Timestamp: time.Now(),
		Resolved:  false,
		Data:      data,
	}
	
	m.Alerts = append(m.Alerts, alert)
	log.Printf("ALERT [%s]: %s", severity, message)
}

// GetSystemMetrics returns current system metrics
func (m *Monitor) GetSystemMetrics() SystemMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return m.Metrics
}

// GetWorkerMetrics returns metrics for all workers
func (m *Monitor) GetWorkerMetrics() map[string]WorkerMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Return a copy to avoid race conditions
	result := make(map[string]WorkerMetrics)
	for id, metrics := range m.Workers {
		result[id] = metrics
	}
	
	return result
}

// GetBuildMetrics returns metrics for all builds
func (m *Monitor) GetBuildMetrics() map[string]BuildMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Return a copy to avoid race conditions
	result := make(map[string]BuildMetrics)
	for id, metrics := range m.Builds {
		result[id] = metrics
	}
	
	return result
}

// GetAlerts returns all alerts
func (m *Monitor) GetAlerts() []Alert {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Return a copy to avoid race conditions
	result := make([]Alert, len(m.Alerts))
	copy(result, m.Alerts)
	
	return result
}

// StartServer starts the HTTP server for monitoring API
func (m *Monitor) StartServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/metrics", m.handleMetrics)
	mux.HandleFunc("/api/workers", m.handleWorkers)
	mux.HandleFunc("/api/builds", m.handleBuilds)
	mux.HandleFunc("/api/alerts", m.handleAlerts)
	mux.HandleFunc("/api/health", m.handleHealth)
	
	m.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", m.Config.Port),
		Handler: mux,
	}
	
	go func() {
		log.Printf("Monitor HTTP server listening on port %d", m.Config.Port)
		if err := m.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()
	
	// Start metrics collection routine
	go m.startMetricsCollection()
	
	return nil
}

// Shutdown gracefully shuts down the monitor
func (m *Monitor) Shutdown() error {
	close(m.shutdown)
	
	if m.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		m.httpServer.Shutdown(ctx)
	}
	
	return nil
}

// HandleMetrics handles system metrics requests (exported for testing)
func (m *Monitor) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	m.handleMetrics(w, r)
}

// HandleWorkers handles worker metrics requests (exported for testing)
func (m *Monitor) HandleWorkers(w http.ResponseWriter, r *http.Request) {
	m.handleWorkers(w, r)
}

// HandleBuilds handles build metrics requests (exported for testing)
func (m *Monitor) HandleBuilds(w http.ResponseWriter, r *http.Request) {
	m.handleBuilds(w, r)
}

// HandleAlerts handles alerts requests (exported for testing)
func (m *Monitor) HandleAlerts(w http.ResponseWriter, r *http.Request) {
	m.handleAlerts(w, r)
}

// HandleHealth handles health check requests (exported for testing)
func (m *Monitor) HandleHealth(w http.ResponseWriter, r *http.Request) {
	m.handleHealth(w, r)
}

// handleMetrics handles system metrics requests
func (m *Monitor) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := m.GetSystemMetrics()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// handleWorkers handles worker metrics requests (exported for testing)
func (m *Monitor) handleWorkers(w http.ResponseWriter, r *http.Request) {
	workers := m.GetWorkerMetrics()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workers)
}

// handleBuilds handles build metrics requests (exported for testing)
func (m *Monitor) handleBuilds(w http.ResponseWriter, r *http.Request) {
	builds := m.GetBuildMetrics()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(builds)
}

// handleAlerts handles alerts requests (exported for testing)
func (m *Monitor) handleAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := m.GetAlerts()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

// handleHealth handles health check requests (exported for testing)
func (m *Monitor) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// startMetricsCollection starts the metrics collection routine
func (m *Monitor) startMetricsCollection() {
	ticker := time.NewTicker(m.Config.MetricsInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Check for worker timeouts and trigger alerts if needed
			m.checkWorkerTimeouts()
			
			// Check alert thresholds and trigger alerts if needed
			m.checkAlertThresholds()
			
		case <-m.shutdown:
			return
		}
	}
}

// checkWorkerTimeouts checks for worker timeouts and triggers alerts
func (m *Monitor) checkWorkerTimeouts() {
	now := time.Now()
	timeout := 5 * time.Minute // Configurable timeout
	
	for workerID, metrics := range m.Workers {
		if now.Sub(metrics.LastSeen) > timeout {
			// Check if we already have an alert for this worker
			hasAlert := false
			for _, alert := range m.Alerts {
				if alert.Type == "worker_timeout" && alert.Data["worker_id"] == workerID && !alert.Resolved {
					hasAlert = true
					break
				}
			}
			
			if !hasAlert {
				m.TriggerAlert("worker_timeout", "warning", 
					fmt.Sprintf("Worker %s has not been seen for %v", workerID, now.Sub(metrics.LastSeen)),
					map[string]interface{}{
						"worker_id": workerID,
						"last_seen": metrics.LastSeen,
					})
			}
		}
	}
}

// CheckAlertThresholds checks metrics against alert thresholds (exported for testing)
func (m *Monitor) CheckAlertThresholds() {
	m.checkAlertThresholds()
}

// checkAlertThresholds checks metrics against alert thresholds
func (m *Monitor) checkAlertThresholds() {
	// Check build failure rate
	if m.Metrics.TotalBuilds > 0 {
		failureRate := float64(m.Metrics.FailedBuilds) / float64(m.Metrics.TotalBuilds)
		if threshold, exists := m.Config.AlertThresholds["build_failure_rate"]; exists && failureRate > threshold {
			m.TriggerAlert("high_failure_rate", "critical",
				fmt.Sprintf("Build failure rate %.2f%% exceeds threshold %.2f%%", failureRate*100, threshold*100),
				map[string]interface{}{
					"failure_rate": failureRate,
					"threshold":    threshold,
				})
		}
	}
	
	// Check queue length
	if threshold, exists := m.Config.AlertThresholds["queue_length"]; exists && int64(m.Metrics.QueueLength) > int64(threshold) {
		m.TriggerAlert("queue_length", "warning",
			fmt.Sprintf("Build queue length %d exceeds threshold %d", m.Metrics.QueueLength, int(threshold)),
			map[string]interface{}{
				"queue_length": m.Metrics.QueueLength,
				"threshold":    threshold,
			})
	}
}