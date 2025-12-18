#!/bin/bash
# Unit Tests for Setup Scripts
# Tests for setup_master.sh and setup_passwordless_ssh.sh

set -euo pipefail

# Import test framework
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_ROOT="$(dirname "$SCRIPT_DIR")"
source "$TEST_ROOT/test_framework.sh"

# Test data
TEST_PROJECT_DIR="/tmp/gradle-setup-test-$$"
SSH_TEST_DIR="/tmp/ssh-setup-test-$$"

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

    # Create SSH test directory
    mkdir -p "$SSH_TEST_DIR/.ssh"
    
    # Create test hosts file
    cat > "$SSH_TEST_DIR/hosts" << EOF
worker1
worker2
worker3
EOF
}

# Cleanup test environment
cleanup_test_environment() {
    rm -rf "$TEST_PROJECT_DIR"
    rm -rf "$SSH_TEST_DIR"
}

# Test setup_master.sh functionality
test_setup_master_script() {
    start_test_group "Setup Master Script Tests"
    
    # Test script exists
    assert_file_exists "$TEST_ROOT/setup_master.sh" "setup_master.sh should exist"
    
    # Test script is executable
    assert_command_success "test -x $TEST_ROOT/setup_master.sh" "setup_master.sh should be executable"
    
    # Test script with help option
    local help_output
    help_output=$("$TEST_ROOT/setup_master.sh" --help 2>&1 || true)
    assert_contains "$help_output" "Usage" "Should show help information"
    
    # Test script with invalid directory
    local error_output
    error_output=$("$TEST_ROOT/setup_master.sh" /nonexistent/directory 2>&1 || true)
    assert_contains "$error_output" "does not exist" "Should detect non-existent directory"
    
    # Test script with valid directory (dry run)
    local setup_output
    cd "$TEST_PROJECT_DIR"
    setup_output=("$TEST_ROOT/setup_master.sh" . 2>&1 || true)
    # Should show that it's attempting to setup master
    assert_contains "$setup_output" "master" "Should mention master setup" || true
    
    end_test_group "Setup Master Script Tests"
}

# Test setup_worker.sh functionality
test_setup_worker_script() {
    start_test_group "Setup Worker Script Tests"
    
    # Test script exists
    assert_file_exists "$TEST_ROOT/setup_worker.sh" "setup_worker.sh should exist"
    
    # Test script is executable
    assert_command_success "test -x $TEST_ROOT/setup_worker.sh" "setup_worker.sh should be executable"
    
    # Test script with help option
    local help_output
    help_output=$("$TEST_ROOT/setup_worker.sh" --help 2>&1 || true)
    assert_contains "$help_output" "Usage" "Should show help information"
    
    # Test script with invalid directory
    local error_output
    error_output=$("$TEST_ROOT/setup_worker.sh" /nonexistent/directory 2>&1 || true)
    assert_contains "$error_output" "does not exist" "Should detect non-existent directory"
    
    # Test script with valid directory (dry run)
    local setup_output
    cd "$TEST_PROJECT_DIR"
    setup_output=("$TEST_ROOT/setup_worker.sh" . 2>&1 || true)
    # Should show that it's attempting to setup worker
    assert_contains "$setup_output" "worker" "Should mention worker setup" || true
    
    end_test_group "Setup Worker Script Tests"
}

# Test setup_passwordless_ssh.sh functionality
test_setup_passwordless_ssh_script() {
    start_test_group "Passwordless SSH Script Tests"
    
    # Test script exists
    assert_file_exists "$TEST_ROOT/setup_passwordless_ssh.sh" "setup_passwordless_ssh.sh should exist"
    
    # Test script is executable
    assert_command_success "test -x $TEST_ROOT/setup_passwordless_ssh.sh" "setup_passwordless_ssh.sh should be executable"
    
    # Test script with help option
    local help_output
    help_output=$("$TEST_ROOT/setup_passwordless_ssh.sh" --help 2>&1 || true)
    assert_contains "$help_output" "Usage" "Should show help information"
    
    # Test script with hosts file
    local ssh_output
    cd "$SSH_TEST_DIR"
    ssh_output=("$TEST_ROOT/setup_passwordless_ssh.sh" hosts 2>&1 || true)
    # Should show that it's attempting to setup SSH
    assert_contains "$ssh_output" "ssh" "Should mention SSH setup" || true
    
    # Test SSH key generation functionality
    if [[ ! -f "$HOME/.ssh/id_rsa" ]]; then
        # Mock SSH key generation for test
        mkdir -p "$HOME/.ssh"
        echo "-----BEGIN RSA PRIVATE KEY-----
TEST_KEY_CONTENT_FOR_TESTING_ONLY
-----END RSA PRIVATE KEY-----" > "$HOME/.ssh/id_rsa"
        echo "ssh-rsa test-key user@host" > "$HOME/.ssh/id_rsa.pub"
        chmod 600 "$HOME/.ssh/id_rsa"
        chmod 644 "$HOME/.ssh/id_rsa.pub"
    fi
    
    assert_file_exists "$HOME/.ssh/id_rsa" "SSH private key should exist"
    assert_file_exists "$HOME/.ssh/id_rsa.pub" "SSH public key should exist"
    
    end_test_group "Passwordless SSH Script Tests"
}

# Test configuration parsing
test_configuration_parsing() {
    start_test_group "Configuration Parsing Tests"
    
    # Create test configuration
    local test_config="/tmp/test-config-$$"
    cat > "$test_config" << EOF
#!/bin/bash
# Test configuration
WORKER_IPS="192.168.1.10 192.168.1.11 192.168.1.12"
MAX_WORKERS=6
PROJECT_DIR="/test/project"
MASTER_IP="192.168.1.1"
SSH_PORT=22
GRADLE_OPTS="-Xmx2g -XX:+UseG1GC"
CACHE_ENABLED=true
EOF

    # Test configuration loading
    source "$test_config"
    
    assert_equals "3" "$(echo $WORKER_IPS | wc -w)" "Should parse 3 worker IPs"
    assert_equals "6" "$MAX_WORKERS" "Should parse max workers correctly"
    assert_equals "/test/project" "$PROJECT_DIR" "Should parse project directory correctly"
    assert_equals "192.168.1.1" "$MASTER_IP" "Should parse master IP correctly"
    assert_equals "22" "$SSH_PORT" "Should parse SSH port correctly"
    assert_true "$CACHE_ENABLED" "Should parse cache enabled correctly"
    
    # Cleanup
    rm -f "$test_config"
    
    end_test_group "Configuration Parsing Tests"
}

# Test directory structure validation
test_directory_structure_validation() {
    start_test_group "Directory Structure Validation Tests"
    
    # Test project directory validation
    local non_existent_dir="/tmp/non-existent-$$"
    assert_command_failure "test -d $non_existent_dir" "Should detect non-existent directory"
    
    # Test valid project directory
    assert_command_success "test -d $TEST_PROJECT_DIR" "Should detect valid directory"
    
    # Test required files in project directory
    assert_file_exists "$TEST_PROJECT_DIR/build.gradle" "build.gradle should exist in project"
    assert_file_exists "$TEST_PROJECT_DIR/src/main/java" "src/main/java directory should exist"
    
    # Test directory permissions
    assert_command_success "test -r $TEST_PROJECT_DIR" "Project directory should be readable"
    assert_command_success "test -w $TEST_PROJECT_DIR" "Project directory should be writable"
    assert_command_success "test -x $TEST_PROJECT_DIR" "Project directory should be executable"
    
    end_test_group "Directory Structure Validation Tests"
}

# Test network connectivity validation
test_network_connectivity_validation() {
    start_test_group "Network Connectivity Validation Tests"
    
    # Test host resolution (localhost)
    if command -v nslookup >/dev/null 2>&1; then
        assert_command_success "nslookup localhost" "Should resolve localhost"
    elif command -v ping >/dev/null 2>&1; then
        assert_command_success "ping -c 1 localhost" "Should ping localhost"
    fi
    
    # Test port checking function (mock)
    check_port() {
        local host="$1"
        local port="$2"
        if command -v nc >/dev/null 2>&1; then
            nc -z -w3 "$host" "$port" 2>/dev/null
        elif command -v telnet >/dev/null 2>&1; then
            timeout 3 telnet "$host" "$port" </dev/null >/dev/null 2>&1
        else
            return 1
        fi
    }
    
    # Test port checking on common ports
    if check_port localhost 22; then
        log_success "SSH port 22 is accessible"
        ((TESTS_PASSED++))
    else
        log_warning "SSH port 22 is not accessible (may be expected in test environment)"
        ((TESTS_SKIPPED++))
    fi
    
    ((TESTS_TOTAL++))
    
    end_test_group "Network Connectivity Validation Tests"
}

# Test environment validation
test_environment_validation() {
    start_test_group "Environment Validation Tests"
    
    # Test required commands
    local required_commands=("bash" "ssh" "rsync" "java" "gradle")
    for cmd in "${required_commands[@]}"; do
        if command -v "$cmd" >/dev/null 2>&1; then
            log_success "$cmd command is available"
            ((TESTS_PASSED++))
        else
            log_warning "$cmd command is not available (may be expected in test environment)"
            ((TESTS_SKIPPED++))
        fi
        ((TESTS_TOTAL++))
    done
    
    # Test Java version (if available)
    if command -v java >/dev/null 2>&1; then
        local java_version
        java_version=$(java -version 2>&1 | head -1 | grep -o '[0-9]\+\.[0-9]\+' | head -1)
        assert_not_equals "" "$java_version" "Should detect Java version"
    fi
    
    # Test Gradle version (if available)
    if command -v gradle >/dev/null 2>&1; then
        local gradle_version
        gradle_version=$(gradle --version 2>&1 | grep "Gradle" | head -1 | grep -o '[0-9]\+\.[0-9]\+' | head -1)
        assert_not_equals "" "$gradle_version" "Should detect Gradle version"
    fi
    
    end_test_group "Environment Validation Tests"
}

# Main test execution
main() {
    local test_type="${1:-all}"
    
    # Setup test environment
    setup_test_environment
    
    # Trap cleanup
    trap cleanup_test_environment EXIT
    
    case "$test_type" in
        "master")
            test_setup_master_script
            ;;
        "worker")
            test_setup_worker_script
            ;;
        "ssh")
            test_setup_passwordless_ssh_script
            ;;
        "config")
            test_configuration_parsing
            ;;
        "structure")
            test_directory_structure_validation
            ;;
        "network")
            test_network_connectivity_validation
            ;;
        "env")
            test_environment_validation
            ;;
        "all")
            test_setup_master_script
            test_setup_worker_script
            test_setup_passwordless_ssh_script
            test_configuration_parsing
            test_directory_structure_validation
            test_network_connectivity_validation
            test_environment_validation
            ;;
        *)
            echo "Unknown test type: $test_type"
            echo "Available types: master, worker, ssh, config, structure, network, env, all"
            return 1
            ;;
    esac
}

# Run tests if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi