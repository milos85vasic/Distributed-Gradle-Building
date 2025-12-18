# Complete Setup Guide

This guide walks you through setting up the Distributed Gradle Building System from scratch.

## System Requirements

### Hardware Requirements
- **Master Node**: Minimum 4GB RAM, 2+ CPU cores
- **Worker Nodes**: Minimum 2GB RAM, 1+ CPU core each
- **Network**: 100Mbps+ recommended between machines
- **Storage**: 10GB+ free space for project on each machine

### Software Requirements
- **Operating System**: Ubuntu 18.04+ / Debian 9+ (other Linux variants may work)
- **Java**: OpenJDK 17 (or your project's required version)
- **Network**: SSH server running on all machines
- **Tools**: git, rsync, curl

## Network Preparation

### 1. Configure Hostnames (Optional but Recommended)

On all machines, edit `/etc/hosts`:
```bash
sudo nano /etc/hosts
```

Add entries for all machines:
```
192.168.1.10  master
192.168.1.11  worker1
192.168.1.12  worker2
```

### 2. Configure SSH

Install SSH server if not present:
```bash
sudo apt update
sudo apt install -y openssh-server
```

Ensure SSH is running:
```bash
sudo systemctl status ssh
```

## Step-by-Step Setup

### Step 1: Clone or Download the Scripts

```bash
# Option 1: Clone the repository
git clone git@github.com:milos85vasic/Distributed-Gradle-Building.git
cd Distributed-Gradle-Building

# Option 2: Download and extract
wget https://github.com/milos85vasic/Distributed-Gradle-Building/archive/main.zip
unzip main.zip
cd Distributed-Gradle-Building-main
```

### Step 2: Make Scripts Executable

```bash
chmod +x setup_*.sh sync_and_build.sh
chmod +x Upstreams/GitHub.sh
```

### Step 3: Setup Passwordless SSH

This is the most critical step. The system requires bidirectional SSH access between all machines.

#### Interactive Setup

Run the script on any machine:
```bash
./setup_passwordless_ssh.sh
```

You'll be prompted for:
- Username (must exist on all machines)
- SSH password (must be same on all machines)
- IP addresses of all machines (including the current one)

#### Manual Setup (Alternative)

If the automated script fails, set up SSH keys manually:

1. Generate SSH key on master:
```bash
ssh-keygen -t ed25519 -N "" -f ~/.ssh/id_ed25519
```

2. Copy to all workers:
```bash
for ip in WORKER_IP1 WORKER_IP2; do
    ssh-copy-id user@$ip
done
```

3. Test SSH:
```bash
ssh user@WORKER_IP1 "echo 'SSH OK'"
```

### Step 4: Install Java

On all machines:
```bash
sudo apt update
sudo apt install -y openjdk-17-jdk
```

Verify installation:
```bash
java -version
javac -version
```

### Step 5: Prepare Your Gradle Project

Ensure your project has:
- A `gradlew` wrapper script
- Valid `build.gradle` or `build.gradle.kts` files

If missing, generate wrapper:
```bash
gradle wrapper --gradle-version 8.0
```

### Step 6: Setup Workers

Run this script on EACH worker machine:
```bash
./setup_worker.sh /absolute/path/to/your/project
```

You'll be prompted for:
- Master IP address

Example path:
```bash
./setup_worker.sh /home/user/workspace/my-gradle-project
```

### Step 7: Setup Master

Run this script on the master machine:
```bash
./setup_master.sh /absolute/path/to/your/project
```

You'll be prompted for:
- Worker IP addresses (space-separated)

Example:
```
Enter worker IP addresses (space-separated): 192.168.1.11 192.168.1.12
```

### Step 8: Verify Setup

Check that configuration file was created:
```bash
cat /absolute/path/to/your/project/.gradlebuild_env
```

You should see:
```bash
export WORKER_IPS="192.168.1.11 192.168.1.12"
export MAX_WORKERS=16  # Varies based on CPU cores
export PROJECT_DIR="/absolute/path/to/your/project"
```

### Step 9: First Build Test

Navigate to your project directory on the master:
```bash
cd /absolute/path/to/your/project
```

Run the build:
```bash
/some/path/Distributed-Gradle-Building/sync_and_build.sh
```

## Advanced Setup Options

### Custom SSH Port

If you use a non-standard SSH port, modify the scripts:
1. Add `-p PORT` to SSH commands
2. Add `-e 'ssh -p PORT'` to rsync commands

### Alternative Java Versions

If your project needs a different Java version:
1. Install the required JDK on all machines
2. Update the setup scripts to install the correct version
3. Ensure `JAVA_HOME` is set appropriately

### VPN or SSH Tunnel Support

For remote workers:
1. Set up VPN or SSH tunnels
2. Use tunnel endpoints as IP addresses
3. Test connectivity before running setup scripts

## Verification Checklist

Before proceeding to regular use, verify:

- [ ] Passwordless SSH works bidirectionally between all machines
- [ ] Java is installed and accessible on all machines
- [ ] Project directory exists with same path on all machines
- [ ] `.gradlebuild_env` configuration file exists on master
- [ ] Gradle wrapper (`gradlew`) exists in project directory
- [ ] Network connectivity is stable between machines
- [ ] Disk space is sufficient on all machines

## Common Setup Issues

### SSH Issues

**Problem**: SSH asks for password despite setup
```
Solution: Check ~/.ssh/authorized_keys permissions:
chmod 600 ~/.ssh/authorized_keys
chmod 700 ~/.ssh
```

**Problem**: "Host key verification failed"
```
Solution: SSH once manually to accept host keys:
ssh user@worker_ip
```

### Java Issues

**Problem**: Multiple Java versions installed
```
Solution: Set JAVA_HOME in ~/.bashrc:
export JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64
```

### Path Issues

**Problem**: Project paths differ between machines
```
Solution: Ensure absolute paths are identical:
master:/home/user/project
worker1:/home/user/project
```

## Post-Setup Configuration

### 1. Optimize Gradle Properties

Add to your project's `gradle.properties`:
```properties
# Recommended settings for this system
org.gradle.parallel=true
org.gradle.caching=true
org.gradle.daemon=true
org.gradle.jvmargs=-Xmx4g -XX:+UseG1GC

# Adjust based on your master's CPU
org.gradle.workers.max=8  # Should match MAX_WORKERS from setup
```

### 2. Configure Git Hooks (Optional)

Create a pre-commit hook to run sync_and_build.sh:
```bash
echo "#!/bin/bash
/path/to/Distributed-Gradle-Building/sync_and_build.sh --dry-run" > .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

### 3. Setup Monitoring (Optional)

Monitor build performance with simple logging:
```bash
# Add to sync_and_build.sh
echo "$(date): Starting build" >> ~/.gradle_build.log
./gradlew "$@"
echo "$(date): Build completed" >> ~/.gradle_build.log
```

## Next Steps

Once setup is complete:
1. Read the [User Guide](USER_GUIDE.md) for day-to-day usage
2. Review [Advanced Configuration](ADVANCED_CONFIG.md) for optimization
3. Check [Troubleshooting Guide](TROUBLESHOOTING.md) for common issues