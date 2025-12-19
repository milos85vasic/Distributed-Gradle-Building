package main

import (
	"log"
)

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
	err = service.StartRPCServer()
	if err != nil {
		log.Fatalf("Failed to start RPC server: %v", err)
	}
}

// Main function for worker
func main() {
	workerMain()
}