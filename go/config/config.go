package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"distributed-gradle-building/types"
)

// LoadCoordinatorConfig loads coordinator configuration with environment variable overrides
func LoadCoordinatorConfig(configPath string) (*types.CoordinatorConfig, error) {
	config := &types.CoordinatorConfig{
		HTTPPort:         8080,
		RPCPort:          8081,
		MaxWorkers:       10,
		QueueSize:        100,
		HeartbeatTimeout: 30 * time.Second,
	}

	// Load from file if exists
	if configPath != "" {
		if err := loadConfigFromFile(configPath, config); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Apply environment variable overrides
	if port := os.Getenv("COORDINATOR_HTTP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.HTTPPort = p
		}
	}

	if port := os.Getenv("COORDINATOR_RPC_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.RPCPort = p
		}
	}

	if workers := os.Getenv("COORDINATOR_MAX_WORKERS"); workers != "" {
		if w, err := strconv.Atoi(workers); err == nil {
			config.MaxWorkers = w
		}
	}

	if queue := os.Getenv("COORDINATOR_QUEUE_SIZE"); queue != "" {
		if q, err := strconv.Atoi(queue); err == nil {
			config.QueueSize = q
		}
	}

	if timeout := os.Getenv("COORDINATOR_HEARTBEAT_TIMEOUT"); timeout != "" {
		if t, err := time.ParseDuration(timeout); err == nil {
			config.HeartbeatTimeout = t
		}
	}

	return config, nil
}

// LoadWorkerConfig loads worker configuration with environment variable overrides
func LoadWorkerConfig(configPath string) (*types.WorkerConfig, error) {
	config := &types.WorkerConfig{
		ID:                  "worker-" + generateWorkerID(),
		CoordinatorURL:      "http://localhost:8080",
		HTTPPort:            9000,
		RPCPort:             9001,
		BuildDir:            "/tmp/builds",
		CacheEnabled:        true,
		MaxConcurrentBuilds: 2,
		WorkerType:          "gradle",
	}

	// Load from file if exists
	if configPath != "" {
		if err := loadConfigFromFile(configPath, config); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Apply environment variable overrides
	if id := os.Getenv("WORKER_ID"); id != "" {
		config.ID = id
	}

	if url := os.Getenv("WORKER_COORDINATOR_URL"); url != "" {
		config.CoordinatorURL = url
	}

	if port := os.Getenv("WORKER_HTTP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.HTTPPort = p
		}
	}

	if port := os.Getenv("WORKER_RPC_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.RPCPort = p
		}
	}

	if dir := os.Getenv("WORKER_BUILD_DIR"); dir != "" {
		config.BuildDir = dir
	}

	if cache := os.Getenv("WORKER_CACHE_ENABLED"); cache != "" {
		if c, err := strconv.ParseBool(cache); err == nil {
			config.CacheEnabled = c
		}
	}

	if concurrent := os.Getenv("WORKER_MAX_CONCURRENT_BUILDS"); concurrent != "" {
		if c, err := strconv.Atoi(concurrent); err == nil {
			config.MaxConcurrentBuilds = c
		}
	}

	if wtype := os.Getenv("WORKER_TYPE"); wtype != "" {
		config.WorkerType = wtype
	}

	return config, nil
}

// LoadCacheConfig loads cache configuration with environment variable overrides
func LoadCacheConfig(configPath string) (*types.CacheConfig, error) {
	config := &types.CacheConfig{
		Port:          8082,
		StorageType:   "filesystem",
		StorageDir:    "/tmp/cache",
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		S3Bucket:      "",
		S3Region:      "us-west-2",
		MaxCacheSize:  1024 * 1024 * 1024, // 1GB
		TTL:           3600,               // 1 hour
	}

	// Load from file if exists
	if configPath != "" {
		if err := loadConfigFromFile(configPath, config); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Apply environment variable overrides
	if port := os.Getenv("CACHE_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Port = p
		}
	}

	if storage := os.Getenv("CACHE_STORAGE_TYPE"); storage != "" {
		config.StorageType = storage
	}

	if dir := os.Getenv("CACHE_STORAGE_DIR"); dir != "" {
		config.StorageDir = dir
	}

	if addr := os.Getenv("CACHE_REDIS_ADDR"); addr != "" {
		config.RedisAddr = addr
	}

	if password := os.Getenv("CACHE_REDIS_PASSWORD"); password != "" {
		config.RedisPassword = password
	}

	if bucket := os.Getenv("CACHE_S3_BUCKET"); bucket != "" {
		config.S3Bucket = bucket
	}

	if region := os.Getenv("CACHE_S3_REGION"); region != "" {
		config.S3Region = region
	}

	if size := os.Getenv("CACHE_MAX_SIZE"); size != "" {
		if s, err := strconv.ParseInt(size, 10, 64); err == nil {
			config.MaxCacheSize = s
		}
	}

	if ttl := os.Getenv("CACHE_TTL"); ttl != "" {
		if t, err := strconv.Atoi(ttl); err == nil {
			config.TTL = time.Duration(t) * time.Second
		}
	}

	return config, nil
}

// LoadMonitorConfig loads monitor configuration with environment variable overrides
func LoadMonitorConfig(configPath string) (*types.MonitorConfig, error) {
	config := &types.MonitorConfig{
		Port:             8083,
		MetricsInterval:  30 * time.Second,
		AlertThresholds:  make(map[string]float64),
		EnablePrometheus: false,
		EnableJaeger:     false,
		JaegerEndpoint:   "http://localhost:14268/api/traces",
	}

	// Set default alert thresholds
	config.AlertThresholds["cpu_usage"] = 80.0
	config.AlertThresholds["memory_usage"] = 85.0
	config.AlertThresholds["error_rate"] = 5.0
	config.AlertThresholds["response_time"] = 1000.0

	// Load from file if exists
	if configPath != "" {
		if err := loadConfigFromFile(configPath, config); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Apply environment variable overrides
	if port := os.Getenv("MONITOR_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.Port = p
		}
	}

	if interval := os.Getenv("MONITOR_METRICS_INTERVAL"); interval != "" {
		if i, err := time.ParseDuration(interval); err == nil {
			config.MetricsInterval = i
		}
	}

	if prometheus := os.Getenv("MONITOR_ENABLE_PROMETHEUS"); prometheus != "" {
		if p, err := strconv.ParseBool(prometheus); err == nil {
			config.EnablePrometheus = p
		}
	}

	if jaeger := os.Getenv("MONITOR_ENABLE_JAEGER"); jaeger != "" {
		if j, err := strconv.ParseBool(jaeger); err == nil {
			config.EnableJaeger = j
		}
	}

	if endpoint := os.Getenv("MONITOR_JAEGER_ENDPOINT"); endpoint != "" {
		config.JaegerEndpoint = endpoint
	}

	return config, nil
}

// loadConfigFromFile loads configuration from JSON file
func loadConfigFromFile(configPath string, config interface{}) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // Config file is optional
	}

	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	return nil
}

// generateWorkerID generates a simple worker ID
func generateWorkerID() string {
	// In a real implementation, use a proper ID generator
	return "unknown"
}

// SaveConfig saves configuration to JSON file
func SaveConfig(configPath string, config interface{}) error {
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	return nil
}
