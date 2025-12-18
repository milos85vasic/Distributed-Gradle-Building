#!/bin/bash

# Load testing for distributed build system
# This test validates system behavior under high load conditions

# Import test framework
source "$(dirname "$0")/../test_framework.sh"

# Test configuration
source "$(dirname "$0")/../test_config.sh"

# Load test configuration
LOAD_TEST_BUILD_COUNT=1000
LOAD_TEST_CONCURRENT_BUILDS=50
LOAD_TEST_DURATION=300
LOAD_TEST_RAMP_UP=30

# Test setup
setup_load_test() {
    echo "Setting up load test environment..."
    
    # Create test workspace
    TEST_WORKSPACE=$(mktemp -d)
    export TEST_WORKSPACE
    
    # Create a variety of project sizes for realistic load testing
    for i in $(seq 1 100); do
        PROJECT_DIR="$TEST_WORKSPACE/load-project-$i"
        mkdir -p "$PROJECT_DIR/src/main/java/com/example/loadtest"
        
        # Create varied project complexity
        local project_size=$((i % 4))
        case $project_size in
            0)
                create_small_project "$PROJECT_DIR" "$i"
                ;;
            1)
                create_medium_project "$PROJECT_DIR" "$i"
                ;;
            2)
                create_large_project "$PROJECT_DIR" "$i"
                ;;
            3)
                create_complex_project "$PROJECT_DIR" "$i"
                ;;
        esac
    done
    
    echo "Created 100 test projects with varying complexity in $TEST_WORKSPACE"
}

create_small_project() {
    local project_dir=$1
    local project_id=$2
    
    cat > "$project_dir/build.gradle" << EOF
plugins {
    id 'java'
}

group 'com.example.loadtest'
version '1.0.0'

repositories {
    mavenCentral()
}

task compileJava {
    doLast {
        println "Compiling small project $project_id..."
        mkdir "build/classes/java/main"
    }
}
EOF

    cat > "$project_dir/src/main/java/com/example/loadtest/Small$project_id.java" << EOF
package com.example.loadtest;

public class Small$project_id {
    public static void main(String[] args) {
        System.out.println("Small project $project_id completed!");
    }
}
EOF
}

create_medium_project() {
    local project_dir=$1
    local project_id=$2
    
    cat > "$project_dir/build.gradle" << EOF
plugins {
    id 'java'
}

group 'com.example.loadtest'
version '1.0.0'

repositories {
    mavenCentral()
}

dependencies {
    implementation 'com.google.guava:guava:31.0-jre'
    testImplementation 'junit:junit:4.13.2'
}

task compileJava {
    doLast {
        println "Compiling medium project $project_id..."
        mkdir "build/classes/java/main"
    }
}

task processResources {
    doLast {
        println "Processing resources for project $project_id..."
        mkdir "build/resources/main"
    }
}

task test {
    doLast {
        println "Testing medium project $project_id..."
        mkdir "build/test-results"
    }
}
EOF

    # Create multiple source files
    for j in $(seq 1 3); do
        cat > "$project_dir/src/main/java/com/example/loadtest/Medium${project_id}_$j.java" << EOF
package com.example.loadtest;

public class Medium${project_id}_$j {
    public void process() {
        System.out.println("Processing file $j in medium project $project_id");
    }
}
EOF
    done
}

create_large_project() {
    local project_dir=$1
    local project_id=$2
    
    cat > "$project_dir/build.gradle" << EOF
plugins {
    id 'java'
}

group 'com.example.loadtest'
version '1.0.0'

repositories {
    mavenCentral()
}

dependencies {
    implementation 'com.google.guava:guava:31.0-jre'
    implementation 'org.apache.commons:commons-lang3:3.12.0'
    implementation 'org.springframework:spring-core:5.3.20'
    testImplementation 'junit:junit:4.13.2'
    testImplementation 'org.mockito:mockito-core:4.5.1'
}

task compileJava {
    doLast {
        println "Compiling large project $project_id..."
        mkdir "build/classes/java/main"
    }
}

task processResources {
    doLast {
        println "Processing resources for large project $project_id..."
        mkdir "build/resources/main"
        mkdir "build/resources/main/static"
        mkdir "build/resources/main/templates"
    }
}

task test {
    doLast {
        println "Testing large project $project_id..."
        mkdir "build/test-results"
    }
}

task javadoc {
    doLast {
        println "Generating javadoc for large project $project_id..."
        mkdir "build/docs/javadoc"
    }
}

task jar {
    doLast {
        println "Creating JAR for large project $project_id..."
        mkdir "build/libs"
    }
}
EOF

    # Create many source files
    for j in $(seq 1 10); do
        cat > "$project_dir/src/main/java/com/example/loadtest/Large${project_id}_$j.java" << EOF
package com.example.loadtest;

import com.google.common.base.Strings;
import org.apache.commons.lang3.StringUtils;

public class Large${project_id}_$j {
    private String data = "Large project $project_id file $j data";
    
    public String processData() {
        return StringUtils.capitalize(Strings.nullToEmpty(data));
    }
}
EOF
    done
}

create_complex_project() {
    local project_dir=$1
    local project_id=$2
    
    cat > "$project_dir/build.gradle" << EOF
plugins {
    id 'java'
    id 'application'
    id 'maven-publish'
}

group 'com.example.loadtest'
version '1.0.0'

repositories {
    mavenCentral()
    maven { url 'https://repo.spring.io/milestone' }
}

dependencies {
    implementation 'com.google.guava:guava:31.0-jre'
    implementation 'org.apache.commons:commons-lang3:3.12.0'
    implementation 'org.springframework.boot:spring-boot-starter:2.7.0'
    implementation 'org.springframework.boot:spring-boot-starter-web:2.7.0'
    implementation 'org.springframework.boot:spring-boot-starter-data-jpa:2.7.0'
    implementation 'org.hibernate:hibernate-core:5.6.9.Final'
    implementation 'org.postgresql:postgresql:42.3.4'
    implementation 'redis.clients:jedis:4.2.3'
    
    testImplementation 'junit:junit:4.13.2'
    testImplementation 'org.springframework.boot:spring-boot-starter-test:2.7.0'
    testImplementation 'org.mockito:mockito-core:4.5.1'
    testImplementation 'com.h2database:h2:2.1.212'
}

application {
    mainClass = 'com.example.loadtest.Complex${project_id}Application'
}

publishing {
    publications {
        maven(MavenPublication) {
            from components.java
        }
    }
}

task compileJava {
    doLast {
        println "Compiling complex project $project_id..."
        mkdir "build/classes/java/main"
    }
}

task processResources {
    doLast {
        println "Processing resources for complex project $project_id..."
        mkdir -p "build/resources/main/static/css"
        mkdir -p "build/resources/main/static/js"
        mkdir -p "build/resources/main/templates"
        mkdir -p "build/resources/main/config"
    }
}

task test {
    doLast {
        println "Testing complex project $project_id..."
        mkdir -p "build/test-results/unit"
        mkdir -p "build/test-results/integration"
    }
}

task integrationTest {
    doLast {
        println "Running integration tests for complex project $project_id..."
        mkdir "build/test-results/integration"
    }
}

task javadoc {
    doLast {
        println "Generating javadoc for complex project $project_id..."
        mkdir "build/docs/javadoc"
    }
}

task sourceJar {
    doLast {
        println "Creating source JAR for complex project $project_id..."
        mkdir "build/libs"
    }
}

task jar {
    doLast {
        println "Creating JAR for complex project $project_id..."
        mkdir "build/libs"
    }
}

task dockerImage {
    doLast {
        println "Building Docker image for complex project $project_id..."
        mkdir "build/docker"
    }
}
EOF

    # Create complex source structure
    mkdir -p "$project_dir/src/main/java/com/example/loadtest/controller"
    mkdir -p "$project_dir/src/main/java/com/example/loadtest/service"
    mkdir -p "$project_dir/src/main/java/com/example/loadtest/repository"
    mkdir -p "$project_dir/src/main/java/com/example/loadtest/config"
    mkdir -p "$project_dir/src/test/java/com/example/loadtest"
    
    # Application main class
    cat > "$project_dir/src/main/java/com/example/loadtest/Complex${project_id}Application.java" << EOF
package com.example.loadtest;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class Complex${project_id}Application {
    public static void main(String[] args) {
        SpringApplication.run(Complex${project_id}Application.class, args);
        System.out.println("Complex project $project_id started!");
    }
}
EOF

    # Various component classes
    cat > "$project_dir/src/main/java/com/example/loadtest/controller/Controller${project_id}.java" << EOF
package com.example.loadtest.controller;

import org.springframework.web.bind.annotation.*;
import com.example.loadtest.service.Service${project_id};

@RestController
@RequestMapping("/api/v${project_id}")
public class Controller${project_id} {
    private final Service${project_id} service;
    
    public Controller${project_id}(Service${project_id} service) {
        this.service = service;
    }
    
    @GetMapping("/process")
    public String process() {
        return service.process("Request for project $project_id");
    }
}
EOF

    cat > "$project_dir/src/main/java/com/example/loadtest/service/Service${project_id}.java" << EOF
package com.example.loadtest.service;

import org.springframework.stereotype.Service;
import com.google.common.base.Strings;
import org.apache.commons.lang3.StringUtils;

@Service
public class Service${project_id} {
    public String process(String input) {
        return StringUtils.capitalize(Strings.nullToEmpty(input));
    }
}
EOF
}

# Test sustained load handling
test_sustained_load() {
    echo "Testing sustained load handling..."
    
    # Start mock coordinator for load testing
    cd "$TEST_WORKSPACE"
    
    python3 -c "
import http.server
import socketserver
import json
import time
import threading
from collections import defaultdict

class LoadTestCoordinator(http.server.BaseHTTPRequestHandler):
    builds = defaultdict(dict)
    build_counter = 0
    stats = {
        'total_submissions': 0,
        'total_completions': 0,
        'peak_concurrent': 0,
        'current_concurrent': 0,
        'response_times': []
    }
    
    def do_POST(self):
        if self.path == '/api/builds':
            start_time = time.time()
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            
            build_id = f'build-{LoadTestCoordinator.build_counter}'
            LoadTestCoordinator.build_counter += 1
            
            LoadTestCoordinator.stats['total_submissions'] += 1
            LoadTestCoordinator.stats['current_concurrent'] += 1
            LoadTestCoordinator.stats['peak_concurrent'] = max(
                LoadTestCoordinator.stats['peak_concurrent'],
                LoadTestCoordinator.stats['current_concurrent']
            )
            
            LoadTestCoordinator.builds[build_id] = {
                'status': 'running',
                'created_at': start_time,
                'project_name': json.loads(post_data.decode('utf-8')).get('project_name', 'unknown')
            }
            
            # Simulate processing time based on project complexity
            processing_time = 0.5 + (hash(build_id) % 10) * 0.1
            threading.Timer(processing_time, self._complete_build, args=[build_id, start_time]).start()
            
            response = {
                'build_id': build_id,
                'status': 'accepted',
                'queue_position': LoadTestCoordinator.stats['current_concurrent']
            }
            
            self.send_response(202)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(response).encode())
    
    def _complete_build(self, build_id, start_time):
        LoadTestCoordinator.stats['total_completions'] += 1
        LoadTestCoordinator.stats['current_concurrent'] -= 1
        response_time = time.time() - start_time
        LoadTestCoordinator.stats['response_times'].append(response_time)
        LoadTestCoordinator.builds[build_id]['status'] = 'completed'
        LoadTestCoordinator.builds[build_id]['response_time'] = response_time
    
    def do_GET(self):
        if self.path == '/api/stats':
            stats = LoadTestCoordinator.stats.copy()
            if stats['response_times']:
                stats['avg_response_time'] = sum(stats['response_times']) / len(stats['response_times'])
                stats['max_response_time'] = max(stats['response_times'])
                stats['min_response_time'] = min(stats['response_times'])
            else:
                stats['avg_response_time'] = 0
                stats['max_response_time'] = 0
                stats['min_response_time'] = 0
            
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(stats).encode())

PORT = $COORDINATOR_PORT
with socketserver.TCPServer(('', PORT), LoadTestCoordinator) as httpd:
    print(f'Load test coordinator running on port {PORT}')
    httpd.serve_forever()
" &
    COORDINATOR_PID=$!
    
    sleep 2
    
    # Execute sustained load test
    echo "Running sustained load test for $LOAD_TEST_DURATION seconds..."
    
    START_TIME=$(date +%s)
    END_TIME=$((START_TIME + LOAD_TEST_DURATION))
    BUILD_COUNT=0
    
    while [ $(date +%s) -lt $END_TIME ]; do
        # Generate random project index
        PROJECT_INDEX=$((BUILD_COUNT % 100 + 1))
        PROJECT_DIR="$TEST_WORKSPACE/load-project-$PROJECT_INDEX"
        
        # Random task selection based on project complexity
        case $((PROJECT_INDEX % 4)) in
            0)
                TASKS='["compileJava"]'
                ;;
            1)
                TASKS='["compileJava", "processResources", "test"]'
                ;;
            2)
                TASKS='["compileJava", "processResources", "test", "javadoc", "jar"]'
                ;;
            3)
                TASKS='["compileJava", "processResources", "test", "integrationTest", "javadoc", "jar", "dockerImage"]'
                ;;
        esac
        
        BUILD_REQUEST="{
            \"project_name\": \"load-test-$BUILD_COUNT\",
            \"build_type\": \"gradle\",
            \"tasks\": $TASKS,
            \"source_location\": \"$PROJECT_DIR\",
            \"environment\": {
                \"java_home\": \"/usr/lib/jvm/default-java\",
                \"gradle_home\": \"/usr/local/gradle\"
            }
        }"
        
        # Submit build with timeout
        if curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
            -H "Content-Type: application/json" \
            -d "$BUILD_REQUEST" > /dev/null; then
            BUILD_COUNT=$((BUILD_COUNT + 1))
        fi
        
        # Ramp up control
        if [ $BUILD_COUNT -lt $((LOAD_TEST_RAMP_UP / 2)) ]; then
            sleep 0.1  # Slower start
        elif [ $BUILD_COUNT -lt $LOAD_TEST_RAMP_UP ]; then
            sleep 0.05  # Ramp up
        else
            sleep 0.02  # Full speed
        fi
    done
    
    # Wait for remaining builds to complete
    echo "Waiting for builds to complete..."
    sleep 10
    
    # Get load test statistics
    STATS_RESPONSE=$(curl -s "http://localhost:$COORDINATOR_PORT/api/stats")
    
    TOTAL_SUBMISSIONS=$(echo "$STATS_RESPONSE" | jq -r '.total_submissions')
    TOTAL_COMPLETIONS=$(echo "$STATS_RESPONSE" | jq -r '.total_completions')
    PEAK_CONCURRENT=$(echo "$STATS_RESPONSE" | jq -r '.peak_concurrent')
    AVG_RESPONSE_TIME=$(echo "$STATS_RESPONSE" | jq -r '.avg_response_time')
    MAX_RESPONSE_TIME=$(echo "$STATS_RESPONSE" | jq -r '.max_response_time')
    
    echo "Load Test Results:"
    echo "  Total submissions: $TOTAL_SUBMISSIONS"
    echo "  Total completions: $TOTAL_COMPLETIONS"
    echo "  Peak concurrent builds: $PEAK_CONCURRENT"
    echo "  Average response time: ${AVG_RESPONSE_TIME}s"
    echo "  Maximum response time: ${MAX_RESPONSE_TIME}s"
    
    # Evaluate performance metrics
    SUCCESS_RATE=$(echo "scale=2; $TOTAL_COMPLETIONS / $TOTAL_SUBMISSIONS * 100" | bc)
    
    if (( $(echo "$SUCCESS_RATE >= 95" | bc -l) )); then
        test_pass "Load test success rate: ${SUCCESS_RATE}% (target: ≥95%)"
    else
        test_fail "Load test success rate too low: ${SUCCESS_RATE}% (target: ≥95%)"
    fi
    
    if (( $(echo "$AVG_RESPONSE_TIME < 5.0" | bc -l) )); then
        test_pass "Load test avg response time: ${AVG_RESPONSE_TIME}s (target: <5s)"
    else
        test_fail "Load test avg response time too high: ${AVG_RESPONSE_TIME}s (target: <5s)"
    fi
    
    if [ "$PEAK_CONCURRENT" -ge $LOAD_TEST_CONCURRENT_BUILDS ]; then
        test_pass "Load test peak concurrent: $PEAK_CONCURRENT (target: ≥$LOAD_TEST_CONCURRENT_BUILDS)"
    else
        test_fail "Load test peak concurrent too low: $PEAK_CONCURRENT (target: ≥$LOAD_TEST_CONCURRENT_BUILDS)"
    fi
    
    # Stop coordinator
    kill $COORDINATOR_PID 2>/dev/null
}

# Test resource utilization under load
test_resource_utilization() {
    echo "Testing resource utilization under load..."
    
    # Monitor baseline resources
    BASELINE_CPU=$(ps aux | grep -E "(python|bash)" | grep -v grep | awk '{sum+=$3} END {print sum+0}')
    BASELINE_MEMORY=$(ps aux | grep -E "(python|bash)" | grep -v grep | awk '{sum+=$4} END {print sum+0}')
    
    echo "Baseline CPU: ${BASELINE_CPU}%, Memory: ${BASELINE_MEMORY}%"
    
    # Start resource monitor
    python3 -c "
import time
import psutil
import json

# Monitor for 60 seconds
start_time = time.time()
max_cpu = 0
max_memory = 0

while time.time() - start_time < 60:
    cpu_percent = psutil.cpu_percent(interval=1)
    memory_percent = psutil.virtual_memory().percent
    
    max_cpu = max(max_cpu, cpu_percent)
    max_memory = max(max_memory, memory_percent)
    
    time.sleep(1)

stats = {
    'max_cpu': max_cpu,
    'max_memory': max_memory
}

with open('/tmp/resource_stats.json', 'w') as f:
    json.dump(stats, f)
" &
    MONITOR_PID=$!
    
    # Start lightweight coordinator
    python3 -c "
import http.server
import socketserver
import json
import time

class ResourceTestCoordinator(http.server.BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path == '/api/builds':
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            
            build_id = f'build-{int(time.time() * 1000)}'
            
            # Simulate some processing
            time.sleep(0.1)
            
            response = {
                'build_id': build_id,
                'status': 'accepted'
            }
            
            self.send_response(202)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(response).encode())

PORT = $COORDINATOR_PORT
with socketserver.TCPServer(('', PORT), ResourceTestCoordinator) as httpd:
    httpd.serve_forever()
" &
    COORDINATOR_PID=$!
    
    sleep 2
    
    # Generate load
    for i in $(seq 1 100); do
        BUILD_REQUEST="{
            \"project_name\": \"resource-test-$i\",
            \"build_type\": \"gradle\",
            \"tasks\": [\"compileJava\"]
        }"
        
        curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
            -H "Content-Type: application/json" \
            -d "$BUILD_REQUEST" > /dev/null &
        
        # Control concurrency
        if [ $((i % 20)) -eq 0 ]; then
            sleep 1
        fi
    done
    
    wait
    
    # Wait for monitoring to complete
    wait $MONITOR_PID
    
    # Get resource stats
    if [ -f "/tmp/resource_stats.json" ]; then
        RESOURCE_STATS=$(cat /tmp/resource_stats.json)
        MAX_CPU=$(echo "$RESOURCE_STATS" | jq -r '.max_cpu')
        MAX_MEMORY=$(echo "$RESOURCE_STATS" | jq -r '.max_memory')
        
        echo "Resource utilization under load:"
        echo "  Max CPU usage: ${MAX_CPU}%"
        echo "  Max Memory usage: ${MAX_MEMORY}%"
        
        if (( $(echo "$MAX_CPU < 80" | bc -l) )); then
            test_pass "CPU usage under load: ${MAX_CPU}% (target: <80%)"
        else
            test_fail "CPU usage too high under load: ${MAX_CPU}% (target: <80%)"
        fi
        
        if [ "$MAX_MEMORY" -lt 90 ]; then
            test_pass "Memory usage under load: ${MAX_MEMORY}% (target: <90%)"
        else
            test_fail "Memory usage too high under load: ${MAX_MEMORY}% (target: <90%)"
        fi
        
        rm -f "/tmp/resource_stats.json"
    fi
    
    # Stop coordinator
    kill $COORDINATOR_PID 2>/dev/null
}

# Test error handling under load
test_error_handling_under_load() {
    echo "Testing error handling under load..."
    
    # Start coordinator with error injection
    python3 -c "
import http.server
import socketserver
import json
import time
import random

class ErrorHandlingCoordinator(http.server.BaseHTTPRequestHandler):
    request_count = 0
    error_count = 0
    
    def do_POST(self):
        ErrorHandlingCoordinator.request_count += 1
        
        # Inject 5% error rate
        if random.random() < 0.05:
            ErrorHandlingCoordinator.error_count += 1
            self.send_response(500)
            self.end_headers()
            self.wfile.write(b'{\"error\": \"Simulated error\"}')
            return
        
        if self.path == '/api/builds':
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            
            build_id = f'build-{int(time.time() * 1000)}'
            
            response = {
                'build_id': build_id,
                'status': 'accepted'
            }
            
            self.send_response(202)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(response).encode())
    
    def do_GET(self):
        if self.path == '/api/error-stats':
            stats = {
                'total_requests': ErrorHandlingCoordinator.request_count,
                'total_errors': ErrorHandlingCoordinator.error_count,
                'error_rate': ErrorHandlingCoordinator.error_count / max(1, ErrorHandlingCoordinator.request_count)
            }
            
            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps(stats).encode())

PORT = $COORDINATOR_PORT
with socketserver.TCPServer(('', PORT), ErrorHandlingCoordinator) as httpd:
    httpd.serve_forever()
" &
    COORDINATOR_PID=$!
    
    sleep 2
    
    # Generate requests with error handling
    SUCCESSFUL_REQUESTS=0
    FAILED_REQUESTS=0
    TOTAL_REQUESTS=0
    
    for i in $(seq 1 200); do
        BUILD_REQUEST="{
            \"project_name\": \"error-test-$i\",
            \"build_type\": \"gradle\",
            \"tasks\": [\"compileJava\"]
        }"
        
        HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
            -H "Content-Type: application/json" \
            -d "$BUILD_REQUEST")
        
        TOTAL_REQUESTS=$((TOTAL_REQUESTS + 1))
        
        if [ "$HTTP_CODE" = "202" ]; then
            SUCCESSFUL_REQUESTS=$((SUCCESSFUL_REQUESTS + 1))
        else
            FAILED_REQUESTS=$((FAILED_REQUESTS + 1))
        fi
        
        # Some concurrent requests
        if [ $((i % 10)) -eq 0 ]; then
            {
                curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
                    -H "Content-Type: application/json" \
                    -d "{\"project_name\": \"concurrent-error-$i\"}" > /dev/null &
            } &
        fi
    done
    
    wait
    
    # Get error statistics
    ERROR_STATS=$(curl -s "http://localhost:$COORDINATOR_PORT/api/error-stats")
    
    SERVER_REQUESTS=$(echo "$ERROR_STATS" | jq -r '.total_requests')
    SERVER_ERRORS=$(echo "$ERROR_STATS" | jq -r '.total_errors')
    SERVER_ERROR_RATE=$(echo "$ERROR_STATS" | jq -r '.error_rate')
    
    echo "Error handling test results:"
    echo "  Total requests: $TOTAL_REQUESTS"
    echo "  Successful requests: $SUCCESSFUL_REQUESTS"
    echo "  Failed requests: $FAILED_REQUESTS"
    echo "  Server error rate: $(echo "$SERVER_ERROR_RATE * 100" | bc -l)%"
    
    CLIENT_SUCCESS_RATE=$(echo "scale=2; $SUCCESSFUL_REQUESTS / $TOTAL_REQUESTS * 100" | bc)
    
    if (( $(echo "$CLIENT_SUCCESS_RATE >= 90" | bc -l) )); then
        test_pass "Client success rate under error conditions: ${CLIENT_SUCCESS_RATE}% (target: ≥90%)"
    else
        test_fail "Client success rate too low under error conditions: ${CLIENT_SUCCESS_RATE}% (target: ≥90%)"
    fi
    
    # Stop coordinator
    kill $COORDINATOR_PID 2>/dev/null
}

# Test cleanup
test_cleanup() {
    echo "Cleaning up load test..."
    
    # Kill any remaining processes
    pkill -f "python3.*coord" 2>/dev/null || true
    pkill -f "python3.*LoadTest" 2>/dev/null || true
    pkill -f "python3.*Resource" 2>/dev/null || true
    pkill -f "python3.*ErrorHandling" 2>/dev/null || true
    
    # Cleanup test workspace
    rm -rf "$TEST_WORKSPACE"
    
    test_pass "Load test cleanup completed"
}

# Main test execution
main() {
    echo "Starting load tests..."
    
    # Check dependencies
    if ! command -v bc >/dev/null 2>&1; then
        test_fail "bc calculator required for load tests"
        exit 1
    fi
    
    if ! command -v python3 >/dev/null 2>&1; then
        test_fail "python3 required for load test mock services"
        exit 1
    fi
    
    # Run load test suite
    setup_load_test || exit 1
    test_sustained_load || exit 1
    test_resource_utilization || exit 1
    test_error_handling_under_load || exit 1
    test_cleanup || exit 1
    
    echo "Load tests completed successfully!"
}

# Run tests if script is executed directly
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi