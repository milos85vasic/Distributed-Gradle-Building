package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"distributed-gradle-building/types"
)

func TestLoadCoordinatorConfig(t *testing.T) {
	// Test default config
	config, err := LoadCoordinatorConfig("")
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	if config.HTTPPort != 8080 {
		t.Errorf("Expected HTTPPort 8080, got %d", config.HTTPPort)
	}
	if config.RPCPort != 8081 {
		t.Errorf("Expected RPCPort 8081, got %d", config.RPCPort)
	}
	if config.MaxWorkers != 10 {
		t.Errorf("Expected MaxWorkers 10, got %d", config.MaxWorkers)
	}
	if config.QueueSize != 100 {
		t.Errorf("Expected QueueSize 100, got %d", config.QueueSize)
	}
	if config.HeartbeatTimeout != 30*time.Second {
		t.Errorf("Expected HeartbeatTimeout 30s, got %v", config.HeartbeatTimeout)
	}

	// Test environment variable overrides
	os.Setenv("COORDINATOR_HTTP_PORT", "9090")
	os.Setenv("COORDINATOR_RPC_PORT", "9091")
	os.Setenv("COORDINATOR_MAX_WORKERS", "20")
	os.Setenv("COORDINATOR_QUEUE_SIZE", "200")
	os.Setenv("COORDINATOR_HEARTBEAT_TIMEOUT", "60s")

	config, err = LoadCoordinatorConfig("")
	if err != nil {
		t.Fatalf("Failed to load config with env vars: %v", err)
	}

	if config.HTTPPort != 9090 {
		t.Errorf("Expected HTTPPort 9090 from env, got %d", config.HTTPPort)
	}
	if config.RPCPort != 9091 {
		t.Errorf("Expected RPCPort 9091 from env, got %d", config.RPCPort)
	}
	if config.MaxWorkers != 20 {
		t.Errorf("Expected MaxWorkers 20 from env, got %d", config.MaxWorkers)
	}
	if config.QueueSize != 200 {
		t.Errorf("Expected QueueSize 200 from env, got %d", config.QueueSize)
	}
	if config.HeartbeatTimeout != 60*time.Second {
		t.Errorf("Expected HeartbeatTimeout 60s from env, got %v", config.HeartbeatTimeout)
	}

	// Clean up environment variables
	os.Unsetenv("COORDINATOR_HTTP_PORT")
	os.Unsetenv("COORDINATOR_RPC_PORT")
	os.Unsetenv("COORDINATOR_MAX_WORKERS")
	os.Unsetenv("COORDINATOR_QUEUE_SIZE")
	os.Unsetenv("COORDINATOR_HEARTBEAT_TIMEOUT")
}

func TestLoadWorkerConfig(t *testing.T) {
	// Test default config
	config, err := LoadWorkerConfig("")
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	if config.ID == "" {
		t.Error("Expected non-empty worker ID")
	}
	if config.CoordinatorURL != "http://localhost:8080" {
		t.Errorf("Expected CoordinatorURL 'http://localhost:8080', got %s", config.CoordinatorURL)
	}
	if config.HTTPPort != 9000 {
		t.Errorf("Expected HTTPPort 9000, got %d", config.HTTPPort)
	}
	if config.RPCPort != 9001 {
		t.Errorf("Expected RPCPort 9001, got %d", config.RPCPort)
	}
	if config.BuildDir != "/tmp/builds" {
		t.Errorf("Expected BuildDir '/tmp/builds', got %s", config.BuildDir)
	}
	if !config.CacheEnabled {
		t.Error("Expected CacheEnabled true")
	}
	if config.MaxConcurrentBuilds != 2 {
		t.Errorf("Expected MaxConcurrentBuilds 2, got %d", config.MaxConcurrentBuilds)
	}
	if config.WorkerType != "gradle" {
		t.Errorf("Expected WorkerType 'gradle', got %s", config.WorkerType)
	}

	// Test environment variable overrides
	os.Setenv("WORKER_ID", "test-worker-123")
	os.Setenv("WORKER_COORDINATOR_URL", "http://coordinator:8080")
	os.Setenv("WORKER_HTTP_PORT", "9100")
	os.Setenv("WORKER_RPC_PORT", "9101")
	os.Setenv("WORKER_BUILD_DIR", "/var/builds")
	os.Setenv("WORKER_CACHE_ENABLED", "false")
	os.Setenv("WORKER_MAX_CONCURRENT_BUILDS", "5")
	os.Setenv("WORKER_TYPE", "maven")

	config, err = LoadWorkerConfig("")
	if err != nil {
		t.Fatalf("Failed to load config with env vars: %v", err)
	}

	if config.ID != "test-worker-123" {
		t.Errorf("Expected ID 'test-worker-123' from env, got %s", config.ID)
	}
	if config.CoordinatorURL != "http://coordinator:8080" {
		t.Errorf("Expected CoordinatorURL 'http://coordinator:8080' from env, got %s", config.CoordinatorURL)
	}
	if config.HTTPPort != 9100 {
		t.Errorf("Expected HTTPPort 9100 from env, got %d", config.HTTPPort)
	}
	if config.RPCPort != 9101 {
		t.Errorf("Expected RPCPort 9101 from env, got %d", config.RPCPort)
	}
	if config.BuildDir != "/var/builds" {
		t.Errorf("Expected BuildDir '/var/builds' from env, got %s", config.BuildDir)
	}
	if config.CacheEnabled {
		t.Error("Expected CacheEnabled false from env")
	}
	if config.MaxConcurrentBuilds != 5 {
		t.Errorf("Expected MaxConcurrentBuilds 5 from env, got %d", config.MaxConcurrentBuilds)
	}
	if config.WorkerType != "maven" {
		t.Errorf("Expected WorkerType 'maven' from env, got %s", config.WorkerType)
	}

	// Clean up environment variables
	os.Unsetenv("WORKER_ID")
	os.Unsetenv("WORKER_COORDINATOR_URL")
	os.Unsetenv("WORKER_HTTP_PORT")
	os.Unsetenv("WORKER_RPC_PORT")
	os.Unsetenv("WORKER_BUILD_DIR")
	os.Unsetenv("WORKER_CACHE_ENABLED")
	os.Unsetenv("WORKER_MAX_CONCURRENT_BUILDS")
	os.Unsetenv("WORKER_TYPE")
}

func TestLoadCacheConfig(t *testing.T) {
	// Test default config
	config, err := LoadCacheConfig("")
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	if config.Port != 8082 {
		t.Errorf("Expected Port 8082, got %d", config.Port)
	}
	if config.StorageType != "filesystem" {
		t.Errorf("Expected StorageType 'filesystem', got %s", config.StorageType)
	}
	if config.StorageDir != "/tmp/cache" {
		t.Errorf("Expected StorageDir '/tmp/cache', got %s", config.StorageDir)
	}
	if config.RedisAddr != "localhost:6379" {
		t.Errorf("Expected RedisAddr 'localhost:6379', got %s", config.RedisAddr)
	}
	if config.RedisPassword != "" {
		t.Errorf("Expected empty RedisPassword, got %s", config.RedisPassword)
	}
	if config.S3Bucket != "" {
		t.Errorf("Expected empty S3Bucket, got %s", config.S3Bucket)
	}
	if config.S3Region != "us-west-2" {
		t.Errorf("Expected S3Region 'us-west-2', got %s", config.S3Region)
	}
	if config.MaxCacheSize != 1024*1024*1024 {
		t.Errorf("Expected MaxCacheSize 1GB, got %d", config.MaxCacheSize)
	}
	if config.TTL != time.Hour {
		t.Errorf("Expected TTL 1 hour, got %v", config.TTL)
	}

	// Test environment variable overrides
	os.Setenv("CACHE_PORT", "8282")
	os.Setenv("CACHE_STORAGE_TYPE", "redis")
	os.Setenv("CACHE_STORAGE_DIR", "/var/cache")
	os.Setenv("CACHE_REDIS_ADDR", "redis:6379")
	os.Setenv("CACHE_REDIS_PASSWORD", "secret")
	os.Setenv("CACHE_S3_BUCKET", "my-bucket")
	os.Setenv("CACHE_S3_REGION", "us-east-1")
	os.Setenv("CACHE_MAX_SIZE", "2147483648") // 2GB
	os.Setenv("CACHE_TTL", "7200")            // 2 hours in seconds

	config, err = LoadCacheConfig("")
	if err != nil {
		t.Fatalf("Failed to load config with env vars: %v", err)
	}

	if config.Port != 8282 {
		t.Errorf("Expected Port 8282 from env, got %d", config.Port)
	}
	if config.StorageType != "redis" {
		t.Errorf("Expected StorageType 'redis' from env, got %s", config.StorageType)
	}
	if config.StorageDir != "/var/cache" {
		t.Errorf("Expected StorageDir '/var/cache' from env, got %s", config.StorageDir)
	}
	if config.RedisAddr != "redis:6379" {
		t.Errorf("Expected RedisAddr 'redis:6379' from env, got %s", config.RedisAddr)
	}
	if config.RedisPassword != "secret" {
		t.Errorf("Expected RedisPassword 'secret' from env, got %s", config.RedisPassword)
	}
	if config.S3Bucket != "my-bucket" {
		t.Errorf("Expected S3Bucket 'my-bucket' from env, got %s", config.S3Bucket)
	}
	if config.S3Region != "us-east-1" {
		t.Errorf("Expected S3Region 'us-east-1' from env, got %s", config.S3Region)
	}
	if config.MaxCacheSize != 2147483648 {
		t.Errorf("Expected MaxCacheSize 2GB from env, got %d", config.MaxCacheSize)
	}
	if config.TTL != 2*time.Hour {
		t.Errorf("Expected TTL 2 hours from env, got %v", config.TTL)
	}

	// Clean up environment variables
	os.Unsetenv("CACHE_PORT")
	os.Unsetenv("CACHE_STORAGE_TYPE")
	os.Unsetenv("CACHE_STORAGE_DIR")
	os.Unsetenv("CACHE_REDIS_ADDR")
	os.Unsetenv("CACHE_REDIS_PASSWORD")
	os.Unsetenv("CACHE_S3_BUCKET")
	os.Unsetenv("CACHE_S3_REGION")
	os.Unsetenv("CACHE_MAX_SIZE")
	os.Unsetenv("CACHE_TTL")
}

func TestLoadMonitorConfig(t *testing.T) {
	// Test default config
	config, err := LoadMonitorConfig("")
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	if config.Port != 8083 {
		t.Errorf("Expected Port 8083, got %d", config.Port)
	}
	if config.MetricsInterval != 30*time.Second {
		t.Errorf("Expected MetricsInterval 30s, got %v", config.MetricsInterval)
	}
	if len(config.AlertThresholds) != 4 {
		t.Errorf("Expected 4 alert thresholds, got %d", len(config.AlertThresholds))
	}
	if config.AlertThresholds["cpu_usage"] != 80.0 {
		t.Errorf("Expected cpu_usage threshold 80.0, got %f", config.AlertThresholds["cpu_usage"])
	}
	if config.AlertThresholds["memory_usage"] != 85.0 {
		t.Errorf("Expected memory_usage threshold 85.0, got %f", config.AlertThresholds["memory_usage"])
	}
	if config.AlertThresholds["error_rate"] != 5.0 {
		t.Errorf("Expected error_rate threshold 5.0, got %f", config.AlertThresholds["error_rate"])
	}
	if config.AlertThresholds["response_time"] != 1000.0 {
		t.Errorf("Expected response_time threshold 1000.0, got %f", config.AlertThresholds["response_time"])
	}
	if config.EnablePrometheus {
		t.Error("Expected EnablePrometheus false")
	}
	if config.EnableJaeger {
		t.Error("Expected EnableJaeger false")
	}
	if config.JaegerEndpoint != "http://localhost:14268/api/traces" {
		t.Errorf("Expected JaegerEndpoint 'http://localhost:14268/api/traces', got %s", config.JaegerEndpoint)
	}

	// Test environment variable overrides
	os.Setenv("MONITOR_PORT", "8383")
	os.Setenv("MONITOR_METRICS_INTERVAL", "60s")
	os.Setenv("MONITOR_ENABLE_PROMETHEUS", "true")
	os.Setenv("MONITOR_ENABLE_JAEGER", "true")
	os.Setenv("MONITOR_JAEGER_ENDPOINT", "http://jaeger:14268/api/traces")

	config, err = LoadMonitorConfig("")
	if err != nil {
		t.Fatalf("Failed to load config with env vars: %v", err)
	}

	if config.Port != 8383 {
		t.Errorf("Expected Port 8383 from env, got %d", config.Port)
	}
	if config.MetricsInterval != 60*time.Second {
		t.Errorf("Expected MetricsInterval 60s from env, got %v", config.MetricsInterval)
	}
	if !config.EnablePrometheus {
		t.Error("Expected EnablePrometheus true from env")
	}
	if !config.EnableJaeger {
		t.Error("Expected EnableJaeger true from env")
	}
	if config.JaegerEndpoint != "http://jaeger:14268/api/traces" {
		t.Errorf("Expected JaegerEndpoint 'http://jaeger:14268/api/traces' from env, got %s", config.JaegerEndpoint)
	}

	// Clean up environment variables
	os.Unsetenv("MONITOR_PORT")
	os.Unsetenv("MONITOR_METRICS_INTERVAL")
	os.Unsetenv("MONITOR_ENABLE_PROMETHEUS")
	os.Unsetenv("MONITOR_ENABLE_JAEGER")
	os.Unsetenv("MONITOR_JAEGER_ENDPOINT")
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.json")

	configData := `{
		"http_port": 9090,
		"rpc_port": 9091,
		"max_workers": 20,
		"queue_size": 200,
		"heartbeat_timeout": 60000000000
	}`

	if err := os.WriteFile(configFile, []byte(configData), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Load config from file
	var config types.CoordinatorConfig
	if err := loadConfigFromFile(configFile, &config); err != nil {
		t.Fatalf("Failed to load config from file: %v", err)
	}

	if config.HTTPPort != 9090 {
		t.Errorf("Expected HTTPPort 9090 from file, got %d", config.HTTPPort)
	}
	if config.RPCPort != 9091 {
		t.Errorf("Expected RPCPort 9091 from file, got %d", config.RPCPort)
	}
	if config.MaxWorkers != 20 {
		t.Errorf("Expected MaxWorkers 20 from file, got %d", config.MaxWorkers)
	}
	if config.QueueSize != 200 {
		t.Errorf("Expected QueueSize 200 from file, got %d", config.QueueSize)
	}
	if config.HeartbeatTimeout != 60*time.Second {
		t.Errorf("Expected HeartbeatTimeout 60s from file, got %v", config.HeartbeatTimeout)
	}

	// Test with non-existent file (should not error)
	var config2 types.CoordinatorConfig
	if err := loadConfigFromFile("/non/existent/file.json", &config2); err != nil {
		t.Errorf("Expected no error for non-existent file, got %v", err)
	}
}

func TestSaveConfig(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "saved_config.json")

	// Create config to save
	config := &types.CoordinatorConfig{
		HTTPPort:         9090,
		RPCPort:          9091,
		MaxWorkers:       20,
		QueueSize:        200,
		HeartbeatTimeout: 60 * time.Second,
	}

	// Save config
	if err := SaveConfig(configFile, config); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load and verify the saved config
	fileData, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read saved config file: %v", err)
	}

	// Check that it contains expected data
	if string(fileData) == "" {
		t.Fatal("Saved config file is empty")
	}

	// Try to load it back
	var loadedConfig types.CoordinatorConfig
	if err := loadConfigFromFile(configFile, &loadedConfig); err != nil {
		t.Fatalf("Failed to load saved config: %v", err)
	}

	if loadedConfig.HTTPPort != config.HTTPPort {
		t.Errorf("Saved HTTPPort doesn't match: expected %d, got %d", config.HTTPPort, loadedConfig.HTTPPort)
	}
	if loadedConfig.RPCPort != config.RPCPort {
		t.Errorf("Saved RPCPort doesn't match: expected %d, got %d", config.RPCPort, loadedConfig.RPCPort)
	}
}

func TestConfigFileAndEnvVarIntegration(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "coordinator_config.json")

	configData := `{
		"http_port": 7070,
		"rpc_port": 7071,
		"max_workers": 15
	}`

	if err := os.WriteFile(configFile, []byte(configData), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Set environment variable that should override file config
	os.Setenv("COORDINATOR_HTTP_PORT", "8080")

	// Load config - env var should override file
	config, err := LoadCoordinatorConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to load config with file and env: %v", err)
	}

	// HTTP port should come from env var (override)
	if config.HTTPPort != 8080 {
		t.Errorf("Expected HTTPPort 8080 from env var override, got %d", config.HTTPPort)
	}

	// RPC port should come from file (no env var override)
	if config.RPCPort != 7071 {
		t.Errorf("Expected RPCPort 7071 from file, got %d", config.RPCPort)
	}

	// Max workers should come from file (no env var override)
	if config.MaxWorkers != 15 {
		t.Errorf("Expected MaxWorkers 15 from file, got %d", config.MaxWorkers)
	}

	// Clean up
	os.Unsetenv("COORDINATOR_HTTP_PORT")
}
