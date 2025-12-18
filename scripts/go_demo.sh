#!/bin/bash

# Quick Demo for Distributed Gradle Building System - Go Implementation
# This script sets up and demonstrates the Go-based distributed build system

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
DEMO_DIR="$PROJECT_DIR/demo"

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

# Check dependencies
check_dependencies() {
    log "Checking dependencies..."
    
    local missing_deps=()
    
    command -v go >/dev/null 2>&1 || missing_deps+=("go")
    command -v gradle >/dev/null 2>&1 || missing_deps+=("gradle")
    command -v curl >/dev/null 2>&1 || missing_deps+=("curl")
    command -v jq >/dev/null 2>&1 || missing_deps+=("jq")
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        error "Missing required dependencies: ${missing_deps[*]}"
        error "Please install missing dependencies and try again"
        exit 1
    fi
    
    log "All dependencies available ✓"
}

# Create demo project
create_demo_project() {
    log "Creating demo project..."
    
    mkdir -p "$DEMO_DIR"
    
    # Create build.gradle
    cat > "$DEMO_DIR/build.gradle" << 'EOF'
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
    mainClass = 'com.example.DistributedApp'
}

tasks.withType(JavaCompile) {
    options.incremental = true
}

tasks.withType(Test) {
    maxParallelForks = Runtime.runtime.availableProcessors().intdiv(2) ?: 1
}
EOF

    # Create Java source
    mkdir -p "$DEMO_DIR/src/main/java/com/example"
    cat > "$DEMO_DIR/src/main/java/com/example/DistributedApp.java" << 'EOF'
package com.example;

import org.apache.commons.lang3.StringUtils;
import java.util.concurrent.TimeUnit;
import java.util.ArrayList;
import java.util.List;

public class DistributedApp {
    
    public static void main(String[] args) {
        System.out.println("=== Distributed Gradle Build Demo ===");
        System.out.println("Application starting on distributed build system...");
        
        // Simulate some work to make the build more realistic
        List<String> items = new ArrayList<>();
        for (int i = 0; i < 100; i++) {
            items.add(StringUtils.capitalize("item_" + i));
        }
        
        System.out.println("Processed " + items.size() + " items");
        System.out.println("Build was executed using distributed Gradle system!");
        
        try {
            TimeUnit.SECONDS.sleep(1); // Small delay to simulate work
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
        
        System.out.println("=== Demo Complete ===");
    }
}
EOF

    # Create test
    mkdir -p "$DEMO_DIR/src/test/java/com/example"
    cat > "$DEMO_DIR/src/test/java/com/example/DistributedAppTest.java" << 'EOF'
package com.example;

import org.junit.Test;
import static org.junit.Assert.*;

public class DistributedAppTest {
    
    @Test
    public void testDemo() {
        // Simple test to demonstrate the distributed build system
        assertTrue("Demo test should pass", true);
    }
    
    @Test
    public void testStringUtils() {
        String result = org.apache.commons.lang3.StringUtils.capitalize("hello");
        assertEquals("Hello", result);
    }
}
EOF

    log "Demo project created at $DEMO_DIR ✓"
}

# Build and start Go services
start_services() {
    log "Building and starting distributed build services..."
    
    cd "$PROJECT_DIR"
    
    # Use the build script
    if [ -x "scripts/build_and_deploy.sh" ]; then
        ./scripts/build_and_deploy.sh restart
    else
        error "Build script not found or not executable"
        exit 1
    fi
    
    # Wait for services to be ready
    log "Waiting for services to start..."
    local max_wait=30
    local wait_count=0
    
    while [ $wait_count -lt $max_wait ]; do
        if curl -s "http://localhost:8080/api/status" >/dev/null 2>&1 && \
           curl -s "http://localhost:8083/health" >/dev/null 2>&1 && \
           curl -s "http://localhost:8084/health" >/dev/null 2>&1; then
            log "All services are ready ✓"
            return 0
        fi
        
        echo -n "."
        sleep 1
        ((wait_count++))
    done
    
    echo
    error "Timeout waiting for services to be ready"
    return 1
}

# Run demo build using Go API
run_demo_build() {
    log "Running demo build using Go API..."
    
    # Submit build request
    local build_response=$(curl -s -X POST "http://localhost:8080/api/build" \
        -H "Content-Type: application/json" \
        -d "{\"project_path\":\"$DEMO_DIR\",\"task_name\":\"build\",\"cache_enabled\":true}")
    
    if echo "$build_response" | jq -e '.build_id' >/dev/null 2>&1; then
        local build_id=$(echo "$build_response" | jq -r '.build_id')
        log "Build submitted with ID: $build_id ✓"
        
        # Wait for build completion
        log "Waiting for build to complete..."
        local max_wait=300  # 5 minutes
        local wait_count=0
        
        while [ $wait_count -lt $max_wait ]; do
            local status_response=$(curl -s "http://localhost:8080/api/build/$build_id")
            local status=$(echo "$status_response" | jq -r '.status')
            
            if [ "$status" = "completed" ]; then
                local duration=$(echo "$status_response" | jq -r '.duration')
                local worker_id=$(echo "$status_response" | jq -r '.worker_id')
                local cache_hit_rate=$(echo "$status_response" | jq -r '.cache_hit_rate')
                
                log "Build completed successfully! ✓"
                log "  - Duration: $duration"
                log "  - Worker: $worker_id"
                log "  - Cache hit rate: $(echo "$cache_hit_rate * 100" | bc -l | cut -d. -f1)%"
                
                # Show artifacts
                local artifacts=$(echo "$status_response" | jq -r '.artifacts[]')
                if [ "$artifacts" != "null" ] && [ -n "$artifacts" ]; then
                    log "  - Artifacts:"
                    echo "$artifacts" | while read -r artifact; do
                        if [ -n "$artifact" ]; then
                            echo "    $artifact"
                        fi
                    done
                fi
                
                return 0
            elif [ "$status" = "failed" ]; then
                local error_msg=$(echo "$status_response" | jq -r '.error_message')
                error "Build failed: $error_msg"
                return 1
            fi
            
            echo -n "."
            sleep 3
            ((wait_count++))
        done
        
        echo
        error "Build timed out"
        return 1
    else
        error "Failed to submit build request"
        echo "$build_response"
        return 1
    fi
}

# Show system status
show_status() {
    log "System Status:"
    
    # Coordinator status
    local coord_status=$(curl -s "http://localhost:8080/api/status" 2>/dev/null || echo '{"error":"not responding"}')
    if echo "$coord_status" | jq -e '.worker_count' >/dev/null 2>&1; then
        local worker_count=$(echo "$coord_status" | jq -r '.worker_count')
        local queue_len=$(echo "$coord_status" | jq -r '.queue_length')
        local active_builds=$(echo "$coord_status" | jq -r '.active_builds')
        
        echo "  📋 Coordinator:"
        echo "    - Workers: $worker_count"
        echo "    - Queue: $queue_len builds"
        echo "    - Active: $active_builds builds"
    else
        echo "  ❌ Coordinator: Not responding"
    fi
    
    # Cache server status
    if curl -s "http://localhost:8083/health" >/dev/null 2>&1; then
        local cache_stats=$(curl -s "http://localhost:8083/stats")
        local entries=$(echo "$cache_stats" | jq -r '.entries')
        local hit_rate=$(echo "$cache_stats" | jq -r '.hit_rate * 100' | bc -l | cut -d. -f1)
        
        echo "  💾 Cache Server:"
        echo "    - Entries: $entries"
        echo "    - Hit rate: $hit_rate%"
    else
        echo "  ❌ Cache Server: Not responding"
    fi
    
    # Monitor status
    if curl -s "http://localhost:8084/health" >/dev/null 2>&1; then
        local metrics=$(curl -s "http://localhost:8084/api/metrics")
        local total_builds=$(echo "$metrics" | jq -r '.total_builds')
        local success_rate=$(echo "$metrics" | jq -r '.successful_builds / .total_builds * 100' | bc -l | cut -d. -f1)
        
        echo "  📊 Monitor:"
        echo "    - Total builds: $total_builds"
        echo "    - Success rate: $success_rate%"
    else
        echo "  ❌ Monitor: Not responding"
    fi
}

# Run multiple builds to demonstrate caching
run_multi_build_demo() {
    log "Running multiple builds to demonstrate caching..."
    
    # First build (cache miss)
    log "Build #1 - Expected cache miss..."
    run_single_build "$DEMO_DIR" "clean build"
    
    sleep 2
    
    # Second build (cache hit)
    log "Build #2 - Expected cache hit..."
    run_single_build "$DEMO_DIR" "build"
    
    sleep 2
    
    # Test build
    log "Build #3 - Test execution..."
    run_single_build "$DEMO_DIR" "test"
}

run_single_build() {
    local project_path="$1"
    local task_name="$2"
    
    local build_response=$(curl -s -X POST "http://localhost:8080/api/build" \
        -H "Content-Type: application/json" \
        -d "{\"project_path\":\"$project_path\",\"task_name\":\"$task_name\",\"cache_enabled\":true}")
    
    if echo "$build_response" | jq -e '.build_id' >/dev/null 2>&1; then
        local build_id=$(echo "$build_response" | jq -r '.build_id')
        
        # Wait for completion
        local max_wait=120
        local wait_count=0
        
        while [ $wait_count -lt $max_wait ]; do
            local status_response=$(curl -s "http://localhost:8080/api/build/$build_id")
            local status=$(echo "$status_response" | jq -r '.status')
            
            if [ "$status" = "completed" ]; then
                local duration=$(echo "$status_response" | jq -r '.duration')
                local cache_hit_rate=$(echo "$status_response" | jq -r '.cache_hit_rate')
                log "✓ $task_name completed in $duration (cache hit rate: $(echo "$cache_hit_rate * 100" | bc -l | cut -d. -f1)%)"
                return 0
            elif [ "$status" = "failed" ]; then
                error "✗ $task_name failed"
                return 1
            fi
            
            sleep 1
            ((wait_count++))
        done
        
        error "✗ $task_name timed out"
        return 1
    fi
}

# Clean up
cleanup() {
    log "Cleaning up demo..."
    
    # Stop services
    if [ -x "$PROJECT_DIR/scripts/build_and_deploy.sh" ]; then
        "$PROJECT_DIR/scripts/build_and_deploy.sh" stop >/dev/null 2>&1 || true
    fi
    
    # Remove demo project
    if [ -d "$DEMO_DIR" ]; then
        rm -rf "$DEMO_DIR"
        log "Demo project removed"
    fi
    
    log "Cleanup complete ✓"
}

# Main execution
main() {
    local command="${1:-demo}"
    
    case "$command" in
        "demo")
            log "Starting Distributed Gradle Build Demo..."
            check_dependencies
            create_demo_project
            start_services
            show_status
            run_demo_build
            run_multi_build_demo
            show_status
            log "Demo completed successfully!"
            ;;
        "status")
            show_status
            ;;
        "cleanup")
            cleanup
            ;;
        "help"|"-h"|"--help")
            cat << EOF
Usage: $0 [COMMAND]

Commands:
  demo      Run the complete demo (default)
  status    Show system status
  cleanup   Stop services and remove demo files
  help      Show this help message

Examples:
  $0 demo      # Run full demo
  $0 status    # Check system status
  $0 cleanup   # Clean up after demo
EOF
            ;;
        *)
            error "Unknown command: $command"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Set trap for cleanup on script exit
trap cleanup EXIT

# Run main function with all arguments
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi