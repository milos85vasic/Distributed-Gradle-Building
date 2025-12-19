# AI-Powered Distributed Gradle Building - Troubleshooting Guide

## Overview

This guide provides solutions for common issues encountered when deploying and operating the AI-powered distributed Gradle building system.

## Quick Diagnosis

### System Health Check

Run this comprehensive health check script:

```bash
#!/bin/bash
# health-check.sh

echo "=== AI-Powered Distributed Gradle Building Health Check ==="
echo

# Check Docker services
echo "1. Docker Services Status:"
docker-compose ps
echo

# Check service health endpoints
echo "2. Service Health Checks:"
services=("coordinator:8080" "ml-service:8082" "monitor:8084" "cache:8085" "worker-1:8087" "worker-2:8088" "worker-3:8089")

for service in "${services[@]}"; do
    name=$(echo $service | cut -d: -f1)
    port=$(echo $service | cut -d: -f2)

    if curl -f -s http://localhost:$port/health > /dev/null 2>&1; then
        echo "✅ $name: HEALTHY"
    else
        echo "❌ $name: UNHEALTHY"
    fi
done
echo

# Check inter-service connectivity
echo "3. Inter-Service Connectivity:"
if curl -f -s http://localhost:8080/api/workers > /dev/null 2>&1; then
    echo "✅ Coordinator can access workers"
else
    echo "❌ Coordinator cannot access workers"
fi

if curl -f -s http://localhost:8082/api/learning > /dev/null 2>&1; then
    echo "✅ ML service operational"
else
    echo "❌ ML service not responding"
fi
echo

# Check resource usage
echo "4. Resource Usage:"
docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}"
echo

echo "=== Health Check Complete ==="
```

## Service-Specific Issues

### Coordinator Service Issues

#### Workers Not Registering

**Symptoms:**
- Workers show as "unhealthy" in `docker-compose ps`
- Build requests fail with "no workers available"
- Worker logs show connection refused errors

**Diagnosis:**
```bash
# Check worker registration
curl http://localhost:8080/api/workers

# Check worker logs
docker-compose logs worker-1

# Test RPC connectivity
docker-compose exec coordinator nc -zv worker-1 8087
```

**Solutions:**

1. **Network connectivity issues:**
   ```bash
   # Restart networking
   docker-compose down
   docker-compose up -d

   # Check Docker network
   docker network ls
   docker network inspect distributed-gradle_default
   ```

2. **Port conflicts:**
   ```bash
   # Check port usage
   netstat -tulpn | grep :808[0-9]

   # Update ports in docker-compose.yml if needed
   ```

3. **Worker configuration:**
   ```bash
   # Verify worker environment variables
   docker-compose exec worker-1 env | grep COORDINATOR

   # Check worker startup logs
   docker-compose logs --tail=50 worker-1
   ```

#### Build Queue Full

**Symptoms:**
- New builds rejected with "queue is full"
- High memory usage on coordinator
- Slow response times

**Solutions:**
```bash
# Increase queue size
export BUILD_QUEUE_SIZE=200

# Restart coordinator
docker-compose restart coordinator

# Monitor queue length
curl http://localhost:8080/api/metrics | jq '.queue_length'
```

### ML Service Issues

#### Continuous Learning Not Working

**Symptoms:**
- ML predictions remain static
- No model retraining logs
- `/api/learning` shows old timestamps

**Diagnosis:**
```bash
# Check ML service configuration
curl http://localhost:8082/api/learning | jq '.config'

# Check data collection
curl http://localhost:8082/api/stats | jq '.total_build_records'

# Verify inter-service communication
docker-compose exec ml-service curl -f coordinator:8080/api/health
docker-compose exec ml-service curl -f monitor:8084/health
```

**Solutions:**

1. **Enable continuous learning:**
   ```bash
   # Set environment variable
   export ML_CONTINUOUS_LEARNING_ENABLED=true

   # Restart ML service
   docker-compose restart ml-service
   ```

2. **Insufficient data:**
   ```bash
   # Check data collection
   curl http://localhost:8082/api/stats

   # Manually trigger training
   curl -X POST http://localhost:8082/api/train
   ```

3. **Configuration issues:**
   ```bash
   # Verify ML configuration
   docker-compose exec ml-service env | grep ML_

   # Check service connectivity
   docker-compose exec ml-service ping coordinator
   ```

#### Model Accuracy Degradation

**Symptoms:**
- Build time predictions increasingly inaccurate
- Higher than expected failure rates
- ML service logs show rollback messages

**Solutions:**
```bash
# Check current model performance
curl http://localhost:8082/api/learning | jq '.stats.average_accuracy'

# View model versions
curl http://localhost:8082/api/learning | jq '.stats.model_backups'

# Rollback to previous version
curl -X POST http://localhost:8082/api/rollback \
  -H "Content-Type: application/json" \
  -d '{"version": "v4.1703932800"}'

# Force retraining
curl -X POST http://localhost:8082/api/train
```

### Monitor Service Issues

#### Metrics Not Collecting

**Symptoms:**
- Empty metrics in `/api/metrics`
- Dashboard shows no data
- Alert system not triggering

**Diagnosis:**
```bash
# Check monitor configuration
curl http://localhost:8084/api/metrics

# Verify service connectivity
docker-compose exec monitor curl -f coordinator:8080/api/health
docker-compose exec monitor curl -f ml-service:8082/health
```

**Solutions:**

1. **Service discovery issues:**
   ```bash
   # Update monitor configuration
   export MONITOR_COORDINATOR_HOST=coordinator
   export MONITOR_ML_SERVICE_HOST=ml-service

   docker-compose restart monitor
   ```

2. **Metrics collection interval:**
   ```bash
   # Adjust collection frequency
   export METRICS_INTERVAL=15s

   docker-compose restart monitor
   ```

### Cache Service Issues

#### Low Cache Hit Rate

**Symptoms:**
- Cache hit rate below 50%
- Slow build times
- High cache miss rates

**Diagnosis:**
```bash
# Check cache statistics
curl http://localhost:8085/stats

# Monitor cache performance
docker-compose logs --tail=100 cache
```

**Solutions:**

1. **Increase cache size:**
   ```bash
   export MAX_CACHE_SIZE=21474836480  # 20GB

   docker-compose restart cache
   ```

2. **Adjust TTL settings:**
   ```bash
   export TTL_SECONDS=604800  # 7 days

   docker-compose restart cache
   ```

3. **Enable compression:**
   ```bash
   export COMPRESSION=true

   docker-compose restart cache
   ```

#### Cache Storage Full

**Symptoms:**
- Builds failing with storage errors
- Cache evictions increasing rapidly
- High disk I/O

**Solutions:**
```bash
# Check disk usage
df -h /var/lib/docker/volumes/

# Clean old cache entries
docker-compose exec cache find /app/data/cache -type f -mtime +30 -delete

# Increase storage allocation
# Edit docker-compose.yml to increase volume size
```

### Worker Service Issues

#### Build Failures

**Symptoms:**
- Builds failing on specific workers
- Gradle errors in build logs
- Worker status shows "error"

**Diagnosis:**
```bash
# Check worker logs
docker-compose logs --tail=100 worker-1

# Verify worker resources
docker-compose exec worker-1 df -h
docker-compose exec worker-1 free -h

# Check Gradle installation
docker-compose exec worker-1 ./gradlew --version
```

**Solutions:**

1. **Resource constraints:**
   ```bash
   # Increase worker memory
   export JAVA_OPTS="-Xmx8g -Xms4g"

   docker-compose restart worker-1
   ```

2. **Gradle cache issues:**
   ```bash
   # Clear Gradle cache
   docker-compose exec worker-1 rm -rf /app/gradle-home/caches/*

   # Restart worker
   docker-compose restart worker-1
   ```

3. **Network issues:**
   ```bash
   # Check connectivity to cache
   docker-compose exec worker-1 curl -f cache:8083/health
   ```

#### Worker Performance Issues

**Symptoms:**
- Slow build times on specific workers
- High CPU/memory usage
- Worker frequently disconnecting

**Solutions:**
```bash
# Adjust concurrent builds
export MAX_BUILDS=3

# Monitor worker metrics
curl http://localhost:8084/api/metrics | jq '.workers."worker-1"'

# Check for resource leaks
docker-compose exec worker-1 ps aux
docker-compose exec worker-1 top -b -n1
```

## Network and Connectivity Issues

### Service Discovery Problems

**Symptoms:**
- Services cannot communicate with each other
- Intermittent connection failures
- DNS resolution errors

**Solutions:**

1. **Docker network issues:**
   ```bash
   # Recreate network
   docker-compose down
   docker network rm distributed-gradle_default
   docker-compose up -d
   ```

2. **DNS resolution:**
   ```bash
   # Check DNS
   docker-compose exec coordinator nslookup ml-service

   # Restart DNS
   sudo systemctl restart docker
   ```

### Port Conflicts

**Symptoms:**
- Services fail to start with "port already in use"
- Connection refused errors

**Diagnosis:**
```bash
# Check port usage
netstat -tulpn | grep :808[0-9]

# Check Docker port mappings
docker ps --format "table {{.Names}}\t{{.Ports}}"
```

**Solutions:**
```yaml
# Update docker-compose.yml with different ports
services:
  coordinator:
    ports:
      - "8081:8080"  # Changed from 8080
  ml-service:
    ports:
      - "8083:8082"  # Changed from 8082
```

## Performance Issues

### High Memory Usage

**Symptoms:**
- Services consuming excessive memory
- Out of memory errors
- System slowdown

**Solutions:**

1. **ML service memory optimization:**
   ```bash
   # Limit ML service memory
   export GOMEMLIMIT=4GiB

   docker-compose restart ml-service
   ```

2. **Cache memory tuning:**
   ```bash
   # Reduce cache size
   export MAX_CACHE_SIZE=5368709120  # 5GB

   docker-compose restart cache
   ```

3. **Worker memory limits:**
   ```yaml
   # Add memory limits to docker-compose.yml
   services:
     worker-1:
       deploy:
         resources:
           limits:
             memory: 8G
           reservations:
             memory: 4G
   ```

### High CPU Usage

**Symptoms:**
- Services consuming excessive CPU
- Slow response times
- System becoming unresponsive

**Solutions:**

1. **ML training scheduling:**
   ```bash
   # Reduce training frequency
   export ML_RETRAINING_INTERVAL=168h  # Weekly instead of daily

   docker-compose restart ml-service
   ```

2. **Worker CPU limits:**
   ```yaml
   services:
     worker-1:
       deploy:
         resources:
           limits:
             cpus: '4.0'
   ```

### Slow Build Times

**Symptoms:**
- Builds taking longer than expected
- Cache hit rates declining
- Worker utilization low

**Diagnosis:**
```bash
# Check ML predictions
curl -X POST http://localhost:8082/api/predict \
  -H "Content-Type: application/json" \
  -d '{"project_path":"/projects/myapp","task_name":"build"}' \
  | jq '.predicted_time'

# Monitor cache performance
curl http://localhost:8085/stats | jq '.hit_rate'

# Check worker utilization
curl http://localhost:8084/api/metrics | jq '.workers'
```

**Solutions:**

1. **Improve ML predictions:**
   ```bash
   # Force model retraining
   curl -X POST http://localhost:8082/api/train

   # Increase data collection
   export ML_DATA_COLLECTION_INTERVAL=2m
   ```

2. **Cache optimization:**
   ```bash
   # Increase cache TTL
   export TTL_SECONDS=2592000  # 30 days

   # Enable compression
   export COMPRESSION=true
   ```

3. **Worker scaling:**
   ```bash
   # Add more workers
   docker-compose up -d --scale worker-1=2 --scale worker-2=2 --scale worker-3=2
   ```

## Data and Persistence Issues

### Data Loss

**Symptoms:**
- ML models reset to defaults
- Build history lost
- Cache data missing

**Recovery:**
```bash
# Check volume mounts
docker volume ls

# Restore from backup
docker run --rm -v distributed-gradle_ml_data:/data \
  alpine tar xzf /backup/ml-data-20231231.tar.gz -C /

# Restart services
docker-compose restart
```

### Database Corruption

**Symptoms:**
- Services failing to start
- Data inconsistency errors
- ML predictions failing

**Solutions:**
```bash
# Export current data
curl http://localhost:8082/api/export > backup-$(date +%Y%m%d).json

# Reset and reimport
docker-compose exec ml-service rm -rf /app/data/*
curl -X POST http://localhost:8082/api/import \
  -H "Content-Type: application/json" \
  --data-binary @backup-$(date +%Y%m%d).json

# Restart service
docker-compose restart ml-service
```

## Monitoring and Alerting Issues

### Alerts Not Triggering

**Symptoms:**
- No alerts despite high resource usage
- Alert thresholds not working

**Solutions:**
```bash
# Check alert configuration
docker-compose exec monitor env | grep ALERT

# Update thresholds
export ALERT_CPU_THRESHOLD=75.0
export ALERT_MEMORY_THRESHOLD=80.0

docker-compose restart monitor
```

### Metrics Not Visible

**Symptoms:**
- Grafana dashboards empty
- Prometheus metrics missing

**Solutions:**
```bash
# Check Prometheus configuration
curl http://localhost:9090/targets

# Verify service labels
docker-compose ps --format "table {{.Name}}\t{{.Labels}}"

# Restart monitoring stack
docker-compose restart prometheus grafana
```

## Upgrade and Migration Issues

### Service Upgrade Problems

**Symptoms:**
- Services failing after upgrade
- Incompatible API versions
- Data migration issues

**Solutions:**

1. **Rolling upgrade:**
   ```bash
   # Update image version
   export IMAGE_TAG=v1.1.0

   # Rolling restart
   docker-compose up -d coordinator
   sleep 30
   docker-compose up -d ml-service
   sleep 30
   docker-compose up -d monitor cache
   sleep 30
   docker-compose up -d worker-1 worker-2 worker-3
   ```

2. **Data migration:**
   ```bash
   # Backup before upgrade
   curl http://localhost:8082/api/export > pre-upgrade-backup.json

   # Upgrade and verify
   docker-compose pull
   docker-compose up -d

   # Restore if needed
   curl -X POST http://localhost:8082/api/import \
     -H "Content-Type: application/json" \
     --data-binary @pre-upgrade-backup.json
   ```

## Emergency Procedures

### Complete System Reset

**⚠️ WARNING: This will delete all data**

```bash
# Stop all services
docker-compose down

# Remove volumes
docker volume rm distributed-gradle_coordinator_data
docker volume rm distributed-gradle_ml_data
docker volume rm distributed-gradle_cache_data

# Clean restart
docker-compose up -d

# Verify services
./health-check.sh
```

### Single Service Recovery

```bash
# Restart failing service
docker-compose restart ml-service

# Check logs
docker-compose logs --tail=50 ml-service

# Force recreation
docker-compose up -d --force-recreate ml-service
```

## Support Resources

### Log Collection

```bash
#!/bin/bash
# collect-logs.sh

timestamp=$(date +%Y%m%d_%H%M%S)
log_dir="logs_$timestamp"

mkdir -p $log_dir

# Collect all service logs
docker-compose logs > $log_dir/all_services.log

# Collect individual service logs
services=("coordinator" "ml-service" "monitor" "cache" "worker-1" "worker-2" "worker-3")
for service in "${services[@]}"; do
    docker-compose logs $service > $log_dir/${service}.log 2>&1
done

# Collect system information
docker system info > $log_dir/docker_info.txt
docker stats --no-stream > $log_dir/docker_stats.txt

# Create archive
tar czf ${log_dir}.tar.gz $log_dir
rm -rf $log_dir

echo "Logs collected in ${log_dir}.tar.gz"
```

### Diagnostic Commands

```bash
# System information
uname -a
docker --version
docker-compose --version

# Network diagnostics
docker network ls
docker network inspect distributed-gradle_default

# Volume information
docker volume ls
docker volume inspect distributed-gradle_ml_data

# Resource usage
docker system df
df -h
free -h
```

### Getting Help

1. **Check documentation:**
   - [Deployment Guide](DEPLOYMENT_GUIDE.md)
   - [Operations Manual](OPERATIONS_MANUAL.md)
   - [API Reference](API_REFERENCE.md)

2. **Community support:**
   - GitHub Issues: Report bugs and request features
   - GitHub Discussions: Ask questions and share solutions
   - Stack Overflow: Tag questions with `distributed-gradle-building`

3. **Enterprise support:**
   - Contact your system administrator
   - Review internal runbooks and procedures

---

**Remember:** When troubleshooting, always collect logs and system information before making changes. Start with the simplest solutions and work systematically through the issue.</content>
</xai:function_call">### 6. Performance Tuning Guide

<xai:function_call name="write">
<parameter name="filePath">docs/PERFORMANCE_TUNING_GUIDE.md