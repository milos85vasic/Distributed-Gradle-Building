#!/bin/bash

# Test configuration for integration tests
# This file contains configuration values used by integration tests

# Service ports
COORDINATOR_PORT=8080
WORKER1_PORT=8081
WORKER2_PORT=8082
WORKER3_PORT=8083
CACHE_SERVER_PORT=8084
MONITOR_PORT=8085

# Service directories
COORDINATOR_DIR="$(dirname "$0")/../../coordinator"
WORKER_DIR="$(dirname "$0")/../../worker"
CACHE_DIR="$(dirname "$0")/../../cache-server"
MONITOR_DIR="$(dirname "$0")/../../monitor"

# Timeouts
SERVICE_STARTUP_TIMEOUT=30
BUILD_TIMEOUT=300
HEARTBEAT_INTERVAL=5

# Test data
TEST_PROJECT_DIR="$(dirname "$0")/fixtures/demo_projects/java_small"

# Utility functions
wait_for_service() {
    local url=$1
    local timeout=$2
    local interval=1
    local elapsed=0
    
    echo "Waiting for service at $url..."
    while [ $elapsed -lt $timeout ]; do
        if curl -s "$url" >/dev/null 2>&1; then
            echo "Service is ready"
            return 0
        fi
        
        sleep $interval
        elapsed=$((elapsed + interval))
    done
    
    echo "Service not ready after $timeout seconds"
    return 1
}