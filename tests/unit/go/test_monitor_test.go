package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

// TestMonitorStartup tests monitor startup and initialization
func TestMonitorStartup(t *testing.T) {
	// Create temporary directory for monitor
	tempDir, err := os.MkdirTemp("", "monitor-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create monitor configuration
	config := MonitorConfig{
		Host:               "localhost",
		Port:               8082,
		DataDir:            tempDir,
		MetricsInterval:    30 * time.Second,
		AlertThresholds: map[string]float64{
			"cpu_usage":       0.8,
			"memory_usage":    0.85,
			"disk_usage":      0.9,
			"error_rate":      0.05,
			"response_time":    5.0,
		},
		CoordinatorURL:    "http://localhost:8080",
		CacheServerURL:    "http://localhost:8081",
		NotificationConfig: NotificationConfig{
			Enabled: true,
			Email:   false,
			Slack:   false,
			Webhook: "http://localhost:9090/webhook",
		},
		RetentionDays: 30,
	}
	
	// Create monitor instance
	monitor, err := NewMonitor(config)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	
	// Verify monitor initialization
	if monitor.Host != config.Host {
		t.Errorf("Expected host '%s', got '%s'", config.Host, monitor.Host)
	}
	
	if monitor.Port != config.Port {
		t.Errorf("Expected port %d, got %d", config.Port, monitor.Port)
	}
	
	if monitor.Status != "initializing" {
		t.Errorf("Expected status 'initializing', got '%s'", monitor.Status)
	}
	
	if len(monitor.AlertThresholds) != len(config.AlertThresholds) {
		t.Errorf("Expected %d alert thresholds, got %d", len(config.AlertThresholds), len(monitor.AlertThresholds))
	}
}

// TestMetricsCollection tests metrics collection from various sources
func TestMetricsCollection(t *testing.T) {
	// Create temporary directory for monitor
	tempDir, err := os.MkdirTemp("", "monitor-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create monitor configuration
	config := MonitorConfig{
		Host:               "localhost",
		Port:               8082,
		DataDir:            tempDir,
		MetricsInterval:    30 * time.Second,
		AlertThresholds:    map[string]float64{},
		CoordinatorURL:    "http://localhost:8080",
		CacheServerURL:    "http://localhost:8081",
		NotificationConfig: NotificationConfig{},
		RetentionDays:      30,
	}
	
	// Mock coordinator and cache server endpoints
	var collectedMetrics map[string]interface{}
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/metrics":
			// Mock coordinator metrics
			coordinatorMetrics := map[string]interface{}{
				"total_builds":     100,
				"active_builds":    5,
				"completed_builds": 95,
				"failed_builds":     10,
				"total_workers":     3,
				"available_workers": 2,
				"busy_workers":      1,
				"average_build_time": 180,
			}
			
			if r.URL.Query().Get("source") == "coordinator" {
				collectedMetrics["coordinator"] = coordinatorMetrics
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(coordinatorMetrics)
			
		case "/api/cache/statistics":
			// Mock cache server metrics
			cacheMetrics := map[string]interface{}{
				"total_entries": 1500,
				"total_size_mb":  512,
				"hit_rate":       0.75,
				"miss_rate":      0.25,
				"evictions":      50,
			}
			
			if r.URL.Query().Get("source") == "cache" {
				collectedMetrics["cache"] = cacheMetrics
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(cacheMetrics)
		}
	}))
	defer server.Close()
	
	// Update config to use mock server URLs
	config.CoordinatorURL = server.URL
	config.CacheServerURL = server.URL
	
	// Create monitor instance
	monitor, err := NewMonitor(config)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	
	// Test metrics collection
	metrics, err := monitor.CollectMetrics()
	if err != nil {
		t.Errorf("Failed to collect metrics: %v", err)
	}
	
	// Verify collected metrics
	if metrics.CoordinatorMetrics.TotalBuilds != 100 {
		t.Errorf("Expected total builds 100, got %d", metrics.CoordinatorMetrics.TotalBuilds)
	}
	
	if metrics.CoordinatorMetrics.ActiveBuilds != 5 {
		t.Errorf("Expected active builds 5, got %d", metrics.CoordinatorMetrics.ActiveBuilds)
	}
	
	if metrics.CacheMetrics.TotalEntries != 1500 {
		t.Errorf("Expected total cache entries 1500, got %d", metrics.CacheMetrics.TotalEntries)
	}
	
	if metrics.CacheMetrics.HitRate != 0.75 {
		t.Errorf("Expected cache hit rate 0.75, got %f", metrics.CacheMetrics.HitRate)
	}
	
	// Verify local system metrics
	if metrics.SystemMetrics.CPUUsage < 0 || metrics.SystemMetrics.CPUUsage > 1 {
		t.Errorf("CPU usage should be between 0 and 1, got %f", metrics.SystemMetrics.CPUUsage)
	}
	
	if metrics.SystemMetrics.MemoryUsage < 0 || metrics.SystemMetrics.MemoryUsage > 1 {
		t.Errorf("Memory usage should be between 0 and 1, got %f", metrics.SystemMetrics.MemoryUsage)
	}
	
	if metrics.SystemMetrics.DiskUsage < 0 || metrics.SystemMetrics.DiskUsage > 1 {
		t.Errorf("Disk usage should be between 0 and 1, got %f", metrics.SystemMetrics.DiskUsage)
	}
}

// TestAlertGeneration tests alert generation based on thresholds
func TestAlertGeneration(t *testing.T) {
	// Create temporary directory for monitor
	tempDir, err := os.MkdirTemp("", "monitor-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create monitor configuration with alert thresholds
	config := MonitorConfig{
		Host:            "localhost",
		Port:            8082,
		DataDir:         tempDir,
		MetricsInterval: 30 * time.Second,
		AlertThresholds: map[string]float64{
			"cpu_usage":    0.8,
			"memory_usage": 0.85,
			"disk_usage":   0.9,
			"error_rate":   0.05,
			"response_time": 5.0,
		},
		CoordinatorURL:    "http://localhost:8080",
		CacheServerURL:    "http://localhost:8081",
		NotificationConfig: NotificationConfig{
			Enabled: true,
			Email:   false,
			Slack:   false,
			Webhook: "http://localhost:9090/webhook",
		},
		RetentionDays: 30,
	}
	
	// Create monitor instance
	monitor, err := NewMonitor(config)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	
	// Create test metrics that should trigger alerts
	testMetrics := SystemMetrics{
		Timestamp:    time.Now(),
		CPUUsage:     0.9, // Above threshold
		MemoryUsage:  0.7, // Below threshold
		DiskUsage:    0.95, // Above threshold
		LoadAverage:  []float64{1.5, 1.3, 1.2},
		NetworkIO:    NetworkIOMetrics{
			BytesReceived: 1024 * 1024,
			BytesSent:     512 * 1024,
		},
		ErrorRate:     0.1, // Above threshold
		ResponseTime:  3.0, // Below threshold
	}
	
	// Test alert generation
	alerts := monitor.GenerateAlerts(testMetrics)
	
	// Verify alerts
	expectedAlerts := 3 // CPU, disk, error rate
	if len(alerts) != expectedAlerts {
		t.Errorf("Expected %d alerts, got %d", expectedAlerts, len(alerts))
	}
	
	// Check specific alerts
	alertTypes := make(map[string]bool)
	for _, alert := range alerts {
		alertTypes[alert.Type] = true
		
		// Verify alert structure
		if alert.Severity == "" {
			t.Error("Alert should have severity level")
		}
		
		if alert.Message == "" {
			t.Error("Alert should have message")
		}
		
		if alert.Timestamp.IsZero() {
			t.Error("Alert should have timestamp")
		}
	}
	
	if !alertTypes["cpu_usage"] {
		t.Error("Expected CPU usage alert")
	}
	
	if !alertTypes["disk_usage"] {
		t.Error("Expected disk usage alert")
	}
	
	if !alertTypes["error_rate"] {
		t.Error("Expected error rate alert")
	}
}

// TestNotificationSending tests notification sending for alerts
func TestNotificationSending(t *testing.T) {
	// Create temporary directory for monitor
	tempDir, err := os.MkdirTemp("", "monitor-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Track webhook notifications
	var webhookNotifications []Alert
	
	// Mock webhook server
	webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/webhook" {
			var alert Alert
			err := json.NewDecoder(r.Body).Decode(&alert)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			
			webhookNotifications = append(webhookNotifications, alert)
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer webhookServer.Close()
	
	// Create monitor configuration with webhook notifications
	config := MonitorConfig{
		Host:               "localhost",
		Port:               8082,
		DataDir:            tempDir,
		MetricsInterval:    30 * time.Second,
		AlertThresholds:    map[string]float64{},
		CoordinatorURL:     "http://localhost:8080",
		CacheServerURL:     "http://localhost:8081",
		NotificationConfig: NotificationConfig{
			Enabled: true,
			Email:   false,
			Slack:   false,
			Webhook: webhookServer.URL + "/webhook",
		},
		RetentionDays: 30,
	}
	
	// Create monitor instance
	monitor, err := NewMonitor(config)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	
	// Create test alerts
	alerts := []Alert{
		{
			Type:      "cpu_usage",
			Severity:  "warning",
			Message:   "CPU usage is above threshold: 85%",
			Timestamp: time.Now(),
			Value:     0.85,
			Threshold: 0.8,
		},
		{
			Type:      "memory_usage",
			Severity:  "critical",
			Message:   "Memory usage is critical: 95%",
			Timestamp: time.Now(),
			Value:     0.95,
			Threshold: 0.9,
		},
	}
	
	// Send notifications
	err = monitor.SendNotifications(alerts)
	if err != nil {
		t.Errorf("Failed to send notifications: %v", err)
	}
	
	// Verify webhook notifications
	if len(webhookNotifications) != len(alerts) {
		t.Errorf("Expected %d webhook notifications, got %d", len(alerts), len(webhookNotifications))
	}
	
	for i, expectedAlert := range alerts {
		if i >= len(webhookNotifications) {
			t.Errorf("Missing webhook notification for alert %d", i)
			continue
		}
		
		actualAlert := webhookNotifications[i]
		if actualAlert.Type != expectedAlert.Type {
			t.Errorf("Expected alert type '%s', got '%s'", expectedAlert.Type, actualAlert.Type)
		}
		
		if actualAlert.Severity != expectedAlert.Severity {
			t.Errorf("Expected severity '%s', got '%s'", expectedAlert.Severity, actualAlert.Severity)
		}
		
		if actualAlert.Message != expectedAlert.Message {
			t.Errorf("Expected message '%s', got '%s'", expectedAlert.Message, actualAlert.Message)
		}
	}
}

// TestMetricsStorage tests metrics storage and retrieval
func TestMetricsStorage(t *testing.T) {
	// Create temporary directory for monitor
	tempDir, err := os.MkdirTemp("", "monitor-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create monitor configuration
	config := MonitorConfig{
		Host:               "localhost",
		Port:               8082,
		DataDir:            tempDir,
		MetricsInterval:    30 * time.Second,
		AlertThresholds:    map[string]float64{},
		CoordinatorURL:     "http://localhost:8080",
		CacheServerURL:     "http://localhost:8081",
		NotificationConfig: NotificationConfig{},
		RetentionDays:      30,
	}
	
	// Create monitor instance
	monitor, err := NewMonitor(config)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	
	// Create test metrics
	testMetrics := SystemMetrics{
		Timestamp:    time.Now(),
		CPUUsage:     0.45,
		MemoryUsage:  0.67,
		DiskUsage:    0.23,
		LoadAverage:  []float64{1.2, 1.1, 1.0},
		NetworkIO: NetworkIOMetrics{
			BytesReceived: 1024 * 1024,
			BytesSent:     512 * 1024,
		},
		ErrorRate:    0.02,
		ResponseTime: 1.5,
	}
	
	// Store metrics
	err = monitor.StoreMetrics(testMetrics)
	if err != nil {
		t.Errorf("Failed to store metrics: %v", err)
	}
	
	// Retrieve metrics (last hour)
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()
	
	retrievedMetrics, err := monitor.GetMetrics(startTime, endTime)
	if err != nil {
		t.Errorf("Failed to retrieve metrics: %v", err)
	}
	
	if len(retrievedMetrics) == 0 {
		t.Error("Retrieved metrics should not be empty")
	}
	
	// Verify stored metrics
	lastMetrics := retrievedMetrics[len(retrievedMetrics)-1]
	if lastMetrics.CPUUsage != testMetrics.CPUUsage {
		t.Errorf("Expected CPU usage %f, got %f", testMetrics.CPUUsage, lastMetrics.CPUUsage)
	}
	
	if lastMetrics.MemoryUsage != testMetrics.MemoryUsage {
		t.Errorf("Expected memory usage %f, got %f", testMetrics.MemoryUsage, lastMetrics.MemoryUsage)
	}
	
	if lastMetrics.DiskUsage != testMetrics.DiskUsage {
		t.Errorf("Expected disk usage %f, got %f", testMetrics.DiskUsage, lastMetrics.DiskUsage)
	}
}

// TestMonitorAPI tests monitor HTTP API
func TestMonitorAPI(t *testing.T) {
	// Create temporary directory for monitor
	tempDir, err := os.MkdirTemp("", "monitor-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create monitor instance
	config := MonitorConfig{
		Host:               "localhost",
		Port:               8082,
		DataDir:            tempDir,
		MetricsInterval:    30 * time.Second,
		AlertThresholds:    map[string]float64{},
		CoordinatorURL:     "http://localhost:8080",
		CacheServerURL:     "http://localhost:8081",
		NotificationConfig: NotificationConfig{},
		RetentionDays:      30,
	}
	
	monitor, err := NewMonitor(config)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	
	// Create router
	router := mux.NewRouter()
	router.HandleFunc("/api/metrics", monitor.GetMetricsHandler).Methods("GET")
	router.HandleFunc("/api/metrics/current", monitor.GetCurrentMetricsHandler).Methods("GET")
	router.HandleFunc("/api/alerts", monitor.GetAlertsHandler).Methods("GET")
	router.HandleFunc("/api/health", monitor.HealthCheckHandler).Methods("GET")
	router.HandleFunc("/api/config", monitor.GetConfigHandler).Methods("GET")
	router.HandleFunc("/api/config", monitor.UpdateConfigHandler).Methods("PUT")
	
	// Test current metrics endpoint
	req, _ := http.NewRequest("GET", "/api/metrics/current", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d for current metrics, got %d", http.StatusOK, status)
	}
	
	var currentMetrics map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &currentMetrics)
	if err != nil {
		t.Errorf("Failed to unmarshal current metrics: %v", err)
	}
	
	if currentMetrics["timestamp"] == nil {
		t.Error("Current metrics should include timestamp")
	}
	
	if currentMetrics["system_metrics"] == nil {
		t.Error("Current metrics should include system metrics")
	}
	
	// Test alerts endpoint
	req, _ = http.NewRequest("GET", "/api/alerts", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d for alerts, got %d", http.StatusOK, status)
	}
	
	var alerts []Alert
	err = json.Unmarshal(rr.Body.Bytes(), &alerts)
	if err != nil {
		t.Errorf("Failed to unmarshal alerts: %v", err)
	}
	
	// Test health check endpoint
	req, _ = http.NewRequest("GET", "/api/health", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d for health check, got %d", http.StatusOK, status)
	}
	
	var health map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &health)
	if err != nil {
		t.Errorf("Failed to unmarshal health check: %v", err)
	}
	
	if health["status"] == nil {
		t.Error("Health check should include status")
	}
	
	if health["timestamp"] == nil {
		t.Error("Health check should include timestamp")
	}
	
	// Test configuration endpoint
	req, _ = http.NewRequest("GET", "/api/config", nil)
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d for config, got %d", http.StatusOK, status)
	}
	
	var configResponse map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &configResponse)
	if err != nil {
		t.Errorf("Failed to unmarshal config: %v", err)
	}
	
	if configResponse["alert_thresholds"] == nil {
		t.Error("Config should include alert thresholds")
	}
	
	// Test configuration update
	updatedConfig := map[string]interface{}{
		"alert_thresholds": map[string]float64{
			"cpu_usage":    0.7,
			"memory_usage":  0.8,
			"disk_usage":   0.85,
		},
	}
	
	reqBody, _ := json.Marshal(updatedConfig)
	req, _ = http.NewRequest("PUT", "/api/config", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d for config update, got %d", http.StatusOK, status)
	}
}

// TestMetricsRetention tests metrics data retention policy
func TestMetricsRetention(t *testing.T) {
	// Create temporary directory for monitor
	tempDir, err := os.MkdirTemp("", "monitor-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create monitor configuration with short retention period
	config := MonitorConfig{
		Host:               "localhost",
		Port:               8082,
		DataDir:            tempDir,
		MetricsInterval:    30 * time.Second,
		AlertThresholds:     map[string]float64{},
		CoordinatorURL:     "http://localhost:8080",
		CacheServerURL:     "http://localhost:8081",
		NotificationConfig:  NotificationConfig{},
		RetentionDays:      1, // Very short retention for testing
	}
	
	// Create monitor instance
	monitor, err := NewMonitor(config)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	
	// Store old metrics (2 days ago)
	oldTimestamp := time.Now().Add(-48 * time.Hour)
	oldMetrics := SystemMetrics{
		Timestamp:   oldTimestamp,
		CPUUsage:    0.45,
		MemoryUsage:  0.67,
		DiskUsage:   0.23,
		LoadAverage: []float64{1.2, 1.1, 1.0},
		NetworkIO: NetworkIOMetrics{
			BytesReceived: 1024 * 1024,
			BytesSent:     512 * 1024,
		},
		ErrorRate:    0.02,
		ResponseTime: 1.5,
	}
	
	// Store old metrics
	err = monitor.StoreMetrics(oldMetrics)
	if err != nil {
		t.Errorf("Failed to store old metrics: %v", err)
	}
	
	// Store recent metrics (now)
	recentMetrics := SystemMetrics{
		Timestamp:   time.Now(),
		CPUUsage:    0.55,
		MemoryUsage:  0.72,
		DiskUsage:   0.28,
		LoadAverage: []float64{1.3, 1.2, 1.1},
		NetworkIO: NetworkIOMetrics{
			BytesReceived: 2048 * 1024,
			BytesSent:     1024 * 1024,
		},
		ErrorRate:    0.01,
		ResponseTime: 1.8,
	}
	
	// Store recent metrics
	err = monitor.StoreMetrics(recentMetrics)
	if err != nil {
		t.Errorf("Failed to store recent metrics: %v", err)
	}
	
	// Run retention cleanup
	err = monitor.CleanupOldMetrics()
	if err != nil {
		t.Errorf("Failed to cleanup old metrics: %v", err)
	}
	
	// Verify old metrics were removed and recent metrics remain
	startTime := time.Now().Add(-72 * time.Hour)
	endTime := time.Now()
	
	retrievedMetrics, err := monitor.GetMetrics(startTime, endTime)
	if err != nil {
		t.Errorf("Failed to retrieve metrics: %v", err)
	}
	
	if len(retrievedMetrics) != 1 {
		t.Errorf("Expected 1 metrics entry after cleanup, got %d", len(retrievedMetrics))
	}
	
	if len(retrievedMetrics) > 0 && retrievedMetrics[0].CPUUsage != recentMetrics.CPUUsage {
		t.Errorf("Expected recent metrics to remain, got old metrics")
	}
}

// TestAlertingRules tests custom alerting rules
func TestAlertingRules(t *testing.T) {
	// Create temporary directory for monitor
	tempDir, err := os.MkdirTemp("", "monitor-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create monitor configuration with custom alerting rules
	config := MonitorConfig{
		Host:            "localhost",
		Port:            8082,
		DataDir:         tempDir,
		MetricsInterval: 30 * time.Second,
		AlertThresholds: map[string]float64{
			"cpu_usage":           0.8,
			"memory_usage":        0.85,
			"disk_usage":         0.9,
			"error_rate":         0.05,
			"response_time":      5.0,
			"build_failure_rate": 0.1,
		},
		AlertingRules: []AlertRule{
			{
				Name:        "High CPU",
				Metric:      "cpu_usage",
				Operator:    ">",
				Threshold:   0.8,
				Duration:    5 * time.Minute,
				Severity:    "warning",
				Message:     "CPU usage is above 80% for 5 minutes",
			},
			{
				Name:        "Critical Memory",
				Metric:      "memory_usage",
				Operator:    ">",
				Threshold:   0.9,
				Duration:    2 * time.Minute,
				Severity:    "critical",
				Message:     "Memory usage is critically high",
			},
		},
		CoordinatorURL:     "http://localhost:8080",
		CacheServerURL:     "http://localhost:8081",
		NotificationConfig:  NotificationConfig{},
		RetentionDays:      30,
	}
	
	// Create monitor instance
	monitor, err := NewMonitor(config)
	if err != nil {
		t.Fatalf("Failed to create monitor: %v", err)
	}
	
	// Test metrics that trigger one rule but not the other
	testMetrics := SystemMetrics{
		Timestamp:    time.Now(),
		CPUUsage:     0.85, // Triggers high CPU rule
		MemoryUsage:  0.88, // Between 0.85 and 0.9, triggers basic threshold but not custom rule
		DiskUsage:    0.5,
		LoadAverage:  []float64{1.2, 1.1, 1.0},
		NetworkIO:    NetworkIOMetrics{
			BytesReceived: 1024 * 1024,
			BytesSent:     512 * 1024,
		},
		ErrorRate:    0.02,
		ResponseTime:  1.5,
	}
	
	// Generate alerts
	alerts := monitor.GenerateAlerts(testMetrics)
	
	// Verify alerts
	// Should have alerts from both basic thresholds and custom rules
	if len(alerts) < 2 {
		t.Errorf("Expected at least 2 alerts, got %d", len(alerts))
	}
	
	// Check for specific alert types
	alertTypes := make(map[string]bool)
	for _, alert := range alerts {
		alertTypes[alert.Type] = true
	}
	
	if !alertTypes["cpu_usage"] {
		t.Error("Expected CPU usage alert")
	}
	
	if !alertTypes["memory_usage"] {
		t.Error("Expected memory usage alert")
	}
	
	// Check alert severities
	for _, alert := range alerts {
		if alert.Type == "cpu_usage" && alert.Severity != "warning" {
			t.Errorf("CPU alert should have 'warning' severity, got '%s'", alert.Severity)
		}
		
		if alert.Type == "memory_usage" && alert.Severity != "warning" {
			t.Errorf("Memory alert should have 'warning' severity from basic threshold, got '%s'", alert.Severity)
		}
	}
}

// TestMonitorValidation tests monitor configuration validation
func TestMonitorValidation(t *testing.T) {
	testCases := []struct {
		name        string
		config      MonitorConfig
		expectError bool
	}{
		{
			name: "Valid configuration",
			config: MonitorConfig{
				Host:               "localhost",
				Port:               8082,
				DataDir:            "/tmp/monitor",
				MetricsInterval:    30 * time.Second,
				AlertThresholds:    map[string]float64{"cpu_usage": 0.8},
				CoordinatorURL:    "http://localhost:8080",
				CacheServerURL:    "http://localhost:8081",
				NotificationConfig: NotificationConfig{},
				RetentionDays:      30,
			},
			expectError: false,
		},
		{
			name: "Invalid port",
			config: MonitorConfig{
				Host:               "localhost",
				Port:               0,
				DataDir:            "/tmp/monitor",
				MetricsInterval:    30 * time.Second,
				AlertThresholds:    map[string]float64{},
				CoordinatorURL:    "http://localhost:8080",
				CacheServerURL:    "http://localhost:8081",
				NotificationConfig: NotificationConfig{},
				RetentionDays:      30,
			},
			expectError: true,
		},
		{
			name: "Zero metrics interval",
			config: MonitorConfig{
				Host:               "localhost",
				Port:               8082,
				DataDir:            "/tmp/monitor",
				MetricsInterval:    0,
				AlertThresholds:    map[string]float64{},
				CoordinatorURL:    "http://localhost:8080",
				CacheServerURL:    "http://localhost:8081",
				NotificationConfig: NotificationConfig{},
				RetentionDays:      30,
			},
			expectError: true,
		},
		{
			name: "Negative retention days",
			config: MonitorConfig{
				Host:               "localhost",
				Port:               8082,
				DataDir:            "/tmp/monitor",
				MetricsInterval:    30 * time.Second,
				AlertThresholds:    map[string]float64{},
				CoordinatorURL:    "http://localhost:8080",
				CacheServerURL:    "http://localhost:8081",
				NotificationConfig: NotificationConfig{},
				RetentionDays:      -1,
			},
			expectError: true,
		},
		{
			name: "Invalid coordinator URL",
			config: MonitorConfig{
				Host:               "localhost",
				Port:               8082,
				DataDir:            "/tmp/monitor",
				MetricsInterval:    30 * time.Second,
				AlertThresholds:    map[string]float64{},
				CoordinatorURL:    "not-a-url",
				CacheServerURL:     "http://localhost:8081",
				NotificationConfig: NotificationConfig{},
				RetentionDays:      30,
			},
			expectError: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			monitor, err := NewMonitor(tc.config)
			
			if tc.expectError {
				if err == nil {
					t.Error("Expected error but monitor creation succeeded")
				}
				if monitor != nil {
					t.Error("Expected nil monitor on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected success but got error: %v", err)
				}
				if monitor == nil {
					t.Error("Expected non-nil monitor on success")
				}
			}
		})
	}
}