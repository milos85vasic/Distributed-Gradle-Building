#!/bin/bash
# Connection Validation Script
# Validates that all components of the Distributed Gradle Building System are properly connected

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
RESULTS_FILE="$PROJECT_ROOT/connection_validation_results.txt"

# Initialize results
echo "📡 Distributed Gradle Building - Connection Validation" > "$RESULTS_FILE"
echo "=========================================================" >> "$RESULTS_FILE"
echo "Date: $(date)" >> "$RESULTS_FILE"
echo "" >> "$RESULTS_FILE"

# Validation functions
validate_file_exists() {
    local file="$1"
    local description="$2"
    if [[ -f "$file" ]]; then
        echo "✅ $description: $file" >> "$RESULTS_FILE"
        return 0
    else
        echo "❌ $description: $file (MISSING)" >> "$RESULTS_FILE"
        return 1
    fi
}

validate_directory_exists() {
    local dir="$1"
    local description="$2"
    if [[ -d "$dir" ]]; then
        echo "✅ $description: $dir" >> "$RESULTS_FILE"
        return 0
    else
        echo "❌ $description: $dir (MISSING)" >> "$RESULTS_FILE"
        return 1
    fi
}

validate_script_executable() {
    local script="$1"
    local description="$2"
    if [[ -x "$script" ]]; then
        echo "✅ $description: $script (executable)" >> "$RESULTS_FILE"
        return 0
    else
        echo "⚠️  $description: $script (not executable)" >> "$RESULTS_FILE"
        return 1
    fi
}

validate_link_reference() {
    local file="$1"
    local pattern="$2"
    local description="$3"
    if grep -q "$pattern" "$file" 2>/dev/null; then
        echo "✅ $description: Found in $file" >> "$RESULTS_FILE"
        return 0
    else
        echo "❌ $description: Missing in $file" >> "$RESULTS_FILE"
        return 1
    fi
}

# Count total tests
TOTAL_TESTS=0
PASSED_TESTS=0

# Run validations
echo "Running connection validation..." >&2

echo "" >> "$RESULTS_FILE"
echo "🏠 Main Project Structure" >> "$RESULTS_FILE"
echo "------------------------" >> "$RESULTS_FILE"

# Core files
TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/README.md" "Main README" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/QUICK_START.md" "Quick Start Guide" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/CONNECTIONS.md" "Connections Documentation" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/distributed_gradle_build.sh" "Distributed Build Engine" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/sync_and_build.sh" "Main Build Interface" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/setup_master.sh" "Master Setup Script" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/setup_worker.sh" "Worker Setup Script" && PASSED_TESTS=$((PASSED_TESTS + 1))

echo "" >> "$RESULTS_FILE"
echo "📚 Documentation Structure" >> "$RESULTS_FILE"
echo "-------------------------" >> "$RESULTS_FILE"

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_directory_exists "$PROJECT_ROOT/docs" "Documentation Directory" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/docs/README.md" "Documentation Index" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/docs/SETUP_GUIDE.md" "Setup Guide" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/docs/USER_GUIDE.md" "User Guide" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/docs/GO_DEPLOYMENT.md" "Go Deployment Guide" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/docs/API_REFERENCE.md" "API Reference" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/docs/PERFORMANCE.md" "Performance Guide" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/docs/MONITORING.md" "Monitoring Guide" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/docs/CICD.md" "CI/CD Integration Guide" && PASSED_TESTS=$((PASSED_TESTS + 1))

echo "" >> "$RESULTS_FILE"
echo "🌍 Website Structure" >> "$RESULTS_FILE"
echo "-------------------" >> "$RESULTS_FILE"

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_directory_exists "$PROJECT_ROOT/website" "Website Directory" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/website/content/_index.md" "Website Homepage" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_directory_exists "$PROJECT_ROOT/website/content/docs" "Website Documentation" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_directory_exists "$PROJECT_ROOT/website/content/tutorials" "Website Tutorials" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_directory_exists "$PROJECT_ROOT/website/content/video-courses" "Website Video Courses" && PASSED_TESTS=$((PASSED_TESTS + 1))

echo "" >> "$RESULTS_FILE"
echo "🏗️ Go Implementation" >> "$RESULTS_FILE"
echo "-------------------" >> "$RESULTS_FILE"

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_directory_exists "$PROJECT_ROOT/go" "Go Implementation Directory" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/go/main.go" "Go Main Service" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/go/worker.go" "Go Worker Service" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/go/cache_server.go" "Go Cache Server" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/go/monitor.go" "Go Monitor Service" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/go/go.mod" "Go Modules File" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_directory_exists "$PROJECT_ROOT/go/client" "Go Client Library" && PASSED_TESTS=$((PASSED_TESTS + 1))

echo "" >> "$RESULTS_FILE"
echo "🧪 Test Framework" >> "$RESULTS_FILE"
echo "----------------" >> "$RESULTS_FILE"

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_directory_exists "$PROJECT_ROOT/tests" "Test Framework Directory" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/tests/test_framework.sh" "Test Framework Core" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/tests/quick_distributed_verification.sh" "Quick Verification Test" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_directory_exists "$PROJECT_ROOT/tests/comprehensive" "Comprehensive Tests" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_directory_exists "$PROJECT_ROOT/tests/integration" "Integration Tests" && PASSED_TESTS=$((PASSED_TESTS + 1))

echo "" >> "$RESULTS_FILE"
echo "🔧 Scripts Directory" >> "$RESULTS_FILE"
echo "-------------------" >> "$RESULTS_FILE"

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_directory_exists "$PROJECT_ROOT/scripts" "Scripts Directory" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_file_exists "$PROJECT_ROOT/scripts/run-all-tests.sh" "Test Runner Script" && PASSED_TESTS=$((PASSED_TESTS + 1))

echo "" >> "$RESULTS_FILE"
echo "🔗 Cross-Reference Validation" >> "$RESULTS_FILE"
echo "----------------------------" >> "$RESULTS_FILE"

# Validate key cross-references
TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_link_reference "$PROJECT_ROOT/README.md" "CONNECTIONS.md" "README links to CONNECTIONS.md" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_link_reference "$PROJECT_ROOT/README.md" "QUICK_START.md" "README links to QUICK_START.md" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_link_reference "$PROJECT_ROOT/README.md" "docs/" "README links to docs/" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_link_reference "$PROJECT_ROOT/QUICK_START.md" "docs/" "QuickStart links to docs/" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_link_reference "$PROJECT_ROOT/QUICK_START.md" "CONNECTIONS.md" "QuickStart links to CONNECTIONS.md" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_link_reference "$PROJECT_ROOT/docs/README.md" "GO_DEPLOYMENT.md" "Docs index links to Go deployment" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_link_reference "$PROJECT_ROOT/docs/README.md" "API_REFERENCE.md" "Docs index links to API reference" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_link_reference "$PROJECT_ROOT/website/content/_index.md" "docs/" "Website homepage links to docs/" && PASSED_TESTS=$((PASSED_TESTS + 1))

echo "" >> "$RESULTS_FILE"
echo "🔍 Content Consistency" >> "$RESULTS_FILE"
echo "-------------------" >> "$RESULTS_FILE"

# Check for script executability
TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_script_executable "$PROJECT_ROOT/distributed_gradle_build.sh" "Distributed Build Engine" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_script_executable "$PROJECT_ROOT/sync_and_build.sh" "Main Build Interface" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_script_executable "$PROJECT_ROOT/setup_master.sh" "Master Setup Script" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_script_executable "$PROJECT_ROOT/setup_worker.sh" "Worker Setup Script" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_script_executable "$PROJECT_ROOT/tests/test_framework.sh" "Test Framework Core" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_script_executable "$PROJECT_ROOT/tests/quick_distributed_verification.sh" "Quick Verification Test" && PASSED_TESTS=$((PASSED_TESTS + 1))

TOTAL_TESTS=$((TOTAL_TESTS + 1))
validate_script_executable "$PROJECT_ROOT/scripts/run-all-tests.sh" "Test Runner Script" && PASSED_TESTS=$((PASSED_TESTS + 1))

# Summary
echo "" >> "$RESULTS_FILE"
echo "📊 Summary" >> "$RESULTS_FILE"
echo "---------" >> "$RESULTS_FILE"
echo "Total Tests: $TOTAL_TESTS" >> "$RESULTS_FILE"
echo "Passed: $PASSED_TESTS" >> "$RESULTS_FILE"
echo "Failed: $((TOTAL_TESTS - PASSED_TESTS))" >> "$RESULTS_FILE"
echo "Success Rate: $(( PASSED_TESTS * 100 / TOTAL_TESTS ))%" >> "$RESULTS_FILE"

if [[ $PASSED_TESTS -eq $TOTAL_TESTS ]]; then
    echo "" >> "$RESULTS_FILE"
    echo "🎉 All connections are properly established!" >> "$RESULTS_FILE"
    exit 0
else
    echo "" >> "$RESULTS_FILE"
    echo "⚠️  Some connections need attention." >> "$RESULTS_FILE"
    exit 1
fi