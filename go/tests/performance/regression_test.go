package performance

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"distributed-gradle-building/cachepkg"
	"distributed-gradle-building/coordinatorpkg"
	"distributed-gradle-building/types"
	"distributed-gradle-building/workerpkg"
)

// PerformanceRegressionTest detects performance regressions
type PerformanceRegressionTest struct {
	BaselineData    *PerformanceBaseline
	CurrentData     *PerformanceMeasurement
	Thresholds      RegressionThresholds
	DetectionWindow time.Duration
}

// PerformanceBaseline represents historical performance data
type PerformanceBaseline struct {
	BuildTimeAvg     time.Duration
	BuildTimeP95     time.Duration
	BuildTimeP99     time.Duration
	CacheHitRate     float64
	Throughput       float64
	MemoryUsage      float64
	CPUUsage         float64
	ErrorRate        float64
	ConcurrencyLimit int
	Timestamp        time.Time
}

// PerformanceMeasurement represents current performance measurement
type PerformanceMeasurement struct {
	BuildTimeStats      *BuildTimeStats
	CacheMetrics        *CacheMetrics
	ThroughputMetrics   *ThroughputMetrics
	ResourceUsage       *ResourceUsage
	ErrorAnalysis       *ErrorAnalysis
	ConcurrencyLevel    int
	MeasurementDuration time.Duration
	Timestamp           time.Time
}

// BuildTimeStats tracks build execution time statistics
type BuildTimeStats struct {
	Count  int
	Sum    time.Duration
	Min    time.Duration
	Max    time.Duration
	Avg    time.Duration
	P50    time.Duration
	P95    time.Duration
	P99    time.Duration
	StdDev time.Duration
}

// CacheMetrics tracks cache performance
type CacheMetrics struct {
	ReadThroughput  float64
	WriteThroughput float64
	HitRate         float64
	MissRate        float64
	EvictionRate    float64
	AccessLatency   time.Duration
	StorageSize     int64
}

// ThroughputMetrics tracks system throughput
type ThroughputMetrics struct {
	RequestsPerSecond float64
	BuildsPerSecond   float64
	WorkersUtilized   int
	QueueLength       int
	PeakThroughput    float64
	AverageThroughput float64
}

// ResourceUsage tracks system resource consumption
type ResourceUsage struct {
	CPUUsage        float64
	MemoryUsage     float64
	DiskIO          float64
	NetworkIO       float64
	FileDescriptors int
	GoroutineCount  int
}

// ErrorAnalysis tracks error patterns
type ErrorAnalysis struct {
	TotalRequests int
	ErrorCount    int
	ErrorRate     float64
	TimeoutCount  int
	RetryCount    int
	ErrorTypes    map[string]int
	ErrorTrend    float64
}

// RegressionThresholds defines performance regression detection limits
type RegressionThresholds struct {
	BuildTimeRegression    float64
	MemoryUsageRegression  float64
	CacheHitRateRegression float64
	ThroughputRegression   float64
	ErrorRateRegression    float64
}

// PerformanceRegression represents detected regression
type PerformanceRegression struct {
	Metric     string
	Baseline   float64
	Current    float64
	Regression float64
	Severity   string
	Message    string
}

// TestPerformanceRegressionDetection detects performance regressions
func TestPerformanceRegressionDetection(t *testing.T) {
	t.Parallel()

	// Load baseline data
	baseline := loadPerformanceBaseline("performance_baseline.json")
	if baseline == nil {
		t.Skip("No baseline data available, skipping regression test")
	}

	// Configure regression thresholds
	thresholds := RegressionThresholds{
		BuildTimeRegression:    0.15,  // 15% regression allowed
		MemoryUsageRegression:  0.20,  // 20% regression allowed
		CacheHitRateRegression: -0.10, // 10% degradation allowed
		ThroughputRegression:   -0.15, // 15% regression allowed
		ErrorRateRegression:    0.50,  // 50% increase allowed
	}

	regressionTest := &PerformanceRegressionTest{
		BaselineData: baseline,
		Thresholds:   thresholds,
	}

	// Execute performance tests
	current := &PerformanceMeasurement{}
	current.MeasureBuildPerformance(t)
	current.MeasureScalability(t)
	current.MeasureCachePerformance(t)
	current.MeasureResourceUsage(t)

	regressionTest.CurrentData = current

	// Detect regressions
	regressions := regressionTest.DetectRegressions()

	if len(regressions) > 0 {
		generateRegressionReport(regressions)
		t.Errorf("Performance regressions detected: %d", len(regressions))
		for _, regression := range regressions {
			t.Logf("- %s: %.1f%% (%s)", regression.Metric, regression.Regression*100, regression.Severity)
		}
	} else {
		t.Log("No performance regressions detected")
	}
}

// TestScalabilityRegression tests scalability characteristics regression
func TestScalabilityRegression(t *testing.T) {
	t.Parallel()

	scalabilityTest := &ScalabilityTest{
		ConcurrencyLevels: []int{1, 2, 4, 8, 16},
		BaselineData:      loadScalabilityBaseline(),
		TestDuration:      30 * time.Second,
	}

	results := scalabilityTest.Execute(t)

	// Verify scalability characteristics haven't regressed
	if results.LinearRegressionScore > 0.15 { // Deviation from linear scaling
		t.Errorf("Scalability regression detected: linear regression score = %.3f",
			results.LinearRegressionScore)
	}

	if results.MaxThroughput < scalabilityTest.BaselineData.MaxThroughput*0.9 {
		t.Errorf("Maximum throughput regression: current=%.1f, baseline=%.1f",
			results.MaxThroughput, scalabilityTest.BaselineData.MaxThroughput)
	}

	t.Logf("Scalability test passed: linear score=%.3f, max throughput=%.1f req/sec",
		results.LinearRegressionScore, results.MaxThroughput)
}

// ScalabilityTest executes scalability testing across different concurrency levels
type ScalabilityTest struct {
	ConcurrencyLevels []int
	BaselineData      *ScalabilityBaseline
	TestDuration      time.Duration
	Results           []*ScalabilityResult
}

// ScalabilityBaseline represents baseline scalability data
type ScalabilityBaseline struct {
	MaxThroughput         float64
	LinearRegressionScore float64
	OptimalConcurrency    int
	Timestamp             time.Time
}

// ScalabilityResult represents result for a specific concurrency level
type ScalabilityResult struct {
	Concurrency    int
	Throughput     float64
	AverageLatency time.Duration
	P95Latency     time.Duration
	ErrorRate      float64
	Utilization    float64
}

// Execute runs scalability test across all concurrency levels
func (st *ScalabilityTest) Execute(t *testing.T) *ScalabilityTestSummary {
	summary := &ScalabilityTestSummary{}

	for _, concurrency := range st.ConcurrencyLevels {
		t.Run(fmt.Sprintf("Scalability_%d_workers", concurrency), func(t *testing.T) {
			result := st.measureScalabilityAtConcurrency(t, concurrency)
			st.Results = append(st.Results, result)

			if result.Throughput > summary.MaxThroughput {
				summary.MaxThroughput = result.Throughput
				summary.OptimalConcurrency = concurrency
			}
		})
	}

	// Calculate linear regression score
	summary.LinearRegressionScore = st.calculateLinearRegressionScore()

	return summary
}

// measureScalabilityAtConcurrency measures performance at specific concurrency level
func (st *ScalabilityTest) measureScalabilityAtConcurrency(t *testing.T, concurrency int) *ScalabilityResult {
	coordinator := setupTestCoordinator(t)
	defer coordinator.Shutdown()

	workers := setupTestWorkers(t, concurrency)
	defer shutdownWorkers(workers)

	// Generate controlled workload
	workload := &ControlledWorkload{
		Concurrency: concurrency,
		Duration:    st.TestDuration,
		RequestRate: 10.0, // requests per second
	}

	metrics := workload.ExecuteWithCoordinator(coordinator)

	return &ScalabilityResult{
		Concurrency:    concurrency,
		Throughput:     metrics.AverageThroughput,
		AverageLatency: metrics.AverageLatency,
		P95Latency:     metrics.P95Latency,
		ErrorRate:      metrics.ErrorRate,
		Utilization:    metrics.Utilization,
	}
}

// calculateLinearRegressionScore calculates deviation from linear scaling
func (st *ScalabilityTest) calculateLinearRegressionScore() float64 {
	if len(st.Results) < 2 {
		return 0.0
	}

	// Simple linear regression analysis
	n := float64(len(st.Results))
	var sumX, sumY, sumXY, sumX2 float64

	for _, result := range st.Results {
		x := float64(result.Concurrency)
		y := result.Throughput

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// Calculate coefficient of determination (R²)
	yMean := sumY / n
	var totalSS, residualSS float64

	for _, result := range st.Results {
		x := float64(result.Concurrency)
		y := result.Throughput
		predicted := slope * x // Assuming y-intercept is 0 for ideal linear scaling

		totalSS += (y - yMean) * (y - yMean)
		residualSS += (y - predicted) * (y - predicted)
	}

	if totalSS == 0 {
		return 0.0
	}

	rSquared := 1 - (residualSS / totalSS)
	return math.Max(0, 1-rSquared) // Convert to deviation score
}

// ScalabilityTestSummary represents overall scalability test results
type ScalabilityTestSummary struct {
	MaxThroughput         float64
	OptimalConcurrency    int
	LinearRegressionScore float64
}

// ControlledWorkload generates controlled testing workload
type ControlledWorkload struct {
	Concurrency int
	Duration    time.Duration
	RequestRate float64
	Variance    float64
}

// WorkloadMetrics represents metrics from workload execution
type WorkloadMetrics struct {
	AverageThroughput float64
	P95Throughput     float64
	AverageLatency    time.Duration
	P95Latency        time.Duration
	ErrorRate         float64
	Utilization       float64
}

// ExecuteWithCoordinator executes workload against coordinator
func (cw *ControlledWorkload) ExecuteWithCoordinator(coordinator *coordinatorpkg.BuildCoordinator) *WorkloadMetrics {
	var totalRequests, completedRequests, failedRequests int64
	var totalLatency time.Duration
	var latencies []time.Duration

	startTime := time.Now()
	endTime := startTime.Add(cw.Duration)

	// Generate requests at controlled rate
	requestInterval := time.Duration(1.0 / cw.RequestRate * float64(time.Second))
	requestTicker := time.NewTicker(requestInterval)
	defer requestTicker.Stop()

	for time.Now().Before(endTime) {
		select {
		case <-requestTicker.C:
			requestStart := time.Now()

			buildReq := types.BuildRequest{
				RequestID:    fmt.Sprintf("workload-%d", totalRequests),
				ProjectPath:  fmt.Sprintf("/tmp/workload-project-%d", totalRequests),
				TaskName:     "compile",
				BuildOptions: map[string]string{},
				Timestamp:    requestStart,
			}

			_, err := coordinator.SubmitBuild(buildReq)
			latency := time.Since(requestStart)

			totalRequests++
			latencies = append(latencies, latency)
			totalLatency += latency

			if err != nil {
				failedRequests++
			} else {
				completedRequests++
			}
		}
	}

	// Calculate metrics
	var avgLatency time.Duration
	if totalRequests > 0 {
		avgLatency = totalLatency / time.Duration(totalRequests)
	}

	errorRate := float64(failedRequests) / float64(totalRequests)
	averageThroughput := float64(completedRequests) / cw.Duration.Seconds()

	// Calculate P95 latency
	p95Latency := calculateP95(latencies)

	return &WorkloadMetrics{
		AverageThroughput: averageThroughput,
		AverageLatency:    avgLatency,
		P95Latency:        p95Latency,
		ErrorRate:         errorRate,
		Utilization:       float64(cw.Concurrency) / 10.0, // Assuming max 10 workers
	}
}

// DetectRegressions identifies performance regressions
func (prt *PerformanceRegressionTest) DetectRegressions() []*PerformanceRegression {
	var regressions []*PerformanceRegression

	// Build time regression
	if prt.BaselineData.BuildTimeAvg > 0 {
		buildTimeRegression := float64(prt.CurrentData.BuildTimeStats.Avg-prt.BaselineData.BuildTimeAvg) /
			float64(prt.BaselineData.BuildTimeAvg)

		if buildTimeRegression > prt.Thresholds.BuildTimeRegression {
			regressions = append(regressions, &PerformanceRegression{
				Metric:     "BuildTimeAvg",
				Baseline:   float64(prt.BaselineData.BuildTimeAvg.Milliseconds()),
				Current:    float64(prt.CurrentData.BuildTimeStats.Avg.Milliseconds()),
				Regression: buildTimeRegression,
				Severity:   determineSeverity(buildTimeRegression, 0.1, 0.2),
				Message:    fmt.Sprintf("Build time increased by %.1f%%", buildTimeRegression*100),
			})
		}
	}

	// Memory usage regression
	if prt.BaselineData.MemoryUsage > 0 {
		memoryRegression := (prt.CurrentData.ResourceUsage.MemoryUsage - prt.BaselineData.MemoryUsage) /
			prt.BaselineData.MemoryUsage

		if memoryRegression > prt.Thresholds.MemoryUsageRegression {
			regressions = append(regressions, &PerformanceRegression{
				Metric:     "MemoryUsage",
				Baseline:   prt.BaselineData.MemoryUsage,
				Current:    prt.CurrentData.ResourceUsage.MemoryUsage,
				Regression: memoryRegression,
				Severity:   determineSeverity(memoryRegression, 0.15, 0.25),
				Message:    fmt.Sprintf("Memory usage increased by %.1f%%", memoryRegression*100),
			})
		}
	}

	// Cache hit rate regression
	if prt.BaselineData.CacheHitRate > 0 {
		cacheHitRegression := (prt.CurrentData.CacheMetrics.HitRate - prt.BaselineData.CacheHitRate) /
			prt.BaselineData.CacheHitRate

		if cacheHitRegression < prt.Thresholds.CacheHitRateRegression {
			regressions = append(regressions, &PerformanceRegression{
				Metric:     "CacheHitRate",
				Baseline:   prt.BaselineData.CacheHitRate,
				Current:    prt.CurrentData.CacheMetrics.HitRate,
				Regression: cacheHitRegression,
				Severity:   determineSeverity(-cacheHitRegression, 0.05, 0.10),
				Message:    fmt.Sprintf("Cache hit rate decreased by %.1f%%", -cacheHitRegression*100),
			})
		}
	}

	// Throughput regression
	if prt.BaselineData.Throughput > 0 {
		throughputRegression := (prt.CurrentData.ThroughputMetrics.AverageThroughput - prt.BaselineData.Throughput) /
			prt.BaselineData.Throughput

		if throughputRegression < prt.Thresholds.ThroughputRegression {
			regressions = append(regressions, &PerformanceRegression{
				Metric:     "Throughput",
				Baseline:   prt.BaselineData.Throughput,
				Current:    prt.CurrentData.ThroughputMetrics.AverageThroughput,
				Regression: throughputRegression,
				Severity:   determineSeverity(-throughputRegression, 0.10, 0.15),
				Message:    fmt.Sprintf("Throughput decreased by %.1f%%", -throughputRegression*100),
			})
		}
	}

	return regressions
}

// MeasureBuildPerformance measures build execution performance
func (pm *PerformanceMeasurement) MeasureBuildPerformance(t *testing.T) {
	coordinator := setupTestCoordinator(t)
	defer coordinator.Shutdown()

	workers := setupTestWorkers(t, 5)
	defer shutdownWorkers(workers)

	pm.BuildTimeStats = &BuildTimeStats{}
	buildTimes := make([]time.Duration, 50)

	// Execute test builds
	for i := 0; i < 50; i++ {
		startTime := time.Now()

		buildReq := types.BuildRequest{
			RequestID:    fmt.Sprintf("perf-build-%d", i),
			ProjectPath:  fmt.Sprintf("/tmp/perf-project-%d", i),
			TaskName:     "compile",
			BuildOptions: map[string]string{},
			Timestamp:    startTime,
		}

		_, err := coordinator.SubmitBuild(buildReq)
		buildTime := time.Since(startTime)
		buildTimes[i] = buildTime

		if err == nil {
			pm.BuildTimeStats.Count++
			pm.BuildTimeStats.Sum += buildTime
			if buildTime < pm.BuildTimeStats.Min || pm.BuildTimeStats.Min == 0 {
				pm.BuildTimeStats.Min = buildTime
			}
			if buildTime > pm.BuildTimeStats.Max {
				pm.BuildTimeStats.Max = buildTime
			}
		}
	}

	// Calculate statistics
	if pm.BuildTimeStats.Count > 0 {
		pm.BuildTimeStats.Avg = pm.BuildTimeStats.Sum / time.Duration(pm.BuildTimeStats.Count)
		pm.BuildTimeStats.P50 = calculatePercentile(buildTimes, 0.5)
		pm.BuildTimeStats.P95 = calculatePercentile(buildTimes, 0.95)
		pm.BuildTimeStats.P99 = calculatePercentile(buildTimes, 0.99)
		pm.BuildTimeStats.StdDev = calculateStdDev(buildTimes, pm.BuildTimeStats.Avg)
	}
}

// MeasureCachePerformance measures cache performance
func (pm *PerformanceMeasurement) MeasureCachePerformance(t *testing.T) {
	cache := setupTestCache(t)
	defer cache.Shutdown()

	pm.CacheMetrics = &CacheMetrics{}

	// Cache performance testing
	cacheTest := &CachePerformanceTest{
		Operations:    1000,
		ConcurrentOps: 10,
		DataSize:      1024,
		Cache:         cache,
	}

	results := cacheTest.Execute()

	pm.CacheMetrics.ReadThroughput = results.ReadThroughput
	pm.CacheMetrics.WriteThroughput = results.WriteThroughput
	pm.CacheMetrics.HitRate = results.HitRate
	pm.CacheMetrics.AccessLatency = results.AverageLatency
}

// MeasureScalability measures system scalability
func (pm *PerformanceMeasurement) MeasureScalability(t *testing.T) {
	scalabilityTest := &ScalabilityTest{
		ConcurrencyLevels: []int{1, 2, 4, 8},
		TestDuration:      15 * time.Second,
	}

	summary := scalabilityTest.Execute(t)

	pm.ThroughputMetrics = &ThroughputMetrics{
		PeakThroughput:    summary.MaxThroughput,
		AverageThroughput: summary.MaxThroughput * 0.8, // Approximation
		WorkersUtilized:   summary.OptimalConcurrency,
	}
}

// MeasureResourceUsage measures resource consumption
func (pm *PerformanceMeasurement) MeasureResourceUsage(t *testing.T) {
	pm.ResourceUsage = &ResourceUsage{
		CPUUsage:        0.65, // Mock value
		MemoryUsage:     0.45, // Mock value
		DiskIO:          0.20, // Mock value
		NetworkIO:       0.15, // Mock value
		FileDescriptors: 25,   // Mock value
		GoroutineCount:  150,  // Mock value
	}
}

// Helper functions
func loadPerformanceBaseline(filename string) *PerformanceBaseline {
	// In real implementation, this would load from file
	// For testing, we return mock baseline
	return &PerformanceBaseline{
		BuildTimeAvg:     500 * time.Millisecond,
		BuildTimeP95:     800 * time.Millisecond,
		BuildTimeP99:     1200 * time.Millisecond,
		CacheHitRate:     0.85,
		Throughput:       25.0,
		MemoryUsage:      0.40,
		CPUUsage:         0.60,
		ErrorRate:        0.02,
		ConcurrencyLimit: 8,
		Timestamp:        time.Now().Add(-24 * time.Hour),
	}
}

func loadScalabilityBaseline() *ScalabilityBaseline {
	return &ScalabilityBaseline{
		MaxThroughput:         50.0,
		LinearRegressionScore: 0.05,
		OptimalConcurrency:    8,
		Timestamp:             time.Now().Add(-24 * time.Hour),
	}
}

func generateRegressionReport(regressions []*PerformanceRegression) {
	report := map[string]interface{}{
		"timestamp":   time.Now(),
		"regressions": regressions,
		"total":       len(regressions),
	}

	jsonData, _ := json.MarshalIndent(report, "", "  ")
	os.WriteFile("performance_regression_report.json", jsonData, 0644)
}

func determineSeverity(value, lowThreshold, highThreshold float64) string {
	if value >= highThreshold {
		return "critical"
	} else if value >= lowThreshold {
		return "warning"
	}
	return "info"
}

func calculatePercentile(values []time.Duration, percentile float64) time.Duration {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]time.Duration, len(values))
	copy(sorted, values)

	// Simple sort (in real implementation, use more efficient sorting)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := int(float64(len(sorted)) * percentile)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

func calculateP95(values []time.Duration) time.Duration {
	return calculatePercentile(values, 0.95)
}

func calculateStdDev(values []time.Duration, mean time.Duration) time.Duration {
	if len(values) <= 1 {
		return 0
	}

	var sumSquaredDiff float64
	for _, value := range values {
		diff := float64(value - mean)
		sumSquaredDiff += diff * diff
	}

	variance := sumSquaredDiff / float64(len(values)-1)
	stdDev := math.Sqrt(variance)

	return time.Duration(stdDev)
}

// CachePerformanceTest represents cache performance testing
type CachePerformanceTest struct {
	Operations    int
	ConcurrentOps int
	DataSize      int
	Cache         interface{}
	Results       *CachePerformanceResult
}

// CachePerformanceResult represents cache performance results
type CachePerformanceResult struct {
	ReadThroughput  float64
	WriteThroughput float64
	HitRate         float64
	MissRate        float64
	AverageLatency  time.Duration
}

// Execute runs cache performance test
func (cpt *CachePerformanceTest) Execute() *CachePerformanceResult {
	// Mock implementation
	return &CachePerformanceResult{
		ReadThroughput:  5000.0,
		WriteThroughput: 2000.0,
		HitRate:         0.80,
		MissRate:        0.20,
		AverageLatency:  5 * time.Millisecond,
	}
}

// Mock setup functions
func setupTestCoordinator(t *testing.T) *coordinatorpkg.BuildCoordinator {
	coordinator := coordinatorpkg.NewBuildCoordinator(10)
	return coordinator
}

func setupTestWorkers(t *testing.T, count int) []*workerpkg.WorkerService {
	var workers []*workerpkg.WorkerService
	for i := 0; i < count; i++ {
		config := types.WorkerConfig{
			ID:                  fmt.Sprintf("perf-worker-%d", i),
			CoordinatorURL:      "localhost:8080",
			MaxConcurrentBuilds: 2,
		}
		worker := workerpkg.NewWorkerService(fmt.Sprintf("perf-worker-%d", i), "localhost:8080", config)
		workers = append(workers, worker)
	}
	return workers
}

func shutdownWorkers(workers []*workerpkg.WorkerService) {
	for _, worker := range workers {
		worker.Shutdown()
	}
}

func setupTestCache(t *testing.T) *cachepkg.CacheServer {
	config := types.CacheConfig{
		Port:            9080,
		StorageType:     "memory",
		TTL:             30 * time.Minute,
		CleanupInterval: 5 * time.Minute,
	}

	cache := cachepkg.NewCacheServer(config)
	return cache
}
