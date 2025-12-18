#!/bin/bash

set -e

PROJECT_DIR="$1"

if [[ -z "$PROJECT_DIR" ]]; then
  echo "Usage: $0 <project_directory>"
  exit 1
fi

PROJECT_DIR=$(realpath "$PROJECT_DIR")

read -p "Enter worker IP addresses (space-separated): " -a WORKER_IPS

sudo apt update
sudo apt install -y rsync openjdk-17-jdk

# Test SSH
for ip in "${WORKER_IPS[@]}"; do
  if ! ssh -o BatchMode=yes -o ConnectTimeout=5 "$ip" echo "OK" >/dev/null 2>&1; then
    echo "Error: Passwordless SSH to $ip failed."
    exit 1
  fi
done

# Auto-detect for parallel
CORES=$(nproc --all)
MAX_WORKERS=$((CORES * 2))  # Overcommit safely

# Config file
CONFIG_FILE="$PROJECT_DIR/.gradlebuild_env"
cat <<EOF > "$CONFIG_FILE"
export WORKER_IPS="${WORKER_IPS[*]}"
export MAX_WORKERS=$MAX_WORKERS
export PROJECT_DIR="$PROJECT_DIR"
EOF

# Optional: Simple remote cache server (Python HTTP on port 5000)
echo "Starting simple HTTP build cache server (optional, for remote caching)..."
echo "Run: cd $PROJECT_DIR && mkdir -p build-cache && python3 -m http.server 5000 --directory build-cache"
echo "Then add to gradle.properties: see instructions below."

echo "Master setup complete. Config in $CONFIG_FILE"
