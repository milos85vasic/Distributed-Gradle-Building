# Makefile for Distributed Gradle Building System Test Suite
# Provides automation commands for running different types of tests

.PHONY: help test test-unit test-integration test-performance test-security test-load test-e2e test-all setup cleanup docker-setup docker-teardown clean

# Default target
help:
	@echo "Distributed Gradle Building System - Test Suite"
	@echo ""
	@echo "Usage: make [TARGET]"
	@echo ""
	@echo "Test Targets:"
	@echo "  test          Run default test suite (unit + integration)"
	@echo "  test-unit     Run unit tests only"
	@echo "  test-integration    Run integration tests only"
	@echo "  test-performance    Run performance tests only"
	@echo "  test-security       Run security tests only"
	@echo "  test-load           Run load tests only"
	@echo "  test-e2e            Run end-to-end tests only"
	@echo "  test-all            Run all test suites"
	@echo ""
	@echo "Environment Targets:"
	@echo "  setup               Setup test environment"
	@echo "  cleanup             Cleanup test environment"
	@echo "  docker-setup        Start test Docker containers"
	@echo "  docker-teardown     Stop test Docker containers"
	@echo "  clean               Clean all generated files"
	@echo ""
	@echo "Report Targets:"
	@echo "  report-html         Generate HTML test report"
	@echo "  report-junit        Generate JUnit XML report"
	@echo "  report-coverage     Generate coverage report"
	@echo ""
	@echo "Development Targets:"
	@echo "  lint                Run linting on test files"
	@echo "  format              Format test files"
	@echo "  deps                Install test dependencies"

# Test execution
test: test-unit test-integration

test-unit:
	@echo "Running unit tests..."
	@./test_framework.sh unit

test-integration:
	@echo "Running integration tests..."
	@./test_framework.sh integration
	@cd integration && ./test_end_to_end.sh

test-performance:
	@echo "Running performance tests..."
	@./test_framework.sh performance
	@cd performance && ./test_coordinator_performance.sh

test-security:
	@echo "Running security tests..."
	@./test_framework.sh security
	@cd security && ./test_authentication_security.sh

test-load:
	@echo "Running load tests..."
	@./test_framework.sh load

test-e2e:
	@echo "Running end-to-end tests..."
	@./test_framework.sh e2e

test-all:
	@echo "Running all test suites..."
	@./test_framework.sh all

# Environment management
setup:
	@echo "Setting up test environment..."
	@mkdir -p results logs fixtures/demo_projects fixtures/test_data fixtures/mock_services
	@./test_framework.sh setup
	@if [ ! -f fixtures/mock_services/responses.json ]; then \
		echo '{}' > fixtures/mock_services/responses.json; \
	fi

cleanup:
	@echo "Cleaning up test environment..."
	@./test_framework.sh cleanup
	@find . -name "*.tmp" -delete 2>/dev/null || true

docker-setup:
	@echo "Starting test Docker environment..."
	@if command -v docker-compose >/dev/null 2>&1; then \
		docker-compose -f test_docker-compose.yml up -d; \
		sleep 10; \
		echo "Waiting for services to be ready..."; \
		for port in 8080 8081 8082; do \
			echo "Checking port $$port..."; \
			timeout 30 bash -c "until curl -f http://localhost:$$port/health; do sleep 1; done"; \
		done; \
		echo "Docker environment ready"; \
	else \
		echo "docker-compose not found. Please install Docker Compose."; \
		exit 1; \
	fi

docker-teardown:
	@echo "Stopping test Docker environment..."
	@if command -v docker-compose >/dev/null 2>&1; then \
		docker-compose -f test_docker-compose.yml down; \
		docker-compose -f test_docker-compose.yml rm -f; \
	else \
		echo "docker-compose not found. Please install Docker Compose."; \
		exit 1; \
	fi

clean: cleanup
	@echo "Cleaning all generated files..."
	@rm -rf results/* logs/* fixtures/demo_projects/* fixtures/test_data/*
	@find . -name "*.log" -delete 2>/dev/null || true
	@find . -name "*.out" -delete 2>/dev/null || true

# Report generation
report-html:
	@echo "Generating HTML test report..."
	@if [ -f results/test_results.json ]; then \
		if command -v jq >/dev/null 2>&1; then \
			jq '.' results/test_results.json > results/test_results_formatted.json; \
			echo "HTML report generation not yet implemented"; \
		else \
			echo "jq not found. Please install jq for JSON processing."; \
		fi; \
	else \
		echo "No test results found. Run tests first."; \
	fi

report-junit:
	@echo "Generating JUnit XML report..."
	@if [ -f results/test_results.json ]; then \
		echo "JUnit report generation not yet implemented"; \
	else \
		echo "No test results found. Run tests first."; \
	fi

report-coverage:
	@echo "Generating coverage report..."
	@cd ../go && go test -coverprofile=../tests/results/coverage.out ./...
	@if [ -f results/coverage.out ]; then \
		go tool cover -html=results/coverage.out -o results/coverage.html; \
		echo "Coverage report generated: results/coverage.html"; \
	else \
		echo "Coverage report generation failed"; \
	fi

# Development utilities
lint:
	@echo "Running linting on test files..."
	@if command -v shellcheck >/dev/null 2>&1; then \
		shellcheck test_framework.sh unit/*.sh integration/*.sh performance/*.sh security/*.sh load/*.sh; \
	else \
		echo "shellcheck not found. Please install ShellCheck."; \
	fi

format:
	@echo "Formatting test files..."
	@if command -v shfmt >/dev/null 2>&1; then \
		shfmt -w -i 4 test_framework.sh unit/*.sh integration/*.sh performance/*.sh security/*.sh load/*.sh; \
	else \
		echo "shfmt not found. Please install shfmt."; \
	fi

deps:
	@echo "Installing test dependencies..."
	@if command -v jq >/dev/null 2>&1; then \
		echo "jq is already installed"; \
	else \
		echo "Installing jq..."; \
		if command -v brew >/dev/null 2>&1; then \
			brew install jq; \
		elif command -v apt-get >/dev/null 2>&1; then \
			sudo apt-get update && sudo apt-get install -y jq; \
		elif command -v yum >/dev/null 2>&1; then \
			sudo yum install -y jq; \
		else \
			echo "Please install jq manually"; \
		fi; \
	fi
	@if command -v curl >/dev/null 2>&1; then \
		echo "curl is already installed"; \
	else \
		echo "Please install curl for HTTP testing"; \
	fi

# CI/CD helpers
ci-test:
	@echo "Running CI test suite..."
	@make setup
	@make test-all
	@make report-junit
	@make cleanup

ci-performance:
	@echo "Running CI performance tests..."
	@make docker-setup
	@make test-performance
	@make docker-teardown

ci-security:
	@echo "Running CI security tests..."
	@make docker-setup
	@make test-security
	@make docker-teardown

# Development helper targets
watch-test:
	@echo "Watching for changes and running tests..."
	@if command -v inotifywait >/dev/null 2>&1; then \
		while inotifywait -e modify -r . --include='.*\.(sh|go|js|json)$$'; do \
			echo "Changes detected, running tests..."; \
			make test-unit; \
		done; \
	else \
		echo "inotifywait not found. Please install inotify-tools."; \
	fi

quick-test:
	@echo "Running quick tests (unit tests only)...";
	@TEST_TIMEOUT=30 make test-unit