# Distributed Gradle Building - Agent Guide

This repository contains scripts and configuration for setting up distributed Gradle builds across multiple machines using a master-worker architecture.

## Project Overview

This is a shell-based distributed build system for Gradle projects that enables parallel compilation across multiple worker machines. The system uses rsync for code synchronization and SSH for communication between master and worker nodes.

## Key Components

### Shell Scripts
- `setup_master.sh` - Configures the master build node
- `setup_worker.sh` - Configures worker nodes
- `setup_passwordless_ssh.sh` - Sets up SSH key authentication between machines
- `sync_and_build.sh` - Syncs code to workers and executes distributed build
- `Upstreams/GitHub.sh` - Contains upstream repository reference

### Configuration
- `.gradlebuild_env` - Environment configuration file created by setup_master.sh

## Essential Commands

### Initial Setup
1. Set up passwordless SSH between all machines:
   ```bash
   ./setup_passwordless_ssh.sh
   ```

2. Configure master node (run on master):
   ```bash
   ./setup_master.sh <project_directory>
   ```

3. Configure worker nodes (run on each worker):
   ```bash
   ./setup_worker.sh <project_directory>
   ```

### Building
Execute distributed build (run from repository root):
```bash
./sync_and_build.sh
```

## Setup Process

### Prerequisites
- Linux machines (Ubuntu/Debian based on apt usage)
- OpenJDK 17
- SSH access between machines
- Same project directory path on all machines

### Master Configuration
The master setup script:
- Installs required packages (rsync, openjdk-17-jdk)
- Prompts for worker IP addresses
- Tests SSH connectivity to all workers
- Auto-detects CPU cores and calculates optimal worker count
- Creates `.gradlebuild_env` configuration file
- Optionally sets up HTTP build cache server

### Worker Configuration
The worker setup script:
- Installs required packages
- Creates project directory structure
- Configures system for receiving build tasks

## Code Patterns and Conventions

### Shell Script Patterns
- All scripts use `set -e` for strict error handling
- Functions use snake_case naming
- Configuration sourced from `.gradlebuild_env` file
- Error messages output to stderr with clear context

### Build Configuration
- Uses `--parallel --max-workers=<count>` Gradle flags
- Worker count calculated as `(CPU_CORES * 2)` for safe overcommitment
- Excludes build artifacts from rsync: `build/`, `.gradle/`, `.idea/`, `*.iml`

## Environment Variables

The `.gradlebuild_env` file contains:
- `WORKER_IPS` - Space-separated list of worker IP addresses
- `MAX_WORKERS` - Maximum number of parallel workers
- `PROJECT_DIR` - Absolute path to project directory

## Testing and Validation

### SSH Testing
The master setup script automatically tests SSH connectivity to all workers before proceeding.

### Build Validation
After building, the system runs:
```bash
./gradlew --status
```
To show daemon status across the distributed environment.

## Common Gotchas

1. **SSH Authentication**: Must run `setup_passwordless_ssh.sh` before any other setup
2. **Network Connectivity**: All machines must be able to reach each other via SSH
3. **Java Version**: Requires OpenJDK 17 on all machines
4. **Directory Paths**: Project directory must exist with same path on all machines
5. **Firewall**: Ensure SSH ports (22) and optional cache server port (5000) are open

## Optional Features

### Build Caching
The master setup can optionally start an HTTP build cache server:
```bash
cd $PROJECT_DIR && mkdir -p build-cache && python3 -m http.server 5000 --directory build-cache
```

## Repository Structure

```
Distributed-Gradle-Building/
├── README.md (empty)
├── setup_master.sh
├── setup_worker.sh
├── setup_passwordless_ssh.sh
├── sync_and_build.sh
└── Upstreams/
    └── GitHub.sh
```

## Working with This Project

### When Modifying Scripts
- Maintain error handling with `set -e`
- Test SSH connectivity before remote operations
- Use absolute paths for consistency across machines
- Preserve existing rsync exclusion patterns

### When Debugging
- Check `.gradlebuild_env` configuration
- Verify SSH connectivity manually: `ssh <worker-ip> echo "OK"`
- Monitor Gradle daemon status: `./gradlew --status`
- Check rsync logs for sync issues

### Security Notes
- Scripts handle SSH passwords (use with caution in production)
- SSH keys are stored in standard `~/.ssh/` locations
- Consider using SSH agent forwarding instead of password authentication in production environments