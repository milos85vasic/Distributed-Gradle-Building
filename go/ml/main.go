package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"distributed-gradle-building/ml/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MLServer struct {
	mlService  *service.MLService
	port       int
	httpServer *http.Server
	shutdown   chan struct{}

	// Prometheus metrics
	predictionsTotal    prometheus.Counter
	predictionsDuration prometheus.Histogram
	trainingTotal       prometheus.Counter
}

func NewMLServer(port int) *MLServer {
	// Create Prometheus metrics
	predictionsTotal := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ml_predictions_total",
			Help: "Total number of predictions made by ML service",
		},
	)

	predictionsDuration := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "ml_prediction_duration_seconds",
			Help:    "Duration of prediction requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	trainingTotal := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "ml_training_total",
			Help: "Total number of model training sessions",
		},
	)

	// Register metrics
	prometheus.MustRegister(predictionsTotal, predictionsDuration, trainingTotal)

	return &MLServer{
		mlService:           service.NewMLService(),
		port:                port,
		shutdown:            make(chan struct{}),
		predictionsTotal:    predictionsTotal,
		predictionsDuration: predictionsDuration,
		trainingTotal:       trainingTotal,
	}
}

func (s *MLServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/predict", s.handlePredict)
	mux.HandleFunc("/api/train", s.handleTrain)
	mux.HandleFunc("/api/scaling", s.handleScalingAdvice)
	mux.HandleFunc("/api/stats", s.handleStats)
	mux.HandleFunc("/api/learning", s.handleLearningStats)
	mux.HandleFunc("/api/rollback", s.handleRollback)
	mux.HandleFunc("/api/export", s.handleExport)
	mux.HandleFunc("/api/import", s.handleImport)
	mux.Handle("/metrics", promhttp.Handler())

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	log.Printf("ML Service starting on port %d", s.port)
	return s.httpServer.ListenAndServe()
}

func (s *MLServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stats := s.mlService.GetLearningStats()
	health := map[string]any{
		"status":    "healthy",
		"service":   "ml",
		"version":   "1.0.0",
		"timestamp": time.Now().Format(time.RFC3339),
		"stats":     stats,
	}

	json.NewEncoder(w).Encode(health)
}

func (s *MLServer) handlePredict(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

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

	// Record metrics
	s.predictionsTotal.Inc()
	s.predictionsDuration.Observe(time.Since(start).Seconds())

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

	// Record training metric
	s.trainingTotal.Inc()

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

func (s *MLServer) handleLearningStats(w http.ResponseWriter, r *http.Request) {
	response := s.mlService.GetLearningStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *MLServer) handleRollback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Version string `json:"version"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Version == "" {
		http.Error(w, "Version is required", http.StatusBadRequest)
		return
	}

	if err := s.mlService.RollbackToVersion(req.Version); err != nil {
		http.Error(w, fmt.Sprintf("Rollback failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "rollback_completed", "version": req.Version})
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
	port := 8085
	if len(os.Args) > 1 {
		if p, err := strconv.Atoi(os.Args[1]); err == nil {
			port = p
		}
	}

	server := NewMLServer(port)

	log.Printf("Distributed Gradle Building - ML Service")
	log.Printf("Version: 1.0.0")
	log.Printf("Listening on port %d", port)
	log.Printf("Available endpoints:")
	log.Printf("  GET  /health - Health check")
	log.Printf("  POST /api/predict - Predict build time and resources")
	log.Printf("  POST /api/train - Train ML models")
	log.Printf("  GET  /api/scaling - Get scaling advice")
	log.Printf("  GET  /api/stats - Get service statistics")
	log.Printf("  GET  /api/learning - Get learning statistics")
	log.Printf("  POST /api/rollback - Rollback to previous model")
	log.Printf("  GET  /api/export - Export ML data")
	log.Printf("  POST /api/import - Import ML data")
	log.Printf("Continuous learning: ENABLED")

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal %v, shutting down ML service...", sig)

	// Graceful shutdown
	if err := server.Shutdown(); err != nil {
		log.Printf("Error during ML service shutdown: %v", err)
	}

	log.Printf("ML service shutdown complete")
}

// Shutdown gracefully shuts down the ML server
func (s *MLServer) Shutdown() error {
	close(s.shutdown)

	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(ctx)
	}

	return nil
}
