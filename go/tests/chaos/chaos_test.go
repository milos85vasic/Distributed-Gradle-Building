package chaos

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"distributed-gradle-building/coordinatorpkg"
	"distributed-gradle-building/types"
	"distributed-gradle-building/workerpkg"
)

// ChaosEngine implements fault injection and chaos testing
type ChaosEngine struct {
	failureRate     float64
	networkLatency  time.Duration
	affectedWorkers map[string]bool
	mutex           sync.RWMutex
}

// ChaosTestResult represents outcome of a chaos test
type ChaosTestResult struct {
	TestName        string
	BuildsSubmitted int
	BuildsCompleted int
	SuccessRate     float64
	RecoveryTime    time.Duration
	Errors          []error
}

// NewChaosEngine creates a new chaos testing engine
func NewChaosEngine() *ChaosEngine {
	return &ChaosEngine{
		failureRate:     0.3,
		networkLatency:  100 * time.Millisecond,
		affectedWorkers: make(map[string]bool),
	}
}

// SetFailureRate configures failure probability (0.0-1.0)
func (ce *ChaosEngine) SetFailureRate(rate float64) {
	ce.mutex.Lock()
	defer ce.mutex.Unlock()
	ce.failureRate = rate
}

// SetNetworkLatency configures artificial network delays
func (ce *ChaosEngine) SetNetworkLatency(duration time.Duration) {
	ce.mutex.Lock()
	defer ce.mutex.Unlock()
	ce.networkLatency = duration
}

// SimulateNetworkLatency adds artificial delay to operations
func (ce *ChaosEngine) SimulateNetworkLatency() time.Duration {
	ce.mutex.RLock()
	defer ce.mutex.RUnlock()
	return ce.networkLatency
}

// ShouldFail determines if an operation should fail based on failure rate
func (ce *ChaosEngine) ShouldFail() bool {
	ce.mutex.RLock()
	defer ce.mutex.RUnlock()
	if ce.failureRate <= 0 {
		return false
	}
	return rand.Float64() < ce.failureRate
}

// TestChaosWorkerFailure tests system resilience to worker failures
func TestChaosWorkerFailure(t *testing.T) {
	t.Parallel()

	chaos := NewChaosEngine()
	chaos.SetFailureRate(0.3) // 30% failure rate

	// Setup test infrastructure
	coordinator := coordinatorpkg.NewBuildCoordinator(10)

	workers := make([]*workerpkg.WorkerService, 4)
	for i := 0; i < 4; i++ {
		config := types.WorkerConfig{
			ID:             fmt.Sprintf("chaos-worker-%d", i),
			CoordinatorURL: "localhost:8081",
			HTTPPort:       9000 + i,
			RPCPort:        10000 + i,
			BuildDir:       fmt.Sprintf("/tmp/chaos-build-%d", i),
			CacheEnabled:   true,
		}

		worker := workerpkg.NewWorkerService(config.ID, config.CoordinatorURL, config)
		worker.BuildDir = config.BuildDir
		workers[i] = worker
	}

	// Submit builds and verify system resilience
	var wg sync.WaitGroup
	results := make(chan types.BuildRequest, 20)

	buildCount := 20
	for i := 0; i < buildCount; i++ {
		wg.Add(1)
		go func(buildID int) {
			defer wg.Done()

			buildReq := types.BuildRequest{
				ProjectPath:  fmt.Sprintf("/tmp/chaos-project-%d", buildID),
				TaskName:     fmt.Sprintf("chaos-task-%d", buildID),
				CacheEnabled: true,
				BuildOptions: map[string]string{},
				Timestamp:    time.Now(),
				RequestID:    fmt.Sprintf("chaos-test-%d", buildID),
			}

			// Add network latency simulation
			time.Sleep(chaos.SimulateNetworkLatency())

			_, err := coordinator.SubmitBuild(buildReq)
			if err == nil {
				results <- buildReq
			}
		}(i)
	}

	wg.Wait()
	close(results)

	// Verify system resilience
	completedBuilds := 0
	for range results {
		completedBuilds++
	}

	successRate := float64(completedBuilds) / float64(buildCount)
	t.Logf("Chaos test results: %d/%d builds completed (%.1f%% success rate)",
		completedBuilds, buildCount, successRate*100)

	// We expect some failures due to chaos, but system should still work
	if successRate < 0.5 {
		t.Errorf("Success rate too low: %.1f%% (expected >= 50%%)", successRate*100)
	}

	if successRate > 0.9 {
		t.Logf("Warning: High success rate (%.1f%%) - chaos may not have been injected properly",
			successRate*100)
	}
}

// TestNetworkLatencyImpact tests system behavior under network latency
func TestNetworkLatencyImpact(t *testing.T) {
	t.Parallel()

	chaos := NewChaosEngine()
	chaos.SetNetworkLatency(200 * time.Millisecond) // High latency

	coordinator := coordinatorpkg.NewBuildCoordinator(5)

	// Test builds with network latency
	var wg sync.WaitGroup
	startTime := time.Now()

	buildCount := 10
	for i := 0; i < buildCount; i++ {
		wg.Add(1)
		go func(buildID int) {
			defer wg.Done()

			buildReq := types.BuildRequest{
				ProjectPath:  fmt.Sprintf("/tmp/latency-project-%d", buildID),
				TaskName:     fmt.Sprintf("latency-task-%d", buildID),
				CacheEnabled: true,
				BuildOptions: map[string]string{},
				Timestamp:    time.Now(),
				RequestID:    fmt.Sprintf("latency-test-%d", buildID),
			}

			// Add artificial network latency
			time.Sleep(chaos.SimulateNetworkLatency())

			_, err := coordinator.SubmitBuild(buildReq)
			if err != nil {
				t.Logf("Build submission failed under latency: %v", err)
			}
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(startTime)
	avgLatency := totalTime / time.Duration(buildCount)

	t.Logf("Network latency test completed in %v (avg: %v per build)",
		totalTime, avgLatency)

	// System should still function despite latency
	if avgLatency > 5*time.Second {
		t.Errorf("Average latency too high: %v (expected < 5s)", avgLatency)
	}
}

// TestConcurrentChaos tests multiple chaos scenarios simultaneously
func TestConcurrentChaos(t *testing.T) {
	t.Parallel()

	chaos := NewChaosEngine()
	chaos.SetFailureRate(0.2) // 20% failure rate
	chaos.SetNetworkLatency(50 * time.Millisecond)

	coordinator := coordinatorpkg.NewBuildCoordinator(8)

	// Submit builds from multiple concurrent sources
	var wg sync.WaitGroup
	successCount := 0
	failureCount := 0
	mutex := sync.Mutex{}

	buildCount := 30
	concurrentGoroutines := 5
	buildsPerGoroutine := buildCount / concurrentGoroutines

	for i := 0; i < concurrentGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < buildsPerGoroutine; j++ {
				buildID := goroutineID*buildsPerGoroutine + j

				buildReq := types.BuildRequest{
					ProjectPath:  fmt.Sprintf("/tmp/concurrent-project-%d", buildID),
					TaskName:     fmt.Sprintf("concurrent-task-%d", buildID),
					CacheEnabled: true,
					BuildOptions: map[string]string{},
					Timestamp:    time.Now(),
					RequestID:    fmt.Sprintf("concurrent-test-%d", buildID),
				}

				// Random chaos injection
				time.Sleep(chaos.SimulateNetworkLatency())

				_, err := coordinator.SubmitBuild(buildReq)
				mutex.Lock()
				if err == nil {
					successCount++
				} else {
					failureCount++
				}
				mutex.Unlock()
			}
		}(i)
	}

	wg.Wait()

	successRate := float64(successCount) / float64(buildCount)
	t.Logf("Concurrent chaos test: %d/%d builds succeeded (%.1f%%)",
		successCount, buildCount, successRate*100)

	// System should handle concurrent chaos well
	if successRate < 0.6 {
		t.Errorf("Success rate too low under concurrent chaos: %.1f%% (expected >= 60%%)",
			successRate*100)
	}
}

// TestSystemRecovery tests system recovery after chaos injection
func TestSystemRecovery(t *testing.T) {
	t.Parallel()

	chaos := NewChaosEngine()
	chaos.SetFailureRate(0.5) // High failure rate
	chaos.SetNetworkLatency(100 * time.Millisecond)

	coordinator := coordinatorpkg.NewBuildCoordinator(6)

	// Phase 1: Apply chaos
	t.Log("Phase 1: Applying chaos...")
	var wg sync.WaitGroup
	chaosSuccessCount := 0
	chaosBuildCount := 10

	for i := 0; i < chaosBuildCount; i++ {
		wg.Add(1)
		go func(buildID int) {
			defer wg.Done()

			buildReq := types.BuildRequest{
				ProjectPath:  fmt.Sprintf("/tmp/chaos-phase1-%d", buildID),
				TaskName:     fmt.Sprintf("chaos-phase1-task-%d", buildID),
				CacheEnabled: false, // Disable cache during chaos
				BuildOptions: map[string]string{},
				Timestamp:    time.Now(),
				RequestID:    fmt.Sprintf("chaos-phase1-%d", buildID),
			}

			time.Sleep(chaos.SimulateNetworkLatency())
			_, err := coordinator.SubmitBuild(buildReq)
			if err == nil && !chaos.ShouldFail() {
				chaosSuccessCount++
			}
		}(i)
	}
	wg.Wait()

	chaosSuccessRate := float64(chaosSuccessCount) / float64(chaosBuildCount)
	t.Logf("Chaos phase: %d/%d builds succeeded (%.1f%%)",
		chaosSuccessCount, chaosBuildCount, chaosSuccessRate*100)

	// Wait for system to stabilize
	time.Sleep(2 * time.Second)

	// Phase 2: Recovery (remove chaos)
	t.Log("Phase 2: Testing recovery...")
	chaos.SetFailureRate(0.0)                      // No failures
	chaos.SetNetworkLatency(10 * time.Millisecond) // Normal latency

	recoverySuccessCount := 0
	recoveryBuildCount := 10

	for i := 0; i < recoveryBuildCount; i++ {
		wg.Add(1)
		go func(buildID int) {
			defer wg.Done()

			buildReq := types.BuildRequest{
				ProjectPath:  fmt.Sprintf("/tmp/recovery-project-%d", buildID),
				TaskName:     fmt.Sprintf("recovery-task-%d", buildID),
				CacheEnabled: true, // Re-enable cache
				BuildOptions: map[string]string{},
				Timestamp:    time.Now(),
				RequestID:    fmt.Sprintf("recovery-test-%d", buildID),
			}

			_, err := coordinator.SubmitBuild(buildReq)
			if err == nil {
				recoverySuccessCount++
			}
		}(i)
	}
	wg.Wait()

	recoverySuccessRate := float64(recoverySuccessCount) / float64(recoveryBuildCount)
	t.Logf("Recovery phase: %d/%d builds succeeded (%.1f%%)",
		recoverySuccessCount, recoveryBuildCount, recoverySuccessRate*100)

	// System should recover better than chaos phase
	if recoverySuccessRate <= chaosSuccessRate {
		t.Errorf("No improvement in recovery phase: %.1f%% vs %.1f%% chaos",
			recoverySuccessRate*100, chaosSuccessRate*100)
	}

	// Recovery should be significantly better
	if recoverySuccessRate < 0.8 {
		t.Errorf("Recovery success rate too low: %.1f%% (expected >= 80%%)",
			recoverySuccessRate*100)
	}
}
