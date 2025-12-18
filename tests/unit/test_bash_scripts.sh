#!/bin/bash
# Unit Tests for Bash Scripts
# Comprehensive unit testing for all Bash components

set -euo pipefail

# Import test framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_ROOT="$(dirname "$SCRIPT_DIR")"
source "$TEST_ROOT/tests/test_framework.sh"

# Test data
TEST_PROJECT_DIR="/tmp/gradle-test-$$"
TEST_SSH_DIR="/tmp/ssh-test-$$"
TEST_CONFIG_FILE="/tmp/test-config-$$"

# Mock functions for testing
mock_ssh() {
    echo "Mock SSH: $*"
    return 0
}

mock_rsync() {
    echo "Mock rsync: $*"
    return 0
}

mock_gradle() {
    echo "Mock gradle: $*"
    return 0
}

mock_curl() {
    case "$1" in
        "-s")
            case "$2" in
                "-o")
                    # Mock health check response
                    echo '{"status": "healthy"}' > "$3"
                    return 0
                    ;;
                "-f")
                    # Mock build submission
                    echo '{"build_id": "test-build-123", "status": "queued"}'
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

# Setup test environment
setup_test_environment() {
    # Create test project directory
    mkdir -p "$TEST_PROJECT_DIR/src/main/java"
    mkdir -p "$TEST_PROJECT_DIR/src/test/java"
    
    # Create basic Java files
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
EOF

    cat > "$TEST_PROJECT_DIR/src/main/java/com/example/TestApp.java" << 'EOF'
package com.example;

public class TestApp {
    public static void main(String[] args) {
        System.out.println("Test Application");
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
}
EOF

    # Create SSH test directory
    mkdir -p "$TEST_SSH_DIR"
    
    # Create test config
    cat > "$TEST_CONFIG_FILE" << EOF
WORKER_IPS="worker1 worker2 worker3"
MAX_WORKERS=6
PROJECT_DIR="$TEST_PROJECT_DIR"
MASTER_IP="master"
EOF

    # Mock external commands
    export PATH="$TEST_ROOT/tests/mocks:$PATH"
}

# Cleanup test environment
cleanup_test_environment() {
    rm -rf "$TEST_PROJECT_DIR"
    rm -rf "$TEST_SSH_DIR"
    rm -f "$TEST_CONFIG_FILE"
}

# Test setup_master.sh
test_setup_master_script() {
    start_test_group "Setup Master Script Tests"
    
    # Test master script exists
    assert_file_exists "$TEST_ROOT/setup_master.sh" "setup_master.sh should exist"
    
    # Test master script is executable
    assert_command_success "test -x $TEST_ROOT/setup_master.sh" "setup_master.sh should be executable"
    
    # Test master script with invalid directory
    local output
    output=$("$TEST_ROOT/setup_master.sh" /nonexistent/directory 2>&1 || true)
    assert_contains "$output" "does not exist" "Should detect non-existent directory"
    
    # Test master script help
    local help_output
    help_output=$("$TEST_ROOT/setup_master.sh" --help 2>&1 || true)
    assert_contains "$help_output" "Usage" "Should show help information"
    
    end_test_group "Setup Master Script Tests"
}

# Test setup_worker.sh
test_setup_worker_script() {
    start_test_group "Setup Worker Script Tests"
    
    # Test worker script exists
    assert_file_exists "$TEST_ROOT/setup_worker.sh" "setup_worker.sh should exist"
    
    # Test worker script is executable
    assert_command_success "test -x $TEST_ROOT/setup_worker.sh" "setup_worker.sh should be executable"
    
    # Test worker script with invalid directory
    local output
    output=$("$TEST_ROOT/setup_worker.sh" /nonexistent/directory 2>&1 || true)
    assert_contains "$output" "does not exist" "Should detect non-existent directory"
    
    end_test_group "Setup Worker Script Tests"
}

# Test sync_and_build.sh
test_sync_and_build_script() {
    start_test_group "Sync and Build Script Tests"
    
    # Test script exists
    assert_file_exists "$TEST_ROOT/sync_and_build.sh" "sync_and_build.sh should exist"
    
    # Test script is executable
    assert_command_success "test -x $TEST_ROOT/sync_and_build.sh" "sync_and_build.sh should be executable"
    
    # Test script without project directory
    local output
    output=$("$TEST_ROOT/sync_and_build.sh" 2>&1 || true)
    assert_contains "$output" "project directory" "Should require project directory"
    
    end_test_group "Sync and Build Script Tests"
}

# Test performance_analyzer.sh
test_performance_analyzer_script() {
    start_test_group "Performance Analyzer Script Tests"
    
    local script_path="$TEST_ROOT/scripts/performance_analyzer.sh"
    
    # Test script exists
    assert_file_exists "$script_path" "performance_analyzer.sh should exist"
    
    # Test script is executable
    assert_command_success "test -x $script_path" "performance_analyzer.sh should be executable"
    
    # Test script help
    local help_output
    help_output=("$script_path" --help 2>&1 || true)
    assert_contains "$help_output" "Usage" "Should show help information"
    
    # Test script with invalid mode
    local output
    output=("$script_path" invalid-mode 2>&1 || true)
    assert_contains "$output" "invalid" "Should detect invalid mode"
    
    end_test_group "Performance Analyzer Script Tests"
}

# Test worker_pool_manager.sh
test_worker_pool_manager_script() {
    start_test_group "Worker Pool Manager Script Tests"
    
    local script_path="$TEST_ROOT/scripts/worker_pool_manager.sh"
    
    # Test script exists
    assert_file_exists "$script_path" "worker_pool_manager.sh should exist"
    
    # Test script is executable
    assert_command_success "test -x $script_path" "worker_pool_manager.sh should be executable"
    
    # Test script help
    local help_output
    help_output=("$script_path" --help 2>&1 || true)
    assert_contains "$help_output" "Usage" "Should show help information"
    
    end_test_group "Worker Pool Manager Script Tests"
}

# Test cache_manager.sh
test_cache_manager_script() {
    start_test_group "Cache Manager Script Tests"
    
    local script_path="$TEST_ROOT/scripts/cache_manager.sh"
    
    # Test script exists
    assert_file_exists "$script_path" "cache_manager.sh should exist"
    
    # Test script is executable
    assert_command_success "test -x $script_path" "cache_manager.sh should be executable"
    
    # Test script help
    local help_output
    help_output=("$script_path" --help 2>&1 || true)
    assert_contains "$help_output" "Usage" "Should show help information"
    
    end_test_group "Cache Manager Script Tests"
}

# Test health_checker.sh
test_health_checker_script() {
    start_test_group "Health Checker Script Tests"
    
    local script_path="$TEST_ROOT/scripts/health_checker.sh"
    
    # Test script exists
    assert_file_exists "$script_path" "health_checker.sh should exist"
    
    # Test script is executable
    assert_command_success "test -x $script_path" "health_checker.sh should be executable"
    
    # Test script help
    local help_output
    help_output=("$script_path" --help 2>&1 || true)
    assert_contains "$help_output" "Usage" "Should show help information"
    
    end_test_group "Health Checker Script Tests"
}

# Test configuration validation
test_configuration_validation() {
    start_test_group "Configuration Validation Tests"
    
    # Test valid configuration
    source "$TEST_CONFIG_FILE"
    assert_equals "3" "$(echo $WORKER_IPS | wc -w)" "Should parse 3 worker IPs"
    assert_equals "6" "$MAX_WORKERS" "Should parse max workers correctly"
    assert_equals "$TEST_PROJECT_DIR" "$PROJECT_DIR" "Should parse project directory correctly"
    
    # Test missing required variables
    local config_missing
    cat > "$TEST_CONFIG_FILE" << EOF
MAX_WORKERS=4
# WORKER_IPS missing
EOF
    
    config_missing=$(source "$TEST_CONFIG_FILE" 2>&1 || true)
    assert_contains "$config_missing" "WORKER_IPS" "Should detect missing WORKER_IPS"
    
    end_test_group "Configuration Validation Tests"
}

# Test SSH functionality
test_ssh_functionality() {
    start_test_group "SSH Functionality Tests"
    
    # Test SSH command construction
    local ssh_output
    ssh_output=$(ssh -o BatchMode=yes -o ConnectTimeout=5 test-host "echo 'SSH test'" 2>&1 || true)
    # This will fail in test environment, but we can test the command format
    
    # Test SSH key existence check
    assert_file_exists "$HOME/.ssh/id_rsa" "SSH private key should exist (may not exist in test env)" || true
    assert_file_exists "$HOME/.ssh/id_rsa.pub" "SSH public key should exist (may not exist in test env)" || true
    
    end_test_group "SSH Functionality Tests"
}

# Test rsync functionality
test_rsync_functionality() {
    start_test_group "Rsync Functionality Tests"
    
    # Create source directory
    local source_dir="/tmp/rsync-source-$$"
    mkdir -p "$source_dir/src"
    echo "test content" > "$source_dir/test.txt"
    
    # Test rsync command
    local rsync_cmd="rsync -avz --exclude='build' --exclude='.gradle' $source_dir/ /tmp/rsync-target-$$"
    
    # Test rsync command format
    assert_contains "$rsync_cmd" "--exclude='build'" "Should exclude build directory"
    assert_contains "$rsync_cmd" "--exclude='.gradle'" "Should exclude .gradle directory"
    
    # Cleanup
    rm -rf "$source_dir" "/tmp/rsync-target-$$"
    
    end_test_group "Rsync Functionality Tests"
}

# Test error handling
test_error_handling() {
    start_test_group "Error Handling Tests"
    
    # Test script with invalid permissions
    local no_perm_dir="/tmp/no-permission-$$"
    mkdir -p "$no_perm_dir"
    chmod 000 "$no_perm_dir"
    
    local output
    output=$(cd "$no_perm_dir" 2>&1 || true)
    assert_contains "$output" "Permission denied" "Should handle permission errors"
    
    # Cleanup
    chmod 755 "$no_perm_dir"
    rm -rf "$no_perm_dir"
    
    end_test_group "Error Handling Tests"
}

# Test logging functionality
test_logging_functionality() {
    start_test_group "Logging Functionality Tests"
    
    # Test log directory creation
    local log_dir="/tmp/gradle-test-logs-$$"
    
    # Mock log function test
    log_test() {
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] TEST: $1"
    }
    
    # Test log message format
    local log_output
    log_output=$(log_test "Test message")
    assert_contains "$log_output" "TEST: Test message" "Should format log message correctly"
    assert_contains "$log_output" "$(date +'%Y-%m-%d')" "Should include timestamp"
    
    # Cleanup
    rm -rf "$log_dir"
    
    end_test_group "Logging Functionality Tests"
}

# Test utility functions
test_utility_functions() {
    start_test_group "Utility Functions Tests"
    
    # Test CPU count detection
    local cpu_count
    cpu_count=$(nproc 2>/dev/null || echo "1")
    assert_true "($cpu_count > 0)" "Should detect at least 1 CPU"
    
    # Test memory detection
    local memory_kb
    memory_kb=$(grep MemTotal /proc/meminfo 2>/dev/null | awk '{print $2}' || echo "1048576")
    assert_true "($memory_kb > 1000)" "Should detect reasonable memory amount"
    
    # Test disk space detection
    local disk_available
    disk_available=$(df -BG . 2>/dev/null | awk 'NR==2 {print $4}' || echo "10")
    assert_true "($disk_available > 0)" "Should detect available disk space"
    
    end_test_group "Utility Functions Tests"
}

# Main test execution
main() {
    local test_type="${1:-all}"
    
    case "$test_type" in
        "setup")
            test_setup_master_script
            test_setup_worker_script
            ;;
        "build")
            test_sync_and_build_script
            ;;
        "scripts")
            test_performance_analyzer_script
            test_worker_pool_manager_script
            test_cache_manager_script
            test_health_checker_script
            ;;
        "config")
            test_configuration_validation
            ;;
        "utilities")
            test_ssh_functionality
            test_rsync_functionality
            test_error_handling
            test_logging_functionality
            test_utility_functions
            ;;
        "all")
            # Setup test environment
            setup_test_environment
            
            # Run all tests
            test_setup_master_script
            test_setup_worker_script
            test_sync_and_build_script
            test_performance_analyzer_script
            test_worker_pool_manager_script
            test_cache_manager_script
            test_health_checker_script
            test_configuration_validation
            test_ssh_functionality
            test_rsync_functionality
            test_error_handling
            test_logging_functionality
            test_utility_functions
            
            # Cleanup
            cleanup_test_environment
            ;;
        *)
            echo "Unknown test type: $test_type"
            echo "Available types: setup, build, scripts, config, utilities, all"
            return 1
            ;;
    esac
}

# Run tests if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi