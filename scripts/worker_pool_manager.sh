#!/bin/bash

# Worker Pool Manager for Distributed Gradle Building
# Manages and monitors multiple worker pools for different build scenarios

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${CONFIG_FILE:-$SCRIPT_DIR/../workers.conf}"
LOCK_DIR="${SCRIPT_DIR}/../.worker_locks"
WORKER_STATUS_FILE="${SCRIPT_DIR}/../.worker_status"

# Pool types and default configurations
declare -A WORKER_POOLS=(
    ["compilation"]="high-cpu build-1 build-2 build-3 build-4"
    ["testing"]="medium-cpu test-1 test-2"
    ["packaging"]="balanced build-5 build-6"
    ["default"]="all-pool build-1 build-2 build-3 build-4 build-5 build-6"
)

# Pool resource requirements
declare -A POOL_REQUIREMENTS=(
    ["compilation"]="cpu:high memory:medium storage:medium"
    ["testing"]="cpu:medium memory:high storage:low"
    ["packaging"]="cpu:medium memory:medium storage:high"
    ["default"]="cpu:medium memory:medium storage:medium"
)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case "$level" in
        "INFO")  echo -e "${GREEN}[INFO]${NC}  $message" ;;
        "WARN")  echo -e "${YELLOW}[WARN]${NC}  $message" ;;
        "ERROR") echo -e "${RED}[ERROR]${NC} $message" ;;
        "DEBUG") echo -e "${BLUE}[DEBUG]${NC} $message" ;;
        *)       echo -e "$message" ;;
    esac
}

# Initialize directories
init_directories() {
    mkdir -p "$LOCK_DIR" "$WORKER_STATUS_FILE"
    touch "$WORKER_STATUS_FILE/workers"
    touch "$WORKER_STATUS_FILE/pools"
    touch "$WORKER_STATUS_FILE/health"
}

# Load configuration
load_config() {
    if [[ -f "$CONFIG_FILE" ]]; then
        source "$CONFIG_FILE"
        log "INFO" "Loaded configuration from $CONFIG_FILE"
    else
        log "WARN" "Configuration file not found: $CONFIG_FILE"
        create_default_config
    fi
}

# Create default configuration
create_default_config() {
    cat > "$CONFIG_FILE" << 'EOF'
# Worker Pool Configuration
# Format: POOL_NAME="type:worker1 worker2 worker3"

# Example configurations
COMPILATION_POOL="compilation:build-1 build-2 build-3 build-4"
TESTING_POOL="testing:test-1 test-2"
PACKAGING_POOL="packaging:build-5 build-6"
DEFAULT_POOL="default:build-1 build-2 build-3 build-4 build-5 build-6"

# Health check settings
HEALTH_CHECK_INTERVAL=30
HEALTH_CHECK_TIMEOUT=10

# Pool scaling settings
AUTO_SCALE_ENABLED=true
MIN_WORKERS_PER_POOL=1
MAX_WORKERS_PER_POOL=8

# Resource monitoring
RESOURCE_MONITORING=true
CPU_THRESHOLD=80
MEMORY_THRESHOLD=85
EOF
    log "INFO" "Created default configuration at $CONFIG_FILE"
}

# Acquire lock for worker operation
acquire_lock() {
    local operation="$1"
    local lock_file="$LOCK_DIR/${operation}.lock"
    local timeout=30
    local count=0
    
    while [[ $count -lt $timeout ]]; do
        if mkdir "$lock_file" 2>/dev/null; then
            echo $$ > "$lock_file/pid"
            return 0
        fi
        sleep 1
        ((count++))
    done
    
    log "ERROR" "Failed to acquire lock for $operation"
    return 1
}

# Release lock
release_lock() {
    local operation="$1"
    local lock_file="$LOCK_DIR/${operation}.lock"
    
    if [[ -d "$lock_file" ]]; then
        rm -rf "$lock_file"
    fi
}

# Get workers for a specific pool
get_pool_workers() {
    local pool_name="$1"
    local pool_config
    
    # Load pool configuration
    case "$pool_name" in
        "compilation")
            if [[ -n "${COMPILATION_POOL:-}" ]]; then
                pool_config="${COMPILATION_POOL#*:}"
            else
                pool_config="${WORKER_POOLS[$pool_name]}"
            fi
            ;;
        "testing")
            if [[ -n "${TESTING_POOL:-}" ]]; then
                pool_config="${TESTING_POOL#*:}"
            else
                pool_config="${WORKER_POOLS[$pool_name]}"
            fi
            ;;
        "packaging")
            if [[ -n "${PACKAGING_POOL:-}" ]]; then
                pool_config="${PACKAGING_POOL#*:}"
            else
                pool_config="${WORKER_POOLS[$pool_name]}"
            fi
            ;;
        *)
            if [[ -n "${DEFAULT_POOL:-}" ]]; then
                pool_config="${DEFAULT_POOL#*:}"
            else
                pool_config="${WORKER_POOLS[$pool_name]}"
            fi
            ;;
    esac
    
    echo "$pool_config"
}

# Check worker health
check_worker_health() {
    local worker_ip="$1"
    local timeout="${HEALTH_CHECK_TIMEOUT:-10}"
    
    # Check SSH connectivity
    if ! timeout "$timeout" ssh -o ConnectTimeout=5 -o BatchMode=yes "$worker_ip" "echo 'OK'" >/dev/null 2>&1; then
        echo "failed"
        return 1
    fi
    
    # Check system resources
    local cpu_usage mem_usage disk_usage
    cpu_usage=$(timeout "$timeout" ssh -o ConnectTimeout=5 -o BatchMode=yes "$worker_ip" \
        "top -bn1 | grep 'Cpu(s)' | awk '{print \$2}' | cut -d'%' -f1" 2>/dev/null || echo "100")
    
    mem_usage=$(timeout "$timeout" ssh -o ConnectTimeout=5 -o BatchMode=yes "$worker_ip" \
        "free | grep '^Mem:' | awk '{printf \"%.0f\", \$3/\$2 * 100.0}'" 2>/dev/null || echo "100")
    
    disk_usage=$(timeout "$timeout" ssh -o ConnectTimeout=5 -o BatchMode=yes "$worker_ip" \
        "df / | awk 'NR==2{printf \"%.0f\", \$5}'" 2>/dev/null || echo "100")
    
    # Determine health status
    local cpu_threshold="${CPU_THRESHOLD:-80}"
    local mem_threshold="${MEMORY_THRESHOLD:-85}"
    
    if (( $(echo "$cpu_usage > $cpu_threshold" | bc -l) )); then
        echo "overloaded"
    elif (( $(echo "$mem_usage > $mem_threshold" | bc -l) )); then
        echo "overloaded"
    elif [[ "${disk_usage%?}" -gt 90 ]]; then
        echo "overloaded"
    else
        echo "healthy"
    fi
}

# Update worker status
update_worker_status() {
    local worker_ip="$1"
    local status="$2"
    local pool_name="$3"
    
    # Update worker status file
    awk -F, -v worker="$worker_ip" -v status="$status" -v pool="$pool_name" '
    BEGIN { found=0 }
    $1 == worker { 
        $2=status; $3=pool; found=1; print $0; next 
    }
    { print }
    END { 
        if(!found) printf "%s,%s,%s,%s\n", worker, status, pool, systime()
    }' "$WORKER_STATUS_FILE/workers" > "$WORKER_STATUS_FILE/workers.tmp" && \
    mv "$WORKER_STATUS_FILE/workers.tmp" "$WORKER_STATUS_FILE/workers"
}

# Get available workers from a pool
get_available_workers() {
    local pool_name="$1"
    local workers
    workers=$(get_pool_workers "$pool_name")
    
    for worker in $workers; do
        local status=$(check_worker_health "$worker")
        
        if [[ "$status" == "healthy" ]]; then
            echo "$worker"
        fi
    done
}

# Get optimal number of workers for a task
get_optimal_worker_count() {
    local task_type="$1"
    local available_workers
    available_workers=$(get_available_workers "$task_type" | wc -w)
    
    local min_workers="${MIN_WORKERS_PER_POOL:-1}"
    local max_workers="${MAX_WORKERS_PER_POOL:-8}"
    
    # Task-specific optimization
    case "$task_type" in
        "compilation")
            # CPU-intensive, use all available workers
            echo "$((available_workers < max_workers ? available_workers : max_workers))"
            ;;
        "testing")
            # Memory-intensive, limit workers
            echo "$((available_workers < 4 ? available_workers : 4))"
            ;;
        "packaging")
            # I/O-intensive, moderate workers
            echo "$((available_workers < 3 ? available_workers : 3))"
            ;;
        *)
            echo "$((available_workers < 4 ? available_workers : 4))"
            ;;
    esac
}

# Scale worker pool
scale_pool() {
    local pool_name="$1"
    local desired_workers="$2"
    local current_workers
    current_workers=$(get_available_workers "$pool_name" | wc -w)
    
    if [[ "${AUTO_SCALE_ENABLED:-true}" != "true" ]]; then
        log "DEBUG" "Auto-scaling disabled, skipping pool scaling"
        return 0
    fi
    
    if (( desired_workers > current_workers )); then
        log "INFO" "Scaling up $pool_name pool from $current_workers to $desired_workers workers"
        # In a real implementation, this would start new instances
        # For now, just log the action
    elif (( desired_workers < current_workers )); then
        log "INFO" "Scaling down $pool_name pool from $current_workers to $desired_workers workers"
        # In a real implementation, this would terminate instances
    fi
}

# Monitor all worker pools
monitor_pools() {
    log "INFO" "Starting worker pool monitoring"
    
    while true; do
        local interval="${HEALTH_CHECK_INTERVAL:-30}"
        
        # Check each pool type
        for pool_type in compilation testing packaging default; do
            local workers
            workers=$(get_pool_workers "$pool_type")
            
            log "DEBUG" "Checking pool: $pool_type"
            
            for worker in $workers; do
                local status
                status=$(check_worker_health "$worker")
                
                update_worker_status "$worker" "$status" "$pool_type"
                
                if [[ "$status" == "failed" ]]; then
                    log "WARN" "Worker $worker in pool $pool_type has failed"
                elif [[ "$status" == "overloaded" ]]; then
                    log "WARN" "Worker $worker in pool $pool_type is overloaded"
                fi
            done
        done
        
        sleep "$interval"
    done
}

# Display pool status
show_pool_status() {
    echo "Worker Pool Status"
    echo "================="
    echo ""
    
    for pool_type in compilation testing packaging default; do
        local pool_workers
        pool_workers=$(get_pool_workers "$pool_type")
        
        echo "Pool: $pool_type"
        echo "Workers: $pool_workers"
        echo ""
        
        printf "%-15s %-10s %-15s %s\n" "Worker" "Status" "Pool" "Last Check"
        printf "%-15s %-10s %-15s %s\n" "-------" "------" "----" "----------"
        
        for worker in $pool_workers; do
            local status
            local last_check
            status=$(check_worker_health "$worker")
            last_check=$(date '+%H:%M:%S')
            
            local status_color
            case "$status" in
                "healthy") status_color="$GREEN" ;;
                "overloaded") status_color="$YELLOW" ;;
                "failed") status_color="$RED" ;;
                *) status_color="$NC" ;;
            esac
            
            printf "%-15s ${status_color}%-10s${NC} %-15s %s\n" "$worker" "$status" "$pool_type" "$last_check"
        done
        
        echo ""
    done
}

# Select best workers for a task
select_workers() {
    local task_type="$1"
    local task_count="${2:-1}"
    local workers
    local selected_workers=()
    
    # Get available workers for the task type
    workers=$(get_available_workers "$task_type")
    
    # Select the best workers based on recent performance
    while [[ ${#selected_workers[@]} -lt $task_count ]] && [[ -n "$workers" ]]; do
        # Simple selection - could be enhanced with performance metrics
        local worker=$(echo "$workers" | head -1)
        selected_workers+=("$worker")
        workers=$(echo "$workers" | tail -n +2)
    done
    
    # Return selected workers
    printf '%s\n' "${selected_workers[@]}"
}

# Execute task on worker pool
execute_on_pool() {
    local task_type="$1"
    local gradle_task="$2"
    local workers_needed="${3:-1}"
    
    log "INFO" "Executing $gradle_task on $task_type pool with $workers_needed workers"
    
    # Get optimal workers
    local optimal_count
    optimal_count=$(get_optimal_worker_count "$task_type")
    
    if (( workers_needed > optimal_count )); then
        log "WARN" "Requested $workers_needed workers, but only $optimal_count are optimal for $task_type"
    fi
    
    # Select workers
    local selected_workers
    readarray -t selected_workers < <(select_workers "$task_type" "$workers_needed")
    
    if [[ ${#selected_workers[@]} -eq 0 ]]; then
        log "ERROR" "No available workers in $task_type pool"
        return 1
    fi
    
    log "INFO" "Selected workers: ${selected_workers[*]}"
    
    # Update .gradlebuild_env with selected workers
    local worker_ips
    worker_ips=$(IFS=','; echo "${selected_workers[*]}")
    
    if [[ -f ".gradlebuild_env" ]]; then
        sed -i.bak "s/export WORKER_IPS=.*/export WORKER_IPS=\"${worker_ips//,/ }\"/" .gradlebuild_env
        sed -i "s/export MAX_WORKERS=.*/export MAX_WORKERS=$workers_needed/" .gradlebuild_env
    else
        log "ERROR" ".gradlebuild_env not found"
        return 1
    fi
    
    # Execute the build
    log "INFO" "Starting build: $gradle_task"
    if ./sync_and_build.sh "$gradle_task"; then
        log "INFO" "Build completed successfully"
        return 0
    else
        log "ERROR" "Build failed"
        return 1
    fi
}

# Auto-detect best pool for task
detect_best_pool() {
    local gradle_task="$1"
    
    # Simple detection based on task name
    case "$gradle_task" in
        *"compile"*|*"assemble"*)
            echo "compilation"
            ;;
        *"test"*)
            echo "testing"
            ;;
        *"bundle"*)
            echo "packaging"
            ;;
        *)
            echo "default"
            ;;
    esac
}

# Interactive pool management
interactive_mode() {
    while true; do
        echo ""
        echo "Worker Pool Manager"
        echo "==================="
        echo "1) Show Pool Status"
        echo "2) Execute Task on Pool"
        echo "3) Monitor Pools (continuous)"
        echo "4) Auto-detect Best Pool for Task"
        echo "5) Scale Pool"
        echo "6) Exit"
        echo ""
        read -p "Select option (1-6): " choice
        
        case "$choice" in
            1)
                show_pool_status
                ;;
            2)
                read -p "Enter Gradle task: " gradle_task
                read -p "Enter number of workers (default: auto): " workers
                workers=${workers:-auto}
                
                local pool_type
                pool_type=$(detect_best_pool "$gradle_task")
                
                if [[ "$workers" == "auto" ]]; then
                    workers=$(get_optimal_worker_count "$pool_type")
                fi
                
                execute_on_pool "$pool_type" "$gradle_task" "$workers"
                ;;
            3)
                log "INFO" "Starting continuous monitoring (Ctrl+C to stop)"
                monitor_pools
                ;;
            4)
                read -p "Enter Gradle task: " gradle_task
                local best_pool
                best_pool=$(detect_best_pool "$gradle_task")
                echo "Best pool for '$gradle_task': $best_pool"
                echo "Reason: ${POOL_REQUIREMENTS[$best_pool]}"
                ;;
            5)
                read -p "Enter pool name: " pool_name
                read -p "Enter desired worker count: " desired_count
                scale_pool "$pool_name" "$desired_count"
                ;;
            6)
                break
                ;;
            *)
                log "ERROR" "Invalid option"
                ;;
        esac
    done
}

# Main function
main() {
    # Check dependencies
    if ! command -v bc >/dev/null 2>&1; then
        log "ERROR" "bc calculator is required but not installed"
        exit 1
    fi
    
    # Initialize
    init_directories
    load_config
    
    # Command line handling
    case "${1:-interactive}" in
        "status")
            show_pool_status
            ;;
        "execute")
            if [[ $# -lt 2 ]]; then
                log "ERROR" "Usage: $0 execute <gradle_task> [workers] [pool_type]"
                exit 1
            fi
            local gradle_task="$2"
            local workers="${3:-auto}"
            local pool_type="${4:-auto}"
            
            if [[ "$pool_type" == "auto" ]]; then
                pool_type=$(detect_best_pool "$gradle_task")
            fi
            
            if [[ "$workers" == "auto" ]]; then
                workers=$(get_optimal_worker_count "$pool_type")
            fi
            
            execute_on_pool "$pool_type" "$gradle_task" "$workers"
            ;;
        "monitor")
            log "INFO" "Starting pool monitoring"
            monitor_pools
            ;;
        "detect")
            if [[ $# -lt 2 ]]; then
                log "ERROR" "Usage: $0 detect <gradle_task>"
                exit 1
            fi
            local pool_type
            pool_type=$(detect_best_pool "$2")
            echo "$pool_type"
            ;;
        "interactive"|"")
            interactive_mode
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [command] [options]"
            echo ""
            echo "Commands:"
            echo "  status              - Show pool status"
            echo "  execute <task> [workers] [pool] - Execute task on specific pool"
            echo "  monitor             - Start continuous monitoring"
            echo "  detect <task>       - Detect best pool for task"
            echo "  interactive         - Start interactive mode (default)"
            echo "  help                - Show this help"
            echo ""
            echo "Examples:"
            echo "  $0 execute assembleDebug 4 compilation"
            echo "  $0 detect testDebugUnitTest"
            echo "  $0 status"
            ;;
        *)
            log "ERROR" "Unknown command: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"