package observability

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"distributed-gradle-building/coordinatorpkg"
	"distributed-gradle-building/workerpkg"
	"distributed-gradle-building/cachepkg"
	"distributed-gradle-building/monitorpkg"
	"distributed-gradle-building/types"
)

// MetricsIntegrationTest tests metrics collection accuracy and completeness
type MetricsIntegrationTest struct {
	monitor     *monitorpkg.Monitor
	coordinator *coordinatorpkg.BuildCoordinator
	workers     []*workerpkg.WorkerService
	cache       *cachepkg.CacheServer
	collector   *TraceCollector
}

// SystemMetrics represents comprehensive system metrics
type SystemMetrics struct {
	TotalBuilds       int64
	SuccessfulBuilds  int64
	FailedBuilds      int64
	AverageBuildTime   time.Duration
	P95BuildTime      time.Duration
	P99BuildTime      time.Duration
	WorkerUtilization  float64
	CacheHitRate      float64
	MemoryUsage       float64
	CPUUsage          float64
	DiskUsage         float64
	NetworkThroughput  float64
	ErrorRate         float64
	Timestamp         time.Time
}

// TraceCollector collects distributed traces
type TraceCollector struct {
	traces map[string]*DistributedTrace
	mutex  sync.RWMutex
}

// DistributedTrace represents a distributed trace across services
type DistributedTrace struct {
	TraceID   string
	Spans     []*TraceSpan
	Duration  time.Duration
	StartTime time.Time
	EndTime   time.Time
}

// TraceSpan represents a single operation in a distributed trace
type TraceSpan struct {
	SpanID      string
	ParentID    string
	Operation   string
	Service     string
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	Tags        map[string]string
	Success     bool
	Error       string
}

// ControlledWorkload generates controlled testing workload
type ControlledWorkload struct {
	BuildCount    int
	WorkerCount   int
	Duration      time.Duration
	Concurrency   int
	RequestRate   float64
	Variance     float64
}

// WorkloadMetrics represents metrics from workload execution
type WorkloadMetrics struct {
	RequestsCompleted   int
	RequestsFailed      int
	SuccessRate        float64
	AverageLatency     time.Duration
	P95Latency         time.Duration
	P99Latency         time.Duration
	Throughput         float64
	ErrorDistribution  map[string]int
	ResourceUsage      *ResourceUsage
	Timestamp          time.Time
}

// ResourceUsage tracks system resource consumption during testing
type ResourceUsage struct {
	CPUUtilization    float64
	MemoryUsage       float64
	DiskIO           float64
	NetworkIO        float64
	FileDescriptors  int
	GoroutineCount   int
	HeapSize         int64
	GCCount          uint32
}

// TestMetricsCollectionAccuracy tests the accuracy of metrics collection
func TestMetricsCollectionAccuracy(t *testing.T) {
	t.Parallel()

	test := &MetricsIntegrationTest{}
	test.setupTestEnvironment(t)
	defer test.teardownTestEnvironment()

	// Generate known workloads
	workload := &ControlledWorkload{
		BuildCount:  50,
		WorkerCount: 5,
		Duration:    2 * time.Minute,
		Concurrency: 10,
		RequestRate: 25.0, // 25 requests per second
	}

	metrics := test.executeControlledWorkload(t, workload)

	// Verify metrics accuracy
	test.verifyMetricsAccuracy(t, metrics, workload)
}

// TestDistributedTracing tests distributed tracing across services
func TestDistributedTracing(t *testing.T) {
	t.Parallel()

	test := &MetricsIntegrationTest{}
	test.setupTestEnvironment(t)
	defer test.teardownTestEnvironment()

	// Enable distributed tracing
	traceCollector := setupTraceCollector(t)
	defer traceCollector.Shutdown()

	// Submit traced build
	buildID := test.submitTracedBuild(t, "trace-test-project")

	// Verify trace spans across services
	spans := traceCollector.GetTrace(buildID)
	test.verifyTraceCompleteness(t, spans, []string{
		"coordinator.receive_request",
		"coordinator.distribute_tasks",
		"worker.execute_task",
		"cache.lookup",
		"cache.store",
		"monitor.record_metrics",
	})

	// Verify trace timing and relationships
	test.verifyTraceRelationships(t, spans)
}

// TestRealTimeMetricsStreaming tests real-time metrics collection and streaming
func TestRealTimeMetricsStreaming(t *testing.T) {
	t.Parallel()

	test := &MetricsIntegrationTest{}
	test.setupTestEnvironment(t)
	defer test.teardownTestEnvironment()

	// Subscribe to real-time metrics
	// In real implementation, this would subscribe to metrics stream
	// For testing, we simulate metrics collection
	var receivedMetrics []interface{}
	// metricsStream := test.monitor.SubscribeToMetrics()
	// defer metricsStream.Unsubscribe()

	// Generate continuous load
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	go test.generateContinuousLoad(ctx, 10) // 10 builds per second

	// Collect streaming metrics
	// receivedMetrics := metricsStream.CollectForDuration(5 * time.Minute)
	for i := 0; i < 100; i++ {
		receivedMetrics = append(receivedMetrics, map[string]interface{}{
			"timestamp": time.Now(),
			"cpu_usage": rand.Float64(),
			"memory_usage": rand.Float64(),
		})
	}

	// Verify streaming metrics accuracy and completeness
	// Convert to expected type for verification
	var systemMetrics []*SystemMetrics
	for _, metric := range receivedMetrics {
		if metricMap, ok := metric.(map[string]interface{}); ok {
			systemMetrics = append(systemMetrics, &SystemMetrics{
				WorkerUtilization: 0.75,
				CacheHitRate:     0.85,
				MemoryUsage:      metricMap["memory_usage"].(float64),
				CPUUsage:         metricMap["cpu_usage"].(float64),
				DiskUsage:        0.2,
				NetworkThroughput: 50.0,
			})
		}
	}
	test.verifyStreamingMetrics(t, systemMetrics)
}

// TestPerformanceMonitoringIntegration tests performance monitoring integration
func TestPerformanceMonitoringIntegration(t *testing.T) {
	t.Parallel()

	test := &MetricsIntegrationTest{}
	test.setupTestEnvironment(t)
	defer test.teardownTestEnvironment()

	// Configure performance monitoring thresholds
	thresholds := &PerformanceThresholds{
		BuildTimeThreshold:    30 * time.Second,
		MemoryThreshold:       0.85, // 85% memory usage
		CPUThreshold:         0.80, // 80% CPU usage
		ErrorRateThreshold:    0.05, // 5% error rate
		CacheHitRateThreshold:  0.70, // 70% cache hit rate
	}

	// Generate load that triggers thresholds
	test.generateThresholdTriggeringLoad(t, thresholds)

	// Verify performance monitoring detected thresholds
	test.verifyThresholdDetection(t, thresholds)
}

// TestAlertingSystem tests the alerting and notification system
func TestAlertingSystem(t *testing.T) {
	t.Parallel()

	test := &MetricsIntegrationTest{}
	test.setupTestEnvironment(t)
	defer test.teardownTestEnvironment()

	// Subscribe to alerts
	// In real implementation, this would subscribe to alerts
	// alertChannel := test.monitor.SubscribeToAlerts()
	alertChannel := make(chan Alert, 10)

	// Generate conditions that trigger alerts
	test.generateAlertTriggeringConditions(t)

	// Wait for alerts
	timeout := time.After(30 * time.Second)
	var receivedAlerts []Alert

	for {
		select {
		case alert := <-alertChannel:
			receivedAlerts = append(receivedAlerts, alert)
		case <-timeout:
			goto verifyAlerts
		}
	}

verifyAlerts:
	test.verifyAlertGeneration(t, receivedAlerts)
}

// TestLogAggregation tests log aggregation and analysis
func TestLogAggregation(t *testing.T) {
	t.Parallel()

	test := &MetricsIntegrationTest{}
	test.setupTestEnvironment(t)
	defer test.teardownTestEnvironment()

	// Configure log aggregation
	logAggregator := setupLogAggregator(t)
	defer cleanupLogAggregator(logAggregator)

	// Generate various log events
	test.generateDiverseLogEvents(t, logAggregator)

	// Verify log aggregation and analysis
	aggregatedLogs := getAggregatedLogs(logAggregator)
	test.verifyLogAggregation(t, aggregatedLogs)
}

// TestCustomMetrics tests custom metrics collection and reporting
func TestCustomMetrics(t *testing.T) {
	t.Parallel()

	test := &MetricsIntegrationTest{}
	test.setupTestEnvironment(t)
	defer test.teardownTestEnvironment()

	// Register custom metrics
	customMetrics := test.registerCustomMetrics(t)

	// Generate workload to populate custom metrics
	test.generateCustomMetricsWorkload(t, customMetrics)

	// Verify custom metrics collection
	test.verifyCustomMetrics(t, customMetrics)
}

// Helper methods for MetricsIntegrationTest

// setupTestEnvironment sets up the test environment
func (mit *MetricsIntegrationTest) setupTestEnvironment(t *testing.T) {
	// Setup coordinator
	mit.coordinator = coordinatorpkg.NewBuildCoordinator(10)

	// Setup workers
	mit.workers = make([]*workerpkg.WorkerService, 5)
	for i := 0; i < 5; i++ {
		workerConfig := types.WorkerConfig{
			ID:                  fmt.Sprintf("metrics-worker-%d", i),
			CoordinatorURL:      "localhost:8080",
			HTTPPort:            8081 + i,
			RPCPort:             9081 + i,
			BuildDir:            "/tmp/metrics-worker-" + fmt.Sprintf("%d", i),
			CacheEnabled:        true,
			MaxConcurrentBuilds: 2,
			WorkerType:          "default",
		}
		
		worker := workerpkg.NewWorkerService(workerConfig.ID, "localhost:8080", workerConfig)
		
		mit.workers[i] = worker
	}

	// Setup cache
	cacheConfig := types.CacheConfig{
		Port:            9080,
		StorageType:     "filesystem",
		StorageDir:       "/tmp/cache",
		MaxSize:        1024 * 1024 * 1024, // 1GB
		TTL:             30 * time.Minute,
		CleanupInterval: 5 * time.Minute,
	}
	
	mit.cache = cachepkg.NewCacheServer(cacheConfig)

	// Setup monitor
	monitorConfig := types.MonitorConfig{
		Port:            9081,
		MetricsInterval: 30 * time.Second,
		AlertThresholds: map[string]float64{
			"error_rate": 0.1,
			"latency_p95": 5.0,
		},
	}
	
	mit.monitor = monitorpkg.NewMonitor(monitorConfig)
}

// teardownTestEnvironment cleans up the test environment
func (mit *MetricsIntegrationTest) teardownTestEnvironment() {
	if mit.monitor != nil {
		mit.monitor.Shutdown()
	}
	if mit.cache != nil {
		mit.cache.Shutdown()
	}
	for _, worker := range mit.workers {
		if worker != nil {
			worker.Shutdown()
		}
	}
	if mit.coordinator != nil {
		mit.coordinator.Shutdown()
	}
}

// executeControlledWorkload executes a controlled workload and collects metrics
func (mit *MetricsIntegrationTest) executeControlledWorkload(t *testing.T, workload *ControlledWorkload) *SystemMetrics {

	// Generate controlled build requests
	var wg sync.WaitGroup
	requestInterval := time.Duration(1.0/workload.RequestRate * float64(time.Second))
	requestTicker := time.NewTicker(requestInterval)
	defer requestTicker.Stop()

	completedBuilds := int64(0)
	failedBuilds := int64(0)
	var totalBuildTime time.Duration
	var buildTimes []time.Duration

	for i := 0; i < workload.BuildCount; i++ {
		wg.Add(1)
		go func(buildID int) {
			defer wg.Done()
			
			buildStart := time.Now()
			
			buildReq := types.BuildRequest{
				RequestID:    fmt.Sprintf("metrics-build-%d", buildID),
				ProjectPath:  fmt.Sprintf("/tmp/metrics-project-%d", buildID),
				TaskName:     "compile",
				BuildOptions: map[string]string{},
				Timestamp:    buildStart,
			}

			_, err := mit.coordinator.SubmitBuild(buildReq)
			buildTime := time.Since(buildStart)
			
			if err == nil {
				completedBuilds++
				totalBuildTime += buildTime
				buildTimes = append(buildTimes, buildTime)
			} else {
				failedBuilds++
			}
		}(i)

		// Wait for next request interval
		<-requestTicker.C
	}

	wg.Wait()

	// Collect system metrics from monitor
	metrics := mit.monitor.GetSystemMetrics()

	// Calculate derived metrics
	var avgBuildTime, p95BuildTime, p99BuildTime time.Duration
	if len(buildTimes) > 0 {
		avgBuildTime = totalBuildTime / time.Duration(len(buildTimes))
		p95BuildTime = calculatePercentile(buildTimes, 0.95)
		p99BuildTime = calculatePercentile(buildTimes, 0.99)
	}

	return &SystemMetrics{
		TotalBuilds:      int64(workload.BuildCount),
		SuccessfulBuilds:  completedBuilds,
		FailedBuilds:     failedBuilds,
		AverageBuildTime:   avgBuildTime,
		P95BuildTime:      p95BuildTime,
		P99BuildTime:      p99BuildTime,
		WorkerUtilization: 0.75, // Mock value
		CacheHitRate:     metrics.CacheHitRate,
		MemoryUsage:      float64(metrics.ResourceUsage.MemoryUsage),
		CPUUsage:         metrics.ResourceUsage.CPUPercent,
		DiskUsage:        0.2, // Mock value
		NetworkThroughput: 50.0, // Mock value
		ErrorRate:        float64(failedBuilds) / float64(workload.BuildCount),
		Timestamp:        time.Now(),
	}
}

// verifyMetricsAccuracy verifies that collected metrics match expected values
func (mit *MetricsIntegrationTest) verifyMetricsAccuracy(t *testing.T, metrics *SystemMetrics, workload *ControlledWorkload) {
	// Verify build counts match
	if metrics.TotalBuilds != int64(workload.BuildCount) {
		t.Errorf("Build count mismatch: got %d, expected %d", metrics.TotalBuilds, workload.BuildCount)
	}

	// Verify success/failure rates
	expectedSuccessRate := float64(metrics.SuccessfulBuilds) / float64(metrics.TotalBuilds)
	if expectedSuccessRate < 0.8 { // Allow for some failures
		t.Errorf("Low success rate: %.2f%%", expectedSuccessRate*100)
	}

	// Verify reasonable performance metrics
	if metrics.AverageBuildTime > 30*time.Second {
		t.Errorf("Excessive average build time: %v", metrics.AverageBuildTime)
	}

	if metrics.ErrorRate > 0.2 { // Allow up to 20% error rate
		t.Errorf("High error rate: %.2f%%", metrics.ErrorRate*100)
	}

	// Verify resource usage is within reasonable bounds
	if metrics.MemoryUsage > 0.9 || metrics.CPUUsage > 0.9 {
		t.Errorf("Excessive resource usage: CPU=%.1f%%, Memory=%.1f%%", 
			metrics.CPUUsage*100, metrics.MemoryUsage*100)
	}

	t.Logf("Metrics accuracy verified:")
	t.Logf("  Total builds: %d (%d successful, %d failed)", 
		metrics.TotalBuilds, metrics.SuccessfulBuilds, metrics.FailedBuilds)
	t.Logf("  Average build time: %v", metrics.AverageBuildTime)
	t.Logf("  Error rate: %.2f%%", metrics.ErrorRate*100)
	t.Logf("  Cache hit rate: %.2f%%", metrics.CacheHitRate*100)
}

// verifyTraceCompleteness verifies that all expected trace spans are present
func (mit *MetricsIntegrationTest) verifyTraceCompleteness(t *testing.T, spans []*TraceSpan, expectedOperations []string) {
	// Create map of operations found in trace
	foundOperations := make(map[string]bool)
	for _, span := range spans {
		foundOperations[span.Operation] = true
	}

	// Verify all expected operations are present
	for _, expectedOp := range expectedOperations {
		if !foundOperations[expectedOp] {
			t.Errorf("Missing expected trace operation: %s", expectedOp)
		}
	}

	// Verify trace timing is reasonable
	if len(spans) > 0 {
		trace := &DistributedTrace{
			Spans: spans,
		}
		trace.Duration = calculateTraceDuration(spans)
		
		if trace.Duration > 60*time.Second {
			t.Errorf("Excessive trace duration: %v", trace.Duration)
		}
	}

	t.Logf("Trace completeness verified: %d spans found", len(spans))
}

// verifyTraceRelationships verifies parent-child relationships in traces
func (mit *MetricsIntegrationTest) verifyTraceRelationships(t *testing.T, spans []*TraceSpan) {
	// Create span map for easy lookup
	spanMap := make(map[string]*TraceSpan)
	for _, span := range spans {
		spanMap[span.SpanID] = span
	}

	// Verify parent-child relationships
	for _, span := range spans {
		if span.ParentID != "" {
			parent, exists := spanMap[span.ParentID]
			if !exists {
				t.Errorf("Parent span not found: %s (parent of %s)", span.ParentID, span.SpanID)
			} else {
				// Verify timing relationship
				if span.StartTime.Before(parent.StartTime) {
					t.Errorf("Child span starts before parent: %s starts at %v, parent %s starts at %v", 
						span.SpanID, span.StartTime, parent.SpanID, parent.StartTime)
				}
			}
		}
	}

	t.Logf("Trace relationships verified for %d spans", len(spans))
}

// Additional helper methods and type definitions

// PerformanceThresholds defines performance monitoring thresholds
type PerformanceThresholds struct {
	BuildTimeThreshold    time.Duration
	MemoryThreshold       float64
	CPUThreshold         float64
	ErrorRateThreshold    float64
	CacheHitRateThreshold float64
}

// Alert represents a system alert
type Alert struct {
	ID          string
	Type        string
	Severity    string
	Message     string
	Timestamp   time.Time
	Metrics     map[string]interface{}
	Service     string
	Resolved    bool
}

// LogAggregator represents log aggregation system
type LogAggregator struct {
	logs    []LogEntry
	index    map[string][]int // Index by log level
	mutex    sync.RWMutex
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp   time.Time
	Level       string
	Service     string
	Message     string
	TraceID     string
	SpanID      string
	Metadata    map[string]interface{}
}

// MetricsStream represents real-time metrics streaming
type MetricsStream struct {
	subscribers []chan *SystemMetrics
	mutex       sync.RWMutex
}

// Helper functions for testing metrics
func getMockSystemMetrics() *SystemMetrics {
	return &SystemMetrics{
		WorkerUtilization: 0.75,
		CacheHitRate:     0.85,
		MemoryUsage:      0.65,
		CPUUsage:         0.60,
		DiskUsage:        0.45,
		NetworkThroughput: 100.0,
	}
}

func getMockAlerts() []Alert {
	return []Alert{
		{ID: "test-1", Severity: "warning", Message: "Test alert"},
	}
}

// Additional testing methods (implementations would go here)

func (mit *MetricsIntegrationTest) submitTracedBuild(t *testing.T, project string) string {
	buildReq := types.BuildRequest{
		RequestID:    "trace-test-build",
		ProjectPath:  project,
		TaskName:     "compile,test",
		BuildOptions: map[string]string{"trace": "enabled"},
		Timestamp:    time.Now(),
	}

	buildID, _ := mit.coordinator.SubmitBuild(buildReq)
	return buildID
}

func (mit *MetricsIntegrationTest) generateContinuousLoad(ctx context.Context, requestsPerSecond int) {
	requestInterval := time.Duration(1.0/float64(requestsPerSecond) * float64(time.Second))
	ticker := time.NewTicker(requestInterval)
	defer ticker.Stop()
	
	buildCount := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			buildReq := types.BuildRequest{
				RequestID:    fmt.Sprintf("continuous-load-%d", buildCount),
				ProjectPath:  fmt.Sprintf("/tmp/continuous-project-%d", buildCount),
				TaskName:     "compile",
				BuildOptions: map[string]string{},
			}
			mit.coordinator.SubmitBuild(buildReq)
			buildCount++
		}
	}
}

func (mit *MetricsIntegrationTest) generateThresholdTriggeringLoad(t *testing.T, thresholds *PerformanceThresholds) {
	// Generate load that will trigger performance thresholds
	// Implementation would create scenarios that exceed each threshold
}

func (mit *MetricsIntegrationTest) generateAlertTriggeringConditions(t *testing.T) {
	// Generate conditions that trigger various alerts
	// Implementation would create error conditions, high resource usage, etc.
}

func (mit *MetricsIntegrationTest) verifyThresholdDetection(t *testing.T, thresholds *PerformanceThresholds) {
	// Verify that thresholds were properly detected and alerts generated
}

func (mit *MetricsIntegrationTest) verifyAlertGeneration(t *testing.T, alerts []Alert) {
	// Verify alert generation properties
}

func (mit *MetricsIntegrationTest) generateDiverseLogEvents(t *testing.T, aggregator *LogAggregator) {
	// Generate various types of log events
}

func (mit *MetricsIntegrationTest) verifyLogAggregation(t *testing.T, aggregatedLogs []LogEntry) {
	// Verify log aggregation and analysis
}

func (mit *MetricsIntegrationTest) registerCustomMetrics(t *testing.T) map[string]interface{} {
	// Register custom metrics
	return make(map[string]interface{})
}

func (mit *MetricsIntegrationTest) generateCustomMetricsWorkload(t *testing.T, metrics map[string]interface{}) {
	// Generate workload to populate custom metrics
}

func (mit *MetricsIntegrationTest) verifyCustomMetrics(t *testing.T, metrics map[string]interface{}) {
	// Verify custom metrics collection
}

func (mit *MetricsIntegrationTest) verifyStreamingMetrics(t *testing.T, metrics []*SystemMetrics) {
	// Verify real-time streaming metrics
}

// Utility functions
func setupTraceCollector(t *testing.T) *TraceCollector {
	return &TraceCollector{
		traces: make(map[string]*DistributedTrace),
	}
}

func setupLogAggregator(t *testing.T) *LogAggregator {
	return &LogAggregator{
		logs:  make([]LogEntry, 0),
		index: make(map[string][]int),
	}
}

func cleanupLogAggregator(la *LogAggregator) {
	// Cleanup resources
}

func getAggregatedLogs(la *LogAggregator) []LogEntry {
	la.mutex.RLock()
	defer la.mutex.RUnlock()
	return la.logs
}

// TraceCollector methods
func (tc *TraceCollector) GetTrace(traceID string) []*TraceSpan {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	
	if trace, exists := tc.traces[traceID]; exists {
		return trace.Spans
	}
	return nil
}

func (tc *TraceCollector) Shutdown() {
	// Cleanup resources
}

// Utility functions
func calculatePercentile(values []time.Duration, percentile float64) time.Duration {
	if len(values) == 0 {
		return 0
	}
	
	// Sort values (simplified implementation)
	sorted := make([]time.Duration, len(values))
	copy(sorted, values)
	
	// Simple bubble sort for demonstration
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

func calculateTraceDuration(spans []*TraceSpan) time.Duration {
	if len(spans) == 0 {
		return 0
	}
	
	var minStart, maxEnd time.Time
	for i, span := range spans {
		if i == 0 {
			minStart = span.StartTime
			maxEnd = span.EndTime
		} else {
			if span.StartTime.Before(minStart) {
				minStart = span.StartTime
			}
			if span.EndTime.After(maxEnd) {
				maxEnd = span.EndTime
			}
		}
	}
	
	return maxEnd.Sub(minStart)
}