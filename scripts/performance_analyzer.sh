#!/bin/bash

# Distributed Gradle Build Performance Analyzer
# Analyzes build performance metrics and provides recommendations

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="${PROJECT_DIR:-$(pwd)}"
RESULTS_DIR="${PROJECT_DIR}/performance_results"
METRICS_FILE="${RESULTS_DIR}/metrics.csv"
ANALYSIS_FILE="${RESULTS_DIR}/analysis.txt"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Initialize results directory
mkdir -p "$RESULTS_DIR"

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
    
    echo "[$timestamp] [$level] $message" >> "$RESULTS_DIR/analyzer.log"
}

# Initialize metrics CSV
initialize_metrics() {
    if [[ ! -f "$METRICS_FILE" ]]; then
        echo "timestamp,build_duration,worker_count,cache_enabled,tasks_executed,tasks_cached,cpu_usage,memory_usage,network_io,disk_io,exit_code" > "$METRICS_FILE"
        log "INFO" "Created metrics file: $METRICS_FILE"
    fi
}

# Collect system metrics
collect_system_metrics() {
    local cpu_usage mem_usage disk_io network_io
    
    # CPU usage
    cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)
    
    # Memory usage
    mem_usage=$(free | grep '^Mem:' | awk '{printf "%.1f", $3/$2 * 100.0}')
    
    # Disk I/O
    disk_io=$(iostat -d 1 2 2>/dev/null | tail -n 1 | awk '{print $3+$4}' || echo "0")
    
    # Network I/O (simplified)
    network_io=$(cat /proc/net/dev | grep -E "(eth|en)" | head -1 | awk '{print $2+$10}' || echo "0")
    
    echo "$cpu_usage,$mem_usage,$network_io,$disk_io"
}

# Parse Gradle build output
parse_gradle_output() {
    local build_output="$1"
    local tasks_executed tasks_up_to_date tasks_from_cache
    
    # Extract task information
    if echo "$build_output" | grep -q "actionable tasks:"; then
        local task_info=$(echo "$build_output" | grep "actionable tasks:" | tail -1)
        tasks_executed=$(echo "$task_info" | sed -E 's/.* ([0-9]+) executed.*/\1/')
        tasks_up_to_date=$(echo "$task_info" | sed -E 's/.* ([0-9]+) up-to-date.*/\1/')
        tasks_from_cache=$(echo "$task_info" | sed -E 's/.* ([0-9]+) from cache.*/\1/')
    else
        tasks_executed="0"
        tasks_up_to_date="0"
        tasks_from_cache="0"
    fi
    
    echo "$tasks_executed,$tasks_up_to_date,$tasks_from_cache"
}

# Run single build experiment
run_build_experiment() {
    local worker_count="$1"
    local cache_enabled="$2"
    local build_task="${3:-assemble}"
    
    log "INFO" "Running experiment: workers=$worker_count, cache=$cache_enabled, task=$build_task"
    
    # Update environment
    if [[ -f ".gradlebuild_env" ]]; then
        sed -i.bak "s/export MAX_WORKERS=.*/export MAX_WORKERS=$worker_count/" .gradlebuild_env
        
        # Update gradle.properties for cache
        if [[ "$cache_enabled" == "true" ]]; then
            grep -q "^org.gradle.caching=true" gradle.properties || \
                echo "org.gradle.caching=true" >> gradle.properties
        else
            sed -i '/^org.gradle.caching=/d' gradle.properties
        fi
    else
        log "ERROR" ".gradlebuild_env not found. Please run setup_master.sh first."
        return 1
    fi
    
    # Clear caches if needed
    if [[ "$cache_enabled" == "false" ]]; then
        log "INFO" "Clearing caches for clean test"
        ./gradlew clean >/dev/null 2>&1 || true
    fi
    
    # Collect pre-build metrics
    local pre_metrics=$(collect_system_metrics)
    
    # Run build with timing
    local build_start=$(date +%s.%N)
    local build_output
    local exit_code
    
    # Capture build output and timing
    build_output=$(timeout 3600 ./sync_and_build.sh "$build_task" 2>&1)
    exit_code=$?
    local build_end=$(date +%s.%N)
    
    # Calculate duration
    local build_duration=$(echo "$build_end - $build_start" | bc)
    
    # Parse build output
    local task_metrics=$(parse_gradle_output "$build_output")
    
    # Collect post-build metrics
    local post_metrics=$(collect_system_metrics)
    
    # Save metrics
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local record="$timestamp,$build_duration,$worker_count,$cache_enabled,$task_metrics,$pre_metrics,$exit_code"
    
    echo "$record" >> "$METRICS_FILE"
    
    log "INFO" "Build completed: duration=${build_duration}s, exit_code=$exit_code"
    
    # Save build output for debugging
    if [[ $exit_code -ne 0 ]]; then
        echo "$build_output" > "$RESULTS_DIR/build_error_${timestamp//[: /]/_}.log"
        log "WARN" "Build failed. Output saved to $RESULTS_DIR/"
    fi
    
    return $exit_code
}

# Analyze collected metrics
analyze_metrics() {
    local csv_file="$1"
    
    log "INFO" "Analyzing performance metrics from $csv_file"
    
    if [[ ! -f "$csv_file" || $(wc -l < "$csv_file") -lt 2 ]]; then
        log "ERROR" "No metrics data found. Run some experiments first."
        return 1
    fi
    
    # Create analysis file
    {
        echo "Performance Analysis Report"
        echo "========================="
        echo "Generated: $(date)"
        echo ""
        
        # Basic statistics
        echo "Build Duration Statistics:"
        echo "-------------------------"
        
        awk -F',' 'NR>1 {print $2}' "$csv_file" | \
            awk '{sum+=$1; sumsq+=$1*$1; min=(min==""||$1<min)?$1:min; max=(max==""||$1>max)?$1:max} END {
            if(NR>0) {
                printf "  Average: %.2f seconds\n", sum/NR;
                printf "  Min: %.2f seconds\n", min;
                printf "  Max: %.2f seconds\n", max;
                printf "  Std Dev: %.2f seconds\n", sqrt(sumsq/NR - (sum/NR)^2);
            }
        }'
        
        echo ""
        echo "Worker Count Analysis:"
        echo "---------------------"
        
        # Group by worker count
        awk -F',' 'NR>1 {workers[$3]++; durations[$3]+=$2; count[$3]++} END {
            for(w in workers) {
                printf "  %s workers: %.2f seconds average (%d runs)\n", w, durations[w]/count[w], count[w];
            }
        }' "$csv_file"
        
        echo ""
        echo "Cache Impact:"
        echo "------------"
        
        # Cache vs no cache comparison
        awk -F',' 'NR>1 {
            if($4=="true") {cache_count++; cache_duration+=$2}
            else {no_cache_count++; no_cache_duration+=$2}
        } END {
            if(cache_count>0 && no_cache_count>0) {
                printf "  With cache: %.2f seconds (%d runs)\n", cache_duration/cache_count, cache_count;
                printf "  Without cache: %.2f seconds (%d runs)\n", no_cache_duration/no_cache_count, no_cache_count;
                printf "  Cache speedup: %.2fx\n", no_cache_duration/no_cache_count / (cache_duration/cache_count);
            } else {
                printf "  Insufficient data for cache comparison\n";
            }
        }' "$csv_file"
        
        echo ""
        echo "Success Rate:"
        echo "------------"
        
        # Success rate analysis
        awk -F',' 'NR>1 {total++; if($11=="0") success++} END {
            if(total>0) {
                printf "  Success rate: %.1f%% (%d/%d)\n", (success/total)*100, success, total;
                printf "  Failure rate: %.1f%% (%d/%d)\n", ((total-success)/total)*100, total-success, total;
            }
        }' "$csv_file"
        
        echo ""
        echo "System Resource Utilization:"
        echo "-------------------------"
        
        # Average resource usage
        awk -F',' 'NR>1 {
            cpu+=$7; mem+=$8; network+=$9; disk+=$10; count++
        } END {
            if(count>0) {
                printf "  Average CPU usage: %.1f%%\n", cpu/count;
                printf "  Average Memory usage: %.1f%%\n", mem/count;
                printf "  Average Network I/O: %.0f bytes\n", network/count;
                printf "  Average Disk I/O: %.0f operations\n", disk/count;
            }
        }' "$csv_file"
        
        echo ""
        echo "Optimal Configuration:"
        echo "--------------------"
        
        # Find fastest configuration
        awk -F',' 'NR>1 {
            if($11=="0" && ($2<fastest_time || fastest_time=="")) {
                fastest_time=$2;
                fastest_workers=$3;
                fastest_cache=$4;
            }
        } END {
            if(fastest_time!="") {
                printf "  Fastest build: %.2f seconds\n", fastest_time;
                printf("  Workers: %s\n", fastest_workers);
                printf("  Cache enabled: %s\n", fastest_cache);
            } else {
                printf("  No successful builds found\n");
            }
        }' "$csv_file"
        
        echo ""
        echo "Recommendations:"
        echo "---------------"
        
        # Generate recommendations based on data
        awk -F',' 'NR>1 {
            if($4=="true") {cache_count++; cache_duration+=$2}
            else {no_cache_count++; no_cache_duration+=$2}
        } END {
            if(cache_count>0 && no_cache_count>0) {
                cache_improvement = (no_cache_duration/no_cache_count) / (cache_duration/cache_count);
                if(cache_improvement > 1.5) {
                    printf("  • Cache provides %.1fx speedup - keep it enabled\n", cache_improvement);
                }
            }
        }' "$csv_file"
        
        # Worker scaling recommendation
        awk -F',' 'NR>1 {
            workers[$3]++; durations[$3]+=$2; count[$3]++
        } END {
            best_workers=""; best_time=999999;
            for(w in workers) {
                avg_time = durations[w]/count[w];
                if(avg_time < best_time) {
                    best_time = avg_time;
                    best_workers = w;
                }
            }
            if(best_workers!="") {
                printf("  • Optimal worker count: %s (avg: %.2fs)\n", best_workers, best_time);
            }
        }' "$csv_file"
        
        # Failure analysis
        awk -F',' 'NR>1 && $11!="0" {failures++} END {
            total=NR-1;
            if(total>0 && failures>0) {
                failure_rate = failures/total * 100;
                if(failure_rate > 20) {
                    printf("  • High failure rate (%.1f%%) - investigate common causes\n", failure_rate);
                }
            }
        }' "$csv_file"
        
    } > "$ANALYSIS_FILE"
    
    log "INFO" "Analysis completed. Report saved to: $ANALYSIS_FILE"
    
    # Display summary to console
    echo ""
    log "INFO" "=== Performance Summary ==="
    
    # Show key metrics
    awk -F',' 'NR>1 {sum+=$2; count++} END {
        if(count>0) printf "Average build time: %.2f seconds (%d builds)\n", sum/count, count;
    }' "$csv_file"
    
    awk -F',' 'NR>1 {total++; if($11=="0") success++} END {
        if(total>0) printf "Success rate: %.1f%% (%d/%d)\n", (success/total)*100, success, total;
    }' "$csv_file"
    
    # Show optimal configuration
    awk -F',' 'NR>1 {
        if($11=="0" && ($2<fastest_time || fastest_time=="")) {
            fastest_time=$2;
            fastest_workers=$3;
            fastest_cache=$4;
        }
    } END {
        if(fastest_time!="") {
            printf "Best configuration: %s workers, cache=%s (%.2fs)\n", fastest_workers, fastest_cache, fastest_time;
        }
    }' "$csv_file"
}

# Generate performance report (ASCII charts)
generate_charts() {
    local csv_file="$1"
    
    log "INFO" "Generating performance charts"
    
    if [[ ! -f "$csv_file" || $(wc -l < "$csv_file") -lt 2 ]]; then
        log "WARN" "Insufficient data for charts"
        return 1
    fi
    
    local chart_file="$RESULTS_DIR/performance_charts.txt"
    
    {
        echo "Performance Charts"
        echo "================="
        echo ""
        
        # Build duration timeline
        echo "Build Duration Timeline:"
        echo "----------------------"
        awk -F',' 'NR>1 {
            gsub(/[:-]/," ", $1);
            split($1, dt, " ");
            printf "%s %6.2fs\n", dt[4], $2;
        }' "$csv_file" | sort -k1 | \
        awk '{
            if(NR==1) {
                min=$2; max=$2;
            } else {
                if($2<min) min=$2;
                if($2>max) max=$2;
            }
            times[NR]=$2;
            labels[NR]=$1;
        } END {
            width=50;
            range=max-min;
            if(range==0) range=1;
            
            for(i=1; i<=NR; i++) {
                stars=int((times[i]-min)/range*width);
                printf "%s %5.2fs ", labels[i], times[i];
                for(j=0; j<stars; j++) printf "*";
                printf "\n";
            }
        }'
        
        echo ""
        echo "Worker Count vs Duration:"
        echo "-----------------------"
        awk -F',' 'NR>1 {workers[$3]++; durations[$3]+=$2; count[$3]++} END {
            for(w in workers) {
                avg=durations[w]/count[w];
                data[w]=avg;
            }
            
            min=999999; max=0;
            for(w in data) {
                if(data[w]<min) min=data[w];
                if(data[w]>max) max=data[w];
            }
            
            for(w in data) {
                printf " %s workers: ", w;
                width=30;
                range=max-min;
                if(range==0) range=1;
                stars=int((data[w]-min)/range*width);
                for(i=0; i<stars; i++) printf "█";
                printf " %.2fs\n", data[w];
            }
        }' "$csv_file"
        
        echo ""
        echo "Cache Impact:"
        echo "------------"
        awk -F',' 'NR>1 {
            if($4=="true") {cache_count++; cache_duration+=$2}
            else {no_cache_count++; no_cache_duration+=$2}
        } END {
            if(cache_count>0) {
                cache_avg=cache_duration/cache_count;
                printf(" With Cache:    ");
                width=20;
                for(i=0; i<width*cache_avg/60; i++) printf "█";
                printf(" %.2fs (%d builds)\n", cache_avg, cache_count);
            }
            
            if(no_cache_count>0) {
                no_cache_avg=no_cache_duration/no_cache_count;
                printf(" Without Cache: ");
                width=20;
                for(i=0; i<width*no_cache_avg/60; i++) printf "█";
                printf(" %.2fs (%d builds)\n", no_cache_avg, no_cache_count);
            }
        }' "$csv_file"
        
    } > "$chart_file"
    
    log "INFO" "Charts saved to: $chart_file"
}

# Interactive mode for running experiments
run_experiments_interactive() {
    log "INFO" "Starting interactive performance analysis"
    
    # Get experiment parameters from user
    echo "Available build tasks:"
    echo "1) assemble"
    echo "2) assembleDebug" 
    echo "3) assembleRelease"
    echo "4) test"
    echo "5) custom task"
    read -p "Select build task (1-5): " task_choice
    
    local build_task
    case "$task_choice" in
        1) build_task="assemble" ;;
        2) build_task="assembleDebug" ;;
        3) build_task="assembleRelease" ;;
        4) build_task="test" ;;
        5) read -p "Enter custom Gradle task: " build_task ;;
        *) build_task="assemble" ;;
    esac
    
    echo ""
    read -p "Enter worker counts to test (space-separated, e.g., '1 2 4 8'): " worker_input
    IFS=' ' read -ra worker_counts <<< "$worker_input"
    
    echo ""
    read -p "Test with cache enabled? (y/n): " cache_choice
    local cache_enabled=("false")
    [[ "$cache_choice" =~ ^[Yy]$ ]] && cache_enabled=("false" "true")
    
    echo ""
    read -p "Number of runs per configuration: " runs_per_config
    runs_per_config=${runs_per_config:-1}
    
    log "INFO" "Starting experiments: ${#worker_counts[@]} worker configs × ${#cache_enabled[@]} cache configs × $runs_per_config runs"
    
    local total_experiments=$((${#worker_counts[@]} * ${#cache_enabled[@]} * runs_per_config))
    local current_experiment=0
    
    # Run experiments
    for cache in "${cache_enabled[@]}"; do
        for workers in "${worker_counts[@]}"; do
            for ((run=1; run<=runs_per_config; run++)); do
                ((current_experiment++))
                
                log "INFO" "Experiment $current_experiment/$total_experiments: workers=$workers, cache=$cache, run=$run"
                
                if run_build_experiment "$workers" "$cache" "$build_task"; then
                    log "INFO" "✅ Experiment completed successfully"
                else
                    log "WARN" "❌ Experiment failed"
                fi
                
                # Small delay between experiments
                sleep 2
            done
        done
    done
    
    # Analyze results
    echo ""
    log "INFO" "All experiments completed. Analyzing results..."
    analyze_metrics "$METRICS_FILE"
    generate_charts "$METRICS_FILE"
    
    # Show analysis
    echo ""
    cat "$ANALYSIS_FILE"
    
    echo ""
    log "INFO" "Detailed results available in: $RESULTS_DIR/"
}

# Quick performance test
quick_test() {
    log "INFO" "Running quick performance test"
    
    # Test with common configurations
    local configs=(
        "1 false"
        "2 false" 
        "4 false"
        "8 false"
        "4 true"
        "8 true"
    )
    
    for config in "${configs[@]}"; do
        IFS=' ' read -ra params <<< "$config"
        local workers="${params[0]}"
        local cache="${params[1]}"
        
        log "INFO" "Testing: workers=$workers, cache=$cache"
        
        if run_build_experiment "$workers" "$cache" "assemble"; then
            log "INFO" "✅ Test completed"
        else
            log "WARN" "❌ Test failed"
        fi
    done
    
    # Quick analysis
    analyze_metrics "$METRICS_FILE"
    generate_charts "$METRICS_FILE"
    
    echo ""
    log "INFO" "Quick test completed. Results in $RESULTS_DIR/"
}

# Main menu
main_menu() {
    while true; do
        echo ""
        echo "Distributed Gradle Build Performance Analyzer"
        echo "=========================================="
        echo "1) Run Interactive Experiments"
        echo "2) Quick Performance Test"
        echo "3) Analyze Existing Metrics"
        echo "4) Generate Charts"
        echo "5) View Results"
        echo "6) Exit"
        echo ""
        read -p "Select option (1-6): " choice
        
        case "$choice" in
            1) run_experiments_interactive ;;
            2) quick_test ;;
            3) analyze_metrics "$METRICS_FILE" ;;
            4) generate_charts "$METRICS_FILE" ;;
            5) 
                if [[ -f "$ANALYSIS_FILE" ]]; then
                    cat "$ANALYSIS_FILE"
                else
                    log "WARN" "No analysis file found. Run experiments first."
                fi
                ;;
            6) break ;;
            *) log "ERROR" "Invalid option. Please select 1-6." ;;
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
    
    if ! command -v timeout >/dev/null 2>&1; then
        log "ERROR" "timeout command is required but not installed"
        exit 1
    fi
    
    # Check if we're in a Gradle project
    if [[ ! -f "gradlew" ]]; then
        log "ERROR" "gradlew not found. Please run from a Gradle project directory."
        exit 1
    fi
    
    # Check if distributed build is configured
    if [[ ! -f ".gradlebuild_env" ]]; then
        log "ERROR" ".gradlebuild_env not found. Please run setup_master.sh first."
        exit 1
    fi
    
    # Initialize metrics collection
    initialize_metrics
    
    log "INFO" "Performance analyzer initialized in $PROJECT_DIR"
    
    # Command line argument handling
    case "${1:-}" in
        "test") quick_test ;;
        "analyze") analyze_metrics "$METRICS_FILE" ;;
        "charts") generate_charts "$METRICS_FILE" ;;
        "interactive"|"") main_menu ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [command]"
            echo ""
            echo "Commands:"
            echo "  test        - Run quick performance test"
            echo "  analyze     - Analyze existing metrics"
            echo "  charts      - Generate ASCII charts"
            echo "  interactive - Start interactive mode (default)"
            echo "  help        - Show this help"
            ;;
        *) log "ERROR" "Unknown command: $1"; main_menu ;;
    esac
}

# Run main function with all arguments
main "$@"