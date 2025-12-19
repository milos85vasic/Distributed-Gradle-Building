package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"distributed-gradle-building/client"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <coordinator-url> <project-path> [task-name]\n", os.Args[0])
		os.Exit(1)
	}

	coordinatorURL := os.Args[1]
	projectPath := os.Args[2]
	taskName := "build"
	if len(os.Args) > 3 {
		taskName = os.Args[3]
	}

	// Create client
	buildClient := client.NewClient(coordinatorURL)

	// Check system health
	fmt.Println("Checking system health...")
	if err := buildClient.HealthCheck(); err != nil {
		log.Fatalf("Health check failed: %v", err)
	}
	fmt.Println("✓ System is healthy")

	// Get system status
	fmt.Println("Getting system status...")
	status, err := buildClient.GetSystemStatus()
	if err != nil {
		log.Fatalf("Failed to get system status: %v", err)
	}
	fmt.Printf("✓ System has %d workers, %d builds in queue, %d active builds\n",
		status.WorkerCount, status.QueueLength, status.ActiveBuilds)

	// Verify project exists
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		log.Fatalf("Project path does not exist: %s", projectPath)
	}
	fmt.Printf("✓ Project found at: %s\n", projectPath)

	// Submit build request
	fmt.Printf("Submitting build request for task: %s...\n", taskName)
	buildReq := client.BuildRequest{
		ProjectPath:  projectPath,
		TaskName:     taskName,
		CacheEnabled: true,
		BuildOptions: map[string]string{
			"parallel": "true",
		},
	}

	buildResp, err := buildClient.SubmitBuild(buildReq)
	if err != nil {
		log.Fatalf("Failed to submit build: %v", err)
	}
	fmt.Printf("✓ Build submitted with ID: %s\n", buildResp.BuildID)

	// Wait for build completion
	fmt.Println("Waiting for build to complete...")
	finalStatus, err := buildClient.WaitForBuild(buildResp.BuildID, 30*time.Minute)
	if err != nil {
		log.Fatalf("Build wait failed: %v", err)
	}

	if finalStatus.Success {
		fmt.Printf("✓ Build completed successfully in %v\n", finalStatus.Duration)
		fmt.Printf("✓ Worker: %s\n", finalStatus.WorkerID)
		fmt.Printf("✓ Cache hit rate: %.2f%%\n", finalStatus.CacheHitRate*100)

		if len(finalStatus.Artifacts) > 0 {
			fmt.Println("✓ Build artifacts:")
			for _, artifact := range finalStatus.Artifacts {
				relPath, _ := filepath.Rel(projectPath, artifact)
				fmt.Printf("  - %s\n", relPath)
			}
		}
	} else {
		fmt.Printf("✗ Build failed after %v\n", finalStatus.Duration)
		fmt.Printf("✗ Error: %s\n", finalStatus.ErrorMessage)
		os.Exit(1)
	}
}
