package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// Test monitor functionality
func TestMonitor_Creation(t *testing.T) {
	config := &MonitorConfig{
		Port:         8085,
		CoordinatorURL: "http://localhost:8080",
		Interval:      30,
		AlertThresholds: AlertThresholds{
			CPUUsage:    80.0,
			MemoryUsage: 85.0,
			BuildTime:   10 * time.Minute,
			ErrorRate:   5.0,
		},
	}
	
	monitor := NewMonitor(config)
	
	if monitor == nil {
		t.Fatal("Failed to create monitor")
	}
	
	if monitor.config.Port != config.Port {
		t.Errorf("Expected port %d, got %d", config.Port, monitor.config.Port)
	}
	
	if monitor.config.CoordinatorURL != config.CoordinatorURL {
		t.Errorf("Expected coordinator URL %s, got %s", config.CoordinatorURL, monitor.config.CoordinatorURL)
	}
	
	if monitor.config.AlertThresholds.CPUUsage != config.AlertThresholds.CPUUsage {
		t.Errorf("Expected CPU threshold %f, got %f", config.AlertThresholds.CPUUsage, monitor.config.AlertThresholds.CPUUsage)
	}
}

func TestMonitor_CollectMetrics(t *testing.T) {
	config := &MonitorConfig{
		Port:         8085,
		CoordinatorURL: "http://localhost:8080",
		Interval:      30,
	}
	
	monitor := NewMonitor(config)
	
	// Test metrics collection
	metrics := monitor.collectSystemMetrics()
	
	if metrics.CPUUsage < 0 || metrics.CPUUsage > 100 {
		t.Errorf("Expected CPU usage between 0-100, got %f", metrics.CPUUsage)
	}
	
	if metrics.MemoryUsage < 0 || metrics.MemoryUsage > 100 {
		t.Errorf("Expected memory usage between 0-100, got %f", metrics.MemoryUsage)
	}
	
	if metrics.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

func TestMonitor_CollectBuildMetrics(t *testing.T) {
	config := &MonitorConfig{
		Port:         8085,
		CoordinatorURL: "http://localhost:8080",
		Interval:      30,
	}
	
	monitor := NewMonitor(config)
	
	// Test build metrics collection
	buildMetrics, err := monitor.collectBuildMetrics()
	
	if err != nil {
		t.Logf("Build metrics collection failed (expected in test): %v", err)
	}
	
	if buildMetrics != nil {
		if buildMetrics.TotalBuilds < 0 {
			t.Errorf("Expected non-negative total builds, got %d", buildMetrics.TotalBuilds)
		}
		
		if buildMetrics.ActiveBuilds < 0 {
			t.Errorf("Expected non-negative active builds, got %d", buildMetrics.ActiveBuilds)
		}
		
		if buildMetrics.AverageBuildTime < 0 {
			t.Errorf("Expected non-negative average build time, got %v", buildMetrics.AverageBuildTime)
		}
	}
}

func TestMonitor_HTTPHandlers(t *testing.T) {
	config := &MonitorConfig{
		Port:         8085,
		CoordinatorURL: "http://localhost:8080",
		Interval:      30,
	}
	
	monitor := NewMonitor(config)
	
	// Test metrics endpoint
	req, _ := http.NewRequest("GET", "/api/metrics", nil)
	w := httptest.NewRecorder()
	
	monitor.handleMetrics(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	
	// Check expected metrics keys
	expectedKeys := []string{"cpu_usage", "memory_usage", "build_metrics", "worker_metrics", "timestamp"}
	for _, key := range expectedKeys {
		if _, exists := response[key]; !exists {
			t.Errorf("Expected metric key %s in response", key)
		}
	}
	
	// Test health check endpoint
	req, _ = http.NewRequest("GET", "/health", nil)
	w = httptest.NewRecorder()
	
	monitor.handleHealth(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var healthResponse map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &healthResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal health response: %v", err)
	}
	
	if healthResponse["status"] != "healthy" {
		t.Errorf("Expected healthy status, got %s", healthResponse["status"])
	}
}

func TestMonitor_AlertChecking(t *testing.T) {
	config := &MonitorConfig{
		Port:         8085,
		CoordinatorURL: "http://localhost:8080",
		Interval:      30,
		AlertThresholds: AlertThresholds{
			CPUUsage:    50.0, // Low threshold for testing
			MemoryUsage:  50.0,
			BuildTime:   1 * time.Minute,
			ErrorRate:   1.0,
		},
	}
	
	monitor := NewMonitor(config)
	
	// Create metrics that should trigger alerts
	metrics := SystemMetrics{
		CPUUsage:    80.0, // Above threshold
		MemoryUsage:  90.0, // Above threshold
		BuildMetrics: BuildMetricsSnapshot{
			AverageBuildTime: 5 * time.Minute, // Above threshold
			ErrorRate:       5.0,             // Above threshold
		},
		Timestamp: time.Now(),
	}
	
	// Check alerts
	alerts := monitor.checkAlerts(metrics)
	
	if len(alerts) == 0 {
		t.Error("Expected alerts for exceeded thresholds")
	}
	
	// Verify alert types
	alertTypes := make(map[string]bool)
	for _, alert := range alerts {
		alertTypes[alert.Type] = true
	}
	
	expectedAlerts := []string{"cpu_usage", "memory_usage", "build_time", "error_rate"}
	for _, expected := range expectedAlerts {
		if !alertTypes[expected] {
			t.Errorf("Expected alert type %s", expected)
		}
	}
}

func TestMonitor_StoreMetrics(t *testing.T) {
	config := &MonitorConfig{
		Port:         8085,
		CoordinatorURL: "http://localhost:8080",
		Interval:      30,
	}
	
	monitor := NewMonitor(config)
	
	// Create test metrics
	metrics := SystemMetrics{
		CPUUsage:   75.5,
		MemoryUsage: 60.2,
		BuildMetrics: BuildMetricsSnapshot{
			TotalBuilds:      100,
			ActiveBuilds:     5,
			CompletedBuilds:  95,
			AverageBuildTime: 3 * time.Minute,
			ErrorRate:        2.5,
		},
		Timestamp: time.Now(),
	}
	
	// Store metrics
	err := monitor.storeMetrics(metrics)
	if err != nil {
		t.Fatalf("Failed to store metrics: %v", err)
	}
	
	// Retrieve stored metrics
	retrieved, err := monitor.getRecentMetrics(1)
	if err != nil {
		t.Fatalf("Failed to retrieve metrics: %v", err)
	}
	
	if len(retrieved) == 0 {
		t.Error("Expected at least one metric record")
	}
	
	latest := retrieved[0]
	if latest.CPUUsage != metrics.CPUUsage {
		t.Errorf("Expected CPU usage %f, got %f", metrics.CPUUsage, latest.CPUUsage)
	}
	
	if latest.MemoryUsage != metrics.MemoryUsage {
		t.Errorf("Expected memory usage %f, got %f", metrics.MemoryUsage, latest.MemoryUsage)
	}
}

func TestMonitor_Configuration(t *testing.T) {
	config := &MonitorConfig{
		Port:         8085,
		CoordinatorURL: "http://localhost:8080",
		Interval:      30,
		RetentionDays: 7,
		LogLevel:      "info",
		AlertThresholds: AlertThresholds{
			CPUUsage:    80.0,
			MemoryUsage:  85.0,
			BuildTime:   10 * time.Minute,
			ErrorRate:   5.0,
		},
		Notifications: NotificationConfig{
			Email:     []string{"admin@example.com"},
			Webhook:   "http://localhost:9090/webhook",
			Slack:     "#alerts",
		},
	}
	
	monitor := NewMonitor(config)
	
	if monitor.config.Port != config.Port {
		t.Errorf("Expected port %d, got %d", config.Port, monitor.config.Port)
	}
	
	if monitor.config.Interval != config.Interval {
		t.Errorf("Expected interval %d, got %d", config.Interval, monitor.config.Interval)
	}
	
	if monitor.config.RetentionDays != config.RetentionDays {
		t.Errorf("Expected retention days %d, got %d", config.RetentionDays, monitor.config.RetentionDays)
	}
	
	if monitor.config.LogLevel != config.LogLevel {
		t.Errorf("Expected log level %s, got %s", config.LogLevel, monitor.config.LogLevel)
	}
}

func TestMonitor_Start(t *testing.T) {
	config := &MonitorConfig{
		Port:         8085,
		CoordinatorURL: "http://localhost:8080",
		Interval:      1, // 1 second for testing
		RetentionDays: 1,
	}
	
	monitor := NewMonitor(config)
	
	// Start monitor in goroutine
	go func() {
		err := monitor.Start()
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("Monitor start failed: %v", err)
		}
	}()
	
	// Give it time to start
	time.Sleep(2 * time.Second)
	
	// Stop monitor
	err := monitor.Stop()
	if err != nil {
		t.Errorf("Failed to stop monitor: %v", err)
	}
	
	// Give it time to stop
	time.Sleep(1 * time.Second)
}

func TestMonitor_LoadConfig(t *testing.T) {
	// Create test config file
	testConfig := MonitorConfig{
		Port:         8085,
		CoordinatorURL: "http://localhost:8080",
		Interval:      30,
		RetentionDays: 7,
		LogLevel:      "debug",
		AlertThresholds: AlertThresholds{
			CPUUsage:    75.0,
			MemoryUsage:  80.0,
			BuildTime:   8 * time.Minute,
			ErrorRate:   3.0,
		},
	}
	
	configJSON, _ := json.Marshal(testConfig)
	testConfigPath := "/tmp/test_monitor_config.json"
	
	err := ioutil.WriteFile(testConfigPath, configJSON, 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
	defer os.Remove(testConfigPath)
	
	// Load config
	loadedConfig, err := loadMonitorConfig(testConfigPath)
	if err != nil {
		t.Fatalf("Failed to load monitor config: %v", err)
	}
	
	if loadedConfig.Port != testConfig.Port {
		t.Errorf("Expected port %d, got %d", testConfig.Port, loadedConfig.Port)
	}
	
	if loadedConfig.CoordinatorURL != testConfig.CoordinatorURL {
		t.Errorf("Expected coordinator URL %s, got %s", testConfig.CoordinatorURL, loadedConfig.CoordinatorURL)
	}
	
	if loadedConfig.Interval != testConfig.Interval {
		t.Errorf("Expected interval %d, got %d", testConfig.Interval, loadedConfig.Interval)
	}
}

func TestMonitor_ExportMetrics(t *testing.T) {
	config := &MonitorConfig{
		Port:         8085,
		CoordinatorURL: "http://localhost:8080",
		Interval:      30,
	}
	
	monitor := NewMonitor(config)
	
	// Store some test metrics
	metrics := SystemMetrics{
		CPUUsage:   75.5,
		MemoryUsage: 60.2,
		Timestamp:   time.Now(),
	}
	
	err := monitor.storeMetrics(metrics)
	if err != nil {
		t.Fatalf("Failed to store metrics: %v", err)
	}
	
	// Export metrics to JSON
	exportPath := "/tmp/metrics_export.json"
	err = monitor.exportMetrics(exportPath, "json")
	if err != nil {
		t.Fatalf("Failed to export metrics: %v", err)
	}
	defer os.Remove(exportPath)
	
	// Verify export file exists and contains data
	data, err := ioutil.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("Failed to read exported metrics: %v", err)
	}
	
	if len(data) == 0 {
		t.Error("Export file is empty")
	}
	
	var exportedMetrics []SystemMetrics
	err = json.Unmarshal(data, &exportedMetrics)
	if err != nil {
		t.Fatalf("Failed to unmarshal exported metrics: %v", err)
	}
	
	if len(exportedMetrics) == 0 {
		t.Error("No metrics were exported")
	}
}

func TestMonitor_MetricsAggregation(t *testing.T) {
	config := &MonitorConfig{
		Port:         8085,
		CoordinatorURL: "http://localhost:8080",
		Interval:      30,
	}
	
	monitor := NewMonitor(config)
	
	// Store test metrics over time
	now := time.Now()
	for i := 0; i < 10; i++ {
		metrics := SystemMetrics{
			CPUUsage:   float64(70 + i), // 70-79%
			MemoryUsage: float64(50 + i), // 50-59%
			Timestamp:   now.Add(time.Duration(i) * time.Minute),
		}
		
		err := monitor.storeMetrics(metrics)
		if err != nil {
			t.Fatalf("Failed to store metrics %d: %v", i, err)
		}
	}
	
	// Get aggregated metrics
	aggregated, err := monitor.getAggregatedMetrics(now, now.Add(9*time.Minute))
	if err != nil {
		t.Fatalf("Failed to get aggregated metrics: %v", err)
	}
	
	if aggregated.CPUUsage.Average < 74 || aggregated.CPUUsage.Average > 75 {
		t.Errorf("Expected CPU average ~74.5, got %f", aggregated.CPUUsage.Average)
	}
	
	if aggregated.MemoryUsage.Average < 54 || aggregated.MemoryUsage.Average > 55 {
		t.Errorf("Expected memory average ~54.5, got %f", aggregated.MemoryUsage.Average)
	}
	
	if aggregated.CPUUsage.Min != 70.0 {
		t.Errorf("Expected CPU min 70.0, got %f", aggregated.CPUUsage.Min)
	}
	
	if aggregated.CPUUsage.Max != 79.0 {
		t.Errorf("Expected CPU max 79.0, got %f", aggregated.CPUUsage.Max)
	}
}