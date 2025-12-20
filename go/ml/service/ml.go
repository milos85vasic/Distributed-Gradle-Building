package service

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

// ContinuousLearningConfig contains configuration for automated ML retraining
type ContinuousLearningConfig struct {
	Enabled                bool          `json:"enabled"`
	RetrainingInterval     time.Duration `json:"retraining_interval"`
	DataCollectionInterval time.Duration `json:"data_collection_interval"`
	MinDataPoints          int           `json:"min_data_points"`
	PerformanceThreshold   float64       `json:"performance_threshold"`
	CoordinatorHost        string        `json:"coordinator_host"`
	CoordinatorPort        int           `json:"coordinator_port"`
	MonitorHost            string        `json:"monitor_host"`
	MonitorPort            int           `json:"monitor_port"`
}

// ModelBackup represents a backup of ML models for rollback
type ModelBackup struct {
	Timestamp time.Time `json:"timestamp"`
	Models    MLModels  `json:"models"`
	Accuracy  float64   `json:"accuracy"`
	Version   string    `json:"version"`
}

// ContinuousLearningStats tracks continuous learning performance
type ContinuousLearningStats struct {
	LastDataCollection  time.Time     `json:"last_data_collection"`
	LastRetraining      time.Time     `json:"last_retraining"`
	DataPointsCollected int           `json:"data_points_collected"`
	RetrainingCount     int           `json:"retraining_count"`
	AverageAccuracy     float64       `json:"average_accuracy"`
	LastAccuracyCheck   time.Time     `json:"last_accuracy_check"`
	IsCollecting        bool          `json:"is_collecting"`
	IsRetraining        bool          `json:"is_retraining"`
	ModelBackups        []ModelBackup `json:"model_backups"`
	CurrentVersion      string        `json:"current_version"`
}

// MLService provides machine learning capabilities for build optimization
type MLService struct {
	BuildHistory       []BuildRecord            `json:"build_history"`
	WorkerMetrics      []WorkerMetric           `json:"worker_metrics"`
	CacheMetrics       []CacheMetric            `json:"cache_metrics"`
	Models             MLModels                 `json:"models"`
	ContinuousLearning ContinuousLearningConfig `json:"continuous_learning"`
	LearningStats      ContinuousLearningStats  `json:"learning_stats"`
	mutex              sync.RWMutex
	shutdown           chan struct{}
}

// BuildRecord represents a historical build record for ML training
type BuildRecord struct {
	BuildID      string            `json:"build_id"`
	ProjectPath  string            `json:"project_path"`
	TaskName     string            `json:"task_name"`
	WorkerID     string            `json:"worker_id"`
	StartTime    time.Time         `json:"start_time"`
	EndTime      time.Time         `json:"end_time"`
	Duration     time.Duration     `json:"duration"`
	Success      bool              `json:"success"`
	CacheHitRate float64           `json:"cache_hit_rate"`
	CPUUsage     float64           `json:"cpu_usage"`
	MemoryUsage  float64           `json:"memory_usage"`
	DiskUsage    float64           `json:"disk_usage"`
	BuildOptions map[string]string `json:"build_options"`
	ErrorMessage string            `json:"error_message,omitempty"`
}

// WorkerMetric represents worker performance metrics
type WorkerMetric struct {
	WorkerID     string        `json:"worker_id"`
	Timestamp    time.Time     `json:"timestamp"`
	CPUUsage     float64       `json:"cpu_usage"`
	MemoryUsage  float64       `json:"memory_usage"`
	DiskUsage    float64       `json:"disk_usage"`
	ActiveBuilds int           `json:"active_builds"`
	QueueLength  int           `json:"queue_length"`
	ResponseTime time.Duration `json:"response_time"`
}

// CacheMetric represents cache performance metrics
type CacheMetric struct {
	Timestamp   time.Time `json:"timestamp"`
	CacheHits   int64     `json:"cache_hits"`
	CacheMisses int64     `json:"cache_misses"`
	CacheSize   int64     `json:"cache_size"`
	HitRate     float64   `json:"hit_rate"`
}

// MLModels contains trained machine learning models
type MLModels struct {
	BuildTimePredictor BuildTimeModel `json:"build_time_predictor"`
	ResourcePredictor  ResourceModel  `json:"resource_predictor"`
	ScalingPredictor   ScalingModel   `json:"scaling_predictor"`
	FailurePredictor   FailureModel   `json:"failure_predictor"`
	CachePredictor     CacheModel     `json:"cache_predictor"`
}

// BuildTimeModel predicts build duration
type BuildTimeModel struct {
	// Simple linear regression model for build time prediction
	Weights     map[string]float64 `json:"weights"`
	Bias        float64            `json:"bias"`
	Features    []string           `json:"features"`
	Accuracy    float64            `json:"accuracy"`
	LastTrained time.Time          `json:"last_trained"`
}

// ResourceModel predicts resource requirements
type ResourceModel struct {
	CPUWeights  map[string]float64 `json:"cpu_weights"`
	MemWeights  map[string]float64 `json:"mem_weights"`
	DiskWeights map[string]float64 `json:"disk_weights"`
	Bias        float64            `json:"bias"`
	Accuracy    float64            `json:"accuracy"`
	LastTrained time.Time          `json:"last_trained"`
}

// ScalingModel predicts scaling needs
type ScalingModel struct {
	QueueThreshold   float64          `json:"queue_threshold"`
	CPULoadThreshold float64          `json:"cpu_load_threshold"`
	Patterns         []ScalingPattern `json:"patterns"`
	LastTrained      time.Time        `json:"last_trained"`
}

// ScalingPattern represents a scaling pattern
type ScalingPattern struct {
	HourOfDay          int     `json:"hour_of_day"`
	DayOfWeek          int     `json:"day_of_week"`
	ExpectedLoad       float64 `json:"expected_load"`
	RecommendedWorkers int     `json:"recommended_workers"`
}

// FailureModel predicts build failures
type FailureModel struct {
	ErrorPatterns map[string]float64 `json:"error_patterns"`
	RiskFactors   []string           `json:"risk_factors"`
	Accuracy      float64            `json:"accuracy"`
	LastTrained   time.Time          `json:"last_trained"`
}

// CacheModel predicts cache performance
type CacheModel struct {
	HitRateWeights map[string]float64 `json:"hit_rate_weights"`
	Bias           float64            `json:"bias"`
	Accuracy       float64            `json:"accuracy"`
	LastTrained    time.Time          `json:"last_trained"`
}

// PredictionResult represents a prediction result
type PredictionResult struct {
	BuildID       string                `json:"build_id"`
	PredictedTime time.Duration         `json:"predicted_time"`
	Confidence    float64               `json:"confidence"`
	ResourceNeeds ResourcePrediction    `json:"resource_needs"`
	ScalingAdvice ScalingRecommendation `json:"scaling_advice"`
	FailureRisk   float64               `json:"failure_risk"`
	CacheHitRate  float64               `json:"cache_hit_rate"`
}

// ResourcePrediction predicts resource requirements
type ResourcePrediction struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"memory"`
	Disk   float64 `json:"disk"`
}

// ScalingRecommendation suggests scaling actions
type ScalingRecommendation struct {
	Action        string  `json:"action"` // "scale_up", "scale_down", "maintain"
	WorkersNeeded int     `json:"workers_needed"`
	Confidence    float64 `json:"confidence"`
	Reason        string  `json:"reason"`
}

// NewMLService creates a new ML service instance
func NewMLService() *MLService {
	service := &MLService{
		BuildHistory:  make([]BuildRecord, 0),
		WorkerMetrics: make([]WorkerMetric, 0),
		CacheMetrics:  make([]CacheMetric, 0),
		Models: MLModels{
			BuildTimePredictor: BuildTimeModel{
				Weights: make(map[string]float64),
				Bias:    300.0, // Default 5 minutes
			},
			ResourcePredictor: ResourceModel{
				CPUWeights:  make(map[string]float64),
				MemWeights:  make(map[string]float64),
				DiskWeights: make(map[string]float64),
				Bias:        0.5,
			},
			ScalingPredictor: ScalingModel{
				QueueThreshold:   5.0,
				CPULoadThreshold: 0.8,
				Patterns:         make([]ScalingPattern, 0),
			},
			FailurePredictor: FailureModel{
				ErrorPatterns: make(map[string]float64),
				RiskFactors:   make([]string, 0),
			},
			CachePredictor: CacheModel{
				HitRateWeights: make(map[string]float64),
				Bias:           0.7, // Default 70% hit rate
			},
		},
		ContinuousLearning: ContinuousLearningConfig{
			Enabled:                true,
			RetrainingInterval:     24 * time.Hour,  // Retrain daily
			DataCollectionInterval: 5 * time.Minute, // Collect data every 5 minutes
			MinDataPoints:          50,              // Need at least 50 data points
			PerformanceThreshold:   0.7,             // Retrain if accuracy drops below 70%
			CoordinatorHost:        "coordinator",
			CoordinatorPort:        8080,
			MonitorHost:            "monitor",
			MonitorPort:            8084,
		},
		LearningStats: ContinuousLearningStats{
			LastDataCollection: time.Now(),
			LastRetraining:     time.Now(),
			CurrentVersion:     "v1.0",
		},
		shutdown: make(chan struct{}),
	}

	// Load configuration from environment
	service.loadContinuousLearningConfig()

	return service
}

// loadContinuousLearningConfig loads continuous learning configuration from environment variables
func (ml *MLService) loadContinuousLearningConfig() {
	// Environment variable overrides
	if enabled := getEnvAsBool("ML_CONTINUOUS_LEARNING_ENABLED", ml.ContinuousLearning.Enabled); enabled != ml.ContinuousLearning.Enabled {
		ml.ContinuousLearning.Enabled = enabled
	}

	if interval := getEnvAsDuration("ML_RETRAINING_INTERVAL", ml.ContinuousLearning.RetrainingInterval); interval != ml.ContinuousLearning.RetrainingInterval {
		ml.ContinuousLearning.RetrainingInterval = interval
	}

	if interval := getEnvAsDuration("ML_DATA_COLLECTION_INTERVAL", ml.ContinuousLearning.DataCollectionInterval); interval != ml.ContinuousLearning.DataCollectionInterval {
		ml.ContinuousLearning.DataCollectionInterval = interval
	}

	if points := getEnvAsInt("ML_MIN_DATA_POINTS", ml.ContinuousLearning.MinDataPoints); points != ml.ContinuousLearning.MinDataPoints {
		ml.ContinuousLearning.MinDataPoints = points
	}

	if threshold := getEnvAsFloat("ML_PERFORMANCE_THRESHOLD", ml.ContinuousLearning.PerformanceThreshold); threshold != ml.ContinuousLearning.PerformanceThreshold {
		ml.ContinuousLearning.PerformanceThreshold = threshold
	}

	if host := getEnvString("ML_COORDINATOR_HOST", ml.ContinuousLearning.CoordinatorHost); host != ml.ContinuousLearning.CoordinatorHost {
		ml.ContinuousLearning.CoordinatorHost = host
	}

	if port := getEnvAsInt("ML_COORDINATOR_PORT", ml.ContinuousLearning.CoordinatorPort); port != ml.ContinuousLearning.CoordinatorPort {
		ml.ContinuousLearning.CoordinatorPort = port
	}

	if host := getEnvString("ML_MONITOR_HOST", ml.ContinuousLearning.MonitorHost); host != ml.ContinuousLearning.MonitorHost {
		ml.ContinuousLearning.MonitorHost = host
	}

	if port := getEnvAsInt("ML_MONITOR_PORT", ml.ContinuousLearning.MonitorPort); port != ml.ContinuousLearning.MonitorPort {
		ml.ContinuousLearning.MonitorPort = port
	}
}

// Build represents a completed build for ML analysis
type Build struct {
	ID           string
	ProjectPath  string
	TaskName     string
	WorkerID     string
	StartTime    time.Time
	EndTime      time.Time
	Success      bool
	CacheHitRate float64
	CPUUsage     float64
	MemoryUsage  float64
	DiskUsage    float64
	BuildOptions map[string]string
	ErrorMessage string
}

// RecordBuild adds a build record to the history
func (ml *MLService) RecordBuild(build Build) {
	ml.mutex.Lock()
	defer ml.mutex.Unlock()

	record := BuildRecord{
		BuildID:      build.ID,
		ProjectPath:  build.ProjectPath,
		TaskName:     build.TaskName,
		WorkerID:     build.WorkerID,
		StartTime:    build.StartTime,
		EndTime:      build.EndTime,
		Duration:     build.EndTime.Sub(build.StartTime),
		Success:      build.Success,
		CacheHitRate: build.CacheHitRate,
		CPUUsage:     build.CPUUsage,
		MemoryUsage:  build.MemoryUsage,
		DiskUsage:    build.DiskUsage,
		BuildOptions: build.BuildOptions,
		ErrorMessage: build.ErrorMessage,
	}

	// Keep only last 10000 records to prevent memory issues
	if len(ml.BuildHistory) >= 10000 {
		ml.BuildHistory = ml.BuildHistory[1:]
	}
	ml.BuildHistory = append(ml.BuildHistory, record)
}

// RecordWorkerMetrics adds worker performance metrics
func (ml *MLService) RecordWorkerMetrics(metrics WorkerMetric) {
	ml.mutex.Lock()
	defer ml.mutex.Unlock()

	// Keep only last 5000 metrics
	if len(ml.WorkerMetrics) >= 5000 {
		ml.WorkerMetrics = ml.WorkerMetrics[1:]
	}
	ml.WorkerMetrics = append(ml.WorkerMetrics, metrics)
}

// RecordCacheMetrics adds cache performance metrics
func (ml *MLService) RecordCacheMetrics(metrics CacheMetric) {
	ml.mutex.Lock()
	defer ml.mutex.Unlock()

	// Keep only last 5000 metrics
	if len(ml.CacheMetrics) >= 5000 {
		ml.CacheMetrics = ml.CacheMetrics[1:]
	}
	ml.CacheMetrics = append(ml.CacheMetrics, metrics)
}

// PredictBuildTime predicts the duration of a build
func (ml *MLService) PredictBuildTime(projectPath, taskName string, buildOptions map[string]string) (time.Duration, float64) {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()

	if len(ml.BuildHistory) < 10 {
		// Not enough data, return default
		return 5 * time.Minute, 0.5
	}

	// Simple prediction based on historical averages for similar builds
	var totalDuration time.Duration
	var count int
	var similarBuilds []BuildRecord

	for _, record := range ml.BuildHistory {
		if record.ProjectPath == projectPath && record.TaskName == taskName && record.Success {
			similarBuilds = append(similarBuilds, record)
			totalDuration += record.Duration
			count++
		}
	}

	if count == 0 {
		// No similar builds, use overall average
		for _, record := range ml.BuildHistory {
			if record.Success {
				totalDuration += record.Duration
				count++
			}
		}
	}

	if count == 0 {
		return 5 * time.Minute, 0.3
	}

	averageDuration := totalDuration / time.Duration(count)
	confidence := math.Min(0.9, float64(count)/100.0) // Higher confidence with more data

	return averageDuration, confidence
}

// PredictResourceNeeds predicts resource requirements for a build
func (ml *MLService) PredictResourceNeeds(projectPath, taskName string) ResourcePrediction {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()

	if len(ml.BuildHistory) < 5 {
		return ResourcePrediction{CPU: 0.5, Memory: 0.5, Disk: 0.3}
	}

	var totalCPU, totalMem, totalDisk float64
	var count int

	for _, record := range ml.BuildHistory {
		if record.ProjectPath == projectPath && record.TaskName == taskName && record.Success {
			totalCPU += record.CPUUsage
			totalMem += record.MemoryUsage
			totalDisk += record.DiskUsage
			count++
		}
	}

	if count == 0 {
		// Use overall averages
		for _, record := range ml.BuildHistory {
			if record.Success {
				totalCPU += record.CPUUsage
				totalMem += record.MemoryUsage
				totalDisk += record.DiskUsage
				count++
			}
		}
	}

	if count == 0 {
		return ResourcePrediction{CPU: 0.5, Memory: 0.5, Disk: 0.3}
	}

	return ResourcePrediction{
		CPU:    totalCPU / float64(count),
		Memory: totalMem / float64(count),
		Disk:   totalDisk / float64(count),
	}
}

// PredictScalingNeeds predicts if scaling is needed
func (ml *MLService) PredictScalingNeeds(queueLength int, avgCPULoad float64, currentWorkers int) ScalingRecommendation {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()

	// Simple scaling logic based on queue and CPU load
	if queueLength > 10 || avgCPULoad > 0.9 {
		neededWorkers := int(math.Ceil(float64(queueLength)/3.0 + float64(currentWorkers)))
		if neededWorkers > currentWorkers {
			return ScalingRecommendation{
				Action:        "scale_up",
				WorkersNeeded: neededWorkers,
				Confidence:    0.8,
				Reason:        fmt.Sprintf("High queue (%d) or CPU load (%.2f)", queueLength, avgCPULoad),
			}
		}
	} else if queueLength < 2 && avgCPULoad < 0.3 && currentWorkers > 1 {
		neededWorkers := int(math.Max(1, float64(currentWorkers)-1))
		return ScalingRecommendation{
			Action:        "scale_down",
			WorkersNeeded: neededWorkers,
			Confidence:    0.6,
			Reason:        fmt.Sprintf("Low queue (%d) and CPU load (%.2f)", queueLength, avgCPULoad),
		}
	}

	return ScalingRecommendation{
		Action:        "maintain",
		WorkersNeeded: currentWorkers,
		Confidence:    0.9,
		Reason:        "System operating within normal parameters",
	}
}

// PredictFailureRisk predicts the risk of build failure
func (ml *MLService) PredictFailureRisk(projectPath, taskName string) float64 {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()

	if len(ml.BuildHistory) < 5 {
		return 0.1 // Default low risk
	}

	var failures, total int
	for _, record := range ml.BuildHistory {
		if record.ProjectPath == projectPath && record.TaskName == taskName {
			total++
			if !record.Success {
				failures++
			}
		}
	}

	if total == 0 {
		return 0.1
	}

	failureRate := float64(failures) / float64(total)

	// Adjust based on recent failures (last 10 builds)
	recentFailures := 0
	recentTotal := 0
	recentRecords := ml.BuildHistory[len(ml.BuildHistory)-int(math.Min(10, float64(len(ml.BuildHistory)))):]

	for _, record := range recentRecords {
		if record.ProjectPath == projectPath && record.TaskName == taskName {
			recentTotal++
			if !record.Success {
				recentFailures++
			}
		}
	}

	if recentTotal > 0 {
		recentFailureRate := float64(recentFailures) / float64(recentTotal)
		// Weight recent failures more heavily
		failureRate = (failureRate*0.7 + recentFailureRate*0.3)
	}

	return math.Min(failureRate, 0.95) // Cap at 95%
}

// PredictCacheHitRate predicts cache hit rate for a build
func (ml *MLService) PredictCacheHitRate(projectPath, taskName string) float64 {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()

	if len(ml.BuildHistory) < 5 {
		return 0.7 // Default hit rate
	}

	var totalHitRate float64
	var count int

	for _, record := range ml.BuildHistory {
		if record.ProjectPath == projectPath && record.TaskName == taskName {
			totalHitRate += record.CacheHitRate
			count++
		}
	}

	if count == 0 {
		// Use overall average
		for _, record := range ml.BuildHistory {
			totalHitRate += record.CacheHitRate
			count++
		}
	}

	if count == 0 {
		return 0.7
	}

	return totalHitRate / float64(count)
}

// GetBuildInsights provides comprehensive build insights
func (ml *MLService) GetBuildInsights(projectPath, taskName string, buildOptions map[string]string) PredictionResult {
	predictedTime, timeConfidence := ml.PredictBuildTime(projectPath, taskName, buildOptions)
	resourceNeeds := ml.PredictResourceNeeds(projectPath, taskName)
	failureRisk := ml.PredictFailureRisk(projectPath, taskName)
	cacheHitRate := ml.PredictCacheHitRate(projectPath, taskName)

	// Generate a build ID for this prediction
	buildID := fmt.Sprintf("prediction_%d", time.Now().Unix())

	return PredictionResult{
		BuildID:       buildID,
		PredictedTime: predictedTime,
		Confidence:    timeConfidence,
		ResourceNeeds: resourceNeeds,
		FailureRisk:   failureRisk,
		CacheHitRate:  cacheHitRate,
		ScalingAdvice: ScalingRecommendation{
			Action:        "maintain",
			WorkersNeeded: 1,
			Confidence:    0.5,
			Reason:        "Prediction for individual build",
		},
	}
}

// TrainModels trains all ML models with current data
func (ml *MLService) TrainModels() error {
	ml.mutex.Lock()
	defer ml.mutex.Unlock()

	if len(ml.BuildHistory) < 20 {
		return fmt.Errorf("insufficient data for training: need at least 20 build records, have %d", len(ml.BuildHistory))
	}

	// Train build time predictor
	ml.trainBuildTimeModel()

	// Train resource predictor
	ml.trainResourceModel()

	// Train failure predictor
	ml.trainFailureModel()

	// Train cache predictor
	ml.trainCacheModel()

	// Train scaling patterns
	ml.trainScalingPatterns()

	return nil
}

// trainBuildTimeModel trains the build time prediction model
func (ml *MLService) trainBuildTimeModel() {
	// Simple feature engineering and linear regression
	features := make(map[string][]float64)
	targets := make([]float64, 0)

	for _, record := range ml.BuildHistory {
		if !record.Success {
			continue
		}

		featureKey := record.ProjectPath + ":" + record.TaskName
		durationSeconds := record.Duration.Seconds()

		if features[featureKey] == nil {
			features[featureKey] = make([]float64, 0)
		}

		features[featureKey] = append(features[featureKey], durationSeconds)
		targets = append(targets, durationSeconds)
	}

	// Calculate simple averages for each feature combination
	for feature, values := range features {
		if len(values) > 0 {
			sum := 0.0
			for _, v := range values {
				sum += v
			}
			ml.Models.BuildTimePredictor.Weights[feature] = sum / float64(len(values))
		}
	}

	ml.Models.BuildTimePredictor.LastTrained = time.Now()
	ml.Models.BuildTimePredictor.Accuracy = 0.75 // Placeholder accuracy
}

// trainResourceModel trains the resource prediction model
func (ml *MLService) trainResourceModel() {
	projectResources := make(map[string][]float64) // CPU
	projectMemory := make(map[string][]float64)    // Memory
	projectDisk := make(map[string][]float64)      // Disk

	for _, record := range ml.BuildHistory {
		if !record.Success {
			continue
		}

		key := record.ProjectPath + ":" + record.TaskName

		if projectResources[key] == nil {
			projectResources[key] = make([]float64, 0)
			projectMemory[key] = make([]float64, 0)
			projectDisk[key] = make([]float64, 0)
		}

		projectResources[key] = append(projectResources[key], record.CPUUsage)
		projectMemory[key] = append(projectMemory[key], record.MemoryUsage)
		projectDisk[key] = append(projectDisk[key], record.DiskUsage)
	}

	// Calculate averages
	for key, values := range projectResources {
		if len(values) > 0 {
			sum := 0.0
			for _, v := range values {
				sum += v
			}
			ml.Models.ResourcePredictor.CPUWeights[key] = sum / float64(len(values))
		}
	}

	for key, values := range projectMemory {
		if len(values) > 0 {
			sum := 0.0
			for _, v := range values {
				sum += v
			}
			ml.Models.ResourcePredictor.MemWeights[key] = sum / float64(len(values))
		}
	}

	for key, values := range projectDisk {
		if len(values) > 0 {
			sum := 0.0
			for _, v := range values {
				sum += v
			}
			ml.Models.ResourcePredictor.DiskWeights[key] = sum / float64(len(values))
		}
	}

	ml.Models.ResourcePredictor.LastTrained = time.Now()
	ml.Models.ResourcePredictor.Accuracy = 0.8
}

// trainFailureModel trains the failure prediction model
func (ml *MLService) trainFailureModel() {
	errorCounts := make(map[string]int)
	totalCounts := make(map[string]int)

	for _, record := range ml.BuildHistory {
		key := record.ProjectPath + ":" + record.TaskName
		totalCounts[key]++

		if !record.Success && record.ErrorMessage != "" {
			errorCounts[record.ErrorMessage]++
		}
	}

	// Calculate failure rates
	for errorMsg, count := range errorCounts {
		ml.Models.FailurePredictor.ErrorPatterns[errorMsg] = float64(count) / float64(len(ml.BuildHistory))
	}

	ml.Models.FailurePredictor.LastTrained = time.Now()
	ml.Models.FailurePredictor.Accuracy = 0.7
}

// trainCacheModel trains the cache prediction model
func (ml *MLService) trainCacheModel() {
	projectCacheRates := make(map[string][]float64)

	for _, record := range ml.BuildHistory {
		key := record.ProjectPath + ":" + record.TaskName

		if projectCacheRates[key] == nil {
			projectCacheRates[key] = make([]float64, 0)
		}

		projectCacheRates[key] = append(projectCacheRates[key], record.CacheHitRate)
	}

	// Calculate averages
	for key, rates := range projectCacheRates {
		if len(rates) > 0 {
			sum := 0.0
			for _, rate := range rates {
				sum += rate
			}
			ml.Models.CachePredictor.HitRateWeights[key] = sum / float64(len(rates))
		}
	}

	ml.Models.CachePredictor.LastTrained = time.Now()
	ml.Models.CachePredictor.Accuracy = 0.75
}

// trainScalingPatterns trains scaling patterns based on historical data
func (ml *MLService) trainScalingPatterns() {
	hourlyPatterns := make(map[string][]int) // "hour:weekday" -> worker counts

	for _, metric := range ml.WorkerMetrics {
		key := fmt.Sprintf("%d:%d", metric.Timestamp.Hour(), int(metric.Timestamp.Weekday()))
		if hourlyPatterns[key] == nil {
			hourlyPatterns[key] = make([]int, 0)
		}
		hourlyPatterns[key] = append(hourlyPatterns[key], metric.ActiveBuilds)
	}

	// Create scaling patterns
	patterns := make([]ScalingPattern, 0)
	for key, counts := range hourlyPatterns {
		if len(counts) > 0 {
			sum := 0
			for _, count := range counts {
				sum += count
			}
			avgLoad := float64(sum) / float64(len(counts))

			var hour, weekday int
			fmt.Sscanf(key, "%d:%d", &hour, &weekday)

			// Recommend workers based on load
			recommendedWorkers := int(math.Ceil(avgLoad / 2.0))
			if recommendedWorkers < 1 {
				recommendedWorkers = 1
			}

			patterns = append(patterns, ScalingPattern{
				HourOfDay:          hour,
				DayOfWeek:          weekday,
				ExpectedLoad:       avgLoad,
				RecommendedWorkers: recommendedWorkers,
			})
		}
	}

	// Sort patterns by expected load (descending)
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].ExpectedLoad > patterns[j].ExpectedLoad
	})

	ml.Models.ScalingPredictor.Patterns = patterns
	ml.Models.ScalingPredictor.LastTrained = time.Now()
}

// GetStatistics returns ML service statistics
func (ml *MLService) GetStatistics() map[string]any {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()

	stats := make(map[string]any)

	stats["total_build_records"] = len(ml.BuildHistory)
	stats["total_worker_metrics"] = len(ml.WorkerMetrics)
	stats["total_cache_metrics"] = len(ml.CacheMetrics)

	if len(ml.BuildHistory) > 0 {
		stats["oldest_build"] = ml.BuildHistory[0].StartTime
		stats["newest_build"] = ml.BuildHistory[len(ml.BuildHistory)-1].StartTime
	}

	stats["models_trained"] = map[string]any{
		"build_time": !ml.Models.BuildTimePredictor.LastTrained.IsZero(),
		"resource":   !ml.Models.ResourcePredictor.LastTrained.IsZero(),
		"scaling":    !ml.Models.ScalingPredictor.LastTrained.IsZero(),
		"failure":    !ml.Models.FailurePredictor.LastTrained.IsZero(),
		"cache":      !ml.Models.CachePredictor.LastTrained.IsZero(),
	}

	return stats
}

// GetLearningStats returns continuous learning statistics
func (ml *MLService) GetLearningStats() map[string]any {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()

	return map[string]any{
		"config": ml.ContinuousLearning,
		"stats":  ml.LearningStats,
	}
}

// RollbackToVersion rolls back to a specific model version
func (ml *MLService) RollbackToVersion(version string) error {
	ml.mutex.Lock()
	defer ml.mutex.Unlock()

	return ml.rollbackToBackup(version)
}

// Helper functions for environment variable parsing
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if value == "true" || value == "1" || value == "yes" {
			return true
		}
		if value == "false" || value == "0" || value == "no" {
			return false
		}
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// ExportData exports all ML data for backup/analysis
func (ml *MLService) ExportData() ([]byte, error) {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()

	return json.MarshalIndent(ml, "", "  ")
}

// ImportData imports ML data from backup
func (ml *MLService) ImportData(data []byte) error {
	ml.mutex.Lock()
	defer ml.mutex.Unlock()

	return json.Unmarshal(data, ml)
}

// StartContinuousLearning starts the automated learning process
func (ml *MLService) StartContinuousLearning() {
	if !ml.ContinuousLearning.Enabled {
		log.Println("Continuous learning is disabled")
		return
	}

	log.Println("Starting continuous learning process")

	// Start data collection goroutine
	go ml.dataCollectionLoop()

	// Start retraining trigger goroutine
	go ml.retrainingTriggerLoop()
}

// StopContinuousLearning stops the automated learning process
func (ml *MLService) StopContinuousLearning() {
	close(ml.shutdown)
}

// dataCollectionLoop continuously collects data from coordinator and monitor
func (ml *MLService) dataCollectionLoop() {
	ticker := time.NewTicker(ml.ContinuousLearning.DataCollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ml.collectDataFromServices()
		case <-ml.shutdown:
			return
		}
	}
}

// retrainingTriggerLoop checks for retraining conditions
func (ml *MLService) retrainingTriggerLoop() {
	ticker := time.NewTicker(1 * time.Hour) // Check every hour
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ml.checkRetrainingConditions()
		case <-ml.shutdown:
			return
		}
	}
}

// collectDataFromServices collects build and worker data from coordinator and monitor
func (ml *MLService) collectDataFromServices() {
	ml.LearningStats.IsCollecting = true
	defer func() { ml.LearningStats.IsCollecting = false }()

	// Collect data from monitor service
	ml.collectFromMonitor()

	// Collect data from coordinator
	ml.collectFromCoordinator()

	ml.LearningStats.LastDataCollection = time.Now()
	ml.LearningStats.DataPointsCollected = len(ml.BuildHistory) + len(ml.WorkerMetrics) + len(ml.CacheMetrics)
}

// collectFromMonitor collects data from the monitor service
func (ml *MLService) collectFromMonitor() {
	monitorURL := fmt.Sprintf("http://%s:%d/api/metrics", ml.ContinuousLearning.MonitorHost, ml.ContinuousLearning.MonitorPort)

	resp, err := http.Get(monitorURL)
	if err != nil {
		log.Printf("Failed to collect data from monitor: %v", err)
		return
	}
	defer resp.Body.Close()

	var monitorData struct {
		Workers map[string]any `json:"workers"`
		Builds  map[string]any `json:"builds"`
		System  map[string]any `json:"system"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&monitorData); err != nil {
		log.Printf("Failed to decode monitor data: %v", err)
		return
	}

	// Process worker metrics
	for workerID, workerData := range monitorData.Workers {
		if workerMap, ok := workerData.(map[string]any); ok {
			metric := WorkerMetric{
				WorkerID:  workerID,
				Timestamp: time.Now(),
			}

			if status, ok := workerMap["status"].(string); ok {
				// Convert status to metrics
				if status == "active" {
					metric.ActiveBuilds = 1
				}
			}

			if cpu, ok := workerMap["cpu_usage"].(float64); ok {
				metric.CPUUsage = cpu
			}

			if mem, ok := workerMap["memory_usage"].(float64); ok {
				metric.MemoryUsage = mem
			}

			ml.RecordWorkerMetrics(metric)
		}
	}

	// Process build data
	for buildID, buildData := range monitorData.Builds {
		if buildMap, ok := buildData.(map[string]any); ok {
			build := Build{
				ID: buildID,
			}

			if projectPath, ok := buildMap["project_path"].(string); ok {
				build.ProjectPath = projectPath
			}

			if taskName, ok := buildMap["task_name"].(string); ok {
				build.TaskName = taskName
			}

			if workerID, ok := buildMap["worker_id"].(string); ok {
				build.WorkerID = workerID
			}

			if success, ok := buildMap["success"].(bool); ok {
				build.Success = success
			}

			if startTimeStr, ok := buildMap["start_time"].(string); ok {
				if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
					build.StartTime = startTime
				}
			}

			if endTimeStr, ok := buildMap["end_time"].(string); ok {
				if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
					build.EndTime = endTime
				}
			}

			if cacheHitRate, ok := buildMap["cache_hit_rate"].(float64); ok {
				build.CacheHitRate = cacheHitRate
			}

			if resourceUsage, ok := buildMap["resource_usage"].(map[string]any); ok {
				if cpu, ok := resourceUsage["cpu_percent"].(float64); ok {
					build.CPUUsage = cpu
				}
				if mem, ok := resourceUsage["memory_usage"].(float64); ok {
					build.MemoryUsage = mem
				}
				if disk, ok := resourceUsage["disk_io"].(float64); ok {
					build.DiskUsage = disk
				}
			}

			if errorMsg, ok := buildMap["error_message"].(string); ok {
				build.ErrorMessage = errorMsg
			}

			ml.RecordBuild(build)
		}
	}
}

// collectFromCoordinator collects data from the coordinator service
func (ml *MLService) collectFromCoordinator() {
	coordinatorURL := fmt.Sprintf("http://%s:%d/api/workers", ml.ContinuousLearning.CoordinatorHost, ml.ContinuousLearning.CoordinatorPort)

	resp, err := http.Get(coordinatorURL)
	if err != nil {
		log.Printf("Failed to collect data from coordinator: %v", err)
		return
	}
	defer resp.Body.Close()

	var workers []struct {
		ID       string    `json:"id"`
		Status   string    `json:"status"`
		LastPing time.Time `json:"last_ping"`
		Metrics  struct {
			BuildCount       int           `json:"build_count"`
			TotalBuildTime   time.Duration `json:"total_build_time"`
			AverageBuildTime time.Duration `json:"average_build_time"`
			SuccessRate      float64       `json:"success_rate"`
		} `json:"metrics"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&workers); err != nil {
		log.Printf("Failed to decode coordinator data: %v", err)
		return
	}

	// Convert coordinator data to worker metrics
	for _, worker := range workers {
		metric := WorkerMetric{
			WorkerID:     worker.ID,
			Timestamp:    time.Now(),
			CPUUsage:     0.5, // Default values, would be enhanced with actual metrics
			MemoryUsage:  0.5,
			DiskUsage:    0.3,
			ActiveBuilds: 0,
			QueueLength:  0,
			ResponseTime: time.Since(worker.LastPing),
		}

		if worker.Status == "busy" {
			metric.ActiveBuilds = 1
		}

		ml.RecordWorkerMetrics(metric)
	}
}

// checkRetrainingConditions checks if models should be retrained
func (ml *MLService) checkRetrainingConditions() {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()

	now := time.Now()
	shouldRetrain := false
	reason := ""

	// Check time-based retraining
	if now.Sub(ml.LearningStats.LastRetraining) >= ml.ContinuousLearning.RetrainingInterval {
		shouldRetrain = true
		reason = "scheduled retraining interval reached"
	}

	// Check data volume
	if len(ml.BuildHistory) >= ml.ContinuousLearning.MinDataPoints {
		newDataCount := 0
		for _, record := range ml.BuildHistory {
			if record.StartTime.After(ml.LearningStats.LastRetraining) {
				newDataCount++
			}
		}

		if newDataCount >= ml.ContinuousLearning.MinDataPoints/2 {
			shouldRetrain = true
			reason = fmt.Sprintf("sufficient new data collected (%d records)", newDataCount)
		}
	}

	// Check performance degradation with detailed validation
	if now.Sub(ml.LearningStats.LastAccuracyCheck) >= 6*time.Hour { // Check accuracy every 6 hours
		performance := ml.validateModelPerformance()
		avgAccuracy := ml.calculateAverageModelAccuracy()

		// Update stored accuracies
		ml.Models.BuildTimePredictor.Accuracy = performance["build_time"]
		ml.Models.ResourcePredictor.Accuracy = performance["resource"]
		ml.Models.FailurePredictor.Accuracy = performance["failure"]
		ml.Models.CachePredictor.Accuracy = performance["cache"]
		ml.LearningStats.LastAccuracyCheck = now

		if avgAccuracy < ml.ContinuousLearning.PerformanceThreshold {
			shouldRetrain = true
			reason = fmt.Sprintf("model accuracy degraded (%.2f < %.2f)", avgAccuracy, ml.ContinuousLearning.PerformanceThreshold)
		}
	}

	if shouldRetrain {
		go ml.triggerRetraining(reason)
	}
}

// backupCurrentModels creates a backup of current models
func (ml *MLService) backupCurrentModels() {
	backup := ModelBackup{
		Timestamp: time.Now(),
		Models:    ml.Models, // This creates a copy
		Accuracy:  ml.calculateAverageModelAccuracy(),
		Version:   ml.LearningStats.CurrentVersion,
	}

	// Keep only last 5 backups
	if len(ml.LearningStats.ModelBackups) >= 5 {
		ml.LearningStats.ModelBackups = ml.LearningStats.ModelBackups[1:]
	}
	ml.LearningStats.ModelBackups = append(ml.LearningStats.ModelBackups, backup)

	log.Printf("Created model backup version %s with accuracy %.3f", backup.Version, backup.Accuracy)
}

// rollbackToBackup restores models from a backup
func (ml *MLService) rollbackToBackup(version string) error {
	for _, backup := range ml.LearningStats.ModelBackups {
		if backup.Version == version {
			ml.Models = backup.Models // Restore models
			ml.LearningStats.CurrentVersion = version
			ml.LearningStats.AverageAccuracy = backup.Accuracy
			log.Printf("Rolled back to model version %s (accuracy: %.3f)", version, backup.Accuracy)
			return nil
		}
	}
	return fmt.Errorf("backup version %s not found", version)
}

// generateModelVersion creates a new version string
func (ml *MLService) generateModelVersion() string {
	return fmt.Sprintf("v%d.%d", ml.LearningStats.RetrainingCount+1, time.Now().Unix())
}

// triggerRetraining initiates model retraining with backup and rollback capabilities
func (ml *MLService) triggerRetraining(reason string) {
	log.Printf("Triggering model retraining: %s", reason)

	ml.LearningStats.IsRetraining = true
	defer func() { ml.LearningStats.IsRetraining = false }()

	// Create backup before retraining
	ml.backupCurrentModels()

	// Generate new version
	newVersion := ml.generateModelVersion()

	if err := ml.TrainModels(); err != nil {
		log.Printf("Retraining failed: %v", err)
		// Attempt rollback on failure
		if len(ml.LearningStats.ModelBackups) > 0 {
			lastBackup := ml.LearningStats.ModelBackups[len(ml.LearningStats.ModelBackups)-1]
			if rollbackErr := ml.rollbackToBackup(lastBackup.Version); rollbackErr != nil {
				log.Printf("Rollback failed: %v", rollbackErr)
			} else {
				log.Printf("Rolled back to previous version due to training failure")
			}
		}
		return
	}

	// Validate new models
	newAccuracy := ml.calculateAverageModelAccuracy()

	// Check if new models are significantly worse
	previousAccuracy := ml.LearningStats.AverageAccuracy
	accuracyDrop := previousAccuracy - newAccuracy

	if accuracyDrop > 0.2 { // More than 20% accuracy drop
		log.Printf("New model accuracy (%.3f) significantly worse than previous (%.3f), rolling back", newAccuracy, previousAccuracy)
		if len(ml.LearningStats.ModelBackups) > 0 {
			lastBackup := ml.LearningStats.ModelBackups[len(ml.LearningStats.ModelBackups)-1]
			if rollbackErr := ml.rollbackToBackup(lastBackup.Version); rollbackErr != nil {
				log.Printf("Rollback failed: %v", rollbackErr)
			}
		}
		return
	}

	// Accept new models
	ml.LearningStats.LastRetraining = time.Now()
	ml.LearningStats.RetrainingCount++
	ml.LearningStats.AverageAccuracy = newAccuracy
	ml.LearningStats.LastAccuracyCheck = time.Now()
	ml.LearningStats.CurrentVersion = newVersion

	log.Printf("Model retraining completed successfully - new version %s (accuracy: %.3f)", newVersion, newAccuracy)
}

// calculateAverageModelAccuracy calculates average accuracy across all models
func (ml *MLService) calculateAverageModelAccuracy() float64 {
	accuracies := []float64{
		ml.Models.BuildTimePredictor.Accuracy,
		ml.Models.ResourcePredictor.Accuracy,
		ml.Models.FailurePredictor.Accuracy,
		ml.Models.CachePredictor.Accuracy,
	}

	sum := 0.0
	count := 0
	for _, acc := range accuracies {
		if acc > 0 {
			sum += acc
			count++
		}
	}

	if count == 0 {
		return 0.5 // Default accuracy
	}

	return sum / float64(count)
}

// validateModelPerformance validates model performance against recent data
func (ml *MLService) validateModelPerformance() map[string]float64 {
	performance := make(map[string]float64)

	if len(ml.BuildHistory) < 10 {
		// Not enough data for validation
		performance["build_time"] = 0.5
		performance["resource"] = 0.5
		performance["failure"] = 0.5
		performance["cache"] = 0.5
		return performance
	}

	// Validate build time predictions
	recentBuilds := ml.BuildHistory[len(ml.BuildHistory)-int(math.Min(50, float64(len(ml.BuildHistory)))):]
	buildTimeErrors := 0.0
	buildTimeCount := 0

	for _, record := range recentBuilds {
		if !record.Success {
			continue
		}

		predicted, _ := ml.PredictBuildTime(record.ProjectPath, record.TaskName, nil)
		actual := record.Duration

		// Calculate prediction error as percentage
		if predicted > 0 {
			error := math.Abs(predicted.Seconds()-actual.Seconds()) / predicted.Seconds()
			buildTimeErrors += error
			buildTimeCount++
		}
	}

	if buildTimeCount > 0 {
		avgError := buildTimeErrors / float64(buildTimeCount)
		performance["build_time"] = math.Max(0, 1.0-avgError) // Convert error to accuracy
	} else {
		performance["build_time"] = 0.5
	}

	// Validate resource predictions
	resourceErrors := 0.0
	resourceCount := 0

	for _, record := range recentBuilds {
		if !record.Success {
			continue
		}

		predicted := ml.PredictResourceNeeds(record.ProjectPath, record.TaskName)
		actualCPU := record.CPUUsage
		actualMem := record.MemoryUsage
		actualDisk := record.DiskUsage

		// Calculate average resource prediction error
		cpuError := math.Abs(predicted.CPU-actualCPU) / math.Max(actualCPU, 0.1)
		memError := math.Abs(predicted.Memory-actualMem) / math.Max(actualMem, 0.1)
		diskError := math.Abs(predicted.Disk-actualDisk) / math.Max(actualDisk, 0.1)

		avgResourceError := (cpuError + memError + diskError) / 3.0
		resourceErrors += avgResourceError
		resourceCount++
	}

	if resourceCount > 0 {
		avgError := resourceErrors / float64(resourceCount)
		performance["resource"] = math.Max(0, 1.0-avgError)
	} else {
		performance["resource"] = 0.5
	}

	// Validate failure predictions
	failureCorrect := 0
	failureTotal := 0

	for _, record := range recentBuilds {
		predictedRisk := ml.PredictFailureRisk(record.ProjectPath, record.TaskName)
		actualFailure := !record.Success

		// Consider prediction correct if risk > 0.5 and failed, or risk < 0.5 and succeeded
		correct := (predictedRisk > 0.5) == actualFailure
		if correct {
			failureCorrect++
		}
		failureTotal++
	}

	if failureTotal > 0 {
		performance["failure"] = float64(failureCorrect) / float64(failureTotal)
	} else {
		performance["failure"] = 0.5
	}

	// Validate cache predictions
	cacheErrors := 0.0
	cacheCount := 0

	for _, record := range recentBuilds {
		predicted := ml.PredictCacheHitRate(record.ProjectPath, record.TaskName)
		actual := record.CacheHitRate

		error := math.Abs(predicted-actual) / math.Max(actual, 0.1)
		cacheErrors += error
		cacheCount++
	}

	if cacheCount > 0 {
		avgError := cacheErrors / float64(cacheCount)
		performance["cache"] = math.Max(0, 1.0-avgError)
	} else {
		performance["cache"] = 0.5
	}

	return performance
}

// updateModelAccuracies updates model accuracies based on validation
func (ml *MLService) updateModelAccuracies() {
	performance := ml.validateModelPerformance()

	ml.Models.BuildTimePredictor.Accuracy = performance["build_time"]
	ml.Models.ResourcePredictor.Accuracy = performance["resource"]
	ml.Models.FailurePredictor.Accuracy = performance["failure"]
	ml.Models.CachePredictor.Accuracy = performance["cache"]

	ml.LearningStats.LastAccuracyCheck = time.Now()
}
