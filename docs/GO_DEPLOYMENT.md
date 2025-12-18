# Go Implementation Deployment Guide

This guide provides detailed instructions for deploying the Go-based distributed Gradle building system.

---

## Overview

The Go implementation offers a high-performance, scalable alternative to the Bash scripts, providing:

- RESTful HTTP APIs for integration
- Real-time monitoring and metrics
- Advanced caching with configurable storage backends
- Worker pool management with auto-scaling
- Comprehensive health checking and alerting

---

## System Requirements

### Minimum Requirements

- **CPU**: 2 cores (4+ recommended)
- **Memory**: 4GB RAM (8GB+ recommended)
- **Storage**: 20GB free space (50GB+ for large projects)
- **Network**: 1Gbps connection between nodes
- **OS**: Linux/macOS/Windows (Linux recommended)

### Software Dependencies

- **Go**: 1.19 or later
- **Gradle**: 7.0 or later
- **Java**: JDK 11 or later (matching project requirements)
- **Git**: For source control
- **cURL**: For API testing (optional)
- **jq**: For JSON processing (optional)

---

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Coordinator   │    │   Cache Server  │    │     Monitor     │
│   (Port 8080)   │    │   (Port 8083)   │    │   (Port 8084)   │
│                 │    │                 │    │                 │
│ • Build Queue   │◄──►│ • Distributed   │◄──►│ • Metrics       │
│ • Worker Pool   │    │   Cache Storage │    │ • Health Checks │
│ • RPC Server    │    │ • HTTP API      │    │ • Alerts        │
│ • HTTP API      │    │ • Cleanup       │    │ • Reports       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
         ┌───────────────────────▼───────────────────────┐
         │              Worker Nodes                     │
         │  (Ports 8082, 8083, 8084, ...)              │
         │                                               │
         │ • Task Execution                              │
         │ • Local Caching                              │
         │ • Resource Monitoring                        │
         │ • Heartbeat Reporting                        │
         └───────────────────────────────────────────────┘
```

---

## Quick Start

### 1. Build and Deploy

```bash
# Clone and navigate to project
git clone <repository-url>
cd Distributed-Gradle-Building

# Build all components
./scripts/build_and_deploy.sh build

# Start all services
./scripts/build_and_deploy.sh start

# Check service health
./scripts/build_and_deploy.sh status

# Run demo build
./scripts/build_and_deploy.sh demo
```

### 2. Verify Installation

```bash
# Check coordinator status
curl http://localhost:8080/api/status

# Check cache server
curl http://localhost:8083/health

# Check monitor
curl http://localhost:8084/health
```

---

## Component Configuration

### Coordinator Configuration

The coordinator doesn't require a configuration file by default. It uses command-line flags and environment variables:

```bash
# Start coordinator with custom settings
cd go
./coordinator \
  --http-port=8080 \
  --rpc-port=8081 \
  --max-workers=10 \
  --log-level=info
```

### Worker Configuration

Create `worker_config.json` for each worker:

```json
{
  "id": "worker-001",
  "host": "localhost",
  "port": 8082,
  "coordinator": "localhost:8081",
  "cache_dir": "/tmp/gradle-cache-worker-001",
  "work_dir": "/tmp/gradle-work-001",
  "max_builds": 5,
  "capabilities": ["gradle", "java", "testing", "compilation"],
  "gradle_home": "/tmp/gradle-home-001"
}
```

**Configuration Fields**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique worker identifier |
| `host` | string | Yes | Worker hostname/IP |
| `port` | int | Yes | Worker RPC port |
| `coordinator` | string | Yes | Coordinator address (host:port) |
| `cache_dir` | string | No | Local cache directory |
| `work_dir` | string | No | Working directory |
| `max_builds` | int | No | Maximum concurrent builds |
| `capabilities` | array | No | Worker capabilities |
| `gradle_home` | string | No | Gradle user home directory |

### Cache Server Configuration

Create `cache_config.json`:

```json
{
  "host": "localhost",
  "port": 8083,
  "storage_type": "filesystem",
  "storage_dir": "/tmp/gradle-cache-server",
  "max_cache_size": 10737418240,
  "ttl_seconds": 86400,
  "compression": true,
  "authentication": false,
  "auth_token": ""
}
```

**Configuration Fields**:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `host` | string | "localhost" | Server bind address |
| `port` | int | 8083 | Server port |
| `storage_type` | string | "filesystem" | Storage backend type |
| `storage_dir` | string | "/tmp/gradle-cache-server" | Storage directory |
| `max_cache_size` | int64 | 10GB | Maximum cache size in bytes |
| `ttl_seconds` | int | 86400 | Default TTL for entries |
| `compression` | bool | true | Enable compression |
| `authentication` | bool | false | Enable authentication |
| `auth_token` | string | "" | Authentication token |

### Monitor Configuration

Create `monitor_config.json`:

```json
{
  "host": "localhost",
  "port": 8084,
  "data_dir": "/tmp/gradle-monitor",
  "check_interval_seconds": 30,
  "history_retention_days": 30,
  "alert_thresholds": {
    "cpu_usage": 80.0,
    "memory_usage": 90.0,
    "disk_usage": 85.0,
    "build_failure_rate": 20.0,
    "response_time": 5.0
  }
}
```

---

## Production Deployment

### 1. Single Machine Deployment

For development or small teams:

```bash
# Use the deployment script
./scripts/build_and_deploy.sh start

# Or manually start components
cd go

# Start coordinator
nohup ./coordinator > logs/coordinator.log 2>&1 &

# Start cache server
nohup ./cache_server > logs/cache_server.log 2>&1 &

# Start monitor
nohup ./monitor > logs/monitor.log 2>&1 &

# Start workers
nohup ./worker worker_config.json > logs/worker-001.log 2>&1 &
```

### 2. Multi-Machine Deployment

For production environments:

#### Master Node Setup

```bash
# On master node
cd /opt/distributed-gradle-build/go

# Start coordinator
systemctl start gradle-coordinator

# Start cache server
systemctl start gradle-cache-server

# Start monitor
systemctl start gradle-monitor
```

#### Worker Node Setup

```bash
# On each worker node
cd /opt/distributed-gradle-build/go

# Start worker with master configuration
./worker worker_config.json
```

### 3. Systemd Service Configuration

Create systemd service files for automatic startup:

#### Coordinator Service

```ini
# /etc/systemd/system/gradle-coordinator.service
[Unit]
Description=Gradle Distributed Build Coordinator
After=network.target

[Service]
Type=simple
User=gradle
WorkingDirectory=/opt/distributed-gradle-build/go
ExecStart=/opt/distributed-gradle-build/go/coordinator
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

#### Worker Service

```ini
# /etc/systemd/system/gradle-worker@.service
[Unit]
Description=Gradle Distributed Build Worker %i
After=network.target

[Service]
Type=simple
User=gradle
WorkingDirectory=/opt/distributed-gradle-build/go
ExecStart=/opt/distributed-gradle-build/go/worker /opt/distributed-gradle-build/go/worker_%i_config.json
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Enable and start services:

```bash
systemctl enable gradle-coordinator
systemctl start gradle-coordinator

systemctl enable gradle-worker@1
systemctl start gradle-worker@1
```

---

## Docker Deployment

### 1. Dockerfile

```dockerfile
# Dockerfile
FROM golang:1.19-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN go build -o coordinator main.go
RUN go build -o worker worker.go
RUN go build -o cache_server cache_server.go
RUN go build -o monitor monitor.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates gradle openjdk11-jre curl

WORKDIR /app
COPY --from=builder /app/coordinator .
COPY --from=builder /app/worker .
COPY --from=builder /app/cache_server .
COPY --from=builder /app/monitor .
COPY *.json ./

RUN mkdir -p /data/cache /data/monitor /data/logs

EXPOSE 8080 8081 8082 8083 8084

CMD ["./coordinator"]
```

### 2. Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  coordinator:
    build: .
    ports:
      - "8080:8080"
      - "8081:8081"
    volumes:
      - ./data:/data
    command: ["./coordinator"]

  cache-server:
    build: .
    ports:
      - "8083:8083"
    volumes:
      - ./data/cache:/data/cache
    command: ["./cache_server"]

  monitor:
    build: .
    ports:
      - "8084:8084"
    volumes:
      - ./data/monitor:/data/monitor
    command: ["./monitor"]

  worker-1:
    build: .
    environment:
      - WORKER_ID=worker-1
      - COORDINATOR=coordinator:8081
    command: ["./worker", "worker_config.json"]
    depends_on:
      - coordinator

  worker-2:
    build: .
    environment:
      - WORKER_ID=worker-2
      - COORDINATOR=coordinator:8081
    command: ["./worker", "worker_2_config.json"]
    depends_on:
      - coordinator

  worker-3:
    build: .
    environment:
      - WORKER_ID=worker-3
      - COORDINATOR=coordinator:8081
    command: ["./worker", "worker_3_config.json"]
    depends_on:
      - coordinator
```

Deploy with Docker Compose:

```bash
docker-compose up -d
```

---

## Kubernetes Deployment

### 1. Namespace and ConfigMap

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: gradle-distributed-build

---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gradle-build-config
  namespace: gradle-distributed-build
data:
  coordinator-config.json: |
    {
      "http_port": 8080,
      "rpc_port": 8081,
      "max_workers": 10
    }
  cache-config.json: |
    {
      "host": "0.0.0.0",
      "port": 8083,
      "storage_dir": "/data/cache",
      "max_cache_size": 10737418240
    }
  monitor-config.json: |
    {
      "host": "0.0.0.0",
      "port": 8084,
      "data_dir": "/data/monitor"
    }
```

### 2. Coordinator Deployment

```yaml
# k8s/coordinator.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gradle-coordinator
  namespace: gradle-distributed-build
spec:
  replicas: 1
  selector:
    matchLabels:
      app: gradle-coordinator
  template:
    metadata:
      labels:
        app: gradle-coordinator
    spec:
      containers:
      - name: coordinator
        image: gradle-distributed-build:latest
        ports:
        - containerPort: 8080
        - containerPort: 8081
        command: ["./coordinator"]
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1"

---
apiVersion: v1
kind: Service
metadata:
  name: gradle-coordinator-service
  namespace: gradle-distributed-build
spec:
  selector:
    app: gradle-coordinator
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  - name: rpc
    port: 8081
    targetPort: 8081
  type: ClusterIP
```

### 3. Worker Deployment

```yaml
# k8s/worker.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gradle-worker
  namespace: gradle-distributed-build
spec:
  replicas: 3
  selector:
    matchLabels:
      app: gradle-worker
  template:
    metadata:
      labels:
        app: gradle-worker
    spec:
      containers:
      - name: worker
        image: gradle-distributed-build:latest
        command: ["./worker"]
        args: ["/etc/config/worker-config.json"]
        resources:
          requests:
            memory: "1Gi"
            cpu: "1"
          limits:
            memory: "2Gi"
            cpu: "2"
        volumeMounts:
        - name: config
          mountPath: /etc/config
      volumes:
      - name: config
        configMap:
          name: gradle-build-config
```

---

## Monitoring and Observability

### 1. Metrics Collection

The system provides comprehensive metrics through:

- **HTTP API Endpoints**: Real-time metrics via REST APIs
- **Log Files**: Detailed logging for each component
- **Health Checks**: Endpoint status monitoring
- **Performance Reports**: Historical analysis and recommendations

### 2. Prometheus Integration

Add Prometheus metrics exporter:

```yaml
# prometheus.yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'gradle-coordinator'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    
  - job_name: 'gradle-cache-server'
    static_configs:
      - targets: ['localhost:8083']
    metrics_path: '/metrics'
    
  - job_name: 'gradle-monitor'
    static_configs:
      - targets: ['localhost:8084']
    metrics_path: '/metrics'
```

### 3. Grafana Dashboard

Create a Grafana dashboard for monitoring:

- Build throughput over time
- Worker utilization rates
- Cache hit rates
- System resource usage
- Alert history

---

## Troubleshooting

### Common Issues

#### 1. Services Fail to Start

**Problem**: Services exit immediately after starting

**Solution**:
```bash
# Check logs
./scripts/build_and_deploy.sh status
tail -f logs/coordinator.log
tail -f logs/cache_server.log

# Check port availability
netstat -tlnp | grep :8080
lsof -i :8080

# Check dependencies
go version
gradle --version
java -version
```

#### 2. Workers Not Registering

**Problem**: Workers don't appear in coordinator status

**Solution**:
```bash
# Check worker configuration
cat worker_config.json

# Test network connectivity
telnet coordinator-host 8081

# Check worker logs
tail -f logs/worker-001.log

# Verify RPC connectivity
curl -X POST http://localhost:8081/rpc -d '{"method":"Coordinator.Ping"}'
```

#### 3. Build Failures

**Problem**: Builds fail with permission or path errors

**Solution**:
```bash
# Check project permissions
ls -la /path/to/project
chmod -R 755 /path/to/project

# Verify Gradle installation
which gradle
gradle --version

# Test local build
cd /path/to/project
gradle build
```

#### 4. Cache Issues

**Problem**: Cache misses or storage errors

**Solution**:
```bash
# Check cache directory
ls -la /tmp/gradle-cache-server
df -h /tmp/gradle-cache-server

# Check cache server logs
tail -f logs/cache_server.log

# Test cache API
curl -X GET http://localhost:8083/stats
```

### Debug Mode

Enable debug logging:

```bash
# Set log level
export LOG_LEVEL=debug

# Start with verbose output
./coordinator --log-level=debug
./worker --log-level=debug
./cache_server --log-level=debug
./monitor --log-level=debug
```

### Health Check Script

Create a comprehensive health check script:

```bash
#!/bin/bash
# health_check.sh

echo "Checking distributed build system health..."

# Check coordinator
if curl -sf http://localhost:8080/api/status >/dev/null; then
    echo "✓ Coordinator is healthy"
else
    echo "✗ Coordinator is not responding"
    exit 1
fi

# Check cache server
if curl -sf http://localhost:8083/health >/dev/null; then
    echo "✓ Cache server is healthy"
else
    echo "✗ Cache server is not responding"
    exit 1
fi

# Check monitor
if curl -sf http://localhost:8084/health >/dev/null; then
    echo "✓ Monitor is healthy"
else
    echo "✗ Monitor is not responding"
    exit 1
fi

# Check workers
worker_count=$(curl -sf http://localhost:8080/api/workers | jq length 2>/dev/null || echo "0")
echo "✓ $worker_count workers registered"

echo "All components are healthy!"
```

---

## Performance Tuning

### 1. Worker Pool Optimization

- **CPU Allocation**: Allocate 2 CPU cores per worker for Java compilation
- **Memory Allocation**: Minimum 2GB RAM per worker for large projects
- **Network Optimization**: Use 10Gbps network for cache transfers
- **Storage**: Use SSD storage for Gradle cache

### 2. Cache Optimization

```json
{
  "max_cache_size": 53687091200,  // 50GB
  "compression": true,
  "ttl_seconds": 604800,        // 7 days
  "cleanup_interval": 3600       // 1 hour
}
```

### 3. Build Configuration

```properties
# gradle.properties
org.gradle.parallel=true
org.gradle.caching=true
org.gradle.workers.max=4
org.gradle.jvmargs=-Xmx4g -XX:+UseG1GC
```

---

## Security Considerations

### 1. Network Security

- Use TLS for all API endpoints in production
- Implement authentication tokens for cache server
- Restrict network access using firewall rules
- Use VPN for inter-node communication

### 2. File System Security

```bash
# Set appropriate permissions
chmod 750 /opt/distributed-gradle-build
chmod 640 worker_config.json
chmod 640 cache_config.json
chmod 640 monitor_config.json
```

### 3. Process Isolation

- Run services under dedicated user account
- Use containerization (Docker/Kubernetes) for isolation
- Implement resource limits to prevent abuse
- Regular security updates for Go runtime and dependencies

---

## Backup and Recovery

### 1. Configuration Backup

```bash
#!/bin/bash
# backup_configs.sh

BACKUP_DIR="/backup/gradle-build-$(date +%Y%m%d)"
mkdir -p "$BACKUP_DIR"

# Backup configurations
cp *.json "$BACKUP_DIR/"
cp *.properties "$BACKUP_DIR/"

# Backup cache data
tar -czf "$BACKUP_DIR/cache.tar.gz" /tmp/gradle-cache-server/

# Backup monitoring data
tar -czf "$BACKUP_DIR/monitor.tar.gz" /tmp/gradle-monitor/

echo "Backup completed: $BACKUP_DIR"
```

### 2. Service Recovery

```bash
#!/bin/bash
# recovery.sh

# Stop all services
./scripts/build_and_deploy.sh stop

# Restore configurations
cp /backup/latest/*.json .

# Restart services
./scripts/build_and_deploy.sh start

# Verify health
./scripts/build_and_deploy.sh status
```

---

## API Client Libraries

### Go Client

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type GradleBuildClient struct {
    BaseURL string
    Client  *http.Client
}

func NewClient(baseURL string) *GradleBuildClient {
    return &GradleBuildClient{
        BaseURL: baseURL,
        Client:  &http.Client{},
    }
}

func (c *GradleBuildClient) SubmitBuild(projectPath, taskName string) (string, error) {
    req := map[string]interface{}{
        "project_path":  projectPath,
        "task_name":     taskName,
        "cache_enabled": true,
    }
    
    data, _ := json.Marshal(req)
    resp, err := c.Client.Post(
        c.BaseURL+"/api/build",
        "application/json",
        bytes.NewBuffer(data),
    )
    
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    var result map[string]string
    json.NewDecoder(resp.Body).Decode(&result)
    
    return result["build_id"], nil
}
```

### Python Client

```python
import requests
import json

class GradleBuildClient:
    def __init__(self, base_url):
        self.base_url = base_url.rstrip('/')
    
    def submit_build(self, project_path, task_name="build"):
        """Submit a build request"""
        url = f"{self.base_url}/api/build"
        payload = {
            "project_path": project_path,
            "task_name": task_name,
            "cache_enabled": True
        }
        
        response = requests.post(url, json=payload)
        response.raise_for_status()
        
        return response.json()["build_id"]
    
    def get_build_status(self, build_id):
        """Get build status"""
        url = f"{self.base_url}/api/build/{build_id}"
        response = requests.get(url)
        response.raise_for_status()
        
        return response.json()
    
    def get_workers(self):
        """Get worker status"""
        url = f"{self.base_url}/api/workers"
        response = requests.get(url)
        response.raise_for_status()
        
        return response.json()
    
    def get_metrics(self):
        """Get system metrics"""
        url = f"{self.base_url}/api/metrics"
        response = requests.get(url)
        response.raise_for_status()
        
        return response.json()

# Usage example
client = GradleBuildClient("http://localhost:8080")
build_id = client.submit_build("/path/to/project", "build")
status = client.get_build_status(build_id)
print(f"Build {build_id} status: {status['status']}")
```

---

## Migration from Bash to Go

### 1. Step-by-Step Migration

1. **Install Go Dependencies**
   ```bash
   cd go
   go mod tidy
   ```

2. **Build Go Components**
   ```bash
   ./scripts/build_and_deploy.sh build
   ```

3. **Parallel Operation**
   ```bash
   # Keep Bash running
   ./sync_and_build.sh &
   
   # Start Go services
   ./scripts/build_and_deploy.sh start
   ```

4. **Gradual Migration**
   ```bash
   # Test with small projects first
   curl -X POST http://localhost:8080/api/build \
     -H "Content-Type: application/json" \
     -d '{"project_path":"/path/to/small/project","task_name":"build"}'
   ```

5. **Complete Migration**
   ```bash
   # Switch to Go-only operation
   ./scripts/build_and_deploy.sh start
   # Replace existing build calls with API calls
   ```

### 2. Feature Comparison

| Feature | Bash Implementation | Go Implementation |
|---------|-------------------|-------------------|
| Build Distribution | SSH-based | RPC/HTTP API |
| Worker Management | Process-based | Service-based |
| Caching | HTTP server | Advanced caching with backends |
| Monitoring | Script-based | Real-time metrics and alerts |
| Scalability | Limited by SSH | Highly scalable |
| API | None | RESTful HTTP API |
| Configuration | Environment variables | JSON configuration files |
| Logging | File-based | Structured logging |
| Error Handling | Basic | Comprehensive error handling |

---

This deployment guide provides comprehensive instructions for deploying and managing the Go-based distributed Gradle building system in various environments, from development to production.