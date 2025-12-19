package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"distributed-gradle-building/ml/service"
)

// TestMLServiceIntegration tests the ML service integration
func TestMLServiceIntegration(t *testing.T) {
	// Create ML service
	mlService := service.NewMLService()

	// Add test data
	addTestDataToMLService(mlService)

	// Test predictions
	t.Run("BuildTimePrediction", func(t *testing.T) {
		duration, confidence := mlService.PredictBuildTime("/test/project", "build", nil)

		if duration <= 0 {
			t.Errorf("Expected positive build duration, got %v", duration)
		}

		if confidence < 0 || confidence > 1 {
			t.Errorf("Expected confidence between 0 and 1, got %v", confidence)
		}

		t.Logf("Predicted build time: %v with confidence %.2f", duration, confidence)
	})

	t.Run("ResourcePrediction", func(t *testing.T) {
		resources := mlService.PredictResourceNeeds("/test/project", "build")

		if resources.CPU < 0 || resources.CPU > 1 {
			t.Errorf("Expected CPU usage between 0 and 1, got %v", resources.CPU)
		}

		if resources.Memory < 0 || resources.Memory > 1 {
			t.Errorf("Expected memory usage between 0 and 1, got %v", resources.Memory)
		}

		t.Logf("Predicted resources: CPU=%.2f, Memory=%.2f, Disk=%.2f",
			resources.CPU, resources.Memory, resources.Disk)
	})

	t.Run("FailureRiskPrediction", func(t *testing.T) {
		risk := mlService.PredictFailureRisk("/test/project", "build")

		if risk < 0 || risk > 1 {
			t.Errorf("Expected failure risk between 0 and 1, got %v", risk)
		}

		t.Logf("Predicted failure risk: %.2f", risk)
	})

	t.Run("CacheHitRatePrediction", func(t *testing.T) {
		hitRate := mlService.PredictCacheHitRate("/test/project", "build")

		if hitRate < 0 || hitRate > 1 {
			t.Errorf("Expected cache hit rate between 0 and 1, got %v", hitRate)
		}

		t.Logf("Predicted cache hit rate: %.2f", hitRate)
	})

	t.Run("ScalingAdvice", func(t *testing.T) {
		advice := mlService.PredictScalingNeeds(5, 0.7, 2)

		if advice.WorkersNeeded < 0 {
			t.Errorf("Expected non-negative worker count, got %d", advice.WorkersNeeded)
		}

		if advice.Confidence < 0 || advice.Confidence > 1 {
			t.Errorf("Expected confidence between 0 and 1, got %v", advice.Confidence)
		}

		t.Logf("Scaling advice: %s to %d workers (confidence: %.2f)",
			advice.Action, advice.WorkersNeeded, advice.Confidence)
	})

	t.Run("ModelTraining", func(t *testing.T) {
		err := mlService.TrainModels()
		if err != nil {
			t.Errorf("Model training failed: %v", err)
		}

		t.Logf("Model training completed successfully")
	})
}

// TestMLAPIIntegration tests the ML service HTTP API
func TestMLAPIIntegration(t *testing.T) {
	// Create ML service and server
	mlServer := setupMLServer(t)

	// Start test server
	server := httptest.NewServer(mlServer)
	defer server.Close()

	// Test health endpoint
	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/health")
		if err != nil {
			t.Fatalf("Health check failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var health map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
			t.Errorf("Failed to decode health response: %v", err)
		}

		if health["status"] != "healthy" {
			t.Errorf("Expected status 'healthy', got '%s'", health["status"])
		}
	})

	// Test prediction endpoint
	t.Run("BuildPrediction", func(t *testing.T) {
		payload := map[string]interface{}{
			"project_path":  "/test/project",
			"task_name":     "build",
			"build_options": map[string]string{"clean": "true"},
		}

		jsonData, _ := json.Marshal(payload)
		resp, err := http.Post(server.URL+"/api/predict", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Prediction request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var prediction service.PredictionResult
		if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
			t.Errorf("Failed to decode prediction response: %v", err)
		}

		if prediction.BuildID == "" {
			t.Errorf("Expected non-empty build ID")
		}

		t.Logf("Received prediction: BuildID=%s, Time=%v, Confidence=%.2f",
			prediction.BuildID, prediction.PredictedTime, prediction.Confidence)
	})

	// Test scaling advice endpoint
	t.Run("ScalingAdvice", func(t *testing.T) {
		url := fmt.Sprintf("%s/api/scaling?queue_length=10&cpu_load=0.8&current_workers=2", server.URL)
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("Scaling advice request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var advice service.ScalingRecommendation
		if err := json.NewDecoder(resp.Body).Decode(&advice); err != nil {
			t.Errorf("Failed to decode scaling advice: %v", err)
		}

		t.Logf("Scaling recommendation: %s %d workers (confidence: %.2f)",
			advice.Action, advice.WorkersNeeded, advice.Confidence)
	})

	// Test statistics endpoint
	t.Run("Statistics", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/stats")
		if err != nil {
			t.Fatalf("Statistics request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var stats map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
			t.Errorf("Failed to decode statistics: %v", err)
		}

		if stats["total_build_records"] == nil {
			t.Errorf("Expected total_build_records in statistics")
		}

		t.Logf("ML service statistics: %d build records, %d worker metrics",
			int(stats["total_build_records"].(float64)),
			int(stats["total_worker_metrics"].(float64)))
	})
}

// setupMLServer creates a test ML server handler
func setupMLServer(t *testing.T) http.Handler {
	mlService := service.NewMLService()
	addTestDataToMLService(mlService)

	// Create a simple HTTP handler for testing
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	mux.HandleFunc("/api/predict", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ProjectPath  string            `json:"project_path"`
			TaskName     string            `json:"task_name"`
			BuildOptions map[string]string `json:"build_options"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		prediction := mlService.GetBuildInsights(req.ProjectPath, req.TaskName, req.BuildOptions)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(prediction)
	})

	mux.HandleFunc("/api/scaling", func(w http.ResponseWriter, r *http.Request) {
		queueLengthStr := r.URL.Query().Get("queue_length")
		cpuLoadStr := r.URL.Query().Get("cpu_load")
		workersStr := r.URL.Query().Get("current_workers")

		// Parse parameters with defaults
		queueLength := 0
		if queueLengthStr != "" {
			fmt.Sscanf(queueLengthStr, "%d", &queueLength)
		}

		cpuLoad := 0.5
		if cpuLoadStr != "" {
			fmt.Sscanf(cpuLoadStr, "%f", &cpuLoad)
		}

		currentWorkers := 1
		if workersStr != "" {
			fmt.Sscanf(workersStr, "%d", &currentWorkers)
		}

		advice := mlService.PredictScalingNeeds(queueLength, cpuLoad, currentWorkers)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(advice)
	})

	mux.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		stats := mlService.GetStatistics()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	return mux
}

// addTestDataToMLService adds test data to the ML service
func addTestDataToMLService(mlService *service.MLService) {
	testBuilds := []service.Build{
		{
			ID:           "test-build-1",
			ProjectPath:  "/test/project",
			TaskName:     "build",
			WorkerID:     "worker-1",
			StartTime:    time.Now().Add(-time.Hour),
			EndTime:      time.Now().Add(-time.Hour + 5*time.Minute),
			Success:      true,
			CacheHitRate: 0.8,
			CPUUsage:     0.6,
			MemoryUsage:  0.5,
			DiskUsage:    0.3,
		},
		{
			ID:           "test-build-2",
			ProjectPath:  "/test/project",
			TaskName:     "test",
			WorkerID:     "worker-2",
			StartTime:    time.Now().Add(-30 * time.Minute),
			EndTime:      time.Now().Add(-30*time.Minute + 3*time.Minute),
			Success:      true,
			CacheHitRate: 0.9,
			CPUUsage:     0.4,
			MemoryUsage:  0.3,
			DiskUsage:    0.2,
		},
		{
			ID:           "test-build-3",
			ProjectPath:  "/test/project",
			TaskName:     "build",
			WorkerID:     "worker-1",
			StartTime:    time.Now().Add(-15 * time.Minute),
			EndTime:      time.Now().Add(-15*time.Minute + 8*time.Minute),
			Success:      false,
			CacheHitRate: 0.3,
			CPUUsage:     0.8,
			MemoryUsage:  0.7,
			DiskUsage:    0.5,
			ErrorMessage: "Compilation failed",
		},
		// Add more test data for training
		{
			ID:           "test-build-4",
			ProjectPath:  "/test/project",
			TaskName:     "build",
			WorkerID:     "worker-1",
			StartTime:    time.Now().Add(-45 * time.Minute),
			EndTime:      time.Now().Add(-45*time.Minute + 6*time.Minute),
			Success:      true,
			CacheHitRate: 0.7,
			CPUUsage:     0.5,
			MemoryUsage:  0.4,
			DiskUsage:    0.2,
		},
		{
			ID:           "test-build-5",
			ProjectPath:  "/test/project",
			TaskName:     "test",
			WorkerID:     "worker-2",
			StartTime:    time.Now().Add(-60 * time.Minute),
			EndTime:      time.Now().Add(-60*time.Minute + 4*time.Minute),
			Success:      true,
			CacheHitRate: 0.85,
			CPUUsage:     0.45,
			MemoryUsage:  0.35,
			DiskUsage:    0.25,
		},
		{
			ID:           "test-build-6",
			ProjectPath:  "/another/project",
			TaskName:     "build",
			WorkerID:     "worker-1",
			StartTime:    time.Now().Add(-75 * time.Minute),
			EndTime:      time.Now().Add(-75*time.Minute + 7*time.Minute),
			Success:      true,
			CacheHitRate: 0.6,
			CPUUsage:     0.7,
			MemoryUsage:  0.6,
			DiskUsage:    0.4,
		},
		{
			ID:           "test-build-7",
			ProjectPath:  "/another/project",
			TaskName:     "test",
			WorkerID:     "worker-2",
			StartTime:    time.Now().Add(-90 * time.Minute),
			EndTime:      time.Now().Add(-90*time.Minute + 5*time.Minute),
			Success:      false,
			CacheHitRate: 0.4,
			CPUUsage:     0.9,
			MemoryUsage:  0.8,
			DiskUsage:    0.6,
			ErrorMessage: "Test failure",
		},
		{
			ID:           "test-build-8",
			ProjectPath:  "/test/project",
			TaskName:     "build",
			WorkerID:     "worker-1",
			StartTime:    time.Now().Add(-105 * time.Minute),
			EndTime:      time.Now().Add(-105*time.Minute + 4*time.Minute),
			Success:      true,
			CacheHitRate: 0.75,
			CPUUsage:     0.55,
			MemoryUsage:  0.45,
			DiskUsage:    0.35,
		},
		{
			ID:           "test-build-9",
			ProjectPath:  "/test/project",
			TaskName:     "test",
			WorkerID:     "worker-2",
			StartTime:    time.Now().Add(-120 * time.Minute),
			EndTime:      time.Now().Add(-120*time.Minute + 2*time.Minute),
			Success:      true,
			CacheHitRate: 0.95,
			CPUUsage:     0.35,
			MemoryUsage:  0.25,
			DiskUsage:    0.15,
		},
		{
			ID:           "test-build-10",
			ProjectPath:  "/another/project",
			TaskName:     "build",
			WorkerID:     "worker-1",
			StartTime:    time.Now().Add(-135 * time.Minute),
			EndTime:      time.Now().Add(-135*time.Minute + 9*time.Minute),
			Success:      false,
			CacheHitRate: 0.2,
			CPUUsage:     0.95,
			MemoryUsage:  0.85,
			DiskUsage:    0.7,
			ErrorMessage: "Memory error",
		},
		// Add more builds to reach 20+ records
		{
			ID:           "test-build-11",
			ProjectPath:  "/test/project",
			TaskName:     "build",
			WorkerID:     "worker-1",
			StartTime:    time.Now().Add(-150 * time.Minute),
			EndTime:      time.Now().Add(-150*time.Minute + 5*time.Minute),
			Success:      true,
			CacheHitRate: 0.8,
			CPUUsage:     0.6,
			MemoryUsage:  0.5,
			DiskUsage:    0.3,
		},
		{
			ID:           "test-build-12",
			ProjectPath:  "/test/project",
			TaskName:     "test",
			WorkerID:     "worker-2",
			StartTime:    time.Now().Add(-165 * time.Minute),
			EndTime:      time.Now().Add(-165*time.Minute + 3*time.Minute),
			Success:      true,
			CacheHitRate: 0.9,
			CPUUsage:     0.4,
			MemoryUsage:  0.3,
			DiskUsage:    0.2,
		},
		{
			ID:           "test-build-13",
			ProjectPath:  "/another/project",
			TaskName:     "build",
			WorkerID:     "worker-1",
			StartTime:    time.Now().Add(-180 * time.Minute),
			EndTime:      time.Now().Add(-180*time.Minute + 6*time.Minute),
			Success:      true,
			CacheHitRate: 0.7,
			CPUUsage:     0.65,
			MemoryUsage:  0.55,
			DiskUsage:    0.35,
		},
		{
			ID:           "test-build-14",
			ProjectPath:  "/another/project",
			TaskName:     "test",
			WorkerID:     "worker-2",
			StartTime:    time.Now().Add(-195 * time.Minute),
			EndTime:      time.Now().Add(-195*time.Minute + 4*time.Minute),
			Success:      true,
			CacheHitRate: 0.8,
			CPUUsage:     0.5,
			MemoryUsage:  0.4,
			DiskUsage:    0.3,
		},
		{
			ID:           "test-build-15",
			ProjectPath:  "/test/project",
			TaskName:     "build",
			WorkerID:     "worker-1",
			StartTime:    time.Now().Add(-210 * time.Minute),
			EndTime:      time.Now().Add(-210*time.Minute + 7*time.Minute),
			Success:      false,
			CacheHitRate: 0.3,
			CPUUsage:     0.8,
			MemoryUsage:  0.7,
			DiskUsage:    0.5,
			ErrorMessage: "Build timeout",
		},
		{
			ID:           "test-build-16",
			ProjectPath:  "/test/project",
			TaskName:     "test",
			WorkerID:     "worker-2",
			StartTime:    time.Now().Add(-225 * time.Minute),
			EndTime:      time.Now().Add(-225*time.Minute + 3*time.Minute),
			Success:      true,
			CacheHitRate: 0.85,
			CPUUsage:     0.45,
			MemoryUsage:  0.35,
			DiskUsage:    0.25,
		},
		{
			ID:           "test-build-17",
			ProjectPath:  "/another/project",
			TaskName:     "build",
			WorkerID:     "worker-1",
			StartTime:    time.Now().Add(-240 * time.Minute),
			EndTime:      time.Now().Add(-240*time.Minute + 8*time.Minute),
			Success:      true,
			CacheHitRate: 0.6,
			CPUUsage:     0.7,
			MemoryUsage:  0.6,
			DiskUsage:    0.4,
		},
		{
			ID:           "test-build-18",
			ProjectPath:  "/another/project",
			TaskName:     "test",
			WorkerID:     "worker-2",
			StartTime:    time.Now().Add(-255 * time.Minute),
			EndTime:      time.Now().Add(-255*time.Minute + 5*time.Minute),
			Success:      false,
			CacheHitRate: 0.4,
			CPUUsage:     0.9,
			MemoryUsage:  0.8,
			DiskUsage:    0.6,
			ErrorMessage: "Assertion failed",
		},
		{
			ID:           "test-build-19",
			ProjectPath:  "/test/project",
			TaskName:     "build",
			WorkerID:     "worker-1",
			StartTime:    time.Now().Add(-270 * time.Minute),
			EndTime:      time.Now().Add(-270*time.Minute + 5*time.Minute),
			Success:      true,
			CacheHitRate: 0.75,
			CPUUsage:     0.55,
			MemoryUsage:  0.45,
			DiskUsage:    0.35,
		},
		{
			ID:           "test-build-20",
			ProjectPath:  "/test/project",
			TaskName:     "test",
			WorkerID:     "worker-2",
			StartTime:    time.Now().Add(-285 * time.Minute),
			EndTime:      time.Now().Add(-285*time.Minute + 4*time.Minute),
			Success:      true,
			CacheHitRate: 0.9,
			CPUUsage:     0.4,
			MemoryUsage:  0.3,
			DiskUsage:    0.2,
		},
		{
			ID:           "test-build-21",
			ProjectPath:  "/another/project",
			TaskName:     "build",
			WorkerID:     "worker-1",
			StartTime:    time.Now().Add(-300 * time.Minute),
			EndTime:      time.Now().Add(-300*time.Minute + 6*time.Minute),
			Success:      true,
			CacheHitRate: 0.7,
			CPUUsage:     0.6,
			MemoryUsage:  0.5,
			DiskUsage:    0.3,
		},
	}

	for _, build := range testBuilds {
		mlService.RecordBuild(build)
	}

	// Add worker metrics
	testMetrics := []service.WorkerMetric{
		{
			WorkerID:     "worker-1",
			Timestamp:    time.Now(),
			CPUUsage:     0.6,
			MemoryUsage:  0.5,
			DiskUsage:    0.3,
			ActiveBuilds: 1,
			QueueLength:  0,
			ResponseTime: 100 * time.Millisecond,
		},
		{
			WorkerID:     "worker-2",
			Timestamp:    time.Now(),
			CPUUsage:     0.4,
			MemoryUsage:  0.3,
			DiskUsage:    0.2,
			ActiveBuilds: 0,
			QueueLength:  1,
			ResponseTime: 80 * time.Millisecond,
		},
	}

	for _, metric := range testMetrics {
		mlService.RecordWorkerMetrics(metric)
	}
}
