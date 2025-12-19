# Distributed Gradle Building System

## 🚀 Overview

The Distributed Gradle Building System enables **true distributed compilation** across multiple worker machines, utilizing their combined CPU cores and memory for faster builds.

> **⚠️ IMPORTANT:** The original `sync_and_build.sh` has been upgraded from a simple file sync script to use **real distributed building** that actually utilizes worker resources.

## 🎯 Key Features

- ✅ **True Distributed Building** - Tasks executed on multiple workers
- ✅ **Resource Utilization** - CPU and memory from all workers used
- ✅ **Load Balancing** - Intelligent task distribution based on worker capacity
- ✅ **Real-time Monitoring** - Track CPU, memory, and network usage per worker
- ✅ **Build Artifact Collection** - Consolidate results from all workers
- ✅ **Fault Tolerance** - Automatic retry and worker failure handling
- ✅ **Performance Metrics** - Detailed build statistics and optimization insights

## 📋 Architecture

```
┌─────────────┐     SSH/Rsync     ┌─────────────┐
│   Master    │ ◄──────────────► │  Worker 1   │
│             │                 │ 8 cores, 16GB│
│ distributes │                 └─────────────┘
│   tasks     │                      │
│             │                      │
└─────────────┘                      ▼
      │                     ┌─────────────┐
      │                     │  Worker 2   │
      ▼                     │ 8 cores, 16GB│
┌─────────────┐             └─────────────┘
│  Worker 3   │                      │
│ 8 cores, 16GB│                      │
└─────────────┘                      ▼
                               ┌─────────────┐
                               │  Worker N   │
                               │ 8 cores, 16GB│
                               └─────────────┘
```

## 🛠️ Installation

### Choose Your Implementation

This project offers **two complete implementations**:

#### **Bash Implementation** (Quick Start)
- Simple to set up and understand
- Works with existing SSH infrastructure
- Ideal for small to medium teams
- **Get Started in 5 minutes** → [Quick Start Guide](QUICK_START.md)

#### **Go Implementation** (Enterprise Scale)
- RESTful API and RPC interfaces
- Advanced monitoring and caching
- Horizontal scaling capabilities
- Microservices architecture
- **Get Started** → [Go Deployment Guide](docs/GO_DEPLOYMENT.md)

---

### Prerequisites (Both Implementations)

- Linux-based system (Ubuntu/Debian recommended)
- OpenSSH server and client
- Java 17+ JDK
- Passwordless SSH between machines
- Gradle project structure

## 🚀 Quick Start

### Choose Your Implementation:

#### **Option 1: Bash Implementation** (5-minute setup)
1. **Master Setup:**
   ```bash
   ./setup_master.sh /path/to/gradle/project
   ```

2. **Worker Setup** (run on each worker):
   ```bash
   ./setup_worker.sh /path/to/gradle/project
   ```

3. **Start Distributed Build:**
   ```bash
   cd /path/to/gradle/project
   ./sync_and_build.sh assemble
   ```

#### **Option 2: Go Implementation** (Enterprise scale)
1. **Deploy Services:**
   ```bash
   # See docs/GO_DEPLOYMENT.md for detailed setup
   cd go
   go mod tidy
   ./deploy_all_services.sh
   ```

2. **Configure Client:**
   ```bash
   # Use Go client or REST API
   ./client/example/main
   ```

📖 **Need help?** → [Complete Setup Guide](docs/SETUP_GUIDE.md)

### 1. Passwordless SSH Configuration

Run the automated SSH setup:

```bash
./setup_passwordless_ssh.sh
```

Or manually configure:

```bash
# On master, generate SSH key if not exists
ssh-keygen -t rsa -b 4096

# Copy to each worker
ssh-copy-id user@worker1-ip
ssh-copy-id user@worker2-ip
# ... for all workers
```

### 2. Master Machine Setup

The master coordinates the distributed build:

```bash
./setup_master.sh /path/to/your/gradle/project
```

This will:
- Install required packages (rsync, openjdk-17-jdk, jq)
- Test SSH connectivity to all workers
- Detect system resources
- Create configuration in `.gradlebuild_env`
- Set up distributed build directories

### 3. Worker Machine Setup

Each worker contributes its CPU and memory:

```bash
./setup_worker.sh /path/to/your/gradle/project
```

This will:
- Install required packages
- Detect worker resources (CPU, memory, disk)
- Create worker directories and monitoring
- Start health monitoring service
- Configure worker for distributed tasks

## 🚀 Usage

### Basic Distributed Build

```bash
cd /path/to/gradle/project
./sync_and_build.sh assemble    # Distributed assembly
./sync_and_build.sh test        # Distributed testing
./sync_and_build.sh clean       # Clean distributed artifacts
```

### Advanced Usage

```bash
# Use specific task
./sync_and_build.sh compileJava

# Direct distributed build (bypasses sync_and_build.sh wrapper)
./distributed_gradle_build.sh assemble

# Monitor distributed build in real-time
tail -f .distributed/logs/distributed_build.log

# View worker resource usage
./distributed_gradle_build.sh --status
```

## 📊 Resource Utilization

### Before (Local Parallel Build)
- **CPU:** 1 machine × 8 cores = 8 cores
- **Memory:** 1 machine × 16GB = 16GB
- **Concurrency:** Local threads only

### After (Distributed Build with 4 Workers)
- **CPU:** 4 machines × 8 cores = 32 cores (**4× improvement**)
- **Memory:** 4 machines × 16GB = 64GB (**4× improvement**)
- **Concurrency:** 4 machines × parallel tasks

## 🔍 Monitoring and Metrics

### Real-time Monitoring

```bash
# Watch distributed build progress
watch -n 1 'cat .distributed/metrics/build_metrics.json | jq'

# Monitor individual workers
for worker in $WORKER_IPS; do
    ssh $worker "cat $PROJECT_DIR/.distributed/metrics/worker_metrics.json"
done
```

### Performance Metrics

The system tracks:
- Build duration and task completion rates
- CPU and memory utilization per worker
- Network transfer and synchronization time
- Success/failure rates and retry statistics
- Artifact collection efficiency

### Available Metrics

```bash
# Build metrics
jq '.metrics' .distributed/metrics/build_metrics.json

# Worker metrics
jq '.worker_cpu_usage' .distributed/metrics/build_metrics.json
jq '.worker_memory_usage' .distributed/metrics/build_metrics.json

# Task distribution
jq '.tasks' .distributed/metrics/build_metrics.json
```

## 🧪 Testing and Verification

### Run Comprehensive Tests

```bash
# Quick verification
./tests/quick_distributed_verification.sh

# Comprehensive test suite
./tests/comprehensive/test_distributed_build.sh all

# Integration tests
./tests/integration/test_distributed_build_verification.sh all
```

### Test Categories

1. **Task Analysis Tests** - Verify task decomposition
2. **Resource Utilization Tests** - Validate CPU/memory usage
3. **Distributed Execution Tests** - Test remote task execution
4. **Artifact Collection Tests** - Verify build result consolidation
5. **Performance Tests** - Compare local vs distributed builds
6. **Error Handling Tests** - Test fault tolerance and recovery

## 🔧 Configuration

### Environment Variables

Create `.gradlebuild_env` in your project root:

```bash
export WORKER_IPS="192.168.1.101 192.168.1.102 192.168.1.103"
export MAX_WORKERS=8
export PROJECT_DIR="/path/to/your/project"
export DISTRIBUTED_BUILD=true
export WORKER_TIMEOUT=300
export BUILD_CACHE_ENABLED=true
```

### Gradle Configuration

Add to `gradle.properties`:

```properties
# Enable distributed build optimizations
org.gradle.parallel=true
org.gradle.workers.max=8
org.gradle.caching=true
org.gradle.configureondemand=true

# Network optimizations for distributed builds
org.gradle.daemon=true
org.gradle.jvmargs=-Xmx4g -XX:+UseG1GC
```

## 🚨 Troubleshooting

### Common Issues

1. **SSH Connection Failed**
   ```bash
   # Test connectivity
   ssh worker-ip "echo 'OK'"
   
   # Setup passwordless SSH
   ssh-copy-id user@worker-ip
   ```

2. **Workers Not Utilized**
   ```bash
   # Check worker status
   ./distributed_gradle_build.sh --check-workers
   
   # Verify SSH and Gradle on workers
   ssh worker-ip "cd $PROJECT_DIR && ./gradlew --version"
   ```

3. **Build Artifacts Missing**
   ```bash
   # Check artifact collection
   ls -la build/distributed/
   
   # Force collection
   ./distributed_gradle_build.sh --collect-artifacts
   ```

### Debug Mode

Enable detailed logging:

```bash
export DEBUG=true
./sync_and_build.sh assemble
```

### Log Locations

- Master logs: `.distributed/logs/`
- Worker logs: `$PROJECT_DIR/.distributed/logs/`
- Build metrics: `.distributed/metrics/build_metrics.json`

## 📈 Performance Optimization

### Scaling Guidelines

| Workers | CPU Cores | Memory | Expected Speedup |
|---------|-----------|--------|-----------------|
| 1       | 8         | 16GB   | 1.0x (baseline) |
| 2       | 16        | 32GB   | 1.7-1.9x        |
| 4       | 32        | 64GB   | 2.8-3.5x        |
| 8       | 64        | 128GB  | 4.5-6.0x        |

### Optimization Tips

1. **Balance Worker Resources** - Use similar spec machines
2. **Network Optimization** - Fast network between master/workers
3. **Build Cache** - Enable distributed build cache
4. **Task Granularity** - Smaller tasks distribute better
5. **Memory Allocation** - Allocate sufficient heap per worker

## 🔒 Security Considerations

- Use SSH key authentication, not passwords
- Limit SSH access to specific users
- Monitor network traffic between machines
- Regular security updates on all nodes
- Isolate build environment from production systems

## 📚 API Reference

### sync_and_build.sh

Upgraded wrapper that calls `distributed_gradle_build.sh`.

```bash
./sync_and_build.sh [task] [options]
```

### distributed_gradle_build.sh

Main distributed build engine.

```bash
./distributed_gradle_build.sh [task] [--verbose] [--dry-run]
```

**Options:**
- `assemble` - Build project (default)
- `test` - Run tests distributed
- `clean` - Clean distributed artifacts
- `--status` - Show worker status
- `--check-workers` - Verify worker connectivity
- `--collect-artifacts` - Force artifact collection

### Worker API

Workers expose status via JSON API:

```bash
ssh worker-ip "cat $PROJECT_DIR/.distributed/metrics/worker_metrics.json"
```

## 🔄 Migration from Local Builds

### Step 1: Backup Current Setup
```bash
cp sync_and_build.sh sync_and_build.sh.backup
```

### Step 2: Setup Workers
```bash
./setup_master.sh /path/to/project
# On each worker:
./setup_worker.sh /path/to/project
```

### Step 3: Test Migration
```bash
./tests/quick_distributed_verification.sh
```

### Step 4: Update CI/CD
Replace local build commands with distributed ones:
```bash
# Before
./gradlew assemble

# After  
./sync_and_build.sh assemble
```

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 🌐 Connected Resources

### 📚 Documentation
- **[Complete Documentation](docs/)** - Comprehensive guides and references
- **[API Reference](docs/API_REFERENCE.md)** - RESTful API documentation
- **[Setup Guide](docs/SETUP_GUIDE.md)** - Step-by-step setup instructions
- **[User Guide](docs/USER_GUIDE.md)** - Day-to-day usage guide

### 🔗 **Component Connections**
- **[📡 CONNECTIONS.md](CONNECTIONS.md)** - How all components work together
- **[🚀 QUICK_START.md](QUICK_START.md)** - 5-minute setup guide

### 🌍 Website & Community
- **[Project Website](website/)** - Full documentation site with tutorials
- **[Live Demo](https://demo.distributed-gradle-building.com)** - Try it online
- **[Video Courses](website/content/video-courses/)** - Beginner to advanced tutorials

### 🔧 Go Implementation
- **[Go Services](go/)** - Enterprise-scale microservices implementation
- **[Go Deployment Guide](docs/GO_DEPLOYMENT.md)** - Deploy Go services
- **[Client Library](go/client/)** - Go client with examples

### 🧪 Testing
- **[Test Framework](tests/)** - Comprehensive test suite
- **[Run All Tests](scripts/run-all-tests.sh)** - Complete test runner
- **[Validate Connections](scripts/validate_connections.sh)** - Check all component connections

## 📞 Support

- **Issues:** [GitHub Issues](https://github.com/your-repo/issues)
- **Documentation:** [docs/](docs/)
- **Website:** [website/](website/)
- **Discussions:** [GitHub Discussions](https://github.com/your-repo/discussions)

---

**🚀 Ready to accelerate your Gradle builds with distributed computing?**
- 🚀 **[Quick Start](QUICK_START.md)** - Get running in 5 minutes
- 🏗️ **[Go Enterprise](go/)** - Scale with microservices
- 🌐 **[Full Website](website/)** - Complete documentation and tutorials