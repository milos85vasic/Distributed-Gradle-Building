#!/bin/bash

set -e

if [[ -f ".gradlebuild_env" ]]; then
  source .gradlebuild_env
else
  echo "ERROR: .gradlebuild_env not found. Run setup_master.sh first."
  exit 1
fi

# Check if the distributed build script exists
DISTRIBUTED_SCRIPT="$(dirname "$0")/distributed_gradle_build.sh"
if [[ -f "$DISTRIBUTED_SCRIPT" ]]; then
  echo "🚀 Using True Distributed Build System"
  echo "====================================="
  echo ""
  echo "The original sync_and_build.sh has been upgraded to use"
  echo "true distributed building with worker CPU/memory utilization."
  echo ""
  echo "Old behavior: File sync + local parallel build"
  echo "New behavior: Real distributed build across workers"
  echo ""
  
  # Execute the true distributed build
  exec "$DISTRIBUTED_SCRIPT" "$@"
else
  echo "ERROR: Distributed build script not found at $DISTRIBUTED_SCRIPT"
  echo ""
  echo "This script has been upgraded to use true distributed building."
  echo "Please ensure distributed_gradle_build.sh is available."
  exit 1
fi
