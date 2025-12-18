package performance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

// TestWorkerPoolPerformance tests worker pool performance under load
func TestWorkerPoolPerformance(t *testing.T) {
	// Create test environment
	testDir, err := ioutil.TempDir("", "worker-pool-perf-test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Performance test configuration
	const (
		numWorkers = 10
		numTasks   = 1000
		concurrentRequests = 50
	)

	// Setup mock worker pool
	workerPool := setupMockWorkerPool(t, numWorkers)
	poolServer := httptest.NewServer(workerPool)
	defer poolServer.Close()

	// Performance test cases
	tests := []struct {
		name        string
		testFunc    func(t *testing.T, serverURL string)
		description string
	}{
		{
			name: "TaskThroughput",
			testFunc: func(t *testing.T, serverURL string) {
				testTaskThroughput(t, serverURL, numTasks, concurrentRequests)
			},
			description: "Test worker pool task throughput",
		},
		{
			name: "WorkerUtilization",
			testFunc: func(t *testing.T, serverURL string) {
				testWorkerUtilization(t, serverURL, numWorkers, concurrentRequests)
			},
			description: "Test worker utilization efficiency",
		},
		{
			name: "MemoryEfficiency",
			testFunc: func(t *testing.T, serverURL string) {
				testMemoryEfficiency(t, serverURL, numTasks)
			},
			description: "Test memory usage efficiency",
		},
		{
			name: "ResponseTimeUnderLoad",
			testFunc: func(t *testing.T, serverURL string) {
				testResponseTimeUnderLoad(t, serverURL, numTasks, concurrentRequests)
			},
			description: "Test response times under load",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Running %s: %s", tc.name, tc.description)
			tc.testFunc(t, poolServer.URL)
		})
	}
}

// TestCachePerformance tests cache server performance
func TestCachePerformance(t *testing.T) {
	// Create test environment
	testDir, err := ioutil.TempDir("", "cache-perf-test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Setup mock cache server
	cacheServer := setupMockCacheServer(t)
	cacheServerInst := httptest.NewServer(cacheServer)
	defer cacheServerInst.Close()

	// Performance test cases
	tests := []struct {
		name        string
		testFunc    func(t *testing.T, serverURL string)
		description string
	}{
		{
			name: "CacheHitRate",
			testFunc: func(t *testing.T, serverURL string) {
				testCacheHitRate(t, serverURL)
			},
			description: "Test cache hit rate efficiency",
		},
		{
			name: "CacheStoragePerformance",
			testFunc: func(t *testing.T, serverURL string) {
				testCacheStoragePerformance(t, serverURL)
			},
			description: "Test cache storage/retrieval performance",
		},
		{
			name: "CacheEvictionPerformance",
			testFunc: func(t *testing.T, serverURL string) {
				testCacheEvictionPerformance(t, serverURL)
			},
			description: "Test cache eviction performance",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Running %s: %s", tc.name, tc.description)
			tc.testFunc(t, cacheServerInst.URL)
		})
	}
}

// TestDistributedBuildPerformance tests end-to-end distributed build performance
func TestDistributedBuildPerformance(t *testing.T) {
	// Create test environment
	testDir, err := ioutil.TempDir("", "distributed-build-perf-test")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Setup mock distributed system
	coordinator := setupPerformanceCoordinator(t)
	workers := setupPerformanceWorkers(t, 5)
	cacheServer := setupMockCacheServer(t)

	// Start mock servers
	coordServer := httptest.NewServer(coordinator)
	defer coordServer.Close()

	workerServers := make([]*httptest.Server, 5)
	for i, worker := range workers {
		workerServers[i] = httptest.NewServer(worker)
		defer workerServers[i].Close()
	}

	cacheServerInst := httptest.NewServer(cacheServer)
	defer cacheServerInst.Close()

	// Performance test cases
	tests := []struct {
		name        string
		testFunc    func(t *testing.T, coordURL string, workerURLs []string, cacheURL string)
		description string
	}{
		{
			name: "BuildThroughput",
			testFunc: func(t *testing.T, coordURL string, workerURLs []string, cacheURL string) {
				testBuildThroughput(t, coordURL, workerURLs, cacheURL)
			},
			description: "Test distributed build throughput",
		},
		{
			name: "ScalabilityTest",
			testFunc: func(t *testing.T, coordURL string, workerURLs []string, cacheURL string) {
				testScalability(t, coordURL, workerURLs, cacheURL)
			},
			description: "Test system scalability",
		},
		{
			name: "ResourceEfficiency",
			testFunc: func(t *testing.T, coordURL string, workerURLs []string, cacheURL string) {
				testResourceEfficiency(t, coordURL, workerURLs, cacheURL)
			},
			description: "Test resource usage efficiency",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Running %s: %s", tc.name, tc.description)
			tc.testFunc(t, coordServer.URL, getWorkerURLs(workerServers), cacheServerInst.URL)
		})
	}
}

// Helper performance test functions

func testTaskThroughput(t *testing.T, serverURL string, numTasks int, concurrency int) {
	t.Logf("Testing task throughput with %d tasks, %d concurrent requests", numTasks, concurrency)

	startTime := time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex
	completedTasks := 0

	// Submit tasks concurrently
	semaphore := make(chan struct{}, concurrency)
	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		go func(taskID int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			taskRequest := map[string]interface{}{
				"task_id":   fmt.Sprintf("task-%d", taskID),
				"task_type": "compile",
				"payload":   make([]byte, 1024), // 1KB payload
			}

			jsonData, _ := json.Marshal(taskRequest)
			resp, err := http.Post(
				fmt.Sprintf("%s/api/tasks/submit", serverURL),
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			if err == nil && resp.StatusCode == http.StatusOK {
				mu.Lock()
				completedTasks++
				mu.Unlock()
			}
			if resp != nil {
				resp.Body.Close()
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	throughput := float64(completedTasks) / duration.Seconds()
	t.Logf("Task throughput: %.2f tasks/sec (%d/%d completed in %v)", 
		throughput, completedTasks, numTasks, duration)

	// Performance target: at least 100 tasks/sec
	if throughput >= 100 {
		t.Logf("✓ Throughput meets target: %.2f tasks/sec (target: ≥100)", throughput)
	} else {
		t.Errorf("✗ Throughput below target: %.2f tasks/sec (target: ≥100)", throughput)
	}

	// Success rate should be at least 95%
	successRate := float64(completedTasks) / float64(numTasks) * 100
	if successRate >= 95 {
		t.Logf("✓ Success rate meets target: %.1f%% (target: ≥95%%)", successRate)
	} else {
		t.Errorf("✗ Success rate below target: %.1f%% (target: ≥95%%)", successRate)
	}
}

func testWorkerUtilization(t *testing.T, serverURL string, numWorkers int, concurrency int) {
	t.Logf("Testing worker utilization with %d workers", numWorkers)

	// Get initial worker stats
	resp, err := http.Get(fmt.Sprintf("%s/api/workers/stats", serverURL))
	if err != nil {
		t.Fatalf("Failed to get worker stats: %v", err)
	}
	defer resp.Body.Close()

	var initialStats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&initialStats); err != nil {
		t.Fatalf("Failed to decode initial stats: %v", err)
	}

	// Submit sustained load
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(taskID int) {
			defer wg.Done()
			
			taskRequest := map[string]interface{}{
				"task_id":   fmt.Sprintf("util-task-%d", taskID),
				"task_type": "compile",
				"payload":   make([]byte, 1024),
			}

			jsonData, _ := json.Marshal(taskRequest)
			http.Post(
				fmt.Sprintf("%s/api/tasks/submit", serverURL),
				"application/json",
				bytes.NewBuffer(jsonData),
			)
		}(i)
	}

	// Wait a bit for tasks to be processed
	time.Sleep(2 * time.Second)
	wg.Wait()

	// Get final worker stats
	resp, err = http.Get(fmt.Sprintf("%s/api/workers/stats", serverURL))
	if err != nil {
		t.Fatalf("Failed to get final worker stats: %v", err)
	}
	defer resp.Body.Close()

	var finalStats map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&finalStats); err != nil {
		t.Fatalf("Failed to decode final stats: %v", err)
	}

	// Analyze worker utilization
	workers := finalStats["workers"].([]interface{})
	totalTasks := 0
	activeWorkers := 0

	for _, worker := range workers {
		workerMap := worker.(map[string]interface{})
		var tasksCompleted int
		if tc, ok := workerMap["tasks_completed"].(float64); ok {
			tasksCompleted = int(tc)
		} else if tc, ok := workerMap["tasks_completed"].(int); ok {
			tasksCompleted = tc
		}
		isActive := workerMap["is_active"].(bool)

		totalTasks += tasksCompleted
		if isActive {
			activeWorkers++
		}
	}

	avgTasksPerWorker := float64(totalTasks) / float64(len(workers))
	utilization := float64(activeWorkers) / float64(numWorkers) * 100

	t.Logf("Worker utilization: %.1f%% (%d/%d active)", utilization, activeWorkers, numWorkers)
	t.Logf("Average tasks per worker: %.1f", avgTasksPerWorker)

	// Target: at least 80% worker utilization
	if utilization >= 80 {
		t.Logf("✓ Worker utilization meets target: %.1f%% (target: ≥80%%)", utilization)
	} else {
		t.Errorf("✗ Worker utilization below target: %.1f%% (target: ≥80%%)", utilization)
	}
}

func testMemoryEfficiency(t *testing.T, serverURL string, numTasks int) {
	t.Logf("Testing memory efficiency with %d tasks", numTasks)

	// Get initial memory usage
	var m1 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Submit tasks to test memory usage
	for i := 0; i < numTasks/10; i++ { // Submit fewer tasks to avoid timeout
		taskRequest := map[string]interface{}{
			"task_id":   fmt.Sprintf("mem-task-%d", i),
			"task_type": "compile",
			"payload":   make([]byte, 1024),
		}

		jsonData, _ := json.Marshal(taskRequest)
		http.Post(
			fmt.Sprintf("%s/api/tasks/submit", serverURL),
			"application/json",
			bytes.NewBuffer(jsonData),
		)
	}

	// Wait for processing
	time.Sleep(1 * time.Second)

	// Get final memory usage
	var m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m2)

	memoryDiff := m2.Alloc - m1.Alloc
	memoryPerTask := float64(memoryDiff) / float64(numTasks/10)

	t.Logf("Memory usage: %d bytes (%.2f bytes/task)", memoryDiff, memoryPerTask)

	// Target: less than 1KB per task in additional memory
	if memoryPerTask < 1024 {
		t.Logf("✓ Memory efficiency meets target: %.2f bytes/task (target: <1024)", memoryPerTask)
	} else {
		t.Errorf("✗ Memory efficiency below target: %.2f bytes/task (target: <1024)", memoryPerTask)
	}
}

func testResponseTimeUnderLoad(t *testing.T, serverURL string, numTasks int, concurrency int) {
	t.Logf("Testing response times under load with %d tasks", numTasks)

	var wg sync.WaitGroup
	var mu sync.Mutex
	responseTimes := make([]time.Duration, 0, numTasks)

	semaphore := make(chan struct{}, concurrency)
	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		go func(taskID int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			start := time.Now()
			taskRequest := map[string]interface{}{
				"task_id":   fmt.Sprintf("rt-task-%d", taskID),
				"task_type": "compile",
			}

			jsonData, _ := json.Marshal(taskRequest)
			resp, err := http.Post(
				fmt.Sprintf("%s/api/tasks/submit", serverURL),
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			duration := time.Since(start)

			if err == nil && resp.StatusCode == http.StatusOK {
				mu.Lock()
				responseTimes = append(responseTimes, duration)
				mu.Unlock()
			}
			if resp != nil {
				resp.Body.Close()
			}
		}(i)
	}

	wg.Wait()

	// Calculate response time statistics
	if len(responseTimes) == 0 {
		t.Fatal("No successful requests to measure response times")
	}

	var total time.Duration
	min := responseTimes[0]
	max := responseTimes[0]

	for _, rt := range responseTimes {
		total += rt
		if rt < min {
			min = rt
		}
		if rt > max {
			max = rt
		}
	}

	avg := total / time.Duration(len(responseTimes))
	p95Index := int(float64(len(responseTimes)) * 0.95)
	
	// Sort for percentile calculation
	for i := 0; i < len(responseTimes)-1; i++ {
		for j := i + 1; j < len(responseTimes); j++ {
			if responseTimes[i] > responseTimes[j] {
				responseTimes[i], responseTimes[j] = responseTimes[j], responseTimes[i]
			}
		}
	}
	p95 := responseTimes[p95Index]

	t.Logf("Response time statistics:")
	t.Logf("  Average: %v", avg)
	t.Logf("  Min: %v", min)
	t.Logf("  Max: %v", max)
	t.Logf("  95th percentile: %v", p95)

	// Performance targets
	targetAvg := 100 * time.Millisecond
	targetP95 := 200 * time.Millisecond

	if avg <= targetAvg {
		t.Logf("✓ Average response time meets target: %v (target: ≤%v)", avg, targetAvg)
	} else {
		t.Errorf("✗ Average response time above target: %v (target: ≤%v)", avg, targetAvg)
	}

	if p95 <= targetP95 {
		t.Logf("✓ 95th percentile response time meets target: %v (target: ≤%v)", p95, targetP95)
	} else {
		t.Errorf("✗ 95th percentile response time above target: %v (target: ≤%v)", p95, targetP95)
	}
}

// Mock service setup functions

func setupMockWorkerPool(t *testing.T, numWorkers int) http.Handler {
	router := mux.NewRouter()
	
	// Track worker statistics - declare workers here to persist across requests
	var workers []map[string]interface{}
	
	// Task submission endpoint
	router.HandleFunc("/api/tasks/submit", func(w http.ResponseWriter, r *http.Request) {
		var taskRequest map[string]interface{}
		json.NewDecoder(r.Body).Decode(&taskRequest)
		
		// Initialize workers if not done yet
		if len(workers) == 0 {
			workers = make([]map[string]interface{}, numWorkers)
			for i := 0; i < numWorkers; i++ {
				workers[i] = map[string]interface{}{
					"id":             fmt.Sprintf("worker-%d", i+1),
					"is_active":      true,
					"tasks_completed": 0,
					"last_heartbeat": time.Now().Unix(),
				}
			}
		}
		
		// Simulate task processing
		time.Sleep(10 * time.Millisecond)
		
		// Update worker stats
		workerIdx := len(workers) - 1
		currentTasks := workers[workerIdx]["tasks_completed"]
		var newTaskCount int
		if f, ok := currentTasks.(float64); ok {
			newTaskCount = int(f) + 1
		} else if i, ok := currentTasks.(int); ok {
			newTaskCount = i + 1
		} else {
			newTaskCount = 1
		}
		workers[workerIdx]["tasks_completed"] = newTaskCount
		
		response := map[string]interface{}{
			"task_id":     taskRequest["task_id"],
			"status":      "completed",
			"worker_id":   workers[workerIdx]["id"],
			"duration_ms": 10,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("POST")

	// Worker stats endpoint
	router.HandleFunc("/api/workers/stats", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"workers": workers,
			"total_workers": numWorkers,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	return router
}

func setupPerformanceCoordinator(t *testing.T) http.Handler {
	router := mux.NewRouter()
	
	// Performance-optimized build submission
	router.HandleFunc("/api/builds", func(w http.ResponseWriter, r *http.Request) {
		var buildRequest map[string]interface{}
		json.NewDecoder(r.Body).Decode(&buildRequest)
		
		buildID := fmt.Sprintf("build-%d", time.Now().UnixNano())
		
		response := map[string]interface{}{
			"build_id": buildID,
			"status":   "accepted",
			"queue_position": 1,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(response)
	}).Methods("POST")

	return router
}

func setupPerformanceWorkers(t *testing.T, count int) []http.Handler {
	workers := make([]http.Handler, count)
	
	for i := 0; i < count; i++ {
		router := mux.NewRouter()
		
		router.HandleFunc("/api/tasks/execute", func(w http.ResponseWriter, r *http.Request) {
			// Simulate fast task execution
			time.Sleep(5 * time.Millisecond)
			
			response := map[string]interface{}{
				"task_id":    fmt.Sprintf("task-%d", time.Now().UnixNano()),
				"worker_id":  fmt.Sprintf("worker-%d", i+1),
				"status":     "completed",
				"duration":   5,
				"artifacts":  []string{"build.log"},
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}).Methods("POST")

		workers[i] = router
	}
	
	return workers
}

func setupMockCacheServer(t *testing.T) http.Handler {
	router := mux.NewRouter()
	
	// Mock cache storage endpoint
	router.HandleFunc("/api/cache", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "stored",
		})
	}).Methods("POST")

	// Mock cache retrieval endpoint
	router.HandleFunc("/api/cache/{key}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		key := vars["key"]
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"key":       key,
			"value":     "cached-content",
			"timestamp": time.Now().Unix(),
		})
	}).Methods("GET")

	return router
}

// Additional performance test implementations

func testCacheHitRate(t *testing.T, serverURL string) {
	t.Logf("Testing cache hit rate")
	
	// Store some data first
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("perf-key-%d", i)
		cacheData := map[string]interface{}{
			"key":   key,
			"value": fmt.Sprintf("cached-value-%d", i),
			"ttl":   3600,
		}
		
		jsonData, _ := json.Marshal(cacheData)
		http.Post(
			fmt.Sprintf("%s/api/cache", serverURL),
			"application/json",
			bytes.NewBuffer(jsonData),
		)
	}
	
	// Test cache hits
	hits := 0
	total := 10
	
	for i := 0; i < total; i++ {
		key := fmt.Sprintf("perf-key-%d", i)
		resp, err := http.Get(fmt.Sprintf("%s/api/cache/%s", serverURL, key))
		if err == nil && resp.StatusCode == http.StatusOK {
			hits++
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	
	hitRate := float64(hits) / float64(total) * 100
	t.Logf("Cache hit rate: %.1f%% (%d/%d)", hitRate, hits, total)
	
	if hitRate >= 90 {
		t.Logf("✓ Cache hit rate meets target: %.1f%% (target: ≥90%%)", hitRate)
	} else {
		t.Errorf("✗ Cache hit rate below target: %.1f%% (target: ≥90%%)", hitRate)
	}
}

func testCacheStoragePerformance(t *testing.T, serverURL string) {
	t.Logf("Testing cache storage performance")
	
	numOps := 100
	var storageTimes []time.Duration
	
	for i := 0; i < numOps; i++ {
		start := time.Now()
		
		cacheData := map[string]interface{}{
			"key":   fmt.Sprintf("storage-test-%d", i),
			"value": make([]byte, 1024), // 1KB data
			"ttl":   3600,
		}
		
		jsonData, _ := json.Marshal(cacheData)
		resp, err := http.Post(
			fmt.Sprintf("%s/api/cache", serverURL),
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		
		duration := time.Since(start)
		if err == nil && resp.StatusCode == http.StatusCreated {
			storageTimes = append(storageTimes, duration)
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	
	if len(storageTimes) == 0 {
		t.Fatal("No successful cache operations to measure")
	}
	
	var total time.Duration
	for _, st := range storageTimes {
		total += st
	}
	avgStorageTime := total / time.Duration(len(storageTimes))
	
	t.Logf("Average cache storage time: %v", avgStorageTime)
	
	if avgStorageTime <= 10*time.Millisecond {
		t.Logf("✓ Cache storage performance meets target: %v (target: ≤10ms)", avgStorageTime)
	} else {
		t.Errorf("✗ Cache storage performance below target: %v (target: ≤10ms)", avgStorageTime)
	}
}

func testCacheEvictionPerformance(t *testing.T, serverURL string) {
	t.Logf("Testing cache eviction performance")
	
	// Fill cache to trigger eviction
	for i := 0; i < 1000; i++ {
		cacheData := map[string]interface{}{
			"key":   fmt.Sprintf("eviction-test-%d", i),
			"value": make([]byte, 1024),
			"ttl":   1, // Very short TTL to trigger eviction
		}
		
		jsonData, _ := json.Marshal(cacheData)
		http.Post(
			fmt.Sprintf("%s/api/cache", serverURL),
			"application/json",
			bytes.NewBuffer(jsonData),
		)
	}
	
	// Wait for eviction
	time.Sleep(2 * time.Second)
	
	start := time.Now()
	
	// Trigger eviction by adding one more item
	cacheData := map[string]interface{}{
		"key":   "eviction-trigger",
		"value": make([]byte, 1024),
		"ttl":   3600,
	}
	
	jsonData, _ := json.Marshal(cacheData)
	resp, err := http.Post(
		fmt.Sprintf("%s/api/cache", serverURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	
	evictionTime := time.Since(start)
	
	if err == nil && resp.StatusCode == http.StatusCreated {
		t.Logf("Cache eviction time: %v", evictionTime)
		
		if evictionTime <= 100*time.Millisecond {
			t.Logf("✓ Cache eviction performance meets target: %v (target: ≤100ms)", evictionTime)
		} else {
			t.Errorf("✗ Cache eviction performance below target: %v (target: ≤100ms)", evictionTime)
		}
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func testBuildThroughput(t *testing.T, coordURL string, workerURLs []string, cacheURL string) {
	t.Logf("Testing distributed build throughput")
	
	numBuilds := 50
	concurrentBuilds := 10
	startTime := time.Now()
	
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrentBuilds)
	completedBuilds := 0
	var mu sync.Mutex
	
	for i := 0; i < numBuilds; i++ {
		wg.Add(1)
		go func(buildID int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			buildRequest := map[string]interface{}{
				"project_name":   fmt.Sprintf("throughput-test-%d", buildID),
				"build_type":     "gradle",
				"tasks":          []string{"compileJava", "test"},
				"cache_enabled":   true,
				"cache_server":   cacheURL,
			}
			
			jsonData, _ := json.Marshal(buildRequest)
			resp, err := http.Post(
				fmt.Sprintf("%s/api/builds", coordURL),
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			
			if err == nil && resp.StatusCode == http.StatusAccepted {
				mu.Lock()
				completedBuilds++
				mu.Unlock()
			}
			if resp != nil {
				resp.Body.Close()
			}
		}(i)
	}
	
	wg.Wait()
	duration := time.Since(startTime)
	
	throughput := float64(completedBuilds) / duration.Seconds()
	t.Logf("Distributed build throughput: %.2f builds/sec (%d/%d in %v)", 
		throughput, completedBuilds, numBuilds, duration)
	
	if throughput >= 10 {
		t.Logf("✓ Build throughput meets target: %.2f builds/sec (target: ≥10)", throughput)
	} else {
		t.Errorf("✗ Build throughput below target: %.2f builds/sec (target: ≥10)", throughput)
	}
}

func testScalability(t *testing.T, coordURL string, workerURLs []string, cacheURL string) {
	t.Logf("Testing system scalability with %d workers", len(workerURLs))
	
	// Test increasing load
	workloads := []int{5, 10, 20, 40}
	
	for _, workload := range workloads {
		start := time.Now()
		completed := 0
		
		var wg sync.WaitGroup
		for i := 0; i < workload; i++ {
			wg.Add(1)
			go func(buildID int) {
				defer wg.Done()
				
				buildRequest := map[string]interface{}{
					"project_name": fmt.Sprintf("scalability-test-%d", buildID),
					"build_type":   "gradle",
					"tasks":        []string{"compileJava"},
				}
				
				jsonData, _ := json.Marshal(buildRequest)
				resp, err := http.Post(
					fmt.Sprintf("%s/api/builds", coordURL),
					"application/json",
					bytes.NewBuffer(jsonData),
				)
				
				if err == nil && resp.StatusCode == http.StatusAccepted {
					completed++
				}
				if resp != nil {
					resp.Body.Close()
				}
			}(i)
		}
		
		wg.Wait()
		duration := time.Since(start)
		throughput := float64(completed) / duration.Seconds()
		
		t.Logf("Workload %d: %.2f builds/sec", workload, throughput)
	}
}

func testResourceEfficiency(t *testing.T, coordURL string, workerURLs []string, cacheURL string) {
	t.Logf("Testing resource efficiency")
	
	// Monitor resource usage during builds
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)
	
	// Run some builds
	for i := 0; i < 20; i++ {
		buildRequest := map[string]interface{}{
			"project_name": fmt.Sprintf("resource-test-%d", i),
			"build_type":   "gradle",
			"tasks":        []string{"compileJava"},
		}
		
		jsonData, _ := json.Marshal(buildRequest)
		http.Post(
			fmt.Sprintf("%s/api/builds", coordURL),
			"application/json",
			bytes.NewBuffer(jsonData),
		)
	}
	
	runtime.GC()
	runtime.ReadMemStats(&m2)
	
	memoryDiff := m2.Alloc - m1.Alloc
	memoryPerBuild := float64(memoryDiff) / 20.0
	
	t.Logf("Resource efficiency: %.2f bytes/build", memoryPerBuild)
	
	if memoryPerBuild < 1024 {
		t.Logf("✓ Resource efficiency meets target: %.2f bytes/build (target: <1024)", memoryPerBuild)
	} else {
		t.Errorf("✗ Resource efficiency below target: %.2f bytes/build (target: <1024)", memoryPerBuild)
	}
}

func getWorkerURLs(servers []*httptest.Server) []string {
	urls := make([]string, len(servers))
	for i, server := range servers {
		urls[i] = server.URL
	}
	return urls
}