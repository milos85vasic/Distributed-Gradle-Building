package monitorpkg

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"distributed-gradle-building/types"
)

func TestNewMonitor(t *testing.T) {
	config := types.MonitorConfig{
		Port:            8080,
		MetricsInterval: 30 * time.Second,
		AlertThresholds: map[string]float64{
			"build_failure_rate": 0.2,
			"queue_length":       10,
		},
	}

	monitor := NewMonitor(config)
	if monitor == nil {
		t.Fatal("Expected monitor to be created")
	}
	if monitor.Config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", monitor.Config.Port)
	}
	if len(monitor.Workers) != 0 {
		t.Errorf("Expected 0 workers, got %d", len(monitor.Workers))
	}
	if len(monitor.Builds) != 0 {
		t.Errorf("Expected 0 builds, got %d", len(monitor.Builds))
	}
	if len(monitor.Alerts) != 0 {
		t.Errorf("Expected 0 alerts, got %d", len(monitor.Alerts))
	}
	if monitor.Metrics.TotalBuilds != 0 {
		t.Errorf("Expected 0 total builds, got %d", monitor.Metrics.TotalBuilds)
	}
}

func TestRecordBuildStart(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	monitor.RecordBuildStart("build-123", "worker-1")

	if len(monitor.Builds) != 1 {
		t.Errorf("Expected 1 build, got %d", len(monitor.Builds))
	}

	buildMetrics, exists := monitor.Builds["build-123"]
	if !exists {
		t.Fatal("Build metrics not found")
	}
	if buildMetrics.BuildID != "build-123" {
		t.Errorf("Expected BuildID build-123, got %s", buildMetrics.BuildID)
	}
	if buildMetrics.WorkerID != "worker-1" {
		t.Errorf("Expected WorkerID worker-1, got %s", buildMetrics.WorkerID)
	}
	if buildMetrics.Status != "running" {
		t.Errorf("Expected Status running, got %s", buildMetrics.Status)
	}
	if buildMetrics.StartTime.IsZero() {
		t.Error("Expected non-zero StartTime")
	}

	if monitor.Metrics.TotalBuilds != 1 {
		t.Errorf("Expected TotalBuilds 1, got %d", monitor.Metrics.TotalBuilds)
	}
}

func TestRecordBuildComplete(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	// First record build start
	monitor.RecordBuildStart("build-123", "worker-1")

	// Then record completion
	artifacts := []string{"app.jar", "test-results.xml"}
	resourceUsage := ResourceMetrics{
		CPUPercent:  0.75,
		MemoryUsage: 1024 * 1024 * 500, // 500MB
		DiskIO:      1024 * 1024 * 100, // 100MB
		NetworkIO:   1024 * 1024 * 50,  // 50MB
	}

	monitor.RecordBuildComplete("build-123", true, 2*time.Minute, true, artifacts, resourceUsage)

	buildMetrics, exists := monitor.Builds["build-123"]
	if !exists {
		t.Fatal("Build metrics not found")
	}
	if buildMetrics.Status != "completed" {
		t.Errorf("Expected Status completed, got %s", buildMetrics.Status)
	}
	if buildMetrics.EndTime.IsZero() {
		t.Error("Expected non-zero EndTime")
	}
	if buildMetrics.Duration != 2*time.Minute {
		t.Errorf("Expected Duration 2m, got %v", buildMetrics.Duration)
	}
	if !buildMetrics.CacheHit {
		t.Error("Expected CacheHit true")
	}
	if len(buildMetrics.Artifacts) != 2 {
		t.Errorf("Expected 2 artifacts, got %d", len(buildMetrics.Artifacts))
	}
	if buildMetrics.ResourceUsage.CPUPercent != 0.75 {
		t.Errorf("Expected CPUPercent 0.75, got %f", buildMetrics.ResourceUsage.CPUPercent)
	}

	if monitor.Metrics.SuccessfulBuilds != 1 {
		t.Errorf("Expected SuccessfulBuilds 1, got %d", monitor.Metrics.SuccessfulBuilds)
	}
	if monitor.Metrics.FailedBuilds != 0 {
		t.Errorf("Expected FailedBuilds 0, got %d", monitor.Metrics.FailedBuilds)
	}
	if monitor.Metrics.AverageBuildTime != 120.0 {
		t.Errorf("Expected AverageBuildTime 120.0, got %f", monitor.Metrics.AverageBuildTime)
	}
}

func TestRecordBuildComplete_Failure(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	monitor.RecordBuildStart("build-123", "worker-1")
	monitor.RecordBuildComplete("build-123", false, 30*time.Second, false, nil, ResourceMetrics{})

	if monitor.Metrics.SuccessfulBuilds != 0 {
		t.Errorf("Expected SuccessfulBuilds 0, got %d", monitor.Metrics.SuccessfulBuilds)
	}
	if monitor.Metrics.FailedBuilds != 1 {
		t.Errorf("Expected FailedBuilds 1, got %d", monitor.Metrics.FailedBuilds)
	}
	if monitor.Metrics.AverageBuildTime != 30.0 {
		t.Errorf("Expected AverageBuildTime 30.0, got %f", monitor.Metrics.AverageBuildTime)
	}
}

func TestRecordBuildComplete_NonexistentBuild(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	// Should not panic when recording completion for non-existent build
	monitor.RecordBuildComplete("nonexistent", true, 0, false, nil, ResourceMetrics{})

	if len(monitor.Builds) != 0 {
		t.Errorf("Expected 0 builds, got %d", len(monitor.Builds))
	}
}

func TestUpdateWorkerMetrics(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	now := time.Now()
	monitor.UpdateWorkerMetrics("worker-1", "active", 2, 10, 1, now, 0.75, 1024*1024*500)

	if len(monitor.Workers) != 1 {
		t.Errorf("Expected 1 worker, got %d", len(monitor.Workers))
	}

	workerMetrics, exists := monitor.Workers["worker-1"]
	if !exists {
		t.Fatal("Worker metrics not found")
	}
	if workerMetrics.ID != "worker-1" {
		t.Errorf("Expected ID worker-1, got %s", workerMetrics.ID)
	}
	if workerMetrics.Status != "active" {
		t.Errorf("Expected Status active, got %s", workerMetrics.Status)
	}
	if workerMetrics.ActiveBuilds != 2 {
		t.Errorf("Expected ActiveBuilds 2, got %d", workerMetrics.ActiveBuilds)
	}
	if workerMetrics.CompletedBuilds != 10 {
		t.Errorf("Expected CompletedBuilds 10, got %d", workerMetrics.CompletedBuilds)
	}
	if workerMetrics.FailedBuilds != 1 {
		t.Errorf("Expected FailedBuilds 1, got %d", workerMetrics.FailedBuilds)
	}
	if workerMetrics.CPUUsage != 0.75 {
		t.Errorf("Expected CPUUsage 0.75, got %f", workerMetrics.CPUUsage)
	}
	if workerMetrics.MemoryUsage != 1024*1024*500 {
		t.Errorf("Expected MemoryUsage 524288000, got %d", workerMetrics.MemoryUsage)
	}

	if monitor.Metrics.ActiveWorkers != 1 {
		t.Errorf("Expected ActiveWorkers 1, got %d", monitor.Metrics.ActiveWorkers)
	}
}

func TestUpdateWorkerMetrics_UpdateExisting(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	now := time.Now()
	monitor.UpdateWorkerMetrics("worker-1", "active", 1, 5, 0, now, 0.5, 1024*1024*250)
	monitor.UpdateWorkerMetrics("worker-1", "idle", 0, 6, 1, now.Add(time.Minute), 0.1, 1024*1024*100)

	if len(monitor.Workers) != 1 {
		t.Errorf("Expected 1 worker, got %d", len(monitor.Workers))
	}

	workerMetrics := monitor.Workers["worker-1"]
	if workerMetrics.Status != "idle" {
		t.Errorf("Expected Status idle, got %s", workerMetrics.Status)
	}
	if workerMetrics.ActiveBuilds != 0 {
		t.Errorf("Expected ActiveBuilds 0, got %d", workerMetrics.ActiveBuilds)
	}
	if workerMetrics.CompletedBuilds != 6 {
		t.Errorf("Expected CompletedBuilds 6, got %d", workerMetrics.CompletedBuilds)
	}
	if workerMetrics.FailedBuilds != 1 {
		t.Errorf("Expected FailedBuilds 1, got %d", workerMetrics.FailedBuilds)
	}
}

func TestTriggerAlert(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	data := map[string]interface{}{
		"build_id": "build-123",
		"reason":   "timeout",
	}
	monitor.TriggerAlert("build_timeout", "critical", "Build build-123 timed out", data)

	if len(monitor.Alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(monitor.Alerts))
	}

	alert := monitor.Alerts[0]
	if alert.Type != "build_timeout" {
		t.Errorf("Expected Type build_timeout, got %s", alert.Type)
	}
	if alert.Severity != "critical" {
		t.Errorf("Expected Severity critical, got %s", alert.Severity)
	}
	if alert.Message != "Build build-123 timed out" {
		t.Errorf("Expected Message 'Build build-123 timed out', got %s", alert.Message)
	}
	if alert.Resolved {
		t.Error("Expected Resolved false")
	}
	if alert.Timestamp.IsZero() {
		t.Error("Expected non-zero Timestamp")
	}
	if alert.Data["build_id"] != "build-123" {
		t.Errorf("Expected Data.build_id build-123, got %v", alert.Data["build_id"])
	}
}

func TestGetSystemMetrics(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	// Update some metrics
	monitor.RecordBuildStart("build-1", "worker-1")
	monitor.RecordBuildComplete("build-1", true, 2*time.Minute, true, nil, ResourceMetrics{})
	monitor.UpdateWorkerMetrics("worker-1", "active", 1, 5, 0, time.Now(), 0.5, 1024*1024*250)

	metrics := monitor.GetSystemMetrics()
	if metrics.TotalBuilds != 1 {
		t.Errorf("Expected TotalBuilds 1, got %d", metrics.TotalBuilds)
	}
	if metrics.SuccessfulBuilds != 1 {
		t.Errorf("Expected SuccessfulBuilds 1, got %d", metrics.SuccessfulBuilds)
	}
	if metrics.ActiveWorkers != 1 {
		t.Errorf("Expected ActiveWorkers 1, got %d", metrics.ActiveWorkers)
	}
	if metrics.AverageBuildTime != 120.0 {
		t.Errorf("Expected AverageBuildTime 120.0, got %f", metrics.AverageBuildTime)
	}
	if metrics.LastUpdate.IsZero() {
		t.Error("Expected non-zero LastUpdate")
	}
}

func TestGetWorkerMetrics(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	monitor.UpdateWorkerMetrics("worker-1", "active", 1, 5, 0, time.Now(), 0.5, 1024*1024*250)
	monitor.UpdateWorkerMetrics("worker-2", "idle", 0, 3, 1, time.Now(), 0.1, 1024*1024*100)

	workers := monitor.GetWorkerMetrics()
	if len(workers) != 2 {
		t.Errorf("Expected 2 workers, got %d", len(workers))
	}

	if worker1, ok := workers["worker-1"]; !ok {
		t.Error("Worker-1 not found")
	} else if worker1.Status != "active" {
		t.Errorf("Expected worker-1 status active, got %s", worker1.Status)
	}

	if worker2, ok := workers["worker-2"]; !ok {
		t.Error("Worker-2 not found")
	} else if worker2.Status != "idle" {
		t.Errorf("Expected worker-2 status idle, got %s", worker2.Status)
	}
}

func TestGetBuildMetrics(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	monitor.RecordBuildStart("build-1", "worker-1")
	monitor.RecordBuildStart("build-2", "worker-2")
	monitor.RecordBuildComplete("build-1", true, 2*time.Minute, true, nil, ResourceMetrics{})

	builds := monitor.GetBuildMetrics()
	if len(builds) != 2 {
		t.Errorf("Expected 2 builds, got %d", len(builds))
	}

	if build1, ok := builds["build-1"]; !ok {
		t.Error("Build-1 not found")
	} else if build1.Status != "completed" {
		t.Errorf("Expected build-1 status completed, got %s", build1.Status)
	}

	if build2, ok := builds["build-2"]; !ok {
		t.Error("Build-2 not found")
	} else if build2.Status != "running" {
		t.Errorf("Expected build-2 status running, got %s", build2.Status)
	}
}

func TestGetAlerts(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	monitor.TriggerAlert("alert1", "warning", "Test alert 1", nil)
	monitor.TriggerAlert("alert2", "critical", "Test alert 2", nil)

	alerts := monitor.GetAlerts()
	if len(alerts) != 2 {
		t.Errorf("Expected 2 alerts, got %d", len(alerts))
	}

	if alerts[0].Type != "alert1" {
		t.Errorf("Expected first alert type alert1, got %s", alerts[0].Type)
	}
	if alerts[1].Type != "alert2" {
		t.Errorf("Expected second alert type alert2, got %s", alerts[1].Type)
	}
}

func TestHandleMetrics(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	req := httptest.NewRequest("GET", "/api/metrics", nil)
	w := httptest.NewRecorder()

	monitor.HandleMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var metrics SystemMetrics
	if err := json.NewDecoder(w.Body).Decode(&metrics); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if metrics.TotalBuilds != 0 {
		t.Errorf("Expected TotalBuilds 0, got %d", metrics.TotalBuilds)
	}
}

func TestHandleWorkers(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	monitor.UpdateWorkerMetrics("worker-1", "active", 1, 5, 0, time.Now(), 0.5, 1024*1024*250)

	req := httptest.NewRequest("GET", "/api/workers", nil)
	w := httptest.NewRecorder()

	monitor.HandleWorkers(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var workers map[string]WorkerMetrics
	if err := json.NewDecoder(w.Body).Decode(&workers); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if len(workers) != 1 {
		t.Errorf("Expected 1 worker, got %d", len(workers))
	}
}

func TestHandleBuilds(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	monitor.RecordBuildStart("build-1", "worker-1")

	req := httptest.NewRequest("GET", "/api/builds", nil)
	w := httptest.NewRecorder()

	monitor.HandleBuilds(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var builds map[string]BuildMetrics
	if err := json.NewDecoder(w.Body).Decode(&builds); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if len(builds) != 1 {
		t.Errorf("Expected 1 build, got %d", len(builds))
	}
}

func TestHandleAlerts(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	monitor.TriggerAlert("test", "warning", "Test alert", nil)

	req := httptest.NewRequest("GET", "/api/alerts", nil)
	w := httptest.NewRecorder()

	monitor.HandleAlerts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var alerts []Alert
	if err := json.NewDecoder(w.Body).Decode(&alerts); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if len(alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(alerts))
	}
}

func TestHandleHealth(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	monitor.HandleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var health map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if status, ok := health["status"].(string); !ok || status != "healthy" {
		t.Errorf("Expected status healthy, got %v", health["status"])
	}
	if version, ok := health["version"].(string); !ok || version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %v", health["version"])
	}
}

func TestDetectBuildTimeAnomaly(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	// Test with insufficient data
	metrics := SystemMetrics{TotalBuilds: 5, AverageBuildTime: 1000.0}
	if monitor.detectBuildTimeAnomaly(metrics) {
		t.Error("Expected no anomaly with insufficient data")
	}

	// Test with normal build time
	metrics = SystemMetrics{TotalBuilds: 20, AverageBuildTime: 300.0}
	if monitor.detectBuildTimeAnomaly(metrics) {
		t.Error("Expected no anomaly with normal build time")
	}

	// Test with anomalous build time
	metrics = SystemMetrics{TotalBuilds: 20, AverageBuildTime: 500.0} // > 450 (300 * 1.5)
	if !monitor.detectBuildTimeAnomaly(metrics) {
		t.Error("Expected anomaly with high build time")
	}
}

func TestDetectWorkerPerformanceAnomaly(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	// Test normal worker
	worker := WorkerMetrics{CPUUsage: 0.5, MemoryUsage: 1024 * 1024 * 500}
	if monitor.detectWorkerPerformanceAnomaly(worker) {
		t.Error("Expected no anomaly with normal worker")
	}

	// Test high CPU usage
	worker = WorkerMetrics{CPUUsage: 0.95, MemoryUsage: 1024 * 1024 * 500}
	if !monitor.detectWorkerPerformanceAnomaly(worker) {
		t.Error("Expected anomaly with high CPU usage")
	}

	// Test high memory usage
	worker = WorkerMetrics{CPUUsage: 0.5, MemoryUsage: 1024 * 1024 * 2000} // > 1GB
	if !monitor.detectWorkerPerformanceAnomaly(worker) {
		t.Error("Expected anomaly with high memory usage")
	}
}

func TestDetectQueueLengthAnomaly(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	// Test normal queue length
	if monitor.detectQueueLengthAnomaly(10) {
		t.Error("Expected no anomaly with normal queue length")
	}

	// Test anomalous queue length
	if !monitor.detectQueueLengthAnomaly(25) {
		t.Error("Expected anomaly with high queue length")
	}
}

func TestDetectCacheHitRateAnomaly(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	// Test normal cache hit rate
	if monitor.detectCacheHitRateAnomaly(0.75) {
		t.Error("Expected no anomaly with normal cache hit rate")
	}

	// Test anomalous cache hit rate
	if !monitor.detectCacheHitRateAnomaly(0.3) {
		t.Error("Expected anomaly with low cache hit rate")
	}
}

func TestDetectResourceUsageAnomaly(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	// Test normal resource usage
	resources := ResourceMetrics{CPUPercent: 0.5, MemoryUsage: 1024 * 1024 * 1000}
	if monitor.detectResourceUsageAnomaly(resources) {
		t.Error("Expected no anomaly with normal resource usage")
	}

	// Test high CPU usage
	resources = ResourceMetrics{CPUPercent: 0.96, MemoryUsage: 1024 * 1024 * 1000}
	if !monitor.detectResourceUsageAnomaly(resources) {
		t.Error("Expected anomaly with high CPU usage")
	}

	// Test high memory usage
	resources = ResourceMetrics{CPUPercent: 0.5, MemoryUsage: 1024 * 1024 * 3000} // > 2GB
	if !monitor.detectResourceUsageAnomaly(resources) {
		t.Error("Expected anomaly with high memory usage")
	}
}

func TestGetAnomalyStatistics(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	// Trigger some alerts
	monitor.TriggerAlert("build_time_anomaly", "warning", "Test anomaly 1", nil)
	monitor.TriggerAlert("worker_performance_anomaly", "warning", "Test anomaly 2", nil)
	monitor.TriggerAlert("build_time_anomaly", "warning", "Test anomaly 3", nil)
	monitor.TriggerAlert("other_alert", "info", "Non-anomaly alert", nil)

	stats := monitor.GetAnomalyStatistics()
	if total, ok := stats["total_anomalies"].(int); !ok || total != 4 {
		t.Errorf("Expected total_anomalies 4, got %v", stats["total_anomalies"])
	}

	anomaliesByType, ok := stats["anomalies_by_type"].(map[string]int)
	if !ok {
		t.Fatal("Expected anomalies_by_type to be map[string]int")
	}
	if anomaliesByType["build_time_anomaly"] != 2 {
		t.Errorf("Expected 2 build_time_anomaly alerts, got %d", anomaliesByType["build_time_anomaly"])
	}
	if anomaliesByType["worker_performance_anomaly"] != 1 {
		t.Errorf("Expected 1 worker_performance_anomaly alert, got %d", anomaliesByType["worker_performance_anomaly"])
	}
}

func TestCheckAlertThresholds(t *testing.T) {
	config := types.MonitorConfig{
		AlertThresholds: map[string]float64{
			"build_failure_rate": 0.2,
			"queue_length":       10,
		},
	}
	monitor := NewMonitor(config)

	// Set up metrics that should trigger alerts
	monitor.Metrics.TotalBuilds = 100
	monitor.Metrics.FailedBuilds = 30 // 30% failure rate > 20% threshold
	monitor.Metrics.QueueLength = 15  // > 10 threshold

	monitor.CheckAlertThresholds()

	if len(monitor.Alerts) != 2 {
		t.Errorf("Expected 2 alerts, got %d", len(monitor.Alerts))
	}

	// Check for high failure rate alert
	foundFailureAlert := false
	foundQueueAlert := false
	for _, alert := range monitor.Alerts {
		if alert.Type == "high_failure_rate" {
			foundFailureAlert = true
			if alert.Severity != "critical" {
				t.Errorf("Expected severity critical for failure rate alert, got %s", alert.Severity)
			}
		}
		if alert.Type == "queue_length" {
			foundQueueAlert = true
			if alert.Severity != "warning" {
				t.Errorf("Expected severity warning for queue length alert, got %s", alert.Severity)
			}
		}
	}
	if !foundFailureAlert {
		t.Error("Expected high_failure_rate alert")
	}
	if !foundQueueAlert {
		t.Error("Expected queue_length alert")
	}
}

func TestCheckAlertThresholds_NoAlerts(t *testing.T) {
	config := types.MonitorConfig{
		AlertThresholds: map[string]float64{
			"build_failure_rate": 0.2,
			"queue_length":       10,
		},
	}
	monitor := NewMonitor(config)

	// Set up metrics that should NOT trigger alerts
	monitor.Metrics.TotalBuilds = 100
	monitor.Metrics.FailedBuilds = 15 // 15% failure rate < 20% threshold
	monitor.Metrics.QueueLength = 5   // < 10 threshold

	monitor.CheckAlertThresholds()

	if len(monitor.Alerts) != 0 {
		t.Errorf("Expected 0 alerts, got %d", len(monitor.Alerts))
	}
}

func TestShutdown(t *testing.T) {
	config := types.MonitorConfig{}
	monitor := NewMonitor(config)

	// Shutdown should not panic even without server started
	err := monitor.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}
