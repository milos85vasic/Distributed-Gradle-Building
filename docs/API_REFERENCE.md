# AI-Powered Distributed Gradle Building - API Reference

## Overview

This document provides comprehensive API documentation for all services in the AI-powered distributed Gradle building system.

## Service Endpoints

| Service | Base URL | Purpose |
|---------|----------|---------|
| Coordinator | `http://localhost:8080` | Build orchestration and queue management |
| ML Service | `http://localhost:8082` | Machine learning predictions and continuous learning |
| Monitor | `http://localhost:8084` | System monitoring and metrics |
| Cache | `http://localhost:8085` | Distributed caching service |
| Workers | `http://localhost:8087-8089` | Build execution nodes |

## Coordinator Service API

### Build Management

#### Submit Build
**POST** `/api/build`

Submit a new build request to the distributed system.

**Request Body:**
```json
{
  "project_path": "/path/to/gradle/project",
  "task_name": "build",
  "worker_id": "worker-1",
  "cache_enabled": true,
  "build_options": {
    "gradle_version": "7.6",
    "java_version": "11"
  }
}
```

**Response:**
```json
{
  "build_id": "build-1640995200"
}
```

**Status Codes:**
- `200` - Build queued successfully
- `400` - Invalid request
- `500` - Internal server error

#### Get Build Status
**GET** `/api/builds/{build_id}`

Retrieve the status and details of a specific build.

**Response:**
```json
{
  "request_id": "build-1640995200",
  "success": true,
  "worker_id": "worker-1",
  "build_duration": 45000000000,
  "artifacts": ["build/libs/app.jar", "build/reports/tests/test/index.html"],
  "error_message": "",
  "metrics": {
    "build_steps": [
      {
        "name": "compileJava",
        "duration": 15000000000,
        "status": "success",
        "start_time": "2023-12-31T12:00:00Z",
        "end_time": "2023-12-31T12:00:15Z"
      }
    ],
    "cache_hit_rate": 0.75,
    "compiled_files": 45,
    "test_results": {
      "total": 150,
      "passed": 148,
      "failed": 2,
      "skipped": 0,
      "duration": 30000000000
    },
    "resource_usage": {
      "cpu_usage": 0.85,
      "memory_usage": 2147483648,
      "disk_io": 104857600,
      "network_io": 52428800
    }
  },
  "timestamp": "2023-12-31T12:00:45Z"
}
```

### Worker Management

#### List Workers
**GET** `/api/workers`

Retrieve information about all registered workers.

**Response:**
```json
[
  {
    "id": "worker-1",
    "host": "worker-1",
    "port": 8087,
    "status": "active",
    "capabilities": ["gradle", "maven", "java-8", "java-11", "java-17"],
    "last_ping": "2023-12-31T12:00:30Z",
    "builds": [
      {
        "request_id": "build-1640995200",
        "project_path": "/projects/myapp",
        "task_name": "build",
        "timestamp": "2023-12-31T12:00:00Z"
      }
    ],
    "metrics": {
      "build_count": 150,
      "total_build_time": 7200000000000,
      "average_build_time": 48000000000,
      "success_rate": 0.95,
      "last_build_time": "2023-12-31T12:00:00Z"
    }
  }
]
```

### System Health

#### Health Check
**GET** `/api/health`

Check the health status of the coordinator service.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2023-12-31T12:00:30Z",
  "version": "1.0.0"
}
```

## ML Service API

### Build Predictions

#### Get Build Insights
**POST** `/api/predict`

Get comprehensive ML-driven insights for a build request.

**Request Body:**
```json
{
  "project_path": "/projects/myapp",
  "task_name": "build",
  "build_options": {
    "gradle_version": "7.6"
  }
}
```

**Response:**
```json
{
  "build_id": "prediction_1704028800",
  "predicted_time": 45000000000,
  "confidence": 0.85,
  "resource_needs": {
    "cpu": 0.8,
    "memory": 0.75,
    "disk": 0.4
  },
  "scaling_advice": {
    "action": "maintain",
    "workers_needed": 3,
    "confidence": 0.9,
    "reason": "Current load within optimal range"
  },
  "failure_risk": 0.15,
  "cache_hit_rate": 0.72
}
```

#### Scaling Recommendations
**GET** `/api/scaling?queue_length={n}&cpu_load={f}&current_workers={n}`

Get scaling recommendations based on current system load.

**Parameters:**
- `queue_length` (int): Current build queue length
- `cpu_load` (float): Average CPU utilization (0.0-1.0)
- `current_workers` (int): Current number of active workers

**Response:**
```json
{
  "action": "scale_up",
  "workers_needed": 5,
  "confidence": 0.85,
  "reason": "High queue length indicates need for more workers"
}
```

### Model Management

#### Train Models
**POST** `/api/train`

Manually trigger ML model retraining with current data.

**Response:**
```json
{
  "status": "training_completed",
  "models_trained": ["build_time", "resource", "failure", "cache"],
  "training_duration": 300000000000,
  "new_accuracy": 0.82
}
```

**Status Codes:**
- `200` - Training completed successfully
- `400` - Insufficient data for training
- `500` - Training failed

#### Get Learning Statistics
**GET** `/api/learning`

Retrieve continuous learning statistics and configuration.

**Response:**
```json
{
  "config": {
    "enabled": true,
    "retraining_interval": 86400000000000,
    "data_collection_interval": 300000000000,
    "min_data_points": 50,
    "performance_threshold": 0.7,
    "coordinator_host": "coordinator",
    "coordinator_port": 8080,
    "monitor_host": "monitor",
    "monitor_port": 8084
  },
  "stats": {
    "last_data_collection": "2023-12-31T11:55:00Z",
    "last_retraining": "2023-12-30T12:00:00Z",
    "data_points_collected": 1250,
    "retraining_count": 5,
    "average_accuracy": 0.83,
    "last_accuracy_check": "2023-12-31T06:00:00Z",
    "is_collecting": false,
    "is_retraining": false,
    "model_backups": [
      {
        "timestamp": "2023-12-30T12:00:00Z",
        "accuracy": 0.81,
        "version": "v5.1704028800"
      }
    ],
    "current_version": "v5.1704115200"
  }
}
```

#### Rollback Models
**POST** `/api/rollback`

Rollback to a previous model version.

**Request Body:**
```json
{
  "version": "v5.1704028800"
}
```

**Response:**
```json
{
  "status": "rollback_completed",
  "version": "v5.1704028800"
}
```

### Data Management

#### Export ML Data
**GET** `/api/export`

Export all ML service data for backup or analysis.

**Response Headers:**
```
Content-Type: application/json
Content-Disposition: attachment; filename=ml-data.json
```

**Response Body:** Complete JSON export of all ML data, models, and statistics.

#### Import ML Data
**POST** `/api/import`

Import ML data from a backup file.

**Request Body:** JSON data exported via `/api/export`

**Response:**
```json
{
  "status": "import_completed",
  "data_points_imported": 1250,
  "models_loaded": 4
}
```

### Statistics

#### Get ML Statistics
**GET** `/api/stats`

Retrieve general ML service statistics.

**Response:**
```json
{
  "total_build_records": 1250,
  "total_worker_metrics": 5000,
  "total_cache_metrics": 2500,
  "oldest_build": "2023-11-01T00:00:00Z",
  "newest_build": "2023-12-31T12:00:00Z",
  "models_trained": {
    "build_time": true,
    "resource": true,
    "scaling": true,
    "failure": true,
    "cache": true
  }
}
```

## Monitor Service API

### System Metrics

#### Get System Metrics
**GET** `/api/metrics`

Retrieve comprehensive system metrics from all services.

**Response:**
```json
{
  "workers": {
    "worker-1": {
      "id": "worker-1",
      "status": "active",
      "cpu_usage": 0.75,
      "memory_usage": 0.65,
      "active_builds": 2,
      "completed_builds": 150,
      "failed_builds": 5,
      "last_seen": "2023-12-31T12:00:30Z",
      "average_build_time": 45.5
    }
  },
  "builds": {
    "build-1640995200": {
      "build_id": "build-1640995200",
      "worker_id": "worker-1",
      "status": "completed",
      "start_time": "2023-12-31T12:00:00Z",
      "end_time": "2023-12-31T12:00:45Z",
      "duration": 45000000000,
      "cache_hit": true,
      "artifacts": ["build/libs/app.jar"],
      "resource_usage": {
        "cpu_percent": 0.85,
        "memory_usage": 2147483648,
        "disk_io": 104857600,
        "network_io": 52428800
      }
    }
  },
  "system": {
    "total_builds": 500,
    "successful_builds": 475,
    "failed_builds": 25,
    "active_workers": 3,
    "queue_length": 2,
    "average_build_time": 42.3,
    "cache_hit_rate": 0.78,
    "last_update": "2023-12-31T12:00:30Z",
    "resource_usage": {
      "cpu_percent": 0.65,
      "memory_usage": 4294967296,
      "disk_io": 209715200,
      "network_io": 104857600
    }
  }
}
```

### Health Monitoring

#### Health Check
**GET** `/health`

Basic health check for the monitor service.

**Response:**
```json
{
  "status": "healthy",
  "service": "monitor",
  "timestamp": "2023-12-31T12:00:30Z"
}
```

#### Basic Metrics
**GET** `/metrics`

Basic service metrics (Prometheus format).

**Response:**
```
# HELP monitor_workers_total Total number of registered workers
# TYPE monitor_workers_total gauge
monitor_workers_total 3

# HELP monitor_builds_total Total number of builds tracked
# TYPE monitor_builds_total gauge
monitor_builds_total 500
```

## Cache Service API

### Cache Operations

#### Health Check
**GET** `/health`

Check cache service health.

**Response:**
```json
{
  "status": "healthy",
  "service": "cache",
  "uptime": "24h30m45s"
}
```

#### Cache Statistics
**GET** `/stats`

Retrieve cache performance statistics.

**Response:**
```json
{
  "total_entries": 1250,
  "cache_size_bytes": 5368709120,
  "hit_rate": 0.78,
  "hits": 9750,
  "misses": 2750,
  "evictions": 125,
  "compression_ratio": 0.65
}
```

## Worker Service API

### Build Execution

#### Execute Build
**RPC Call** `WorkerService.Build`

Execute a build on a worker node (internal RPC interface).

**Request:**
```go
type BuildRequest struct {
    ProjectPath  string
    TaskName     string
    WorkerID     string
    CacheEnabled bool
    BuildOptions map[string]string
    Timestamp    time.Time
    RequestID    string
}
```

**Response:**
```go
type BuildResponse struct {
    Success       bool
    WorkerID      string
    BuildDuration time.Duration
    Artifacts     []string
    ErrorMessage  string
    Metrics       BuildMetrics
    RequestID     string
    Timestamp     time.Time
}
```

### Worker Management

#### Register Worker
**RPC Call** `BuildCoordinator.RegisterWorker`

Register a worker with the coordinator.

#### Heartbeat
**RPC Call** `BuildCoordinator.Heartbeat`

Send heartbeat from worker to coordinator.

#### Unregister Worker
**RPC Call** `BuildCoordinator.UnregisterWorker`

Remove worker from the coordinator pool.

## Error Codes

### Common HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | Success |
| 400 | Bad Request - Invalid parameters |
| 401 | Unauthorized - Authentication required |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource doesn't exist |
| 409 | Conflict - Resource state conflict |
| 422 | Unprocessable Entity - Validation failed |
| 500 | Internal Server Error - Unexpected error |
| 503 | Service Unavailable - Service temporarily down |

### ML Service Specific Errors

| Error Code | Description |
|------------|-------------|
| `INSUFFICIENT_DATA` | Not enough data for training/prediction |
| `TRAINING_FAILED` | Model training failed |
| `INVALID_MODEL_VERSION` | Requested model version doesn't exist |
| `ROLLBACK_FAILED` | Model rollback failed |

### Coordinator Specific Errors

| Error Code | Description |
|------------|-------------|
| `QUEUE_FULL` | Build queue is at capacity |
| `NO_WORKERS_AVAILABLE` | No workers available for build execution |
| `WORKER_NOT_FOUND` | Specified worker doesn't exist |
| `BUILD_NOT_FOUND` | Specified build doesn't exist |

## Rate Limiting

### API Rate Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| `/api/build` | 100/hour | Per IP |
| `/api/predict` | 1000/hour | Per service |
| `/api/learning` | 60/minute | Per service |
| `/api/metrics` | 120/minute | Per service |

### Burst Limits

- Burst allowance: 2x normal rate for 10 seconds
- Rate limit headers included in responses:
  - `X-RateLimit-Limit`: Maximum requests per window
  - `X-RateLimit-Remaining`: Remaining requests in current window
  - `X-RateLimit-Reset`: Time when limit resets (Unix timestamp)

## Authentication

### API Keys

Most administrative endpoints require API key authentication:

```
Authorization: Bearer <api-key>
```

### Worker Authentication

Workers authenticate via mutual TLS certificates during registration.

## Webhooks

### Build Completion Webhook

Configure webhooks to receive build completion notifications:

```json
{
  "webhook_url": "https://your-app.com/webhooks/build-completed",
  "events": ["build.completed", "build.failed"],
  "secret": "webhook-secret-key"
}
```

**Webhook Payload:**
```json
{
  "event": "build.completed",
  "build_id": "build-1640995200",
  "project_path": "/projects/myapp",
  "task_name": "build",
  "success": true,
  "duration": 45000000000,
  "worker_id": "worker-1",
  "timestamp": "2023-12-31T12:00:45Z"
}
```

## SDKs and Libraries

### Go Client

```go
import "github.com/your-org/distributed-gradle-building/client"

// Create client
client := distributedgradle.NewClient("http://localhost:8080", "your-api-key")

// Submit build
buildID, err := client.SubmitBuild(&BuildRequest{
    ProjectPath: "/projects/myapp",
    TaskName:    "build",
})

// Get predictions
insights, err := client.GetBuildInsights("/projects/myapp", "build")
```

### Python Client

```python
from distributed_gradle import Client

client = Client("http://localhost:8080", api_key="your-api-key")

# Submit build
build_id = client.submit_build({
    "project_path": "/projects/myapp",
    "task_name": "build"
})

# Get predictions
insights = client.get_build_insights("/projects/myapp", "build")
```

### JavaScript Client

```javascript
import { DistributedGradleClient } from 'distributed-gradle-client';

const client = new DistributedGradleClient({
  baseURL: 'http://localhost:8080',
  apiKey: 'your-api-key'
});

// Submit build
const buildId = await client.submitBuild({
  projectPath: '/projects/myapp',
  taskName: 'build'
});

// Get predictions
const insights = await client.getBuildInsights('/projects/myapp', 'build');
```

## Versioning

API versioning follows semantic versioning:

- **Major version**: Breaking changes (URL path changes)
- **Minor version**: New features (backward compatible)
- **Patch version**: Bug fixes (backward compatible)

Current version: `v1.0.0`

Version is included in response headers:
```
X-API-Version: v1.0.0
```

## Changelog

### v1.0.0 (Current)
- Initial release with core build orchestration
- ML-powered predictions and continuous learning
- Comprehensive monitoring and alerting
- Distributed caching with compression
- Horizontal scaling support

### Upcoming Features
- **v1.1.0**: Advanced ML models (time series forecasting)
- **v1.2.0**: Multi-cloud deployment support
- **v2.0.0**: GraphQL API (breaking change)

---

For additional support or questions about the API, please refer to the [Operations Manual](OPERATIONS_MANUAL.md) or create an issue in the project repository.</content>
</xai:function_call">### 4. Configuration Guide

<xai:function_call name="write">
<parameter name="filePath">docs/CONFIGURATION_GUIDE.md