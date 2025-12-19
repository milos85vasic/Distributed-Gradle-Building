#!/bin/bash
# Distributed Gradle Building System - Test Framework
# Comprehensive test automation for all components

set -euo pipefail

# Test configuration
TEST_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$TEST_DIR")"
RESULTS_DIR="$TEST_DIR/results"
LOGS_DIR="$TEST_DIR/logs"
FIXTURES_DIR="$TEST_DIR/fixtures"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Test counters
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Timing
TEST_START_TIME=""
TEST_END_TIME=""

# Logging functions
log_info() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')] INFO: $1${NC}"
}

log_success() {
    echo -e "${GREEN}[$(date +'%H:%M:%S')] PASS: $1${NC}"
    ((TESTS_PASSED++))
}

log_error() {
    echo -e "${RED}[$(date +'%H:%M:%S')] FAIL: $1${NC}"
    ((TESTS_FAILED++))
}

log_warning() {
    echo -e "${YELLOW}[$(date +'%H:%M:%S')] WARN: $1${NC}"
}

log_skip() {
    echo -e "${PURPLE}[$(date +'%H:%M:%S')] SKIP: $1${NC}"
    ((TESTS_SKIPPED++))
}

log_debug() {
    if [[ "${DEBUG:-}" == "true" ]]; then
        echo -e "${CYAN}[$(date +'%H:%M:%S')] DEBUG: $1${NC}"
    fi
}

# Setup test environment
setup_test_environment() {
    log_info "Setting up test environment..."
    
    # Create directories
    mkdir -p "$RESULTS_DIR" "$LOGS_DIR" "$FIXTURES_DIR"
    mkdir -p "$FIXTURES_DIR/demo_projects"
    mkdir -p "$FIXTURES_DIR/test_data"
    mkdir -p "$FIXTURES_DIR/mock_services"
    
    # Create test results JSON
    echo '{"test_suite": "", "start_time": "", "end_time": "", "duration": 0, "tests": [], "summary": {"total": 0, "passed": 0, "failed": 0, "skipped": 0, "pass_rate": 0.0}}' > "$RESULTS_DIR/test_results.json"
    
    TEST_START_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    log_info "Test environment ready"
}

# Cleanup test environment
cleanup_test_environment() {
    log_info "Cleaning up test environment..."
    
    # Stop any running services
    stop_test_services
    
    # Clean up temporary files
    find /tmp -name "gradle-test-*" -type d -exec rm -rf {} + 2>/dev/null || true
    find /tmp -name "gradle-cache-test-*" -type d -exec rm -rf {} + 2>/dev/null || true
    
    TEST_END_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    update_test_results
    
    log_info "Test environment cleaned up"
}

# Update test results JSON
update_test_results() {
    local duration=0
    if [[ -n "$TEST_START_TIME" && -n "$TEST_END_TIME" ]]; then
        # Cross-platform date calculation
        if command -v gdate >/dev/null 2>&1; then
            # GNU date (Linux)
            duration=$(gdate -d "$TEST_END_TIME" +%s) - $(gdate -d "$TEST_START_TIME" +%s)
        else
            # BSD date (macOS) - simplified calculation
            duration=$(($(date -j -f "%Y-%m-%d %H:%M:%S" "$TEST_END_TIME" +%s) - $(date -j -f "%Y-%m-%d %H:%M:%S" "$TEST_START_TIME" +%s)))
        fi
    fi
    
    local pass_rate=0
    if ((TESTS_TOTAL > 0)); then
        pass_rate=$(echo "scale=2; $TESTS_PASSED / $TESTS_TOTAL * 100" | bc -l)
    fi
    
    jq --arg start_time "$TEST_START_TIME" \
       --arg end_time "$TEST_END_TIME" \
       --arg duration "$duration" \
       --arg total "$TESTS_TOTAL" \
       --arg passed "$TESTS_PASSED" \
       --arg failed "$TESTS_FAILED" \
       --arg skipped "$TESTS_SKIPPED" \
       --arg pass_rate "$pass_rate" \
       '.start_time = $start_time | .end_time = $end_time | .duration = ($duration | tonumber) | .summary.total = ($total | tonumber) | .summary.passed = ($passed | tonumber) | .summary.failed = ($failed | tonumber) | .summary.skipped = ($skipped | tonumber) | .summary.pass_rate = ($pass_rate | tonumber)' \
       "$RESULTS_DIR/test_results.json" > "$RESULTS_DIR/test_results.json.tmp" && \
    mv "$RESULTS_DIR/test_results.json.tmp" "$RESULTS_DIR/test_results.json"
}

# Test assertion helpers
assert_equals() {
    local expected="$1"
    local actual="$2"
    local message="${3:-Expected '$expected', got '$actual'}"
    
    ((TESTS_TOTAL++))
    
    if [[ "$expected" == "$actual" ]]; then
        log_success "$message"
        return 0
    else
        log_error "$message"
        return 1
    fi
}

assert_not_equals() {
    local not_expected="$1"
    local actual="$2"
    local message="${3:-Expected not '$not_expected', got '$actual'}"
    
    ((TESTS_TOTAL++))
    
    if [[ "$not_expected" != "$actual" ]]; then
        log_success "$message"
        return 0
    else
        log_error "$message"
        return 1
    fi
}

assert_not_empty() {
    local value="$1"
    local message="${2:-Value should not be empty}"
    
    ((TESTS_TOTAL++))
    
    if [[ -n "$value" ]]; then
        log_success "$message"
        return 0
    else
        log_error "$message"
        return 1
    fi
}

assert_true() {
    local condition="$1"
    local message="${2:-Expected true, got '$condition'}"
    
    ((TESTS_TOTAL++))
    
    if [[ "$condition" == "true" || "$condition" == "0" ]]; then
        log_success "$message"
        return 0
    else
        log_error "$message"
        return 1
    fi
}

assert_false() {
    local condition="$1"
    local message="${2:-Expected false, got '$condition'}"
    
    ((TESTS_TOTAL++))
    
    if [[ "$condition" == "false" || "$condition" == "1" ]]; then
        log_success "$message"
        return 0
    else
        log_error "$message"
        return 1
    fi
}

assert_file_exists() {
    local file_path="$1"
    local message="${2:-File should exist: $file_path}"
    
    ((TESTS_TOTAL++))
    
    if [[ -f "$file_path" ]]; then
        log_success "$message"
        return 0
    else
        log_error "$message"
        return 1
    fi
}

assert_file_not_exists() {
    local file_path="$1"
    local message="${2:-File should not exist: $file_path}"
    
    ((TESTS_TOTAL++))
    
    if [[ ! -f "$file_path" ]]; then
        log_success "$message"
        return 0
    else
        log_error "$message"
        return 1
    fi
}

assert_command_success() {
    local command="$1"
    local message="${2:-Command should succeed: $command}"
    
    ((TESTS_TOTAL++))
    
    if eval "$command" >/dev/null 2>&1; then
        log_success "$message"
        return 0
    else
        log_error "$message"
        return 1
    fi
}

assert_contains() {
    local haystack="$1"
    local needle="$2"
    local message="${3:-Expected '$haystack' to contain '$needle'}"
    
    ((TESTS_TOTAL++))
    
    if [[ "$haystack" == *"$needle"* ]]; then
        log_success "$message"
        return 0
    else
        log_error "$message"
        return 1
    fi
}

assert_http_status() {
    local expected_status="$1"
    local url="$2"
    local message="${3:-Expected HTTP $expected_status for $url}"
    
    ((TESTS_TOTAL++))
    
    local status_code
    status_code=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")
    
    if [[ "$status_code" == "$expected_status" ]]; then
        log_success "$message"
        return 0
    else
        log_error "$message (got $status_code)"
        return 1
    fi
}

assert_contains() {
    local haystack="$1"
    local needle="$2"
    local message="${3:-Expected '$haystack' to contain '$needle'}"
    
    ((TESTS_TOTAL++))
    
    if [[ "$haystack" == *"$needle"* ]]; then
        log_success "$message"
        return 0
    else
        log_error "$message"
        return 1
    fi
}

assert_not_contains() {
    local haystack="$1"
    local needle="$2"
    local message="${3:-Expected '$haystack' to not contain '$needle'}"
    
    ((TESTS_TOTAL++))
    
    if [[ "$haystack" != *"$needle"* ]]; then
        log_success "$message"
        return 0
    else
        log_error "$message"
        return 1
    fi
}

# Test group functions
start_test_group() {
    local group_name="$1"
    echo
    log_info "=== Starting test group: $group_name ==="
    echo
}

end_test_group() {
    local group_name="$1"
    echo
    log_info "=== Completed test group: $group_name ==="
    echo
}

# Wait for service to be ready
wait_for_service() {
    local url="$1"
    local timeout="${2:-30}"
    local interval="${3:-2}"
    
    log_info "Waiting for service at $url to be ready..."
    
    local elapsed=0
    while ((elapsed < timeout)); do
        if curl -sf "$url" >/dev/null 2>&1; then
            log_info "Service is ready"
            return 0
        fi
        
        sleep $interval
        elapsed=$((elapsed + interval))
        echo -n "."
    done
    
    echo
    log_error "Service did not become ready within $timeout seconds"
    return 1
}

# Stop test services
stop_test_services() {
    log_info "Stopping test services..."
    
    # Stop Go services
    if command -v lsof >/dev/null 2>&1; then
        for port in 8080 8081 8082 8083 8084; do
            local pid=$(lsof -ti :$port 2>/dev/null || true)
            if [[ -n "$pid" ]]; then
                kill -TERM "$pid" 2>/dev/null || true
                sleep 2
                kill -KILL "$pid" 2>/dev/null || true
                log_info "Stopped service on port $port (PID: $pid)"
            fi
        done
    fi
    
    # Clean up any Docker containers
    if command -v docker >/dev/null 2>&1; then
        docker ps -q --filter "name=gradle-test-" | xargs -r docker stop 2>/dev/null || true
        docker ps -aq --filter "name=gradle-test-" | xargs -r docker rm 2>/dev/null || true
    fi
}

# Create test project
create_test_project() {
    local project_name="$1"
    local project_type="${2:-java}"
    local project_path="$FIXTURES_DIR/demo_projects/$project_name"
    
    log_info "Creating test project: $project_name"
    
    mkdir -p "$project_path"
    
    case "$project_type" in
        "java")
            create_java_project "$project_path"
            ;;
        "android")
            create_android_project "$project_path"
            ;;
        "kotlin")
            create_kotlin_project "$project_path"
            ;;
        *)
            log_error "Unknown project type: $project_type"
            return 1
            ;;
    esac
    
    echo "$project_path"
}

create_java_project() {
    local project_path="$1"
    
    # Create build.gradle
    cat > "$project_path/build.gradle" << 'EOF'
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
    options.fork = true
}

tasks.withType(Test) {
    maxParallelForks = 1
    testLogging {
        events "passed", "skipped", "failed"
    }
}
EOF

    # Create source structure
    mkdir -p "$project_path/src/main/java/com/example"
    mkdir -p "$project_path/src/test/java/com/example"
    
    # Create main class
    cat > "$project_path/src/main/java/com/example/TestApp.java" << 'EOF'
package com.example;

import org.apache.commons.lang3.StringUtils;

public class TestApp {
    public static void main(String[] args) {
        System.out.println("Distributed Gradle Build Test Application");
        System.out.println("Testing: " + StringUtils.capitalize("distributed builds"));
    }
}
EOF

    # Create test class
    cat > "$project_path/src/test/java/com/example/TestAppTest.java" << 'EOF'
package com.example;

import org.junit.Test;
import static org.junit.Assert.*;

public class TestAppTest {
    
    @Test
    public void testApp() {
        // Simple test to verify build system
        assertTrue("Test should pass", true);
    }
    
    @Test
    public void testStringUtils() {
        String result = org.apache.commons.lang3.StringUtils.capitalize("hello");
        assertEquals("Hello", result);
    }
}
EOF

    # Create gradle wrapper properties
    mkdir -p "$project_path/gradle/wrapper"
    cat > "$project_path/gradle/wrapper/gradle-wrapper.properties" << 'EOF'
distributionBase=GRADLE_USER_HOME
distributionPath=wrapper/dists
distributionUrl=https\://services.gradle.org/distributions/gradle-7.4.2-bin.zip
zipStoreBase=GRADLE_USER_HOME
zipStorePath=wrapper/dists
EOF

    log_info "Created Java test project at $project_path"
}

create_android_project() {
    local project_path="$1"
    
    # Create basic Android project structure
    mkdir -p "$project_path/app/src/main/java/com/example"
    mkdir -p "$project_path/app/src/test/java/com/example"
    mkdir -p "$project_path/app/src/main/res/values"
    
    # Create build.gradle
    cat > "$project_path/build.gradle" << 'EOF'
buildscript {
    repositories {
        google()
        mavenCentral()
    }
    dependencies {
        classpath 'com.android.tools.build:gradle:7.0.4'
    }
}

plugins {
    id 'com.android.application'
}

android {
    compileSdk 31
    buildToolsVersion "30.0.3"
    
    defaultConfig {
        applicationId "com.example.testapp"
        minSdk 21
        targetSdk 31
        versionCode 1
        versionName "1.0"
        testInstrumentationRunner "androidx.test.runner.AndroidJUnitRunner"
    }
    
    buildTypes {
        release {
            minifyEnabled false
            proguardFiles getDefaultProguardFile('proguard-android.txt'), 'proguard-rules.pro'
        }
    }
    
    compileOptions {
        sourceCompatibility JavaVersion.VERSION_1_8
        targetCompatibility JavaVersion.VERSION_1_8
    }
    
    testOptions {
        unitTests.all {
            testLogging {
                events "passed", "skipped", "failed"
            }
        }
    }
}

dependencies {
    implementation 'androidx.appcompat:appcompat:1.3.1'
    testImplementation 'junit:junit:4.13.2'
    androidTestImplementation 'androidx.test.ext:junit:1.1.3'
}
EOF

    # Create main activity
    cat > "$project_path/app/src/main/java/com/example/MainActivity.java" << 'EOF'
package com.example.testapp;

import android.app.Activity;
import android.os.Bundle;

public class MainActivity extends Activity {
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);
    }
}
EOF

    # Create test
    cat > "$project_path/app/src/test/java/com/example/ExampleUnitTest.java" << 'EOF'
package com.example.testapp;

import org.junit.Test;
import static org.junit.Assert.*;

public class ExampleUnitTest {
    @Test
    public void addition_isCorrect() {
        assertEquals(4, 2 + 2);
    }
}
EOF

    log_info "Created Android test project at $project_path"
}

create_kotlin_project() {
    local project_path="$1"
    
    # Create build.gradle.kts
    cat > "$project_path/build.gradle.kts" << 'EOF'
plugins {
    application
    kotlin("jvm") version "1.6.10"
}

group = "com.example"
version = "1.0-SNAPSHOT"

repositories {
    mavenCentral()
}

dependencies {
    implementation(kotlin("stdlib"))
    implementation("org.apache.commons:commons-lang3:3.12.0")
    testImplementation(kotlin("test"))
    testImplementation("junit:junit:4.13.2")
}

application {
    mainClass.set("com.example.TestAppKt")
}

tasks.withType<org.jetbrains.kotlin.gradle.tasks.KotlinCompile> {
    kotlinOptions.jvmTarget = "1.8"
    kotlinOptions.incremental = true
}
EOF

    # Create source structure
    mkdir -p "$project_path/src/main/kotlin/com/example"
    mkdir -p "$PROJECT_PATH/src/test/kotlin/com/example"
    
    # Create main class
    cat > "$project_path/src/main/kotlin/com/example/TestApp.kt" << 'EOF'
package com.example

import org.apache.commons.lang3.StringUtils

fun main() {
    println("Distributed Gradle Build Kotlin Test Application")
    println("Testing: " + StringUtils.capitalize("kotlin distributed builds"))
}
EOF

    # Create test
    cat > "$project_path/src/test/kotlin/com/example/TestAppTest.kt" << 'EOF'
package com.example

import org.junit.Test
import org.junit.Assert.*

class TestAppTest {
    @Test
    fun testApp() {
        assertTrue("Kotlin test should pass", true)
    }
    
    @Test
    fun testStringUtils() {
        val result = StringUtils.capitalize("hello")
        assertEquals("Hello", result)
    }
}
EOF

    log_info "Created Kotlin test project at $project_path"
}

# Run test suite
run_test_suite() {
    local test_type="${1:-all}"
    log_info "Running test suite: $test_type"
    
    # Setup environment
    setup_test_environment
    
    # Trap cleanup
    trap cleanup_test_environment EXIT
    
    # Run tests based on type
    case "$test_type" in
        "unit")
            run_unit_tests
            ;;
        "integration")
            run_integration_tests
            ;;
        "performance")
            run_performance_tests
            ;;
        "security")
            run_security_tests
            ;;
        "load")
            run_load_tests
            ;;
        "e2e"|"end-to-end")
            run_end_to_end_tests
            ;;
        "all")
            run_unit_tests
            run_integration_tests
            run_performance_tests
            run_security_tests
            run_load_tests
            run_end_to_end_tests
            ;;
        *)
            log_error "Unknown test type: $test_type"
            echo "Available types: unit, integration, performance, security, load, e2e, all"
            return 1
            ;;
    esac
    
    # Print summary
    print_test_summary
    
    # Return appropriate exit code
    if ((TESTS_FAILED > 0)); then
        return 1
    else
        return 0
    fi
}

print_test_summary() {
    echo
    echo "==============================================="
    echo "Test Suite Summary"
    echo "==============================================="
    echo "Total Tests:  $TESTS_TOTAL"
    echo "Passed:       $TESTS_PASSED"
    echo "Failed:       $TESTS_FAILED"
    echo "Skipped:      $TESTS_SKIPPED"
    echo
    
    if ((TESTS_TOTAL > 0)); then
        local pass_rate
        pass_rate=$(echo "scale=2; $TESTS_PASSED / $TESTS_TOTAL * 100" | bc -l)
        echo "Pass Rate:    ${pass_rate}%"
        echo
    fi
    
    if ((TESTS_FAILED > 0)); then
        echo -e "${RED}❌ Some tests failed${NC}"
        return 1
    else
        echo -e "${GREEN}✅ All tests passed${NC}"
        return 0
    fi
}

# Placeholder for specific test types (to be implemented)
run_unit_tests() {
    start_test_group "Unit Tests"
    
    # Bash script unit tests
    run_bash_unit_tests
    
    # Go unit tests
    run_go_unit_tests
    
    end_test_group "Unit Tests"
}

run_integration_tests() {
    start_test_group "Integration Tests"

    # API integration tests
    run_api_integration_tests

    # Component integration tests
    run_component_integration_tests

    # End-to-end tests
    run_end_to_end_tests

    end_test_group "Integration Tests"
}

run_api_integration_tests() {
    log_info "Running API integration tests..."

    # Test coordinator API
    if curl -s "http://localhost:8080/api/status" >/dev/null 2>&1; then
        log_success "Coordinator API is accessible"
    else
        log_error "Coordinator API is not accessible"
    fi

    # Test cache server API
    if curl -s "http://localhost:8083/health" >/dev/null 2>&1; then
        log_success "Cache server API is accessible"
    else
        log_error "Cache server API is not accessible"
    fi

    # Test monitor API
    if curl -s "http://localhost:8084/health" >/dev/null 2>&1; then
        log_success "Monitor API is accessible"
    else
        log_error "Monitor API is not accessible"
    fi
}

run_component_integration_tests() {
    log_info "Running component integration tests..."

    # Test Go integration tests
    if cd "$PROJECT_DIR/go" && go test ./tests/integration -v -timeout 30s >/dev/null 2>&1; then
        log_success "Go component integration tests passed"
    else
        log_error "Go component integration tests failed"
    fi
}

run_end_to_end_tests() {
    log_info "Running end-to-end tests..."

    # Run the comprehensive end-to-end test
    if bash "$TEST_DIR/integration/test_end_to_end.sh" >/dev/null 2>&1; then
        log_success "End-to-end tests passed"
    else
        log_error "End-to-end tests failed"
    fi
}

run_performance_tests() {
    start_test_group "Performance Tests"

    # Build performance tests
    run_build_performance_tests

    # Cache performance tests
    run_cache_performance_tests

    # Worker performance tests
    run_worker_performance_tests

    # Load testing
    run_load_tests

    end_test_group "Performance Tests"
}

run_build_performance_tests() {
    log_info "Running build performance tests..."

    # Test Go build performance
    local start_time=$(date +%s)
    if cd "$PROJECT_DIR/go" && go test -bench=. -run=^$ -count=3 ./... >/dev/null 2>&1; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        log_success "Build performance test completed in ${duration}s"
    else
        log_error "Build performance test failed"
    fi
}

run_cache_performance_tests() {
    log_info "Running cache performance tests..."

    # Test cache operations performance
    if bash "$TEST_DIR/performance/test_cache_performance.sh" >/dev/null 2>&1; then
        log_success "Cache performance tests passed"
    else
        log_error "Cache performance tests failed"
    fi
}

run_worker_performance_tests() {
    log_info "Running worker performance tests..."

    # Test worker pool performance
    if bash "$TEST_DIR/performance/test_worker_pool_performance_test.go" >/dev/null 2>&1; then
        log_success "Worker performance tests passed"
    else
        log_error "Worker performance tests failed"
    fi
}

run_load_tests() {
    log_info "Running load tests..."

    # Test system under load
    if bash "$TEST_DIR/load/test_stress_test.sh" >/dev/null 2>&1; then
        log_success "Load tests passed"
    else
        log_error "Load tests failed"
    fi
}

run_security_tests() {
    start_test_group "Security Tests"
    
    # Authentication tests
    run_authentication_tests
    
    # Data protection tests
    run_data_protection_tests
    
    end_test_group "Security Tests"
}

run_load_tests() {
    start_test_group "Load Tests"
    
    # Concurrent build tests
    run_concurrent_build_tests
    
    # Stress tests
    run_stress_tests
    
    end_test_group "Load Tests"
}

run_end_to_end_tests() {
    start_test_group "End-to-End Tests"
    
    # Complete workflow tests
    run_workflow_tests
    
    end_test_group "End-to-End Tests"
}

# Placeholder test runners (to be implemented in separate files)
run_bash_unit_tests() {
    log_info "Running Bash unit tests..."
    source "$TEST_DIR/unit/test_bash_scripts.sh"
}

run_go_unit_tests() {
    log_info "Running Go unit tests..."
    cd "$PROJECT_DIR"
    go test -v ./go/...
}

# Main execution
main() {
    local command="${1:-help}"
    
    case "$command" in
        "unit"|"integration"|"performance"|"security"|"load"|"e2e"|"all")
            run_test_suite "$command"
            ;;
        "setup")
            setup_test_environment
            ;;
        "cleanup")
            cleanup_test_environment
            ;;
        "help"|"-h"|"--help")
            cat << EOF
Usage: $0 [COMMAND] [OPTIONS]

Commands:
  unit          Run unit tests only
  integration   Run integration tests only
  performance   Run performance tests only
  security      Run security tests only
  load          Run load tests only
  e2e           Run end-to-end tests only
  all           Run all test suites (default)
  setup         Setup test environment
  cleanup       Cleanup test environment
  help          Show this help message

Examples:
  $0                    # Run all tests
  $0 unit              # Run unit tests only
  $0 integration       # Run integration tests only
  
Environment Variables:
  DEBUG=true           Enable debug logging
  TEST_TIMEOUT=300      Test timeout in seconds
EOF
            ;;
        *)
            log_error "Unknown command: $command"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi