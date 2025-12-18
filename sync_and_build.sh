#!/bin/bash

set -e

if [[ -f ".gradlebuild_env" ]]; then
  source .gradlebuild_env
else
  echo "Run setup_master.sh first."
  exit 1
fi

cd "$PROJECT_DIR"

# Sync to workers
echo "Syncing code to workers..."
for ip in $WORKER_IPS; do
  rsync -a --delete -e ssh \
    --exclude='build/' --exclude='.gradle/' --exclude='*.iml' --exclude='.idea/' \
    "$PROJECT_DIR"/ "$ip":"$PROJECT_DIR"/
done

# Build with high parallelism
echo "Starting Gradle build with --parallel --max-workers=$MAX_WORKERS"
./gradlew assemble --parallel --max-workers=\( MAX_WORKERS " \)@"  # Or :app:bundle, clean, etc.

echo "Build complete."
./gradlew --status  # Optional: Show daemon status
