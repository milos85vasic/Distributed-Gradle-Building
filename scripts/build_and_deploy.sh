#!/bin/bash

# Distributed Gradle Building System - Build and Deploy Script
# This script builds and deploys the Go-based distributed build system

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Directory paths
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
GO_DIR="$PROJECT_DIR/go"
LOGS_DIR="$PROJECT_DIR/logs"
DATA_DIR="$PROJECT_DIR/data"

# Configuration
COORDINATOR_PORT=8080
COORDINATOR_RPC_PORT=8081
WORKER_BASE_PORT=8082
CACHE_SERVER_PORT=8083
MONITOR_PORT=8084

# Process tracking
PIDS_FILE="$DATA_DIR/pids.txt"

# Logging
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

# Ensure required directories exist
setup_directories() {
    log "Creating necessary directories..."
    mkdir -p "$LOGS_DIR" "$DATA_DIR" "$GO_DIR/bin"
    
    # Create data directories for services
    mkdir -p "$DATA_DIR/cache-server"
    mkdir -p "$DATA_DIR/monitor"
    mkdir -p "$DATA_DIR/gradle-cache"
}

# Build Go components
build_components() {
    log "Building Go components..."
    
    cd "$GO_DIR"
    
    # Initialize Go module if needed
    if [ ! -f "go.mod" ]; then
        go mod init distributed-gradle-building
    fi
    
    # Download dependencies
    go mod tidy
    
    # Build coordinator
    log "Building build coordinator..."
    go build -o bin/coordinator main.go
    
    # Build worker
    log "Building worker node..."
    go build -o bin/worker worker.go
    
    # Build cache server
    log "Building cache server..."
    go build -o bin/cache_server cache_server.go
    
    # Build monitor
    log "Building monitor..."
    go build -o bin/monitor monitor.go
    
    log "All components built successfully"
}

# Stop all running services
stop_services() {
    log "Stopping existing services..."
    
    if [ -f "$PIDS_FILE" ]; then
        while read -r pid; do
            if kill -0 "$pid" 2>/dev/null; then
                kill "$pid"
                log "Stopped process $pid"
            fi
        done < "$PIDS_FILE"
        rm -f "$PIDS_FILE"
    fi
    
    # Kill any remaining processes on our ports
    lsof -ti :$COORDINATOR_PORT | xargs -r kill 2>/dev/null || true
    lsof -ti :$COORDINATOR_RPC_PORT | xargs -r kill 2>/dev/null || true
    lsof -ti :$WORKER_BASE_PORT | xargs -r kill 2>/dev/null || true
    lsof -ti :$CACHE_SERVER_PORT | xargs -r kill 2>/dev/null || true
    lsof -ti :$MONITOR_PORT | xargs -r kill 2>/dev/null || true
}

# Start services
start_services() {
    log "Starting distributed build services..."
    
    cd "$GO_DIR"
    
    # Start coordinator
    log "Starting build coordinator..."
    ./bin/coordinator > "$LOGS_DIR/coordinator.log" 2>&1 &
    local coordinator_pid=$!
    echo $coordinator_pid >> "$PIDS_FILE"
    log "Coordinator started with PID $coordinator_pid (HTTP: $COORDINATOR_PORT, RPC: $COORDINATOR_RPC_PORT)"
    
    # Wait for coordinator to start
    sleep 2
    
    # Start cache server
    log "Starting cache server..."
    ./bin/cache_server > "$LOGS_DIR/cache_server.log" 2>&1 &
    local cache_pid=$!
    echo $cache_pid >> "$PIDS_FILE"
    log "Cache server started with PID $cache_pid (Port: $CACHE_SERVER_PORT)"
    
    # Start monitor
    log "Starting monitor..."
    ./bin/monitor > "$LOGS_DIR/monitor.log" 2>&1 &
    local monitor_pid=$!
    echo $monitor_pid >> "$PIDS_FILE"
    log "Monitor started with PID $monitor_pid (Port: $MONITOR_PORT)"
    
    # Start workers
    local worker_count=3
    for ((i=1; i<=worker_count; i++)); do
        local worker_port=$((WORKER_BASE_PORT + i - 1))
        local worker_id="worker-$(printf "%03d" $i)"
        
        # Create worker config
        cat > "worker_${i}_config.json" << EOF
{
  "id": "$worker_id",
  "host": "localhost",
  "port": $worker_port,
  "coordinator": "localhost:$COORDINATOR_RPC_PORT",
  "cache_dir": "$DATA_DIR/gradle-cache/worker-$i",
  "work_dir": "$DATA_DIR/gradle-work/worker-$i",
  "max_builds": 5,
  "capabilities": ["gradle", "java", "testing"],
  "gradle_home": "$DATA_DIR/gradle-home/worker-$i"
}
EOF
        
        log "Starting worker $worker_id..."
        ./bin/worker "worker_${i}_config.json" > "$LOGS_DIR/worker_${i}.log" 2>&1 &
        local worker_pid=$!
        echo $worker_pid >> "$PIDS_FILE"
        log "Worker $worker_id started with PID $worker_pid (Port: $worker_port)"
        
        # Small delay between worker starts
        sleep 1
    done
    
    log "All services started successfully"
}

# Wait for services to be ready
wait_for_ready() {
    log "Waiting for services to be ready..."
    
    local max_wait=30
    local wait_count=0
    
    while [ $wait_count -lt $max_wait ]; do
        if curl -s "http://localhost:$COORDINATOR_PORT/api/status" >/dev/null 2>&1 && \
           curl -s "http://localhost:$CACHE_SERVER_PORT/health" >/dev/null 2>&1 && \
           curl -s "http://localhost:$MONITOR_PORT/health" >/dev/null 2>&1; then
            log "All services are ready"
            return 0
        fi
        
        echo -n "."
        sleep 1
        ((wait_count++))
    done
    
    echo
    warn "Timeout waiting for services to be ready"
    return 1
}

# Check service health
check_health() {
    log "Checking service health..."
    
    # Check coordinator
    if curl -s "http://localhost:$COORDINATOR_PORT/api/status" | jq -r '.worker_count' >/dev/null 2>&1; then
        local worker_count=$(curl -s "http://localhost:$COORDINATOR_PORT/api/status" | jq -r '.worker_count')
        log "✓ Coordinator is running with $worker_count workers"
    else
        error "✗ Coordinator is not responding"
    fi
    
    # Check cache server
    if curl -s "http://localhost:$CACHE_SERVER_PORT/stats" >/dev/null 2>&1; then
        log "✓ Cache server is running"
    else
        error "✗ Cache server is not responding"
    fi
    
    # Check monitor
    if curl -s "http://localhost:$MONITOR_PORT/health" >/dev/null 2>&1; then
        log "✓ Monitor is running"
    else
        error "✗ Monitor is not responding"
    fi
    
    # Check workers
    log "Checking worker registration..."
    sleep 3  # Give workers time to register
    local registered_workers=$(curl -s "http://localhost:$COORDINATOR_PORT/api/workers" | jq 'length' 2>/dev/null || echo "0")
    log "✓ $registered_workers workers registered with coordinator"
}

# Run demo build
run_demo() {
    log "Running demo build..."
    
    # Create a simple demo project if it doesn't exist
    local demo_project="$DATA_DIR/demo-project"
    if [ ! -d "$demo_project" ]; then
        log "Creating demo project..."
        mkdir -p "$demo_project"
        cat > "$demo_project/build.gradle" << 'EOF'
plugins {
    id 'java'
    id 'application'
}

repositories {
    mavenCentral()
}

dependencies {
    implementation 'org.apache.commons:commons-lang3:3.12.0'
    testImplementation 'junit:junit:4.13.2'
}

application {
    mainClass = 'com.example.App'
}
EOF
        
        mkdir -p "$demo_project/src/main/java/com/example"
        cat > "$demo_project/src/main/java/com/example/App.java" << 'EOF'
package com.example;

import org.apache.commons.lang3.StringUtils;

public class App {
    public static void main(String[] args) {
        System.out.println("Hello, Distributed Gradle Build!");
        System.out.println("StringUtils.capitalize(\"hello\") = " + StringUtils.capitalize("hello"));
    }
}
EOF
        
        mkdir -p "$demo_project/src/test/java/com/example"
        cat > "$demo_project/src/test/java/com/example/AppTest.java" << 'EOF'
package com.example;

import org.junit.Test;
import static org.junit.Assert.*;

public class AppTest {
    @Test
    public void testApp() {
        assertEquals("Hello", "Hello");
    }
}
EOF
    fi
    
    # Submit build request
    log "Submitting build request..."
    local build_response=$(curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/build" \
        -H "Content-Type: application/json" \
        -d "{\"project_path\":\"$demo_project\",\"task_name\":\"build\",\"cache_enabled\":true}")
    
    if echo "$build_response" | jq -e '.build_id' >/dev/null 2>&1; then
        local build_id=$(echo "$build_response" | jq -r '.build_id')
        log "✓ Build submitted with ID: $build_id"
        log "Check monitor at: http://localhost:$MONITOR_PORT/api/builds"
    else
        error "✗ Failed to submit build"
        echo "$build_response"
    fi
}

# Show usage information
show_usage() {
    cat << EOF
Usage: $0 [COMMAND]

Commands:
  build     Build all Go components
  start     Start all services
  stop      Stop all services
  restart   Stop and restart all services
  status    Check service health
  demo      Run a demo build
  help      Show this help message

Examples:
  $0 build           # Build all components
  $0 start           # Start all services
  $0 status          # Check service health
  $0 demo            # Run demo build

Service Ports:
  Coordinator HTTP: $COORDINATOR_PORT
  Coordinator RPC:  $COORDINATOR_RPC_PORT
  Cache Server:     $CACHE_SERVER_PORT
  Monitor:          $MONITOR_PORT
  Workers:          $WORKER_BASE_PORT-$((WORKER_BASE_PORT + 2))
EOF
}

# Main execution
main() {
    local command="${1:-help}"
    
    case "$command" in
        "build")
            setup_directories
            build_components
            ;;
        "start")
            setup_directories
            stop_services
            start_services
            wait_for_ready
            check_health
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            setup_directories
            stop_services
            start_services
            wait_for_ready
            check_health
            ;;
        "status")
            check_health
            ;;
        "demo")
            run_demo
            ;;
        "help"|"-h"|"--help")
            show_usage
            ;;
        *)
            error "Unknown command: $command"
            show_usage
            exit 1
            ;;
    esac
}

# Check dependencies
check_dependencies() {
    local missing_deps=()
    
    command -v go >/dev/null 2>&1 || missing_deps+=("go")
    command -v curl >/dev/null 2>&1 || missing_deps+=("curl")
    command -v jq >/dev/null 2>&1 || missing_deps+=("jq")
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        error "Missing required dependencies: ${missing_deps[*]}"
        error "Please install missing dependencies and try again"
        exit 1
    fi
}

# Run main function with all arguments
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    check_dependencies
    main "$@"
fi