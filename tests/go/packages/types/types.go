package types

import "time"

// BuildRequest represents a distributed build request
type BuildRequest struct {
	ProjectPath   string
	TaskName      string
	WorkerID      string
	CacheEnabled  bool
	BuildOptions  map[string]string
	Timestamp     time.Time
	RequestID     string
}

// BuildResponse represents the response from a build worker
type BuildResponse struct {
	Success       bool
	WorkerID      string
	BuildDuration time.Duration
	Artifacts     []string
	ErrorMessage  string
	Metrics       BuildMetrics
	RequestID     string
	Timestamp     time.Time
}

// BuildMetrics contains detailed build performance metrics
type BuildMetrics struct {
	BuildSteps     []BuildStep
	CacheHitRate   float64
	CompiledFiles  int
	TestResults    TestResults
	ResourceUsage  ResourceMetrics
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
	TotalTests    int
	PassedTests   int
	FailedTests   int
	SkippedTests  int
	TestDuration  time.Duration
	Coverage      float64
}

// ResourceMetrics tracks resource utilization during build
type ResourceMetrics struct {
	CPUPercent    float64
	MemoryUsage   int64
	DiskIO        int64
	NetworkIO     int64
	MaxMemoryUsed int64
}

// WorkerInfo contains information about registered workers
type WorkerInfo struct {
	ID           string
	Address      string
	Status       string
	Capabilities []string
	LastSeen     time.Time
	BuildCount   int
	IsActive     bool
}

// CoordinatorConfig holds configuration for the coordinator
type CoordinatorConfig struct {
	HTTPPort       int
	RPCPort        int
	MaxWorkers     int
	QueueSize      int
	HeartbeatTimeout time.Duration
}

// WorkerConfig holds configuration for worker nodes
type WorkerConfig struct {
	ID             string
	CoordinatorURL string
	HTTPPort       int
	RPCPort        int
	BuildDir       string
	CacheEnabled   bool
	MaxConcurrentBuilds int
	WorkerType     string
}

// CacheConfig holds configuration for the cache server
type CacheConfig struct {
	Port            int
	StorageType    string
	StorageDir     string
	MaxSize        int64
	TTL            time.Duration
	CleanupInterval time.Duration
}

// MonitorConfig holds configuration for the monitoring system
type MonitorConfig struct {
	Port            int
	MetricsInterval time.Duration
	AlertThresholds map[string]float64
}

// ClientConfig holds configuration for the HTTP client
type ClientConfig struct {
	BaseURL      string
	Timeout      time.Duration
	MaxRetries   int
	RetryDelay   time.Duration
	AuthEnabled  bool
	AuthToken    string
}