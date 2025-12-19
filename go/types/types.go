package types

import (
	"sync"
	"time"
)

// BuildRequest represents a distributed build request
type BuildRequest struct {
	ProjectPath  string            `json:"project_path"`
	TaskName     string            `json:"task_name"`
	WorkerID     string            `json:"worker_id"`
	CacheEnabled bool              `json:"cache_enabled"`
	BuildOptions map[string]string `json:"build_options"`
	Timestamp    time.Time         `json:"timestamp"`
	RequestID    string            `json:"request_id"`
}

// BuildResponse represents response from a build worker
type BuildResponse struct {
	Success       bool          `json:"success"`
	WorkerID      string        `json:"worker_id"`
	BuildDuration time.Duration `json:"build_duration"`
	Artifacts     []string      `json:"artifacts"`
	ErrorMessage  string        `json:"error_message"`
	Metrics       BuildMetrics  `json:"metrics"`
	RequestID     string        `json:"request_id"`
	Timestamp     time.Time     `json:"timestamp"`
}

// BuildMetrics contains detailed build performance metrics
type BuildMetrics struct {
	BuildSteps    []BuildStep     `json:"build_steps"`
	CacheHitRate  float64         `json:"cache_hit_rate"`
	CompiledFiles int             `json:"compiled_files"`
	TestResults   TestResults     `json:"test_results"`
	ResourceUsage ResourceMetrics `json:"resource_usage"`
}

// BuildStep represents individual build step metrics
type BuildStep struct {
	Name        string        `json:"name"`
	Duration    time.Duration `json:"duration"`
	MemoryUsage int64         `json:"memory_usage"`
	CPUUsage    float64       `json:"cpu_usage"`
	CacheHit    bool          `json:"cache_hit"`
}

// TestResults contains test execution results
type TestResults struct {
	TotalTests   int           `json:"total_tests"`
	PassedTests  int           `json:"passed_tests"`
	FailedTests  int           `json:"failed_tests"`
	SkippedTests int           `json:"skipped_tests"`
	TestDuration time.Duration `json:"test_duration"`
	Coverage     float64       `json:"coverage"`
}

// ResourceMetrics tracks resource utilization during build
type ResourceMetrics struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryUsage   int64   `json:"memory_usage"`
	DiskIO        int64   `json:"disk_io"`
	NetworkIO     int64   `json:"network_io"`
	MaxMemoryUsed int64   `json:"max_memory_used"`
}

// WorkerInfo contains information about registered workers
type WorkerInfo struct {
	ID           string    `json:"id"`
	Address      string    `json:"address"`
	Status       string    `json:"status"`
	Capabilities []string  `json:"capabilities"`
	LastSeen     time.Time `json:"last_seen"`
	BuildCount   int       `json:"build_count"`
	IsActive     bool      `json:"is_active"`
}

// WorkerPool manages distributed build workers
type WorkerPool struct {
	Workers    map[string]*Worker `json:"workers"`
	MaxWorkers int                `json:"max_workers"`
	Mutex      sync.RWMutex       `json:"-"`
}

// Worker represents a distributed build worker
type Worker struct {
	ID           string          `json:"id"`
	Host         string          `json:"host"`
	Port         int             `json:"port"`
	Status       string          `json:"status"`
	LastCheckin  time.Time       `json:"last_checkin"`
	BuildCount   int             `json:"build_count"`
	Capabilities []string        `json:"capabilities"`
	Resources    ResourceMetrics `json:"resources"`
}

// CoordinatorConfig holds configuration for coordinator
type CoordinatorConfig struct {
	HTTPPort         int           `json:"http_port"`
	RPCPort          int           `json:"rpc_port"`
	MaxWorkers       int           `json:"max_workers"`
	QueueSize        int           `json:"queue_size"`
	HeartbeatTimeout time.Duration `json:"heartbeat_timeout"`
}

// WorkerConfig holds configuration for worker nodes
type WorkerConfig struct {
	ID                  string `json:"id"`
	CoordinatorURL      string `json:"coordinator_url"`
	HTTPPort            int    `json:"http_port"`
	RPCPort             int    `json:"rpc_port"`
	BuildDir            string `json:"build_dir"`
	CacheEnabled        bool   `json:"cache_enabled"`
	MaxConcurrentBuilds int    `json:"max_concurrent_builds"`
	WorkerType          string `json:"worker_type"`
}

// CacheConfig holds configuration for cache server
type CacheConfig struct {
	Port            int           `json:"port"`
	StorageType     string        `json:"storage_type"`
	StorageDir      string        `json:"storage_dir"`
	RedisAddr       string        `json:"redis_addr"`
	RedisPassword   string        `json:"redis_password"`
	S3Bucket        string        `json:"s3_bucket"`
	S3Region        string        `json:"s3_region"`
	MaxCacheSize    int64         `json:"max_cache_size"`
	TTL             time.Duration `json:"ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// MonitorConfig holds configuration for monitoring service
type MonitorConfig struct {
	Port             int                `json:"port"`
	MetricsInterval  time.Duration      `json:"metrics_interval"`
	AlertThresholds  map[string]float64 `json:"alert_thresholds"`
	EnablePrometheus bool               `json:"enable_prometheus"`
	EnableJaeger     bool               `json:"enable_jaeger"`
	JaegerEndpoint   string             `json:"jaeger_endpoint"`
}

// ClientConfig holds configuration for HTTP client
type ClientConfig struct {
	BaseURL     string        `json:"base_url"`
	Timeout     time.Duration `json:"timeout"`
	MaxRetries  int           `json:"max_retries"`
	RetryDelay  time.Duration `json:"retry_delay"`
	AuthEnabled bool          `json:"auth_enabled"`
	AuthToken   string        `json:"auth_token"`
}
