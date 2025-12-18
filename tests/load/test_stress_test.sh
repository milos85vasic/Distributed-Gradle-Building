#!/bin/bash

# Distributed Gradle Building - Stress Test
# Tests system behavior under extreme load conditions

set -euo pipefail

# Import test framework
source "$(dirname "$0")/../test_framework.sh"

# Configuration
STRESS_PROJECT_NAME="stress-test"
MAX_CONCURRENT_BUILDS=20
BUILD_TIMEOUT=600
COORDINATOR_PORT=${COORDINATOR_PORT:-8082}
RESULTS_DIR="${TEST_RESULTS_DIR}/stress_test"
CPU_THRESHOLD=90
MEMORY_THRESHOLD=85

# Test data setup
setup_stress_test() {
    echo "Setting up stress test environment..."
    
    # Create test results directory
    mkdir -p "$RESULTS_DIR"
    
    # Create multiple test projects with varying complexity
    for i in $(seq 1 $MAX_CONCURRENT_BUILDS); do
        project_dir="${TEST_WORKSPACE}/${STRESS_PROJECT_NAME}-${i}"
        mkdir -p "$project_dir/src/main/java"
        
        # Create different complexity levels
        complexity=$((i % 4))
        
        case $complexity in
            0) # Simple project
                cat > "$project_dir/src/main/java/Simple${i}.java" << EOF
public class Simple${i} {
    public static void main(String[] args) {
        System.out.println("Simple stress test ${i}");
    }
}
EOF
                ;;
            1) # Medium complexity
                cat > "$project_dir/src/main/java/Medium${i}.java" << EOF
public class Medium${i} {
    public static void main(String[] args) {
        System.out.println("Medium stress test ${i}");
        for (int i = 0; i < 10000; i++) {
            double result = Math.sqrt(i) * Math.log(i + 1);
        }
    }
}
EOF
                ;;
            2) # CPU intensive
                cat > "$project_dir/src/main/java/CPU${i}.java" << EOF
public class CPU${i} {
    public static void main(String[] args) {
        System.out.println("CPU intensive stress test ${i}");
        long start = System.currentTimeMillis();
        while (System.currentTimeMillis() - start < 5000) { // 5 seconds of CPU work
            for (int i = 0; i < 1000; i++) {
                Math.pow(Math.random(), Math.random());
            }
        }
    }
}
EOF
                ;;
            3) # Memory intensive
                cat > "$project_dir/src/main/java/Memory${i}.java" << EOF
import java.util.ArrayList;
import java.util.List;

public class Memory${i} {
    public static void main(String[] args) {
        System.out.println("Memory intensive stress test ${i}");
        List<byte[]> memoryList = new ArrayList<>();
        for (int i = 0; i < 100; i++) {
            memoryList.add(new byte[1024 * 100]); // 100KB chunks
        }
        System.out.println("Allocated " + (memoryList.size() * 100) + "KB");
        memoryList.clear();
        System.gc();
    }
}
EOF
                ;;
        esac
        
        # Create build.gradle
        cat > "$project_dir/build.gradle" << EOF
plugins {
    application
    java
}

application {
    mainClass = "${complexity^}${i}"
}

java {
    sourceCompatibility = JavaVersion.VERSION_11
    targetCompatibility = JavaVersion.VERSION_11
}

tasks.withType(JavaCompile) {
    options.encoding = 'UTF-8'
}
EOF
        
        echo "Created stress test project: $project_dir (complexity: $complexity)"
    done
    
    test_pass "Stress test environment setup completed"
}

# Start coordinator for stress testing
start_stress_coordinator() {
    echo "Starting coordinator for stress testing..."
    
    cd "$TEST_WORKSPACE"
    export COORDINATOR_PORT="$COORDINATOR_PORT"
    export LOG_LEVEL="WARN"  # Reduce log noise during stress test
    export MAX_CONCURRENT_BUILDS="$MAX_CONCURRENT_BUILDS"
    
    # Use Go coordinator
    if command -v "${PROJECT_DIR}/go/bin/coordinator" >/dev/null 2>&1; then
        "${PROJECT_DIR}/go/bin/coordinator" > "${RESULTS_DIR}/coordinator_stress.log" 2>&1 &
        COORDINATOR_PID=$!
        sleep 3
        test_pass "Stress coordinator started (PID: $COORDINATOR_PID)"
    else
        test_fail "Coordinator binary not found"
        return 1
    fi
}

# Monitor system resources
monitor_system_resources() {
    echo "Starting system resource monitoring..."
    
    # Create resource monitoring script
    cat > "${RESULTS_DIR}/monitor_resources.sh" << 'EOF'
#!/bin/bash
RESULTS_FILE="$1"
INTERVAL="$2"

while true; do
    timestamp=$(date +%s%3N)  # Milliseconds
    cpu=$(top -l 1 -n 0 | grep "CPU usage" | awk '{print $3}' | sed 's/%//')
    memory_pressure=$(memory_pressure | grep "System-wide memory free percentage" | awk '{print $5}' | sed 's/%//')
    
    # Count running processes
    build_processes=$(ps aux | grep -v grep | grep -E "(java|gradle)" | wc -l)
    
    echo "$timestamp,$cpu,$memory_pressure,$build_processes" >> "$RESULTS_FILE"
    sleep "$INTERVAL"
done
EOF
    
    chmod +x "${RESULTS_DIR}/monitor_resources.sh"
    
    # Start monitoring in background
    "${RESULTS_DIR}/monitor_resources.sh" "${RESULTS_DIR}/system_resources.csv" "1" &
    RESOURCE_MONITOR_PID=$!
    
    test_pass "System resource monitoring started (PID: $RESOURCE_MONITOR_PID)"
}

# Submit stress load
submit_stress_load() {
    echo "Submitting stress load: $MAX_CONCURRENT_BUILDS concurrent builds..."
    
    local submission_start=$(date +%s)
    local submission_count=0
    local submission_failures=0
    
    # Submit builds in batches to avoid overwhelming the system
    local batch_size=5
    local total_batches=$(((MAX_CONCURRENT_BUILDS + batch_size - 1) / batch_size))
    
    for batch in $(seq 1 $total_batches); do
        echo "Submitting batch $batch of $total_batches..."
        
        local start_idx=$(((batch - 1) * batch_size + 1))
        local end_idx=$((batch * batch_size))
        if [ "$end_idx" -gt "$MAX_CONCURRENT_BUILDS" ]; then
            end_idx=$MAX_CONCURRENT_BUILDS
        fi
        
        # Submit batch
        for i in $(seq $start_idx $end_idx); do
            {
                project_dir="${TEST_WORKSPACE}/${STRESS_PROJECT_NAME}-${i}"
                complexity=$((i % 4))
                
                build_response=$(curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
                    -H "Content-Type: application/json" \
                    -d "{
                        \"project_path\": \"$project_dir\",
                        \"task_name\": \"build\",
                        \"cache_enabled\": $(($i % 2 == 0 ? "true" : "false")),
                        \"priority\": \"normal\"
                    }")
                
                build_id=$(echo "$build_response" | jq -r '.build_id // empty')
                submission_time=$(date +%s%3N)
                
                if [ -n "$build_id" ]; then
                    echo "${build_id},${submission_time},${complexity}" >> "${RESULTS_DIR}/submissions.csv"
                    submission_count=$((submission_count + 1))
                else
                    echo "$(date +%s%3N),failure,${complexity}" >> "${RESULTS_DIR}/submissions.csv"
                    submission_failures=$((submission_failures + 1))
                fi
            } &
        done
        
        # Wait for batch submission to complete
        wait
        
        # Brief pause between batches
        sleep 1
    done
    
    local submission_end=$(date +%s)
    local submission_time=$((submission_end - submission_start))
    
    echo "Stress load submission completed:"
    echo "- Time: ${submission_time}s"
    echo "- Successful submissions: $submission_count"
    echo "- Failed submissions: $submission_failures"
    
    test_pass "Stress load submitted (${submission_count} successful, ${submission_failures} failed)"
}

# Monitor stress execution
monitor_stress_execution() {
    echo "Monitoring stress test execution..."
    
    local execution_start=$(date +%s)
    local monitoring_interval=5
    local peak_active_builds=0
    local peak_cpu_usage=0
    local peak_memory_usage=0
    
    while true; do
        local current_time=$(date +%s)
        local elapsed_time=$((current_time - execution_start))
        
        # Get build status from coordinator
        status_response=$(curl -s "http://localhost:$COORDINATOR_PORT/api/builds/status" 2>/dev/null || echo "{}")
        active_builds=$(echo "$status_response" | jq -r '.active_builds // 0')
        completed_builds=$(echo "$status_response" | jq -r '.completed_builds // 0')
        failed_builds=$(echo "$status_response" | jq -r '.failed_builds // 0')
        
        # Update peaks
        if [ "$active_builds" -gt "$peak_active_builds" ]; then
            peak_active_builds=$active_builds
        fi
        
        # Get current resource usage
        if [ -f "${RESULTS_DIR}/system_resources.csv" ]; then
            current_cpu=$(tail -1 "${RESULTS_DIR}/system_resources.csv" | cut -d',' -f2)
            current_memory=$(tail -1 "${RESULTS_DIR}/system_resources.csv" | cut -d',' -f3)
            
            if [ -n "$current_cpu" ] && [ "$current_cpu" -gt "$peak_cpu_usage" ]; then
                peak_cpu_usage=$current_cpu
            fi
            
            if [ -n "$current_memory" ] && [ "$current_memory" -gt "$peak_memory_usage" ]; then
                peak_memory_usage=$current_memory
            fi
        fi
        
        echo "[${elapsed_time}s] Active: $active_builds, Completed: $completed_builds, Failed: $failed_builds"
        echo "Peak CPU: ${peak_cpu_usage}%, Peak Memory: ${peak_memory_usage}%, Peak Active Builds: $peak_active_builds"
        
        # Check if all builds are complete
        local total_completed=$((completed_builds + failed_builds))
        if [ "$active_builds" -eq 0 ] && [ "$total_completed" -ge "$MAX_CONCURRENT_BUILDS" ]; then
            test_pass "All stress builds completed (Success: $completed_builds, Failed: $failed_builds)"
            break
        fi
        
        # Check for timeout
        if [ "$elapsed_time" -gt "$BUILD_TIMEOUT" ]; then
            test_fail "Stress test timeout after ${BUILD_TIMEOUT}s"
            break
        fi
        
        # Check resource thresholds
        if [ "$peak_cpu_usage" -gt "$CPU_THRESHOLD" ]; then
            test_fail "CPU usage exceeded threshold: ${peak_cpu_usage}% > ${CPU_THRESHOLD}%"
        fi
        
        if [ "$peak_memory_usage" -gt "$MEMORY_THRESHOLD" ]; then
            test_fail "Memory usage exceeded threshold: ${peak_memory_usage}% > ${MEMORY_THRESHOLD}%"
        fi
        
        sleep $monitoring_interval
    done
}

# Generate stress test report
generate_stress_report() {
    echo "Generating stress test report..."
    
    cat > "${RESULTS_DIR}/stress_test_report.json" << EOF
{
    "test_configuration": {
        "max_concurrent_builds": $MAX_CONCURRENT_BUILDS,
        "build_timeout": $BUILD_TIMEOUT,
        "coordinator_port": $COORDINATOR_PORT,
        "cpu_threshold": $CPU_THRESHOLD,
        "memory_threshold": $MEMORY_THRESHOLD
    },
    "performance_metrics": {
        "peak_concurrent_builds": $peak_active_builds,
        "peak_cpu_usage": $peak_cpu_usage,
        "peak_memory_usage": $peak_memory_usage
    },
    "results": {
        "successful_submissions": $submission_count,
        "failed_submissions": $submission_failures,
        "success_rate": $(echo "scale=2; $submission_count * 100 / $MAX_CONCURRENT_BUILDS" | bc -l)
    }
}
EOF
    
    # Generate CSV summary if data is available
    if [ -f "${RESULTS_DIR}/submissions.csv" ]; then
        echo "timestamp,build_id,complexity" > "${RESULTS_DIR}/submissions_summary.csv"
        grep -v "^#" "${RESULTS_DIR}/submissions.csv" >> "${RESULTS_DIR}/submissions_summary.csv"
    fi
    
    test_pass "Stress test report generated"
}

# Cleanup stress test
cleanup_stress_test() {
    echo "Cleaning up stress test environment..."
    
    # Stop coordinator
    if [ -n "${COORDINATOR_PID:-}" ]; then
        kill $COORDINATOR_PID 2>/dev/null || true
        wait $COORDINATOR_PID 2>/dev/null || true
    fi
    
    # Stop resource monitoring
    if [ -n "${RESOURCE_MONITOR_PID:-}" ]; then
        kill $RESOURCE_MONITOR_PID 2>/dev/null || true
    fi
    
    # Clean up test projects
    rm -rf "${TEST_WORKSPACE}/${STRESS_PROJECT_NAME}"*
    
    # Clean up temporary monitoring script
    rm -f "${RESULTS_DIR}/monitor_resources.sh"
    
    test_pass "Stress test cleanup completed"
}

# Main stress test execution
main() {
    echo "Starting system stress test..."
    
    setup_stress_test || exit 1
    start_stress_coordinator || exit 1
    monitor_system_resources || exit 1
    submit_stress_load || exit 1
    monitor_stress_execution || exit 1
    generate_stress_report || exit 1
    cleanup_stress_test || exit 1
    
    echo "System stress test completed successfully!"
}

# Run tests if script is executed directly
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi