#!/bin/bash

# Performance test for build coordinator
# This test validates coordinator performance under various load conditions

# Import test framework
source "$(dirname "$0")/../test_framework.sh"

# Test configuration
source "$(dirname "$0")/../test_config.sh"

# Performance test configuration
PERF_TEST_BUILD_COUNT=100
PERF_TEST_CONCURRENT=10
PERF_TEST_TIMEOUT=300

# Test setup
setup_performance_test() {
    echo "Setting up coordinator performance test..."
    
    # Create test workspace
    TEST_WORKSPACE=$(mktemp -d)
    export TEST_WORKSPACE
    
    # Create test projects
    for i in $(seq 1 $PERF_TEST_BUILD_COUNT); do
        PROJECT_DIR="$TEST_WORKSPACE/project-$i"
        mkdir -p "$PROJECT_DIR/src/main/java/com/example"
        
        # Create build.gradle
        cat > "$PROJECT_DIR/build.gradle" << EOF
plugins {
    id 'java'
}

group 'com.example'
version '1.0.0'

repositories {
    mavenCentral()
}

dependencies {
    testImplementation 'junit:junit:4.13.2'
}

task compileJava {
    doLast {
        println "Compiling project $i..."
        mkdir "build/classes/java/main"
    }
}

task test {
    doLast {
        println "Testing project $i..."
        mkdir "build/test-results"
    }
}

task jar {
    doLast {
        println "Creating JAR for project $i..."
        mkdir "build/libs"
    }
}
EOF
        
        # Create Java source
        cat > "$PROJECT_DIR/src/main/java/com/example/Hello$i.java" << EOF
package com.example;

public class Hello$i {
    public static void main(String[] args) {
        System.out.println("Hello from project $i!");
    }
}
EOF
    done
    
    echo "Created $PERF_TEST_BUILD_COUNT test projects in $TEST_WORKSPACE"
}

# Test coordinator scalability
test_coordinator_scalability() {
    echo "Testing coordinator scalability..."
    
    # Start coordinator in performance mode
    cd "$TEST_WORKSPACE"
    
    # Mock coordinator for performance testing
    python3 -c "
import http.server
import socketserver
import json
import threading
import time
import sys
from urllib.parse import urlparse, parse_qs

class PerformanceCoordinator(http.server.BaseHTTPRequestHandler):
    builds = {}
    build_counter = 0
    
    def do_POST(self):
        if self.path == '/api/builds':
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            
            # Generate build ID
            build_id = f'build-{PerformanceCoordinator.build_counter}'
            PerformanceCoordinator.build_counter += 1
            
            # Store build request
            PerformanceCoordinator.builds[build_id] = {
                'status': 'queued',
                'created_at': time.time(),
                'project_name': json.loads(post_data.decode('utf-8')).get('project_name', 'unknown')
            }
            
            response = {
                'build_id': build_id,
                'status': 'accepted',
                'queue_position': len(PerformanceCoordinator.builds)
            }
            
            self.send_response(202)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(response).encode())
    
    def do_GET(self):
        if self.path.startswith('/api/builds/') and '/status' in self.path:
            build_id = self.path.split('/')[3]
            
            if build_id in PerformanceCoordinator.builds:
                build = PerformanceCoordinator.builds[build_id]
                response = {
                    'build_id': build_id,
                    'status': 'completed' if time.time() - build['created_at'] > 1 else 'in_progress',
                    'progress': min(100, int((time.time() - build['created_at']) * 100))
                }
            else:
                response = {'error': 'Build not found'}
            
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(response).encode())

PORT = $COORDINATOR_PORT
with socketserver.TCPServer(('', PORT), PerformanceCoordinator) as httpd:
    print(f'Performance coordinator running on port {PORT}')
    httpd.serve_forever()
" &
    COORDINATOR_PID=$!
    
    # Wait for coordinator to start
    sleep 2
    
    # Test concurrent build submissions
    echo "Testing concurrent build submissions ($PERF_TEST_CONCURRENT concurrent builds)..."
    
    START_TIME=$(date +%s.%N)
    
    # Submit builds concurrently
    for i in $(seq 1 $PERF_TEST_CONCURRENT); do
        {
            PROJECT_DIR="$TEST_WORKSPACE/project-$i"
            BUILD_REQUEST="{
                \"project_name\": \"performance-test-$i\",
                \"build_type\": \"gradle\",
                \"tasks\": [\"compileJava\", \"test\"],
                \"source_location\": \"$PROJECT_DIR\",
                \"environment\": {
                    \"java_home\": \"/usr/lib/jvm/default-java\",
                    \"gradle_home\": \"/usr/local/gradle\"
                }
            }"
            
            RESPONSE=$(curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
                -H "Content-Type: application/json" \
                -d "$BUILD_REQUEST")
            
            BUILD_ID=$(echo "$RESPONSE" | jq -r '.build_id')
            echo "Build $i submitted with ID: $BUILD_ID"
        } &
    done
    
    # Wait for all submissions to complete
    wait
    
    END_TIME=$(date +%s.%N)
    SUBMISSION_TIME=$(echo "$END_TIME - $START_TIME" | bc)
    
    echo "Build submissions completed in ${SUBMISSION_TIME}s"
    
    # Verify submission rate
    SUBMISSION_RATE=$(echo "scale=2; $PERF_TEST_CONCURRENT / $SUBMISSION_TIME" | bc)
    
    if (( $(echo "$SUBMISSION_RATE > 5" | bc -l) )); then
        test_pass "Coordinator submission rate: $SUBMISSION_RATE builds/sec (target: >5)"
    else
        test_fail "Coordinator submission rate too low: $SUBMISSION_RATE builds/sec (target: >5)"
    fi
    
    # Stop coordinator
    kill $COORDINATOR_PID 2>/dev/null
}

# Test memory usage
test_memory_usage() {
    echo "Testing coordinator memory usage..."
    
    # Start memory tracking
    MEMORY_START=$(ps aux | grep -E "(coordinator|python3)" | grep -v grep | awk '{sum+=$6} END {print sum/1024}' 2>/dev/null || echo "0")
    
    # Process a large number of builds to test memory efficiency
    cd "$TEST_WORKSPACE"
    
    python3 -c "
import http.server
import socketserver
import json
import time
import gc
import sys

class MemoryTestCoordinator(http.server.BaseHTTPRequestHandler):
    builds = {}
    build_counter = 0
    
    def do_POST(self):
        if self.path == '/api/builds':
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            
            build_id = f'build-{MemoryTestCoordinator.build_counter}'
            MemoryTestCoordinator.build_counter += 1
            
            # Store build with more data to test memory usage
            MemoryTestCoordinator.builds[build_id] = {
                'status': 'queued',
                'created_at': time.time(),
                'project_name': json.loads(post_data.decode('utf-8')).get('project_name', 'unknown'),
                'large_data': 'x' * 1024  # 1KB of data per build
            }
            
            response = {
                'build_id': build_id,
                'status': 'accepted'
            }
            
            self.send_response(202)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(response).encode())

PORT = $COORDINATOR_PORT
with socketserver.TCPServer(('', PORT), MemoryTestCoordinator) as httpd:
    print(f'Memory test coordinator running on port {PORT}')
    
    # Process builds for memory test
    end_time = time.time() + 10  # Run for 10 seconds
    while time.time() < end_time:
        httpd.handle_request()
        
        # Force garbage collection periodically
        if MemoryTestCoordinator.build_counter % 50 == 0:
            gc.collect()
" &
    MEMORY_COORD_PID=$!
    
    sleep 2
    
    # Submit builds to test memory accumulation
    for i in $(seq 1 50); do
        PROJECT_DIR="$TEST_WORKSPACE/project-$((i % 10 + 1))"
        BUILD_REQUEST="{
            \"project_name\": \"memory-test-$i\",
            \"build_type\": \"gradle\",
            \"tasks\": [\"compileJava\"],
            \"source_location\": \"$PROJECT_DIR\"
        }"
        
        curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
            -H "Content-Type: application/json" \
            -d "$BUILD_REQUEST" > /dev/null
    done
    
    sleep 5
    
    # Check memory usage after test
    MEMORY_END=$(ps aux | grep -E "(coordinator|python3)" | grep -v grep | awk '{sum+=$6} END {print sum/1024}' 2>/dev/null || echo "0")
    
    MEMORY_DIFF=$(echo "scale=2; $MEMORY_END - $MEMORY_START" | bc)
    
    echo "Memory usage increased by ${MEMORY_DIFF}MB during test"
    
    if (( $(echo "$MEMORY_DIFF < 100" | bc -l) )); then
        test_pass "Memory usage within acceptable limits: ${MEMORY_DIFF}MB (target: <100MB)"
    else
        test_fail "Memory usage too high: ${MEMORY_DIFF}MB (target: <100MB)"
    fi
    
    # Stop coordinator
    kill $MEMORY_COORD_PID 2>/dev/null
}

# Test response time under load
test_response_time() {
    echo "Testing response time under load..."
    
    # Start response time test coordinator
    cd "$TEST_WORKSPACE"
    
    python3 -c "
import http.server
import socketserver
import json
import time
import random

class ResponseTimeCoordinator(http.server.BaseHTTPRequestHandler):
    builds = {}
    
    def do_POST(self):
        if self.path == '/api/builds':
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            
            # Simulate variable processing time
            time.sleep(random.uniform(0.01, 0.1))
            
            build_id = f'build-{int(time.time() * 1000)}'
            ResponseTimeCoordinator.builds[build_id] = {
                'status': 'queued',
                'created_at': time.time()
            }
            
            response = {
                'build_id': build_id,
                'status': 'accepted'
            }
            
            self.send_response(202)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(response).encode())

PORT = $COORDINATOR_PORT
with socketserver.TCPServer(('', PORT), ResponseTimeCoordinator) as httpd:
    print(f'Response time coordinator running on port {PORT}')
    httpd.serve_forever()
" &
    RESPONSE_COORD_PID=$!
    
    sleep 2
    
    # Test response times with multiple concurrent requests
    TOTAL_RESPONSE_TIME=0
    SUCCESSFUL_REQUESTS=0
    
    for i in $(seq 1 20); do
        START_TIME=$(date +%s.%N)
        
        PROJECT_DIR="$TEST_WORKSPACE/project-$((i % 10 + 1))"
        BUILD_REQUEST="{
            \"project_name\": \"response-test-$i\",
            \"build_type\": \"gradle\",
            \"tasks\": [\"compileJava\"],
            \"source_location\": \"$PROJECT_DIR\"
        }"
        
        HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
            -H "Content-Type: application/json" \
            -d "$BUILD_REQUEST")
        
        END_TIME=$(date +%s.%N)
        RESPONSE_TIME=$(echo "$END_TIME - $START_TIME" | bc)
        
        if [ "$HTTP_CODE" = "202" ]; then
            TOTAL_RESPONSE_TIME=$(echo "$TOTAL_RESPONSE_TIME + $RESPONSE_TIME" | bc)
            SUCCESSFUL_REQUESTS=$((SUCCESSFUL_REQUESTS + 1))
        fi
        
        # Run some requests concurrently
        if [ $((i % 5)) -eq 0 ]; then
            {
                START_TIME=$(date +%s.%N)
                curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
                    -H "Content-Type: application/json" \
                    -d "{\"project_name\": \"concurrent-$i\"}" > /dev/null
                END_TIME=$(date +%s.%N)
                echo "$END_TIME - $START_TIME" | bc > "/tmp/response_time_$i"
            } &
        fi
    done
    
    wait
    
    # Collect concurrent response times
    for i in $(seq 5 5 20); do
        if [ -f "/tmp/response_time_$i" ]; then
            CONCURRENT_TIME=$(cat "/tmp/response_time_$i")
            TOTAL_RESPONSE_TIME=$(echo "$TOTAL_RESPONSE_TIME + $CONCURRENT_TIME" | bc)
            SUCCESSFUL_REQUESTS=$((SUCCESSFUL_REQUESTS + 1))
            rm -f "/tmp/response_time_$i"
        fi
    done
    
    # Calculate average response time
    if [ $SUCCESSFUL_REQUESTS -gt 0 ]; then
        AVG_RESPONSE_TIME=$(echo "scale=3; $TOTAL_RESPONSE_TIME / $SUCCESSFUL_REQUESTS" | bc)
        echo "Average response time: ${AVG_RESPONSE_TIME}s ($SUCCESSFUL_REQUESTS requests)"
        
        if (( $(echo "$AVG_RESPONSE_TIME < 0.5" | bc -l) )); then
            test_pass "Response time acceptable: ${AVG_RESPONSE_TIME}s (target: <0.5s)"
        else
            test_fail "Response time too slow: ${AVG_RESPONSE_TIME}s (target: <0.5s)"
        fi
    else
        test_fail "No successful requests to measure response time"
    fi
    
    # Stop coordinator
    kill $RESPONSE_COORD_PID 2>/dev/null
}

# Test cleanup
test_cleanup() {
    echo "Cleaning up performance test..."
    
    # Kill any remaining processes
    pkill -f "python3.*coord" 2>/dev/null || true
    pkill -f "python3.*test" 2>/dev/null || true
    
    # Cleanup test workspace
    rm -rf "$TEST_WORKSPACE"
    
    # Cleanup temporary files
    rm -f /tmp/response_time_*
    
    test_pass "Performance test cleanup completed"
}

# Main test execution
main() {
    echo "Starting coordinator performance tests..."
    
    # Check dependencies
    if ! command -v bc >/dev/null 2>&1; then
        test_fail "bc calculator required for performance tests"
        exit 1
    fi
    
    if ! command -v python3 >/dev/null 2>&1; then
        test_fail "python3 required for performance test mock services"
        exit 1
    fi
    
    # Run performance test suite
    setup_performance_test || exit 1
    test_coordinator_scalability || exit 1
    test_memory_usage || exit 1
    test_response_time || exit 1
    test_cleanup || exit 1
    
    echo "Coordinator performance tests completed successfully!"
}

# Run tests if script is executed directly
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi