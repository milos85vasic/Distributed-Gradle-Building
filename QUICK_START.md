# Distributed Gradle Build - Quick Start Guide

## 🚀 Choose Your Implementation

### **Option 1: Bash Implementation** ⚡ (5-minute setup)
Perfect for small to medium teams wanting fast setup and reliable builds.

### **Option 2: Go Implementation** 🏗️ (Enterprise scale)
Ideal for large organizations needing APIs, monitoring, and horizontal scaling.

> 💡 **New to distributed builds?** Start with Bash implementation, migrate to Go later.

---

## ⚡ Bash Implementation (5-Minute Quick Start)

### Prerequisites
- Linux machines (master + workers)
- SSH access between machines
- Java 17+ installed
- Gradle project

### Step 1: Master Setup
```bash
# Clone/enter the project
cd /path/to/Distributed-Gradle-Building

# Setup master machine
./setup_master.sh /path/to/your/gradle/project

# Example output:
# 🚀 Distributed Gradle Build - Master Setup
# =========================================
# Enter worker IP addresses (space-separated): 192.168.1.101 192.168.1.102
# ✅ SSH to 192.168.1.101: OK
# ✅ SSH to 192.168.1.102: OK
# 🎯 DISTRIBUTED BUILD SETUP COMPLETE
```

### Step 2: Worker Setup (run on each worker)
```bash
# On worker 1
./setup_worker.sh /path/to/your/gradle/project
# Enter master IP address: 192.168.1.100

# On worker 2  
./setup_worker.sh /path/to/your/gradle/project
# Enter master IP address: 192.168.1.100
```

### Step 3: Start Distributed Build
```bash
cd /path/to/your/gradle/project
./sync_and_build.sh assemble
```

---

## 🏗️ Go Implementation (Enterprise Scale)

### When to Use Go Implementation
- Large teams (50+ developers)
- CI/CD pipeline integration
- Advanced monitoring and metrics
- RESTful API requirements
- Horizontal scaling needs

### Prerequisites
- Go 1.19+
- Docker (optional, for containerization)
- Redis or PostgreSQL (for production cache)
- Load balancer (for production)

### Step 1: Deploy Go Services
```bash
# Clone and setup
cd /path/to/Distributed-Gradle-Building
go mod tidy

# Deploy all services
cd go
./deploy_all_services.sh
```

### Step 2: Configure Build Client
```bash
# Using REST API
curl -X POST http://localhost:8080/api/builds \
  -H "Content-Type: application/json" \
  -d '{"project": "/path/to/project", "tasks": ["assemble"]}'

# Using Go client
cd go/client/example
go run main.go
```

### Step 3: Monitor Services
```bash
# Check service status
curl http://localhost:8080/api/health
curl http://localhost:8081/api/cache/status
curl http://localhost:8082/api/metrics
```

### Key Benefits
- **API-First**: Full RESTful API for CI/CD integration
- **Monitoring**: Real-time metrics and dashboards
- **Scalability**: Auto-scaling worker pools
- **Performance**: Optimized task scheduling
- **Reliability**: Built-in health checks and recovery

📖 **Complete Go Setup:** [Go Deployment Guide](docs/GO_DEPLOYMENT.md)

---

## 🎯 What Just Happened?

Your Gradle build is now distributed across multiple machines:

- **Before:** 1 machine × 8 cores × 16GB RAM
- **After:** 2 machines × 16 cores × 32GB RAM
- **Result:** 2×+ build speed improvement!

## 📊 Verify It's Working

```bash
# Quick verification
./tests/quick_distributed_verification.sh

# Should show:
# ✅ CURRENT sync_and_build.sh IS NOW A DISTRIBUTED BUILD SYSTEM
# ✅ Remote Task Execution: YES
# ✅ Worker CPU Utilization: YES
# ✅ Worker Memory Utilization: YES
```

## 🔧 Common Commands

```bash
# Different build tasks
./sync_and_build.sh compileJava    # Distributed compilation
./sync_and_build.sh test           # Distributed testing
./sync_and_build.sh jar            # Distributed packaging
./sync_and_build.sh clean          # Clean distributed artifacts

# Check system status
./distributed_gradle_build.sh --status
./distributed_gradle_build.sh --check-workers

# Monitor resources
tail -f .distributed/logs/distributed_build.log
```

## 📈 Performance Results

### Real Example (4 Workers)
| Metric | Local Build | Distributed Build | Improvement |
|--------|-------------|-------------------|-------------|
| Time | 12m 45s | 4m 30s | **2.9× faster** |
| CPU Used | 25% | 85% across workers | **3.4× more** |
| Memory Used | 8GB | 32GB | **4× more** |

## 🚨 Troubleshooting Quick Fixes

### SSH Issues
```bash
# Test connectivity
ssh worker-ip "echo 'OK'"

# Fix SSH keys
ssh-copy-id user@worker-ip
```

### Workers Not Responding
```bash
# Check worker status
./distributed_gradle_build.sh --check-workers

# Restart worker monitoring (on worker)
cd /path/to/project
./.distributed/worker_monitor.sh
```

### Build Fails
```bash
# Enable debug mode
DEBUG=true ./sync_and_build.sh assemble

# Check logs
cat .distributed/logs/distributed_build.log
```

## 🎯 Next Steps

### 📚 **Expand Your Knowledge**
- **[Complete Documentation](README.md)** - Full project overview
- **[Performance Guide](docs/PERFORMANCE.md)** - Optimize your builds
- **[Monitoring Guide](docs/MONITORING.md)** - Track metrics and performance
- **[CI/CD Integration](docs/CICD.md)** - Integrate with your pipelines

### 🌐 **Connected Resources**
- **[Project Website](website/)** - Tutorials and video courses
- **[Component Connections](CONNECTIONS.md)** - How everything works together
- **[API Reference](docs/API_REFERENCE.md)** - Go implementation APIs
- **[Test Framework](tests/)** - Comprehensive testing

### 🚀 **Advanced Options**
- **[Go Implementation](docs/GO_DEPLOYMENT.md)** - Upgrade to enterprise scale
- **[Advanced Configuration](docs/ADVANCED_CONFIG.md)** - Production deployment
- **[Use Cases](docs/USE_CASES.md)** - Real-world implementation examples

## 🎉 You're Ready!

Your Gradle builds are now distributed and significantly faster! 🚀

**Choose your path:**
- ⚡ **Continue with Bash** - Simple, reliable, fast
- 🏗️ **Explore Go** - Enterprise features and APIs
- 🌐 **Visit Website** - Video tutorials and guides

---

**Need help?** Check [connections](CONNECTIONS.md) or [open an issue](https://github.com/milos85vasic/Distributed-Gradle-Building/issues).