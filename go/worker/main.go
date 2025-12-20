package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// WorkerConfig contains worker configuration
type WorkerConfig struct {
	ID                  string `json:"id"`
	CoordinatorHost     string `json:"coordinator_host"`
	CoordinatorRPCPort  int    `json:"coordinator_rpc_port"`
	HTTPPort            int    `json:"http_port"`
	RPCPort             int    `json:"rpc_port"`
	BuildDir            string `json:"build_dir"`
	CacheEnabled        bool   `json:"cache_enabled"`
	MaxConcurrentBuilds int    `json:"max_concurrent_builds"`
	WorkerType          string `json:"worker_type"`
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

type BuildRequest struct {
	ProjectPath  string
	TaskName     string
	WorkerID     string
	CacheEnabled bool
	BuildOptions map[string]string
	Timestamp    time.Time
	RequestID    string
}

// RPC argument and reply types for heartbeat
type HeartbeatArgs struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type HeartbeatReply struct {
	Message string `json:"message"`
}

// WorkerService represents a build worker
type WorkerService struct {
	config     *WorkerConfig
	rpcServer  *rpc.Server
	httpServer *http.Server
	shutdown   chan struct{}
}

// loadWorkerConfig loads worker configuration from file and environment variables
func loadWorkerConfig(filename string) (*WorkerConfig, error) {
	config := &WorkerConfig{
		ID:                  getEnvOrDefault("WORKER_ID", "worker-1"),
		CoordinatorHost:     getEnvOrDefault("COORDINATOR_HOST", "coordinator"),
		CoordinatorRPCPort:  getEnvIntOrDefault("COORDINATOR_RPC_PORT", 8081),
		HTTPPort:            8080, // Not used in current implementation
		RPCPort:             getEnvIntOrDefault("WORKER_PORT", 8082),
		BuildDir:            getEnvOrDefault("BUILD_DIR", "/tmp/worker-builds"),
		CacheEnabled:        getEnvBoolOrDefault("CACHE_ENABLED", true),
		MaxConcurrentBuilds: getEnvIntOrDefault("MAX_BUILDS", 5),
		WorkerType:          getEnvOrDefault("WORKER_TYPE", "standard"),
	}

	// Try to load from file if it exists
	if data, err := os.ReadFile(filename); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			log.Printf("Warning: failed to parse config file: %v", err)
		}
	}

	return config, nil
}

// Helper functions for environment variables
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// NewWorkerService creates a new worker service
func NewWorkerService(config *WorkerConfig) *WorkerService {
	return &WorkerService{
		config:   config,
		shutdown: make(chan struct{}),
	}
}

// registerWithCoordinator registers the worker with the coordinator
func (ws *WorkerService) registerWithCoordinator() error {
	log.Printf("Registering worker %s with coordinator", ws.config.ID)

	// Connect to coordinator RPC server
	client, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", ws.config.CoordinatorHost, ws.config.CoordinatorRPCPort))
	if err != nil {
		return fmt.Errorf("failed to connect to coordinator: %v", err)
	}
	defer client.Close()

	// Prepare registration args
	args := RegisterWorkerArgs{
		ID:           ws.config.ID,
		Host:         ws.config.ID, // Use worker ID as host for now
		Port:         ws.config.RPCPort,
		Capabilities: []string{"gradle", "java"},
		Status:       "idle",
	}

	var reply RegisterWorkerReply

	// Call RegisterWorker RPC method
	err = client.Call("BuildCoordinator.RegisterWorker", args, &reply)
	if err != nil {
		return fmt.Errorf("RPC registration failed: %v", err)
	}

	log.Printf("Worker %s registered successfully: %s", ws.config.ID, reply.Message)
	return nil
}

// startRPCServer starts the RPC server
func (ws *WorkerService) startRPCServer() error {
	log.Printf("Starting RPC server on port %d", ws.config.RPCPort)

	// Create RPC server
	rpcServer := rpc.NewServer()
	rpcServer.Register(ws)

	// Start listening
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", ws.config.RPCPort))
	if err != nil {
		return fmt.Errorf("failed to start RPC listener: %v", err)
	}

	log.Printf("RPC server listening on port %d", ws.config.RPCPort)
	go rpcServer.Accept(listener)

	return nil
}

// startHTTPServer starts the HTTP server for metrics
func (ws *WorkerService) startHTTPServer() error {
	log.Printf("Starting HTTP server on port %d", ws.config.HTTPPort)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	ws.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", ws.config.HTTPPort),
		Handler: mux,
	}

	log.Printf("HTTP server listening on port %d", ws.config.HTTPPort)
	go func() {
		if err := ws.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

// Heartbeat sends periodic heartbeat to coordinator
func (ws *WorkerService) Heartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Printf("Worker %s sending heartbeat", ws.config.ID)

		// Connect to coordinator RPC
		client, err := rpc.Dial("tcp", fmt.Sprintf("%s:%d", ws.config.CoordinatorHost, ws.config.CoordinatorRPCPort))
		if err != nil {
			log.Printf("Failed to connect to coordinator for heartbeat: %v", err)
			continue
		}

		// Send heartbeat
		args := HeartbeatArgs{
			ID:        ws.config.ID,
			Status:    "idle",
			Timestamp: time.Now(),
		}

		var reply HeartbeatReply
		err = client.Call("BuildCoordinator.Heartbeat", args, &reply)
		client.Close()

		if err != nil {
			log.Printf("Heartbeat failed: %v", err)
		} else {
			log.Printf("Heartbeat successful")
		}
	}
}

// Build executes a build request (RPC method)
func (ws *WorkerService) Build(request BuildRequest, response *string) error {
	log.Printf("Received build request %s for project %s", request.RequestID, request.ProjectPath)

	// Execute the build
	err := ws.executeBuild(request)
	if err != nil {
		log.Printf("Build %s failed: %v", request.RequestID, err)
		return err
	}

	*response = fmt.Sprintf("Build request %s completed successfully", request.RequestID)
	return nil
}

// executeBuild executes a Gradle build
func (ws *WorkerService) executeBuild(request BuildRequest) error {
	log.Printf("Executing build %s in %s", request.RequestID, request.ProjectPath)

	// Change to project directory
	if err := os.Chdir(request.ProjectPath); err != nil {
		return fmt.Errorf("failed to change to project directory: %v", err)
	}

	// Execute gradle build
	cmd := exec.Command("gradle", request.TaskName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gradle build failed: %v", err)
	}

	log.Printf("Build %s completed successfully", request.RequestID)
	return nil
}

// Shutdown gracefully shuts down the worker service
func (ws *WorkerService) Shutdown() error {
	close(ws.shutdown)

	if ws.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return ws.httpServer.Shutdown(ctx)
	}

	return nil
}

// Main worker application entry point
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
	err = service.registerWithCoordinator()
	if err != nil {
		log.Fatalf("Failed to register with coordinator: %v", err)
	}

	// Start RPC server
	err = service.startRPCServer()
	if err != nil {
		log.Fatalf("Failed to start RPC server: %v", err)
	}

	// Start HTTP server for metrics
	err = service.startHTTPServer()
	if err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}

	log.Printf("Worker %s started successfully", config.ID)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal %v, shutting down...", sig)

	// Graceful shutdown
	if err := service.Shutdown(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Printf("Worker %s shutdown complete", config.ID)
}

// Main function for worker
func main() {
	workerMain()
}
