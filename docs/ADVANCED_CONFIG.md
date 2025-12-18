# Advanced Configuration Guide

This guide covers advanced configuration options for optimizing the Distributed Gradle Building System for specific use cases.

## Performance Tuning

### CPU Optimization

#### Adaptive Worker Counting

Create a smart worker calculation script:
```bash
#!/bin/bash
# calculate_workers.sh

# Get system information
CORES=$(nproc --all)
MEM_GB=$(free -g | awk '/^Mem:/{print $7}')
LOAD_AVG=$(uptime | awk -F'[a-z]:' '{print $2}' | awk '{print $1}' | tr -d ',')

# Calculate optimal workers
CPU_WORKERS=$((CORES * 2))
MEM_WORKERS=$((MEM_GB / 2))  # 2GB per worker
LOAD_WORKERS=$(echo "$LOAD_AVG < 0.7" | bc)

# Take minimum of constraints
MAX_WORKERS=$CPU_WORKERS
[ $MEM_WORKERS -lt $MAX_WORKERS ] && MAX_WORKERS=$MEM_WORKERS
[ $LOAD_WORKERS -eq 0 ] && MAX_WORKERS=$((MAX_WORKERS / 2))

echo $MAX_WORKERS
```

Integrate into build system:
```bash
# In sync_and_build.sh
source $(dirname $0)/calculate_workers.sh
export MAX_WORKERS=$(./calculate_workers.sh)
```

#### CPU Pinning for Predictable Performance

```bash
# Add to sync_and_build.sh
export GRADLE_OPTS="$GRADLE_OPTS -XX:+UseG1GC"
export GRADLE_OPTS="$GRADLE_OPTS -XX:ParallelGCThreads=$MAX_WORKERS"
export GRADLE_OPTS="$GRADLE_OPTS -XX:ConcGCThreads=2"
```

### Memory Optimization

#### Dynamic Memory Allocation

```bash
#!/bin/bash
# memory_config.sh

# System memory info
TOTAL_MEM=$(free -g | awk '/^Mem:/{print $2}')
AVAIL_MEM=$(free -g | awk '/^Mem:/{print $7}')

# Calculate optimal Gradle heap
HEAP_SIZE=$((AVAIL_MEM * 70 / 100))  # 70% of available
METASPACE_SIZE=512

# Ensure minimum values
[ $HEAP_SIZE -lt 2 ] && HEAP_SIZE=2
[ $HEAP_SIZE -gt 16 ] && HEAP_SIZE=16

echo "org.gradle.jvmargs=-Xmx${HEAP_SIZE}g -XX:MaxMetaspaceSize=${METASPACE_SIZE}m"
```

#### Per-Project Memory Profiles

Create multiple gradle.properties files:
```bash
# gradle.properties.large (for large projects)
org.gradle.jvmargs=-Xmx8g -XX:MaxMetaspaceSize=1g -XX:+UseG1GC
org.gradle.workers.max=8

# gradle.properties.small (for small projects)
org.gradle.jvmargs=-Xmx2g -XX:MaxMetaspaceSize=512m -XX:+UseG1GC
org.gradle.workers.max=4

# gradle.properties.memory-constrained
org.gradle.jvmargs=-Xmx1g -XX:MaxMetaspaceSize=256m -XX:+UseSerialGC
org.gradle.workers.max=2
```

### I/O Optimization

#### SSD Optimization

```bash
# Add to sync_and_build.sh for SSD-backed systems
export GRADLE_OPTS="$GRADLE_OPTS -XX:+UseNUMA"
export GRADLE_OPTS="$GRADLE_OPTS -Djava.io.tmpdir=/tmp/gradle_tmp"

# Create tmp directory on SSD
mkdir -p /tmp/gradle_tmp
chmod 777 /tmp/gradle_tmp
```

#### Parallel File Operations

```bash
# Enhanced rsync with parallel transfer
rsync_parallel() {
    local src=$1
    local dst=$2
    local worker_count=${3:-4}
    
    find "$src" -type f -print0 | parallel -0 -j $worker_count \
        "rsync -az --relative {} $dst"
}

# Use in sync_and_build.sh
for ip in $WORKER_IPS; do
    echo "Parallel syncing to $ip..."
    rsync_parallel "$PROJECT_DIR" "$ip:$PROJECT_DIR" 8
done
```

## Network Optimization

### SSH Configuration

#### Connection Pooling

Create `/etc/ssh/ssh_config.d/gradle_build.conf`:
```bash
Host *
    ServerAliveInterval 60
    ServerAliveCountMax 3
    Compression no
    ControlMaster auto
    ControlPath ~/.ssh/cm-%r@%h:%p
    ControlPersist 10m
    
Host gradle-*
    User build-user
    ConnectTimeout 30
    StrictHostKeyChecking no
```

#### Network Bonding for Redundancy

For systems with multiple network interfaces:
```bash
# Create network bonds
sudo nmcli connection add type bond bond0 mode=balance-rr
sudo nmcli connection add type bond-slave ifname eth0 master bond0
sudo nmcli connection add type bond-slave ifname eth1 master bond0
```

### Advanced Rsync Configuration

#### Block-Based Syncing for Large Files

```bash
rsync_advanced() {
    local src=$1
    local dst=$2
    local ip=$3
    
    rsync -a --delete --partial --partial-dir=.rsync-partial \
        --inplace --numeric-ids --acls --xattrs \
        --exclude='build/' --exclude='.gradle/' --exclude='*.iml' --exclude='.idea/' \
        --exclude='*.class' --exclude='*.jar' --exclude='*.aar' \
        "$src" "$dst"
}
```

#### Checksum-Based Verification

```bash
verify_sync() {
    local src=$1
    local dst=$2
    
    # Generate checksums
    find "$src" -type f -exec md5sum {} \; | sort > /tmp/src_checksums
    
    # Check on remote
    ssh $ip "cd $dst && find . -type f -exec md5sum {} \; | sort" > /tmp/dst_checksums
    
    # Compare
    diff /tmp/src_checksums /tmp/dst_checksums
}
```

## Build Cache Strategies

### Remote Cache Configuration

#### Nginx-based Cache Server

```nginx
# /etc/nginx/sites-available/gradle-cache
server {
    listen 8080;
    server_name _;
    
    root /var/lib/gradle-cache;
    
    location / {
        # PUT for cache writes
        dav_methods PUT DELETE MKCOL COPY MOVE;
        create_full_put_path on;
        dav_access group:rw all:r;
        
        # Cache headers
        add_header Cache-Control "public, max-age=31536000";
        
        # Authentication
        auth_basic "Gradle Cache";
        auth_basic_user_file /etc/nginx/.htpasswd;
        
        # Security
        client_max_body_size 100M;
        limit_except GET HEAD PUT DELETE {
            deny all;
        }
    }
}
```

#### Cache Warming Strategy

```bash
#!/bin/bash
# warm_cache.sh

PROJECT_DIR=$1
CACHE_SERVER=$2

# Identify recent changes
RECENT_FILES=$(find "$PROJECT_DIR" -name "*.java" -o -name "*.kt" -o -name "*.xml" \
    -newer "$PROJECT_DIR/.git/COMMIT_EDITMSG" 2>/dev/null | head -20)

# Build only changed modules
for file in $RECENT_FILES; do
    MODULE=$(echo "$file" | awk -F'/' '{print $2}')
    echo "Warming cache for module: $MODULE"
    ./gradlew ":$MODULE:compileJava" --build-cache
done

# Sync cache to server
rsync -av ~/.gradle/caches/build-cache-1/ "$CACHE_SERVER:/var/lib/gradle-cache/"
```

### Multi-Level Caching

#### Local Cache + Remote Cache

```properties
# gradle.properties
org.gradle.caching=true

# Local cache configuration
org.gradle.cache.fixed-size=true
org.gradle.cache.max-size=10g

# Remote cache configuration
org.gradle.buildCache.remote.enabled=true
org.gradle.buildCache.remote.url=http://cache-server:8080/
org.gradle.buildCache.remote.push=true
org.gradle.buildCache.remote.allow-insecure=true
```

#### Cache Hierarchies

```bash
# Cache hierarchy script
#!/bin/bash
# cache_hierarchy.sh

# Level 1: In-memory (RAM disk)
if [ ! -d /tmp/gradle-ram ]; then
    sudo mount -t tmpfs -o size=2g tmpfs /tmp/gradle-ram
fi

# Level 2: SSD cache
export GRADLE_USER_HOME="/var/gradle-ssd"

# Level 3: Remote cache
./gradlew "$@" --build-cache --refresh-keys
```

## Security Configuration

### SSH Security Hardening

#### Key-Based Authentication Only

```bash
#!/bin/bash
# secure_ssh.sh

# Disable password authentication
sudo sed -i 's/#PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config
sudo sed -i 's/PasswordAuthentication yes/PasswordAuthentication no/' /etc/ssh/sshd_config

# Use specific key types
echo "PubkeyAcceptedKeyTypes ssh-ed25519,ssh-rsa" | sudo tee -a /etc/ssh/sshd_config

# Limit access to build users
echo "AllowGroups build-users" | sudo tee -a /etc/ssh/sshd_config

sudo systemctl restart sshd
```

#### SSH Certificate Authority

```bash
# Setup SSH CA for temporary build access
ssh-keygen -s /path/to/ca_key -I build-agent -n build-user -V +1h \
    /path/to/agent_key.pub

# On workers, trust the CA
echo "@cert-authority * $(cat /path/to/ca_key.pub)" >> ~/.ssh/known_hosts
```

### Network Security

#### VPN Tunnel Configuration

```bash
#!/bin/bash
# setup_vpn.sh

# Create VPN interface
sudo ip tunnel add tunl1 mode gre local $LOCAL_IP remote $REMOTE_IP ttl 255
sudo ip addr add 10.0.0.1/30 dev tunl1
sudo ip link set tunl1 up

# Configure firewall
sudo iptables -A INPUT -i tunl1 -j ACCEPT
sudo iptables -A OUTPUT -o tunl1 -j ACCEPT
```

#### WireGuard Setup

```bash
# /etc/wireguard/wg0.conf
[Interface]
PrivateKey = <private_key>
Address = 10.0.0.1/24
ListenPort = 51820

[Peer]
PublicKey = <worker_public_key>
AllowedIPs = 10.0.0.2/32

[Peer]
PublicKey = <worker2_public_key>
AllowedIPs = 10.0.0.3/32
```

## Monitoring and Observability

### Build Performance Monitoring

#### Prometheus Metrics

```python
#!/usr/bin/env python3
# gradle_exporter.py

from prometheus_client import start_http_server, Gauge, Histogram
import subprocess
import time
import re

# Metrics
BUILD_DURATION = Histogram('gradle_build_duration_seconds', 'Build duration')
WORKER_COUNT = Gauge('gradle_active_workers', 'Active worker count')
CACHE_HIT_RATE = Gauge('gradle_cache_hit_rate', 'Cache hit rate')

def collect_metrics():
    while True:
        # Parse Gradle output
        result = subprocess.run(['./gradlew', 'build', '--info'], 
                              capture_output=True, text=True)
        
        # Extract metrics
        duration_match = re.search(r'Build completed in ([\d.]+) s', result.stdout)
        if duration_match:
            BUILD_DURATION.observe(float(duration_match.group(1)))
        
        # Export metrics
        yield
        time.sleep(30)

if __name__ == '__main__':
    start_http_server(8081)
    for _ in collect_metrics():
        pass
```

#### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Gradle Build Performance",
    "panels": [
      {
        "title": "Build Duration",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(gradle_build_duration_seconds_sum[5m])",
            "legendFormat": "Build Time"
          }
        ]
      },
      {
        "title": "Worker Utilization",
        "type": "stat",
        "targets": [
          {
            "expr": "gradle_active_workers",
            "legendFormat": "Active Workers"
          }
        ]
      }
    ]
  }
}
```

### Health Checks

#### System Health Monitoring

```bash
#!/bin/bash
# health_check.sh

check_worker_health() {
    local worker_ip=$1
    
    # SSH connectivity
    if ! ssh -o ConnectTimeout=5 -o BatchMode=yes $worker_ip echo "OK" >/dev/null; then
        echo "FAIL: SSH to $worker_ip"
        return 1
    fi
    
    # Disk space
    local disk_usage=$(ssh $worker_ip "df $PROJECT_DIR | awk 'NR==2{print $5}' | tr -d '%")
    if [ $disk_usage -gt 90 ]; then
        echo "WARN: Low disk space on $worker_ip: ${disk_usage}%"
    fi
    
    # Memory usage
    local mem_usage=$(ssh $worker_ip "free | awk 'NR==2{printf \"%.0f\", $3/$2*100}'")
    if [ $mem_usage -gt 90 ]; then
        echo "WARN: High memory usage on $worker_ip: ${mem_usage}%"
    fi
    
    echo "OK: $worker_ip healthy"
    return 0
}

# Check all workers
for ip in $WORKER_IPS; do
    check_worker_health $ip
done
```

## Integration with External Systems

### CI/CD Integration

#### Jenkins Pipeline

```groovy
pipeline {
    agent any
    
    environment {
        GRADLE_BUILD_SCRIPT = '/opt/distributed-gradle/sync_and_build.sh'
        WORKER_IPS = 'worker1 worker2 worker3'
        PROJECT_DIR = "${WORKSPACE}"
    }
    
    stages {
        stage('Setup') {
            steps {
                sh './setup_master.sh ${PROJECT_DIR}'
            }
        }
        
        stage('Build') {
            steps {
                sh '${GRADLE_BUILD_SCRIPT} assemble test'
            }
        }
        
        stage('Deploy') {
            steps {
                // Deployment steps
            }
        }
    }
    
    post {
        always {
            sh './gradlew --stop'
        }
    }
}
```

#### GitLab CI

```yaml
# .gitlab-ci.yml
variables:
  GRADLE_OPTS: "-Dorg.gradle.daemon=false"
  GRADLE_USER_HOME: "$CI_PROJECT_DIR/.gradle"

build:
  stage: build
  before_script:
    - ./setup_master.sh $CI_PROJECT_DIR
  script:
    - ./sync_and_build.sh assemble test
  artifacts:
    reports:
      junit: "**/build/test-results/test/TEST-*.xml"
```

### Container Integration

#### Docker Setup

```dockerfile
# Dockerfile.build-node
FROM ubuntu:22.04

# Install dependencies
RUN apt-get update && apt-get install -y \
    openjdk-17-jdk \
    rsync \
    openssh-client \
    && rm -rf /var/lib/apt/lists/*

# Create build user
RUN useradd -m -s /bin/bash builder

# Setup SSH
COPY ssh/ /home/builder/.ssh/
RUN chown -R builder:builder /home/builder/.ssh
RUN chmod 700 /home/builder/.ssh
RUN chmod 600 /home/builder/.ssh/*

USER builder

# Mount point for projects
VOLUME ["/workspace"]

WORKDIR /workspace
```

#### Kubernetes Deployment

```yaml
# k8s/gradle-master.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gradle-master
spec:
  replicas: 1
  selector:
    matchLabels:
      app: gradle-master
  template:
    metadata:
      labels:
        app: gradle-master
    spec:
      containers:
      - name: gradle-master
        image: gradle-build:latest
        command: ["./sync_and_build.sh"]
        env:
        - name: WORKER_IPS
          value: "gradle-worker-1.gradle-worker-2"
        - name: MAX_WORKERS
          value: "8"
        volumeMounts:
        - name: project-source
          mountPath: /workspace
      volumes:
      - name: project-source
        persistentVolumeClaim:
          claimName: gradle-project-pvc
```

## Custom Build Scenarios

### Multi-Environment Builds

#### Environment-Specific Configurations

```bash
#!/bin/bash
# multi_env_build.sh

ENVIRONMENT=${1:-dev}
PROJECT_DIR=${2:-$(pwd)}

# Load environment-specific config
case $ENVIRONMENT in
    "dev")
        source .gradlebuild_env.dev
        BUILD_TASK="assembleDebug"
        ;;
    "staging")
        source .gradlebuild_env.staging
        BUILD_TASK="assembleStaging"
        ;;
    "prod")
        source .gradlebuild_env.prod
        BUILD_TASK="assembleRelease"
        ;;
esac

# Execute build
cd "$PROJECT_DIR"
/path/to/sync_and_build.sh "$BUILD_TASK"
```

### Feature Branch Builds

#### Branch-Aware Worker Allocation

```bash
#!/bin/bash
# branch_aware_build.sh

CURRENT_BRANCH=$(git branch --show-current)
PROJECT_DIR=$(pwd)

# Allocate workers based on branch type
case $CURRENT_BRANCH in
    "main"|"develop")
        WORKER_COUNT=16
        ;;
    "feature/*"|"bugfix/*")
        WORKER_COUNT=8
        ;;
    "hotfix/*")
        WORKER_COUNT=4
        ;;
    *)
        WORKER_COUNT=2
        ;;
esac

# Update config
sed -i "s/export MAX_WORKERS=.*/export MAX_WORKERS=$WORKER_COUNT/" .gradlebuild_env

# Build
./sync_and_build.sh assemble
```

## Backup and Recovery

### Configuration Backup

```bash
#!/bin/bash
# backup_config.sh

BACKUP_DIR="/backup/gradle-config/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

# Backup configuration
cp .gradlebuild_env "$BACKUP_DIR/"
cp gradle.properties "$BACKUP_DIR/"
cp ~/.ssh/known_hosts "$BACKUP_DIR/"
cp -r ~/.ssh/keys "$BACKUP_DIR/"

# Backup worker information
for ip in $WORKER_IPS; do
    mkdir -p "$BACKUP_DIR/workers/$ip"
    ssh $ip "cat ~/.ssh/known_hosts" > "$BACKUP_DIR/workers/$ip/known_hosts"
    ssh $ip "df -h" > "$BACKUP_DIR/workers/$ip/disk_info"
done

# Create restore script
cat > "$BACKUP_DIR/restore.sh" << 'EOF'
#!/bin/bash
# Restore configuration from backup

BACKUP_DIR=$(dirname $0)

# Restore main config
cp "$BACKUP_DIR/.gradlebuild_env" .
cp "$BACKUP_DIR/gradle.properties" .
cp "$BACKUP_DIR/known_hosts" ~/.ssh/

echo "Configuration restored. Test with: ./sync_and_build.sh test"
EOF

chmod +x "$BACKUP_DIR/restore.sh"
echo "Backup saved to: $BACKUP_DIR"
```

### Disaster Recovery

```bash
#!/bin/bash
# disaster_recovery.sh

# Identify failed master
FAILED_MASTER=$1
BACKUP_MASTER=$2

if [ -z "$FAILED_MASTER" ] || [ -z "$BACKUP_MASTER" ]; then
    echo "Usage: $0 <failed_master_ip> <backup_master_ip>"
    exit 1
fi

# Promote backup to master
ssh $BACKUP_MASTER "
    # Update configuration
    sed -i 's/export MASTER_IP=.*/export MASTER_IP=$BACKUP_MASTER/' .gradlebuild_env
    
    # Start cache server if needed
    if ! pgrep -f 'python3.*http.server' > /dev/null; then
        cd $PROJECT_DIR
        mkdir -p build-cache
        nohup python3 -m http.server 8080 --directory build-cache &
    fi
"

# Update workers to point to new master
for ip in $WORKER_IPS; do
    if [ "$ip" != "$BACKUP_MASTER" ]; then
        ssh $ip "
            sed -i 's/$FAILED_MASTER/$BACKUP_MASTER/g' .gradlebuild_env
        "
    fi
done

echo "Master promotion completed. New master: $BACKUP_MASTER"
```

## Testing and Validation

### Performance Benchmarking

```bash
#!/bin/bash
# benchmark_builds.sh

PROJECT_DIR=$1
ITERATIONS=${2:-5}

mkdir -p benchmark_results
cd "$PROJECT_DIR"

# Test configurations
CONFIGS=("local" "distributed" "cached")
WORKERS=(1 4 8 16)

for config in "${CONFIGS[@]}"; do
    echo "Benchmarking: $config"
    
    for workers in "${WORKERS[@]}"; do
        echo "Testing with $workers workers"
        
        for i in $(seq 1 $ITERATIONS); do
            START=$(date +%s)
            
            case $config in
                "local")
                    ./gradlew clean assemble --no-daemon
                    ;;
                "distributed")
                    export MAX_WORKERS=$workers
                    ./sync_and_build.sh clean assemble
                    ;;
                "cached")
                    export MAX_WORKERS=$workers
                    ./sync_and_build.sh clean assemble --build-cache
                    ;;
            esac
            
            END=$(date +%s)
            DURATION=$((END - START))
            
            echo "$config,$workers,$i,$DURATION" >> benchmark_results/results.csv
        done
    done
done

echo "Benchmark complete. Results in benchmark_results/results.csv"
```

This advanced configuration guide provides comprehensive options for optimizing the distributed Gradle building system for specific environments and use cases.