#!/bin/bash

# Distributed Gradle Building - Cache Performance Test
# Tests cache server performance under various scenarios

set -euo pipefail

# Import test framework
source "$(dirname "$0")/../test_framework.sh"

# Configuration
CACHE_TEST_PROJECT="cache-test"
CACHE_TYPES=("local" "distributed" "hybrid")
CACHE_SIZES=("small" "medium" "large")
COORDINATOR_PORT=${COORDINATOR_PORT:-8082}
CACHE_SERVER_PORT=${CACHE_SERVER_PORT:-8083}
RESULTS_DIR="${TEST_RESULTS_DIR}/cache_performance"
TEST_ITERATIONS=5

# Test data setup
setup_cache_test() {
    echo "Setting up cache performance test environment..."
    
    # Create test results directory
    mkdir -p "$RESULTS_DIR"
    
    # Create test projects with different sizes
    for size in "${CACHE_SIZES[@]}"; do
        project_dir="${TEST_WORKSPACE}/${CACHE_TEST_PROJECT}-${size}"
        mkdir -p "$project_dir/src/main/java"
        mkdir -p "$project_dir/src/main/resources"
        
        case $size in
            "small")
                # Small project with minimal dependencies
                cat > "$project_dir/src/main/java/SmallApp.java" << EOF
public class SmallApp {
    public static void main(String[] args) {
        System.out.println("Small cache test application");
    }
}
EOF
                cat > "$project_dir/build.gradle" << EOF
plugins {
    application
    java
}

application {
    mainClass = 'SmallApp'
}

java {
    sourceCompatibility = JavaVersion.VERSION_11
    targetCompatibility = JavaVersion.VERSION_11
}
EOF
                ;;
            "medium")
                # Medium project with multiple dependencies
                cat > "$project_dir/src/main/java/MediumApp.java" << EOF
import java.util.ArrayList;
import java.util.List;
import java.io.File;

public class MediumApp {
    public static void main(String[] args) {
        System.out.println("Medium cache test application");
        
        List<String> data = new ArrayList<>();
        for (int i = 0; i < 1000; i++) {
            data.add("Item " + i);
        }
        
        // Create some resource files during build
        File resource = new File("generated-resource.txt");
        try {
            if (resource.createNewFile()) {
                System.out.println("Generated resource file");
            }
        } catch (Exception e) {
            e.printStackTrace();
        }
    }
}
EOF
                cat > "$project_dir/build.gradle" << EOF
plugins {
    application
    java
}

repositories {
    mavenCentral()
}

dependencies {
    implementation 'org.apache.commons:commons-lang3:3.12.0'
    implementation 'com.google.guava:guava:31.1-jre'
}

application {
    mainClass = 'MediumApp'
}

java {
    sourceCompatibility = JavaVersion.VERSION_11
    targetCompatibility = JavaVersion.VERSION_11
}

task generateResources {
    doLast {
        def resourceDir = new File("\$buildDir/generated-resources")
        resourceDir.mkdirs()
        new File(resourceDir, "generated.properties").text = "timestamp=\${System.currentTimeMillis()}"
    }
}

compileJava.dependsOn generateResources

sourceSets {
    main {
        resources {
            srcDir "\$buildDir/generated-resources"
        }
    }
}
EOF
                ;;
            "large")
                # Large project with extensive dependencies and resources
                cat > "$project_dir/src/main/java/LargeApp.java" << EOF
import java.util.*;
import java.util.concurrent.*;
import java.io.*;
import java.nio.file.*;

public class LargeApp {
    public static void main(String[] args) throws Exception {
        System.out.println("Large cache test application");
        
        // Simulate complex processing
        ExecutorService executor = Executors.newFixedThreadPool(4);
        List<Future<?>> futures = new ArrayList<>();
        
        for (int i = 0; i < 100; i++) {
            final int taskId = i;
            futures.add(executor.submit(() -> {
                List<Integer> data = new ArrayList<>();
                Random random = new Random(taskId);
                
                for (int j = 0; j < 10000; j++) {
                    data.add(random.nextInt());
                }
                
                Collections.sort(data);
                return data.size();
            }));
        }
        
        // Wait for all tasks to complete
        for (Future<?> future : futures) {
            future.get();
        }
        
        executor.shutdown();
        
        // Generate large resource files
        Path resourceDir = Paths.get("large-resources");
        Files.createDirectories(resourceDir);
        
        for (int i = 0; i < 10; i++) {
            Path resourceFile = resourceDir.resolve("resource-" + i + ".txt");
            List<String> content = new ArrayList<>();
            for (int j = 0; j < 1000; j++) {
                content.add("Resource line " + j + " for file " + i);
            }
            Files.write(resourceFile, content);
        }
        
        System.out.println("Large application processing completed");
    }
}
EOF
                cat > "$project_dir/build.gradle" << EOF
plugins {
    application
    java
}

repositories {
    mavenCentral()
}

dependencies {
    implementation 'org.apache.commons:commons-lang3:3.12.0'
    implementation 'com.google.guava:guava:31.1-jre'
    implementation 'org.apache.httpcomponents:httpclient:4.5.13'
    implementation 'org.springframework:spring-context:5.3.20'
    implementation 'com.fasterxml.jackson.core:jackson-databind:2.13.3'
    implementation 'org.slf4j:slf4j-api:1.7.36'
    implementation 'ch.qos.logback:logback-classic:1.2.11'
    testImplementation 'junit:junit:4.13.2'
}

application {
    mainClass = 'LargeApp'
}

java {
    sourceCompatibility = JavaVersion.VERSION_11
    targetCompatibility = JavaVersion.VERSION_11
}

task generateLargeResources {
    doLast {
        def resourceDir = new File("\$buildDir/large-resources")
        resourceDir.mkdirs()
        
        // Generate multiple resource files
        (1..20).each { i ->
            def resourceFile = new File(resourceDir, "large-resource-\${i}.properties")
            def content = new StringBuilder()
            content.append("timestamp=\${System.currentTimeMillis()}\\n")
            content.append("resource-id=\${i}\\n")
            content.append("random-data=\${UUID.randomUUID().toString()}\\n")
            
            // Add more content for larger files
            (1..1000).each { j ->
                content.append("line-\${j}=\${Math.random()}\\n")
            }
            
            resourceFile.text = content.toString()
        }
    }
}

processResources.dependsOn generateLargeResources

sourceSets {
    main {
        resources {
            srcDir "\$buildDir/large-resources"
        }
    }
}

tasks.withType(JavaCompile) {
    options.encoding = 'UTF-8'
    options.fork = true
    options.forkOptions.memoryMaximumSize = '2g'
}
EOF
                ;;
        esac
        
        echo "Created cache test project: $project_dir (size: $size)"
    done
    
    test_pass "Cache test environment setup completed"
}

# Start cache server for testing
start_cache_server() {
    local cache_type="$1"
    echo "Starting $cache_type cache server..."
    
    cd "$TEST_WORKSPACE"
    export CACHE_SERVER_PORT="$CACHE_SERVER_PORT"
    export CACHE_TYPE="$cache_type"
    export LOG_LEVEL="INFO"
    export CACHE_SIZE="1024"  # MB
    
    # Use Go cache server
    if command -v "${PROJECT_DIR}/go/bin/cache_server" >/dev/null 2>&1; then
        "${PROJECT_DIR}/go/bin/cache_server" > "${RESULTS_DIR}/cache_server_${cache_type}.log" 2>&1 &
        CACHE_SERVER_PID=$!
        sleep 2
        
        # Verify cache server is running
        if curl -s "http://localhost:$CACHE_SERVER_PORT/api/health" >/dev/null; then
            test_pass "$cache_type cache server started (PID: $CACHE_SERVER_PID)"
        else
            test_fail "$cache_type cache server failed to start"
            return 1
        fi
    else
        test_fail "Cache server binary not found"
        return 1
    fi
}

# Test cache performance
test_cache_performance() {
    local cache_type="$1"
    local project_size="$2"
    echo "Testing $cache_type cache performance with $project_size project..."
    
    local project_dir="${TEST_WORKSPACE}/${CACHE_TEST_PROJECT}-${project_size}"
    local total_build_time=0
    local min_build_time=999999
    local max_build_time=0
    local cache_hits=0
    local cache_misses=0
    
    # Test multiple iterations
    for iteration in $(seq 1 $TEST_ITERATIONS); do
        echo "  Iteration $iteration of $TEST_ITERATIONS..."
        
        # Clear cache for first iteration only
        if [ "$iteration" -eq 1 ]; then
            echo "  Clearing cache for cold start test..."
            curl -s -X DELETE "http://localhost:$CACHE_SERVER_PORT/api/cache/clear" >/dev/null || true
            sleep 1
        fi
        
        # Submit build request
        local build_start=$(date +%s%3N)
        
        build_response=$(curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
            -H "Content-Type: application/json" \
            -d "{
                \"project_path\": \"$project_dir\",
                \"task_name\": \"build\",
                \"cache_enabled\": true,
                \"cache_type\": \"$cache_type\"
            }")
        
        build_id=$(echo "$build_response" | jq -r '.build_id // empty')
        
        if [ -n "$build_id" ]; then
            # Wait for build completion
            while true; do
                status_response=$(curl -s "http://localhost:$COORDINATOR_PORT/api/builds/$build_id/status")
                build_status=$(echo "$status_response" | jq -r '.status // "unknown"')
                
                if [ "$build_status" = "completed" ]; then
                    build_end=$(date +%s%3N)
                    build_time=$((build_end - build_start))
                    
                    # Get cache statistics
                    cache_stats=$(curl -s "http://localhost:$CACHE_SERVER_PORT/api/stats")
                    iteration_hits=$(echo "$cache_stats" | jq -r '.cache_hits // 0')
                    iteration_misses=$(echo "$cache_stats" | jq -r '.cache_misses // 0')
                    
                    # Update metrics
                    total_build_time=$((total_build_time + build_time))
                    if [ "$build_time" -lt "$min_build_time" ]; then
                        min_build_time=$build_time
                    fi
                    if [ "$build_time" -gt "$max_build_time" ]; then
                        max_build_time=$build_time
                    fi
                    
                    cache_hits=$((cache_hits + iteration_hits))
                    cache_misses=$((cache_misses + iteration_misses))
                    
                    echo "    Build completed in ${build_time}ms (Cache hits: $iteration_hits, misses: $iteration_misses)"
                    break
                elif [ "$build_status" = "failed" ]; then
                    test_fail "Build $build_id failed"
                    break
                fi
                
                sleep 2
            done
        else
            test_fail "Build submission failed: $build_response"
        fi
    done
    
    # Calculate average and performance metrics
    local avg_build_time=$((total_build_time / TEST_ITERATIONS))
    local cache_hit_rate=0
    local total_cache_requests=$((cache_hits + cache_misses))
    if [ "$total_cache_requests" -gt 0 ]; then
        cache_hit_rate=$((cache_hits * 100 / total_cache_requests))
    fi
    
    # Store results
    echo "${cache_type},${project_size},${avg_build_time},${min_build_time},${max_build_time},${cache_hits},${cache_misses},${cache_hit_rate}" >> "${RESULTS_DIR}/cache_performance.csv"
    
    test_pass "$cache_type cache test with $project_size project completed"
    test_pass "  Average build time: ${avg_build_time}ms, Cache hit rate: ${cache_hit_rate}%"
}

# Test cache capacity and eviction
test_cache_capacity() {
    local cache_type="$1"
    echo "Testing $cache_type cache capacity and eviction..."
    
    # Create multiple builds to fill cache
    for i in $(seq 1 20); do
        project_dir="${TEST_WORKSPACE}/${CACHE_TEST_PROJECT}-medium"
        
        # Modify project slightly to create different cache entries
        sed "s/MediumApp/MediumApp${i}/" "$project_dir/src/main/java/MediumApp.java" > "${project_dir}/src/main/java/MediumApp${i}.java"
        sed -i.bak "s/mainClass = 'MediumApp'/mainClass = 'MediumApp${i}'/" "$project_dir/build.gradle"
        
        # Submit build
        build_response=$(curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
            -H "Content-Type: application/json" \
            -d "{
                \"project_path\": \"$project_dir\",
                \"task_name\": \"build\",
                \"cache_enabled\": true,
                \"cache_type\": \"$cache_type\"
            }")
        
        build_id=$(echo "$build_response" | jq -r '.build_id // empty')
        
        if [ -n "$build_id" ]; then
            echo "  Submitted capacity test build $i (ID: $build_id)"
        fi
        
        # Restore original files
        mv "${project_dir}/build.gradle.bak" "$project_dir/build.gradle"
        rm -f "${project_dir}/src/main/java/MediumApp${i}.java"
    done
    
    # Wait for builds to complete
    sleep 30
    
    # Check cache statistics
    cache_stats=$(curl -s "http://localhost:$CACHE_SERVER_PORT/api/stats")
    cache_size=$(echo "$cache_stats" | jq -r '.cache_size_mb // 0')
    cache_entries=$(echo "$cache_stats" | jq -r '.cache_entries // 0')
    evictions=$(echo "$cache_stats" | jq -r '.evictions // 0')
    
    echo "${cache_type},${cache_size},${cache_entries},${evictions}" >> "${RESULTS_DIR}/cache_capacity.csv"
    
    test_pass "$cache_type cache capacity test completed"
    test_pass "  Cache size: ${cache_size}MB, Entries: $cache_entries, Evictions: $evictions"
}

# Test cache concurrency
test_cache_concurrency() {
    local cache_type="$1"
    echo "Testing $cache_type cache concurrency..."
    
    # Submit multiple builds concurrently
    local concurrent_builds=5
    local build_pids=()
    
    for i in $(seq 1 $concurrent_builds); do
        {
            project_dir="${TEST_WORKSPACE}/${CACHE_TEST_PROJECT}-medium"
            
            build_response=$(curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
                -H "Content-Type: application/json" \
                -d "{
                    \"project_path\": \"$project_dir\",
                    \"task_name\": \"build\",
                    \"cache_enabled\": true,
                    \"cache_type\": \"$cache_type\"
                }")
            
            build_id=$(echo "$build_response" | jq -r '.build_id // empty')
            echo "$build_id" >> "${RESULTS_DIR}/concurrent_builds_$cache_type.txt"
        } &
        
        build_pids+=($!)
    done
    
    # Wait for all submissions
    for pid in "${build_pids[@]}"; do
        wait $pid
    done
    
    # Wait for builds to complete
    sleep 60
    
    # Check cache statistics for concurrent access patterns
    cache_stats=$(curl -s "http://localhost:$CACHE_SERVER_PORT/api/stats")
    concurrent_hits=$(echo "$cache_stats" | jq -r '.concurrent_hits // 0')
    concurrent_misses=$(echo "$cache_stats" | jq -r '.concurrent_misses // 0')
    
    echo "${cache_type},${concurrent_hits},${concurrent_misses}" >> "${RESULTS_DIR}/cache_concurrency.csv"
    
    test_pass "$cache_type cache concurrency test completed"
    test_pass "  Concurrent cache hits: $concurrent_hits, misses: $concurrent_misses"
}

# Generate cache performance report
generate_cache_report() {
    echo "Generating cache performance report..."
    
    # Initialize CSV files
    echo "cache_type,project_size,avg_build_time_ms,min_build_time_ms,max_build_time_ms,cache_hits,cache_misses,cache_hit_rate_percent" > "${RESULTS_DIR}/cache_performance.csv"
    echo "cache_type,cache_size_mb,cache_entries,evictions" > "${RESULTS_DIR}/cache_capacity.csv"
    echo "cache_type,concurrent_hits,concurrent_misses" > "${RESULTS_DIR}/cache_concurrency.csv"
    
    # Generate JSON summary
    cat > "${RESULTS_DIR}/cache_performance_report.json" << EOF
{
    "test_configuration": {
        "cache_types": [$(printf '"%s",' "${CACHE_TYPES[@]}" | sed 's/,$//')],
        "project_sizes": [$(printf '"%s",' "${CACHE_SIZES[@]}" | sed 's/,$//')],
        "test_iterations": $TEST_ITERATIONS,
        "coordinator_port": $COORDINATOR_PORT,
        "cache_server_port": $CACHE_SERVER_PORT
    },
    "results": {
        "performance_data_file": "cache_performance.csv",
        "capacity_data_file": "cache_capacity.csv",
        "concurrency_data_file": "cache_concurrency.csv"
    },
    "test_completed_at": "$(date -Iseconds)"
}
EOF
    
    test_pass "Cache performance report generated"
}

# Cleanup cache test
cleanup_cache_test() {
    echo "Cleaning up cache test environment..."
    
    # Stop cache server
    if [ -n "${CACHE_SERVER_PID:-}" ]; then
        kill $CACHE_SERVER_PID 2>/dev/null || true
        wait $CACHE_SERVER_PID 2>/dev/null || true
    fi
    
    # Clean up test projects
    rm -rf "${TEST_WORKSPACE}/${CACHE_TEST_PROJECT}"*
    
    # Clean up temporary files
    rm -f "${RESULTS_DIR}/concurrent_builds_"*.txt
    
    test_pass "Cache test cleanup completed"
}

# Main test execution
main() {
    echo "Starting cache performance tests..."
    
    # Initialize results
    setup_cache_test || exit 1
    generate_cache_report || exit 1
    
    # Test each cache type
    for cache_type in "${CACHE_TYPES[@]}"; do
        echo "Testing $cache_type cache configuration..."
        
        # Start cache server
        start_cache_server "$cache_type" || exit 1
        
        # Test different project sizes
        for project_size in "${CACHE_SIZES[@]}"; do
            test_cache_performance "$cache_type" "$project_size" || exit 1
        done
        
        # Test cache capacity
        test_cache_capacity "$cache_type" || exit 1
        
        # Test cache concurrency
        test_cache_concurrency "$cache_type" || exit 1
        
        # Cleanup for next cache type
        cleanup_cache_test
    done
    
    echo "Cache performance tests completed successfully!"
}

# Run tests if script is executed directly
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi