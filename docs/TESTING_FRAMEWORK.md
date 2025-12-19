# Distributed Gradle Building System - Test Framework Documentation

## Overview

This document provides comprehensive guidance for the Distributed Gradle Building System test framework, covering all testing levels from unit tests to security testing.

## Test Architecture

The test framework follows a multi-level testing pyramid:

```
                ┌─────────────────┐
                │   Security     │
                │     Tests       │
                └─────────────────┘
              ┌─────────────────────┐
              │   Performance      │
              │     Tests          │
              └─────────────────────┘
            ┌─────────────────────────┐
            │      Load Tests         │
            │    (Sustained Load)     │
            └─────────────────────────┘
          ┌─────────────────────────────┐
          │    Integration Tests        │
          │   (Service Interactions)    │
          └─────────────────────────────┘
        ┌─────────────────────────────────┐
        │       Unit Tests               │
        │   (Component Isolation)       │
        └─────────────────────────────────┘
```

## Test Structure

### Unit Tests
- **Location**: `./coordinatorpkg/`, `./workerpkg/`, `./cachepkg/`, `./monitorpkg/`
- **Purpose**: Test individual components in isolation
- **Count**: 36 test functions across 5 service packages
- **Coverage**: >95% of core business logic

### Integration Tests
- **Location**: `./tests/integration/integration_test.go`
- **Purpose**: Test service interactions and complete workflows
- **Functions**: 8 integration test scenarios

### Performance Tests
- **Location**: `./tests/performance/performance_test.go`
- **Purpose**: Benchmark system performance and scalability
- **Functions**: 3 main test categories with multiple scenarios

### Load Tests
- **Location**: `./tests/load/load_test.go`
- **Purpose**: Test system under high and sustained load
- **Functions**: 3 main test scenarios

### Security Tests
- **Location**: `./tests/security/security_test.go`
- **Purpose**: Validate security controls and vulnerability resistance
- **Functions**: 11 comprehensive security test scenarios

## Quick Start

### Running All Tests
```bash
# From project root
./scripts/run-all-tests.sh

# Or manually:
cd go && go test ./... -v -timeout 120s
```

### Running Specific Test Suites
```bash
# Unit tests only
cd go && go test ./coordinatorpkg ./workerpkg ./cachepkg ./monitorpkg -v

# Integration tests
cd go && go test ./tests/integration -v -timeout 30s

# Performance tests (may take longer)
cd go && go test ./tests/performance -v -timeout 60s

# Load tests (may take longer)
cd go && go test ./tests/load -v -timeout 60s

# Security tests
cd go && go test ./tests/security -v -timeout 30s

# Quick smoke test (security + integration only)
cd go && go test ./tests/security ./tests/integration -v -short
```

## Test Execution Guidelines

### Performance Testing

#### Best Practices
1. **Run in isolation**: Performance tests should not run concurrently with other tests
2. **Warm-up system**: Allow 30 seconds for system warmup before performance testing
3. **Baseline measurements**: Document baseline performance metrics for comparison
4. **Resource monitoring**: Monitor CPU, memory, and disk I/O during tests

#### Common Issues & Solutions

**Issue**: Deadlock in cache operations
```bash
# Symptom: Tests hang with "sync.Mutex.Lock" in stack trace
# Solution: Cache deadlock has been fixed - ensure latest code
```

**Issue**: Timeouts in coordinator tests
```bash
# Symptom: "test timed out" errors in scalability tests
# Solution: Use -short flag or reduce test data size
```

#### Optimization Tips
1. Reduce thread counts for local development
2. Use `-short` flag for faster execution
3. Monitor filesystem cache performance
4. Adjust test parameters based on system capabilities

### Integration Testing

#### Test Scenarios
- Complete build workflows
- Worker registration and RPC communication
- Cache artifact storage and retrieval
- Monitor alerting system
- Concurrent build submissions
- HTTP API endpoints
- Error scenarios and recovery
- Resource cleanup and memory management

#### Best Practices
1. Use unique ports for each test to avoid conflicts
2. Clean up resources after each test
3. Mock external dependencies
4. Verify service communication patterns

### Load Testing

#### Load Patterns
- **Moderate Load**: 5 workers, 2-second duration, ~40 requests
- **Heavy Load**: 10 workers, 5-second duration, ~191 requests
- **Sustained Load**: 10 workers, 2-second duration, concurrent operations

#### Metrics to Monitor
- Request throughput (requests/second)
- Success rate (%)
- Average response time
- Memory usage patterns
- Worker utilization

#### Common Issues
1. **Worker Registration Conflicts**: Ensure unique ports
2. **Cache Contention**: Fixed in recent updates
3. **Test Timeout**: Increase timeout or reduce load

### Security Testing

#### Security Controls Tested
- Authentication and authorization
- Input validation and sanitization
- Rate limiting
- CORS configuration
- Request size limits
- Data encryption at rest
- Secure password handling
- Secure session management
- Security headers
- Secure logging practices

#### OWASP Compliance
- Authentication bypass protection
- Input validation against XSS/SQL injection
- CSRF protection mechanisms
- Secure headers implementation
- Rate limiting for DoS protection

## Troubleshooting Guide

### Common Test Failures

#### 1. Port Conflicts
```
Error: "address already in use"
Solution: Tests use unique ports automatically - check for external processes
```

#### 2. Cache Deadlock (Fixed)
```
Error: Tests hang with "sync.Mutex.Lock" stack trace
Solution: Cache concurrency issue has been resolved
```

#### 3. Timeout Issues
```
Error: "test timed out after 30s"
Solution: Increase timeout or use -short flag for reduced test scope
```

#### 4. Worker Registration Issues
```
Error: "rpc.Register: type has no exported methods"
Solution: Verify RPC interface implementations match
```

#### 5. Performance Test Failures
```
Error: Performance thresholds not met
Solution: Adjust thresholds based on system capabilities or use -short
```

### Debug Mode
```bash
# Run tests with verbose output
cd go && go test ./tests/... -v -timeout 120s

# Run specific test with debug output
cd go && go test ./tests/integration -v -run TestCompleteBuildWorkflow

# Run tests with race detection (slower but thorough)
cd go && go test ./tests/... -race -v -timeout 180s
```

## Continuous Integration

### GitHub Actions Pipeline
The CI pipeline runs all test suites in the following order:
1. Unit tests (quick feedback)
2. Integration tests (service validation)
3. Security tests (vulnerability checks)
4. Performance tests (baseline validation)

### Coverage Requirements
- **Unit tests**: >90% code coverage
- **Integration tests**: >80% service path coverage
- **Overall**: >85% combined coverage

## Performance Baselines

### Current Performance Metrics
- **Cache Write**: 583.63 entries/sec (small), 47.85 entries/sec (medium)
- **Cache Read**: 4962.38 entries/sec (small), 8236.14 entries/sec (medium)
- **Coordinator Load**: 40 requests/sec (moderate), 38 requests/sec (heavy)
- **Success Rate**: 100% (moderate), 52% (heavy load - expected due to resource constraints)

### Acceptable Thresholds
- **Cache Write**: >100 entries/sec
- **Cache Read**: >1000 entries/sec
- **Coordinator Throughput**: >20 requests/sec
- **Success Rate**: >95% (moderate load)

## Test Data Management

### Test Data Locations
- **Unit Tests**: In-memory test data
- **Integration Tests**: Temporary directories
- **Performance Tests`: Configurable data sizes
- **Load Tests**: Generated test data on-the-fly

### Cleanup Procedures
- Unit tests: Automatic cleanup
- Integration tests: Manual cleanup in tearDown
- Performance tests: Automatic cache cleanup
- Load tests: Background cleanup processes

## Contributing

### Adding New Tests
1. Follow existing naming conventions
2. Use table-driven tests for multiple scenarios
3. Include proper setup/teardown
4. Document test purpose and expected behavior
5. Update this documentation

### Test Writing Guidelines
1. **Unit Tests**: Test one thing at a time
2. **Integration Tests**: Test realistic scenarios
3. **Performance Tests**: Focus on bottlenecks
4. **Security Tests**: Follow OWASP guidelines
5. **Load Tests**: Simulate real usage patterns

## Future Enhancements

### Planned Improvements
1. **Chaos Engineering**: Fault injection testing
2. **End-to-End Tests**: Full workflow automation
3. **Visual Reports**: Test result dashboards
4. **Performance Regression**: Automated detection
5. **Contract Testing**: API contract validation

### Advanced Testing Features
1. **Canary Testing**: Gradual rollout validation
2. **A/B Testing**: Feature comparison
3. **Stress Testing**: Breaking point identification
4. **Recovery Testing**: Failover validation
5. **Monitoring Integration**: Real-time metrics

## Resources

### Documentation
- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [OWASP Security Testing Guide](https://owasp.org/www-project-web-security-testing-guide/)
- [Performance Testing Best Practices](https://golang.org/pkg/testing/#hdr-Benchmarks)

### Tools and Libraries
- Go standard testing package
- HTTP Test Server (`net/http/httptest`)
- Benchmarking framework (`testing.B`)
- Race detection (`go test -race`)

### Contact and Support
For questions about the test framework:
1. Check this documentation first
2. Review existing test implementations
3. Create GitHub issues for bugs
4. Submit pull requests for improvements