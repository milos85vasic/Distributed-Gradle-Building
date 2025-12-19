---
title: "User Guide"
weight: 20
---

# User Guide

This guide will help you effectively use the Distributed Gradle Building System in your development workflow.

## Basic Usage

### Running Your First Distributed Build

```bash
# Basic distributed build
./sync_and_build.sh

# With specific Gradle project
./sync_and_build.sh --project /path/to/gradle/project

# With custom worker count
./sync_and_build.sh --workers 6

# With specific build configuration
./sync_and_build.sh --config release
```

### Monitoring Build Progress

```bash
# Monitor active builds
./scripts/health_checker.sh --status

# View worker utilization
./scripts/performance_analyzer.sh --workers

# Follow build logs
tail -f logs/build.log
```

## Project Configuration

### Supported Project Types

The Distributed Gradle Building System supports:

- **Java Projects** (Spring Boot, Android, Enterprise Java)
- **Multi-Module Projects** (monorepos, microservices)
- **Gradle Plugins** (custom plugins and tooling)
- **Kotlin Projects** (JVM, Android, Multiplatform)

### Project Requirements

Your Gradle project should meet these requirements:

```gradle
// Minimum Gradle version
plugins {
    id 'java' // or 'java-library', 'application', 'com.android.application', etc.
}

// Ensure dependency resolution is compatible
repositories {
    mavenCentral()
    google() // for Android projects
}

// Optional: Configure for optimal distributed performance
tasks.withType(JavaCompile) {
    options.incremental = true
    options.fork = true
}
```

### Configuration File

Create `.distributed-gradle.yml` in your project root:

```yaml
# Distributed Gradle Configuration
project:
  name: "my-application"
  type: "java"  # java, android, kotlin
  modules: ["app", "core", "shared"]

distribution:
  workers:
    min: 2
    max: 8
    timeout: 1800  # 30 minutes
  
  strategy: "dependency-aware"  # round-robin, load-balanced, dependency-aware
  
  caching:
    enabled: true
    strategy: "full"  # full, incremental, selective
    
  modules:
    parallel: true
    max_parallel: 4
    
monitoring:
  log_level: "INFO"
  metrics_enabled: true
  performance_tracking: true

optimization:
  compile_parallelism: true
  test_distribution: true
  resource_monitoring: true
```

## Advanced Usage

### Build Strategies

#### 1. Round-Robin Distribution

```bash
# Distribute tasks evenly across all workers
./sync_and_build.sh --strategy round-robin
```

#### 2. Load-Balanced Distribution

```bash
# Distribute based on worker capacity and current load
./sync_and_build.sh --strategy load-balanced
```

#### 3. Dependency-Aware Distribution

```bash
# Optimize distribution based on project dependency graph
./sync_and_build.sh --strategy dependency-aware
```

### Caching Strategies

#### Full Caching

```yaml
caching:
  enabled: true
  strategy: "full"
  cache_dependencies: true
  cache_outputs: true
  invalidation_strategy: "semantic"  # semantic, timestamp, always
```

#### Incremental Caching

```yaml
caching:
  enabled: true
  strategy: "incremental"
  only_changed: true
  cache_depth: 3  # dependency levels to cache
```

#### Selective Caching

```yaml
caching:
  enabled: true
  strategy: "selective"
  include_tasks: ["compile", "test", "jar"]
  exclude_modules: ["integration-test"]
  size_limit_mb: 1000
```

## Performance Optimization

### Worker Pool Management

```bash
# Add temporary workers
./scripts/worker_pool_manager.sh --add worker-temp1 --host 192.168.1.50

# Remove workers
./scripts/worker_pool_manager.sh --remove worker-temp1

# Scale up for big builds
./scripts/worker_pool_manager.sh --scale --target 12

# Optimize for current load
./scripts/worker_pool_manager.sh --optimize --metrics
```

### Resource Allocation

```yaml
# Resource optimization configuration
resources:
  workers:
    default_memory: "4g"
    max_memory: "8g"
    cpu_cores: 4
    disk_space: "20g"
  
  build:
    compile_memory: "2g"
    test_memory: "1g"
    jvm_options: [
      "-XX:+UseG1GC",
      "-XX:+UseStringDeduplication",
      "-XX:+OptimizeStringConcat"
    ]
    
  network:
    bandwidth_limit: "1Gbps"
    compression: true
    chunk_size: "10MB"
```

### Gradle Optimization

```gradle
// Optimize Gradle for distributed builds
org.gradle.parallel=true
org.gradle.caching=true
org.gradle.daemon=true
org.gradle.jvmargs=-Xmx2048m -XX:+UseG1GC
org.gradle.configureondemand=true

// Optimize for your specific project
android {
    compileOptions {
        incremental = true
    }
    
    testOptions {
        maxParallelForks = 4
    }
}

java {
    sourceCompatibility = JavaVersion.VERSION_11
    targetCompatibility = JavaVersion.VERSION_11
}
```

## Integration with CI/CD

### Jenkins Integration

```groovy
pipeline {
    agent any
    
    environment {
        DISTRIBUTED_GRADLE_HOME = '/opt/distributed-gradle'
    }
    
    stages {
        stage('Build') {
            steps {
                sh './sync_and_build.sh --workers 8 --config production'
            }
        }
        
        stage('Test') {
            steps {
                sh './sync_and_build.sh --stage test --workers 6'
            }
        }
    }
    
    post {
        always {
            archiveArtifacts artifacts: 'build/libs/*.jar', fingerprint: true
            junit 'build/test-results/test/TEST-*.xml'
        }
    }
}
```

### GitHub Actions Integration

```yaml
name: Distributed Gradle Build

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Distributed Gradle
      run: |
        git clone https://github.com/milos85vasic/Distributed-Gradle-Building
        cd Distributed-Gradle-Building
        ./setup_master.sh
    
    - name: Run Distributed Build
      run: |
        cd Distributed-Gradle-Building
        ./sync_and_build.sh --project .. --workers 4
```

### Docker Integration

```dockerfile
FROM gradle:7.4.2-jdk11

# Install distributed gradle
RUN git clone https://github.com/milos85vasic/Distributed-Gradle-Building
WORKDIR /Distributed-Gradle-Building
RUN ./setup_master.sh

# Configure workers (environment variables)
ENV WORKER_COUNT=4
ENV CACHE_SIZE=10GB

# Set entrypoint
ENTRYPOINT ["./sync_and_build.sh"]
```

## Monitoring and Debugging

### Real-time Monitoring

```bash
# Start monitoring dashboard
./scripts/performance_analyzer.sh --dashboard --port 8080

# Monitor specific build
./scripts/health_checker.sh --monitor build-12345

# Worker performance metrics
./scripts/performance_analyzer.sh --workers --detailed
```

### Log Analysis

```bash
# Analyze build logs for performance issues
./scripts/performance_analyzer.sh --analyze-logs build-2023-12-01.log

# Find bottlenecks
./scripts/performance_analyzer.sh --bottlenecks --output bottlenecks.json

# Performance trends
./scripts/performance_analyzer.sh --trends --days 30
```

### Debug Mode

```bash
# Enable debug logging
./sync_and_build.sh --debug --verbose

# Test specific module
./sync_and_build.sh --module core --debug

# Dry run (no actual execution)
./sync_and_build.sh --dry-run --plan-only
```

## Troubleshooting

### Common Issues

#### Slow Build Times

```bash
# Check worker utilization
./scripts/performance_analyzer.sh --workers --utilization

# Optimize worker pool
./scripts/worker_pool_manager.sh --optimize

# Check network connectivity
./scripts/health_checker.sh --network-test
```

#### Build Failures

```bash
# Check build logs
./scripts/health_checker.sh --logs --level ERROR

# Test individual workers
./scripts/health_checker.sh --worker worker-1 --test

# Validate project configuration
./go/client/example/main.go --validate-project /path/to/project
```

#### Memory Issues

```bash
# Check memory usage
./scripts/performance_analyzer.sh --memory --detailed

# Adjust JVM options
./sync_and_build.sh --jvm-options "-Xmx6g -XX:+UseG1GC"

# Monitor memory trends
./scripts/performance_analyzer.sh --memory --trends
```

## Best Practices

### Project Structure

```
my-project/
├── .distributed-gradle.yml    # Configuration file
├── build.gradle               # Main build file
├── settings.gradle            # Gradle settings
├── gradle.properties          # Gradle properties
└── modules/                   # Project modules
    ├── core/
    ├── app/
    └── shared/
```

### Performance Tips

1. **Enable Gradle caching** for maximum performance
2. **Optimize module dependencies** to enable parallelization
3. **Use appropriate worker count** based on project size
4. **Monitor resource usage** to identify bottlenecks
5. **Regular maintenance** of cache and worker pools

### Security Considerations

```yaml
security:
  ssh:
    key_rotation: true
    access_control: true
    audit_logging: true
    
  network:
    encryption: true
    authentication: true
    firewall_rules: true
```

---

**Ready to learn more?** Check our [advanced configuration](/docs/advanced-config) guide or [watch our video tutorials](/video-courses).