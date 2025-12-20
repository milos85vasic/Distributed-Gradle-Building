package monitor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"distributed-gradle-building/monitorpkg"
	"distributed-gradle-building/types"
)

func TestMonitor_Creation(t *testing.T) {
	config := types.MonitorConfig{
		Port:            8084,
		MetricsInterval: time.Second,
		AlertThresholds: map[string]float64{
			"build_failure_rate": 0.1,
			"queue_length":       10,
		},
	}

	monitor := monitorpkg.NewMonitor(config)

	if monitor == nil {
		t.Fatal("Failed to create monitor")
	}

	if monitor.Config.Port != 8084 {
		t.Errorf("Expected port 8084, got %d", monitor.Config.Port)
	}
}

func TestMonitor_RecordBuildStart(t *testing.T) {
	config := types.MonitorConfig{
		Port:            8084,
		MetricsInterval: time.Second,
		AlertThresholds: map[string]float64{},
	}

	monitor := monitorpkg.NewMonitor(config)

	// Record build start
	monitor.RecordBuildStart("build-1", "worker-1")

	// Check system metrics
	metrics := monitor.GetSystemMetrics()
	if metrics.TotalBuilds != 1 {
		t.Errorf("Expected 1 total build, got %d", metrics.TotalBuilds)
	}

	// Check build metrics
	builds := monitor.GetBuildMetrics()
	build, exists := builds["build-1"]
	if !exists {
		t.Fatal("Expected build-1 to exist in metrics")
	}

	if build.WorkerID != "worker-1" {
		t.Errorf("Expected worker ID worker-1, got %s", build.WorkerID)
	}

	if build.Status != "running" {
		t.Errorf("Expected status running, got %s", build.Status)
	}
}

func TestMonitor_RecordBuildComplete(t *testing.T) {
	config := types.MonitorConfig{
		Port:            8084,
		MetricsInterval: time.Second,
		AlertThresholds: map[string]float64{},
	}

	monitor := monitorpkg.NewMonitor(config)

	// Record build start and completion
	monitor.RecordBuildStart("build-1", "worker-1")

	resourceUsage := monitorpkg.ResourceMetrics{
		CPUPercent:    75.5,
		MemoryUsage:   1024 * 1024 * 100, // 100MB
		DiskIO:        1024 * 1024 * 10,  // 10MB
		NetworkIO:     1024 * 1024 * 5,   // 5MB
		MaxMemoryUsed: 1024 * 1024 * 120, // 120MB
	}

	monitor.RecordBuildComplete("build-1", true, time.Second*5, true, []string{"artifact1.jar", "artifact2.jar"}, resourceUsage)

	// Check system metrics
	metrics := monitor.GetSystemMetrics()
	if metrics.SuccessfulBuilds != 1 {
		t.Errorf("Expected 1 successful build, got %d", metrics.SuccessfulBuilds)
	}

	if metrics.AverageBuildTime != 5.0 {
		t.Errorf("Expected average build time 5.0, got %f", metrics.AverageBuildTime)
	}

	// Check build metrics
	builds := monitor.GetBuildMetrics()
	build := builds["build-1"]

	if build.Status != "completed" {
		t.Errorf("Expected status completed, got %s", build.Status)
	}

	if !build.CacheHit {
		t.Error("Expected cache hit to be true")
	}

	if len(build.Artifacts) != 2 {
		t.Errorf("Expected 2 artifacts, got %d", len(build.Artifacts))
	}
}

func TestMonitor_UpdateWorkerMetrics(t *testing.T) {
	config := types.MonitorConfig{
		Port:            8084,
		MetricsInterval: time.Second,
		AlertThresholds: map[string]float64{},
	}

	monitor := monitorpkg.NewMonitor(config)

	// Update worker metrics
	monitor.UpdateWorkerMetrics("worker-1", "active", 2, 10, 1, time.Now(), 45.5, 1024*1024*200)

	// Check worker metrics
	workers := monitor.GetWorkerMetrics()
	worker, exists := workers["worker-1"]
	if !exists {
		t.Fatal("Expected worker-1 to exist in metrics")
	}

	if worker.Status != "active" {
		t.Errorf("Expected status active, got %s", worker.Status)
	}

	if worker.ActiveBuilds != 2 {
		t.Errorf("Expected 2 active builds, got %d", worker.ActiveBuilds)
	}

	if worker.CPUUsage != 45.5 {
		t.Errorf("Expected CPU usage 45.5, got %f", worker.CPUUsage)
	}

	// Check system metrics
	metrics := monitor.GetSystemMetrics()
	if metrics.ActiveWorkers != 1 {
		t.Errorf("Expected 1 active worker, got %d", metrics.ActiveWorkers)
	}
}

func TestMonitor_TriggerAlert(t *testing.T) {
	config := types.MonitorConfig{
		Port:            8084,
		MetricsInterval: time.Second,
		AlertThresholds: map[string]float64{},
	}

	monitor := monitorpkg.NewMonitor(config)

	// Trigger an alert
	data := map[string]any{
		"worker_id": "worker-1",
		"error":     "Connection timeout",
	}

	monitor.TriggerAlert("worker_error", "critical", "Worker connection failed", data)

	// Check alerts
	alerts := monitor.GetAlerts()
	if len(alerts) != 1 {
		t.Fatalf("Expected 1 alert, got %d", len(alerts))
	}

	alert := alerts[0]
	if alert.Type != "worker_error" {
		t.Errorf("Expected alert type worker_error, got %s", alert.Type)
	}

	if alert.Severity != "critical" {
		t.Errorf("Expected severity critical, got %s", alert.Severity)
	}

	if alert.Message != "Worker connection failed" {
		t.Errorf("Expected message 'Worker connection failed', got %s", alert.Message)
	}

	if alert.Resolved {
		t.Error("Expected alert to be unresolved")
	}

	if alert.Data["worker_id"] != "worker-1" {
		t.Errorf("Expected worker_id worker-1, got %v", alert.Data["worker_id"])
	}
}

func TestMonitor_HTTPHandlers(t *testing.T) {
	config := types.MonitorConfig{
		Port:            8084,
		MetricsInterval: time.Second,
		AlertThresholds: map[string]float64{},
	}

	monitor := monitorpkg.NewMonitor(config)

	// Test metrics handler
	req, _ := http.NewRequest("GET", "/api/metrics", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(monitor.HandleMetrics)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Metrics handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var metrics monitorpkg.SystemMetrics
	err := json.Unmarshal(rr.Body.Bytes(), &metrics)
	if err != nil {
		t.Fatalf("Failed to parse metrics JSON: %v", err)
	}

	if metrics.LastUpdate.IsZero() {
		t.Error("Expected LastUpdate to be set")
	}

	// Test workers handler
	req, _ = http.NewRequest("GET", "/api/workers", nil)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(monitor.HandleWorkers)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Workers handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Test builds handler
	req, _ = http.NewRequest("GET", "/api/builds", nil)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(monitor.HandleBuilds)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Builds handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Test alerts handler
	req, _ = http.NewRequest("GET", "/api/alerts", nil)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(monitor.HandleAlerts)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Alerts handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Test health handler
	req, _ = http.NewRequest("GET", "/api/health", nil)
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(monitor.HandleHealth)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Health handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var health map[string]any
	err = json.Unmarshal(rr.Body.Bytes(), &health)
	if err != nil {
		t.Fatalf("Failed to parse health JSON: %v", err)
	}

	if health["status"] != "healthy" {
		t.Errorf("Expected status healthy, got %v", health["status"])
	}
}

func TestMonitor_AlertThresholds(t *testing.T) {
	config := types.MonitorConfig{
		Port:            8084,
		MetricsInterval: time.Millisecond * 50, // Very short interval for testing
		AlertThresholds: map[string]float64{
			"build_failure_rate": 0.1, // 10% failure rate threshold
		},
	}

	monitor := monitorpkg.NewMonitor(config)

	// Create some successful builds
	for i := 0; i < 5; i++ {
		monitor.RecordBuildStart(fmt.Sprintf("build-success-%d", i), "worker-1")
		resourceUsage := monitorpkg.ResourceMetrics{}
		monitor.RecordBuildComplete(fmt.Sprintf("build-success-%d", i), true, time.Second, false, []string{}, resourceUsage)
	}

	// Create some failed builds to exceed threshold
	for i := 0; i < 3; i++ {
		monitor.RecordBuildStart(fmt.Sprintf("build-failed-%d", i), "worker-1")
		resourceUsage := monitorpkg.ResourceMetrics{}
		monitor.RecordBuildComplete(fmt.Sprintf("build-failed-%d", i), false, time.Second, false, []string{}, resourceUsage)
	}

	// Manually trigger alert thresholds check
	monitor.CheckAlertThresholds()

	// Check if alert was triggered
	alerts := monitor.GetAlerts()

	// Look for high failure rate alert
	foundAlert := false
	for _, alert := range alerts {
		if alert.Type == "high_failure_rate" && !alert.Resolved {
			foundAlert = true
			if alert.Severity != "critical" {
				t.Errorf("Expected severity critical, got %s", alert.Severity)
			}
			break
		}
	}

	if !foundAlert {
		t.Error("Expected high failure rate alert to be triggered")
	}
}

func TestMonitor_Shutdown(t *testing.T) {
	config := types.MonitorConfig{
		Port:            8084,
		MetricsInterval: time.Second,
		AlertThresholds: map[string]float64{},
	}

	monitor := monitorpkg.NewMonitor(config)

	err := monitor.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}
