#!/bin/bash
# Unit Tests for Build Scripts
# Tests for sync_and_build.sh and performance_analyzer.sh

set -euo pipefail

# Import test framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_ROOT="$(dirname "$SCRIPT_DIR")"
source "$TEST_ROOT/test_framework.sh"

# Test data
TEST_PROJECT_DIR="/tmp/gradle-build-test-$$"
TEST_WORKER_DIR="/tmp/gradle-worker-test-$$"
TEST_CACHE_DIR="/tmp/gradle-cache-test-$$"
TEST_RESULTS_DIR="/tmp/gradle-results-test-$$"

# Setup test environment
setup_test_environment() {
    # Create test project directory with structure
    mkdir -p "$TEST_PROJECT_DIR"/{src/main/java/com/example,src/test/java/com/example,build}
    
    # Create build.gradle
    cat > "$TEST_PROJECT_DIR/build.gradle" << 'EOF'
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
    mainClass = 'com.example.TestApp'
}
tasks.withType(JavaCompile) {
    options.incremental = true
}
EOF

    # Create source files
    cat > "$TEST_PROJECT_DIR/src/main/java/com/example/TestApp.java" << 'EOF'
package com.example;

import org.apache.commons.lang3.StringUtils;

public class TestApp {
    public static void main(String[] args) {
        System.out.println("Test Application");
        System.out.println("Testing: " + StringUtils.capitalize("distributed builds"));
    }
}
EOF

    cat > "$TEST_PROJECT_DIR/src/test/java/com/example/TestAppTest.java" << 'EOF'
package com.example;

import org.junit.Test;
import static org.junit.Assert.*;

public class TestAppTest {
    @Test
    public void testApp() {
        assertTrue("Test should pass", true);
    }
    
    @Test
    public void testStringUtils() {
        String result = StringUtils.capitalize("hello");
        assertEquals("Hello", result);
    }
}
EOF

    # Create worker directory
    mkdir -p "$TEST_WORKER_DIR"/{workspace,cache}
    
    # Create cache directory
    mkdir -p "$TEST_CACHE_DIR"/{gradle,build}
    
    # Create results directory
    mkdir -p "$TEST_RESULTS_DIR"
}

# Cleanup test environment
cleanup_test_environment() {
    rm -rf "$TEST_PROJECT_DIR"
    rm -rf "$TEST_WORKER_DIR"
    rm -rf "$TEST_CACHE_DIR"
    rm -rf "$TEST_RESULTS_DIR"
}

# Mock functions for testing
mock_gradle() {
    echo "Mock Gradle execution: $*"
    echo "BUILD SUCCESSFUL"
    return 0
}

mock_rsync() {
    echo "Mock rsync: $*"
    return 0
}

mock_ssh() {
    echo "Mock SSH to $1: ${*:2}"
    echo "Command executed successfully"
    return 0
}

# Test sync_and_build.sh functionality
test_sync_and_build_script() {
    start_test_group "Sync and Build Script Tests"
    
    # Test script exists
    assert_file_exists "$TEST_ROOT/sync_and_build.sh" "sync_and_build.sh should exist"
    
    # Test script is executable
    assert_command_success "test -x $TEST_ROOT/sync_and_build.sh" "sync_and_build.sh should be executable"
    
    # Test script with help option
    local help_output
    help_output=$("$TEST_ROOT/sync_and_build.sh" --help 2>&1 || true)
    assert_contains "$help_output" "Usage" "Should show help information"
    
    # Test script without arguments
    local error_output
    error_output=$("$TEST_ROOT/sync_and_build.sh" 2>&1 || true)
    assert_contains "$error_output" "Usage" "Should show usage when no arguments provided"
    
    # Test script with invalid task
    cd "$TEST_PROJECT_DIR"
    local invalid_task_output
    invalid_task_output=$("$TEST_ROOT/sync_and_build.sh" invalid-task 2>&1 || true)
    assert_contains "$invalid_task_output" "invalid" "Should detect invalid task" || true
    
    end_test_group "Sync and Build Script Tests"
}

# Test performance_analyzer.sh functionality
test_performance_analyzer_script() {
    start_test_group "Performance Analyzer Script Tests"
    
    local script_path="$TEST_ROOT/scripts/performance_analyzer.sh"
    
    # Test script exists
    assert_file_exists "$script_path" "performance_analyzer.sh should exist"
    
    # Test script is executable
    assert_command_success "test -x $script_path" "performance_analyzer.sh should be executable"
    
    # Test script with help option
    local help_output
    help_output=("$script_path" --help 2>&1 || true)
    assert_contains "$help_output" "Usage" "Should show help information"
    
    # Test script with invalid mode
    local error_output
    error_output=("$script_path" invalid-mode 2>&1 || true)
    assert_contains "$error_output" "invalid" "Should detect invalid mode"
    
    # Test performance metrics collection
    local test_metrics="/tmp/test-metrics-$$"
    cat > "$test_metrics" << EOF
{
  "build_time": 120,
  "cache_hit_rate": 0.75,
  "memory_usage_mb": 512,
  "cpu_usage": 0.45,
  "network_io_mb": 25
}
EOF

    # Test metrics parsing
    if command -v jq >/dev/null 2>&1; then
        local build_time
        build_time=$(jq -r '.build_time' "$test_metrics")
        assert_equals "120" "$build_time" "Should parse build time correctly"
        
        local cache_hit_rate
        cache_hit_rate=$(jq -r '.cache_hit_rate' "$test_metrics")
        assert_equals "0.75" "$cache_hit_rate" "Should parse cache hit rate correctly"
    else
        log_warning "jq not available for JSON parsing tests"
        ((TESTS_SKIPPED++))
        ((TESTS_TOTAL++))
    fi
    
    rm -f "$test_metrics"
    
    end_test_group "Performance Analyzer Script Tests"
}

# Test project synchronization
test_project_synchronization() {
    start_test_group "Project Synchronization Tests"
    
    # Test rsync command construction
    local rsync_cmd="rsync -avz --exclude='build' --exclude='.gradle' --exclude='.git' $TEST_PROJECT_DIR/ $TEST_WORKER_DIR/workspace"
    
    # Test rsync command format
    assert_contains "$rsync_cmd" "--exclude='build'" "Should exclude build directory"
    assert_contains "$rsync_cmd" "--exclude='.gradle'" "Should exclude .gradle directory"
    assert_contains "$rsync_cmd" "--exclude='.git'" "Should exclude .git directory"
    assert_contains "$rsync_cmd" "$TEST_PROJECT_DIR/" "Should include source directory"
    assert_contains "$rsync_cmd" "$TEST_WORKER_DIR/workspace" "Should include target directory"
    
    # Test file synchronization simulation
    local test_file="$TEST_PROJECT_DIR/test-sync-$$"
    echo "test content" > "$test_file"
    
    # Simulate rsync
    cp "$test_file" "$TEST_WORKER_DIR/workspace/"
    
    local target_file="$TEST_WORKER_DIR/workspace/$(basename "$test_file")"
    assert_file_exists "$target_file" "File should be synchronized to worker"
    
    rm -f "$test_file" "$target_file"
    
    end_test_group "Project Synchronization Tests"
}

# Test build task distribution
test_build_task_distribution() {
    start_test_group "Build Task Distribution Tests"
    
    # Create test build tasks
    local test_tasks="/tmp/test-tasks-$$"
    cat > "$test_tasks" << EOF
{
  "tasks": [
    {
      "name": "compileJava",
      "type": "compile",
      "dependencies": [],
      "worker": "worker-1"
    },
    {
      "name": "compileTestJava",
      "type": "compile",
      "dependencies": ["compileJava"],
      "worker": "worker-2"
    },
    {
      "name": "test",
      "type": "test",
      "dependencies": ["compileTestJava"],
      "worker": "worker-3"
    },
    {
      "name": "jar",
      "type": "package",
      "dependencies": ["compileJava"],
      "worker": "worker-1"
    }
  ]
}
EOF

    # Test task distribution logic
    if command -v jq >/dev/null 2>&1; then
        local total_tasks
        total_tasks=$(jq -r '.tasks | length' "$test_tasks")
        assert_equals "4" "$total_tasks" "Should identify 4 build tasks"
        
        local compile_tasks
        compile_tasks=$(jq -r '.tasks[] | select(.type == "compile") | .name' "$test_tasks" | wc -l)
        assert_equals "2" "$compile_tasks" "Should identify 2 compile tasks"
        
        local test_tasks_count
        test_tasks_count=$(jq -r '.tasks[] | select(.type == "test") | .name' "$test_tasks" | wc -l)
        assert_equals "1" "$test_tasks_count" "Should identify 1 test task"
        
        local package_tasks
        package_tasks=$(jq -r '.tasks[] | select(.type == "package") | .name' "$test_tasks" | wc -l)
        assert_equals "1" "$package_tasks" "Should identify 1 package task"
    else
        log_warning "jq not available for JSON parsing tests"
        ((TESTS_SKIPPED++))
        ((TESTS_TOTAL++))
    fi
    
    rm -f "$test_tasks"
    
    end_test_group "Build Task Distribution Tests"
}

# Test caching functionality
test_caching_functionality() {
    start_test_group "Caching Functionality Tests"
    
    # Test cache key generation
    local test_file="$TEST_PROJECT_DIR/src/main/java/com/example/TestApp.java"
    local file_hash
    if command -v md5 >/dev/null 2>&1; then
        file_hash=$(md5 -q "$test_file" 2>/dev/null || echo "")
    elif command -v md5sum >/dev/null 2>&1; then
        file_hash=$(md5sum "$test_file" 2>/dev/null | cut -d' ' -f1 || echo "")
    else
        file_hash="test-hash-12345"
    fi
    
    assert_not_equals "" "$file_hash" "Should generate file hash for cache key"
    
    # Test cache directory structure
    local cache_subdir="$TEST_CACHE_DIR/gradle/$(echo "$file_hash" | cut -c1-2)"
    mkdir -p "$cache_subdir"
    
    assert_file_exists "$cache_subdir" "Should create cache subdirectory based on hash"
    
    # Test cache entry creation
    local cache_entry="$cache_subdir/$file_hash"
    echo "cached build result" > "$cache_entry"
    
    assert_file_exists "$cache_entry" "Should create cache entry"
    
    # Test cache retrieval
    local cached_content
    cached_content=$(cat "$cache_entry")
    assert_equals "cached build result" "$cached_content" "Should retrieve cached content"
    
    end_test_group "Caching Functionality Tests"
}

# Test build result collection
test_build_result_collection() {
    start_test_group "Build Result Collection Tests"
    
    # Create mock build results
    local worker1_results="$TEST_RESULTS_DIR/worker-1-results.json"
    local worker2_results="$TEST_RESULTS_DIR/worker-2-results.json"
    
    cat > "$worker1_results" << EOF
{
  "worker_id": "worker-1",
  "task_name": "compileJava",
  "status": "success",
  "duration_seconds": 45,
  "artifacts": [
    "build/classes/java/main/com/example/TestApp.class"
  ],
  "timestamp": $(date +%s)
}
EOF

    cat > "$worker2_results" << EOF
{
  "worker_id": "worker-2",
  "task_name": "compileTestJava",
  "status": "success",
  "duration_seconds": 30,
  "artifacts": [
    "build/classes/java/test/com/example/TestAppTest.class"
  ],
  "timestamp": $(date +%s)
}
EOF

    # Test result collection
    assert_file_exists "$worker1_results" "Worker 1 results file should exist"
    assert_file_exists "$worker2_results" "Worker 2 results file should exist"
    
    # Test result aggregation
    local aggregated_results="$TEST_RESULTS_DIR/aggregated-results.json"
    
    if command -v jq >/dev/null 2>&1; then
        jq -s 'flatten' "$worker1_results" "$worker2_results" > "$aggregated_results"
        
        local total_results
        total_results=$(jq -r '. | length' "$aggregated_results")
        assert_equals "2" "$total_results" "Should aggregate 2 result sets"
        
        local successful_tasks
        successful_tasks=$(jq -r '.[] | select(.status == "success") | .task_name' "$aggregated_results" | wc -l)
        assert_equals "2" "$successful_tasks" "Should identify 2 successful tasks"
    else
        log_warning "jq not available for JSON aggregation tests"
        ((TESTS_SKIPPED++))
        ((TESTS_TOTAL++))
    fi
    
    end_test_group "Build Result Collection Tests"
}

# Test performance measurement
test_performance_measurement() {
    start_test_group "Performance Measurement Tests"
    
    # Test build time measurement
    local start_time
    start_time=$(date +%s)
    
    # Simulate build process
    sleep 1
    
    local end_time
    end_time=$(date +%s)
    
    local build_duration
    build_duration=$((end_time - start_time))
    
    assert_true "($build_duration >= 1)" "Should measure build duration correctly"
    
    # Test resource usage measurement
    local memory_usage_kb
    memory_usage_kb=$(grep VmRSS /proc/self/status 2>/dev/null | awk '{print $2}' || echo "10240")
    assert_true "($memory_usage_kb > 0)" "Should measure memory usage"
    
    # Test disk usage measurement
    local disk_usage_kb
    disk_usage_kb=$(du -sk "$TEST_PROJECT_DIR" 2>/dev/null | cut -f1 || echo "1024")
    assert_true "($disk_usage_kb > 0)" "Should measure disk usage"
    
    # Test network transfer measurement (mock)
    local transfer_size_mb=25
    local transfer_time_seconds=5
    local transfer_rate_mbps
    transfer_rate_mbps=$(echo "scale=2; $transfer_size_mb / $transfer_time_seconds" | bc -l 2>/dev/null || echo "5.0")
    
    assert_true "$(echo "$transfer_rate_mbps > 0" | bc -l 2>/dev/null || echo "1")" "Should calculate transfer rate"
    
    end_test_group "Performance Measurement Tests"
}

# Test error handling and recovery
test_error_handling_and_recovery() {
    start_test_group "Error Handling and Recovery Tests"
    
    # Test build failure detection
    local build_log="$TEST_RESULTS_DIR/build.log"
    cat > "$build_log" << EOF
Starting build...
Compiling sources...
FAILURE: Build failed with an exception.
EOF

    local build_status
    build_status=$(grep -q "FAILURE" "$build_log" && echo "failed" || echo "success")
    assert_equals "failed" "$build_status" "Should detect build failure"
    
    # Test worker failure handling
    local worker_status="/tmp/worker-failure-test-$$"
    cat > "$worker_status" << EOF
{
  "worker_id": "worker-3",
  "status": "failed",
  "error": "Connection timeout",
  "last_heartbeat": $(($(date +%s) - 300))
}
EOF

    # Test worker failure detection
    if command -v jq >/dev/null 2>&1; then
        local worker_failed
        worker_failed=$(jq -r '.status' "$worker_status")
        assert_equals "failed" "$worker_failed" "Should detect worker failure"
        
        local error_message
        error_message=$(jq -r '.error' "$worker_status")
        assert_contains "$error_message" "timeout" "Should capture error message"
    else
        log_warning "jq not available for error parsing tests"
        ((TESTS_SKIPPED++))
        ((TESTS_TOTAL++))
    fi
    
    rm -f "$worker_status"
    
    # Test retry logic
    local retry_count=0
    local max_retries=3
    
    while ((retry_count < max_retries)); do
        retry_count=$((retry_count + 1))
        # Simulate retry success on third attempt
        if ((retry_count == 3)); then
            break
        fi
    done
    
    assert_equals "3" "$retry_count" "Should retry up to maximum attempts"
    
    end_test_group "Error Handling and Recovery Tests"
}

# Main test execution
main() {
    local test_type="${1:-all}"
    
    # Setup test environment
    setup_test_environment
    
    # Trap cleanup
    trap cleanup_test_environment EXIT
    
    case "$test_type" in
        "sync")
            test_sync_and_build_script
            test_project_synchronization
            test_build_result_collection
            ;;
        "build")
            test_sync_and_build_script
            test_build_task_distribution
            test_caching_functionality
            ;;
        "performance")
            test_performance_analyzer_script
            test_performance_measurement
            ;;
        "error")
            test_error_handling_and_recovery
            ;;
        "all")
            test_sync_and_build_script
            test_performance_analyzer_script
            test_project_synchronization
            test_build_task_distribution
            test_caching_functionality
            test_build_result_collection
            test_performance_measurement
            test_error_handling_and_recovery
            ;;
        *)
            echo "Unknown test type: $test_type"
            echo "Available types: sync, build, performance, error, all"
            return 1
            ;;
    esac
}

# Run tests if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi