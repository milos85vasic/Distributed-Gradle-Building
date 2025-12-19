# Monitoring and Metrics Guide

## 📊 Overview

The Distributed Gradle Build System provides comprehensive monitoring capabilities to track build performance, resource utilization, and system health across all worker machines.

## 🔍 Monitoring Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Master Node   │    │   Worker 1      │    │   Worker N      │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ Metrics     │ │    │ │ Monitoring  │ │    │ │ Monitoring  │ │
│ │ Collection  │ │◄──►│ │ Agent       │ │◄──►│ │ Agent       │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
│                 │    │                 │    │                 │
│ ┌─────────────┐ │    │ ┌─────────────┐ │    │ ┌─────────────┐ │
│ │ Performance │ │    │ │ Resource    │ │    │ │ Resource    │ │
│ │ Dashboard  │ │    │ │ Tracking    │ │    │ │ Tracking    │ │
│ └─────────────┘ │    │ └─────────────┘ │    │ └─────────────┘ │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📈 Available Metrics

### Build Performance Metrics

```json
{
  "build_id": "build-1703123456",
  "start_time": 1703123456,
  "end_time": 1703123656,
  "metrics": {
    "total_tasks": 12,
    "completed_tasks": 11,
    "failed_tasks": 1,
    "build_duration_seconds": 100,
    "success_rate": 91.7
  }
}
```

### Worker Resource Metrics

```json
{
  "worker_cpu_usage": {
    "192.168.1.101": 75.2,
    "192.168.1.102": 68.4,
    "192.168.1.103": 82.1
  },
  "worker_memory_usage": {
    "192.168.1.101": 65.3,
    "192.168.1.102": 78.9,
    "192.168.1.103": 45.6
  }
}
```

### Task Distribution Metrics

```json
{
  "tasks": [
    {
      "name": "compileJava",
      "worker": "192.168.1.101",
      "duration_seconds": 45,
      "artifacts_count": 156,
      "status": "success"
    }
  ]
}
```

## 🛠️ Monitoring Setup

### 1. Enable Monitoring

```bash
# Start monitoring on master
export MONITORING_ENABLED=true
./distributed_gradle_build.sh assemble

# Enable verbose monitoring
export DEBUG=true
export METRICS_LEVEL=verbose
./sync_and_build.sh test
```

### 2. Configure Monitoring Intervals

```bash
# In .gradlebuild_env
export MONITOR_INTERVAL_SECONDS=5
export WORKER_MONITOR_INTERVAL=10
export METRICS_RETENTION_DAYS=30
```

### 3. Setup Monitoring Daemons

```bash
# Start master monitoring daemon
./scripts/performance_analyzer.sh start

# Start worker monitoring (run on each worker)
./setup_worker.sh /path/to/project --with-monitoring
```

## 📊 Real-time Monitoring

### Command Line Monitoring

```bash
# Live build monitoring
watch -n 1 'cat .distributed/metrics/build_metrics.json | jq'

# Worker resource monitoring
for worker in $WORKER_IPS; do
    echo "=== Worker: $worker ==="
    ssh $worker "cat $PROJECT_DIR/.distributed/metrics/worker_metrics.json | jq"
done

# Task progress monitoring
tail -f .distributed/logs/distributed_build.log | grep -E "(START|COMPLETE|FAIL)"
```

### Web Dashboard (Optional)

```bash
# Start monitoring web server
./scripts/performance_analyzer.sh --web-dashboard --port 8080

# Access at: http://localhost:8080
```

## 📈 Performance Analysis

### Build Trend Analysis

```bash
#!/bin/bash
# analyze_build_trends.sh

METRICS_DIR=".distributed/metrics"
DAYS=7

echo "Build Performance Analysis (Last $DAYS Days)"
echo "========================================="

# Extract build durations
jq -r '.metrics.build_duration_seconds' "$METRICS_DIR"/build_*.json | \
  awk '{sum+=$1; count++} END {print "Average Build Time:", sum/count, "seconds"}'

# Calculate success rate
total_builds=$(ls "$METRICS_DIR"/build_*.json | wc -l)
successful_builds=$(jq -r '.metrics.success_rate' "$METRICS_DIR"/build_*.json | \
  awk '{if($1>90) count++} END {print count}')

echo "Success Rate: $(echo "scale=1; $successful_builds*100/$total_builds" | bc -l)%"

# Worker utilization analysis
echo "Average Worker CPU Usage:"
jq -r '.metrics.worker_cpu_usage | to_entries[] | "\(.key): \(.value)%"' "$METRICS_DIR"/build_*.json | \
  awk -F: '{cpu[$1]+=$2; count[$1]++} END {for(w in cpu) print w, cpu[w]/count[w]"%"}'
```

### Resource Utilization Analysis

```bash
#!/bin/bash
# resource_utilization_report.sh

echo "Resource Utilization Report"
echo "========================="

# CPU utilization by worker
echo "CPU Utilization:"
for worker in $WORKER_IPS; do
    cpu_usage=$(ssh $worker "top -bn1 | grep 'Cpu'" | awk '{print $2}' | sed 's/%us,//')
    echo "  $worker: ${cpu_usage}%"
done

# Memory usage by worker
echo "Memory Utilization:"
for worker in $WORKER_IPS; do
    mem_usage=$(ssh $worker "free | grep Mem | awk '{printf \"%.1f\", $3/$2 * 100.0}'")
    echo "  $worker: ${mem_usage}%"
done

# Network utilization
echo "Network Usage:"
if command -v ifstat >/dev/null 2>&1; then
    ifstat -i eth0 1 1 | tail -1
else
    echo "  ifstat not available for network monitoring"
fi
```

### Performance Bottleneck Detection

```bash
#!/bin/bash
# detect_bottlenecks.sh

LATEST_METRICS=".distributed/metrics/build_metrics.json"

# Check for slow workers
slow_workers=$(jq -r '.metrics.worker_cpu_usage | to_entries[] | select(.value < 30) | .key' "$LATEST_METRICS")
if [[ -n "$slow_workers" ]]; then
    echo "⚠️  WARNING: Underutilized workers detected:"
    echo "$slow_workers"
fi

# Check for memory pressure
high_memory_workers=$(jq -r '.metrics.worker_memory_usage | to_entries[] | select(.value > 85) | .key' "$LATEST_METRICS")
if [[ -n "$high_memory_workers" ]]; then
    echo "⚠️  WARNING: High memory usage on workers:"
    echo "$high_memory_workers"
fi

# Check for failed tasks
failed_tasks=$(jq -r '.metrics.failed_tasks' "$LATEST_METRICS")
if [[ $failed_tasks -gt 0 ]]; then
    echo "⚠️  WARNING: $failed_tasks failed tasks detected"
fi
```

## 📊 Custom Metrics

### Adding Custom Metrics

```bash
# In distributed_gradle_build.sh, add custom metric tracking
track_custom_metric() {
    local metric_name="$1"
    local metric_value="$2"
    
    jq ".custom_metrics[\"$metric_name\"] = $metric_value" "$METRICS_FILE" > "${METRICS_FILE}.tmp" && \
    mv "${METRICS_FILE}.tmp" "$METRICS_FILE"
}

# Example usage
track_custom_metric "git_branch" "$(git rev-parse --abbrev-ref HEAD)"
track_custom_metric "test_coverage" "$(./gradlew testCoverage | grep -o '[0-9]*%')"
```

### Integration with External Systems

```bash
# Prometheus integration
export PROMETHEUS_ENABLED=true
export PROMETHEUS_PORT=9090

# Grafana dashboard
export GRAFANA_ENABLED=true
export GRAFANA_DASHBOARD_PATH="/path/to/grafana/dashboards"

# ELK stack integration
export ELK_ENABLED=true
export ELK_HOSTNAME="elk-server.local"
```

## 🔔 Alerting Setup

### Email Alerts

```bash
# setup_alerts.sh

ALERT_EMAIL="build-team@company.com"
SMTP_SERVER="smtp.company.com"

send_alert() {
    local subject="$1"
    local message="$2"
    
    sendemail -f builds@company.com \
              -t "$ALERT_EMAIL" \
              -u "$subject" \
              -m "$message" \
              -s "$SMTP_SERVER"
}

# Build failure alert
check_build_failures() {
    local failed_tasks=$(jq -r '.metrics.failed_tasks' ".distributed/metrics/build_metrics.json")
    
    if [[ $failed_tasks -gt 2 ]]; then
        send_alert "🚨 Build Alert: Multiple Task Failures" \
                  "Distributed build has $failed_tasks failed tasks. Investigation required."
    fi
}

# Performance degradation alert
check_performance() {
    local build_time=$(jq -r '.metrics.build_duration_seconds' ".distributed/metrics/build_metrics.json")
    local baseline_time=300  # 5 minutes baseline
    
    if [[ $build_time -gt $((baseline_time * 2)) ]]; then
        send_alert "🚨 Performance Alert: Build Slowdown" \
                  "Build time ${build_time}s is 2x baseline of ${baseline_time}s."
    fi
}
```

### Slack Integration

```bash
# slack_alerts.sh

SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"

send_slack_alert() {
    local color="$1"
    local title="$2"
    local message="$3"
    
    curl -X POST -H 'Content-type: application/json' \
        --data "{\"attachments\": [{\"color\": \"$color\", \"title\": \"$title\", \"text\": \"$message\"}]}" \
        "$SLACK_WEBHOOK_URL"
}

# Example usage
send_slack_alert "danger" "Build Failed" "Distributed build failed on worker 192.168.1.103"
send_slack_alert "warning" "Performance Alert" "Build time exceeded threshold"
send_slack_alert "good" "Build Success" "All tasks completed successfully"
```

## 📈 Performance Dashboards

### Grafana Dashboard Configuration

```json
{
  "dashboard": {
    "title": "Distributed Gradle Build Metrics",
    "panels": [
      {
        "title": "Build Duration Trend",
        "type": "graph",
        "targets": [
          {
            "expr": "gradle_build_duration_seconds",
            "legendFormat": "Build Time (s)"
          }
        ]
      },
      {
        "title": "Worker CPU Usage",
        "type": "graph", 
        "targets": [
          {
            "expr": "gradle_worker_cpu_usage",
            "legendFormat": "{{worker}}"
          }
        ]
      },
      {
        "title": "Task Success Rate",
        "type": "stat",
        "targets": [
          {
            "expr": "gradle_success_rate"
          }
        ]
      }
    ]
  }
}
```

### Prometheus Metrics Export

```bash
# prometheus_exporter.sh

export_metrics_to_prometheus() {
    local metrics_file=".distributed/metrics/build_metrics.json"
    
    # Convert JSON to Prometheus format
    echo "# HELP gradle_build_duration_seconds Time to complete distributed build"
    echo "# TYPE gradle_build_duration_seconds gauge"
    jq -r '.metrics.build_duration_seconds' "$metrics_file" | \
      awk '{print "gradle_build_duration_seconds " $1}'
    
    echo "# HELP gradle_worker_cpu_usage CPU usage percentage per worker"
    echo "# TYPE gradle_worker_cpu_usage gauge"
    jq -r '.metrics.worker_cpu_usage | to_entries[] | "gradle_worker_cpu_usage{worker=\"" .key "\"} " .value' "$metrics_file"
}
```

## 🔍 Log Analysis

### Centralized Log Collection

```bash
# collect_logs.sh

LOG_DIR=".distributed/logs"
ARCHIVE_DIR=".distributed/logs/archive"

# Archive old logs
find "$LOG_DIR" -name "*.log" -mtime +7 -exec gzip {} \;
mkdir -p "$ARCHIVE_DIR"
mv "$LOG_DIR"/*.log.gz "$ARCHIVE_DIR/" 2>/dev/null || true

# Generate log summary
echo "Log Analysis Summary"
echo "==================="
echo "Build logs: $(ls "$LOG_DIR"/build_*.log 2>/dev/null | wc -l)"
echo "Error logs: $(ls "$LOG_DIR"/error_*.log 2>/dev/null | wc -l)"
echo "Worker logs: $(find "$LOG_DIR" -name "worker_*.log" | wc -l)"
```

### Error Pattern Analysis

```bash
# analyze_errors.sh

echo "Error Pattern Analysis"
echo "===================="

# Find most common errors
grep -h "ERROR\|FAILED\|Exception" .distributed/logs/*.log | \
  sort | uniq -c | sort -nr | head -10

# Find worker-specific errors
for worker in $WORKER_IPS; do
    echo "=== Worker $worker Errors ==="
    grep "$worker" .distributed/logs/error_*.log | tail -5
done
```

## 📊 Custom Monitoring Solutions

### Simple Web Dashboard

```bash
# simple_dashboard.py
#!/usr/bin/env python3

import json
import os
from flask import Flask, render_template_string

app = Flask(__name__)

@app.route('/')
def dashboard():
    metrics_file = ".distributed/metrics/build_metrics.json"
    if os.path.exists(metrics_file):
        with open(metrics_file) as f:
            metrics = json.load(f)
    else:
        metrics = {}
    
    return render_template_string('''
    <h1>Distributed Gradle Build Dashboard</h1>
    <h2>Build Metrics</h2>
    <pre>{{ metrics | tojson(indent=2) }}</pre>
    
    <h2>Worker Status</h2>
    <div id="worker-status">Loading...</div>
    
    <script>
        // Refresh every 5 seconds
        setTimeout(() => location.reload(), 5000);
    </script>
    ''', metrics=metrics)

if __name__ == '__main__':
    app.run(debug=True, port=8080)
```

## 🎯 Monitoring Best Practices

### 1. Metrics Collection

- **Frequency:** Monitor every 5-30 seconds depending on build duration
- **Retention:** Keep 30-90 days of historical data
- **Granularity:** Track both aggregate and per-worker metrics
- **Context:** Include build metadata (branch, project version, etc.)

### 2. Alert Thresholds

- **Build Time:** Alert if >2x baseline
- **Failure Rate:** Alert if >5% failure rate
- **Resource Usage:** Alert if CPU <30% or Memory >90%
- **Network Issues:** Alert if >10% packet loss or high latency

### 3. Data Visualization

- Use real-time dashboards for immediate insights
- Historical trends for capacity planning
- Correlate metrics to identify root causes
- Share reports with development team

---

**Need monitoring integration?** See [Performance Guide](PERFORMANCE.md) for optimization strategies.