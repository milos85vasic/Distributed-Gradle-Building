#!/bin/sh
set -e

# Create config directory if it doesn't exist
mkdir -p /app/config

# Generate config file from environment variables
cat > /app/config/worker-config.json << EOF
{
  "ID": "${WORKER_ID:-worker-1}",
  "CoordinatorURL": "http://${COORDINATOR_HOST:-coordinator}:${COORDINATOR_PORT:-8080}",
  "HTTPPort": ${WORKER_PORT:-8082},
  "RPCPort": $((${WORKER_PORT:-8082} + 1)),
  "BuildDir": "/app/work-dir",
  "CacheEnabled": true,
  "MaxConcurrentBuilds": ${MAX_BUILDS:-5},
  "WorkerType": "standard"
}
EOF

echo "Generated config for worker ${WORKER_ID:-worker-1}:"
cat /app/config/worker-config.json

# Start the worker
exec ./worker /app/config/worker-config.json