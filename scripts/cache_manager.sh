#!/bin/bash

# Distributed Gradle Build Cache Manager
# Manages local and remote build caches with optimization and cleanup

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="${PROJECT_DIR:-$(pwd)}"
CACHE_ROOT="${CACHE_ROOT:-$HOME/.gradle-build-cache}"
LOCAL_CACHE_DIR="${CACHE_ROOT}/local"
REMOTE_CACHE_DIR="${CACHE_ROOT}/remote"
CACHE_CONFIG_FILE="${PROJECT_DIR}/cache_config"

# Default settings
DEFAULT_CACHE_SIZE="10g"
DEFAULT_CACHE_TTL="30d"
DEFAULT_CLEANUP_THRESHOLD="80"
DEFAULT_REMOTE_ENABLED="false"
DEFAULT_REMOTE_URL="http://localhost:8080/cache"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging
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
    
    echo "[$timestamp] [$level] $message" >> "${CACHE_ROOT}/cache_manager.log"
}

# Initialize cache directories
init_cache() {
    mkdir -p "$CACHE_ROOT" "$LOCAL_CACHE_DIR" "$REMOTE_CACHE_DIR"
    touch "$CACHE_CONFIG_FILE"
    
    # Set default configuration
    if [[ ! -s "$CACHE_CONFIG_FILE" ]]; then
        cat > "$CACHE_CONFIG_FILE" << EOF
# Cache Manager Configuration
CACHE_SIZE="$DEFAULT_CACHE_SIZE"
CACHE_TTL="$DEFAULT_CACHE_TTL"
CLEANUP_THRESHOLD="$DEFAULT_CLEANUP_THRESHOLD"
REMOTE_ENABLED="$DEFAULT_REMOTE_ENABLED"
REMOTE_URL="$DEFAULT_REMOTE_URL"
CACHE_ENABLED="true"
EOF
        log "INFO" "Created default cache configuration"
    fi
    
    # Load configuration
    load_config
}

# Load configuration
load_config() {
    if [[ -f "$CACHE_CONFIG_FILE" ]]; then
        source "$CACHE_CONFIG_FILE"
        log "DEBUG" "Loaded cache configuration from $CACHE_CONFIG_FILE"
    else
        log "ERROR" "Cache configuration file not found: $CACHE_CONFIG_FILE"
        return 1
    fi
}

# Save configuration
save_config() {
    cat > "$CACHE_CONFIG_FILE" << EOF
# Cache Manager Configuration
CACHE_SIZE="$CACHE_SIZE"
CACHE_TTL="$CACHE_TTL"
CLEANUP_THRESHOLD="$CLEANUP_THRESHOLD"
REMOTE_ENABLED="$REMOTE_ENABLED"
REMOTE_URL="$REMOTE_URL"
CACHE_ENABLED="$CACHE_ENABLED"
EOF
    log "DEBUG" "Saved cache configuration"
}

# Get cache size in bytes
get_cache_size() {
    local cache_path="$1"
    du -sb "$cache_path" 2>/dev/null | cut -f1 || echo "0"
}

# Format bytes to human readable
format_bytes() {
    local bytes="$1"
    local units=('B' 'K' 'M' 'G' 'T')
    local unit=0
    
    while (( bytes > 1024 )) && [[ $unit -lt 4 ]]; do
        bytes=$((bytes / 1024))
        ((unit++))
    done
    
    echo "${bytes}${units[$unit]}"
}

# Check cache health
check_cache_health() {
    local cache_path="$1"
    local cache_name="$2"
    
    if [[ ! -d "$cache_path" ]]; then
        log "WARN" "$cache_name cache directory does not exist"
        return 1
    fi
    
    local cache_size
    cache_size=$(get_cache_size "$cache_path")
    local cache_size_hr
    cache_size_hr=$(format_bytes "$cache_size")
    
    log "DEBUG" "$cache_name cache size: $cache_size_hr"
    
    # Check permissions
    if [[ ! -w "$cache_path" ]]; then
        log "WARN" "$cache_name cache is not writable"
        return 1
    fi
    
    # Check disk space
    local available_space
    available_space=$(df -BG "$cache_path" | awk 'NR==2{print $4}' | tr -d 'G')
    
    if [[ $available_space -lt 1 ]]; then
        log "WARN" "Low disk space near $cache_name cache: ${available_space}G available"
        return 1
    fi
    
    log "INFO" "$cache_name cache is healthy (${cache_size_hr})"
    return 0
}

# Clean local cache
clean_local_cache() {
    log "INFO" "Cleaning local cache in $LOCAL_CACHE_DIR"
    
    local cache_size_before
    cache_size_before=$(get_cache_size "$LOCAL_CACHE_DIR")
    
    # Remove old cache entries based on TTL
    find "$LOCAL_CACHE_DIR" -type f -mtime "+${CACHE_TTL%[a-z]}" -delete
    
    # Remove empty directories
    find "$LOCAL_CACHE_DIR" -type d -empty -delete 2>/dev/null || true
    
    local cache_size_after
    cache_size_after=$(get_cache_size "$LOCAL_CACHE_DIR")
    
    local freed_space
    freed_space=$((cache_size_before - cache_size_after))
    
    log "INFO" "Cleaned local cache: freed $(format_bytes $freed_space)"
}

# Clean remote cache
clean_remote_cache() {
    if [[ "$REMOTE_ENABLED" != "true" ]]; then
        log "DEBUG" "Remote cache disabled, skipping cleanup"
        return 0
    fi
    
    log "INFO" "Cleaning remote cache at $REMOTE_URL"
    
    # Remove old entries (if API supports it)
    local cleanup_url="${REMOTE_URL%/}/cleanup"
    
    if curl -s -X POST "$cleanup_url" >/dev/null 2>&1; then
        log "INFO" "Remote cache cleanup initiated"
    else
        log "WARN" "Remote cache cleanup failed or not supported"
    fi
}

# Sync caches
sync_caches() {
    if [[ "$REMOTE_ENABLED" != "true" ]]; then
        log "DEBUG" "Remote cache disabled, skipping sync"
        return 0
    fi
    
    log "INFO" "Syncing caches"
    
    # Pull changes from remote
    if curl -s "$REMOTE_URL/list" >/dev/null 2>&1; then
        log "DEBUG" "Remote cache is accessible"
        
        # Compare local and remote entries
        # This is a simplified implementation
        local local_entries
        local_entries=$(find "$LOCAL_CACHE_DIR" -type f | wc -l)
        
        log "INFO" "Local cache entries: $local_entries"
        log "INFO" "Sync completed"
    else
        log "WARN" "Remote cache not accessible: $REMOTE_URL"
    fi
}

# Optimize cache
optimize_cache() {
    log "INFO" "Optimizing local cache"
    
    # Check current cache size
    local current_size
    current_size=$(get_cache_size "$LOCAL_CACHE_DIR")
    local max_size_bytes
    max_size_bytes=$(echo "${CACHE_SIZE%[a-z]}" | sed 's/g/*1024*1024*1024/;s/m/*1024*1024/;s/k/*1024/')
    
    # Convert to bytes
    max_size_bytes=$((max_size_bytes))
    
    # If cache is over size limit, clean it
    if [[ $current_size -gt $max_size_bytes ]]; then
        log "INFO" "Cache size exceeded ($(format_bytes $current_size) > $CACHE_SIZE), cleaning..."
        
        # Find and remove least recently used files
        find "$LOCAL_CACHE_DIR" -type f -printf '%T@ %p\n' | \
            sort -n | \
            head -n $(( (current_size - max_size_bytes) / 1024 )) | \
            cut -d' ' -f2- | \
            xargs -r rm
        
        local new_size
        new_size=$(get_cache_size "$LOCAL_CACHE_DIR")
        log "INFO" "Cache optimized: $(format_bytes $new_size)"
    else
        log "INFO" "Cache size is within limits: $(format_bytes $current_size)"
    fi
    
    # Compress cache if beneficial
    if command -v pigz >/dev/null 2>&1; then
        log "DEBUG" "Cache compression available (pigz)"
        # Could implement cache compression here
    fi
}

# Start cache server
start_cache_server() {
    if [[ "$REMOTE_ENABLED" != "true" ]]; then
        log "INFO" "Enabling remote cache server"
        REMOTE_ENABLED="true"
        save_config
    fi
    
    # Determine cache directory to serve
    local serve_dir="${LOCAL_CACHE_DIR}"
    local port="${REMOTE_URL##*:}"
    port="${port%/cache}"
    
    # Kill existing cache server
    pkill -f "python.*http.server.*$port" || true
    
    log "INFO" "Starting cache server on port $port serving $serve_dir"
    
    # Start Python HTTP server
    cd "$serve_dir"
    nohup python3 -m http.server "$port" > "${CACHE_ROOT}/cache_server.log" 2>&1 &
    
    local server_pid=$!
    echo "$server_pid" > "${CACHE_ROOT}/cache_server.pid"
    
    # Wait for server to start
    sleep 2
    
    if kill -0 "$server_pid" 2>/dev/null; then
        log "INFO" "Cache server started (PID: $server_pid, URL: $REMOTE_URL)"
        return 0
    else
        log "ERROR" "Failed to start cache server"
        return 1
    fi
}

# Stop cache server
stop_cache_server() {
    local pid_file="${CACHE_ROOT}/cache_server.pid"
    
    if [[ -f "$pid_file" ]]; then
        local server_pid
        server_pid=$(cat "$pid_file")
        
        if kill -0 "$server_pid" 2>/dev/null; then
            log "INFO" "Stopping cache server (PID: $server_pid)"
            kill "$server_pid"
            
            # Wait for graceful shutdown
            sleep 2
            
            # Force kill if still running
            if kill -0 "$server_pid" 2>/dev/null; then
                kill -9 "$server_pid"
            fi
            
            rm -f "$pid_file"
            log "INFO" "Cache server stopped"
        else
            log "WARN" "Cache server PID file exists but process not running"
            rm -f "$pid_file"
        fi
    else
        log "INFO" "Cache server not running"
    fi
}

# Check cache server status
check_cache_server_status() {
    local pid_file="${CACHE_ROOT}/cache_server.pid"
    
    if [[ -f "$pid_file" ]]; then
        local server_pid
        server_pid=$(cat "$pid_file")
        
        if kill -0 "$server_pid" 2>/dev/null; then
            log "INFO" "Cache server running (PID: $server_pid, URL: $REMOTE_URL)"
            
            # Test connectivity
            if curl -s "$REMOTE_URL/health" >/dev/null 2>&1; then
                log "INFO" "Cache server is responding"
                return 0
            else
                log "WARN" "Cache server running but not responding"
                return 1
            fi
        else
            log "INFO" "Cache server not running"
            return 1
        fi
    else
        log "INFO" "Cache server not running"
        return 1
    fi
}

# Configure Gradle for cache
configure_gradle_cache() {
    if [[ "$CACHE_ENABLED" != "true" ]]; then
        log "INFO" "Cache is disabled in configuration"
        return 0
    fi
    
    log "INFO" "Configuring Gradle for distributed caching"
    
    local gradle_props="gradle.properties"
    
    # Backup existing file
    if [[ -f "$gradle_props" ]]; then
        cp "$gradle_props" "${gradle_props}.backup.$(date +%s)"
    fi
    
    # Create or update gradle.properties
    {
        echo "# Distributed Gradle Build Cache Configuration"
        echo "org.gradle.caching=true"
        echo "org.gradle.cache.fixed-size=true"
        echo "org.gradle.cache.max-size=$CACHE_SIZE"
        
        if [[ "$REMOTE_ENABLED" == "true" ]]; then
            echo ""
            echo "# Remote cache configuration"
            echo "systemProp.org.gradle.caching.remote.enabled=true"
            echo "systemProp.org.gradle.caching.remote.url=$REMOTE_URL"
            echo "systemProp.org.gradle.caching.remote.push=true"
            echo "systemProp.org.gradle.caching.remote.allow-insecure=true"
        fi
        
        # Add any existing properties (excluding our managed ones)
        if [[ -f "${gradle_props}.backup.$(date +%s)" ]]; then
            grep -v -E "(org\.gradle\.caching|systemProp\.org\.gradle\.caching)" \
                "${gradle_props}.backup.$(date +%s)" | \
                grep -v '^#' || true
        fi
    } > "$gradle_props"
    
    log "INFO" "Gradle cache configuration updated"
}

# Generate cache statistics
generate_cache_stats() {
    log "INFO" "Generating cache statistics"
    
    echo "Cache Statistics Report"
    echo "====================="
    echo "Generated: $(date)"
    echo ""
    
    # Local cache stats
    if [[ -d "$LOCAL_CACHE_DIR" ]]; then
        local local_size
        local_size=$(get_cache_size "$LOCAL_CACHE_DIR")
        local local_entries
        local_entries=$(find "$LOCAL_CACHE_DIR" -type f | wc -l)
        local local_dirs
        local_dirs=$(find "$LOCAL_CACHE_DIR" -type d | wc -l)
        
        echo "Local Cache:"
        echo "-----------"
        echo "  Directory: $LOCAL_CACHE_DIR"
        echo "  Size: $(format_bytes $local_size)"
        echo "  Files: $local_entries"
        echo "  Directories: $local_dirs"
        echo ""
    fi
    
    # Remote cache status
    echo "Remote Cache:"
    echo "-------------"
    echo "  Enabled: $REMOTE_ENABLED"
    if [[ "$REMOTE_ENABLED" == "true" ]]; then
        echo "  URL: $REMOTE_URL"
        if check_cache_server_status >/dev/null 2>&1; then
            echo "  Status: Running"
        else
            echo "  Status: Stopped"
        fi
    fi
    echo ""
    
    # Configuration
    echo "Configuration:"
    echo "--------------"
    echo "  Cache Size Limit: $CACHE_SIZE"
    echo "  Cache TTL: $CACHE_TTL"
    echo "  Cleanup Threshold: ${CLEANUP_THRESHOLD}%"
    echo "  Cache Enabled: $CACHE_ENABLED"
    echo ""
    
    # Usage statistics
    echo "Usage:"
    echo "-------"
    
    # Calculate hit rates from Gradle logs if available
    local recent_builds
    recent_builds=$(find "$PROJECT_DIR/build/logs" -name "*.log" -mtime -7 2>/dev/null || true)
    
    if [[ -n "$recent_builds" ]]; then
        local total_hits=0
        local total_requests=0
        
        while IFS= read -r log_file; do
            if [[ -f "$log_file" ]]; then
                local hits
                hits=$(grep -o "Cache hit" "$log_file" | wc -l)
                local requests
                requests=$(grep -o -E "(Cache hit|Cache miss)" "$log_file" | wc -l)
                
                total_hits=$((total_hits + hits))
                total_requests=$((total_requests + requests))
            fi
        done <<< "$recent_builds"
        
        if [[ $total_requests -gt 0 ]]; then
            local hit_rate
            hit_rate=$(( total_hits * 100 / total_requests ))
            echo "  Cache Hit Rate (7 days): ${hit_rate}% ($total_hits/$total_requests)"
        else
            echo "  Cache Hit Rate: No recent build data"
        fi
    else
        echo "  Cache Hit Rate: No build logs available"
    fi
}

# Interactive configuration
interactive_config() {
    echo "Cache Configuration"
    echo "=================="
    echo ""
    
    echo "Current configuration:"
    echo "  Cache Size: $CACHE_SIZE"
    echo "  Cache TTL: $CACHE_TTL"
    echo "  Cleanup Threshold: ${CLEANUP_THRESHOLD}%"
    echo "  Remote Enabled: $REMOTE_ENABLED"
    echo "  Remote URL: $REMOTE_URL"
    echo "  Cache Enabled: $CACHE_ENABLED"
    echo ""
    
    read -p "Modify configuration? (y/n): " modify
    
    if [[ "$modify" =~ ^[Yy]$ ]]; then
        read -p "Cache size (e.g., 5g, 500m): [$CACHE_SIZE] " new_size
        CACHE_SIZE="${new_size:-$CACHE_SIZE}"
        
        read -p "Cache TTL (e.g., 30d, 7d): [$CACHE_TTL] " new_ttl
        CACHE_TTL="${new_ttl:-$CACHE_TTL}"
        
        read -p "Cleanup threshold (%): [$CLEANUP_THRESHOLD] " new_threshold
        CLEANUP_THRESHOLD="${new_threshold:-$CLEANUP_THRESHOLD}"
        
        read -p "Enable remote cache? (y/n): [$REMOTE_ENABLED] " new_remote
        if [[ -n "$new_remote" ]]; then
            if [[ "$new_remote" =~ ^[Yy]$ ]]; then
                REMOTE_ENABLED="true"
            else
                REMOTE_ENABLED="false"
            fi
        fi
        
        if [[ "$REMOTE_ENABLED" == "true" ]]; then
            read -p "Remote cache URL: [$REMOTE_URL] " new_url
            REMOTE_URL="${new_url:-$REMOTE_URL}"
        fi
        
        read -p "Enable cache? (y/n): [$CACHE_ENABLED] " new_cache_enabled
        if [[ -n "$new_cache_enabled" ]]; then
            if [[ "$new_cache_enabled" =~ ^[Yy]$ ]]; then
                CACHE_ENABLED="true"
            else
                CACHE_ENABLED="false"
            fi
        fi
        
        save_config
        log "INFO" "Configuration updated"
        
        # Reconfigure Gradle if needed
        configure_gradle_cache
    fi
}

# Main menu
main_menu() {
    while true; do
        echo ""
        echo "Distributed Gradle Build Cache Manager"
        echo "===================================="
        echo "1) Show Cache Status"
        echo "2) Configure Cache"
        echo "3) Start Cache Server"
        echo "4) Stop Cache Server"
        echo "5) Clean Local Cache"
        echo "6) Clean Remote Cache"
        echo "7) Sync Caches"
        echo "8) Optimize Cache"
        echo "9) Configure Gradle"
        echo "10) Generate Statistics"
        echo "11) Exit"
        echo ""
        read -p "Select option (1-11): " choice
        
        case "$choice" in
            1)
                check_cache_health "$LOCAL_CACHE_DIR" "Local"
                if [[ "$REMOTE_ENABLED" == "true" ]]; then
                    check_cache_server_status
                fi
                ;;
            2)
                interactive_config
                ;;
            3)
                start_cache_server
                ;;
            4)
                stop_cache_server
                ;;
            5)
                clean_local_cache
                ;;
            6)
                clean_remote_cache
                ;;
            7)
                sync_caches
                ;;
            8)
                optimize_cache
                ;;
            9)
                configure_gradle_cache
                ;;
            10)
                generate_cache_stats
                ;;
            11)
                break
                ;;
            *)
                log "ERROR" "Invalid option. Please select 1-11."
                ;;
        esac
    done
}

# Main function
main() {
    # Check if in Gradle project
    if [[ ! -f "build.gradle" ]] && [[ ! -f "build.gradle.kts" ]]; then
        log "ERROR" "Not in a Gradle project directory"
        exit 1
    fi
    
    # Initialize cache
    init_cache
    
    # Command line argument handling
    case "${1:-interactive}" in
        "start")
            start_cache_server
            ;;
        "stop")
            stop_cache_server
            ;;
        "status")
            check_cache_health "$LOCAL_CACHE_DIR" "Local"
            check_cache_server_status
            ;;
        "clean")
            clean_local_cache
            if [[ "$REMOTE_ENABLED" == "true" ]]; then
                clean_remote_cache
            fi
            ;;
        "sync")
            sync_caches
            ;;
        "optimize")
            optimize_cache
            ;;
        "configure")
            configure_gradle_cache
            ;;
        "stats")
            generate_cache_stats
            ;;
        "interactive"|"")
            main_menu
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [command]"
            echo ""
            echo "Commands:"
            echo "  start       - Start cache server"
            echo "  stop        - Stop cache server"
            echo "  status      - Show cache status"
            echo "  clean       - Clean caches"
            echo "  sync        - Sync caches"
            echo "  optimize    - Optimize cache"
            echo "  configure   - Configure Gradle for cache"
            echo "  stats       - Generate statistics"
            echo "  interactive - Start interactive mode (default)"
            echo "  help        - Show this help"
            ;;
        *)
            log "ERROR" "Unknown command: $1"
            main_menu
            ;;
    esac
}

# Run main function
main "$@"