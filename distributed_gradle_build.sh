#!/bin/bash
# True Distributed Gradle Build System
# Implements actual distributed compilation across workers

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BUILD_ID="build-$(date +%s)"
METRICS_FILE="/tmp/distributed-build-metrics-$BUILD_ID.json"

# Load configuration
if [[ -f ".gradlebuild_env" ]]; then
    source .gradlebuild_env
else
    echo "ERROR: .gradlebuild_env not found. Run setup_master.sh first."
    exit 1
fi

# Initialize metrics
init_metrics() {
    cat > "$METRICS_FILE" << EOF
{
  "build_id": "$BUILD_ID",
  "start_time": $(date +%s),
  "workers": [],
  "tasks": [],
  "metrics": {
    "total_tasks": 0,
    "completed_tasks": 0,
    "failed_tasks": 0,
    "worker_cpu_usage": {},
    "worker_memory_usage": {},
    "network_transfer_mb": 0,
    "build_duration_seconds": 0
  }
}
EOF
}

# Update metrics
update_metrics() {
    local key="$1"
    local value="$2"
    if command -v jq >/dev/null 2>&1; then
        jq ".$key = $value" "$METRICS_FILE" > "${METRICS_FILE}.tmp" && mv "${METRICS_FILE}.tmp" "$METRICS_FILE"
    fi
}

# Analyze Gradle build tasks
analyze_tasks() {
    echo "🔍 Analyzing Gradle build tasks..."
    
    # Get task list from Gradle
    cd "$PROJECT_DIR"
    local task_output
    task_output=$(./gradlew tasks --all 2>/dev/null || echo "")
    
    # Extract relevant tasks for distribution
    local tasks=()
    while IFS= read -r line; do
        if [[ $line =~ ^[[:space:]]*([a-zA-Z][a-zA-Z0-9:]*)[[:space:]]+- ]]; then
            tasks+=("${BASH_REMATCH[1]}")
        fi
    done <<< "$task_output"
    
    # Filter tasks that can be parallelized
    local parallelizable_tasks=()
    for task in "${tasks[@]}"; do
        case "$task" in
            compile*|processResources|testClasses|javadoc|jar)
                parallelizable_tasks+=("$task")
                ;;
        esac
    done
    
    echo "  Found ${#parallelizable_tasks[@]} parallelizable tasks"
    printf '%s\n' "${parallelizable_tasks[@]}"
}

# Get worker status
get_worker_status() {
    local worker_ip="$1"
    local worker_info
    
    echo "🔍 Checking worker status: $worker_ip"
    
    # Check SSH connectivity
    if ! ssh -o BatchMode=yes -o ConnectTimeout=5 "$worker_ip" "echo 'OK'" >/dev/null 2>&1; then
        echo "  ❌ Worker $worker_ip: SSH connection failed"
        return 1
    fi
    
    # Get worker resources
    worker_info=$(ssh "$worker_ip" "
        # CPU info
        cpu_cores=\$(nproc 2>/dev/null || echo '4')
        cpu_load=\$(uptime | awk -F'load average:' '{print \$2}' | awk '{print \$1}' | sed 's/,//' || echo '0.5')
        
        # Memory info
        memory_total=\$(free -m 2>/dev/null | awk 'NR==2{print \$2}' || echo '8192')
        memory_available=\$(free -m 2>/dev/null | awk 'NR==2{print \$7}' || echo '4096')
        memory_usage_pct=\$(((memory_total - memory_available) * 100 / memory_total))
        
        # Disk space
        disk_available=\$(df -h . 2>/dev/null | tail -1 | awk '{print \$4}' || echo '10G')
        
        echo '{\"cpu_cores\": '\$cpu_cores', \"cpu_load\": '\$cpu_load', \"memory_total_mb\": '\$memory_total', \"memory_usage_pct\": '\$memory_usage_pct', \"disk_available\": \"'\$disk_available'\"}'
    " 2>/dev/null || echo '{"error": "Failed to get worker info"}')
    
    echo "  ✅ Worker $worker_ip: Available"
    echo "     CPU: $(echo "$worker_info" | jq -r '.cpu_cores // "unknown"') cores, load: $(echo "$worker_info" | jq -r '.cpu_load // "unknown"')"
    echo "     Memory: $(echo "$worker_info" | jq -r '.memory_usage_pct // "unknown"')% used, $(echo "$worker_info" | jq -r '.memory_total_mb // "unknown"')MB total"
    echo "     Disk: $(echo "$worker_info" | jq -r '.disk_available // "unknown"') available"
    
    echo "$worker_info"
    return 0
}

# Select optimal workers for tasks
select_workers() {
    local tasks=("$@")
    local worker_selections=()
    local worker_count=0
    
    echo "🎯 Selecting optimal workers for ${#tasks[@]} tasks..."
    
    # Get status of all workers
    declare -A worker_status
    local available_workers=()
    
    for ip in $WORKER_IPS; do
        if status=$(get_worker_status "$ip"); then
            worker_status["$ip"]="$status"
            available_workers+=("$ip")
        fi
    done
    
    worker_count=${#available_workers[@]}
    
    if [[ $worker_count -eq 0 ]]; then
        echo "  ❌ No workers available"
        return 1
    fi
    
    # Simple round-robin assignment with load consideration
    local task_idx=0
    for task in "${tasks[@]}"; do
        local worker_ip="${available_workers[$((task_idx % worker_count))]}"
        worker_selections+=("$task:$worker_ip")
        echo "    $task → $worker_ip"
        task_idx=$((task_idx + 1))
    done
    
    printf '%s\n' "${worker_selections[@]}"
}

# Sync project to worker
sync_to_worker() {
    local worker_ip="$1"
    local worker_dir="$2"
    
    echo "📤 Syncing project to worker $worker_ip..."
    
    # Create directory on worker
    ssh "$worker_ip" "mkdir -p '$worker_dir'" || {
        echo "  ❌ Failed to create directory on $worker_ip"
        return 1
    }
    
    # Sync project (excluding build artifacts)
    local rsync_output
    if rsync_output=$(rsync -a --delete -e ssh \
        --exclude='build/' --exclude='.gradle/' --exclude='*.iml' --exclude='.idea/' \
        --exclude='node_modules/' --exclude='.git/' \
        "$PROJECT_DIR"/ "$worker_ip":"$worker_dir"/ 2>&1); then
        echo "  ✅ Synced to $worker_ip"
        echo "$rsync_output" | tail -3
    else
        echo "  ❌ Failed to sync to $worker_ip"
        echo "$rsync_output"
        return 1
    fi
}

# Execute task on worker
execute_task_on_worker() {
    local task="$1"
    local worker_ip="$2"
    local worker_dir="$3"
    local task_log="$4"
    
    echo "🔧 Executing $task on worker $worker_ip..."
    
    local task_start_time=$(date +%s)
    
    # Execute the task remotely and monitor resources
    local remote_result
    remote_result=$(ssh "$worker_ip" "
        cd '$worker_dir'
        
        # Start resource monitoring in background
        monitor_pid=\$(
            {
                while true; do
                    cpu_load=\$(uptime | awk -F'load average:' '{print \$2}' | awk '{print \$1}' | sed 's/,//')
                    memory_usage=\$(free -m | awk 'NR==2{printf \"%.1f\", \$3*100/\$2}')
                    echo \"\$(date +%s),\$cpu_load,\$memory_usage\" >> '$task_log.resources'
                    sleep 2
                done
            } &
            echo \$!
        )
        
        # Run the Gradle task
        task_result=0
        if ./gradlew '$task' > '$task_log.stdout' 2> '$task_log.stderr'; then
            task_result=0
        else
            task_result=1
        fi
        
        # Stop monitoring
        kill \$monitor_pid 2>/dev/null || true
        
        # Collect artifacts
        artifacts=''
        if [[ -d 'build' ]]; then
            artifacts=\$(find build -name '*.class' -o -name '*.jar' 2>/dev/null | wc -l)
        fi
        
        # Calculate duration
        task_end_time=\$(date +%s)
        duration=\$((task_end_time - $task_start_time))
        
        echo '{\"task_result\": '\$task_result', \"artifacts_count\": '\$artifacts', \"duration\": '\$duration', \"worker_ip\": \"$worker_ip\"}'
    " 2>&1)
    
    local task_result=$(echo "$remote_result" | jq -r '.task_result // 1')
    local artifacts_count=$(echo "$remote_result" | jq -r '.artifacts_count // 0')
    local duration=$(echo "$remote_result" | jq -r '.duration // 0')
    
    if [[ $task_result -eq 0 ]]; then
        echo "  ✅ Completed $task on $worker_ip (${duration}s, ${artifacts_count} artifacts)"
        update_metrics "metrics.completed_tasks" "$(jq -r '.metrics.completed_tasks + 1' "$METRICS_FILE")"
    else
        echo "  ❌ Failed $task on $worker_ip (${duration}s)"
        update_metrics "metrics.failed_tasks" "$(jq -r '.metrics.failed_tasks + 1' "$METRICS_FILE")"
    fi
    
    # Update worker metrics
    if [[ -f "$task_log.resources" ]]; then
        local avg_cpu=$(awk -F',' '{sum+=$2; count++} END {if(count>0) print sum/count; else print 0}' "$task_log.resources")
        local avg_memory=$(awk -F',' '{sum+=$3; count++} END {if(count>0) print sum/count; else print 0}' "$task_log.resources")
        
        update_metrics "metrics.worker_cpu_usage.\"$worker_ip\"" "$avg_cpu"
        update_metrics "metrics.worker_memory_usage.\"$worker_ip\"" "$avg_memory"
    fi
    
    return $task_result
}

# Collect artifacts from workers
collect_artifacts() {
    echo "📦 Collecting artifacts from workers..."
    
    local collected_artifacts=0
    for ip in $WORKER_IPS; do
        echo "  Collecting from $ip..."
        
        # Create destination directory
        mkdir -p "$PROJECT_DIR/build/distributed/$ip"
        
        # Collect build artifacts
        if rsync -a -e ssh "$ip":"$PROJECT_DIR/build/" "$PROJECT_DIR/build/distributed/$ip/" 2>/dev/null; then
            local artifact_count
            artifact_count=$(find "$PROJECT_DIR/build/distributed/$ip" -name '*.class' -o -name '*.jar' 2>/dev/null | wc -l)
            echo "    Collected $artifact_count artifacts from $ip"
            collected_artifacts=$((collected_artifacts + artifact_count))
        else
            echo "    No artifacts found on $ip or collection failed"
        fi
    done
    
    echo "  Total artifacts collected: $collected_artifacts"
    return 0
}

# Generate final build report
generate_build_report() {
    local end_time=$(date +%s)
    local duration=$((end_time - $(jq -r '.start_time' "$METRICS_FILE")))
    
    update_metrics "end_time" "$end_time"
    update_metrics "metrics.build_duration_seconds" "$duration"
    
    echo ""
    echo "📊 DISTRIBUTED BUILD REPORT"
    echo "=========================="
    echo "Build ID: $BUILD_ID"
    echo "Duration: ${duration}s"
    echo "Workers: $(echo "$WORKER_IPS" | wc -w)"
    echo ""
    
    if command -v jq >/dev/null 2>&1; then
        echo "Tasks Summary:"
        echo "  Total: $(jq -r '.metrics.total_tasks' "$METRICS_FILE")"
        echo "  Completed: $(jq -r '.metrics.completed_tasks' "$METRICS_FILE")"
        echo "  Failed: $(jq -r '.metrics.failed_tasks' "$METRICS_FILE")"
        echo ""
        
        echo "Worker Utilization:"
        jq -r 'to_entries[] | select(.key | startswith("metrics.worker_")) | "  \(.key): \(.value)"' "$METRICS_FILE" | sed 's/metrics.worker_cpu_usage\./CPU (/; s/metrics.worker_memory_usage\./Memory (/; s/$/)/'
        echo ""
        
        echo "Artifacts Collected: $(find "$PROJECT_DIR/build/distributed" -name '*.class' -o -name '*.jar' 2>/dev/null | wc -l)"
    fi
    
    echo "Full metrics saved to: $METRICS_FILE"
}

# Main distributed build function
distributed_build() {
    local target_task="${1:-assemble}"
    
    echo "🚀 Starting Distributed Gradle Build"
    echo "=================================="
    echo "Build ID: $BUILD_ID"
    echo "Target Task: $target_task"
    echo "Workers: $WORKER_IPS"
    echo ""
    
    # Initialize metrics
    init_metrics
    
    # Analyze tasks
    local tasks
    mapfile -t tasks < <(analyze_tasks)
    
    if [[ ${#tasks[@]} -eq 0 ]]; then
        echo "❌ No tasks found to distribute"
        return 1
    fi
    
    update_metrics "metrics.total_tasks" "${#tasks[@]}"
    
    # Select workers
    local task_assignments=()
    mapfile -t task_assignments < <(select_workers "${tasks[@]}")
    
    # Sync to all workers
    echo ""
    echo "📤 Syncing project to all workers..."
    for ip in $WORKER_IPS; do
        sync_to_worker "$ip" "$PROJECT_DIR" &
    done
    wait
    echo "  ✅ Sync completed"
    
    # Execute tasks in parallel
    echo ""
    echo "🔧 Executing distributed tasks..."
    local task_pids=()
    local task_logs=()
    
    for assignment in "${task_assignments[@]}"; do
        IFS=':' read -r task worker_ip <<< "$assignment"
        local task_log="/tmp/${BUILD_ID}-${task}-${worker_ip//./-}.log"
        
        mkdir -p "$(dirname "$task_log")"
        
        execute_task_on_worker "$task" "$worker_ip" "$PROJECT_DIR" "$task_log" &
        task_pids+=($!)
        task_logs+=("$task_log")
    done
    
    # Wait for all tasks to complete
    for pid in "${task_pids[@]}"; do
        wait "$pid"
    done
    
    # Collect artifacts
    collect_artifacts
    
    # Generate report
    generate_build_report
    
    echo ""
    echo "✅ Distributed build completed!"
    return 0
}

# Cleanup function
cleanup() {
    rm -f /tmp/${BUILD_ID}-*.log* 2>/dev/null || true
    rm -f "${METRICS_FILE}.tmp" 2>/dev/null || true
}

# Trap cleanup
trap cleanup EXIT

# Show usage
usage() {
    echo "Usage: $0 [task]"
    echo ""
    echo "Examples:"
    echo "  $0 assemble     # Default: assemble the project"
    echo "  $0 test         # Run distributed tests"
    echo "  $0 clean        # Clean distributed builds"
    echo ""
    echo "Environment:"
    echo "  WORKER_IPS      - Space-separated list of worker IPs"
    echo "  PROJECT_DIR     - Project directory path"
    echo "  MAX_WORKERS     - Maximum parallel workers per machine"
}

# Main execution
main() {
    case "${1:-}" in
        -h|--help|help)
            usage
            exit 0
            ;;
        clean)
            echo "🧹 Cleaning distributed build artifacts..."
            rm -rf "$PROJECT_DIR/build/distributed"
            echo "  ✅ Cleaned"
            exit 0
            ;;
        *)
            distributed_build "${1:-assemble}"
            ;;
    esac
}

# Execute main function
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi