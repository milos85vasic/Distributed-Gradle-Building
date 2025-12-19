package main

import (
	"log"
	"os"
)

// Main monitor application entry point
func monitorMain() {
	// Load configuration
	configFile := "monitor_config.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	config, err := loadMonitorConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load monitor config: %v", err)
	}

	// Create and start monitor
	monitor := NewMonitor(config)

	log.Printf("Starting monitor on port %d", config.Port)
	err = monitor.Start()
	if err != nil {
		log.Fatalf("Monitor failed: %v", err)
	}
}

// Main function for monitor
func main() {
	monitorMain()
}
