# API Reference Guide

This guide describes the programmatic interfaces and configuration options for the Distributed Gradle Building System, covering both Bash and Go implementations.

---

## Overview

The system provides two API interfaces:

1. **Bash Scripts API** - Script-based automation with environment variables
2. **Go Services API** - RESTful HTTP APIs for programmatic integration

---

## Go Services API

The Go implementation provides a comprehensive RESTful API for managing distributed builds.

### Base URLs

- **Coordinator**: `http://localhost:8080/api`
- **Cache Server**: `http://localhost:8083/cache`
- **Monitor**: `http://localhost:8084/api`

### Authentication

If authentication is enabled (configurable), include the token in requests:

```bash
curl -H "X-Auth-Token: your-auth-token" http://localhost:8080/api/status
```

### Coordinator API

#### Submit Build Request

```http
POST /api/build
Content-Type: application/json

{
  "project_path": "/absolute/path/to/project",
  "task_name": "build",
  "cache_enabled": true,
  "build_options": {
    "parallel": "true",
    "max_workers": "4"
  }
}
```

**Response**:
```json
{
  "build_id": "build-1234567890",
  "status": "queued"
}
```

#### Get Build Status

```http
GET /api/build/{build_id}
```

**Response**:
```json
{
  "build_id": "build-1234567890",
  "status": "completed",
  "worker_id": "worker-001",
  "duration": "45.2s",
  "success": true,
  "artifacts": ["/path/to/output.jar"]
}
```

#### Get Worker Status

```http
GET /api/workers
```

**Response**:
```json
{
  "worker-001": {
    "id": "worker-001",
    "host": "worker1.example.com",
    "status": "idle",
    "cpu_usage": 15.2,
    "memory_usage": 1024,
    "build_count": 42
  }
}
```

#### Get System Status

```http
GET /api/status
```

**Response**:
```json
{
  "timestamp": "2023-06-15T10:30:00Z",
  "worker_count": 3,
  "queue_length": 2,
  "active_builds": 1
}
```

### Cache Server API

#### Store Cache Entry

```http
PUT /cache/{key}
Content-Type: application/json

{
  "data": "base64-encoded-data",
  "hash": "sha256-hash",
  "ttl": 86400,
  "metadata": {
    "project": "my-app",
    "version": "1.0.0"
  }
}
```

#### Retrieve Cache Entry

```http
GET /cache/{key}
```

**Response**:
```json
{
  "found": true,
  "data": "base64-encoded-data",
  "metadata": {
    "project": "my-app",
    "version": "1.0.0",
    "created_at": "2023-06-15T10:30:00Z"
  }
}
```

#### Get Cache Statistics

```http
GET /stats
```

**Response**:
```json
{
  "entries": 1234,
  "hit_count": 5678,
  "miss_count": 123,
  "hit_rate": 0.97,
  "storage_size": 1073741824,
  "max_cache_size": 10737418240
}
```

### Monitor API

#### Get System Metrics

```http
GET /api/metrics
```

**Response**:
```json
{
  "total_builds": 1000,
  "successful_builds": 950,
  "failed_builds": 50,
  "average_build_time": "2m30s",
  "total_workers": 5,
  "active_workers": 4,
  "cache_hit_rate": 0.85
}
```

#### Generate Performance Report

```http
GET /api/report?start=2023-06-14T00:00:00Z&end=2023-06-15T23:59:59Z
```

**Response**:
```json
{
  "timestamp": "2023-06-15T10:30:00Z",
  "time_range": {
    "start": "2023-06-14T00:00:00Z",
    "end": "2023-06-15T23:59:59Z"
  },
  "summary": {
    "total_builds": 250,
    "success_rate": 0.96,
    "average_build_time": "2m15s",
    "throughput_per_hour": 10.4,
    "cache_efficiency": 0.87
  },
  "recommendations": [
    "Consider adding more workers to handle peak load",
    "Cache hit rate is optimal"
  ]
}
```

#### Get Active Alerts

```http
GET /api/alerts
```

**Response**:
```json
[
  {
    "id": "alert-123",
    "severity": "warning",
    "message": "Worker response time exceeds threshold",
    "source": "monitor",
    "timestamp": "2023-06-15T10:30:00Z",
    "resolved": false
  }
]
```

#### Register Worker

```http
POST /api/register
Content-Type: application/json

{
  "id": "worker-004",
  "host": "worker4.example.com",
  "status": "idle"
}
```

### Health Check Endpoints

All services provide health check endpoints:

```http
GET /health
```

**Response**:
```json
{
  "status": "healthy",
  "uptime": "2h30m45s"
}
```

---

## Bash Scripts API

---

## Environment Variables

### Core Configuration

#### `.gradlebuild_env`
The primary configuration file that defines the distributed build environment.

**Location**: Project root directory
**Required**: Yes

**Variables**:
```bash
export WORKER_IPS="192.168.1.101 192.168.1.102 192.168.1.103"
export MAX_WORKERS=16
export PROJECT_DIR="/absolute/path/to/project"
export MASTER_IP="192.168.1.100"  # Optional
export PARALLEL_FACTOR=2  # Optional
```

**Variable Details**:

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `WORKER_IPS` | Yes | Space-separated list of worker IP addresses | `"192.168.1.101 192.168.1.102"` |
| `MAX_WORKERS` | Yes | Maximum number of parallel workers | `16` |
| `PROJECT_DIR` | Yes | Absolute path to project directory | `"/home/user/android-app"` |
| `MASTER_IP` | No | Master node IP address | `"192.168.1.100"` |
| `PARALLEL_FACTOR` | No | Parallelization multiplier | `2` |

### Gradle Properties

#### `gradle.properties`
Standard Gradle configuration with distributed build extensions.

**Cache Configuration**:
```properties
# Enable distributed caching
org.gradle.caching=true

# Remote cache settings
systemProp.org.gradle.caching.remote.enabled=true
systemProp.org.gradle.caching.remote.url=http://cache-server:8080/cache/
systemProp.org.gradle.caching.remote.push=true
systemProp.org.gradle.caching.remote.allow-insecure=true
```

**Performance Configuration**:
```properties
# Parallel execution
org.gradle.parallel=true
org.gradle.workers.max=16

# Memory optimization
org.gradle.jvmargs=-Xmx4g -XX:+UseG1GC

# Configuration cache
org.gradle.configuration-cache=true
```

---

## Core Scripts API

### sync_and_build.sh

Primary build orchestration script.

#### Signature
```bash
./sync_and_build.sh [gradle_task] [additional_options...]
```

#### Parameters

| Parameter | Required | Description | Default |
|-----------|----------|-------------|---------|
| `gradle_task` | No | Gradle task to execute | `assemble` |
| `additional_options` | No | Additional Gradle options | None |

#### Return Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Network connectivity error |
| 4 | Build failure |

#### Usage Examples

```bash
# Basic build
./sync_and_build.sh

# Specific task
./sync_and_build.sh assembleDebug

# With additional options
./sync_and_build.sh test --max-workers=8

# Clean build
./sync_and_build.sh clean assemble
```

#### Environment Requirements

**Required Files**:
- `.gradlebuild_env` - Build environment configuration
- `gradlew` - Gradle wrapper
- `gradle.properties` - Gradle configuration (recommended)

**Network Requirements**:
- Passwordless SSH to all workers
- Network connectivity between master and workers
- Firewall rules allowing SSH traffic

---

### Worker Management API

### worker_pool_manager.sh

#### Main Functions

##### `get_pool_workers(pool_name)`

**Description**: Get list of workers for a specific pool type

**Parameters**:
- `pool_name` (string): Pool type (compilation, testing, packaging, default)

**Returns**: Space-separated list of worker IPs

**Example**:
```bash
workers=$(get_pool_workers "compilation")
echo "Compilation workers: $workers"
```

##### `get_available_workers(pool_name)`

**Description**: Get list of healthy workers for a pool

**Parameters**:
- `pool_name` (string): Pool type

**Returns**: Space-separated list of available worker IPs

**Example**:
```bash
available=$(get_available_workers "testing")
for worker in $available; do
    echo "Available test worker: $worker"
done
```

##### `check_worker_health(worker_ip)`

**Description**: Check health status of a specific worker

**Parameters**:
- `worker_ip` (string): Worker IP address

**Returns**: Health status string (healthy, failed, overloaded)

**Example**:
```bash
status=$(check_worker_health "192.168.1.101")
if [[ "$status" == "healthy" ]]; then
    echo "Worker is healthy"
fi
```

#### Pool Configuration

##### `workers.conf`

Worker pool configuration file.

**Format**:
```bash
# Format: POOL_NAME="type:worker1 worker2 worker3"
COMPILATION_POOL="compilation:build-1 build-2 build-3"
TESTING_POOL="testing:test-1 test-2"
PACKAGING_POOL="packaging:build-5 build-6"
DEFAULT_POOL="default:build-1 build-2 build-3 build-4"

# Pool requirements
COMPILE_REQUIREMENTS="cpu:high memory:medium storage:medium"
TEST_REQUIREMENTS="cpu:medium memory:high storage:low"
PACKAGING_REQUIREMENTS="cpu:medium memory:medium storage:high"
```

---

### Cache Management API

### cache_manager.sh

#### Main Functions

##### `get_cache_size(cache_path)`

**Description**: Get size of cache directory in bytes

**Parameters**:
- `cache_path` (string): Path to cache directory

**Returns**: Size in bytes

**Example**:
```bash
size=$(get_cache_size "$HOME/.gradle-build-cache/local")
echo "Local cache size: $size bytes"
```

##### `check_cache_health(cache_path, cache_name)`

**Description**: Perform health check on cache

**Parameters**:
- `cache_path` (string): Path to cache directory
- `cache_name` (string): Descriptive cache name

**Returns**: 0 for healthy, 1 for unhealthy

**Example**:
```bash
if check_cache_health "$LOCAL_CACHE_DIR" "local"; then
    echo "Local cache is healthy"
fi
```

##### `format_bytes(bytes)`

**Description**: Convert bytes to human-readable format

**Parameters**:
- `bytes` (int): Size in bytes

**Returns**: Formatted string (e.g., "1.5G")

**Example**:
```bash
formatted=$(format_bytes 1073741824)
echo "Size: $formatted"  # Output: 1.0G
```

#### Cache Configuration

##### `cache_config`

Cache manager configuration file.

**Format**:
```bash
# Cache Manager Configuration
CACHE_SIZE="10g"
CACHE_TTL="30d"
CLEANUP_THRESHOLD="80"
REMOTE_ENABLED="false"
REMOTE_URL="http://localhost:8080/cache"
CACHE_ENABLED="true"
```

**Configuration Options**:

| Option | Description | Default | Format |
|---------|-------------|----------|---------|
| `CACHE_SIZE` | Maximum cache size | "10g" | Size with suffix (k, m, g) |
| `CACHE_TTL` | Cache entry time-to-live | "30d" | Duration with suffix (s, m, h, d) |
| `CLEANUP_THRESHOLD` | Cleanup trigger threshold | "80" | Percentage (0-100) |
| `REMOTE_ENABLED` | Enable remote cache | "false" | "true"/"false" |
| `REMOTE_URL` | Remote cache server URL | "http://localhost:8080/cache" | Valid HTTP URL |
| `CACHE_ENABLED` | Enable caching | "true" | "true"/"false" |

---

### Performance Analysis API

### performance_analyzer.sh

#### Main Functions

##### `collect_system_metrics()`

**Description**: Collect current system metrics

**Returns**: Comma-separated metrics string

**Format**: `cpu_usage,memory_usage,network_io,disk_io`

**Example**:
```bash
metrics=$(collect_system_metrics)
echo "System metrics: $metrics"
```

##### `parse_gradle_output(build_output)`

**Description**: Parse Gradle build output for performance metrics

**Parameters**:
- `build_output` (string): Gradle build output text

**Returns**: Comma-separated task metrics

**Format**: `tasks_executed,tasks_up_to_date,tasks_from_cache`

**Example**:
```bash
build_output=$(./gradlew assemble 2>&1)
task_metrics=$(parse_gradle_output "$build_output")
echo "Task metrics: $task_metrics"
```

#### Metrics Format

##### CSV Export Format

**File**: `performance_results/metrics.csv`

**Columns**:
```
timestamp,build_duration,worker_count,cache_enabled,tasks_executed,tasks_cached,cpu_usage,memory_usage,network_io,disk_io,exit_code
```

**Example Row**:
```
2023-12-18 14:30:15,125.5,8,true,45,38,75.2,68.1,1024000,512000,0
```

---

### Health Monitoring API

### health_checker.sh

#### Main Functions

##### `get_cpu_usage()`

**Description**: Get current CPU usage percentage

**Returns**: CPU usage as float

**Example**:
```bash
cpu=$(get_cpu_usage)
echo "CPU usage: ${cpu}%"
```

##### `get_memory_usage()`

**Description**: Get current memory usage percentage

**Returns**: Memory usage as float

**Example**:
```bash
memory=$(get_memory_usage)
echo "Memory usage: ${memory}%"
```

##### `check_worker_connectivity()`

**Description**: Check connectivity to all configured workers

**Returns**: Space-separated list of failed worker IPs

**Example**:
```bash
failed=$(check_worker_connectivity)
if [[ -n "$failed" ]]; then
    echo "Failed workers: $failed"
fi
```

#### Health Thresholds

##### `health_config`

Health monitoring configuration file.

**Format**:
```bash
CPU_WARNING=80
CPU_CRITICAL=95
MEMORY_WARNING=85
MEMORY_CRITICAL=95
DISK_WARNING=85
DISK_CRITICAL=95
BUILD_TIME_WARNING=1800
FAILURE_RATE_WARNING=20
```

---

## Configuration Templates

### Minimal Configuration

```bash
# .gradlebuild_env
export WORKER_IPS="192.168.1.101 192.168.1.102"
export MAX_WORKERS=8
export PROJECT_DIR="/home/user/my-project"
```

```properties
# gradle.properties
org.gradle.parallel=true
org.gradle.caching=true
org.gradle.workers.max=8
```

### Advanced Configuration

```bash
# .gradlebuild_env
export WORKER_IPS="build-1 build-2 build-3 build-4"
export MAX_WORKERS=16
export PROJECT_DIR="/opt/projects/android-app"
export MASTER_IP="192.168.1.100"
export PARALLEL_FACTOR=2
```

```properties
# gradle.properties
org.gradle.parallel=true
org.gradle.caching=true
org.gradle.configuration-cache=true
org.gradle.workers.max=16
org.gradle.jvmargs=-Xmx8g -XX:+UseG1GC -XX:MaxGCPauseMillis=200

# Remote cache
systemProp.org.gradle.caching.remote.enabled=true
systemProp.org.gradle.caching.remote.url=http://cache-server:8080/cache/
systemProp.org.gradle.caching.remote.push=true

# Performance tuning
org.gradle.daemon=true
org.gradle.daemon.idletimeout=3600000
org.gradle.vfs.watch=true
```

```bash
# workers.conf
COMPILATION_POOL="compilation:build-1 build-2 build-3 build-4"
TESTING_POOL="testing:test-1 test-2"
PACKAGING_POOL="packaging:build-5 build-6"

COMPILE_REQUIREMENTS="cpu:high memory:medium storage:medium"
TEST_REQUIREMENTS="cpu:medium memory:high storage:low"
PACKAGING_REQUIREMENTS="cpu:medium memory:medium storage:high"
```

```bash
# cache_config
CACHE_SIZE="20g"
CACHE_TTL="7d"
CLEANUP_THRESHOLD="75"
REMOTE_ENABLED="true"
REMOTE_URL="http://cache-server:8080/cache"
CACHE_ENABLED="true"
```

---

## Integration Examples

### Python Integration

```python
import subprocess
import json

def run_distributed_build(task, workers=None):
    """Execute distributed build using Python integration"""
    
    # Build command
    cmd = ["./sync_and_build.sh", task]
    if workers:
        cmd.extend(["--max-workers", str(workers)])
    
    # Execute build
    result = subprocess.run(cmd, capture_output=True, text=True)
    
    return {
        'success': result.returncode == 0,
        'output': result.stdout,
        'error': result.stderr,
        'exit_code': result.returncode
    }

def get_worker_status():
    """Get worker pool status"""
    
    # Use worker pool manager to get status
    result = subprocess.run(
        ["./scripts/worker_pool_manager.sh", "status"],
        capture_output=True, text=True
    )
    
    return {
        'status': result.stdout,
        'success': result.returncode == 0
    }

def check_system_health():
    """Check system health"""
    
    # Use health checker
    result = subprocess.run(
        ["./scripts/health_checker.sh", "check"],
        capture_output=True, text=True
    )
    
    return {
        'healthy': result.returncode == 0,
        'report': result.stdout
    }

# Example usage
if __name__ == "__main__":
    # Check health before build
    health = check_system_health()
    if not health['healthy']:
        print("System health check failed")
        exit(1)
    
    # Run build
    build_result = run_distributed_build("assembleDebug")
    if build_result['success']:
        print("Build completed successfully")
    else:
        print(f"Build failed: {build_result['error']}")
```

### Shell Integration

```bash
#!/bin/bash
# build_pipeline.sh - Example build pipeline

# Health check
echo "=== Health Check ==="
if ! ./scripts/health_checker.sh check; then
    echo "Health check failed - aborting"
    exit 1
fi

# Optimize cache
echo "=== Cache Optimization ==="
./scripts/cache_manager.sh optimize

# Get optimal workers
echo "=== Worker Selection ==="
workers=$(./scripts/worker_pool_manager.sh detect assembleDebug)
echo "Optimal pool: $workers"

# Execute build
echo "=== Build Execution ==="
./scripts/worker_pool_manager.sh execute assembleDebug auto "$workers"

# Post-build analysis
echo "=== Performance Analysis ==="
./scripts/performance_analyzer.sh analyze

# Generate report
echo "=== Health Report ==="
./scripts/health_checker.sh report

echo "Pipeline completed"
```

### CI/CD Integration (Jenkins)

```groovy
pipeline {
    agent { label 'gradle-build' }
    
    environment {
        WORKER_IPS = 'build-worker-1 build-worker-2 build-worker-3'
        MAX_WORKERS = '8'
        PROJECT_DIR = "${WORKSPACE}"
    }
    
    stages {
        stage('Health Check') {
            steps {
                sh './scripts/health_checker.sh check'
            }
        }
        
        stage('Cache Prep') {
            steps {
                sh './scripts/cache_manager.sh optimize'
                sh './scripts/cache_manager.sh configure'
            }
        }
        
        stage('Build') {
            steps {
                sh '''
                    ./scripts/worker_pool_manager.sh execute assembleDebug 8 compilation
                '''
            }
        }
        
        stage('Test') {
            parallel {
                stage('Unit Tests') {
                    steps {
                        sh '''
                            ./scripts/worker_pool_manager.sh execute test 4 testing
                        '''
                    }
                }
                stage('Integration Tests') {
                    steps {
                        sh '''
                            ./scripts/worker_pool_manager.sh execute connectedAndroidTest 2 testing
                        '''
                    }
                }
            }
        }
    }
    
    post {
        always {
            sh '''
                ./scripts/performance_analyzer.sh analyze
                ./scripts/health_checker.sh report
            '''
        }
        
        success {
            sh '''
                ./scripts/cache_manager.sh sync
            '''
        }
    }
}
```

---

## Error Handling

### Common Error Codes

| Code | Description | Resolution |
|------|-------------|------------|
| 1 | General error | Check logs and configuration |
| 2 | Configuration missing | Create `.gradlebuild_env` |
| 3 | Network error | Check SSH connectivity |
| 4 | Build failure | Review build logs |
| 5 | Resource exhaustion | Check system resources |

### Debug Mode

Enable debug output by setting environment variable:

```bash
export DEBUG=1
./sync_and_build.sh assembleDebug
```

### Log Locations

| Component | Log Location |
|-----------|--------------|
| Main build | `build/logs/build.log` |
| Cache manager | `$HOME/.gradle-build-cache/cache_manager.log` |
| Health checker | `health_check.log` |
| Performance analyzer | `performance_results/analyzer.log` |
| Worker pool manager | `.worker_locks/` |

---

This API reference provides comprehensive information for integrating with the Distributed Gradle Building System programmatically. All interfaces are designed to be script-friendly and easily integrable with CI/CD systems.