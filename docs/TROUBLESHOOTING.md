# Troubleshooting Guide

This guide helps diagnose and resolve common issues with the Distributed Gradle Building System.

## Quick Diagnosis Checklist

When something goes wrong, check these in order:
1. [ ] SSH connectivity between machines
2. [ ] Java installation and versions
3. [ ] Project directory existence and permissions
4. [ ] Gradle wrapper availability
5. [ ] Network connectivity and latency
6. [ ] Disk space availability

## SSH Issues

### Passwordless SSH Not Working

**Symptoms:**
- Setup script asks for password
- "Permission denied" errors
- "Host key verification failed"

**Diagnostics:**
```bash
# Test SSH connectivity
ssh -o BatchMode=yes -o ConnectTimeout=5 user@worker_ip echo "OK"

# Check SSH key permissions
ls -la ~/.ssh/
```

**Solutions:**

1. Fix key permissions:
```bash
chmod 700 ~/.ssh
chmod 600 ~/.ssh/id_ed25519
chmod 644 ~/.ssh/id_ed25519.pub
chmod 600 ~/.ssh/authorized_keys
```

2. Regenerate keys if corrupted:
```bash
rm ~/.ssh/id_ed25519*
ssh-keygen -t ed25519 -N "" -f ~/.ssh/id_ed25519
ssh-copy-id user@worker_ip
```

3. Accept host keys manually:
```bash
ssh user@worker_ip exit
```

### SSH Timeout Issues

**Symptoms:**
- "Connection timed out" errors
- Slow or intermittent SSH connections

**Diagnostics:**
```bash
# Test with verbose output
ssh -v user@worker_ip echo "test"

# Check network latency
ping worker_ip

# Test port connectivity
telnet worker_ip 22
```

**Solutions:**

1. Increase SSH timeout:
```bash
# Add to ~/.ssh/config
Host *
    ConnectTimeout 30
    ServerAliveInterval 60
    ServerAliveCountMax 3
```

2. Check firewall settings:
```bash
# On Ubuntu/Debian
sudo ufw status
sudo ufw allow ssh

# On CentOS/RHEL
sudo firewall-cmd --list-all
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --reload
```

3. Check SSH daemon configuration:
```bash
# On worker machines
sudo nano /etc/ssh/sshd_config
# Ensure:
# ClientAliveInterval 300
# ClientAliveCountMax 2
sudo systemctl restart sshd
```

## Gradle Issues

### Gradle Wrapper Not Found

**Symptoms:**
- "gradlew: command not found"
- "Permission denied" on gradlew

**Diagnostics:**
```bash
ls -la gradlew
file gradlew
```

**Solutions:**

1. Make gradlew executable:
```bash
chmod +x gradlew
```

2. Generate wrapper if missing:
```bash
gradle wrapper --gradle-version 8.0
# Or download from your project's source
```

3. Check line endings (Windows ↔ Linux):
```bash
# Convert Windows line endings to Unix
dos2unix gradlew
# Or manually:
sed -i 's/\r$//' gradlew
```

### Gradle Build Failures

**Symptoms:**
- Compilation errors
- OutOfMemoryError
- Dependency resolution failures

**Diagnostics:**
```bash
# Run with verbose output
./gradlew build --info --stacktrace

# Check Java version
java -version
javac -version

# Check available memory
free -h
```

**Solutions:**

1. Increase heap size:
```bash
# Environment variable
export GRADLE_OPTS="-Xmx4g -XX:MaxMetaspaceSize=512m"

# In gradle.properties
org.gradle.jvmargs=-Xmx4g -XX:MaxMetaspaceSize=512m
```

2. Clear Gradle cache:
```bash
./gradlew clean
rm -rf ~/.gradle/caches
./gradlew build --refresh-keys
```

3. Check Java version consistency:
```bash
# On all machines
java -version
# Update if inconsistent
sudo apt update && sudo apt install openjdk-17-jdk
```

## Rsync Issues

### Rsync Permission Errors

**Symptoms:**
- "Permission denied" during sync
- "Operation not permitted" errors

**Diagnostics:**
```bash
# Check directory permissions
ssh worker_ip "ls -la $PROJECT_DIR"

# Test rsync manually
rsync -av --dry-run -e ssh test.txt user@worker_ip:/tmp/
```

**Solutions:**

1. Fix directory ownership:
```bash
# On worker machines
sudo chown -R $USER:$USER $PROJECT_DIR
chmod -R u+w $PROJECT_DIR
```

2. Use different sync user:
```bash
# Update sync_and_build.sh
rsync -a --delete -e "ssh -u other_user" ...
```

3. Create shared directory with proper permissions:
```bash
# On all machines
sudo mkdir -p /shared/gradle-projects
sudo chown $USER:$USER /shared/gradle-projects
sudo chmod 755 /shared/gradle-projects
```

### Rsync Slow Performance

**Symptoms:**
- Sync taking extremely long time
- High network usage during sync

**Diagnostics:**
```bash
# Test network speed
iperf3 -c worker_ip

# Check rsync with stats
rsync -av --stats -e ssh src/ user@worker_ip:/tmp/
```

**Solutions:**

1. Add more exclusions to sync_and_build.sh:
```bash
rsync -a --delete -e ssh \
  --exclude='build/' --exclude='.gradle/' --exclude='*.iml' --exclude='.idea/' \
  --exclude='node_modules/' --exclude='npm_cache/' --exclude='.git/' \
  "$PROJECT_DIR"/ "$ip":"$PROJECT_DIR"/
```

2. Use compression for slow networks:
```bash
rsync -az --delete -e ssh ...
```

3. Limit bandwidth usage:
```bash
rsync -av --delete --bwlimit=10m -e ssh ...
```

## Performance Issues

### Builds Slower Than Expected

**Symptoms:**
- Distributed builds slower than local builds
- High CPU usage but slow progress

**Diagnostics:**
```bash
# Check CPU usage on all machines
for ip in $WORKER_IPS; do
  echo "=== $ip ==="
  ssh $ip "top -bn1 | head -10"
done

# Check network latency
for ip in $WORKER_IPS; do
  echo "=== $ip ==="
  ping -c 3 $ip
done

# Check disk I/O
iostat 1 5
```

**Solutions:**

1. Optimize worker count:
```bash
# Reduce if oversubscribed
echo "export MAX_WORKERS=4" > .gradlebuild_env

# Or calculate based on actual cores
CORES=$(nproc --all)
echo "export MAX_WORKERS=$((CORES))" > .gradlebuild_env
```

2. Enable build caching:
```bash
# Add to gradle.properties
org.gradle.caching=true
org.gradle.buildCache.remote.url=http://master:8080/cache/
```

3. Use faster storage:
```bash
# Check if on SSD vs HDD
lsblk -d -o name,rota
# ROTA=1 means HDD, ROTA=0 means SSD
```

### Memory Issues

**Symptoms:**
- OutOfMemoryError during builds
- System becoming unresponsive
- Gradle daemon crashes

**Diagnostics:**
```bash
# Check memory usage
free -h

# Check Gradle processes
ps aux | grep gradle

# Check system limits
ulimit -a
```

**Solutions:**

1. Increase heap size globally:
```bash
# In ~/.bashrc
export GRADLE_OPTS="-Xmx4g -XX:+UseG1GC"

# In gradle.properties
org.gradle.jvmargs=-Xmx4g -XX:+UseG1GC -XX:MaxGCPauseMillis=200
```

2. Limit parallelism on memory-constrained machines:
```bash
# Calculate based on available memory
MEM_GB=$(free -g | awk '/^Mem:/{print $7}')
MAX_WORKERS=$((MEM_GB / 2))  # 2GB per worker
```

3. Enable daemon recycling:
```bash
# In gradle.properties
org.gradle.daemon=true
org.gradle.daemon.idletimeout=3600000  # 1 hour
```

## Network Issues

### Worker Connectivity Problems

**Symptoms:**
- Some workers unreachable
- Intermittent connection failures
- "Network is unreachable" errors

**Diagnostics:**
```bash
# Test all workers
for ip in $WORKER_IPS; do
  echo "=== Testing $ip ==="
  ping -c 3 $ip
  ssh -o ConnectTimeout=5 $ip "echo 'OK'"
done

# Check routing
traceroute worker_ip

# Check DNS
nslookup worker_ip
dig -x worker_ip
```

**Solutions:**

1. Use IP addresses instead of hostnames
2. Add entries to `/etc/hosts`:
```bash
echo "192.168.1.11 worker1" >> /etc/hosts
echo "192.168.1.12 worker2" >> /etc/hosts
```

3. Set up SSH multiplexing:
```bash
# Add to ~/.ssh/config
Host *
    ControlMaster auto
    ControlPath ~/.ssh/master-%r@%h:%p
    ControlPersist 10m
```

### Network Bandwidth Saturation

**Symptoms:**
- Slow builds during sync phase
- Network congestion affecting other services

**Diagnostics:**
```bash
# Monitor network usage
iftop -i eth0

# Test bandwidth
iperf3 -c worker_ip -t 30
```

**Solutions:**

1. Limit rsync bandwidth:
```bash
rsync -av --delete --bwlimit=50m -e ssh ...
```

2. Schedule builds during off-peak hours
3. Use dedicated network for build traffic

## Configuration Issues

### Environment Variables Not Loading

**Symptoms:**
- "WORKER_IPS: command not found"
- Workers not found during sync

**Diagnostics:**
```bash
# Check if config file exists
ls -la .gradlebuild_env

# Test config loading
source .gradlebuild_env
echo $WORKER_IPS
```

**Solutions:**

1. Ensure config file exists:
```bash
./setup_master.sh /absolute/path/to/project
```

2. Check file permissions:
```bash
chmod 644 .gradlebuild_env
```

3. Verify paths in config are absolute:
```bash
# Should be:
export PROJECT_DIR="/home/user/project"
# Not:
export PROJECT_DIR="~/project"
```

### Project Directory Mismatch

**Symptoms:**
- rsync errors about non-existent directories
- Permission denied errors

**Diagnostics:**
```bash
# Check project exists on all machines
for ip in localhost $WORKER_IPS; do
  echo "=== $ip ==="
  ssh $ip "ls -la $PROJECT_DIR"
done
```

**Solutions:**

1. Create directories on workers:
```bash
for ip in $WORKER_IPS; do
  ssh $ip "mkdir -p $PROJECT_DIR"
done
```

2. Use same absolute paths everywhere:
```bash
# Verify path format
realpath /path/to/project  # Use this output
```

3. Update setup scripts if path format inconsistent

## Advanced Debugging

### Enable Detailed Logging

1. Add debug mode to sync_and_build.sh:
```bash
#!/bin/bash

set -x  # Enable command tracing

# Rest of script...
```

2. Add timing information:
```bash
echo "$(date): Starting sync..."
# sync commands
echo "$(date): Sync completed in $SECONDS seconds"
```

### Isolate Issues

1. Test components individually:
```bash
# Test sync only
rsync -av --dry-run -e ssh $PROJECT_DIR/ worker_ip:$PROJECT_DIR/

# Test build only
./gradlew assemble --parallel --max-workers=4

# Test SSH only
ssh worker_ip "cd $PROJECT_DIR && ./gradlew --version"
```

2. Use build scans for detailed analysis:
```bash
# In gradle.properties
plugins {
    id 'com.gradle.build-scan' version '3.4.1'
}
buildScan {
    publishAlways()
    termsOfServiceUrl = 'https://gradle.com/terms-of-service'
    termsOfServiceAgree = 'yes'
}
```

## Emergency Recovery

### Complete System Reset

If everything fails, start fresh:

1. Clean all machines:
```bash
# On all machines
rm -rf ~/.gradle/caches
rm -rf $PROJECT_DIR
rm -f ~/.gradlebuild_env
```

2. Re-run full setup from scratch

### Partial Recovery

For specific component failures:

1. Recreate SSH keys:
```bash
rm ~/.ssh/id_ed25519*
ssh-keygen -t ed25519 -N ""
./setup_passwordless_ssh.sh
```

2. Recreate Gradle wrapper:
```bash
rm gradlew gradlew.bat
gradle wrapper --gradle-version 8.0
```

3. Reset configuration:
```bash
rm .gradlebuild_env
./setup_master.sh $PROJECT_DIR
```

## Getting Help

### Collect Debug Information

When reporting issues, include:

```bash
# System information
uname -a
java -version
./gradlew --version

# Configuration
cat .gradlebuild_env

# Network test
for ip in $WORKER_IPS; do
  echo "=== $ip ==="
  ping -c 2 $ip
  ssh -o ConnectTimeout=2 $ip "echo 'OK'" || echo "FAILED"
done

# Build attempt with output
./sync_and_build.sh assemble 2>&1 | tee build.log
```

### Common Log Analysis

Look for these patterns in logs:

1. SSH issues: "Permission denied", "Connection timed out"
2. Memory issues: "OutOfMemoryError", "Could not reserve enough space"
3. Permission issues: "Operation not permitted", "Access denied"
4. Network issues: "Network is unreachable", "No route to host"