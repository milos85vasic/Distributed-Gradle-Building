package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"distributed-gradle-building/ml/service"
)

type MLServer struct {
	mlService *service.MLService
	port      int
}

func NewMLServer(port int) *MLServer {
	return &MLServer{
		mlService: service.NewMLService(),
		port:      port,
	}
}

func (s *MLServer) Start() error {
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/api/predict", s.handlePredict)
	http.HandleFunc("/api/train", s.handleTrain)
	http.HandleFunc("/api/scaling", s.handleScalingAdvice)
	http.HandleFunc("/api/stats", s.handleStats)
	http.HandleFunc("/api/export", s.handleExport)
	http.HandleFunc("/api/import", s.handleImport)

	log.Printf("ML Service starting on port %d", s.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil)
}

func (s *MLServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (s *MLServer) handlePredict(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	prediction := s.mlService.GetBuildInsights(req.ProjectPath, req.TaskName, req.BuildOptions)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prediction)
}

func (s *MLServer) handleTrain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.mlService.TrainModels(); err != nil {
		http.Error(w, fmt.Sprintf("Training failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "training_completed"})
}

func (s *MLServer) handleScalingAdvice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queueLengthStr := r.URL.Query().Get("queue_length")
	cpuLoadStr := r.URL.Query().Get("cpu_load")
	workersStr := r.URL.Query().Get("current_workers")

	queueLength, err := strconv.Atoi(queueLengthStr)
	if err != nil {
		queueLength = 0
	}

	cpuLoad, err := strconv.ParseFloat(cpuLoadStr, 64)
	if err != nil {
		cpuLoad = 0.5
	}

	currentWorkers, err := strconv.Atoi(workersStr)
	if err != nil {
		currentWorkers = 1
	}

	advice := s.mlService.PredictScalingNeeds(queueLength, cpuLoad, currentWorkers)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(advice)
}

func (s *MLServer) handleStats(w http.ResponseWriter, r *http.Request) {
	stats := s.mlService.GetStatistics()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *MLServer) handleExport(w http.ResponseWriter, r *http.Request) {
	data, err := s.mlService.ExportData()
	if err != nil {
		http.Error(w, fmt.Sprintf("Export failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=ml-data.json")
	w.Write(data)
}

func (s *MLServer) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := make([]byte, r.ContentLength)
	_, err := r.Body.Read(data)
	if err != nil && err.Error() != "EOF" {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	if err := s.mlService.ImportData(data); err != nil {
		http.Error(w, fmt.Sprintf("Import failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "import_completed"})
}

// AddTestData adds some test data for demonstration
func (s *MLServer) AddTestData() {
	// Add some sample build records
	testBuilds := []service.Build{
		{
			ID:           "build-1",
			ProjectPath:  "/projects/myapp",
			TaskName:     "build",
			WorkerID:     "worker-1",
			StartTime:    time.Now().Add(-time.Hour),
			EndTime:      time.Now().Add(-time.Hour + 5*time.Minute),
			Success:      true,
			CacheHitRate: 0.7,
			CPUUsage:     0.6,
			MemoryUsage:  0.5,
			DiskUsage:    0.3,
		},
		{
			ID:           "build-2",
			ProjectPath:  "/projects/myapp",
			TaskName:     "test",
			WorkerID:     "worker-2",
			StartTime:    time.Now().Add(-30 * time.Minute),
			EndTime:      time.Now().Add(-30*time.Minute + 3*time.Minute),
			Success:      true,
			CacheHitRate: 0.8,
			CPUUsage:     0.4,
			MemoryUsage:  0.3,
			DiskUsage:    0.2,
		},
		{
			ID:           "build-3",
			ProjectPath:  "/projects/web",
			TaskName:     "build",
			WorkerID:     "worker-1",
			StartTime:    time.Now().Add(-15 * time.Minute),
			EndTime:      time.Now().Add(-15*time.Minute + 8*time.Minute),
			Success:      false,
			CacheHitRate: 0.2,
			CPUUsage:     0.8,
			MemoryUsage:  0.7,
			DiskUsage:    0.4,
			ErrorMessage: "Compilation failed",
		},
	}

	for _, build := range testBuilds {
		s.mlService.RecordBuild(build)
	}

	// Add some worker metrics
	testMetrics := []service.WorkerMetric{
		{
			WorkerID:     "worker-1",
			Timestamp:    time.Now(),
			CPUUsage:     0.6,
			MemoryUsage:  0.5,
			DiskUsage:    0.3,
			ActiveBuilds: 2,
			QueueLength:  1,
			ResponseTime: 100 * time.Millisecond,
		},
		{
			WorkerID:     "worker-2",
			Timestamp:    time.Now(),
			CPUUsage:     0.4,
			MemoryUsage:  0.3,
			DiskUsage:    0.2,
			ActiveBuilds: 1,
			QueueLength:  0,
			ResponseTime: 80 * time.Millisecond,
		},
	}

	for _, metric := range testMetrics {
		s.mlService.RecordWorkerMetrics(metric)
	}

	// Add cache metrics
	cacheMetric := service.CacheMetric{
		Timestamp:   time.Now(),
		CacheHits:   150,
		CacheMisses: 50,
		CacheSize:   1024 * 1024 * 100, // 100MB
		HitRate:     0.75,
	}
	s.mlService.RecordCacheMetrics(cacheMetric)
}

func main() {
	port := 8082 // Different port from coordinator (8080) and monitor (8081)

	server := NewMLServer(port)

	// Add test data for demonstration
	server.AddTestData()

	log.Printf("Starting ML Service on port %d", port)
	log.Printf("Endpoints:")
	log.Printf("  GET  /health - Health check")
	log.Printf("  POST /api/predict - Get build predictions")
	log.Printf("  POST /api/train - Train ML models")
	log.Printf("  GET  /api/scaling - Get scaling advice")
	log.Printf("  GET  /api/stats - Get ML service statistics")
	log.Printf("  GET  /api/export - Export ML data")
	log.Printf("  POST /api/import - Import ML data")

	if err := server.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
