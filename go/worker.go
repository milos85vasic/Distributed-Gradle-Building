package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"sync"
	"time"
)

// WorkerNode represents a distributed build worker
type WorkerNode struct {
	ID           string
	Host         string
	Port         int
	Status       string
	Capabilities []string
	Metrics      WorkerMetrics
	mutex        sync.RWMutex
}

// WorkerMetrics tracks worker performance metrics
type WorkerMetrics struct {
	BuildCount       int
	TotalBuildTime   time.Duration
	CpuUsage         float64
	MemoryUsage      uint64
	DiskUsage        uint64
	LastBuildTime    time.Time
	AverageBuildTime time.Duration
}

// WorkerConfig contains worker configuration
type WorkerConfig struct {
	ID           string `json:"ID"`
	Coordinator  string `json:"CoordinatorURL"`
	HTTPPort     int    `json:"HTTPPort"`
	RPCPort      int    `json:"RPCPort"`
	BuildDir     string `json:"BuildDir"`
	CacheEnabled bool   `json:"CacheEnabled"`
	MaxBuilds    int    `json:"MaxConcurrentBuilds"`
	WorkerType   string `json:"WorkerType"`
	// Legacy fields for compatibility
	Host         string   `json:"host,omitempty"`
	Port         int      `json:"port,omitempty"`
	CacheDir     string   `json:"cache_dir,omitempty"`
	WorkDir      string   `json:"work_dir,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
	GradleHome   string   `json:"gradle_home,omitempty"`
}

// RPC argument and reply types
type RegisterWorkerArgs struct {
	ID           string   `json:"id"`
	Host         string   `json:"host"`
	Port         int      `json:"port"`
	Capabilities []string `json:"capabilities"`
	Status       string   `json:"status"`
}

type RegisterWorkerReply struct {
	Message string `json:"message"`
}

type HeartbeatArgs struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type HeartbeatReply struct {
	Message string `json:"message"`
}

// BuildPlugin interface for extending worker capabilities
type BuildPlugin interface {
	PreBuild(projectPath string, options map[string]string) error
	PostBuild(projectPath string, artifacts []string) error
	GetName() string
	GetVersion() string
}

// WorkerService implements the RPC service for build workers
type WorkerService struct {
	worker  *WorkerNode
	config  *WorkerConfig
	plugins []BuildPlugin
	builds  chan BuildRequestCoordinator
	mutex   sync.RWMutex
}

// BuildRequestCoordinator forwarded from coordinator
type BuildRequestCoordinator struct {
	ProjectPath  string
	TaskName     string
	WorkerID     string
	CacheEnabled bool
	BuildOptions map[string]string
	Timestamp    time.Time
	RequestID    string
}

// NewWorkerNode creates a new worker node
func NewWorkerNode(config *WorkerConfig) *WorkerNode {
	host := config.Host
	if host == "" {
		host = "localhost"
	}

	port := config.HTTPPort
	if port == 0 {
		port = config.Port // Fallback
	}

	return &WorkerNode{
		ID:           config.ID,
		Host:         host,
		Port:         port,
		Status:       "idle",
		Capabilities: config.Capabilities,
		Metrics: WorkerMetrics{
			BuildCount: 0,
		},
	}
}

// NewWorkerService creates a new worker service
func NewWorkerService(config *WorkerConfig) *WorkerService {
	worker := NewWorkerNode(config)

	service := &WorkerService{
		worker: worker,
		config: config,
		builds: make(chan BuildRequestCoordinator, 10),
	}

	// Load plugins
	service.loadPlugins()

	// Start build processor
	go service.processBuilds()

	return service
}

// loadPlugins loads build plugins from the plugins directory
func (ws *WorkerService) loadPlugins() {
	pluginsDir := "plugins"
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		return
	}

	files, err := os.ReadDir(pluginsDir)
	if err != nil {
		log.Printf("Error reading plugins directory: %v", err)
		return
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".so" {
			pluginPath := filepath.Join(pluginsDir, file.Name())
			if p, err := plugin.Open(pluginPath); err == nil {
				if pluginInterface, err := p.Lookup("Plugin"); err == nil {
					if buildPlugin, ok := pluginInterface.(BuildPlugin); ok {
						ws.plugins = append(ws.plugins, buildPlugin)
						log.Printf("Loaded plugin: %s", buildPlugin.GetName())
					}
				}
			} else {
				log.Printf("Error loading plugin %s: %v", file.Name(), err)
			}
		}
	}
}

// processBuilds processes build requests from the queue
func (ws *WorkerService) processBuilds() {
	for request := range ws.builds {
		go ws.executeBuild(request)
	}
}

// executeBuild executes a Gradle build request
func (ws *WorkerService) executeBuild(request BuildRequestCoordinator) error {
	ws.worker.mutex.Lock()
	ws.worker.Status = "building"
	startTime := time.Now()
	ws.worker.mutex.Unlock()

	defer func() {
		ws.worker.mutex.Lock()
		duration := time.Since(startTime)
		ws.worker.Metrics.BuildCount++
		ws.worker.Metrics.TotalBuildTime += duration
		ws.worker.Metrics.AverageBuildTime = ws.worker.Metrics.TotalBuildTime / time.Duration(ws.worker.Metrics.BuildCount)
		ws.worker.Metrics.LastBuildTime = time.Now()
		ws.worker.Status = "idle"
		ws.worker.mutex.Unlock()
	}()

	// Run pre-build plugins
	for _, plugin := range ws.plugins {
		if err := plugin.PreBuild(request.ProjectPath, request.BuildOptions); err != nil {
			log.Printf("Pre-build plugin %s failed: %v", plugin.GetName(), err)
			return err
		}
	}

	// Prepare build environment
	gradleHome := ws.config.GradleHome
	if gradleHome == "" {
		gradleHome = "/tmp/gradle-" + request.RequestID
		if err := os.RemoveAll(gradleHome); err != nil {
			log.Printf("Error cleaning gradle home: %v", err)
		}
		if err := os.MkdirAll(gradleHome, 0755); err != nil {
			return fmt.Errorf("failed to create gradle home: %v", err)
		}
	}

	// Execute build
	cmd := exec.Command("gradle", request.TaskName)
	cmd.Dir = request.ProjectPath

	// Set environment variables
	env := os.Environ()
	env = append(env, fmt.Sprintf("GRADLE_USER_HOME=%s", gradleHome))
	if request.CacheEnabled {
		env = append(env, fmt.Sprintf("GRADLE_CACHE_ENABLED=true"))
		env = append(env, fmt.Sprintf("GRADLE_CACHE_DIR=%s", ws.config.CacheDir))
	}
	cmd.Env = env

	// Add build options
	for key, value := range request.BuildOptions {
		cmd.Args = append(cmd.Args, fmt.Sprintf("-P%s=%s", key, value))
	}

	output, err := cmd.CombinedOutput()
	log.Printf("Build output for %s: %s", request.RequestID, string(output))

	if err != nil {
		log.Printf("Build failed for %s: %v", request.RequestID, err)
		return err
	}

	// Find build artifacts
	artifacts := ws.findArtifacts(request.ProjectPath)

	// Run post-build plugins
	for _, plugin := range ws.plugins {
		if err := plugin.PostBuild(request.ProjectPath, artifacts); err != nil {
			log.Printf("Post-build plugin %s failed: %v", plugin.GetName(), err)
		}
	}

	// Cleanup
	if ws.config.GradleHome == "" {
		if err := os.RemoveAll(gradleHome); err != nil {
			log.Printf("Error cleaning up gradle home: %v", err)
		}
	}

	return nil
}

// findArtifacts finds build artifacts in the project directory
func (ws *WorkerService) findArtifacts(projectPath string) []string {
	var artifacts []string

	outputDirs := []string{
		filepath.Join(projectPath, "build/libs"),
		filepath.Join(projectPath, "build/distributions"),
		filepath.Join(projectPath, "build/test-results"),
	}

	for _, dir := range outputDirs {
		if _, err := os.Stat(dir); err == nil {
			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() {
					artifacts = append(artifacts, path)
				}
				return nil
			})
		}
	}

	return artifacts
}

// RegisterWithCoordinator registers the worker with the coordinator via HTTP
func (ws *WorkerService) RegisterWithCoordinator() error {
	// For now, use a simple HTTP POST to register
	// TODO: Implement proper HTTP registration endpoint
	log.Printf("Worker %s attempting to register with coordinator at %s", ws.worker.ID, ws.config.Coordinator)
	// Since RPC is not working, we'll skip registration for now
	return nil
}

// Heartbeat sends periodic heartbeat to coordinator
func (ws *WorkerService) Heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Printf("Worker %s sending heartbeat (status: %s)", ws.worker.ID, ws.worker.Status)
		// TODO: Implement HTTP heartbeat
	}
}

// StartRPCServer starts the worker's RPC server
func (ws *WorkerService) StartRPCServer() error {
	rpcServer := rpc.NewServer()
	rpcServer.Register(ws)

	port := ws.config.RPCPort
	if port == 0 {
		port = ws.config.HTTPPort + 1 // Fallback
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	log.Printf("Worker %s RPC server listening on port %d", ws.worker.ID, port)
	go rpcServer.Accept(listener)
	return nil
}

// RPC Methods
func (ws *WorkerService) Build(request BuildRequestCoordinator, response *string) error {
	select {
	case ws.builds <- request:
		*response = fmt.Sprintf("Build request %s queued", request.RequestID)
		return nil
	default:
		return fmt.Errorf("worker build queue is full")
	}
}

func (ws *WorkerService) GetStatus(args struct{}, status *string) error {
	ws.worker.mutex.RLock()
	*status = ws.worker.Status
	ws.worker.mutex.RUnlock()
	return nil
}

func (ws *WorkerService) GetMetrics(args struct{}, metrics *WorkerMetrics) error {
	ws.worker.mutex.RLock()
	*metrics = ws.worker.Metrics
	ws.worker.mutex.RUnlock()
	return nil
}

func workerMain() {
	// Load configuration
	configFile := "worker_config.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	config, err := loadWorkerConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load worker config: %v", err)
	}

	// Create worker service
	service := NewWorkerService(config)

	// Register with coordinator
	if err := service.RegisterWithCoordinator(); err != nil {
		log.Printf("Warning: Failed to register with coordinator: %v", err)
	}

	// Start heartbeat
	go service.Heartbeat()

	// Start RPC server (disabled for now)
	log.Printf("Starting worker %s on %s:%d", config.ID, config.Host, config.HTTPPort)
	log.Printf("Worker %s started successfully - RPC disabled", config.ID)

	// Keep the process running
	select {}
}

// loadWorkerConfig loads worker configuration from file
func loadWorkerConfig(filename string) (*WorkerConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config WorkerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Set defaults
	if config.Host == "" {
		hostname, _ := os.Hostname()
		config.Host = hostname
	}

	if config.Port == 0 {
		config.Port = 8082
	}

	if config.CacheDir == "" {
		config.CacheDir = "/tmp/gradle-cache"
	}

	if config.WorkDir == "" {
		config.WorkDir = "/tmp/gradle-work"
	}

	if config.MaxBuilds == 0 {
		config.MaxBuilds = 5
	}

	if len(config.Capabilities) == 0 {
		config.Capabilities = []string{"gradle", "java", "testing"}
	}

	return &config, nil
}

// Main function for worker
func main() {
	workerMain()
}
