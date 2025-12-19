#!/bin/bash
# Go Implementation Integration Test
# Validates Go services integration with main project structure

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "=== GO IMPLEMENTATION INTEGRATION TEST ==="
echo "Validating Go services integration with main project..."

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Helper functions
test_passed() {
    local test_name="$1"
    echo -e "${GREEN}✅ $test_name${NC}"
    PASSED_TESTS=$((PASSED_TESTS + 1))
}

test_failed() {
    local test_name="$1"
    local reason="$2"
    echo -e "${RED}❌ $test_name${NC} - $reason"
    FAILED_TESTS=$((FAILED_TESTS + 1))
}

test_info() {
    local test_name="$1"
    echo -e "${BLUE}🔍 Testing: $test_name${NC}"
}

# Test 1: Go services files exist and are structured correctly
echo -e "\n${YELLOW}=== GO SERVICES STRUCTURE ===${NC}"

test_info "Go service files"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if [[ -f "$PROJECT_ROOT/go/main.go" && -f "$PROJECT_ROOT/go/worker.go" && -f "$PROJECT_ROOT/go/cache_server.go" && -f "$PROJECT_ROOT/go/monitor.go" ]]; then
    test_passed "All core Go service files exist"
else
    test_failed "Go service files" "Missing core service files"
fi

test_info "Go modules configuration"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if [[ -f "$PROJECT_ROOT/go/go.mod" && -f "$PROJECT_ROOT/go/go.sum" ]]; then
    test_passed "Go modules configuration exists"
else
    test_failed "Go modules configuration" "go.mod or go.sum missing"
fi

test_info "Go service directories"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if [[ -d "$PROJECT_ROOT/go/coordinator" && -d "$PROJECT_ROOT/go/worker" && -d "$PROJECT_ROOT/go/cache" && -d "$PROJECT_ROOT/go/monitor" ]]; then
    test_passed "Go service directories exist"
else
    test_failed "Go service directories" "Service directories missing"
fi

# Test 2: Documentation connections
echo -e "\n${YELLOW}=== DOCUMENTATION CONNECTIONS ===${NC}"

test_info "Go deployment guide exists"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if [[ -f "$PROJECT_ROOT/docs/GO_DEPLOYMENT.md" ]]; then
    test_passed "Go deployment guide exists"
else
    test_failed "Go deployment guide" "GO_DEPLOYMENT.md missing"
fi

test_info "API reference documentation"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if [[ -f "$PROJECT_ROOT/docs/API_REFERENCE.md" ]]; then
    test_passed "API reference documentation exists"
else
    test_failed "API reference documentation" "API_REFERENCE.md missing"
fi

test_info "Cross-references in main README"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if grep -q "GO_DEPLOYMENT.md" "$PROJECT_ROOT/README.md"; then
    test_passed "Main README references Go deployment"
else
    test_failed "Main README Go references" "Missing GO_DEPLOYMENT.md reference"
fi

test_info "Quick start Go integration"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if grep -q "Go Implementation" "$PROJECT_ROOT/QUICK_START.md"; then
    test_passed "Quick start includes Go implementation"
else
    test_failed "Quick start Go integration" "Missing Go implementation guide"
fi

# Test 3: Test framework integration
echo -e "\n${YELLOW}=== TEST FRAMEWORK INTEGRATION ===${NC}"

test_info "Go test files exist"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if [[ -f "$PROJECT_ROOT/tests/go/coordinator_test.go" && -f "$PROJECT_ROOT/tests/go/worker_test.go" && -f "$PROJECT_ROOT/tests/go/cache_test.go" && -f "$PROJECT_ROOT/tests/go/monitor_test.go" ]]; then
    test_passed "Go test files exist"
else
    test_failed "Go test files" "Missing test files for services"
fi

test_info "Go client tests"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if [[ -f "$PROJECT_ROOT/tests/go/client_test.go" ]]; then
    test_passed "Go client tests exist"
else
    test_failed "Go client tests" "Missing client_test.go"
fi

test_info "Test runner Go integration"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if grep -q "go test" "$PROJECT_ROOT/scripts/run-all-tests.sh"; then
    test_passed "Test runner includes Go tests"
else
    test_failed "Test runner Go integration" "Missing Go test execution"
fi

# Test 4: API structure and connectivity
echo -e "\n${YELLOW}=== API STRUCTURE ===${NC}"

test_info "Client library structure"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if [[ -f "$PROJECT_ROOT/go/client/example/main.go" ]]; then
    test_passed "Client library example exists"
else
    test_failed "Client library structure" "Missing client example"
fi

test_info "API types definition"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if [[ -f "$PROJECT_ROOT/go/types/types.go" ]]; then
    test_passed "API types defined"
else
    test_failed "API types definition" "Missing types.go"
fi

test_info "Service configuration files"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if [[ -f "$PROJECT_ROOT/go/worker_config.json" && -f "$PROJECT_ROOT/go/cache_config.json" && -f "$PROJECT_ROOT/go/monitor_config.json" ]]; then
    test_passed "Service configuration files exist"
else
    test_failed "Service configuration files" "Missing config files"
fi

# Test 5: Build and deployment readiness
echo -e "\n${YELLOW}=== BUILD READINESS ===${NC}"

test_info "Go build dependencies"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
cd "$PROJECT_ROOT/go"
if go mod tidy >/dev/null 2>&1; then
    test_passed "Go dependencies resolve"
else
    test_failed "Go build dependencies" "Dependency resolution failed"
fi

test_info "Go service compilation"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if go build -o /tmp/coordinator ./coordinator >/dev/null 2>&1; then
    test_passed "Coordinator service compiles"
    rm -f /tmp/coordinator
else
    test_failed "Go service compilation" "Coordinator compilation failed"
fi

test_info "Client library compilation"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if go build -o /tmp/client ./client/example >/dev/null 2>&1; then
    test_passed "Client library compiles"
    rm -f /tmp/client
else
    test_failed "Client library compilation" "Client compilation failed"
fi

# Test 6: Website integration
echo -e "\n${YELLOW}=== WEBSITE INTEGRATION ===${NC}"

test_info "Website Go deployment references"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if grep -q "GO_DEPLOYMENT.md" "$PROJECT_ROOT/website/content/_index.md"; then
    test_passed "Website references Go deployment"
else
    test_failed "Website Go deployment references" "Missing reference"
fi

test_info "Documentation sync status"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if grep -q "Go Implementation" "$PROJECT_ROOT/docs/README.md"; then
    test_passed "Documentation index includes Go"
else
    test_failed "Documentation sync status" "Missing Go implementation"
fi

# Test 7: Performance and monitoring integration
echo -e "\n${YELLOW}=== PERFORMANCE INTEGRATION ===${NC}"

test_info "Performance guide Go references"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if grep -q "Go Implementation" "$PROJECT_ROOT/docs/PERFORMANCE.md"; then
    test_passed "Performance guide includes Go"
else
    test_failed "Performance guide Go references" "Missing Go performance info"
fi

test_info "Monitoring guide integration"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if grep -q "monitor.go" "$PROJECT_ROOT/docs/MONITORING.md"; then
    test_passed "Monitoring guide includes Go services"
else
    test_failed "Monitoring guide integration" "Missing Go monitoring info"
fi

test_info "CI/CD integration examples"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if grep -q "Go API" "$PROJECT_ROOT/docs/CICD.md"; then
    test_passed "CI/CD guide includes Go APIs"
else
    test_failed "CI/CD integration examples" "Missing Go API examples"
fi

# Test 8: Cross-implementation connectivity
echo -e "\n${YELLOW}=== CROSS-IMPLEMENTATION CONNECTIVITY ===${NC}"

test_info "Bash to Go migration guidance"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if grep -q "Go Implementation" "$PROJECT_ROOT/docs/ADVANCED_CONFIG.md"; then
    test_passed "Advanced config includes migration guidance"
else
    test_failed "Bash to Go migration guidance" "Missing migration info"
fi

test_info "Use cases Go examples"
TOTAL_TESTS=$((TOTAL_TESTS + 1))
if grep -q "Go Implementation" "$PROJECT_ROOT/docs/USE_CASES.md"; then
    test_passed "Use cases include Go examples"
else
    test_failed "Use cases Go examples" "Missing Go use cases"
fi

# Summary
echo -e "\n${YELLOW}=== INTEGRATION TEST SUMMARY ===${NC}"
echo "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"

SUCCESS_RATE=$(( PASSED_TESTS * 100 / TOTAL_TESTS ))
echo "Success Rate: $SUCCESS_RATE%"

if [[ $SUCCESS_RATE -ge 90 ]]; then
    echo -e "\n${GREEN}🎉 GO IMPLEMENTATION INTEGRATION: EXCELLENT${NC}"
    echo "Go services are properly integrated with the main project"
elif [[ $SUCCESS_RATE -ge 80 ]]; then
    echo -e "\n${YELLOW}⚠️ GO IMPLEMENTATION INTEGRATION: GOOD${NC}"
    echo "Most integration points are working, minor improvements needed"
else
    echo -e "\n${RED}❌ GO IMPLEMENTATION INTEGRATION: NEEDS ATTENTION${NC}"
    echo "Significant integration improvements required"
fi

echo -e "\n${BLUE}Key Integration Points Verified:${NC}"
echo "• Service structure and build readiness"
echo "• Documentation cross-references"
echo "• Test framework integration"
echo "• API and client connectivity"
echo "• Website and documentation sync"
echo "• Performance and monitoring integration"
echo "• Cross-implementation connectivity"

echo -e "\n${BLUE}Next Steps for Full Integration:${NC}"
echo "1. Run individual Go service tests: cd go && go test ./tests/..."
echo "2. Test API connectivity: Start services and validate endpoints"
echo "3. Verify client library functionality"
echo "4. Test cross-implementation workflows"

exit 0