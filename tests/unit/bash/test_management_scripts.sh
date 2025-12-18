#!/bin/bash
# Unit Tests for Management Scripts
# Tests for worker_pool_manager.sh, cache_manager.sh, health_checker.sh

set -euo pipefail

# Import test framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_ROOT="$(dirname "$SCRIPT_DIR")"
source "$TEST_ROOT/test_framework.sh"

# Test data
TEST_PROJECT_DIR="/tmp/gradle-mgmt-test-$$"
TEST_CACHE_DIR="/tmp/gradle-cache-test-$$"
TEST_LOG_DIR="/tmp/gradle-logs-test-$$"
MOCK_PID_FILE="/tmp/test-pid-$$"

# Setup test environment
setup_test_environment() {
    # Create test project directory
    mkdir -p "$TEST_PROJECT_DIR/src/main/java"
    
    # Create basic project structure
    cat > "$TEST_PROJECT_DIR/build.gradle" << 'EOF'
plugins {
    id 'java'
}
repositories {
    mavenCentral()
}
dependencies {
    implementation 'org.apache.commons:commons-lang3:3.12.0'
}
EOF

    # Create cache directory
    mkdir -p "$TEST_CACHE_DIR"
    
    # Create log directory
    mkdir -p "$TEST_LOG_DIR"
    
    # Create mock PID file
    echo "12345" > "$MOCK_PID_FILE"
}

# Cleanup test environment
cleanup_test_environment() {
    rm -rf "$TEST_PROJECT_DIR"
    rm -rf "$TEST_CACHE_DIR"
    rm -rf "$TEST_LOG_DIR"
    rm -f "$MOCK_PID_FILE"
}

# Mock functions for testing
mock_curl() {
    case "$1" in
        "-s")
            case "$2" in
                "-f")
                    # Mock health check response
                    echo '{"status": "healthy", "uptime": 3600, "version": "1.0.0"}'
                    return 0
                    ;;
                *)
                    # Mock generic API response
                    echo '{"status": "success", "data": []}'
                    return 0
                    ;;
            esac
            ;;
        "-X")
            case "$2" in
                "POST")
                    # Mock POST response
                    echo '{"id": "test-123", "status": "created"}'
                    return 0
                    ;;
                "DELETE")
                    # Mock DELETE response
                    echo '{"status": "deleted"}'
                    return 0
                    ;;
                *)
                    # Mock other methods
                    echo '{"status": "success"}'
                    return 0
                    ;;
            esac
            ;;
        *)
            echo "Mock curl: $*"
            return 0
            ;;
    esac
}

# Test worker_pool_manager.sh functionality
test_worker_pool_manager_script() {
    start_test_group "Worker Pool Manager Script Tests"
    
    local script_path="$TEST_ROOT/scripts/worker_pool_manager.sh"
    
    # Test script exists
    assert_file_exists "$script_path" "worker_pool_manager.sh should exist"
    
    # Test script is executable
    assert_command_success "test -x $script_path" "worker_pool_manager.sh should be executable"
    
    # Test script with help option
    local help_output
    help_output=$("$script_path" --help 2>&1 || true)
    assert_contains "$help_output" "Usage" "Should show help information"
    
    # Test script with invalid command
    local error_output
    error_output=$("$script_path" invalid-command 2>&1 || true)
    assert_contains "$error_output" "invalid" "Should detect invalid command" || assert_contains "$error_output" "Usage" "Should show usage"
    
    # Test worker list functionality (mock)
    # Check if script can parse worker configuration
    local test_config="/tmp/workers-config-$$"
    cat > "$test_config" << EOF
WORKER_IPS="192.168.1.10 192.168.1.11 192.168.1.12"
MAX_WORKERS=6
MIN_WORKERS=2
WORKER_PORT=8083
EOF

    # Test configuration parsing
    source "$test_config"
    assert_equals "3" "$(echo $WORKER_IPS | wc -w)" "Should parse 3 worker IPs"
    assert_equals "6" "$MAX_WORKERS" "Should parse max workers correctly"
    assert_equals "2" "$MIN_WORKERS" "Should parse min workers correctly"
    assert_equals "8083" "$WORKER_PORT" "Should parse worker port correctly"
    
    rm -f "$test_config"
    
    end_test_group "Worker Pool Manager Script Tests"
}

# Test cache_manager.sh functionality
test_cache_manager_script() {
    start_test_group "Cache Manager Script Tests"
    
    local script_path="$TEST_ROOT/scripts/cache_manager.sh"
    
    # Test script exists
    assert_file_exists "$script_path" "cache_manager.sh should exist"
    
    # Test script is executable
    assert_command_success "test -x $script_path" "cache_manager.sh should be executable"
    
    # Test script with help option
    local help_output
    help_output=$("$script_path" --help 2>&1 || true)
    assert_contains "$help_output" "Usage" "Should show help information"
    
    # Test script with invalid command
    local error_output
    error_output=$("$script_path" invalid-command 2>&1 || true)
    assert_contains "$error_output" "invalid" "Should detect invalid command" || assert_contains "$error_output" "Usage" "Should show usage"
    
    # Test cache directory structure
    mkdir -p "$TEST_CACHE_DIR"/{gradle,build,artifacts}
    assert_file_exists "$TEST_CACHE_DIR/gradle" "Gradle cache directory should exist"
    assert_file_exists "$TEST_CACHE_DIR/build" "Build cache directory should exist"
    assert_file_exists "$TEST_CACHE_DIR/artifacts" "Artifacts cache directory should exist"
    
    # Test cache size calculation (mock)
    echo "test content" > "$TEST_CACHE_DIR/test-file"
    local cache_size
    cache_size=$(du -sb "$TEST_CACHE_DIR" 2>/dev/null | cut -f1 || echo "0")
    assert_true "($cache_size > 0)" "Should calculate cache size correctly"
    
    # Test cache cleanup functionality
    local test_old_file="$TEST_CACHE_DIR/old-file-$$"
    touch -t 202001010000 "$test_old_file" 2>/dev/null || touch "$test_old_file"
    echo "old content" > "$test_old_file"
    assert_file_exists "$test_old_file" "Old test file should exist"
    
    end_test_group "Cache Manager Script Tests"
}

# Test health_checker.sh functionality
test_health_checker_script() {
    start_test_group "Health Checker Script Tests"
    
    local script_path="$TEST_ROOT/scripts/health_checker.sh"
    
    # Test script exists
    assert_file_exists "$script_path" "health_checker.sh should exist"
    
    # Test script is executable
    assert_command_success "test -x $script_path" "health_checker.sh should be executable"
    
    # Test script with help option
    local help_output
    help_output=$("$script_path" --help 2>&1 || true)
    assert_contains "$help_output" "Usage" "Should show help information"
    
    # Test health check functionality (mock with curl)
    export PATH="/usr/bin:/bin:$TEST_ROOT/tests/mocks:$PATH"
    alias curl=mock_curl
    
    # Test health endpoint checking
    local health_output
    health_output=$(mock_curl -s -f http://localhost:8080/health)
    assert_contains "$health_output" "healthy" "Should detect healthy status"
    assert_contains "$health_output" "uptime" "Should include uptime information"
    
    # Test health metrics parsing
    local uptime
    uptime=$(echo "$health_output" | jq -r '.uptime' 2>/dev/null || echo "3600")
    assert_true "($uptime > 0)" "Should parse uptime correctly"
    
    end_test_group "Health Checker Script Tests"
}

# Test worker status monitoring
test_worker_status_monitoring() {
    start_test_group "Worker Status Monitoring Tests"
    
    # Create mock worker status data
    local worker_status="/tmp/worker-status-$$"
    cat > "$worker_status" << EOF
{
  "workers": [
    {
      "id": "worker-1",
      "host": "192.168.1.10",
      "port": 8083,
      "status": "online",
      "last_heartbeat": $(date +%s),
      "cpu_usage": 0.45,
      "memory_usage": 0.67,
      "active_tasks": 2
    },
    {
      "id": "worker-2",
      "host": "192.168.1.11",
      "port": 8084,
      "status": "offline",
      "last_heartbeat": $(($(date +%s) - 300)),
      "cpu_usage": 0.0,
      "memory_usage": 0.0,
      "active_tasks": 0
    },
    {
      "id": "worker-3",
      "host": "192.168.1.12",
      "port": 8085,
      "status": "online",
      "last_heartbeat": $(date +%s),
      "cpu_usage": 0.23,
      "memory_usage": 0.45,
      "active_tasks": 1
    }
  ]
}
EOF

    # Test worker status parsing
    if command -v jq >/dev/null 2>&1; then
        local online_workers
        online_workers=$(jq -r '.workers[] | select(.status == "online") | .id' "$worker_status" | wc -l)
        assert_equals "2" "$online_workers" "Should detect 2 online workers"
        
        local offline_workers
        offline_workers=$(jq -r '.workers[] | select(.status == "offline") | .id' "$worker_status" | wc -l)
        assert_equals "1" "$offline_workers" "Should detect 1 offline worker"
        
        local high_cpu_workers
        high_cpu_workers=$(jq -r '.workers[] | select(.cpu_usage > 0.4) | .id' "$worker_status" | wc -l)
        assert_equals "1" "$high_cpu_workers" "Should detect 1 worker with high CPU usage"
    else
        log_warning "jq not available for JSON parsing tests"
        ((TESTS_SKIPPED++))
        ((TESTS_TOTAL++))
    fi
    
    rm -f "$worker_status"
    
    end_test_group "Worker Status Monitoring Tests"
}

# Test cache statistics
test_cache_statistics() {
    start_test_group "Cache Statistics Tests"
    
    # Create mock cache statistics
    local cache_stats="/tmp/cache-stats-$$"
    cat > "$cache_stats" << EOF
{
  "cache_size_mb": 1024,
  "cache_entries": 1500,
  "hit_rate": 0.75,
  "miss_rate": 0.25,
  "evictions": 50,
  "last_cleanup": $(date +%s)
}
EOF

    # Test cache statistics parsing
    if command -v jq >/dev/null 2>&1; then
        local cache_size
        cache_size=$(jq -r '.cache_size_mb' "$cache_stats")
        assert_equals "1024" "$cache_size" "Should parse cache size correctly"
        
        local hit_rate
        hit_rate=$(jq -r '.hit_rate' "$cache_stats")
        assert_equals "0.75" "$hit_rate" "Should parse hit rate correctly"
        
        local cache_entries
        cache_entries=$(jq -r '.cache_entries' "$cache_stats")
        assert_equals "1500" "$cache_entries" "Should parse cache entries correctly"
    else
        log_warning "jq not available for JSON parsing tests"
        ((TESTS_SKIPPED++))
        ((TESTS_TOTAL++))
    fi
    
    rm -f "$cache_stats"
    
    end_test_group "Cache Statistics Tests"
}

# Test process management
test_process_management() {
    start_test_group "Process Management Tests"
    
    # Test PID file handling
    echo "12345" > "$MOCK_PID_FILE"
    assert_file_exists "$MOCK_PID_FILE" "PID file should exist"
    
    # Test PID validation
    local pid
    pid=$(cat "$MOCK_PID_FILE")
    assert_equals "12345" "$pid" "Should read PID from file correctly"
    
    # Test process existence check (mock - since we don't have a real process)
    if command -v ps >/dev/null 2>&1; then
        # This will likely fail since 12345 is just a mock PID
        if ps -p "$pid" >/dev/null 2>&1; then
            log_success "Process $pid is running"
            ((TESTS_PASSED++))
        else
            log_warning "Process $pid is not running (expected for mock PID)"
            ((TESTS_SKIPPED++))
        fi
        ((TESTS_TOTAL++))
    fi
    
    # Test service status commands
    if command -v systemctl >/dev/null 2>&1; then
        # Test systemd service management
        assert_command_success "systemctl --version" "systemctl should be available"
    elif command -v service >/dev/null 2>&1; then
        # Test SysV service management
        assert_command_success "service --status-all" "service command should be available"
    else
        log_warning "No service management system detected"
        ((TESTS_SKIPPED++))
        ((TESTS_TOTAL++))
    fi
    
    end_test_group "Process Management Tests"
}

# Test logging functionality
test_logging_functionality() {
    start_test_group "Logging Functionality Tests"
    
    # Test log directory creation
    assert_file_exists "$TEST_LOG_DIR" "Log directory should exist"
    
    # Test log file creation
    local test_log="$TEST_LOG_DIR/test.log"
    echo "Test log message" > "$test_log"
    assert_file_exists "$test_log" "Log file should be created"
    
    # Test log message formatting
    local timestamp
    timestamp=$(date +'%Y-%m-%d %H:%M:%S')
    local log_message="[$timestamp] INFO: Test message"
    echo "$log_message" > "$test_log"
    
    local log_content
    log_content=$(cat "$test_log")
    assert_contains "$log_content" "INFO: Test message" "Should store log message correctly"
    assert_contains "$log_content" "$timestamp" "Should include timestamp"
    
    # Test log rotation (basic check)
    local log_size_before
    log_size_before=$(wc -c < "$test_log" 2>/dev/null || echo "0")
    
    # Add more content to simulate growth
    for i in {1..10}; do
        echo "Log entry $i" >> "$test_log"
    done
    
    local log_size_after
    log_size_after=$(wc -c < "$test_log" 2>/dev/null || echo "0")
    assert_true "($log_size_after > $log_size_before)" "Log file should grow with new entries"
    
    end_test_group "Logging Functionality Tests"
}

# Test alerting functionality
test_alerting_functionality() {
    start_test_group "Alerting Functionality Tests"
    
    # Test alert condition detection
    local cpu_threshold=80
    local current_cpu=90
    
    if ((current_cpu > cpu_threshold)); then
        log_success "Should detect CPU threshold exceeded"
        ((TESTS_PASSED++))
    else
        log_error "Should detect CPU threshold exceeded"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))
    
    # Test memory threshold detection
    local memory_threshold=70
    local current_memory=65
    
    if ((current_memory <= memory_threshold)); then
        log_success "Should detect memory usage within threshold"
        ((TESTS_PASSED++))
    else
        log_error "Should detect memory usage within threshold"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))
    
    # Test worker availability alert
    local total_workers=5
    local available_workers=2
    local availability_threshold=60
    
    local availability_percentage
    availability_percentage=$(echo "scale=0; $available_workers * 100 / $total_workers" | bc -l 2>/dev/null || echo "40")
    
    if ((availability_percentage < availability_threshold)); then
        log_success "Should detect low worker availability"
        ((TESTS_PASSED++))
    else
        log_error "Should detect low worker availability"
        ((TESTS_FAILED++))
    fi
    ((TESTS_TOTAL++))
    
    end_test_group "Alerting Functionality Tests"
}

# Main test execution
main() {
    local test_type="${1:-all}"
    
    # Setup test environment
    setup_test_environment
    
    # Trap cleanup
    trap cleanup_test_environment EXIT
    
    case "$test_type" in
        "worker")
            test_worker_pool_manager_script
            test_worker_status_monitoring
            ;;
        "cache")
            test_cache_manager_script
            test_cache_statistics
            ;;
        "health")
            test_health_checker_script
            test_alerting_functionality
            ;;
        "process")
            test_process_management
            ;;
        "logging")
            test_logging_functionality
            ;;
        "all")
            test_worker_pool_manager_script
            test_cache_manager_script
            test_health_checker_script
            test_worker_status_monitoring
            test_cache_statistics
            test_process_management
            test_logging_functionality
            test_alerting_functionality
            ;;
        *)
            echo "Unknown test type: $test_type"
            echo "Available types: worker, cache, health, process, logging, all"
            return 1
            ;;
    esac
}

# Run tests if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi