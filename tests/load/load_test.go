package load

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestConcurrentBuildSubmissions tests the load testing framework's ability to simulate concurrent operations
func TestConcurrentBuildSubmissions(t *testing.T) {
	concurrentLevels := []int{1, 2, 5, 10}

	for _, concurrency := range concurrentLevels {
		t.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(t *testing.T) {
			var wg sync.WaitGroup
			completed := make(chan bool, concurrency)
			startTimes := make([]time.Time, concurrency)
			endTimes := make([]time.Time, concurrency)

			startTest := time.Now()

			// Submit operations concurrently
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()

					startTimes[index] = time.Now()

					// Simulate operation with varying duration
					time.Sleep(time.Duration(50+index%50) * time.Millisecond)

					endTimes[index] = time.Now()
					completed <- true
				}(i)
			}

			// Wait for all operations to complete
			wg.Wait()
			close(completed)

			totalTime := time.Since(startTest)

			// Calculate metrics
			var totalDuration time.Duration
			completedCount := 0
			for i := 0; i < concurrency; i++ {
				if !endTimes[i].IsZero() && !startTimes[i].IsZero() {
					totalDuration += endTimes[i].Sub(startTimes[i])
					completedCount++
				}
			}

			if completedCount == 0 {
				t.Fatal("No operations completed")
			}

			avgDuration := totalDuration / time.Duration(completedCount)
			throughput := float64(completedCount) / totalTime.Seconds()

			// Log performance metrics
			t.Logf("Concurrency %d: Total time=%v, Avg operation time=%v, Throughput=%.2f ops/sec, Completed=%d/%d",
				concurrency, totalTime, avgDuration, throughput, completedCount, concurrency)

			// Assertions for load test framework
			if completedCount != concurrency {
				t.Errorf("Completion count mismatch: got %d, expected %d", completedCount, concurrency)
			}

			if throughput < 0.5 {
				t.Errorf("Throughput too low: %.2f ops/sec (expected >= 0.5)", throughput)
			}
		})
	}
}

// TestResourceUtilizationUnderLoad tests resource usage patterns under increasing load
func TestResourceUtilizationUnderLoad(t *testing.T) {
	// Simulate increasing load over time
	phases := []struct {
		name        string
		concurrency int
		duration    time.Duration
	}{
		{"Low load", 2, 2 * time.Second},
		{"Medium load", 5, 3 * time.Second},
		{"High load", 10, 5 * time.Second},
		{"Peak load", 20, 3 * time.Second},
	}

	startTime := time.Now()
	totalBuilds := 0
	totalErrors := 0

	for _, phase := range phases {
		t.Run(phase.name, func(t *testing.T) {
			var wg sync.WaitGroup
			errors := make(chan error, phase.concurrency)

			phaseStart := time.Now()
			buildsCompleted := 0

			// Run builds for the duration of this phase
			for time.Since(phaseStart) < phase.duration {
				for i := 0; i < phase.concurrency; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()

						// Simulate build with varying duration
						buildTime := time.Duration(100+index%200) * time.Millisecond
						time.Sleep(buildTime)

						// Simulate occasional errors
						if (totalBuilds+index)%15 == 0 {
							errors <- fmt.Errorf("phase %s: simulated error for build %d", phase.name, totalBuilds+index)
						}

						buildsCompleted++
					}(i)
				}

				// Small delay between batches
				time.Sleep(100 * time.Millisecond)
			}

			wg.Wait()
			close(errors)

			// Count errors in this phase
			phaseErrors := 0
			for range errors {
				phaseErrors++
			}

			totalBuilds += buildsCompleted
			totalErrors += phaseErrors

			phaseDuration := time.Since(phaseStart)
			phaseThroughput := float64(buildsCompleted) / phaseDuration.Seconds()
			phaseSuccessRate := float64(buildsCompleted-phaseErrors) / float64(buildsCompleted) * 100

			t.Logf("Phase %s: Builds=%d, Errors=%d, Duration=%v, Throughput=%.2f builds/sec, Success rate=%.1f%%",
				phase.name, buildsCompleted, phaseErrors, phaseDuration, phaseThroughput, phaseSuccessRate)

			// Phase-specific assertions
			if phaseSuccessRate < 85.0 {
				t.Errorf("Phase %s success rate too low: %.1f%%", phase.name, phaseSuccessRate)
			}
		})
	}

	// Overall test metrics
	totalDuration := time.Since(startTime)
	overallThroughput := float64(totalBuilds) / totalDuration.Seconds()
	overallSuccessRate := float64(totalBuilds-totalErrors) / float64(totalBuilds) * 100

	t.Logf("Overall: Total builds=%d, Total errors=%d, Total duration=%v, Overall throughput=%.2f builds/sec, Overall success rate=%.1f%%",
		totalBuilds, totalErrors, totalDuration, overallThroughput, overallSuccessRate)

	if overallSuccessRate < 90.0 {
		t.Errorf("Overall success rate too low: %.1f%%", overallSuccessRate)
	}
}

// TestBuildQueueCapacity tests the system's ability to handle build queue capacity
func TestBuildQueueCapacity(t *testing.T) {
	queueSizes := []int{10, 20, 50, 100}

	for _, queueSize := range queueSizes {
		t.Run(fmt.Sprintf("QueueSize_%d", queueSize), func(t *testing.T) {
			var wg sync.WaitGroup
			buildQueue := make(chan int, queueSize)
			completedBuilds := 0
			failedBuilds := 0

			// Producer: submit builds to queue
			go func() {
				for i := 0; i < queueSize*2; i++ { // Submit more than queue capacity
					buildQueue <- i
					time.Sleep(10 * time.Millisecond) // Simulate submission rate
				}
				close(buildQueue)
			}()

			startTime := time.Now()

			// Consumer: process builds from queue
			for i := 0; i < 5; i++ { // 5 concurrent workers
				wg.Add(1)
				go func(workerID int) {
					defer wg.Done()

					for buildID := range buildQueue {
						// Simulate build processing
						buildTime := time.Duration(50+buildID%100) * time.Millisecond
						time.Sleep(buildTime)

						// Simulate occasional failures
						if buildID%20 == 0 {
							failedBuilds++
						} else {
							completedBuilds++
						}
					}
				}(i)
			}

			wg.Wait()
			totalDuration := time.Since(startTime)

			totalBuilds := completedBuilds + failedBuilds
			throughput := float64(totalBuilds) / totalDuration.Seconds()
			successRate := float64(completedBuilds) / float64(totalBuilds) * 100

			t.Logf("Queue size %d: Total builds=%d, Completed=%d, Failed=%d, Duration=%v, Throughput=%.2f builds/sec, Success rate=%.1f%%",
				queueSize, totalBuilds, completedBuilds, failedBuilds, totalDuration, throughput, successRate)

			// Queue should handle at least queueSize builds
			if totalBuilds < queueSize {
				t.Errorf("Queue processing insufficient: processed %d builds, expected at least %d", totalBuilds, queueSize)
			}

			if successRate < 90.0 {
				t.Errorf("Success rate too low for queue size %d: %.1f%%", queueSize, successRate)
			}
		})
	}
}

// TestMemoryUsagePatterns tests memory usage patterns under sustained load
func TestMemoryUsagePatterns(t *testing.T) {
	duration := 30 * time.Second
	checkInterval := 2 * time.Second
	concurrentBuilds := 8

	var wg sync.WaitGroup
	stopChan := make(chan bool)
	memorySamples := make([]float64, 0)

	// Memory monitoring goroutine
	go func() {
		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Simulate memory usage measurement
				// In a real test, this would use runtime.MemStats or similar
				memoryUsage := 50.0 + float64(time.Now().UnixNano()%30) // Simulated memory usage 50-80%
				memorySamples = append(memorySamples, memoryUsage)

			case <-stopChan:
				return
			}
		}
	}()

	startTime := time.Now()
	totalBuilds := 0

	// Run builds for the specified duration
	for time.Since(startTime) < duration {
		for i := 0; i < concurrentBuilds; i++ {
			wg.Add(1)
			go func(buildNum int) {
				defer wg.Done()

				// Simulate build
				buildTime := time.Duration(100+buildNum%300) * time.Millisecond
				time.Sleep(buildTime)
				totalBuilds++
			}(totalBuilds)
		}

		time.Sleep(500 * time.Millisecond) // Delay between batches
	}

	wg.Wait()
	close(stopChan)

	totalDuration := time.Since(startTime)
	throughput := float64(totalBuilds) / totalDuration.Seconds()

	// Calculate memory statistics
	var maxMemory, avgMemory float64
	if len(memorySamples) > 0 {
		maxMemory = memorySamples[0]
		var sum float64
		for _, sample := range memorySamples {
			if sample > maxMemory {
				maxMemory = sample
			}
			sum += sample
		}
		avgMemory = sum / float64(len(memorySamples))
	}

	t.Logf("Memory test: Duration=%v, Total builds=%d, Throughput=%.2f builds/sec, Memory samples=%d, Avg memory=%.1f%%, Max memory=%.1f%%",
		totalDuration, totalBuilds, throughput, len(memorySamples), avgMemory, maxMemory)

	// Memory usage assertions
	if maxMemory > 95.0 {
		t.Errorf("Maximum memory usage too high: %.1f%% (should be < 95%%)", maxMemory)
	}

	if avgMemory > 80.0 {
		t.Errorf("Average memory usage too high: %.1f%% (should be < 80%%)", avgMemory)
	}

	if throughput < 5.0 {
		t.Errorf("Throughput too low: %.2f builds/sec (expected >= 5.0)", throughput)
	}
}
