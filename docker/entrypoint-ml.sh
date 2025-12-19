#!/bin/sh
set -e

# Create config directory if it doesn't exist
mkdir -p /app/config

# Generate config file from environment variables
cat > /app/config/ml-config.json << EOF
{
  "Port": ${ML_SERVICE_PORT:-8082},
  "LogLevel": "${LOG_LEVEL:-info}",
  "DataDir": "/app/data",
  "MaxHistoryRecords": 10000,
  "TrainingInterval": 3600
}
EOF

echo "Generated config for ML service:"
cat /app/config/ml-config.json

# Start the ML service
exec ./ml-service