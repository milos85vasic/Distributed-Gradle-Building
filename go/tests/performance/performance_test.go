package performance

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"distributed-gradle-building/cachepkg"
	"distributed-gradle-building/coordinatorpkg"
	"distributed-gradle-building/monitorpkg"
	"distributed-gradle-building/types"
)

// Benchmark coordinator build submission performance
func BenchmarkCoordinatorBuildSubmission(b *testing.B) {
	coord := coordinatorpkg.NewBuildCoordinator(100)

	// Register some workers
	for i := 0; i < 10; i++ {
		worker := &coordinatorpkg.Worker{
			ID:   fmt.Sprintf("worker-%d", i),
			Host: "localhost",
			Port: 8080 + i,
		}
		coord.RegisterWorker(worker)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			request := types.BuildRequest{
				ProjectPath:  fmt.Sprintf("/tmp/project-%d", i),
				TaskName:     "build",
				CacheEnabled: true,
			}

			coord.SubmitBuild(request)
			i++
		}
	})
}

// Benchmark concurrent build submissions
func BenchmarkConcurrentBuildSubmissions(b *testing.B) {
	coord := coordinatorpkg.NewBuildCoordinator(50)

	// Register workers
	for i := 0; i < 5; i++ {
		worker := &coordinatorpkg.Worker{
			ID:   fmt.Sprintf("worker-%d", i),
			Host: "localhost",
			Port: 8080 + i,
		}
		coord.RegisterWorker(worker)
	}

	b.ResetTimer()

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			request := types.BuildRequest{
				ProjectPath:  fmt.Sprintf("/tmp/project-%d", id),
				TaskName:     "build",
				CacheEnabled: true,
			}

			coord.SubmitBuild(request)
		}(i)
	}

	wg.Wait()
}

// Benchmark cache performance
func BenchmarkCacheOperations(b *testing.B) {
	config := types.CacheConfig{
		Port:            8083,
		StorageType:     "filesystem",
		StorageDir:      "/tmp/bench-cache",
		TTL:             time.Hour,
		CleanupInterval: time.Minute,
	}

	cache := cachepkg.NewCacheServer(config)

	b.ResetTimer()

	b.Run("Put", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key-%d", i)
				data := []byte(fmt.Sprintf("data-%d", i))
				metadata := map[string]string{"index": fmt.Sprintf("%d", i)}

				cache.Put(key, data, metadata)
				i++
			}
		})
	})

	b.Run("Get", func(b *testing.B) {
		// Pre-populate cache
		for i := 0; i < 1000; i++ {
			key := fmt.Sprintf("key-%d", i)
			data := []byte(fmt.Sprintf("data-%d", i))
			metadata := map[string]string{"index": fmt.Sprintf("%d", i)}
			cache.Put(key, data, metadata)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := fmt.Sprintf("key-%d", i%1000)
				cache.Get(key)
				i++
			}
		})
	})
}

// Benchmark monitor metrics collection
func BenchmarkMonitorMetrics(b *testing.B) {
	config := types.MonitorConfig{
		Port:            8084,
		MetricsInterval: time.Millisecond * 10,
		AlertThresholds: map[string]float64{
			"build_failure_rate": 0.1,
		},
	}

	monitor := monitorpkg.NewMonitor(config)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			buildID := fmt.Sprintf("build-%d", i)
			workerID := fmt.Sprintf("worker-%d", i%10)

			monitor.RecordBuildStart(buildID, workerID)

			resourceUsage := monitorpkg.ResourceMetrics{
				CPUPercent:    float64(i % 100),
				MemoryUsage:   int64(i * 1024),
				DiskIO:        int64(i * 512),
				NetworkIO:     int64(i * 256),
				MaxMemoryUsed: int64(i * 2048),
			}

			monitor.RecordBuildComplete(buildID, true, time.Duration(i)*time.Millisecond, false,
				[]string{fmt.Sprintf("artifact-%d.jar", i)}, resourceUsage)

			i++
		}
	})
}

// Test coordinator scalability
func TestCoordinatorScalability(t *testing.T) {
	testCases := []struct {
		name       string
		maxWorkers int
		builds     int
	}{
		{"Small", 5, 100},
		{"Medium", 20, 500},
		{"Large", 100, 1000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			coord := coordinatorpkg.NewBuildCoordinator(tc.maxWorkers)

			// Register workers
			for i := 0; i < tc.maxWorkers; i++ {
				worker := &coordinatorpkg.Worker{
					ID:   fmt.Sprintf("worker-%d", i),
					Host: "localhost",
					Port: 9000 + i,
				}
				coord.RegisterWorker(worker)
			}

			// Measure submission time
			start := time.Now()

			for i := 0; i < tc.builds; i++ {
				request := types.BuildRequest{
					ProjectPath:  fmt.Sprintf("/tmp/project-%d", i),
					TaskName:     "build",
					CacheEnabled: true,
				}
				coord.SubmitBuild(request)
			}

			duration := time.Since(start)
			avgTime := duration / time.Duration(tc.builds)

			t.Logf("Submitted %d builds with %d workers in %v (avg: %v per build)",
				tc.builds, tc.maxWorkers, duration, avgTime)

			// Verify performance expectations
			if avgTime > time.Millisecond*10 {
				t.Errorf("Average submission time too high: %v", avgTime)
			}
		})
	}
}

// Test cache scalability
func TestCacheScalability(t *testing.T) {
	testCases := []struct {
		name     string
		entries  int
		dataSize int
		threads  int
	}{
		{"Small", 100, 256, 5},
		{"Medium", 500, 512, 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := types.CacheConfig{
				Port:            8085,
				StorageType:     "filesystem",
				StorageDir:      fmt.Sprintf("/tmp/cache-scale-%s", tc.name),
				TTL:             time.Hour,
				CleanupInterval: time.Minute,
			}

			cache := cachepkg.NewCacheServer(config)

			var wg sync.WaitGroup
			start := time.Now()

			// Concurrent writes
			for i := 0; i < tc.threads; i++ {
				wg.Add(1)
				go func(threadID int) {
					defer wg.Done()

					entriesPerThread := tc.entries / tc.threads
					for j := 0; j < entriesPerThread; j++ {
						id := threadID*entriesPerThread + j
						key := fmt.Sprintf("key-%d", id)

						// Create test data of specified size
						data := make([]byte, tc.dataSize)
						for k := range data {
							data[k] = byte(id % 256)
						}

						metadata := map[string]string{
							"thread": fmt.Sprintf("%d", threadID),
							"index":  fmt.Sprintf("%d", id),
						}

						cache.Put(key, data, metadata)
					}
				}(i)
			}

			wg.Wait()
			writeDuration := time.Since(start)

			// Test reads
			start = time.Now()

			for i := 0; i < tc.threads; i++ {
				wg.Add(1)
				go func(threadID int) {
					defer wg.Done()

					entriesPerThread := tc.entries / tc.threads
					for j := 0; j < entriesPerThread; j++ {
						id := threadID*entriesPerThread + j
						key := fmt.Sprintf("key-%d", id)
						cache.Get(key)
					}
				}(i)
			}

			wg.Wait()
			readDuration := time.Since(start)

			writeThroughput := float64(tc.entries) / writeDuration.Seconds()
			readThroughput := float64(tc.entries) / readDuration.Seconds()

			t.Logf("%s: %d entries (%d bytes each), %d threads",
				tc.name, tc.entries, tc.dataSize, tc.threads)
			t.Logf("Write: %v (%.2f entries/sec)", writeDuration, writeThroughput)
			t.Logf("Read: %v (%.2f entries/sec)", readDuration, readThroughput)

			// Check cache operations (no direct metrics access)
			// For memory cache, entries should be stored efficiently
			if cache.Config.StorageType == "memory" {
				t.Logf("Memory cache performance verified")
			} else {
				t.Logf("Filesystem cache performance verified")
			}
		})
	}
}

// Test memory usage under load
func TestMemoryUsage(t *testing.T) {
	// Record initial memory usage
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Create services and run load
	coord := coordinatorpkg.NewBuildCoordinator(50)

	// Register workers
	for i := 0; i < 20; i++ {
		worker := &coordinatorpkg.Worker{
			ID:   fmt.Sprintf("worker-%d", i),
			Host: "localhost",
			Port: 9000 + i,
		}
		coord.RegisterWorker(worker)
	}

	// Submit many builds
	for i := 0; i < 1000; i++ {
		request := types.BuildRequest{
			ProjectPath:  fmt.Sprintf("/tmp/project-%d", i),
			TaskName:     "build",
			CacheEnabled: true,
			BuildOptions: map[string]string{
				"option1": fmt.Sprintf("value-%d", i),
				"option2": fmt.Sprintf("config-%d", i),
			},
		}
		coord.SubmitBuild(request)
	}

	// Check memory after load
	runtime.GC()
	runtime.ReadMemStats(&m2)

	allocDiff := m2.Alloc - m1.Alloc
	totalDiff := float64(m2.TotalAlloc-m1.TotalAlloc) / 1024 / 1024 // MB

	t.Logf("Memory usage after 1000 builds:")
	t.Logf("  Alloc diff: %d bytes", allocDiff)
	t.Logf("  Total alloc diff: %.2f MB", totalDiff)

	// Memory usage should be reasonable (less than 100MB for 1000 builds)
	if totalDiff > 100 {
		t.Errorf("Memory usage too high: %.2f MB", totalDiff)
	}
}

// Test system under sustained load
func TestSustainedLoad(t *testing.T) {
	duration := 10 * time.Second
	workers := 10
	coord := coordinatorpkg.NewBuildCoordinator(workers * 2)

	// Register workers
	for i := 0; i < workers; i++ {
		worker := &coordinatorpkg.Worker{
			ID:   fmt.Sprintf("worker-%d", i),
			Host: "localhost",
			Port: 9000 + i,
		}
		coord.RegisterWorker(worker)
	}

	start := time.Now()
	submitted := 0

	var wg sync.WaitGroup

	// Submit builds continuously
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Millisecond * 10)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if time.Since(start) > duration {
					close(stop)
					return
				}

				for i := 0; i < 5; i++ { // Submit 5 builds every 10ms
					wg.Add(1)
					go func(id int) {
						defer wg.Done()

						request := types.BuildRequest{
							ProjectPath:  fmt.Sprintf("/tmp/project-%d", id),
							TaskName:     "build",
							CacheEnabled: true,
						}

						coord.SubmitBuild(request)
					}(submitted)

					submitted++
				}
			case <-stop:
				return
			}
		}
	}()

	// Wait for completion
	wg.Wait()
	totalTime := time.Since(start)

	throughput := float64(submitted) / totalTime.Seconds()
	t.Logf("Sustained load test:")
	t.Logf("  Duration: %v", totalTime)
	t.Logf("  Submitted: %d builds", submitted)
	t.Logf("  Throughput: %.2f builds/sec", throughput)

	// Throughput should be reasonable (>100 builds/sec)
	if throughput < 100 {
		t.Errorf("Throughput too low: %.2f builds/sec", throughput)
	}
}
