---
title: "Setup Guide"
weight: 10
---

# Setup Guide

This comprehensive guide will walk you through installing and configuring the Distributed Gradle Building System on your infrastructure.

## Prerequisites

### System Requirements

**Master Node:**
- Operating System: Linux (Ubuntu 18.04+, CentOS 7+, RHEL 7+)
- CPU: 4 cores minimum, 8+ cores recommended
- RAM: 8GB minimum, 16GB+ recommended
- Storage: 50GB free space
- Network: 1Gbps connection to workers
- Software: Git, Java 8+, SSH client

**Worker Nodes:**
- Operating System: Linux (Ubuntu 18.04+, CentOS 7+, RHEL 7+)
- CPU: 2 cores minimum, 4+ cores recommended
- RAM: 4GB minimum, 8GB+ recommended
- Storage: 20GB free space
- Network: 1Gbps connection to master
- Software: Java 8+, SSH server

### Network Requirements

- All nodes must be able to communicate via SSH
- Open ports: 22 (SSH), 8080-8090 (communication)
- No NAT between master and workers (or proper port forwarding)
- Stable network connection with low latency (<50ms)

## Installation Steps

### Step 1: Clone Repository

```bash
# Clone the repository on your master node
git clone https://github.com/milos85vasic/Distributed-Gradle-Building
cd Distributed-Gradle-Building
```

### Step 2: Configure Passwordless SSH

Set up SSH key authentication between master and all worker nodes:

```bash
# Generate SSH key (if not already exists)
ssh-keygen -t rsa -b 4096 -C "distributed-gradle-master"

# Copy SSH key to each worker
ssh-copy-id user@worker1-ip
ssh-copy-id user@worker2-ip
ssh-copy-id user@worker3-ip

# Test connectivity
ssh user@worker1-ip "echo 'SSH connection successful'"
```

### Step 3: Setup Master Node

Run the master setup script:

```bash
# Execute master setup
chmod +x setup_master.sh
./setup_master.sh

# Verify installation
./scripts/health_checker.sh
```

The setup script will:
- Install required dependencies
- Create configuration directories
- Set up the coordinator service
- Configure the cache server
- Generate default configuration files

### Step 4: Setup Worker Nodes

For each worker node, run the worker setup script:

```bash
# On each worker node
ssh user@worker1-ip
git clone https://github.com/milos85vasic/Distributed-Gradle-Building
cd Distributed-Gradle-Building
chmod +x setup_worker.sh
./setup_worker.sh

# Verify worker is running
./scripts/health_checker.sh --worker
```

### Step 5: Configure Workers

Edit the worker configuration file to register workers with the master:

```bash
# On master node
nano go/worker_config.json
```

Add your workers:

```json
{
  "workers": [
    {
      "id": "worker-1",
      "host": "192.168.1.101",
      "port": 8081,
      "ssh_user": "build-user",
      "ssh_key_path": "~/.ssh/id_rsa",
      "capacity": 4,
      "enabled": true
    },
    {
      "id": "worker-2", 
      "host": "192.168.1.102",
      "port": 8082,
      "ssh_user": "build-user",
      "ssh_key_path": "~/.ssh/id_rsa",
      "capacity": 4,
      "enabled": true
    }
  ]
}
```

### Step 6: Verify Installation

Test the complete setup:

```bash
# Run a test build
./tests/test_framework.sh --quick

# Check worker connectivity
./scripts/health_checker.sh --all-workers

# Verify distributed functionality
./go/client/example/main.go --test-distributed
```

## Configuration Options

### Master Configuration (go/worker_config.json)

```json
{
  "coordinator": {
    "host": "0.0.0.0",
    "port": 8080,
    "max_workers": 20,
    "heartbeat_interval": 30,
    "task_timeout": 3600
  },
  "cache": {
    "host": "0.0.0.0", 
    "port": 8090,
    "max_size_gb": 50,
    "cleanup_interval": 3600
  },
  "build": {
    "default_timeout": 1800,
    "max_concurrent_tasks": 100,
    "log_level": "INFO"
  }
}
```

### Worker Configuration (worker_config.json on each worker)

```json
{
  "worker": {
    "id": "worker-1",
    "host": "192.168.1.101",
    "port": 8081,
    "capacity": 4,
    "work_dir": "/tmp/gradle-work",
    "java_home": "/usr/lib/jvm/java-11-openjdk-amd64",
    "gradle_home": "/opt/gradle"
  },
  "build": {
    "timeout": 1800,
    "memory_limit": "4g",
    "jvm_options": "-Xmx3g -XX:+UseG1GC"
  }
}
```

## Post-Installation Tasks

### Enable Services

```bash
# Enable services to start on boot
sudo systemctl enable distributed-gradle-coordinator
sudo systemctl enable distributed-gradle-cache

# Start services immediately
sudo systemctl start distributed-gradle-coordinator
sudo systemctl start distributed-gradle-cache
```

### Configure Firewall

```bash
# Ubuntu/Debian
sudo ufw allow 22/tcp
sudo ufw allow 8080:8090/tcp
sudo ufw enable

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=22/tcp
sudo firewall-cmd --permanent --add-port=8080-8090/tcp
sudo firewall-cmd --reload
```

### Monitoring Setup

```bash
# Install monitoring tools
./scripts/performance_analyzer.sh --setup

# Configure log rotation
sudo nano /etc/logrotate.d/distributed-gradle
```

## Troubleshooting Common Issues

### SSH Connection Issues

```bash
# Test SSH connectivity manually
ssh -v user@worker-ip "echo test"

# Check SSH key permissions
ls -la ~/.ssh/
chmod 600 ~/.ssh/id_rsa
chmod 644 ~/.ssh/id_rsa.pub
```

### Service Startup Issues

```bash
# Check service status
sudo systemctl status distributed-gradle-coordinator
sudo systemctl status distributed-gradle-cache

# View logs
sudo journalctl -u distributed-gradle-coordinator -f
sudo journalctl -u distributed-gradle-cache -f
```

### Worker Registration Issues

```bash
# Check worker connectivity
./scripts/health_checker.sh --worker worker-1

# Verify configuration
./go/client/example/main.go --list-workers
```

## Next Steps

After successful installation:

1. [Read the User Guide](/docs/user-guide) to learn how to use the system
2. [Watch Video Tutorials](/video-courses/beginner) for visual walkthroughs
3. [Configure your first project](/docs/user-guide#project-setup)
4. [Run your first distributed build](/docs/user-guide#first-distributed-build)

---

**Need help?** Check our [troubleshooting guide](/docs/troubleshooting) or [open an issue on GitHub](https://github.com/milos85vasic/Distributed-Gradle-Building/issues).