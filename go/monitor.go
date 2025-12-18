package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Monitor tracks distributed build system performance
type Monitor struct {
	config         *MonitorConfig
	workers        map[string]*WorkerStats
	builds         map[string]*BuildStats
	startTime      time.Time
	httpServer     *http.Server
	metrics        *SystemMetrics
}

// MonitorConfig contains monitor configuration
type MonitorConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	DataDir         string `json:"data_dir"`
	CheckInterval   int    `json:"check_interval_seconds"`
	HistoryRetentionDays int `json:"history_retention_days"`
	AlertThresholds map[string]float64 `json:"alert_thresholds"`
}

// WorkerStats tracks individual worker performance
type WorkerStats struct {
	ID              string    `json:"id"`
	Host            string    `json:"host"`
	Status          string    `json:"status"`
	CPU             float64   `json:"cpu"`
	Memory          uint64    `json:"memory"`
	DiskUsage       uint64    `json:"disk_usage"`
	BuildCount      int       `json:"build_count"`
	LastBuildTime   time.Time `json:"last_build_time"`
	AverageBuildTime time.Duration `json:"average_build_time"`
	LastCheckin     time.Time `json:"last_checkin"`
	ResponseTime    time.Duration `json:"response_time"`
}

// BuildStats tracks build performance metrics
type BuildStats struct {
	ID              string        `json:"id"`
	WorkerID        string        `json:"worker_id"`
	ProjectPath     string        `json:"project_path"`
	TaskName        string        `json:"task_name"`
	Status          string        `json:"status"`
	StartTime       time.Time     `json:"start_time"`
	EndTime         time.Time     `json:"end_time"`
	Duration        time.Duration `json:"duration"`
	Success         bool          `json:"success"`
	CacheHitRate    float64       `json:"cache_hit_rate"`
	Artifacts       []string      `json:"artifacts"`
	ErrorMessage    string        `json:"error_message,omitempty"`
}

// SystemMetrics contains overall system metrics
type SystemMetrics struct {
	TotalBuilds     int64     `json:"total_builds"`
	SuccessfulBuilds int64    `json:"successful_builds"`
	FailedBuilds    int64     `json:"failed_builds"`
	AverageBuildTime time.Duration `json:"average_build_time"`
	TotalWorkers    int       `json:"total_workers"`
	ActiveWorkers   int       `json:"active_workers"`
	BusyWorkers     int       `json:"busy_workers"`
	IdleWorkers     int       `json:"idle_workers"`
	TotalBuildQueue int       `json:"total_build_queue"`
	SystemUptime    time.Duration `json:"system_uptime"`
	CacheHitRate    float64   `json:"cache_hit_rate"`
	NetworkLatency  time.Duration `json:"network_latency"`
}

// Alert represents a system alert
type Alert struct {
	ID          string                 `json:"id"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message"`
	Source      string                 `json:"source"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
}

// PerformanceReport contains system performance analysis
type PerformanceReport struct {
	Timestamp      time.Time             `json:"timestamp"`
	TimeRange      TimeRange             `json:"time_range"`
	Summary        SummaryStats          `json:"summary"`
	WorkerStats    map[string]*WorkerStats `json:"worker_stats"`
	BuildStats     map[string]*BuildStats  `json:"build_stats"`
	Alerts         []Alert               `json:"alerts"`
	Recommendations []string             `json:"recommendations"`
}

// TimeRange defines a time range for reports
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// SummaryStats contains summarized performance metrics
type SummaryStats struct {
	TotalBuilds        int64           `json:"total_builds"`
	SuccessRate        float64         `json:"success_rate"`
	AverageBuildTime   time.Duration   `json:"average_build_time"`
	ThroughputPerHour  float64         `json:"throughput_per_hour"`
	CacheEfficiency    float64         `json:"cache_efficiency"`
	WorkerUtilization  float64         `json:"worker_utilization"`
	ResourceUtilization ResourceStats  `json:"resource_utilization"`
}

// ResourceStats tracks resource utilization
type ResourceStats struct {
	AvgCPU          float64 `json:"avg_cpu"`
	MaxCPU          float64 `json:"max_cpu"`
	AvgMemory       uint64  `json:"avg_memory"`
	MaxMemory       uint64  `json:"max_memory"`
	AvgDiskIO       uint64  `json:"avg_disk_io"`
	MaxDiskIO       uint64  `json:"max_disk_io"`
	NetworkLatency  time.Duration `json:"network_latency"`
}

// NewMonitor creates a new monitoring system
func NewMonitor(config *MonitorConfig) *Monitor {
	if config.DataDir == "" {
		config.DataDir = "/tmp/gradle-monitor"
	}
	
	if config.CheckInterval == 0 {
		config.CheckInterval = 30 // 30 seconds
	}
	
	if config.HistoryRetentionDays == 0 {
		config.HistoryRetentionDays = 30
	}
	
	monitor := &Monitor{
		config:    config,
		workers:   make(map[string]*WorkerStats),
		builds:    make(map[string]*BuildStats),
		startTime: time.Now(),
		metrics:   &SystemMetrics{},
	}
	
	// Initialize metrics
	monitor.initializeMetrics()
	
	return monitor
}

// initializeMetrics sets up initial monitoring metrics
func (m *Monitor) initializeMetrics() {
	m.metrics = &SystemMetrics{
		TotalBuilds:      0,
		SuccessfulBuilds: 0,
		FailedBuilds:     0,
		TotalWorkers:     0,
		ActiveWorkers:    0,
		BusyWorkers:      0,
		IdleWorkers:      0,
		TotalBuildQueue:  0,
		SystemUptime:     time.Duration(0),
		CacheHitRate:     0.0,
		NetworkLatency:   time.Duration(0),
	}
}

// Start starts the monitoring system
func (m *Monitor) Start() error {
	// Create data directory
	if err := os.MkdirAll(m.config.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}
	
	// Start data collection
	go m.collectData()
	
	// Start HTTP server
	return m.StartHTTPServer()
}

// collectData periodically collects system metrics
func (m *Monitor) collectData() {
	ticker := time.NewTicker(time.Duration(m.config.CheckInterval) * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		m.updateWorkerStats()
		m.updateSystemMetrics()
		m.checkAlerts()
		m.saveMetrics()
	}
}

// updateWorkerStats updates worker performance statistics
func (m *Monitor) updateWorkerStats() {
	// This would typically query the coordinator for worker information
	// For now, we'll simulate updating existing workers
	for id, worker := range m.workers {
		// Check if worker is still responsive (simulate with timeout)
		responseTime := m.pingWorker(worker.Host)
		worker.ResponseTime = responseTime
		worker.LastCheckin = time.Now()
		
		// If no response for 5 minutes, mark as offline
		if time.Since(worker.LastCheckin) > 5*time.Minute {
			worker.Status = "offline"
		}
		
		m.workers[id] = worker
	}
}

// pingWorker measures response time to a worker
func (m *Monitor) pingWorker(host string) time.Duration {
	start := time.Now()
	
	// Simple HTTP ping to worker's health endpoint
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s:8082/health", host))
	if err != nil {
		return 5 * time.Second // Timeout
	}
	defer resp.Body.Close()
	
	return time.Since(start)
}

// updateSystemMetrics updates overall system metrics
func (m *Monitor) updateSystemMetrics() {
	m.metrics.SystemUptime = time.Since(m.startTime)
	m.metrics.TotalWorkers = len(m.workers)
	
	// Count worker statuses
	activeCount := 0
	busyCount := 0
	idleCount := 0
	
	for _, worker := range m.workers {
		if worker.Status == "online" || worker.Status == "busy" || worker.Status == "idle" {
			activeCount++
		}
		if worker.Status == "busy" {
			busyCount++
		}
		if worker.Status == "idle" {
			idleCount++
		}
	}
	
	m.metrics.ActiveWorkers = activeCount
	m.metrics.BusyWorkers = busyCount
	m.metrics.IdleWorkers = idleCount
	
	// Calculate success rate
	if m.metrics.TotalBuilds > 0 {
		successRate := float64(m.metrics.SuccessfulBuilds) / float64(m.metrics.TotalBuilds)
		m.metrics.CacheHitRate = successRate * 0.85 // Simulate cache hit rate
	}
}

// checkAlerts checks for alert conditions
func (m *Monitor) checkAlerts() {
	alerts := []Alert{}
	
	// Check worker availability
	if m.metrics.ActiveWorkers < len(m.workers)/2 {
		alert := Alert{
			ID:        fmt.Sprintf("worker-availability-%d", time.Now().Unix()),
			Severity:  "warning",
			Message:   "More than 50% of workers are offline",
			Source:    "monitor",
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"active_workers": m.metrics.ActiveWorkers,
				"total_workers":  m.metrics.TotalWorkers,
			},
			Resolved: false,
		}
		alerts = append(alerts, alert)
	}
	
	// Check build failure rate
	if m.metrics.TotalBuilds > 10 {
		failureRate := float64(m.metrics.FailedBuilds) / float64(m.metrics.TotalBuilds)
		if failureRate > 0.2 { // More than 20% failure rate
			alert := Alert{
				ID:        fmt.Sprintf("build-failure-rate-%d", time.Now().Unix()),
				Severity:  "critical",
				Message:   fmt.Sprintf("High build failure rate: %.2f%%", failureRate*100),
				Source:    "monitor",
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"failure_rate":     failureRate,
					"failed_builds":    m.metrics.FailedBuilds,
					"total_builds":     m.metrics.TotalBuilds,
				},
				Resolved: false,
			}
			alerts = append(alerts, alert)
		}
	}
	
	// Check worker response times
	for _, worker := range m.workers {
		if worker.ResponseTime > 5*time.Second {
			alert := Alert{
				ID:        fmt.Sprintf("worker-response-%s-%d", worker.ID, time.Now().Unix()),
				Severity:  "warning",
				Message:   fmt.Sprintf("Worker %s has high response time: %v", worker.ID, worker.ResponseTime),
				Source:    "monitor",
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"worker_id":     worker.ID,
					"response_time": worker.ResponseTime.String(),
				},
				Resolved: false,
			}
			alerts = append(alerts, alert)
		}
	}
	
	// Save alerts
	if len(alerts) > 0 {
		m.saveAlerts(alerts)
	}
}

// saveAlerts saves alerts to file
func (m *Monitor) saveAlerts(alerts []Alert) {
	alertsFile := filepath.Join(m.config.DataDir, "alerts.json")
	
	var existingAlerts []Alert
	if data, err := os.ReadFile(alertsFile); err == nil {
		json.Unmarshal(data, &existingAlerts)
	}
	
	existingAlerts = append(existingAlerts, alerts...)
	
	data, err := json.MarshalIndent(existingAlerts, "", "  ")
	if err != nil {
		log.Printf("Error marshaling alerts: %v", err)
		return
	}
	
	if err := os.WriteFile(alertsFile, data, 0644); err != nil {
		log.Printf("Error saving alerts: %v", err)
	}
}

// saveMetrics saves metrics to file
func (m *Monitor) saveMetrics() {
	metricsFile := filepath.Join(m.config.DataDir, "metrics.json")
	
	data, err := json.MarshalIndent(m.metrics, "", "  ")
	if err != nil {
		log.Printf("Error marshaling metrics: %v", err)
		return
	}
	
	if err := os.WriteFile(metricsFile, data, 0644); err != nil {
		log.Printf("Error saving metrics: %v", err)
	}
}

// RecordBuild records build completion statistics
func (m *Monitor) RecordBuild(buildID string, workerID string, projectPath string, taskName string, success bool, duration time.Duration, cacheHitRate float64, artifacts []string, errorMessage string) {
	build := &BuildStats{
		ID:            buildID,
		WorkerID:      workerID,
		ProjectPath:   projectPath,
		TaskName:      taskName,
		Status:        "completed",
		StartTime:     time.Now().Add(-duration),
		EndTime:       time.Now(),
		Duration:      duration,
		Success:       success,
		CacheHitRate:  cacheHitRate,
		Artifacts:     artifacts,
		ErrorMessage:  errorMessage,
	}
	
	m.builds[buildID] = build
	
	// Update system metrics
	m.metrics.TotalBuilds++
	if success {
		m.metrics.SuccessfulBuilds++
	} else {
		m.metrics.FailedBuilds++
	}
	
	// Update worker stats
	if worker, exists := m.workers[workerID]; exists {
		worker.BuildCount++
		worker.LastBuildTime = time.Now()
		
		// Update average build time
		totalTime := time.Duration(worker.BuildCount-1) * worker.AverageBuildTime + duration
		worker.AverageBuildTime = totalTime / time.Duration(worker.BuildCount)
		
		m.workers[workerID] = worker
	}
}

// RegisterWorker registers a new worker for monitoring
func (m *Monitor) RegisterWorker(id string, host string, status string) {
	worker := &WorkerStats{
		ID:              id,
		Host:            host,
		Status:          status,
		CPU:             0.0,
		Memory:          0,
		DiskUsage:       0,
		BuildCount:      0,
		LastBuildTime:   time.Time{},
		AverageBuildTime: time.Duration(0),
		LastCheckin:     time.Now(),
		ResponseTime:    time.Duration(0),
	}
	
	m.workers[id] = worker
}

// UnregisterWorker removes a worker from monitoring
func (m *Monitor) UnregisterWorker(id string) {
	delete(m.workers, id)
}

// GenerateReport generates a performance report for the given time range
func (m *Monitor) GenerateReport(start time.Time, end time.Time) (*PerformanceReport, error) {
	report := &PerformanceReport{
		Timestamp:   time.Now(),
		TimeRange:   TimeRange{Start: start, End: end},
		WorkerStats: make(map[string]*WorkerStats),
		BuildStats:  make(map[string]*BuildStats),
		Alerts:      []Alert{},
	}
	
	// Copy current worker stats
	for id, worker := range m.workers {
		workerCopy := *worker
		report.WorkerStats[id] = &workerCopy
	}
	
	// Copy builds in time range
	for id, build := range m.builds {
		if build.StartTime.After(start) && build.EndTime.Before(end) {
			buildCopy := *build
			report.BuildStats[id] = &buildCopy
		}
	}
	
	// Generate summary
	report.Summary = m.generateSummary(report.BuildStats)
	
	// Generate recommendations
	report.Recommendations = m.generateRecommendations(report.Summary, report.WorkerStats)
	
	return report, nil
}

// generateSummary creates summary statistics from build data
func (m *Monitor) generateSummary(builds map[string]*BuildStats) SummaryStats {
	summary := SummaryStats{}
	
	if len(builds) == 0 {
		return summary
	}
	
	var totalDuration time.Duration
	successCount := 0
	
	for _, build := range builds {
		summary.TotalBuilds++
		totalDuration += build.Duration
		
		if build.Success {
			successCount++
		}
	}
	
	summary.SuccessRate = float64(successCount) / float64(summary.TotalBuilds)
	summary.AverageBuildTime = totalDuration / time.Duration(summary.TotalBuilds)
	
	// Calculate hourly throughput
	timeRange := time.Hour // Default to 1 hour
	summary.ThroughputPerHour = float64(summary.TotalBuilds) / timeRange.Hours()
	
	// Simulate cache efficiency and worker utilization
	summary.CacheEfficiency = summary.SuccessRate * 0.85
	summary.WorkerUtilization = float64(m.metrics.BusyWorkers) / float64(m.metrics.TotalWorkers)
	
	return summary
}

// generateRecommendations creates optimization recommendations
func (m *Monitor) generateRecommendations(summary SummaryStats, workers map[string]*WorkerStats) []string {
	recommendations := []string{}
	
	// Low success rate recommendation
	if summary.SuccessRate < 0.8 {
		recommendations = append(recommendations, "Build success rate is below 80%. Check build configuration and dependencies.")
	}
	
	// High build time recommendation
	if summary.AverageBuildTime > 10*time.Minute {
		recommendations = append(recommendations, "Average build time exceeds 10 minutes. Consider optimizing build process or adding more workers.")
	}
	
	// Low cache efficiency recommendation
	if summary.CacheEfficiency < 0.7 {
		recommendations = append(recommendations, "Cache efficiency is below 70%. Review caching strategy and configuration.")
	}
	
	// Worker utilization recommendation
	if summary.WorkerUtilization > 0.9 {
		recommendations = append(recommendations, "Worker utilization is above 90%. Consider adding more workers to handle build load.")
	}
	
	if summary.WorkerUtilization < 0.3 {
		recommendations = append(recommendations, "Worker utilization is below 30%. Consider reducing worker count or increasing build frequency.")
	}
	
	return recommendations
}

// StartHTTPServer starts the monitor's HTTP API
func (m *Monitor) StartHTTPServer() error {
	mux := http.NewServeMux()
	
	// API endpoints
	mux.HandleFunc("/api/metrics", m.handleMetricsRequest)
	mux.HandleFunc("/api/workers", m.handleWorkersRequest)
	mux.HandleFunc("/api/builds", m.handleBuildsRequest)
	mux.HandleFunc("/api/report", m.handleReportRequest)
	mux.HandleFunc("/api/alerts", m.handleAlertsRequest)
	mux.HandleFunc("/api/register", m.handleRegisterRequest)
	mux.HandleFunc("/health", m.handleHealthCheck)
	
	m.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", m.config.Host, m.config.Port),
		Handler: mux,
	}
	
	log.Printf("Starting monitor on %s:%d", m.config.Host, m.config.Port)
	return m.httpServer.ListenAndServe()
}

// HTTP Handlers
func (m *Monitor) handleMetricsRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.metrics)
}

func (m *Monitor) handleWorkersRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.workers)
}

func (m *Monitor) handleBuildsRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m.builds)
}

func (m *Monitor) handleReportRequest(w http.ResponseWriter, r *http.Request) {
	// Parse time range parameters
	start := time.Now().Add(-24 * time.Hour) // Default: last 24 hours
	end := time.Now()
	
	if startStr := r.URL.Query().Get("start"); startStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startStr); err == nil {
			start = parsed
		}
	}
	
	if endStr := r.URL.Query().Get("end"); endStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endStr); err == nil {
			end = parsed
		}
	}
	
	report, err := m.GenerateReport(start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (m *Monitor) handleAlertsRequest(w http.ResponseWriter, r *http.Request) {
	alertsFile := filepath.Join(m.config.DataDir, "alerts.json")
	
	data, err := os.ReadFile(alertsFile)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]Alert{})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (m *Monitor) handleRegisterRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var worker struct {
		ID     string `json:"id"`
		Host   string `json:"host"`
		Status string `json:"status"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&worker); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	m.RegisterWorker(worker.ID, worker.Host, worker.Status)
	
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
}

func (m *Monitor) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"uptime": time.Since(m.startTime).String(),
		"metrics": m.metrics,
	})
}

func main() {
	// Load configuration
	configFile := "monitor_config.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}
	
	config, err := loadMonitorConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load monitor config: %v", err)
	}
	
	// Create and start monitor
	monitor := NewMonitor(config)
	
	log.Printf("Starting monitor on %s:%d", config.Host, config.Port)
	if err := monitor.Start(); err != nil {
		log.Fatalf("Failed to start monitor: %v", err)
	}
}

// loadMonitorConfig loads monitor configuration from file
func loadMonitorConfig(filename string) (*MonitorConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		// Return default config if file doesn't exist
		if os.IsNotExist(err) {
			return &MonitorConfig{
				Host:               "localhost",
				Port:               8084,
				DataDir:            "/tmp/gradle-monitor",
				CheckInterval:      30,
				HistoryRetentionDays: 30,
				AlertThresholds: map[string]float64{
					"cpu_usage":          80.0,
					"memory_usage":       90.0,
					"disk_usage":         85.0,
					"build_failure_rate": 20.0,
					"response_time":      5.0,
				},
			}, nil
		}
		return nil, err
	}
	
	var config MonitorConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	
	return &config, nil
}