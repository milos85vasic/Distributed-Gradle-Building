# User Guide - Day-to-Day Usage

This guide covers how to use the Distributed Gradle Building System effectively after initial setup.

## Basic Usage

### Running a Build

From your project directory on the master machine:
```bash
/path/to/Distributed-Gradle-Building/sync_and_build.sh
```

This will:
1. Sync your code to all worker machines
2. Run the build with optimal parallelism
3. Display build results

### Running Specific Gradle Tasks

Pass any Gradle task as additional arguments:
```bash
# Build specific module
./sync_and_build.sh :app:assembleDebug

# Clean and build
./sync_and_build.sh clean assemble

# Run tests
./sync_and_build.sh test

# Build release bundle
./sync_and_build.sh :app:bundleRelease
```

### Dry Run (Sync Only)

To sync code without building:
```bash
# Edit sync_and_build.sh temporarily to comment out the build command
# or create a separate sync-only script
```

## Understanding the Build Process

### Sync Phase

The rsync process excludes:
- `build/` - Build outputs
- `.gradle/` - Gradle cache and daemon files
- `*.iml` - IntelliJ IDEA module files
- `.idea/` - IDE-specific files

This ensures only source code is distributed.

### Build Phase

The build runs with:
- `--parallel` - Execute tasks in parallel
- `--max-workers=$MAX_WORKERS` - Optimal worker count
- Your additional task arguments

## Common Workflows

### 1. Development Workflow

```bash
# Make changes in your IDE
git add .
git commit -m "Feature implementation"

# Sync and build
./sync_and_build.sh assembleDebug

# Test on device/emulator
# ...
```

### 2. Release Workflow

```bash
# Ensure clean state
./sync_and_build.sh clean

# Build release
./sync_and_build.sh :app:bundleRelease

# Deploy AAB/APK
# ...
```

### 3. Testing Workflow

```bash
# Run all tests
./sync_and_build.sh test

# Run specific test class
./sync_and_build.sh test --tests "*.MyTestClass"

# Run tests with coverage
./sync_and_build.sh test jacocoTestReport
```

## Performance Tips

### 1. Optimize Worker Count

The setup automatically calculates `MAX_WORKERS = CPU_CORES * 2`. Adjust if needed:
```bash
# Edit .gradlebuild_env
export MAX_WORKERS=12  # Custom value
```

### 2. Use Build Caching

Add to `gradle.properties`:
```properties
org.gradle.caching=true
```

### 3. Configure Memory

For large projects, increase JVM memory in `gradle.properties`:
```properties
org.gradle.jvmargs=-Xmx6g -XX:+UseG1GC
```

### 4. Exclude Large Directories

If you have large directories that don't need syncing, modify `sync_and_build.sh`:
```bash
rsync -a --delete -e ssh \
  --exclude='build/' --exclude='.gradle/' --exclude='*.iml' --exclude='.idea/' \
  --exclude='large-data/' --exclude='test-fixtures/' \
  "$PROJECT_DIR"/ "$ip":"$PROJECT_DIR"/
```

## Monitoring and Debugging

### Check Build Status

```bash
# Check Gradle daemon status
./gradlew --status

# View recent builds
./gradlew --continuous --info
```

### Monitor Sync Operations

Add logging to `sync_and_build.sh`:
```bash
echo "$(date): Syncing to $ip"
rsync -a --delete --stats -e ssh \
  --exclude='build/' --exclude='.gradle/' --exclude='*.iml' --exclude='.idea/' \
  "$PROJECT_DIR"/ "$ip":"$PROJECT_DIR"/
echo "$(date): Sync to $ip completed"
```

### Check Worker Status

From master, verify workers are accessible:
```bash
for ip in $WORKER_IPS; do
  echo "Checking $ip..."
  ssh $ip "df -h $PROJECT_DIR && echo 'OK'"
done
```

## Advanced Usage

### 1. Selective Syncing

Create a wrapper script for specific scenarios:
```bash
#!/bin/bash
# quick_build.sh - Skip tests for faster builds
/path/to/Distributed-Gradle-Building/sync_and_build.sh assemble -x test
```

### 2. Staged Builds

For multi-stage builds:
```bash
# Stage 1: Compile only
./sync_and_build.sh compileJava compileKotlin

# Stage 2: Run tests
./sync_and_build.sh test

# Stage 3: Package
./sync_and_build.sh assemble
```

### 3. Continuous Integration

Integrate with CI/CD pipelines:
```bash
# In your CI script
export WORKER_IPS="worker1 worker2"
export MAX_WORKERS=8
export PROJECT_DIR="$CI_PROJECT_DIR"

/path/to/Distributed-Gradle-Building/sync_and_build.sh assemble test
```

## Configuration Management

### Environment-Specific Builds

Create different config files:
```bash
# For dev builds
cp .gradlebuild_env .gradlebuild_env.dev
# Edit .gradlebuild_env.dev for dev-specific settings

# For production builds
cp .gradlebuild_env .gradlebuild_env.prod
# Edit .gradlebuild_env.prod for prod settings
```

Create a wrapper script:
```bash
#!/bin/bash
# build_with_config.sh
ENV_TYPE=${1:-dev}
source ".gradlebuild_env.$ENV_TYPE"
/path/to/Distributed-Gradle-Building/sync_and_build.sh "${@:2}"
```

### Branch-Specific Workers

For different branches with different resource needs:
```bash
# Edit sync_and_build.sh
case "$(git branch --show-current)" in
  "main")
    export MAX_WORKERS=16
    ;;
  "develop")
    export MAX_WORKERS=8
    ;;
  *)
    export MAX_WORKERS=4
    ;;
esac
```

## Team Collaboration

### Shared Configuration

1. Commit configuration templates:
```bash
# gradlebuild_env.template
export WORKER_IPS="<YOUR_WORKER_IPS>"
export MAX_WORKERS="<CALCULATED_WORKERS>"
export PROJECT_DIR="<PROJECT_PATH>"
```

2. Each developer creates their own `.gradlebuild_env` from template

### Work-in-Progress Syncing

For sharing work between developers:
```bash
# Create a WIP sync script
#!/bin/bash
# wip_sync.sh - Sync specific branches to teammates
BRANCH=$(git branch --show-current)
for ip in TEAMMATE_IPS; do
  ssh $ip "cd $PROJECT_DIR && git fetch origin $BRANCH && git checkout $BRANCH"
done
```

## Troubleshooting Common Issues

### Sync Failures

**Problem**: rsync reports permission errors
```bash
# Solution: Check directory permissions
ssh worker_ip "ls -la $PROJECT_DIR"
# Fix if needed
ssh worker_ip "chmod -R u+w $PROJECT_DIR"
```

**Problem**: Sync is very slow
```bash
# Solution: Check network speed and exclude more files
# Add to rsync excludes:
--exclude='node_modules/' --exclude='npm_cache/'
```

### Build Failures

**Problem**: OutOfMemoryError during build
```bash
# Solution: Increase heap size
export GRADLE_OPTS="-Xmx4g"
# Or in gradle.properties:
org.gradle.jvmargs=-Xmx4g
```

**Problem**: Worker machines not keeping up
```bash
# Solution: Reduce parallelism
export MAX_WORKERS=4
# Check CPU usage on workers
ssh worker_ip "top"
```

## Best Practices

### 1. Regular Maintenance
- Clean Gradle caches weekly: `./gradlew clean`
- Monitor disk space on workers
- Update Java versions consistently

### 2. Version Control
- Commit changes before syncing to workers
- Use consistent Git versions across machines
- Tag releases for reproducible builds

### 3. Network Optimization
- Use wired connections for workers
- Configure QoS to prioritize build traffic
- Monitor network latency

### 4. Security
- Use SSH keys instead of passwords
- Regularly rotate SSH keys
- Monitor SSH access logs

## Integration with IDEs

### IntelliJ IDEA

Create a custom run configuration:
1. Run → Edit Configurations
2. Add "Shell Script" configuration
3. Command: `/path/to/sync_and_build.sh`
4. Program arguments: `assembleDebug`

### VS Code

Create a task in `.vscode/tasks.json`:
```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "Distributed Build",
            "type": "shell",
            "command": "/path/to/sync_and_build.sh",
            "args": ["assembleDebug"],
            "group": "build",
            "presentation": {
                "reveal": "always",
                "panel": "new"
            }
        }
    ]
}
```

## Performance Measurement

### Time Your Builds

```bash
# Create timing script
#!/bin/bash
# timed_build.sh
START=$(date +%s)
/path/to/Distributed-Gradle-Building/sync_and_build.sh "$@"
END=$(date +%s)
echo "Build time: $((END-START)) seconds"
```

### Compare with Local Builds

```bash
# Local build timing
START=$(date +%s)
./gradlew "$@"
END=$(date +%s)
echo "Local build time: $((END-START)) seconds"
```

### Optimization Checklist

- [ ] Remote cache configured and working
- [ ] Optimal worker count set
- [ ] Memory allocation sufficient
- [ ] Network is stable and fast
- [ ] Excludes are comprehensive
- [ ] Gradle daemon is warm

## Emergency Procedures

### Worker Failure

If a worker goes down:
1. Remove from worker list in `.gradlebuild_env`
2. Restart build with remaining workers
3. Replace worker when available

### Master Failure

If master machine fails:
1. Promote a worker to master:
   - Copy project to new master
   - Update `.gradlebuild_env`
   - Update worker configurations
2. Rerun setup on new master

### Network Outage

For temporary network issues:
1. Switch to local build mode:
   ```bash
   ./gradlew assemble  # Bypass distribution
   ```
2. Fix network connectivity
3. Resume distributed builds