package main

import (
	"encoding/json"
	"log"
	"net"
	"net/rpc"
	"os"
)

// WorkerConfig contains worker configuration
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

// WorkerService represents a build worker
type WorkerService struct {
	config     *WorkerConfig
	rpcServer  *rpc.Server
	httpServer *net.Listener
}

// loadWorkerConfig loads worker configuration from file
func loadWorkerConfig(filename string) (*WorkerConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		// Return default config if file doesn't exist
		if os.IsNotExist(err) {
			return &WorkerConfig{
				ID:                  "worker-1",
				CoordinatorURL:      "http://localhost:8080",
				HTTPPort:            8081,
				RPCPort:             8082,
				BuildDir:            "/tmp/worker-builds",
				CacheEnabled:        true,
				MaxConcurrentBuilds: 4,
				WorkerType:          "standard",
			}, nil
		}
		return nil, err
	}

	var config WorkerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// NewWorkerService creates a new worker service
func NewWorkerService(config *WorkerConfig) *WorkerService {
	return &WorkerService{
		config: config,
	}
}

// registerWithCoordinator registers the worker with the coordinator
func (ws *WorkerService) registerWithCoordinator() error {
	log.Printf("Registering worker %s with coordinator", ws.config.ID)
	return nil // Simplified implementation
}

// startRPCServer starts the RPC server
func (ws *WorkerService) startRPCServer() error {
	log.Printf("Starting RPC server on port %d", ws.config.RPCPort)
	return nil // Simplified implementation
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
}

// Main function for worker
func main() {
	workerMain()
}
