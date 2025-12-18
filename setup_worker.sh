#!/bin/bash

set -e

PROJECT_DIR="$1"

if [[ -z "$PROJECT_DIR" ]]; then
  echo "Usage: $0 <project_directory>"
  exit 1
fi

PROJECT_DIR=$(realpath "$PROJECT_DIR")

read -p "Enter master IP address: " MASTER_IP

sudo apt update
sudo apt install -y rsync openjdk-17-jdk  # Or your JDK version

mkdir -p "$PROJECT_DIR"

echo "Worker setup complete for $PROJECT_DIR."
