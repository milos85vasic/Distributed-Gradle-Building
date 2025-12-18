#!/bin/bash

# Distributed Gradle Building - Concurrent Builds Test
# Tests system behavior under concurrent build requests

set -euo pipefail

# Import test framework
source "$(dirname "$0")/../test_framework.sh"

# Configuration
TEST_PROJECT_NAME="concurrent-test"
CONCURRENT_BUILDS=5
BUILD_TIMEOUT=300
COORDINATOR_PORT=${COORDINATOR_PORT:-8082}
RESULTS_DIR="${TEST_RESULTS_DIR}/concurrent_builds"

# Test data setup
setup_concurrent_build_test() {
    echo "Setting up concurrent builds test environment..."
    
    # Create test results directory
    mkdir -p "$RESULTS_DIR"
    
    # Create multiple test projects
    for i in $(seq 1 $CONCURRENT_BUILDS); do
        project_dir="${TEST_WORKSPACE}/${TEST_PROJECT_NAME}-${i}"
        mkdir -p "$project_dir/src/main/java"
        
        # Create unique Java source for each project
        cat > "$project_dir/src/main/java/Example${i}.java" << EOF
public class Example${i} {
    public static void main(String[] args) {
        System.out.println("Hello from concurrent build ${i}!");
        
        // Simulate some processing time
        for (int i = 0; i < 1000; i++) {
            Math.sqrt(i);
        }
        
        System.out.println("Concurrent build ${i} completed!");
    }
}
EOF
        
        # Create build.gradle
        cat > "$project_dir/build.gradle" << EOF
plugins {
    application
    java
}

application {
    mainClass = "Example${i}"
}

java {
    sourceCompatibility = JavaVersion.VERSION_11
    targetCompatibility = JavaVersion.VERSION_11
}
EOF
        
        echo "Created test project: $project_dir"
    done
    
    test_pass "Concurrent builds test environment setup completed"
}

# Start coordinator for testing
start_test_coordinator() {
    echo "Starting test coordinator..."
    
    cd "$TEST_WORKSPACE"
    export COORDINATOR_PORT="$COORDINATOR_PORT"
    export LOG_LEVEL="INFO"
    
    # Use Go coordinator
    if command -v "${PROJECT_DIR}/go/bin/coordinator" >/dev/null 2>&1; then
        "${PROJECT_DIR}/go/bin/coordinator" > "${RESULTS_DIR}/coordinator.log" 2>&1 &
        COORDINATOR_PID=$!
        sleep 3
        test_pass "Test coordinator started (PID: $COORDINATOR_PID)"
    else
        test_fail "Coordinator binary not found"
        return 1
    fi
}

# Submit concurrent builds
submit_concurrent_builds() {
    echo "Submitting $CONCURRENT_BUILDS concurrent builds..."
    
    local build_ids=()
    local submission_times=()
    
    # Submit all builds concurrently
    for i in $(seq 1 $CONCURRENT_BUILDS); do
        {
            project_dir="${TEST_WORKSPACE}/${TEST_PROJECT_NAME}-${i}"
            start_time=$(date +%s%N)
            
            # Submit build via API
            build_response=$(curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
                -H "Content-Type: application/json" \
                -d "{
                    \"project_path\": \"$project_dir\",
                    \"task_name\": \"build\",
                    \"cache_enabled\": true
                }")
            
            submission_time=$(date +%s%N)
            submission_duration=$((($submission_time - $start_time) / 1000000)) # Convert to milliseconds
            
            build_id=$(echo "$build_response" | jq -r '.build_id // empty')
            
            if [ -n "$build_id" ]; then
                echo "$build_id:$submission_duration" >> "${RESULTS_DIR}/build_submissions.log"
                test_pass "Build $i submitted successfully (ID: $build_id, ${submission_duration}ms)"
            else
                test_fail "Build $i submission failed: $build_response"
            fi
        } &
    done
    
    # Wait for all submissions to complete
    wait
    
    test_pass "All $CONCURRENT_BUILDS build submissions completed"
}

# Monitor concurrent builds
monitor_concurrent_builds() {
    echo "Monitoring concurrent build execution..."
    
    local active_builds=0
    local completed_builds=0
    local failed_builds=0
    local start_time=$(date +%s)
    local max_build_time=0
    local min_build_time=999999
    local total_build_time=0
    
    # Monitor builds until all complete or timeout
    while true; do
        # Get current build status
        status_response=$(curl -s "http://localhost:$COORDINATOR_PORT/api/builds/status")
        active_builds=$(echo "$status_response" | jq -r '.active_builds // 0')
        completed_builds=$(echo "$status_response" | jq -r '.completed_builds // 0')
        failed_builds=$(echo "$status_response" | jq -r '.failed_builds // 0')
        
        current_time=$(date +%s)
        elapsed_time=$((current_time - start_time))
        
        echo "[$elapsed_time(s)] Active: $active_builds, Completed: $completed_builds, Failed: $failed_builds"
        
        # Check if all builds are complete
        if [ "$active_builds" -eq 0 ] && [ $(("$completed_builds" + "$failed_builds")) -eq $CONCURRENT_BUILDS ]; then
            test_pass "All builds completed (Success: $completed_builds, Failed: $failed_builds)"
            break
        fi
        
        # Check for timeout
        if [ "$elapsed_time" -gt "$BUILD_TIMEOUT" ]; then
            test_fail "Build timeout after ${BUILD_TIMEOUT}s"
            break
        fi
        
        sleep 2
    done
    
    # Collect detailed build results
    for i in $(seq 1 $CONCURRENT_BUILDS); do
        build_details=$(curl -s "http://localhost:$COORDINATOR_PORT/api/builds/${TEST_PROJECT_NAME}-${i}/details")
        if [ -n "$build_details" ]; then
            build_time=$(echo "$build_details" | jq -r '.duration_ms // 0')
            build_status=$(echo "$build_details" | jq -r '.status // "unknown"')
            
            if [ "$build_time" -gt "$max_build_time" ]; then
                max_build_time=$build_time
            fi
            
            if [ "$build_time" -lt "$min_build_time" ] && [ "$build_time" -gt 0 ]; then
                min_build_time=$build_time
            fi
            
            if [ "$build_status" = "success" ]; then
                total_build_time=$((total_build_time + build_time))
            fi
            
            echo "Build ${i}: Status=$build_status, Duration=${build_time}ms" >> "${RESULTS_DIR}/build_details.log"
        fi
    done
    
    # Calculate performance metrics
    avg_build_time=0
    if [ "$completed_builds" -gt 0 ]; then
        avg_build_time=$((total_build_time / completed_builds))
    fi
    
    cat > "${RESULTS_DIR}/performance_metrics.json" << EOF
{
    "concurrent_builds": $CONCURRENT_BUILDS,
    "completed_builds": $completed_builds,
    "failed_builds": $failed_builds,
    "max_build_time_ms": $max_build_time,
    "min_build_time_ms": $min_build_time,
    "avg_build_time_ms": $avg_build_time,
    "success_rate": $(echo "scale=2; $completed_builds * 100 / $CONCURRENT_BUILDS" | bc -l)
}
EOF
    
    test_pass "Performance metrics collected"
}

# Test resource utilization
test_resource_utilization() {
    echo "Testing resource utilization during concurrent builds..."
    
    # Monitor system resources
    (
        while true; do
            timestamp=$(date +%s)
            cpu_usage=$(top -l 1 -n 0 | grep "CPU usage" | awk '{print $3}' | sed 's/%//')
            memory_pressure=$(memory_pressure | grep "System-wide memory free percentage" | awk '{print $5}' | sed 's/%//')
            
            echo "$timestamp,$cpu_usage,$memory_pressure" >> "${RESULTS_DIR}/resource_utilization.csv"
            sleep 1
        done
    ) &
    RESOURCE_MONITOR_PID=$!
    
    # Run a quick concurrent build test
    submit_concurrent_builds
    monitor_concurrent_builds
    
    # Stop resource monitoring
    kill $RESOURCE_MONITOR_PID 2>/dev/null || true
    
    # Analyze resource utilization
    if [ -f "${RESULTS_DIR}/resource_utilization.csv" ]; then
        max_cpu=$(sort -t',' -k2 -nr "${RESULTS_DIR}/resource_utilization.csv" | head -1 | cut -d',' -f2)
        avg_cpu=$(awk -F',' '{sum+=$2; count++} END {print sum/count}' "${RESULTS_DIR}/resource_utilization.csv")
        
        test_pass "Resource utilization monitored (Max CPU: ${max_cpu}%, Avg CPU: ${avg_cpu}%)"
    fi
}

# Cleanup test environment
cleanup_concurrent_build_test() {
    echo "Cleaning up concurrent builds test environment..."
    
    # Stop coordinator
    if [ -n "${COORDINATOR_PID:-}" ]; then
        kill $COORDINATOR_PID 2>/dev/null || true
        wait $COORDINATOR_PID 2>/dev/null || true
    fi
    
    # Stop resource monitor
    if [ -n "${RESOURCE_MONITOR_PID:-}" ]; then
        kill $RESOURCE_MONITOR_PID 2>/dev/null || true
    fi
    
    # Clean up test projects
    rm -rf "${TEST_WORKSPACE}/${TEST_PROJECT_NAME}"*
    
    test_pass "Concurrent builds test cleanup completed"
}

# Main test execution
main() {
    echo "Starting concurrent builds test..."
    
    setup_concurrent_build_test || exit 1
    start_test_coordinator || exit 1
    submit_concurrent_builds || exit 1
    monitor_concurrent_builds || exit 1
    test_resource_utilization || exit 1
    cleanup_concurrent_build_test || exit 1
    
    echo "Concurrent builds test completed successfully!"
}

# Run tests if script is executed directly
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi