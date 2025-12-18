# Scripts Reference Guide

This guide provides detailed documentation for all management scripts included in the Distributed Gradle Building System.

## Overview

The system includes several specialized scripts for different aspects of distributed building:

- **performance_analyzer.sh** - Build performance analysis and optimization
- **worker_pool_manager.sh** - Worker pool management and allocation
- **cache_manager.sh** - Build cache management and optimization
- **health_checker.sh** - System health monitoring and alerting

---

## performance_analyzer.sh

### Purpose
Analyzes build performance metrics and provides optimization recommendations.

### Usage
```bash
./scripts/performance_analyzer.sh [command]
```

### Commands

#### Interactive Mode (default)
Launches an interactive menu for running experiments:
```bash
./scripts/performance_analyzer.sh
./scripts/performance_analyzer.sh interactive
```

#### Quick Test
Runs performance tests with common configurations:
```bash
./scripts/performance_analyzer.sh test
```

#### Analyze Existing Data
Analyzes metrics from previous runs:
```bash
./scripts/performance_analyzer.sh analyze
```

#### Generate Charts
Creates ASCII charts from collected data:
```bash
./scripts/performance_analyzer.sh charts
```

### Features

#### Experiment Configuration
- **Worker counts**: Test 1, 2, 4, 8 workers (customizable)
- **Cache testing**: Compare cached vs non-cached builds
- **Build tasks**: Support any Gradle task (assemble, test, etc.)
- **Multiple runs**: Configure number of runs per configuration

#### Analysis Capabilities
- **Build duration statistics**: Average, min, max, standard deviation
- **Worker count analysis**: Performance scaling across worker counts
- **Cache impact**: Quantify cache speedup and hit rates
- **Success rate analysis**: Build failure rates and patterns
- **Resource utilization**: CPU, memory, network, disk I/O usage

#### Output Formats
- **CSV metrics**: Machine-readable data for further analysis
- **Text analysis**: Human-readable performance report
- **ASCII charts**: Visual representation of performance trends

### Example Workflows

#### Basic Performance Test
```bash
# Quick test with default configurations
./scripts/performance_analyzer.sh test

# View results
cat performance_results/analysis.txt
```

#### Comprehensive Analysis
```bash
# Run interactive experiments
./scripts/performance_analyzer.sh interactive

# Generate detailed charts
./scripts/performance_analyzer.sh charts

# Export data for external analysis
python3 analyze_performance.py performance_results/metrics.csv
```

---

## worker_pool_manager.sh

### Purpose
Manages and optimizes worker pools for different build scenarios.

### Usage
```bash
./scripts/worker_pool_manager.sh [command] [options]
```

### Commands

#### Interactive Mode (default)
Menu-driven pool management:
```bash
./scripts/worker_pool_manager.sh
./scripts/worker_pool_manager.sh interactive
```

#### Show Pool Status
Display current status of all worker pools:
```bash
./scripts/worker_pool_manager.sh status
```

#### Execute on Specific Pool
Run Gradle task on specific worker pool:
```bash
./scripts/worker_pool_manager.sh execute <gradle_task> [workers] [pool_type]
```

Example:
```bash
# Execute assembleDebug on compilation pool with 4 workers
./scripts/worker_pool_manager.sh execute assembleDebug 4 compilation

# Auto-detect pool and workers
./scripts/worker_pool_manager.sh execute test
```

#### Continuous Monitoring
Monitor all worker pools continuously:
```bash
./scripts/worker_pool_manager.sh monitor
```

#### Pool Detection
Auto-detect best pool for a task:
```bash
./scripts/worker_pool_manager.sh detect <gradle_task>
```

### Pool Types

#### Compilation Pool
- **Purpose**: CPU-intensive tasks (compile, assemble)
- **Resource profile**: High CPU, medium memory, medium storage
- **Default workers**: build-1, build-2, build-3, build-4
- **Optimization**: Maximize parallel compilation

#### Testing Pool
- **Purpose**: Memory-intensive tasks (test, integration tests)
- **Resource profile**: Medium CPU, high memory, low storage
- **Default workers**: test-1, test-2
- **Optimization**: Allocate sufficient memory for test execution

#### Packaging Pool
- **Purpose**: I/O-intensive tasks (bundle, package)
- **Resource profile**: Medium CPU, medium memory, high storage
- **Default workers**: build-5, build-6
- **Optimization**: Fast disk I/O for packaging operations

#### Default Pool
- **Purpose**: General-purpose tasks
- **Resource profile**: Balanced resources
- **Default workers**: All available workers
- **Optimization**: Good all-around performance

### Configuration

#### Worker Configuration File
Create `workers.conf` for custom worker pools:
```bash
# Format: POOL_NAME="type:worker1 worker2 worker3"
COMPILATION_POOL="compilation:worker-1 worker-2 worker-3 worker-4"
TESTING_POOL="testing:test-1 test-2"
PACKAGING_POOL="packaging:build-5 build-6"
DEFAULT_POOL="default:build-1 build-2 build-3 build-4 build-5 build-6"
```

#### Resource Requirements
```bash
# Pool resource requirements
# Format: POOL_TYPE="cpu:level memory:level storage:level"
COMPILATION_POOL="cpu:high memory:medium storage:medium"
TESTING_POOL="cpu:medium memory:high storage:low"
PACKAGING_POOL="cpu:medium memory:medium storage:high"
```

### Health Monitoring

#### Worker Health Checks
- **SSH connectivity**: Verify network accessibility
- **CPU usage**: Monitor for overloaded conditions
- **Memory usage**: Check for memory pressure
- **Disk usage**: Ensure sufficient storage space
- **Current task**: Track active Gradle processes

#### Health Status Levels
- **healthy**: All metrics within normal ranges
- **overloaded**: Resource usage above thresholds
- **failed**: SSH connectivity lost

#### Auto-scaling
- **Enabled/Disabled**: Configure via `AUTO_SCALE_ENABLED=true/false`
- **Min/Max workers**: Set limits per pool (`MIN_WORKERS_PER_POOL`, `MAX_WORKERS_PER_POOL`)
- **Scaling triggers**: CPU, memory, disk usage thresholds

### Example Workflows

#### Task-Specific Execution
```bash
# Auto-detect best pool for compilation
./scripts/worker_pool_manager.sh execute assembleDebug

# Force execution on testing pool
./scripts/worker_pool_manager.sh execute test 2 testing

# Monitor pool status
./scripts/worker_pool_manager.sh status
```

#### Pool Management
```bash
# Start continuous monitoring
./scripts/worker_pool_manager.sh monitor &

# Check worker health
./scripts/worker_pool_manager.sh status

# Optimize worker allocation
./scripts/worker_pool_manager.sh execute assemble 4 compilation
```

---

## cache_manager.sh

### Purpose
Manages local and remote build caches with optimization and cleanup.

### Usage
```bash
./scripts/cache_manager.sh [command] [options]
```

### Commands

#### Interactive Mode (default)
Menu-driven cache management:
```bash
./scripts/cache_manager.sh
./scripts/cache_manager.sh interactive
```

#### Start Cache Server
Start remote cache server:
```bash
./scripts/cache_manager.sh start
```

#### Stop Cache Server
Stop remote cache server:
```bash
./scripts/cache_manager.sh stop
```

#### Cache Status
Check cache health and status:
```bash
./scripts/cache_manager.sh status
```

#### Clean Caches
Clean local and remote caches:
```bash
./scripts/cache_manager.sh clean
```

#### Sync Caches
Synchronize local and remote caches:
```bash
./scripts/cache_manager.sh sync
```

#### Optimize Cache
Optimize cache size and organization:
```bash
./scripts/cache_manager.sh optimize
```

#### Configure Gradle
Update Gradle configuration for caching:
```bash
./scripts/cache_manager.sh configure
```

#### Generate Statistics
Generate cache performance statistics:
```bash
./scripts/cache_manager.sh stats
```

### Cache Types

#### Local Cache
- **Location**: `$HOME/.gradle-build-cache/local`
- **Purpose**: Fast local storage of build artifacts
- **Optimization**: TTL-based cleanup, size management
- **Persistence**: Survives system reboots

#### Remote Cache
- **Implementation**: Python HTTP server
- **Port**: Configurable (default: 8080)
- **Protocol**: HTTP GET/PUT for cache operations
- **Authentication**: Basic auth (configurable)

### Configuration

#### Cache Configuration File
Cache settings stored in `cache_config`:
```bash
# Cache Manager Configuration
CACHE_SIZE="10g"              # Maximum cache size
CACHE_TTL="30d"               # Cache time-to-live
CLEANUP_THRESHOLD="80"         # Cleanup trigger (% full)
REMOTE_ENABLED="false"         # Enable remote cache
REMOTE_URL="http://localhost:8080/cache"  # Remote cache URL
CACHE_ENABLED="true"           # Enable caching
```

#### Gradle Integration
Automatic `gradle.properties` updates:
```properties
# Distributed Gradle Build Cache Configuration
org.gradle.caching=true
org.gradle.cache.fixed-size=true
org.gradle.cache.max-size=10g

# Remote cache configuration (when enabled)
systemProp.org.gradle.caching.remote.enabled=true
systemProp.org.gradle.caching.remote.url=http://localhost:8080/cache
systemProp.org.gradle.caching.remote.push=true
systemProp.org.gradle.caching.remote.allow-insecure=true
```

### Cache Operations

#### Size Management
- **Monitoring**: Real-time cache size tracking
- **Thresholds**: Configurable size limits
- **Cleanup**: Automatic LRU-based cleanup
- **Optimization**: Intelligent cache organization

#### Health Checks
- **Directory verification**: Ensure cache directories exist and are writable
- **Disk space**: Monitor available storage
- **Permissions**: Verify read/write access
- **Connectivity**: Test remote cache accessibility

#### Performance Metrics
- **Hit rate calculation**: Track cache effectiveness
- **Size trends**: Monitor cache growth patterns
- **Access patterns**: Analyze cache usage
- **Cleanup efficiency**: Measure optimization success

### Example Workflows

#### Basic Cache Setup
```bash
# Interactive configuration
./scripts/cache_manager.sh

# Start cache server
./scripts/cache_manager.sh start

# Configure Gradle
./scripts/cache_manager.sh configure
```

#### Cache Maintenance
```bash
# Check cache status
./scripts/cache_manager.sh status

# Optimize cache
./scripts/cache_manager.sh optimize

# Generate statistics
./scripts/cache_manager.sh stats
```

---

## health_checker.sh

### Purpose
Monitors system health and provides automated issue detection.

### Usage
```bash
./scripts/health_checker.sh [command] [options]
```

### Commands

#### Health Check (default)
Perform comprehensive health check:
```bash
./scripts/health_checker.sh
./scripts/health_checker.sh check
```

#### Continuous Monitoring
Start continuous health monitoring:
```bash
./scripts/health_checker.sh monitor [interval]
```

Example:
```bash
# Monitor every 60 seconds
./scripts/health_checker.sh monitor 60
```

#### Generate Health Report
Create detailed health report:
```bash
./scripts/health_checker.sh report
```

#### Interactive Mode
Menu-driven health management:
```bash
./scripts/health_checker.sh interactive
```

### Health Monitoring

#### System Metrics
- **CPU usage**: Monitor processor utilization
- **Memory usage**: Track RAM consumption
- **Disk usage**: Monitor storage space
- **Network I/O**: Track network activity
- **Load average**: System load patterns

#### Worker Health
- **SSH connectivity**: Verify worker accessibility
- **Resource usage**: CPU, memory, disk on workers
- **Current tasks**: Track active Gradle processes
- **Uptime**: Monitor worker availability

#### Build Performance
- **Build duration**: Track build time trends
- **Success/failure rates**: Monitor build reliability
- **Cache hit rates**: Track cache effectiveness
- **Resource consumption**: Build-time resource usage

### Alert Thresholds

#### Default Thresholds
```bash
CPU_WARNING=80          # CPU usage warning threshold (%)
CPU_CRITICAL=95         # CPU usage critical threshold (%)
MEMORY_WARNING=85       # Memory usage warning threshold (%)
MEMORY_CRITICAL=95      # Memory usage critical threshold (%)
DISK_WARNING=85        # Disk usage warning threshold (%)
DISK_CRITICAL=95       # Disk usage critical threshold (%)
BUILD_TIME_WARNING=1800 # Build time warning threshold (seconds)
FAILURE_RATE_WARNING=20 # Failure rate warning threshold (%)
```

#### Custom Thresholds
Configure thresholds in `health_config`:
```bash
CPU_WARNING=75
CPU_CRITICAL=90
MEMORY_WARNING=80
MEMORY_CRITICAL=92
# ... other thresholds
```

### Alert System

#### Alert Levels
- **INFO**: General information messages
- **WARNING**: Performance or resource warnings
- **CRITICAL**: System failure or critical resource exhaustion

#### Alert Destinations
- **Console**: Real-time console output with color coding
- **Log files**: Persistent alert storage
- **Slack integration**: Webhook-based notifications
- **Email alerts**: SMTP-based notifications

#### Alert Triggers
- **Resource thresholds**: CPU, memory, disk limits
- **Worker failures**: Connectivity or overload issues
- **Build problems**: Slow builds, high failure rates
- **Cache issues**: Low hit rates, storage problems

### Health Reports

#### Report Content
- **Current status**: Overall system health
- **System information**: OS, Java, Gradle versions
- **Recent issues**: Last 50 health alerts
- **Build history**: Recent build performance data
- **Trends analysis**: Performance patterns over time

#### Report Formats
- **Text reports**: Human-readable health summaries
- **Log files**: Detailed timestamped events
- **Metrics data**: Machine-readable performance data
- **Historical trends**: Long-term health patterns

### Example Workflows

#### Basic Health Check
```bash
# Perform health check
./scripts/health_checker.sh check

# View recent issues
./scripts/health_checker.sh interactive
# Select option 4 to view recent issues
```

#### Continuous Monitoring
```bash
# Start continuous monitoring
./scripts/health_checker.sh monitor 300 &

# Generate weekly report
./scripts/health_checker.sh report
```

#### Threshold Configuration
```bash
# Interactive threshold adjustment
./scripts/health_checker.sh interactive
# Select option 5 to configure thresholds
```

---

## Integration Examples

### Complete Build Workflow

#### 1. Setup and Configuration
```bash
# Configure worker pools
./scripts/worker_pool_manager.sh

# Setup cache
./scripts/cache_manager.sh

# Configure health monitoring
./scripts/health_checker.sh
```

#### 2. Performance Optimization
```bash
# Analyze current performance
./scripts/performance_analyzer.sh test

# Optimize based on results
./scripts/worker_pool_manager.sh execute assembleDebug 8 compilation
```

#### 3. Continuous Operation
```bash
# Start all monitoring services
./scripts/health_checker.sh monitor &
./scripts/worker_pool_manager.sh monitor &

# Execute optimized builds
./scripts/worker_pool_manager.sh execute test
```

#### 4. Maintenance
```bash
# Weekly maintenance
./scripts/cache_manager.sh optimize
./scripts/health_checker.sh report
./scripts/performance_analyzer.sh analyze
```

### Automation Script Example

```bash
#!/bin/bash
# automated_build_pipeline.sh

echo "=== Automated Build Pipeline ==="

# Health check first
if ! ./scripts/health_checker.sh check; then
    echo "Health check failed - aborting build"
    exit 1
fi

# Optimize cache
./scripts/cache_manager.sh optimize

# Execute build with optimal workers
./scripts/worker_pool_manager.sh execute assembleRelease

# Post-build cleanup
./scripts/cache_manager.sh sync

# Generate performance report
./scripts/performance_analyzer.sh analyze

echo "Build pipeline completed"
```

---

## Troubleshooting

### Common Issues

#### Script Permissions
```bash
# Ensure scripts are executable
chmod +x scripts/*.sh
```

#### Dependencies
```bash
# Check for required tools
command -v bc >/dev/null || echo "bc calculator required"
command -v timeout >/dev/null || echo "timeout command required"
```

#### Configuration Files
```bash
# Reset configurations
rm cache_config health_config workers.conf
# Scripts will create defaults on next run
```

#### Logging
```bash
# Check logs for issues
tail -f performance_results/analyzer.log
tail -f health_check.log
tail -f cache_manager.log
```

### Performance Tips

#### Monitoring Overhead
- Adjust health check intervals for production vs development
- Use minimal logging for high-frequency monitoring
- Cache health results to reduce system load

#### Resource Allocation
- Monitor script resource usage during operation
- Adjust worker allocation based on system capabilities
- Consider dedicated monitoring machine for large deployments

#### Network Considerations
- Optimize SSH connection timeouts for remote workers
- Use compression for network-based operations
- Monitor network bandwidth during distributed operations

---

This scripts reference provides comprehensive documentation for all management tools included in the Distributed Gradle Building System. Each script is designed to work independently or as part of an integrated workflow for optimal distributed building performance.