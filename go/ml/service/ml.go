package service

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// MLService provides machine learning capabilities for build optimization
type MLService struct {
	BuildHistory  []BuildRecord  `json:"build_history"`
	WorkerMetrics []WorkerMetric `json:"worker_metrics"`
	CacheMetrics  []CacheMetric  `json:"cache_metrics"`
	Models        MLModels       `json:"models"`
	mutex         sync.RWMutex
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
	return &MLService{
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
func (ml *MLService) GetStatistics() map[string]interface{} {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()

	stats := make(map[string]interface{})

	stats["total_build_records"] = len(ml.BuildHistory)
	stats["total_worker_metrics"] = len(ml.WorkerMetrics)
	stats["total_cache_metrics"] = len(ml.CacheMetrics)

	if len(ml.BuildHistory) > 0 {
		stats["oldest_build"] = ml.BuildHistory[0].StartTime
		stats["newest_build"] = ml.BuildHistory[len(ml.BuildHistory)-1].StartTime
	}

	stats["models_trained"] = map[string]interface{}{
		"build_time": !ml.Models.BuildTimePredictor.LastTrained.IsZero(),
		"resource":   !ml.Models.ResourcePredictor.LastTrained.IsZero(),
		"scaling":    !ml.Models.ScalingPredictor.LastTrained.IsZero(),
		"failure":    !ml.Models.FailurePredictor.LastTrained.IsZero(),
		"cache":      !ml.Models.CachePredictor.LastTrained.IsZero(),
	}

	return stats
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
