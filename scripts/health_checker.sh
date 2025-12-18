#!/bin/bash

# Distributed Gradle Build Health Checker
# Monitors system health and provides automated issue detection

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="${PROJECT_DIR:-$(pwd)}"
HEALTH_LOG="${PROJECT_DIR}/health_check.log"
ALERT_LOG="${PROJECT_DIR}/health_alerts.log"
CONFIG_FILE="${SCRIPT_DIR}/health_config"

# Default thresholds
DEFAULT_CPU_WARNING=80
DEFAULT_CPU_CRITICAL=95
DEFAULT_MEMORY_WARNING=85
DEFAULT_MEMORY_CRITICAL=95
DEFAULT_DISK_WARNING=85
DEFAULT_DISK_CRITICAL=95
DEFAULT_BUILD_TIME_WARNING=1800  # 30 minutes
DEFAULT_FAILURE_RATE_WARNING=20  # 20%

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Health status levels
declare -A STATUS_LEVELS=(
    ["OK"]=0
    ["WARNING"]=1
    ["CRITICAL"]=2
)

# Load configuration
load_config() {
    # Default values
    CPU_WARNING="${DEFAULT_CPU_WARNING}"
    CPU_CRITICAL="${DEFAULT_CPU_CRITICAL}"
    MEMORY_WARNING="${DEFAULT_MEMORY_WARNING}"
    MEMORY_CRITICAL="${DEFAULT_MEMORY_CRITICAL}"
    DISK_WARNING="${DEFAULT_DISK_WARNING}"
    DISK_CRITICAL="${DEFAULT_DISK_CRITICAL}"
    BUILD_TIME_WARNING="${DEFAULT_BUILD_TIME_WARNING}"
    FAILURE_RATE_WARNING="${DEFAULT_FAILURE_RATE_WARNING}"
    
    # Load from config file if exists
    if [[ -f "$CONFIG_FILE" ]]; then
        source "$CONFIG_FILE"
    fi
}

# Logging functions
log_health() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    echo "[$timestamp] [$level] $message" >> "$HEALTH_LOG"
}

log_alert() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    echo "[$timestamp] [$level] $message" >> "$ALERT_LOG"
    
    # Color-coded console output
    case "$level" in
        "CRITICAL") echo -e "${RED}[CRITICAL]${NC} $message" ;;
        "WARNING")  echo -e "${YELLOW}[WARNING]${NC} $message" ;;
        "INFO")     echo -e "${GREEN}[INFO]${NC} $message" ;;
        *)          echo "[$level] $message" ;;
    esac
}

# Get system CPU usage
get_cpu_usage() {
    top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1
}

# Get system memory usage
get_memory_usage() {
    free | grep '^Mem:' | awk '{printf "%.1f", $3/$2 * 100.0}'
}

# Get disk usage
get_disk_usage() {
    df "$PROJECT_DIR" | awk 'NR==2{printf "%.0f", $5}'
}

# Check SSH connectivity to workers
check_worker_connectivity() {
    local failed_workers=()
    
    if [[ -f ".gradlebuild_env" ]]; then
        source ".gradlebuild_env"
        
        if [[ -n "${WORKER_IPS:-}" ]]; then
            for ip in $WORKER_IPS; do
                if ! timeout 10 ssh -o ConnectTimeout=5 -o BatchMode=yes "$ip" "echo 'OK'" >/dev/null 2>&1; then
                    failed_workers+=("$ip")
                fi
            done
        fi
    fi
    
    echo "${failed_workers[@]}"
}

# Get worker resource usage
get_worker_metrics() {
    local ip="$1"
    local metrics="{}"
    
    if timeout 10 ssh -o ConnectTimeout=5 -o BatchMode=yes "$ip" "echo 'OK'" >/dev/null 2>&1; then
        local cpu_usage mem_usage disk_usage
        
        cpu_usage=$(timeout 10 ssh -o ConnectTimeout=5 -o BatchMode=yes "$ip" \
            "top -bn1 | grep 'Cpu(s)' | awk '{print \$2}' | cut -d'%' -f1" 2>/dev/null || echo "0")
        
        mem_usage=$(timeout 10 ssh -o ConnectTimeout=5 -o BatchMode=yes "$ip" \
            "free | grep '^Mem:' | awk '{printf \"%.1f\", \$3/\$2 * 100.0}'" 2>/dev/null || echo "0")
        
        disk_usage=$(timeout 10 ssh -o ConnectTimeout=5 -o BatchMode=yes "$ip" \
            "df / | awk 'NR==2{printf \"%.0f\", \$5}'" 2>/dev/null || echo "0")
        
        metrics="{\"cpu\": $cpu_usage, \"memory\": $mem_usage, \"disk\": $disk_usage}"
    fi
    
    echo "$metrics"
}

# Check build performance
check_build_performance() {
    local recent_builds=()
    local failed_builds=0
    local total_builds=0
    local max_build_time=0
    local total_build_time=0
    
    # Analyze recent build logs if available
    if [[ -d "$PROJECT_DIR/build/logs" ]]; then
        while IFS= read -r -d '' log_file; do
            if [[ -f "$log_file" ]]; then
                local build_time duration
                build_time=$(stat -c %Y "$log_file" 2>/dev/null || echo "0")
                
                # Check if log is from last 24 hours
                local current_time
                current_time=$(date +%s)
                if [[ $((current_time - build_time)) -lt 86400 ]]; then
                    local build_output
                    build_output=$(cat "$log_file" 2>/dev/null || echo "")
                    
                    # Check if build failed
                    if echo "$build_output" | grep -qi "failed\|error\|exception"; then
                        ((failed_builds++))
                    fi
                    
                    # Extract build duration
                    duration=$(echo "$build_output" | grep -o -E "BUILD (SUCCESSFUL|FAILED) in [0-9]+([.,][0-9]+)?" | \
                              grep -o -E "[0-9]+([.,][0-9]+)?" | tr ',' '.' | head -1 || echo "0")
                    
                    if [[ "$duration" != "0" ]]; then
                        local duration_int
                        duration_int=$(printf "%.0f" "$duration" 2>/dev/null || echo "0")
                        total_build_time=$((total_build_time + duration_int))
                        
                        if [[ $duration_int -gt $max_build_time ]]; then
                            max_build_time=$duration_int
                        fi
                    fi
                    
                    ((total_builds++))
                fi
            fi
        done < <(find "$PROJECT_DIR/build/logs" -name "*.log" -mtime -1 -print0 2>/dev/null || true)
    fi
    
    echo "{\"total\": $total_builds, \"failed\": $failed_builds, \"max_time\": $max_build_time, \"total_time\": $total_build_time}"
}

# Check cache performance
check_cache_performance() {
    local cache_hit_rate=0
    local cache_size=0
    
    if [[ -d "$HOME/.gradle/caches" ]]; then
        # Calculate cache size
        cache_size=$(du -sb "$HOME/.gradle/caches" 2>/dev/null | cut -f1 || echo "0")
        
        # Try to get hit rate from recent builds
        local cache_hits=0
        local cache_misses=0
        
        if [[ -d "$PROJECT_DIR/build/logs" ]]; then
            while IFS= read -r -d '' log_file; do
                if [[ -f "$log_file" ]]; then
                    local hits misses
                    hits=$(grep -o "Cache hit" "$log_file" | wc -l)
                    misses=$(grep -o "Cache miss" "$log_file" | wc -l)
                    
                    cache_hits=$((cache_hits + hits))
                    cache_misses=$((cache_misses + misses))
                fi
            done < <(find "$PROJECT_DIR/build/logs" -name "*.log" -mtime -7 -print0 2>/dev/null || true)
        fi
        
        if [[ $((cache_hits + cache_misses)) -gt 0 ]]; then
            cache_hit_rate=$(( cache_hits * 100 / (cache_hits + cache_misses) ))
        fi
    fi
    
    echo "{\"hit_rate\": $cache_hit_rate, \"size_bytes\": $cache_size}"
}

# Perform comprehensive health check
perform_health_check() {
    log_health "INFO" "Starting comprehensive health check"
    
    local overall_status="OK"
    local issues=()
    
    # System health checks
    echo "=== System Health ==="
    
    # CPU check
    local cpu_usage
    cpu_usage=$(get_cpu_usage)
    echo "CPU Usage: ${cpu_usage}%"
    
    if (( $(echo "$cpu_usage > $CPU_CRITICAL" | bc -l) )); then
        issues+=("CRITICAL: CPU usage critical at ${cpu_usage}% (threshold: ${CPU_CRITICAL}%)")
        overall_status="CRITICAL"
    elif (( $(echo "$cpu_usage > $CPU_WARNING" | bc -l) )); then
        issues+=("WARNING: CPU usage high at ${cpu_usage}% (threshold: ${CPU_WARNING}%)")
        if [[ "$overall_status" != "CRITICAL" ]]; then
            overall_status="WARNING"
        fi
    fi
    
    # Memory check
    local memory_usage
    memory_usage=$(get_memory_usage)
    echo "Memory Usage: ${memory_usage}%"
    
    if (( $(echo "$memory_usage > $MEMORY_CRITICAL" | bc -l) )); then
        issues+=("CRITICAL: Memory usage critical at ${memory_usage}% (threshold: ${MEMORY_CRITICAL}%)")
        overall_status="CRITICAL"
    elif (( $(echo "$memory_usage > $MEMORY_WARNING" | bc -l) )); then
        issues+=("WARNING: Memory usage high at ${memory_usage}% (threshold: ${MEMORY_WARNING}%)")
        if [[ "$overall_status" != "CRITICAL" ]]; then
            overall_status="WARNING"
        fi
    fi
    
    # Disk check
    local disk_usage
    disk_usage=$(get_disk_usage)
    echo "Disk Usage: ${disk_usage}%"
    
    if [[ $disk_usage -gt $DISK_CRITICAL ]]; then
        issues+=("CRITICAL: Disk usage critical at ${disk_usage}% (threshold: ${DISK_CRITICAL}%)")
        overall_status="CRITICAL"
    elif [[ $disk_usage -gt $DISK_WARNING ]]; then
        issues+=("WARNING: Disk usage high at ${disk_usage}% (threshold: ${DISK_WARNING}%)")
        if [[ "$overall_status" != "CRITICAL" ]]; then
            overall_status="WARNING"
        fi
    fi
    
    echo ""
    
    # Worker connectivity checks
    echo "=== Worker Connectivity ==="
    local failed_workers
    failed_workers=($(check_worker_connectivity))
    
    if [[ ${#failed_workers[@]} -eq 0 ]]; then
        echo "All workers are reachable"
    else
        echo "Failed workers: ${failed_workers[*]}"
        issues+=("CRITICAL: ${#failed_workers[@]} workers unreachable: ${failed_workers[*]}")
        overall_status="CRITICAL"
    fi
    
    echo ""
    
    # Worker resource checks
    echo "=== Worker Resources ==="
    if [[ -f ".gradlebuild_env" ]]; then
        source ".gradlebuild_env"
        
        if [[ -n "${WORKER_IPS:-}" ]]; then
            for ip in $WORKER_IPS; do
                local metrics
                metrics=$(get_worker_metrics "$ip")
                local cpu mem disk
                
                # Parse JSON-like metrics
                cpu=$(echo "$metrics" | grep -o '"cpu": [0-9.]*' | cut -d: -f2 | tr -d ' ')
                mem=$(echo "$metrics" | grep -o '"memory": [0-9.]*' | cut -d: -f2 | tr -d ' ')
                disk=$(echo "$metrics" | grep -o '"disk": [0-9]*' | cut -d: -f2 | tr -d ' ')
                
                echo "Worker $ip: CPU=${cpu}%, Memory=${mem}%, Disk=${disk}%"
                
                # Check thresholds
                if (( $(echo "$cpu > $CPU_CRITICAL" | bc -l) )) || \
                   (( $(echo "$mem > $MEMORY_CRITICAL" | bc -l) )) || \
                   [[ $disk -gt $DISK_CRITICAL ]]; then
                    issues+=("CRITICAL: Worker $ip overloaded: CPU=${cpu}%, Memory=${mem}%, Disk=${disk}%")
                    if [[ "$overall_status" != "CRITICAL" ]]; then
                        overall_status="CRITICAL"
                    fi
                elif (( $(echo "$cpu > $CPU_WARNING" | bc -l) )) || \
                     (( $(echo "$mem > $MEMORY_WARNING" | bc -l) )) || \
                     [[ $disk -gt $DISK_WARNING ]]; then
                    issues+=("WARNING: Worker $ip stressed: CPU=${cpu}%, Memory=${mem}%, Disk=${disk}%")
                    if [[ "$overall_status" != "CRITICAL" ]]; then
                        overall_status="WARNING"
                    fi
                fi
            done
        fi
    else
        echo "No worker configuration found"
    fi
    
    echo ""
    
    # Build performance checks
    echo "=== Build Performance ==="
    local build_metrics
    build_metrics=$(check_build_performance)
    
    local total_builds failed_builds max_build_time failure_rate
    total_builds=$(echo "$build_metrics" | grep -o '"total": [0-9]*' | cut -d: -f2 | tr -d ' ')
    failed_builds=$(echo "$build_metrics" | grep -o '"failed": [0-9]*' | cut -d: -f2 | tr -d ' ')
    max_build_time=$(echo "$build_metrics" | grep -o '"max_time": [0-9]*' | cut -d: -f2 | tr -d ' ')
    
    if [[ $total_builds -gt 0 ]]; then
        failure_rate=$(( failed_builds * 100 / total_builds ))
        echo "Recent builds: $total_builds total, $failed_builds failed ($failure_rate%)"
        echo "Max build time: ${max_build_time}s"
        
        if [[ $failure_rate -gt $FAILURE_RATE_WARNING ]]; then
            issues+=("CRITICAL: High build failure rate: ${failure_rate}% (threshold: ${FAILURE_RATE_WARNING}%)")
            overall_status="CRITICAL"
        fi
        
        if [[ $max_build_time -gt $BUILD_TIME_WARNING ]]; then
            issues+=("WARNING: Slow build detected: ${max_build_time}s (threshold: ${BUILD_TIME_WARNING}s)")
            if [[ "$overall_status" != "CRITICAL" ]]; then
                overall_status="WARNING"
            fi
        fi
    else
        echo "No recent builds found"
    fi
    
    echo ""
    
    # Cache performance checks
    echo "=== Cache Performance ==="
    local cache_metrics
    cache_metrics=$(check_cache_performance)
    
    local cache_hit_rate cache_size
    cache_hit_rate=$(echo "$cache_metrics" | grep -o '"hit_rate": [0-9]*' | cut -d: -f2 | tr -d ' ')
    cache_size=$(echo "$cache_metrics" | grep -o '"size_bytes": [0-9]*' | cut -d: -f2 | tr -d ' ')
    
    echo "Cache hit rate: ${cache_hit_rate}%"
    echo "Cache size: $(echo "$cache_size" | awk '{print int($1/1024/1024) " MB"}')"
    
    echo ""
    
    # Summary
    echo "=== Health Summary ==="
    echo "Overall Status: $overall_status"
    
    # Log and display issues
    if [[ ${#issues[@]} -gt 0 ]]; then
        echo ""
        echo "Issues detected:"
        for issue in "${issues[@]}"; do
            log_alert "${issue%%:*}" "${issue#*: }"
        done
    else
        echo "All systems healthy"
        log_alert "INFO" "All systems healthy"
    fi
    
    # Log to health log
    log_health "INFO" "Health check completed with status: $overall_status"
    for issue in "${issues[@]}"; do
        log_health "${issue%%:*}" "${issue#*: }"
    done
    
    echo ""
    echo "Health check completed at $(date)"
    
    return "${STATUS_LEVELS[$overall_status]}"
}

# Continuous monitoring
continuous_monitoring() {
    local interval="${1:-300}"  # Default 5 minutes
    
    log_health "INFO" "Starting continuous monitoring with ${interval}s interval"
    
    while true; do
        perform_health_check
        sleep "$interval"
    done
}

# Generate health report
generate_health_report() {
    local report_file="${PROJECT_DIR}/health_report_$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "Distributed Gradle Build Health Report"
        echo "=================================="
        echo "Generated: $(date)"
        echo "Project: $PROJECT_DIR"
        echo ""
        
        echo "=== Current Status ==="
        local current_status
        perform_health_check >/dev/null 2>&1
        current_status=$?
        
        case $current_status in
            0) echo "Overall Status: HEALTHY" ;;
            1) echo "Overall Status: WARNING" ;;
            2) echo "Overall Status: CRITICAL" ;;
            *) echo "Overall Status: UNKNOWN" ;;
        esac
        
        echo ""
        echo "=== System Information ==="
        echo "OS: $(uname -a)"
        echo "Java: $(java -version 2>&1 | head -1)"
        echo "Gradle: $(./gradlew --version 2>/dev/null | grep Gradle || echo 'Not available')"
        echo ""
        
        echo "=== Recent Health Issues ==="
        if [[ -f "$ALERT_LOG" ]]; then
            tail -50 "$ALERT_LOG" | grep -E "(WARNING|CRITICAL)" || echo "No recent issues"
        else
            echo "No alert log found"
        fi
        
        echo ""
        echo "=== Build History ==="
        if [[ -d "$PROJECT_DIR/build/logs" ]]; then
            echo "Recent builds:"
            find "$PROJECT_DIR/build/logs" -name "*.log" -mtime -7 -exec ls -la {} \; || echo "No recent build logs"
        else
            echo "No build logs directory found"
        fi
        
    } > "$report_file"
    
    echo "Health report generated: $report_file"
}

# Interactive mode
interactive_mode() {
    while true; do
        echo ""
        echo "Distributed Gradle Build Health Checker"
        echo "====================================="
        echo "1) Perform Health Check"
        echo "2) Start Continuous Monitoring"
        echo "3) Generate Health Report"
        echo "4) View Recent Issues"
        echo "5) Configure Thresholds"
        echo "6) Exit"
        echo ""
        read -p "Select option (1-6): " choice
        
        case "$choice" in
            1)
                perform_health_check
                ;;
            2)
                read -p "Monitoring interval in seconds (default: 300): " interval
                continuous_monitoring "${interval:-300}"
                ;;
            3)
                generate_health_report
                ;;
            4)
                if [[ -f "$ALERT_LOG" ]]; then
                    echo "Recent issues:"
                    tail -20 "$ALERT_LOG"
                else
                    echo "No alert log found"
                fi
                ;;
            5)
                echo "Current thresholds:"
                echo "CPU Warning: ${CPU_WARNING}%"
                echo "CPU Critical: ${CPU_CRITICAL}%"
                echo "Memory Warning: ${MEMORY_WARNING}%"
                echo "Memory Critical: ${MEMORY_CRITICAL}%"
                echo "Disk Warning: ${DISK_WARNING}%"
                echo "Disk Critical: ${DISK_CRITICAL}%"
                echo "Build Time Warning: ${BUILD_TIME_WARNING}s"
                echo "Failure Rate Warning: ${FAILURE_RATE_WARNING}%"
                echo ""
                read -p "Modify thresholds? (y/n): " modify
                
                if [[ "$modify" =~ ^[Yy]$ ]]; then
                    read -p "CPU Warning %: [$CPU_WARNING] " new_val
                    CPU_WARNING="${new_val:-$CPU_WARNING}"
                    read -p "CPU Critical %: [$CPU_CRITICAL] " new_val
                    CPU_CRITICAL="${new_val:-$CPU_CRITICAL}"
                    read -p "Memory Warning %: [$MEMORY_WARNING] " new_val
                    MEMORY_WARNING="${new_val:-$MEMORY_WARNING}"
                    read -p "Memory Critical %: [$MEMORY_CRITICAL] " new_val
                    MEMORY_CRITICAL="${new_val:-$MEMORY_CRITICAL}"
                    read -p "Disk Warning %: [$DISK_WARNING] " new_val
                    DISK_WARNING="${new_val:-$DISK_WARNING}"
                    read -p "Disk Critical %: [$DISK_CRITICAL] " new_val
                    DISK_CRITICAL="${new_val:-$DISK_CRITICAL}"
                    read -p "Build Time Warning (s): [$BUILD_TIME_WARNING] " new_val
                    BUILD_TIME_WARNING="${new_val:-$BUILD_TIME_WARNING}"
                    read -p "Failure Rate Warning %: [$FAILURE_RATE_WARNING] " new_val
                    FAILURE_RATE_WARNING="${new_val:-$FAILURE_RATE_WARNING}"
                    
                    # Save config
                    cat > "$CONFIG_FILE" << EOF
CPU_WARNING="$CPU_WARNING"
CPU_CRITICAL="$CPU_CRITICAL"
MEMORY_WARNING="$MEMORY_WARNING"
MEMORY_CRITICAL="$MEMORY_CRITICAL"
DISK_WARNING="$DISK_WARNING"
DISK_CRITICAL="$DISK_CRITICAL"
BUILD_TIME_WARNING="$BUILD_TIME_WARNING"
FAILURE_RATE_WARNING="$FAILURE_RATE_WARNING"
EOF
                    echo "Configuration saved"
                fi
                ;;
            6)
                break
                ;;
            *)
                echo "Invalid option. Please select 1-6."
                ;;
        esac
    done
}

# Main function
main() {
    # Check dependencies
    if ! command -v bc >/dev/null 2>&1; then
        echo "ERROR: bc calculator is required but not installed"
        exit 1
    fi
    
    # Load configuration
    load_config
    
    # Command line handling
    case "${1:-check}" in
        "check")
            perform_health_check
            ;;
        "monitor")
            local interval="${2:-300}"
            continuous_monitoring "$interval"
            ;;
        "report")
            generate_health_report
            ;;
        "interactive")
            interactive_mode
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [command] [options]"
            echo ""
            echo "Commands:"
            echo "  check                  - Perform one-time health check (default)"
            echo "  monitor [interval]     - Start continuous monitoring (default: 300s)"
            echo "  report                 - Generate health report"
            echo "  interactive            - Start interactive mode"
            echo "  help                   - Show this help"
            echo ""
            echo "Examples:"
            echo "  $0 check"
            echo "  $0 monitor 60"
            echo "  $0 report"
            ;;
        *)
            echo "Unknown command: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"