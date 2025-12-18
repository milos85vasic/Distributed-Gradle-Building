package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScalabilityUnderLoad tests system behavior with increasing concurrent loads
func TestScalabilityUnderLoad(t *testing.T) {
	// Test coordinator with different concurrent levels
	concurrentLevels := []int{1, 2, 4, 8, 16, 32}
	
	for _, concurrency := range concurrentLevels {
		t.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(t *testing.T) {
			// Setup test coordinator
			coordinator := setupTestCoordinator(t)
			defer coordinator.Stop()
			
			// Submit concurrent builds
			buildIDs := make([]string, concurrency)
			var wg sync.WaitGroup
			
			startTime := time.Now()
			
			// Submit builds concurrently
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					
					project := createTestProject(t, fmt.Sprintf("scalability-test-%d", index))
					buildReq := BuildRequest{
						ProjectPath: project.Path,
						TaskName:    "build",
						CacheEnabled: true,
					}
					
					buildID, err := coordinator.SubmitBuild(buildReq)
					require.NoError(t, err)
					buildIDs[index] = buildID
				}(i)
			}
			
			wg.Wait()
			submissionTime := time.Since(startTime)
			
			// Wait for all builds to complete
			completedBuilds := 0
			maxWaitTime := 5 * time.Minute
			checkInterval := 5 * time.Second
			deadline := time.Now().Add(maxWaitTime)
			
			for time.Now().Before(deadline) && completedBuilds < concurrency {
				time.Sleep(checkInterval)
				completedBuilds = 0
				
				for _, buildID := range buildIDs {
					if buildID != "" {
						status := coordinator.GetBuildStatus(buildID)
						if status.Status == "completed" || status.Status == "failed" {
							completedBuilds++
						}
					}
				}
			}
			
			// Collect performance metrics
			totalTime := time.Since(startTime)
			avgTime := totalTime / time.Duration(concurrency)
			throughput := float64(concurrency) / totalTime.Seconds()
			
			// Record scalability metrics
			scalabilityMetric := ScalabilityMetric{
				ConcurrencyLevel:     concurrency,
				SubmissionTime:       submissionTime,
				TotalExecutionTime:   totalTime,
				AverageBuildTime:     avgTime,
				CompletedBuilds:      completedBuilds,
				FailedBuilds:         concurrency - completedBuilds,
				ThroughputPerSecond:  throughput,
			}
			
			// Assert scalability expectations
			assert.Equal(t, concurrency, completedBuilds, "All builds should complete")
			assert.Greater(t, throughput, 0.1, "Throughput should be reasonable")
			
			// For higher concurrency, total time shouldn't increase linearly
			if concurrency > 4 {
				linearTime := submissionTime * time.Duration(concurrency)
				assert.Less(t, totalTime, linearTime, "System should show concurrency benefits")
			}
			
			t.Logf("Concurrency %d: Completed=%d, Throughput=%.2f builds/sec, Total time=%v", 
				concurrency, completedBuilds, throughput, totalTime)
		})
	}
}

// TestResourceScaling tests resource utilization as load increases
func TestResourceScaling(t *testing.T) {
	coordinator := setupTestCoordinator(t)
	defer coordinator.Stop()
	
	// Monitor resource usage during scaling test
	resourceMonitor := NewResourceMonitor()
	defer resourceMonitor.Stop()
	
	// Incrementally increase load
	maxConcurrency := 16
	buildInterval := 2 * time.Second
	testDuration := 2 * time.Minute
	
	startTime := time.Now()
	currentBuilds := 0
	totalBuilds := 0
	
	for time.Since(startTime) < testDuration {
		if currentBuilds < maxConcurrency {
			// Submit new build
			project := createTestProject(t, fmt.Sprintf("resource-test-%d", totalBuilds))
			buildReq := BuildRequest{
				ProjectPath: project.Path,
				TaskName:    "build",
				CacheEnabled: true,
			}
			
			buildID, err := coordinator.SubmitBuild(buildReq)
			require.NoError(t, err)
			
			currentBuilds++
			totalBuilds++
		} else {
			// Check for completed builds
			activeBuilds := 0
			for i := 0; i < totalBuilds; i++ {
				status := coordinator.GetBuildStatus(fmt.Sprintf("build-%d", i))
				if status.Status == "active" {
					activeBuilds++
				}
			}
			currentBuilds = activeBuilds
		}
		
		time.Sleep(buildInterval)
	}
	
	// Wait for remaining builds to complete
	time.Sleep(30 * time.Second)
	
	// Analyze resource usage
	metrics := resourceMonitor.GetMetrics()
	
	// Resource scaling expectations
	assert.Less(t, metrics.MaxCPUUsage, 90.0, "CPU usage should stay below 90%")
	assert.Less(t, metrics.MaxMemoryUsage, 80.0, "Memory usage should stay below 80%")
	assert.Greater(t, metrics.AvgCPUUsage, 10.0, "CPU usage should show reasonable utilization")
	
	t.Logf("Resource scaling metrics: MaxCPU=%.1f%%, MaxMem=%.1f%%, AvgCPU=%.1f%%, AvgMem=%.1f%%",
		metrics.MaxCPUUsage, metrics.MaxMemoryUsage, 
		metrics.AvgCPUUsage, metrics.AvgMemoryUsage)
}

// TestWorkerPoolScaling tests scaling of worker pool
func TestWorkerPoolScaling(t *testing.T) {
	workerCounts := []int{2, 4, 8, 16}
	
	for _, workerCount := range workerCounts {
		t.Run(fmt.Sprintf("Workers_%d", workerCount), func(t *testing.T) {
			// Setup coordinator with specified worker count
			config := CoordinatorConfig{
				WorkerPoolSize: workerCount,
				CacheEnabled:    true,
			}
			
			coordinator := setupTestCoordinatorWithConfig(t, config)
			defer coordinator.Stop()
			
			// Start worker pool
			workers := setupWorkerPool(t, workerCount)
			defer stopWorkerPool(workers)
			
			// Submit test builds
			buildCount := workerCount * 4 // 4 builds per worker
			buildIDs := make([]string, buildCount)
			
			startTime := time.Now()
			
			for i := 0; i < buildCount; i++ {
				project := createTestProject(t, fmt.Sprintf("worker-test-%d-%d", workerCount, i))
				buildReq := BuildRequest{
					ProjectPath: project.Path,
					TaskName:    "build",
					CacheEnabled: true,
				}
				
				buildID, err := coordinator.SubmitBuild(buildReq)
				require.NoError(t, err)
				buildIDs[i] = buildID
			}
			
			// Wait for completion
			completed := waitForBuilds(t, coordinator, buildIDs, 5*time.Minute)
			totalTime := time.Since(startTime)
			
			// Calculate scaling efficiency
			throughput := float64(completed) / totalTime.Seconds()
			throughputPerWorker := throughput / float64(workerCount)
			
			// Scaling expectations
			assert.Equal(t, buildCount, completed, "All builds should complete")
			assert.Greater(t, throughputPerWorker, 0.2, "Each worker should handle reasonable throughput")
			
			t.Logf("Workers %d: Throughput=%.2f builds/sec (%.2f per worker), Time=%v", 
				workerCount, throughput, throughputPerWorker, totalTime)
		})
	}
}

// TestCacheScaling tests cache performance under different sizes
func TestCacheScaling(t *testing.T) {
	cacheSizes := []int{100, 500, 1000, 2000} // MB
	
	for _, cacheSize := range cacheSizes {
		t.Run(fmt.Sprintf("Cache_%dMB", cacheSize), func(t *testing.T) {
			// Setup cache server
			cache := setupTestCache(t, cacheSize)
			defer cache.Stop()
			
			coordinator := setupTestCoordinator(t)
			defer coordinator.Stop()
			
			// Submit multiple builds with repeated patterns
			buildPatterns := 5
			buildsPerPattern := 10
			totalBuilds := buildPatterns * buildsPerPattern
			
			startTime := time.Now()
			
			for pattern := 0; pattern < buildPatterns; pattern++ {
				for i := 0; i < buildsPerPattern; i++ {
					project := createTestProject(t, fmt.Sprintf("cache-pattern-%d", pattern))
					buildReq := BuildRequest{
						ProjectPath: project.Path,
						TaskName:    "build",
						CacheEnabled: true,
					}
					
					_, err := coordinator.SubmitBuild(buildReq)
					require.NoError(t, err)
				}
			}
			
			// Wait for all builds to complete
			completed := waitForBuildCompletion(t, coordinator, totalBuilds, 5*time.Minute)
			totalTime := time.Since(startTime)
			
			// Get cache statistics
			stats := cache.GetStatistics()
			
			// Cache scaling expectations
			assert.Greater(t, stats.CacheHitRate, 30.0, "Cache hit rate should be reasonable")
			assert.Less(t, stats.AverageRetrievalTime, 100*time.Millisecond, "Cache retrieval should be fast")
			
			// Larger caches should have higher hit rates
			if cacheSize >= 500 {
				assert.Greater(t, stats.CacheHitRate, 50.0, "Larger cache should have better hit rate")
			}
			
			t.Logf("Cache %dMB: HitRate=%.1f%%, AvgRetrieval=%v, Builds=%d, Time=%v",
				cacheSize, stats.CacheHitRate, stats.AverageRetrievalTime, completed, totalTime)
		})
	}
}

// TestNetworkScaling tests network behavior under high load
func TestNetworkScaling(t *testing.T) {
	// Setup HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate build request processing
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"build_id": "test-build-%d"}`, time.Now().UnixNano())
	}))
	defer server.Close()
	
	// Test different concurrent request levels
	concurrentRequests := []int{10, 50, 100, 200}
	
	for _, concurrency := range concurrentRequests {
		t.Run(fmt.Sprintf("Network_Requests_%d", concurrency), func(t *testing.T) {
			var wg sync.WaitGroup
			successfulRequests := 0
			failedRequests := 0
			responseTimes := make([]time.Duration, concurrency)
			
			startTime := time.Now()
			
			// Send concurrent requests
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					
					requestStart := time.Now()
					
					resp, err := http.Post(server.URL, "application/json", strings.NewReader("{}"))
					if err != nil {
						failedRequests++
						return
					}
					defer resp.Body.Close()
					
					if resp.StatusCode == http.StatusOK {
						successfulRequests++
					} else {
						failedRequests++
					}
					
					responseTimes[index] = time.Since(requestStart)
				}(i)
			}
			
			wg.Wait()
			totalTime := time.Since(startTime)
			
			// Calculate network metrics
			avgResponseTime := calculateAverage(responseTimes)
			maxResponseTime := calculateMax(responseTimes)
			throughput := float64(concurrency) / totalTime.Seconds()
			
			// Network scaling expectations
			assert.Greater(t, float64(successfulRequests), float64(concurrency)*0.95, "At least 95% of requests should succeed")
			assert.Less(t, avgResponseTime, 5*time.Second, "Average response time should be reasonable")
			assert.Less(t, maxResponseTime, 10*time.Second, "Maximum response time should be acceptable")
			
			t.Logf("Network %d requests: Success=%d, Failed=%d, AvgTime=%v, MaxTime=%v, Throughput=%.2f req/sec",
				concurrency, successfulRequests, failedRequests, avgResponseTime, maxResponseTime, throughput)
		})
	}
}

// TestDatabaseScaling tests database performance under load
func TestDatabaseScaling(t *testing.T) {
	// Setup test database
	db := setupTestDatabase(t)
	defer db.Close()
	
	coordinator := setupTestCoordinatorWithDB(t, db)
	defer coordinator.Stop()
	
	// Test database operation scaling
	recordCounts := []int{100, 1000, 5000, 10000}
	
	for _, recordCount := range recordCounts {
		t.Run(fmt.Sprintf("DB_Records_%d", recordCount), func(t *testing.T) {
			// Insert test records
			startTime := time.Now()
			
			for i := 0; i < recordCount; i++ {
				buildRecord := BuildRecord{
					ID:         fmt.Sprintf("build-%d-%d", recordCount, i),
					Project:    fmt.Sprintf("test-project-%d", i%10),
					Status:     "completed",
					StartTime:  time.Now().Add(-time.Hour),
					EndTime:    time.Now(),
					Duration:   time.Duration(i%60) * time.Second,
				}
				
				err := db.InsertBuildRecord(buildRecord)
				require.NoError(t, err)
			}
			
			insertTime := time.Since(startTime)
			
			// Test query performance
			queryStart := time.Now()
			
			// Simulate various queries
			for i := 0; i < 100; i++ {
				_, err := db.GetBuildRecord(fmt.Sprintf("build-%d-%d", recordCount, i%recordCount))
				require.NoError(t, err)
				
				_, err = db.GetBuildsByStatus("completed", 100, 0)
				require.NoError(t, err)
			}
			
			queryTime := time.Since(queryStart)
			avgQueryTime := queryTime / 200 // 100 GetBuildRecord + 100 GetBuildsByStatus calls
			
			// Database scaling expectations
			assert.Less(t, insertTime, time.Duration(recordCount)*time.Millisecond, "Insert time should be linear")
			assert.Less(t, avgQueryTime, 100*time.Millisecond, "Query time should be reasonable")
			
			// Test cleanup
			cleanupStart := time.Now()
			err := db.DeleteBuildsByStatus("completed")
			require.NoError(t, err)
			cleanupTime := time.Since(cleanupStart)
			
			t.Logf("DB %d records: Insert=%v, AvgQuery=%v, Cleanup=%v",
				recordCount, insertTime, avgQueryTime, cleanupTime)
		})
	}
}

// Helper functions and data structures

type ScalabilityMetric struct {
	ConcurrencyLevel    int
	SubmissionTime     time.Duration
	TotalExecutionTime time.Duration
	AverageBuildTime   time.Duration
	CompletedBuilds    int
	FailedBuilds       int
	ThroughputPerSecond float64
}

type ResourceMetrics struct {
	MaxCPUUsage    float64
	AvgCPUUsage    float64
	MaxMemoryUsage float64
	AvgMemoryUsage float64
	SampleCount    int
}

type ResourceMonitor struct {
	metrics chan ResourceMetrics
	done    chan bool
}

func NewResourceMonitor() *ResourceMonitor {
	rm := &ResourceMonitor{
		metrics: make(chan ResourceMetrics, 100),
		done:    make(chan bool),
	}
	
	go rm.monitor()
	return rm
}

func (rm *ResourceMonitor) monitor() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	var samples []ResourceMetrics
	
	for {
		select {
		case <-ticker.C:
			// Collect system metrics (implementation depends on platform)
			cpuUsage := getCPUUsage()
			memoryUsage := getMemoryUsage()
			
			samples = append(samples, ResourceMetrics{
				MaxCPUUsage:    cpuUsage,
				AvgCPUUsage:    cpuUsage,
				MaxMemoryUsage: memoryUsage,
				AvgMemoryUsage: memoryUsage,
				SampleCount:    1,
			})
			
		case <-rm.done:
			// Calculate final metrics
			rm.calculateFinalMetrics(samples)
			return
		}
	}
}

func (rm *ResourceMonitor) GetMetrics() ResourceMetrics {
	// Return calculated metrics
	select {
	case metrics := <-rm.metrics:
		return metrics
	default:
		return ResourceMetrics{}
	}
}

func (rm *ResourceMonitor) Stop() {
	close(rm.done)
}

func (rm *ResourceMonitor) calculateFinalMetrics(samples []ResourceMetrics) {
	if len(samples) == 0 {
		return
	}
	
	var maxCPU, avgCPU, maxMem, avgMem float64
	
	for _, sample := range samples {
		if sample.MaxCPUUsage > maxCPU {
			maxCPU = sample.MaxCPUUsage
		}
		if sample.MaxMemoryUsage > maxMem {
			maxMem = sample.MaxMemoryUsage
		}
		avgCPU += sample.AvgCPUUsage
		avgMem += sample.AvgMemoryUsage
	}
	
	avgCPU /= float64(len(samples))
	avgMem /= float64(len(samples))
	
	finalMetrics := ResourceMetrics{
		MaxCPUUsage:    maxCPU,
		AvgCPUUsage:    avgCPU,
		MaxMemoryUsage: maxMem,
		AvgMemoryUsage: avgMem,
		SampleCount:    len(samples),
	}
	
	select {
	case rm.metrics <- finalMetrics:
	default:
	}
}

// Helper functions
func calculateAverage(times []time.Duration) time.Duration {
	if len(times) == 0 {
		return 0
	}
	
	var total time.Duration
	for _, t := range times {
		total += t
	}
	return total / time.Duration(len(times))
}

func calculateMax(times []time.Duration) time.Duration {
	if len(times) == 0 {
		return 0
	}
	
	max := times[0]
	for _, t := range times[1:] {
		if t > max {
			max = t
		}
	}
	return max
}

func waitForBuilds(t *testing.T, coordinator *TestCoordinator, buildIDs []string, timeout time.Duration) int {
	deadline := time.Now().Add(timeout)
	completed := 0
	
	for time.Now().Before(deadline) {
		for i, buildID := range buildIDs {
			if buildID != "" {
				status := coordinator.GetBuildStatus(buildID)
				if status.Status == "completed" || status.Status == "failed" {
					completed++
					buildIDs[i] = "" // Mark as processed
				}
			}
		}
		
		if completed == len(buildIDs) {
			break
		}
		
		time.Sleep(1 * time.Second)
	}
	
	return completed
}

func getCPUUsage() float64 {
	// Placeholder implementation - in real scenario would use system monitoring
	return 25.0 + float64(time.Now().UnixNano()%50) // Simulated CPU usage
}

func getMemoryUsage() float64 {
	// Placeholder implementation - in real scenario would use system monitoring
	return 40.0 + float64(time.Now().UnixNano()%30) // Simulated memory usage
}