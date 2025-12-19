#!/bin/sh
set -e

# Create config directory if it doesn't exist
mkdir -p /app/config

# Generate config file from environment variables
cat > /app/config/cache-config.json << EOF
{
  "Host": "${CACHE_HOST:-}",
  "Port": ${CACHE_PORT:-8083},
  "StorageType": "${STORAGE_TYPE:-filesystem}",
  "StorageDir": "${STORAGE_DIR:-/app/data/cache}",
  "MaxCacheSize": ${MAX_CACHE_SIZE:-10737418240},
  "TTL": ${TTL_SECONDS:-86400},
  "Compression": ${COMPRESSION:-true},
  "Authentication": ${AUTHENTICATION:-false}
}
EOF

echo "Generated config for cache server:"
cat /app/config/cache-config.json

# Start the cache server
exec ./cache_server /app/config/cache-config.json