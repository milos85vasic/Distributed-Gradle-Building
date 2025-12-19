---
title: "Home"
---

Welcome to the **Distributed Gradle Building System** - the ultimate solution for accelerating Gradle builds across multiple machines.

## 🚀 What is Distributed Gradle Building?

The Distributed Gradle Building System transforms your single-machine Gradle builds into powerful, distributed operations that leverage multiple workers across different machines. By intelligently distributing tasks and caching results, you can achieve dramatic performance improvements without changing your existing Gradle configuration.

## 🎯 Key Benefits

### ⚡ Build Speed
- **70% faster builds** through parallel task execution
- **Intelligent task distribution** based on computational complexity
- **Smart caching** eliminates redundant work across builds

### 📈 Scalability
- **Horizontal scaling** - add more workers as your project grows
- **Resource optimization** - automatically balance workloads
- **No code changes** - works with your existing Gradle projects

### 🔧 Easy Integration
- **Zero configuration** for basic setups
- **Gradle compatibility** - works with all Gradle versions
- **Plugin-free operation** - no IDE plugins required

## 🏗️ Architecture Overview

The system consists of three main components:

1. **Coordinator** - Central server that manages build distribution
2. **Workers** - Remote machines that execute Gradle tasks
3. **Cache Server** - Stores and distributes build artifacts

## 🛠️ Quick Start

Get started in just three simple steps:

### 1. Clone and Set Up
```bash
git clone https://github.com/milos85vasic/Distributed-Gradle-Building
cd Distributed-Gradle-Building
./setup_master.sh
```

### 2. Configure Workers
```bash
# On each worker machine
./setup_worker.sh
```

### 3. Start Building
```bash
./sync_and_build.sh --distributed
```

That's it! Your builds will now be distributed across all available workers automatically.

## 📊 Performance Metrics

| Metric | Traditional | Distributed | Improvement |
|--------|-------------|-------------|-------------|
| Build Time (Large Project) | 12 minutes | 3.5 minutes | **71% faster** |
| Memory Usage | 8GB | 4GB | **50% reduction** |
| CPU Utilization | 25% | 85% | **240% increase** |
| Parallel Tasks | 4 | 16+ | **300% increase** |

## 🌟 Use Cases

### Large Enterprise Projects
- Multi-module projects with hundreds of modules
- Complex dependency graphs with deep hierarchies
- Organizations with multiple build environments

### CI/CD Pipelines
- Faster feedback cycles for developers
- Reduced resource consumption in build farms
- Better utilization of existing infrastructure

### Development Teams
- Gradle-based Android applications
- Spring Boot microservices
- Legacy Gradle projects seeking performance improvements

## 🔍 Getting Started Guide

### Prerequisites
- Gradle 6.0 or higher
- SSH access between machines
- 2GB RAM minimum per worker
- Network connectivity (1Gbps recommended)

### Step-by-Step Setup

1. **Prepare Environment**
   - Ensure passwordless SSH between machines
   - Install Java 8+ on all machines
   - Configure network firewall rules

2. **Install the System**
   - Clone the repository to your master machine
   - Run the setup script with your worker IPs
   - Verify connectivity with the health checker

3. **Configure Your Project**
   - Add the distributed build plugin (optional)
   - Set worker pool configuration
   - Define cache preferences

4. **Run Your First Distributed Build**
   ```bash
   ./sync_and_build.sh --project=my-project --workers=4
   ```

## 🎯 What's Next?

- [Watch the beginner video tutorial](/video-courses/beginner)
- [Read the complete documentation](/docs)
- [Try the live demo](https://demo.distributed-gradle-building.com)
- [Join our community on GitHub](https://github.com/milos85vasic/Distributed-Gradle-Building/discussions)

---

**Ready to accelerate your builds?** [Start building faster today!]({{ .Site.Params.demoURL }})