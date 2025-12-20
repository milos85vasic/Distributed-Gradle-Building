package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// MonitorConfig holds configuration for monitoring service
type MonitorConfig struct {
	Port             int                `json:"port"`
	MetricsInterval  time.Duration      `json:"metrics_interval"`
	AlertThresholds  map[string]float64 `json:"alert_thresholds"`
	EnablePrometheus bool               `json:"enable_prometheus"`
	EnableJaeger     bool               `json:"enable_jaeger"`
	JaegerEndpoint   string             `json:"jaeger_endpoint"`
}

// Monitor represents the monitoring service
type Monitor struct {
	config     *MonitorConfig
	httpServer *http.Server
}

// loadMonitorConfig loads monitor configuration from file
func loadMonitorConfig(filename string) (*MonitorConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		// Return default config if file doesn't exist
		if os.IsNotExist(err) {
			return &MonitorConfig{
				Port:            8084,
				MetricsInterval: 30 * time.Second,
				AlertThresholds: map[string]float64{
					"cpu_usage":    80.0,
					"memory_usage": 85.0,
					"disk_usage":   90.0,
				},
				EnablePrometheus: false,
				EnableJaeger:     false,
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

// NewMonitor creates a new monitor service
func NewMonitor(config *MonitorConfig) *Monitor {
	return &Monitor{
		config: config,
	}
}

// Start starts the monitor service
func (m *Monitor) Start() error {
	// Set up HTTP routes
	http.HandleFunc("/health", m.healthHandler)
	http.HandleFunc("/metrics", m.metricsHandler)
	http.HandleFunc("/api/metrics", m.apiMetricsHandler)

	// Start HTTP server
	addr := fmt.Sprintf(":%d", m.config.Port)
	log.Printf("Monitor service started on port %d", m.config.Port)

	m.httpServer = &http.Server{
		Addr:    addr,
		Handler: nil, // Use default mux
	}

	return m.httpServer.ListenAndServe()
}

// healthHandler provides a basic health check endpoint
func (m *Monitor) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "monitor",
	})
}

// metricsHandler provides basic metrics endpoint
func (m *Monitor) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"service": "monitor",
		"metrics": map[string]any{
			"uptime":  "running",
			"version": "1.0.0",
		},
	})
}

// apiMetricsHandler provides detailed metrics for ML service consumption
func (m *Monitor) apiMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Mock data for now - in a real implementation, this would collect from actual monitoring
	response := map[string]any{
		"workers": map[string]any{
			"worker-1": map[string]any{
				"id":            "worker-1",
				"status":        "active",
				"cpu_usage":     0.65,
				"memory_usage":  0.55,
				"active_builds": 1,
			},
			"worker-2": map[string]any{
				"id":            "worker-2",
				"status":        "active",
				"cpu_usage":     0.45,
				"memory_usage":  0.35,
				"active_builds": 0,
			},
			"worker-3": map[string]any{
				"id":            "worker-3",
				"status":        "idle",
				"cpu_usage":     0.25,
				"memory_usage":  0.20,
				"active_builds": 0,
			},
		},
		"builds": map[string]any{
			"build-123": map[string]any{
				"build_id":       "build-123",
				"worker_id":      "worker-1",
				"project_path":   "/projects/myapp",
				"task_name":      "build",
				"success":        true,
				"start_time":     time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
				"end_time":       time.Now().Add(-8 * time.Minute).Format(time.RFC3339),
				"cache_hit_rate": 0.75,
				"resource_usage": map[string]any{
					"cpu_percent":  0.65,
					"memory_usage": 512 * 1024 * 1024, // 512MB
					"disk_io":      1024 * 1024,       // 1MB
					"network_io":   512 * 1024,        // 512KB
				},
				"error_message": "",
			},
			"build-124": map[string]any{
				"build_id":       "build-124",
				"worker_id":      "worker-2",
				"project_path":   "/projects/web",
				"task_name":      "test",
				"success":        false,
				"start_time":     time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
				"end_time":       time.Now().Add(-3 * time.Minute).Format(time.RFC3339),
				"cache_hit_rate": 0.20,
				"resource_usage": map[string]any{
					"cpu_percent":  0.80,
					"memory_usage": 1024 * 1024 * 1024, // 1GB
					"disk_io":      2048 * 1024,        // 2MB
					"network_io":   1024 * 1024,        // 1MB
				},
				"error_message": "Test failures detected",
			},
		},
		"system": map[string]any{
			"total_builds":       150,
			"successful_builds":  135,
			"failed_builds":      15,
			"active_workers":     3,
			"queue_length":       2,
			"average_build_time": 450.5, // seconds
			"cache_hit_rate":     0.72,
			"last_update":        time.Now().Format(time.RFC3339),
		},
	}

	json.NewEncoder(w).Encode(response)
}

// Main monitor application entry point
func monitorMain() {
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

	log.Printf("Starting monitor on port %d", config.Port)
	err = monitor.Start()
	if err != nil {
		log.Fatalf("Monitor failed: %v", err)
	}
}

// Main function for monitor
func main() {
	monitorMain()
}
