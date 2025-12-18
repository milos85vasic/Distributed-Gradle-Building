#!/bin/bash

# Integration test for end-to-end distributed build workflow
# This test validates the complete build system from submission to completion

# Import test framework
source "$(dirname "$0")/../test_framework.sh"

# Test configuration
source "$(dirname "$0")/../test_config.sh"

# Test setup
setup_test() {
    echo "Setting up end-to-end integration test..."
    
    # Create test workspace
    TEST_WORKSPACE=$(mktemp -d)
    export TEST_WORKSPACE
    
    # Create test project
    mkdir -p "$TEST_WORKSPACE/project"
    cd "$TEST_WORKSPACE/project"
    
    # Create a simple Gradle project
    cat > build.gradle << 'EOF'
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
        println "Compiling Java sources..."
        mkdir "build/classes/java/main"
        // Create a simple class file
        echo "public class Hello { public static void main(String[] args) { System.out.println(\"Hello World!\"); } }" > build/classes/java/main/Hello.java
    }
}
EOF
    
    mkdir -p src/main/java/com/example
    cat > src/main/java/com/example/Hello.java << 'EOF'
package com.example;

public class Hello {
    public static void main(String[] args) {
        System.out.println("Hello World!");
    }
}
EOF
    
    echo "Test project created at $TEST_WORKSPACE/project"
}

# Test coordinator startup
test_coordinator_startup() {
    echo "Testing coordinator startup..."
    
    cd "$TEST_WORKSPACE"
    
    # Start coordinator
    $COORDINATOR_DIR/start_coordinator.sh --port $COORDINATOR_PORT --cache-server "localhost:$CACHE_SERVER_PORT" &
    COORDINATOR_PID=$!
    
    # Wait for coordinator to be ready
    wait_for_service "http://localhost:$COORDINATOR_PORT/health" 30
    
    if [ $? -eq 0 ]; then
        test_pass "Coordinator started successfully"
    else
        test_fail "Coordinator failed to start"
        return 1
    fi
}

# Test worker startup and registration
test_worker_startup() {
    echo "Testing worker startup and registration..."
    
    # Start first worker
    $WORKER_DIR/start_worker.sh --coordinator "localhost:$COORDINATOR_PORT" --id worker1 --port $WORKER1_PORT &
    WORKER1_PID=$!
    
    # Start second worker
    $WORKER_DIR/start_worker.sh --coordinator "localhost:$COORDINATOR_PORT" --id worker2 --port $WORKER2_PORT &
    WORKER2_PID=$!
    
    # Wait for workers to register
    sleep 5
    
    # Check worker registration
    RESPONSE=$(curl -s "http://localhost:$COORDINATOR_PORT/api/workers" | jq '.')
    WORKER_COUNT=$(echo "$RESPONSE" | jq '.workers | length')
    
    if [ "$WORKER_COUNT" -eq 2 ]; then
        test_pass "Workers registered successfully"
    else
        test_fail "Expected 2 workers, found $WORKER_COUNT"
        return 1
    fi
}

# Test cache server startup
test_cache_server_startup() {
    echo "Testing cache server startup..."
    
    cd "$TEST_WORKSPACE"
    
    # Start cache server
    $CACHE_DIR/start_cache_server.sh --port $CACHE_SERVER_PORT &
    CACHE_PID=$!
    
    # Wait for cache server to be ready
    wait_for_service "http://localhost:$CACHE_SERVER_PORT/health" 30
    
    if [ $? -eq 0 ]; then
        test_pass "Cache server started successfully"
    else
        test_fail "Cache server failed to start"
        return 1
    fi
}

# Test build submission
test_build_submission() {
    echo "Testing build submission..."
    
    cd "$TEST_WORKSPACE/project"
    
    # Submit build via API
    BUILD_REQUEST='{
        "project_name": "test-project",
        "build_type": "gradle",
        "tasks": ["compileJava", "test"],
        "source_location": "'"$TEST_WORKSPACE/project"'",
        "environment": {
            "java_home": "/usr/lib/jvm/default-java",
            "gradle_home": "/usr/local/gradle"
        }
    }'
    
    RESPONSE=$(curl -s -X POST "http://localhost:$COORDINATOR_PORT/api/builds" \
        -H "Content-Type: application/json" \
        -d "$BUILD_REQUEST")
    
    BUILD_ID=$(echo "$RESPONSE" | jq -r '.build_id')
    
    if [ "$BUILD_ID" != "null" ] && [ "$BUILD_ID" != "" ]; then
        test_pass "Build submitted successfully with ID: $BUILD_ID"
        export BUILD_ID
    else
        test_fail "Failed to submit build: $RESPONSE"
        return 1
    fi
}

# Test build execution and monitoring
test_build_execution() {
    echo "Testing build execution and monitoring..."
    
    if [ -z "$BUILD_ID" ]; then
        test_fail "No build ID available for monitoring"
        return 1
    fi
    
    # Monitor build progress
    MAX_WAIT=60
    WAIT_TIME=0
    
    while [ $WAIT_TIME -lt $MAX_WAIT ]; do
        STATUS_RESPONSE=$(curl -s "http://localhost:$COORDINATOR_PORT/api/builds/$BUILD_ID/status")
        BUILD_STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.status')
        PROGRESS=$(echo "$STATUS_RESPONSE" | jq -r '.progress')
        
        echo "Build status: $BUILD_STATUS, Progress: $PROGRESS%"
        
        if [ "$BUILD_STATUS" = "completed" ]; then
            test_pass "Build completed successfully"
            break
        elif [ "$BUILD_STATUS" = "failed" ]; then
            test_fail "Build failed"
            return 1
        fi
        
        sleep 2
        WAIT_TIME=$((WAIT_TIME + 2))
    done
    
    if [ $WAIT_TIME -ge $MAX_WAIT ]; then
        test_fail "Build timed out after $MAX_WAIT seconds"
        return 1
    fi
}

# Test artifact retrieval
test_artifact_retrieval() {
    echo "Testing artifact retrieval..."
    
    if [ -z "$BUILD_ID" ]; then
        test_fail "No build ID available for artifact retrieval"
        return 1
    fi
    
    # Get build artifacts
    ARTIFACTS_RESPONSE=$(curl -s "http://localhost:$COORDINATOR_PORT/api/builds/$BUILD_ID/artifacts")
    ARTIFACTS=$(echo "$ARTIFACTS_RESPONSE" | jq -r '.artifacts[]')
    
    if [ -n "$ARTIFACTS" ]; then
        test_pass "Artifacts retrieved successfully"
        echo "Retrieved artifacts: $ARTIFACTS"
        
        # Download artifacts
        mkdir -p "$TEST_WORKSPACE/artifacts"
        for artifact in $ARTIFACTS; do
            curl -s "http://localhost:$COORDINATOR_PORT/api/builds/$BUILD_ID/artifacts/$artifact" \
                -o "$TEST_WORKSPACE/artifacts/$artifact"
        done
        
        # Verify artifacts exist
        if [ -f "$TEST_WORKSPACE/artifacts/build.log" ]; then
            test_pass "Build log artifact retrieved"
        else
            test_fail "Build log artifact missing"
        fi
    else
        test_fail "No artifacts found"
        return 1
    fi
}

# Test cleanup
test_cleanup() {
    echo "Testing cleanup..."
    
    # Stop services
    kill $COORDINATOR_PID 2>/dev/null
    kill $WORKER1_PID 2>/dev/null
    kill $WORKER2_PID 2>/dev/null
    kill $CACHE_PID 2>/dev/null
    
    # Wait for services to stop
    sleep 3
    
    # Cleanup test workspace
    rm -rf "$TEST_WORKSPACE"
    
    test_pass "Cleanup completed"
}

# Main test execution
main() {
    echo "Starting end-to-end integration test..."
    
    # Run test suite
    setup_test || exit 1
    test_coordinator_startup || exit 1
    test_cache_server_startup || exit 1
    test_worker_startup || exit 1
    test_build_submission || exit 1
    test_build_execution || exit 1
    test_artifact_retrieval || exit 1
    test_cleanup || exit 1
    
    echo "End-to-end integration test completed successfully!"
}

# Run tests if script is executed directly
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi