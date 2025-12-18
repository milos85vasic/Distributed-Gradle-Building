#!/bin/bash

# Distributed Gradle Building - Long Running Test
# Tests system stability for extended build sessions

set -euo pipefail

# Import test framework
source "$(dirname "$0")/../test_framework.sh"

# Configuration
LONGRUN_PROJECT_NAME="longrun-test"
TEST_DURATION=1800  # 30 minutes
BUILD_INTERVAL=60    # Submit new build every minute
COORDINATOR_PORT=${COORDINATOR_PORT:-8082}
RESULTS_DIR="${TEST_RESULTS_DIR}/longrun_test"
MAX_MEMORY_MB=2048
MAX_DISK_SPACE_MB=10240

# Test data setup
setup_longrun_test() {
    echo "Setting up long running test environment..."
    
    # Create test results directory
    mkdir -p "$RESULTS_DIR"
    
    # Create baseline project for repeated builds
    base_project="${TEST_WORKSPACE}/${LONGRUN_PROJECT_NAME}"
    mkdir -p "$base_project/src/main/java"
    
    # Create Java source with some computational work
    cat > "$base_project/src/main/java/LongRunTest.java" << EOF
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.util.Random;
import java.util.concurrent.TimeUnit;

public class LongRunTest {
    public static void main(String[] args) throws IOException, InterruptedException {
        System.out.println("Starting long running test at " + System.currentTimeMillis());
        
        // Simulate build work
        Random random = new Random();
        int workDuration = 5 + random.nextInt(10); // 5-15 seconds
        int dataSize = 1000 + random.nextInt(4000); // 1-5MB of data
        
        // Generate some temporary data
        File tempFile = new File("temp-data.txt");
        try (FileWriter writer = new FileWriter(tempFile)) {
            for (int i = 0; i < dataSize; i++) {
                writer.write("Data line " + i + ": " + random.nextDouble() + "\\n");
            }
        }
        
        // Simulate processing time
        for (int i = 0; i < workDuration * 100; i++) {
            double result = Math.sqrt(random.nextDouble() * 10000);
            if (i % 100 == 0) {
                System.out.print(".");
                Thread.sleep(10);
            }
        }
        
        // Clean up
        tempFile.delete();
        
        System.out.println("\\nLong running test completed successfully!");
        System.out.println("Processing time: " + workDuration + " seconds");
        System.out.println("Data processed: " + dataSize + " lines");
    }
}
EOF
    
    # Create build.gradle with extended tasks
    cat > "$base_project/build.gradle" << EOF
plugins {
    application
    java
}

application {
    mainClass = "LongRunTest"
}

java {
    sourceCompatibility = JavaVersion.VERSION_11
    targetCompatibility = JavaVersion.VERSION_11
}

// Add custom tasks to extend build time
task generateResources {
    doLast {
        def resourceDir = new File("\$buildDir/generated-resources")
        resourceDir.mkdirs()
        
        // Generate random resource files
        (1..10).each { i ->
            def resourceFile = new File(resourceDir, "resource-\${i}.properties")
            resourceFile.text = "timestamp=\${System.currentTimeMillis()}\\nrandom=\${Math.random()}"
        }
    }
}

// Ensure resources are generated before compilation
compileJava.dependsOn generateResources

sourceSets {
    main {
        resources {
            srcDir "\$buildDir/generated-resources"
        }
    }
}

tasks.withType(JavaCompile) {
    options.encoding = 'UTF-8'
    options.fork = true
    options.forkOptions.memoryMaximumSize = '1g'
}
EOF
    
    # Create Gradle wrapper properties
    mkdir -p "$base_project/gradle/wrapper"
    cat > "$base_project/gradle/wrapper/gradle-wrapper.properties" << EOF
distributionBase=GRADLE_USER_HOME
distributionPath=wrapper/dists
distributionUrl=https\\://services.gradle.org/distributions/gradle-7.6.1-bin.zip
zipStoreBase=GRADLE_USER_HOME
zipStorePath=wrapper/dists
EOF
    
    test_pass "Long running test environment setup completed"
}

# Start coordinator for long running test
start_longrun_coordinator() {
    echo "Starting coordinator for long running test..."
    
    cd "$TEST_WORKSPACE"
    export COORDINATOR_PORT="$COORDINATOR_PORT"
    export LOG_LEVEL="INFO"
    export BUILD_HISTORY_SIZE="1000"
    export CLEANUP_INTERVAL="300"  # Clean up every 5 minutes
    
    # Use Go coordinator
    if command -v "${PROJECT_DIR}/go/bin/coordinator" >/dev/null 2>&1; then
        "${PROJECT_DIR}/go/bin/coordinator" > "${RESULTS_DIR}/coordinator_longrun.log" 2>&1 &
        COORDINATOR_PID=$!
        sleep 3
        test_pass "Long run coordinator started (PID: $COORDINATOR_PID)"
    else
        test_fail "Coordinator binary not found"
        return 1
    fi
}

# Monitor system health during long run
monitor_system_health() {
    echo "Starting long-term system health monitoring..."
    
    # Create comprehensive monitoring script
    cat > "${RESULTS_DIR}/health_monitor.sh" << 'EOF'
#!/bin/bash
RESULTS_DIR="$1"
MONITOR_INTERVAL="$2"
MAX_MEMORY_MB="$3"
MAX_DISK_SPACE_MB="$4"

while true; do
    timestamp=$(date +%s)
    
    # CPU usage
    cpu_usage=$(top -l 1 -n 0 | grep "CPU usage" | awk '{print $3}' | sed 's/%//')
    
    # Memory usage
    memory_pressure=$(memory_pressure | grep "System-wide memory free percentage" | awk '{print $5}' | sed 's/%//')
    memory_used_mb=$(( (100 - memory_pressure) * 100 / 100 * 16384 / 100 ))  # Approximate for 16GB system
    
    # Disk usage for test workspace
    if [ -d "/tmp" ]; then
        disk_usage=$(df -m /tmp | awk 'NR==2 {print $3}')
        disk_available=$(df -m /tmp | awk 'NR==2 {print $4}')
    else
        disk_usage=0
        disk_available=10000
    fi
    
    # Network connections (simplified)
    network_connections=$(netstat -an 2>/dev/null | grep ":$COORDINATOR_PORT" | grep ESTABLISHED | wc -l || echo 0)
    
    # Process count
    java_processes=$(ps aux | grep -v grep | grep java | wc -l)
    gradle_processes=$(ps aux | grep -v grep | grep gradle | wc -l)
    
    # Log system health
    echo "$timestamp,$cpu_usage,$memory_used_mb,$disk_usage,$disk_available,$network_connections,$java_processes,$gradle_processes" >> "${RESULTS_DIR}/system_health.csv"
    
    # Check for resource exhaustion
    if [ "$memory_used_mb" -gt "$MAX_MEMORY_MB" ]; then
        echo "$(date): WARNING - Memory usage exceeded threshold: ${memory_used_mb}MB > ${MAX_MEMORY_MB}MB" >> "${RESULTS_DIR}/health_warnings.log"
    fi
    
    if [ "$disk_usage" -gt "$MAX_DISK_SPACE_MB" ]; then
        echo "$(date): WARNING - Disk usage exceeded threshold: ${disk_usage}MB > ${MAX_DISK_SPACE_MB}MB" >> "${RESULTS_DIR}/health_warnings.log"
    fi
    
    sleep "$MONITOR_INTERVAL"
done
EOF
    
    chmod +x "${RESULTS_DIR}/health_monitor.sh"
    
    # Start health monitoring
    "${RESULTS_DIR}/health_monitor.sh" "$RESULTS_DIR" "30" "$MAX_MEMORY_MB" "$MAX_DISK_SPACE_MB" &
    HEALTH_MONITOR_PID=$!
    
    test_pass "System health monitoring started (PID: $HEALTH_MONITOR_PID)"
}

# Submit builds at regular intervals
submit_periodic_builds() {
    echo "Starting periodic build submissions for ${TEST_DURATION}s..."
    
    local start_time=$(date +%s)
    local build_counter=0
    local successful_builds=0
    local failed_builds=0
    
    while true; do
        local current_time=$(date +%s)
        local elapsed_time=$((current_time - start_time))
        
        # Check if test duration reached
        if [ "$elapsed_time" -ge "$TEST_DURATION" ]; then
            echo "Test duration of ${TEST_DURATION}s reached"
            break
        fi
        
        # Submit new build
        build_counter=$((build_counter + 1))
        build_timestamp=$(date +%s%3N)
        
        echo "[$elapsed_time/${TEST_DURATION}s] Submitting build #$build_counter"
        
        # Create unique project directory for this build
        unique_project="${TEST_WORKSPACE}/${LONGRUN_PROJECT_NAME}-build-${build_counter}"
        cp -r "${TEST_WORKSPACE}/${LONGRUN_PROJECT_NAME}" "$unique_project"
        
        # Randomize the build slightly to ensure differences
        if [ $((build_counter % 3)) -eq 0 ]; then
            echo "test.value=$(date +%s)" >> "$unique_project/src/main/resources/test.properties"
        fi
        
        # Submit build request
        build_response=$(curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
            -H "Content-Type: application/json" \
            -d "{
                \"project_path\": \"$unique_project\",
                \"task_name\": \"build\",
                \"cache_enabled\": $(($build_counter % 2 == 0 ? "true" : "false")),
                \"priority\": \"normal\",
                \"timeout\": 300
            }")
        
        build_id=$(echo "$build_response" | jq -r '.build_id // empty')
        submission_time=$(date +%s%3N)
        
        if [ -n "$build_id" ]; then
            echo "${build_counter},${build_id},${submission_time},submitted" >> "${RESULTS_DIR}/build_lifecycle.csv"
            successful_builds=$((successful_builds + 1))
            echo "Build #$build_counter submitted successfully (ID: $build_id)"
        else
            echo "${build_counter},failed,${submission_time},submission_failed" >> "${RESULTS_DIR}/build_lifecycle.csv"
            failed_builds=$((failed_builds + 1))
            echo "Build #$build_counter submission failed: $build_response"
        fi
        
        # Wait for next submission interval
        sleep $BUILD_INTERVAL
    done
    
    echo "Periodic build submissions completed:"
    echo "- Total builds attempted: $build_counter"
    echo "- Successful submissions: $successful_builds"
    echo "- Failed submissions: $failed_builds"
    
    test_pass "Periodic build submissions completed"
}

# Monitor coordinator stability
monitor_coordinator_stability() {
    echo "Monitoring coordinator stability during long run..."
    
    local start_time=$(date +%s)
    local coordinator_restarts=0
    local last_pid="$COORDINATOR_PID"
    
    while true; do
        local current_time=$(date +%s)
        local elapsed_time=$((current_time - start_time))
        
        # Check if test duration reached
        if [ "$elapsed_time" -ge "$TEST_DURATION" ]; then
            break
        fi
        
        # Check if coordinator is still running
        if ! kill -0 "$COORDINATOR_PID" 2>/dev/null; then
            coordinator_restarts=$((coordinator_restarts + 1))
            echo "$(date): WARNING - Coordinator process died, restarting..." >> "${RESULTS_DIR}/stability_issues.log"
            
            # Restart coordinator
            cd "$TEST_WORKSPACE"
            export COORDINATOR_PORT="$COORDINATOR_PORT"
            export LOG_LEVEL="INFO"
            "${PROJECT_DIR}/go/bin/coordinator" >> "${RESULTS_DIR}/coordinator_longrun.log" 2>&1 &
            COORDINATOR_PID=$!
            sleep 5
        fi
        
        # Test coordinator API responsiveness
        api_response_time=$(curl -o /dev/null -s -w "%{time_total}" "http://localhost:$COORDINATOR_PORT/api/health" 2>/dev/null || echo "0")
        
        if [ "$api_response_time" = "0" ]; then
            echo "$(date): WARNING - Coordinator API not responding" >> "${RESULTS_DIR}/stability_issues.log"
        elif [ "$(echo "$api_response_time > 5.0" | bc -l 2>/dev/null || echo 0)" -eq 1 ]; then
            echo "$(date): WARNING - Coordinator API slow response: ${api_response_time}s" >> "${RESULTS_DIR}/stability_issues.log"
        fi
        
        # Log current status
        status_response=$(curl -s "http://localhost:$COORDINATOR_PORT/api/builds/status" 2>/dev/null || echo "{}")
        active_builds=$(echo "$status_response" | jq -r '.active_builds // 0')
        completed_builds=$(echo "$status_response" | jq -r '.completed_builds // 0')
        failed_builds=$(echo "$status_response" | jq -r '.failed_builds // 0')
        
        echo "[$elapsed_time/${TEST_DURATION}s] Status - Active: $active_builds, Completed: $completed_builds, Failed: $failed_builds, API: ${api_response_time}s"
        
        # Check for memory leaks in coordinator
        if [ -n "$COORDINATOR_PID" ]; then
            coordinator_memory=$(ps -p "$COORDINATOR_PID" -o rss= 2>/dev/null || echo 0)
            coordinator_memory_mb=$((coordinator_memory / 1024))
            
            echo "${current_time},${COORDINATOR_PID},${coordinator_memory_mb},${api_response_time},${active_builds}" >> "${RESULTS_DIR}/coordinator_health.csv"
        fi
        
        sleep 30  # Check every 30 seconds
    done
    
    echo "Coordinator stability monitoring completed"
    echo "- Coordinator restarts: $coordinator_restarts"
    echo "- Total test duration: ${TEST_DURATION}s"
    
    test_pass "Coordinator stability monitoring completed"
}

# Wait for all builds to complete
wait_for_completion() {
    echo "Waiting for all builds to complete..."
    
    local wait_start=$(date +%s)
    local max_wait=600  # Maximum 10 minutes after test end
    
    while true; do
        local current_time=$(date +%s)
        local elapsed_wait=$((current_time - wait_start))
        
        # Get current build status
        status_response=$(curl -s "http://localhost:$COORDINATOR_PORT/api/builds/status" 2>/dev/null || echo "{}")
        active_builds=$(echo "$status_response" | jq -r '.active_builds // 0')
        
        if [ "$active_builds" -eq 0 ]; then
            test_pass "All builds completed successfully"
            break
        fi
        
        if [ "$elapsed_wait" -gt "$max_wait" ]; then
            test_fail "Timeout waiting for builds to complete (${active_builds} still active)"
            break
        fi
        
        echo "Waiting for $active_builds active builds to complete... (${elapsed_wait}s elapsed)"
        sleep 10
    done
}

# Generate long running test report
generate_longrun_report() {
    echo "Generating long running test report..."
    
    # Count total builds from lifecycle log
    total_submissions=$(wc -l < "${RESULTS_DIR}/build_lifecycle.csv" 2>/dev/null || echo 0)
    
    # Analyze coordinator health if data available
    max_coordinator_memory=0
    avg_api_response=0
    if [ -f "${RESULTS_DIR}/coordinator_health.csv" ]; then
        max_coordinator_memory=$(awk -F',' 'NR>1 && $3>max {max=$3} END {print max}' "${RESULTS_DIR}/coordinator_health.csv" 2>/dev/null || echo 0)
        avg_api_response=$(awk -F',' 'NR>1 {sum+=$4; count++} END {if(count>0) print sum/count; else print 0}' "${RESULTS_DIR}/coordinator_health.csv" 2>/dev/null || echo 0)
    fi
    
    # Count stability issues
    stability_issues=$(wc -l < "${RESULTS_DIR}/stability_issues.log" 2>/dev/null || echo 0)
    health_warnings=$(wc -l < "${RESULTS_DIR}/health_warnings.log" 2>/dev/null || echo 0)
    
    cat > "${RESULTS_DIR}/longrun_test_report.json" << EOF
{
    "test_configuration": {
        "test_duration": $TEST_DURATION,
        "build_interval": $BUILD_INTERVAL,
        "coordinator_port": $COORDINATOR_PORT,
        "max_memory_mb": $MAX_MEMORY_MB,
        "max_disk_space_mb": $MAX_DISK_SPACE_MB
    },
    "execution_summary": {
        "total_build_submissions": $total_submissions,
        "coordinator_restarts": $coordinator_restarts,
        "test_completed_at": "$(date -Iseconds)"
    },
    "performance_metrics": {
        "max_coordinator_memory_mb": $max_coordinator_memory,
        "avg_api_response_time": $avg_api_response
    },
    "stability_metrics": {
        "stability_issues": $stability_issues,
        "health_warnings": $health_warnings
    }
}
EOF
    
    test_pass "Long running test report generated"
}

# Cleanup long running test
cleanup_longrun_test() {
    echo "Cleaning up long running test environment..."
    
    # Stop coordinator
    if [ -n "${COORDINATOR_PID:-}" ]; then
        kill $COORDINATOR_PID 2>/dev/null || true
        wait $COORDINATOR_PID 2>/dev/null || true
    fi
    
    # Stop health monitoring
    if [ -n "${HEALTH_MONITOR_PID:-}" ]; then
        kill $HEALTH_MONITOR_PID 2>/dev/null || true
    fi
    
    # Clean up test projects
    rm -rf "${TEST_WORKSPACE}/${LONGRUN_PROJECT_NAME}"*
    
    # Clean up monitoring script
    rm -f "${RESULTS_DIR}/health_monitor.sh"
    
    test_pass "Long running test cleanup completed"
}

# Main test execution
main() {
    echo "Starting long running test (${TEST_DURATION}s)..."
    
    setup_longrun_test || exit 1
    start_longrun_coordinator || exit 1
    monitor_system_health || exit 1
    submit_periodic_builds || exit 1
    monitor_coordinator_stability || exit 1
    wait_for_completion || exit 1
    generate_longrun_report || exit 1
    cleanup_longrun_test || exit 1
    
    echo "Long running test completed successfully!"
}

# Run tests if script is executed directly
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi