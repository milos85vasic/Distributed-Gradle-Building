package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestBuildRequest_JSONSerialization(t *testing.T) {
	original := BuildRequest{
		ProjectPath:  "/projects/myapp",
		TaskName:     "build",
		WorkerID:     "worker-123",
		CacheEnabled: true,
		BuildOptions: map[string]string{
			"clean":    "true",
			"parallel": "4",
		},
		Timestamp: time.Now().Truncate(time.Millisecond),
		RequestID: "req-123",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal BuildRequest: %v", err)
	}

	// Unmarshal from JSON
	var decoded BuildRequest
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal BuildRequest: %v", err)
	}

	// Compare
	if decoded.ProjectPath != original.ProjectPath {
		t.Errorf("ProjectPath mismatch: expected %s, got %s", original.ProjectPath, decoded.ProjectPath)
	}

	if decoded.TaskName != original.TaskName {
		t.Errorf("TaskName mismatch: expected %s, got %s", original.TaskName, decoded.TaskName)
	}

	if decoded.WorkerID != original.WorkerID {
		t.Errorf("WorkerID mismatch: expected %s, got %s", original.WorkerID, decoded.WorkerID)
	}

	if decoded.CacheEnabled != original.CacheEnabled {
		t.Errorf("CacheEnabled mismatch: expected %v, got %v", original.CacheEnabled, decoded.CacheEnabled)
	}

	if len(decoded.BuildOptions) != len(original.BuildOptions) {
		t.Errorf("BuildOptions length mismatch: expected %d, got %d", len(original.BuildOptions), len(decoded.BuildOptions))
	}

	for k, v := range original.BuildOptions {
		if decoded.BuildOptions[k] != v {
			t.Errorf("BuildOptions[%s] mismatch: expected %s, got %s", k, v, decoded.BuildOptions[k])
		}
	}

	if !decoded.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Timestamp mismatch: expected %v, got %v", original.Timestamp, decoded.Timestamp)
	}

	if decoded.RequestID != original.RequestID {
		t.Errorf("RequestID mismatch: expected %s, got %s", original.RequestID, decoded.RequestID)
	}
}

func TestBuildResponse_JSONSerialization(t *testing.T) {
	original := BuildResponse{
		Success:       true,
		WorkerID:      "worker-123",
		BuildDuration: 2*time.Minute + 30*time.Second,
		Artifacts:     []string{"build/libs/myapp.jar", "build/reports/tests"},
		ErrorMessage:  "",
		Metrics: BuildMetrics{
			CacheHitRate:  0.85,
			CompiledFiles: 42,
			TestResults: TestResults{
				TotalTests:   100,
				PassedTests:  95,
				FailedTests:  5,
				SkippedTests: 0,
				TestDuration: 45 * time.Second,
				Coverage:     0.92,
			},
			ResourceUsage: ResourceMetrics{
				CPUPercent:    75.5,
				MemoryUsage:   512 * 1024 * 1024, // 512MB
				DiskIO:        2 * 1024 * 1024,   // 2MB
				NetworkIO:     50 * 1024,         // 50KB
				MaxMemoryUsed: 600 * 1024 * 1024, // 600MB
			},
			BuildSteps: []BuildStep{
				{
					Name:        "compileJava",
					Duration:    30 * time.Second,
					MemoryUsage: 256 * 1024 * 1024,
					CPUUsage:    0.75,
					CacheHit:    true,
				},
				{
					Name:        "test",
					Duration:    45 * time.Second,
					MemoryUsage: 128 * 1024 * 1024,
					CPUUsage:    0.60,
					CacheHit:    false,
				},
			},
		},
		RequestID: "req-123",
		Timestamp: time.Now().Truncate(time.Millisecond),
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal BuildResponse: %v", err)
	}

	// Unmarshal from JSON
	var decoded BuildResponse
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal BuildResponse: %v", err)
	}

	// Compare basic fields
	if decoded.Success != original.Success {
		t.Errorf("Success mismatch: expected %v, got %v", original.Success, decoded.Success)
	}

	if decoded.WorkerID != original.WorkerID {
		t.Errorf("WorkerID mismatch: expected %s, got %s", original.WorkerID, decoded.WorkerID)
	}

	if decoded.BuildDuration != original.BuildDuration {
		t.Errorf("BuildDuration mismatch: expected %v, got %v", original.BuildDuration, decoded.BuildDuration)
	}

	if len(decoded.Artifacts) != len(original.Artifacts) {
		t.Errorf("Artifacts length mismatch: expected %d, got %d", len(original.Artifacts), len(decoded.Artifacts))
	}

	for i, artifact := range original.Artifacts {
		if decoded.Artifacts[i] != artifact {
			t.Errorf("Artifacts[%d] mismatch: expected %s, got %s", i, artifact, decoded.Artifacts[i])
		}
	}

	if decoded.ErrorMessage != original.ErrorMessage {
		t.Errorf("ErrorMessage mismatch: expected %s, got %s", original.ErrorMessage, decoded.ErrorMessage)
	}

	if decoded.RequestID != original.RequestID {
		t.Errorf("RequestID mismatch: expected %s, got %s", original.RequestID, decoded.RequestID)
	}

	if !decoded.Timestamp.Equal(original.Timestamp) {
		t.Errorf("Timestamp mismatch: expected %v, got %v", original.Timestamp, decoded.Timestamp)
	}

	// Compare metrics
	if decoded.Metrics.CacheHitRate != original.Metrics.CacheHitRate {
		t.Errorf("Metrics.CacheHitRate mismatch: expected %f, got %f", original.Metrics.CacheHitRate, decoded.Metrics.CacheHitRate)
	}

	if decoded.Metrics.CompiledFiles != original.Metrics.CompiledFiles {
		t.Errorf("Metrics.CompiledFiles mismatch: expected %d, got %d", original.Metrics.CompiledFiles, decoded.Metrics.CompiledFiles)
	}

	// Compare test results
	if decoded.Metrics.TestResults.TotalTests != original.Metrics.TestResults.TotalTests {
		t.Errorf("TestResults.TotalTests mismatch: expected %d, got %d", original.Metrics.TestResults.TotalTests, decoded.Metrics.TestResults.TotalTests)
	}

	if decoded.Metrics.TestResults.PassedTests != original.Metrics.TestResults.PassedTests {
		t.Errorf("TestResults.PassedTests mismatch: expected %d, got %d", original.Metrics.TestResults.PassedTests, decoded.Metrics.TestResults.PassedTests)
	}

	if decoded.Metrics.TestResults.FailedTests != original.Metrics.TestResults.FailedTests {
		t.Errorf("TestResults.FailedTests mismatch: expected %d, got %d", original.Metrics.TestResults.FailedTests, decoded.Metrics.TestResults.FailedTests)
	}

	if decoded.Metrics.TestResults.SkippedTests != original.Metrics.TestResults.SkippedTests {
		t.Errorf("TestResults.SkippedTests mismatch: expected %d, got %d", original.Metrics.TestResults.SkippedTests, decoded.Metrics.TestResults.SkippedTests)
	}

	if decoded.Metrics.TestResults.TestDuration != original.Metrics.TestResults.TestDuration {
		t.Errorf("TestResults.TestDuration mismatch: expected %v, got %v", original.Metrics.TestResults.TestDuration, decoded.Metrics.TestResults.TestDuration)
	}

	if decoded.Metrics.TestResults.Coverage != original.Metrics.TestResults.Coverage {
		t.Errorf("TestResults.Coverage mismatch: expected %f, got %f", original.Metrics.TestResults.Coverage, decoded.Metrics.TestResults.Coverage)
	}

	// Compare resource usage
	if decoded.Metrics.ResourceUsage.CPUPercent != original.Metrics.ResourceUsage.CPUPercent {
		t.Errorf("ResourceUsage.CPUPercent mismatch: expected %f, got %f", original.Metrics.ResourceUsage.CPUPercent, decoded.Metrics.ResourceUsage.CPUPercent)
	}

	if decoded.Metrics.ResourceUsage.MemoryUsage != original.Metrics.ResourceUsage.MemoryUsage {
		t.Errorf("ResourceUsage.MemoryUsage mismatch: expected %d, got %d", original.Metrics.ResourceUsage.MemoryUsage, decoded.Metrics.ResourceUsage.MemoryUsage)
	}

	if decoded.Metrics.ResourceUsage.DiskIO != original.Metrics.ResourceUsage.DiskIO {
		t.Errorf("ResourceUsage.DiskIO mismatch: expected %d, got %d", original.Metrics.ResourceUsage.DiskIO, decoded.Metrics.ResourceUsage.DiskIO)
	}

	if decoded.Metrics.ResourceUsage.NetworkIO != original.Metrics.ResourceUsage.NetworkIO {
		t.Errorf("ResourceUsage.NetworkIO mismatch: expected %d, got %d", original.Metrics.ResourceUsage.NetworkIO, decoded.Metrics.ResourceUsage.NetworkIO)
	}

	if decoded.Metrics.ResourceUsage.MaxMemoryUsed != original.Metrics.ResourceUsage.MaxMemoryUsed {
		t.Errorf("ResourceUsage.MaxMemoryUsed mismatch: expected %d, got %d", original.Metrics.ResourceUsage.MaxMemoryUsed, decoded.Metrics.ResourceUsage.MaxMemoryUsed)
	}

	// Compare build steps
	if len(decoded.Metrics.BuildSteps) != len(original.Metrics.BuildSteps) {
		t.Errorf("BuildSteps length mismatch: expected %d, got %d", len(original.Metrics.BuildSteps), len(decoded.Metrics.BuildSteps))
	}

	for i, step := range original.Metrics.BuildSteps {
		if decoded.Metrics.BuildSteps[i].Name != step.Name {
			t.Errorf("BuildSteps[%d].Name mismatch: expected %s, got %s", i, step.Name, decoded.Metrics.BuildSteps[i].Name)
		}

		if decoded.Metrics.BuildSteps[i].Duration != step.Duration {
			t.Errorf("BuildSteps[%d].Duration mismatch: expected %v, got %v", i, step.Duration, decoded.Metrics.BuildSteps[i].Duration)
		}

		if decoded.Metrics.BuildSteps[i].MemoryUsage != step.MemoryUsage {
			t.Errorf("BuildSteps[%d].MemoryUsage mismatch: expected %d, got %d", i, step.MemoryUsage, decoded.Metrics.BuildSteps[i].MemoryUsage)
		}

		if decoded.Metrics.BuildSteps[i].CPUUsage != step.CPUUsage {
			t.Errorf("BuildSteps[%d].CPUUsage mismatch: expected %f, got %f", i, step.CPUUsage, decoded.Metrics.BuildSteps[i].CPUUsage)
		}

		if decoded.Metrics.BuildSteps[i].CacheHit != step.CacheHit {
			t.Errorf("BuildSteps[%d].CacheHit mismatch: expected %v, got %v", i, step.CacheHit, decoded.Metrics.BuildSteps[i].CacheHit)
		}
	}
}

func TestWorkerConfig_JSONSerialization(t *testing.T) {
	original := WorkerConfig{
		ID:                  "worker-123",
		CoordinatorURL:      "http://coordinator:8080",
		HTTPPort:            8081,
		RPCPort:             50051,
		BuildDir:            "/tmp/builds",
		CacheEnabled:        true,
		MaxConcurrentBuilds: 4,
		WorkerType:          "gradle",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal WorkerConfig: %v", err)
	}

	// Unmarshal from JSON
	var decoded WorkerConfig
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal WorkerConfig: %v", err)
	}

	// Compare
	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: expected %s, got %s", original.ID, decoded.ID)
	}

	if decoded.CoordinatorURL != original.CoordinatorURL {
		t.Errorf("CoordinatorURL mismatch: expected %s, got %s", original.CoordinatorURL, decoded.CoordinatorURL)
	}

	if decoded.HTTPPort != original.HTTPPort {
		t.Errorf("HTTPPort mismatch: expected %d, got %d", original.HTTPPort, decoded.HTTPPort)
	}

	if decoded.RPCPort != original.RPCPort {
		t.Errorf("RPCPort mismatch: expected %d, got %d", original.RPCPort, decoded.RPCPort)
	}

	if decoded.BuildDir != original.BuildDir {
		t.Errorf("BuildDir mismatch: expected %s, got %s", original.BuildDir, decoded.BuildDir)
	}

	if decoded.CacheEnabled != original.CacheEnabled {
		t.Errorf("CacheEnabled mismatch: expected %v, got %v", original.CacheEnabled, decoded.CacheEnabled)
	}

	if decoded.MaxConcurrentBuilds != original.MaxConcurrentBuilds {
		t.Errorf("MaxConcurrentBuilds mismatch: expected %d, got %d", original.MaxConcurrentBuilds, decoded.MaxConcurrentBuilds)
	}

	if decoded.WorkerType != original.WorkerType {
		t.Errorf("WorkerType mismatch: expected %s, got %s", original.WorkerType, decoded.WorkerType)
	}
}

func TestCoordinatorConfig_JSONSerialization(t *testing.T) {
	original := CoordinatorConfig{
		HTTPPort:         8080,
		RPCPort:          50050,
		MaxWorkers:       10,
		QueueSize:        100,
		HeartbeatTimeout: 60 * time.Second,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal CoordinatorConfig: %v", err)
	}

	// Unmarshal from JSON
	var decoded CoordinatorConfig
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal CoordinatorConfig: %v", err)
	}

	// Compare
	if decoded.HTTPPort != original.HTTPPort {
		t.Errorf("HTTPPort mismatch: expected %d, got %d", original.HTTPPort, decoded.HTTPPort)
	}

	if decoded.RPCPort != original.RPCPort {
		t.Errorf("RPCPort mismatch: expected %d, got %d", original.RPCPort, decoded.RPCPort)
	}

	if decoded.MaxWorkers != original.MaxWorkers {
		t.Errorf("MaxWorkers mismatch: expected %d, got %d", original.MaxWorkers, decoded.MaxWorkers)
	}

	if decoded.QueueSize != original.QueueSize {
		t.Errorf("QueueSize mismatch: expected %d, got %d", original.QueueSize, decoded.QueueSize)
	}

	if decoded.HeartbeatTimeout != original.HeartbeatTimeout {
		t.Errorf("HeartbeatTimeout mismatch: expected %v, got %v", original.HeartbeatTimeout, decoded.HeartbeatTimeout)
	}
}

func TestCacheConfig_JSONSerialization(t *testing.T) {
	original := CacheConfig{
		Port:            8082,
		StorageType:     "filesystem",
		StorageDir:      "/var/cache/builds",
		RedisAddr:       "redis:6379",
		RedisPassword:   "secret",
		S3Bucket:        "build-cache",
		S3Region:        "us-east-1",
		MaxCacheSize:    10 * 1024 * 1024 * 1024, // 10GB
		TTL:             24 * time.Hour,
		CleanupInterval: time.Hour,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal CacheConfig: %v", err)
	}

	// Unmarshal from JSON
	var decoded CacheConfig
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal CacheConfig: %v", err)
	}

	// Compare
	if decoded.Port != original.Port {
		t.Errorf("Port mismatch: expected %d, got %d", original.Port, decoded.Port)
	}

	if decoded.StorageType != original.StorageType {
		t.Errorf("StorageType mismatch: expected %s, got %s", original.StorageType, decoded.StorageType)
	}

	if decoded.StorageDir != original.StorageDir {
		t.Errorf("StorageDir mismatch: expected %s, got %s", original.StorageDir, decoded.StorageDir)
	}

	if decoded.RedisAddr != original.RedisAddr {
		t.Errorf("RedisAddr mismatch: expected %s, got %s", original.RedisAddr, decoded.RedisAddr)
	}

	if decoded.RedisPassword != original.RedisPassword {
		t.Errorf("RedisPassword mismatch: expected %s, got %s", original.RedisPassword, decoded.RedisPassword)
	}

	if decoded.S3Bucket != original.S3Bucket {
		t.Errorf("S3Bucket mismatch: expected %s, got %s", original.S3Bucket, decoded.S3Bucket)
	}

	if decoded.S3Region != original.S3Region {
		t.Errorf("S3Region mismatch: expected %s, got %s", original.S3Region, decoded.S3Region)
	}

	if decoded.MaxCacheSize != original.MaxCacheSize {
		t.Errorf("MaxCacheSize mismatch: expected %d, got %d", original.MaxCacheSize, decoded.MaxCacheSize)
	}

	if decoded.TTL != original.TTL {
		t.Errorf("TTL mismatch: expected %v, got %v", original.TTL, decoded.TTL)
	}

	if decoded.CleanupInterval != original.CleanupInterval {
		t.Errorf("CleanupInterval mismatch: expected %v, got %v", original.CleanupInterval, decoded.CleanupInterval)
	}
}

func TestMonitorConfig_JSONSerialization(t *testing.T) {
	original := MonitorConfig{
		Port:            9090,
		MetricsInterval: 30 * time.Second,
		AlertThresholds: map[string]float64{
			"cpu_usage":    0.8,
			"memory_usage": 0.9,
			"disk_usage":   0.85,
			"build_time":   300, // 5 minutes
		},
		EnablePrometheus: true,
		EnableJaeger:     true,
		JaegerEndpoint:   "jaeger:14268",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal MonitorConfig: %v", err)
	}

	// Unmarshal from JSON
	var decoded MonitorConfig
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal MonitorConfig: %v", err)
	}

	// Compare
	if decoded.Port != original.Port {
		t.Errorf("Port mismatch: expected %d, got %d", original.Port, decoded.Port)
	}

	if decoded.MetricsInterval != original.MetricsInterval {
		t.Errorf("MetricsInterval mismatch: expected %v, got %v", original.MetricsInterval, decoded.MetricsInterval)
	}

	if len(decoded.AlertThresholds) != len(original.AlertThresholds) {
		t.Errorf("AlertThresholds length mismatch: expected %d, got %d", len(original.AlertThresholds), len(decoded.AlertThresholds))
	}

	for k, v := range original.AlertThresholds {
		if decoded.AlertThresholds[k] != v {
			t.Errorf("AlertThresholds[%s] mismatch: expected %f, got %f", k, v, decoded.AlertThresholds[k])
		}
	}

	if decoded.EnablePrometheus != original.EnablePrometheus {
		t.Errorf("EnablePrometheus mismatch: expected %v, got %v", original.EnablePrometheus, decoded.EnablePrometheus)
	}

	if decoded.EnableJaeger != original.EnableJaeger {
		t.Errorf("EnableJaeger mismatch: expected %v, got %v", original.EnableJaeger, decoded.EnableJaeger)
	}

	if decoded.JaegerEndpoint != original.JaegerEndpoint {
		t.Errorf("JaegerEndpoint mismatch: expected %s, got %s", original.JaegerEndpoint, decoded.JaegerEndpoint)
	}
}

func TestWorkerInfo_JSONSerialization(t *testing.T) {
	original := WorkerInfo{
		ID:           "worker-123",
		Address:      "192.168.1.100:8081",
		Status:       "idle",
		Capabilities: []string{"gradle", "maven", "docker"},
		LastSeen:     time.Now().Truncate(time.Millisecond),
		BuildCount:   42,
		IsActive:     true,
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal WorkerInfo: %v", err)
	}

	// Unmarshal from JSON
	var decoded WorkerInfo
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal WorkerInfo: %v", err)
	}

	// Compare
	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: expected %s, got %s", original.ID, decoded.ID)
	}

	if decoded.Address != original.Address {
		t.Errorf("Address mismatch: expected %s, got %s", original.Address, decoded.Address)
	}

	if decoded.Status != original.Status {
		t.Errorf("Status mismatch: expected %s, got %s", original.Status, decoded.Status)
	}

	if len(decoded.Capabilities) != len(original.Capabilities) {
		t.Errorf("Capabilities length mismatch: expected %d, got %d", len(original.Capabilities), len(decoded.Capabilities))
	}

	for i, capability := range original.Capabilities {
		if decoded.Capabilities[i] != capability {
			t.Errorf("Capabilities[%d] mismatch: expected %s, got %s", i, capability, decoded.Capabilities[i])
		}
	}

	if !decoded.LastSeen.Equal(original.LastSeen) {
		t.Errorf("LastSeen mismatch: expected %v, got %v", original.LastSeen, decoded.LastSeen)
	}

	if decoded.BuildCount != original.BuildCount {
		t.Errorf("BuildCount mismatch: expected %d, got %d", original.BuildCount, decoded.BuildCount)
	}

	if decoded.IsActive != original.IsActive {
		t.Errorf("IsActive mismatch: expected %v, got %v", original.IsActive, decoded.IsActive)
	}
}

func TestBuildRequest_DefaultValues(t *testing.T) {
	req := BuildRequest{
		ProjectPath: "/projects/test",
		TaskName:    "build",
		RequestID:   "test-123",
	}

	// Check default values
	if req.CacheEnabled != false {
		t.Errorf("Expected CacheEnabled to default to false, got %v", req.CacheEnabled)
	}

	if req.BuildOptions != nil {
		t.Errorf("Expected BuildOptions to default to nil, got %v", req.BuildOptions)
	}

	if req.WorkerID != "" {
		t.Errorf("Expected WorkerID to default to empty string, got %s", req.WorkerID)
	}

	if !req.Timestamp.IsZero() {
		t.Error("Expected Timestamp to default to zero value")
	}
}

func TestBuildResponse_ErrorCase(t *testing.T) {
	response := BuildResponse{
		Success:      false,
		WorkerID:     "worker-123",
		ErrorMessage: "Build failed: compilation error",
		RequestID:    "req-123",
		Timestamp:    time.Now(),
	}

	// Marshal and unmarshal
	jsonData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal error response: %v", err)
	}

	var decoded BuildResponse
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if decoded.Success != false {
		t.Error("Expected Success to be false for error case")
	}

	if decoded.ErrorMessage != "Build failed: compilation error" {
		t.Errorf("Expected error message 'Build failed: compilation error', got '%s'", decoded.ErrorMessage)
	}

	// Error responses should have empty artifacts
	if decoded.Artifacts != nil && len(decoded.Artifacts) > 0 {
		t.Errorf("Expected empty artifacts for error response, got %v", decoded.Artifacts)
	}
}

func TestBuildMetrics_EmptyBuildSteps(t *testing.T) {
	metrics := BuildMetrics{
		CacheHitRate:  0.0,
		CompiledFiles: 0,
		TestResults: TestResults{
			TotalTests:   0,
			PassedTests:  0,
			FailedTests:  0,
			SkippedTests: 0,
			TestDuration: 0,
			Coverage:     0.0,
		},
		ResourceUsage: ResourceMetrics{
			CPUPercent:    0.0,
			MemoryUsage:   0,
			DiskIO:        0,
			NetworkIO:     0,
			MaxMemoryUsed: 0,
		},
		BuildSteps: []BuildStep{}, // Empty slice
	}

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		t.Fatalf("Failed to marshal metrics with empty build steps: %v", err)
	}

	var decoded BuildMetrics
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal metrics with empty build steps: %v", err)
	}

	if decoded.BuildSteps == nil {
		t.Error("Expected BuildSteps to be empty slice, got nil")
	}

	if len(decoded.BuildSteps) != 0 {
		t.Errorf("Expected 0 build steps, got %d", len(decoded.BuildSteps))
	}
}

func TestTypes_ZeroValues(t *testing.T) {
	// Test that zero values are valid
	var req BuildRequest
	if req.ProjectPath != "" {
		t.Errorf("Expected empty ProjectPath for zero value, got %s", req.ProjectPath)
	}

	var resp BuildResponse
	if resp.Success != false {
		t.Error("Expected Success to be false for zero value")
	}

	var metrics BuildMetrics
	if metrics.CacheHitRate != 0.0 {
		t.Errorf("Expected CacheHitRate to be 0.0 for zero value, got %f", metrics.CacheHitRate)
	}

	var workerInfo WorkerInfo
	if workerInfo.Status != "" {
		t.Errorf("Expected empty Status for zero value, got %s", workerInfo.Status)
	}
}
