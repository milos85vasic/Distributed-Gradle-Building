#!/bin/sh
set -e

# Create config directory if it doesn't exist
mkdir -p /app/config

# Generate config file from environment variables
cat > /app/config/monitor-config.json << EOF
{
  "Port": ${MONITOR_PORT:-8084},
  "CoordinatorURL": "http://${COORDINATOR_HOST:-coordinator}:${COORDINATOR_PORT:-8080}",
  "CacheURL": "http://${CACHE_HOST:-cache}:${CACHE_PORT:-8083}",
  "PollInterval": 30,
  "RetentionDays": 7,
  "AlertEnabled": true,
  "AlertThreshold": 80
}
EOF

echo "Generated config for monitor:"
cat /app/config/monitor-config.json

# Start the monitor
exec ./monitor /app/config/monitor-config.json