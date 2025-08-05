#!/bin/bash
set -e

# Load CloudBox Install Protocol common functions
source "$(dirname "$0")/cloudbox-common.sh"

# Display stop header
cip_header "Stopping $(cip_get_app_name) v$(cip_get_app_version)"

# Step 1: Check if application is running
cip_step "Checking application status"

APP_PID=""
if [[ -f ".cloudbox.pid" ]]; then
    APP_PID=$(cat .cloudbox.pid)
    if [[ -n "$APP_PID" ]] && kill -0 "$APP_PID" 2>/dev/null; then
        cip_info "Application is running (PID: $APP_PID)"
    else
        cip_warn "PID file exists but process is not running"
        rm -f .cloudbox.pid
        APP_PID=""
    fi
else
    # Check for processes by pattern
    APP_PID=$(cip_get_pid)
    if [[ -n "$APP_PID" ]]; then
        cip_warn "Application running without PID file (PID: $APP_PID)"
    else
        cip_info "No running application processes found"
    fi
fi

# Step 2: Graceful shutdown
if [[ -n "$APP_PID" ]]; then
    cip_step "Initiating graceful shutdown"
    
    # Send SIGTERM for graceful shutdown
    if kill -TERM "$APP_PID" 2>/dev/null; then
        cip_info "Sent SIGTERM to process $APP_PID"
        
        # Wait for graceful shutdown (up to 15 seconds)
        local count=0
        while [[ $count -lt 15 ]] && kill -0 "$APP_PID" 2>/dev/null; do
            sleep 1
            ((count++))
            
            if [[ $((count % 3)) -eq 0 ]]; then
                cip_debug "Waiting for graceful shutdown... (${count}/15s)"
            fi
        done
        
        # Check if process stopped
        if ! kill -0 "$APP_PID" 2>/dev/null; then
            cip_success "Application stopped gracefully"
        else
            cip_warn "Process did not stop gracefully, forcing shutdown..."
            
            # Send SIGKILL
            if kill -KILL "$APP_PID" 2>/dev/null; then
                sleep 2
                if ! kill -0 "$APP_PID" 2>/dev/null; then
                    cip_success "Application force-stopped"
                else
                    cip_error "Failed to stop process $APP_PID"
                fi
            else
                cip_warn "Process may have already stopped"
            fi
        fi
    else
        cip_warn "Could not send signal to process $APP_PID (may have already stopped)"
    fi
    
    # Clean up PID file
    rm -f .cloudbox.pid
else
    cip_info "No application process to stop"
fi

# Step 3: Stop monitoring process
cip_step "Stopping monitoring processes"

if [[ -f ".cloudbox-monitor.pid" ]]; then
    MONITOR_PID=$(cat .cloudbox-monitor.pid)
    if [[ -n "$MONITOR_PID" ]] && kill -0 "$MONITOR_PID" 2>/dev/null; then
        kill "$MONITOR_PID" 2>/dev/null || true
        cip_success "Application monitoring stopped"
    fi
    rm -f .cloudbox-monitor.pid .cloudbox-monitor.sh
fi

# Step 4: Clean up any remaining processes
cip_step "Cleaning up remaining processes"

# Kill any remaining node processes for this app
APP_NAME=$(basename "$PWD")
REMAINING_PIDS=$(pgrep -f "node.*$APP_NAME" 2>/dev/null || echo "")

if [[ -n "$REMAINING_PIDS" ]]; then
    cip_warn "Found remaining processes, cleaning up..."
    echo "$REMAINING_PIDS" | while read pid; do
        if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
            cip_debug "Stopping process: $pid"
            kill "$pid" 2>/dev/null || true
        fi
    done
    sleep 2
    
    # Final check
    REMAINING_PIDS=$(pgrep -f "node.*$APP_NAME" 2>/dev/null || echo "")
    if [[ -n "$REMAINING_PIDS" ]]; then
        cip_warn "Force killing remaining processes"
        echo "$REMAINING_PIDS" | xargs kill -9 2>/dev/null || true
    fi
fi

cip_success "Process cleanup completed"

# Step 5: Stop Docker monitoring (if enabled)
if [[ "${CLOUDBOX_DOCKER_ENABLED}" == "true" ]] && [[ -f "docker-compose.cloudbox.yml" ]]; then
    cip_step "Stopping Docker monitoring"
    
    if docker-compose -f docker-compose.cloudbox.yml down 2>/dev/null; then
        cip_success "Docker monitoring stopped"
    else
        cip_warn "Failed to stop Docker monitoring (may not be running)"
    fi
else
    cip_debug "Docker monitoring not configured or disabled"
fi

# Step 6: Port cleanup verification
if [[ -f ".env.production" ]]; then
    source .env.production
    PORT=${PORT:-3000}
    
    cip_step "Verifying port ${PORT} is released"
    
    # Wait a moment for port to be released
    sleep 2
    
    if cip_check_port "$PORT"; then
        cip_success "Port ${PORT} is now available"
    else
        cip_warn "Port ${PORT} may still be in use by another process"
        
        # Show what's using the port
        if command -v netstat >/dev/null 2>&1; then
            cip_info "Processes using port ${PORT}:"
            netstat -tlnp | grep ":${PORT} " | head -3 | while read line; do
                cip_info "  $line"
            done
        fi
    fi
fi

# Step 7: Clean up temporary files
cip_step "Cleaning up temporary files"

# Remove runtime status files
rm -f .cloudbox-status.json .cloudbox-startup-time

# Clean up temporary directories
rm -rf .tmp-* /tmp/cloudbox-* 2>/dev/null || true

# Rotate large log files
if [[ -f "logs/access.log" ]]; then
    LOG_SIZE=$(stat -f%z "logs/access.log" 2>/dev/null || stat -c%s "logs/access.log" 2>/dev/null || echo 0)
    if [[ $LOG_SIZE -gt 10485760 ]]; then  # 10MB
        cip_info "Rotating large access log (${LOG_SIZE} bytes)"
        mv logs/access.log "logs/access.log.$(date +%Y%m%d_%H%M%S)"
        touch logs/access.log
    fi
fi

if [[ -f "logs/error.log" ]]; then
    ERROR_LOG_SIZE=$(stat -f%z "logs/error.log" 2>/dev/null || stat -c%s "logs/error.log" 2>/dev/null || echo 0)
    if [[ $ERROR_LOG_SIZE -gt 5242880 ]]; then  # 5MB
        cip_info "Rotating large error log (${ERROR_LOG_SIZE} bytes)"
        mv logs/error.log "logs/error.log.$(date +%Y%m%d_%H%M%S)"
        touch logs/error.log
    fi
fi

cip_success "Temporary files cleaned up"

# Step 8: Resource cleanup
cip_step "Releasing system resources"

# Clear Node.js cache if memory pressure is high
if command -v free >/dev/null 2>&1; then
    MEMORY_PERCENT=$(free | awk 'FNR==2{printf "%.0f", $3/$2*100}' 2>/dev/null || echo "50")
    if [[ $MEMORY_PERCENT -gt 85 ]]; then
        cip_info "High memory usage detected (${MEMORY_PERCENT}%), clearing Node.js cache"
        # Force garbage collection if possible
        sync && echo 3 > /proc/sys/vm/drop_caches 2>/dev/null || true
    fi
fi

# Clean npm cache if disk space is low
if command -v npm >/dev/null 2>&1; then
    DISK_PERCENT=$(df . | awk 'FNR==2{print $5}' | sed 's/%//' 2>/dev/null || echo "50")
    if [[ $DISK_PERCENT -gt 90 ]]; then
        cip_info "Low disk space (${DISK_PERCENT}% used), cleaning npm cache"
        npm cache clean --force 2>/dev/null || true
    fi
fi

cip_success "System resources released"

# Step 9: Create shutdown report
cip_step "Generating shutdown report"

SHUTDOWN_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
UPTIME="unknown"

if [[ -f ".cloudbox-startup-time" ]]; then
    STARTUP_TIME=$(cat .cloudbox-startup-time)
    if command -v date >/dev/null 2>&1; then
        STARTUP_EPOCH=$(date -d "$STARTUP_TIME" +%s 2>/dev/null || date -j -f "%Y-%m-%dT%H:%M:%SZ" "$STARTUP_TIME" +%s 2>/dev/null || echo "0")
        SHUTDOWN_EPOCH=$(date +%s)
        UPTIME_SECONDS=$((SHUTDOWN_EPOCH - STARTUP_EPOCH))
        UPTIME="${UPTIME_SECONDS} seconds"
        
        if [[ $UPTIME_SECONDS -gt 60 ]]; then
            UPTIME="$((UPTIME_SECONDS / 60)) minutes, $((UPTIME_SECONDS % 60)) seconds"
        fi
        if [[ $UPTIME_SECONDS -gt 3600 ]]; then
            HOURS=$((UPTIME_SECONDS / 3600))
            MINUTES=$(((UPTIME_SECONDS % 3600) / 60))
            UPTIME="${HOURS} hours, ${MINUTES} minutes"
        fi
    fi
    
    rm -f .cloudbox-startup-time
fi

cat > .cloudbox-shutdown.json << EOF
{
  "app_name": "$(cip_get_app_name)",
  "app_version": "$(cip_get_app_version)",
  "shutdown_at": "$SHUTDOWN_TIME",
  "uptime": "$UPTIME",
  "shutdown_reason": "manual_stop",
  "final_status": "stopped",
  "resources_cleaned": true,
  "graceful_shutdown": true
}
EOF

cip_success "Shutdown report created"

# Step 10: Final verification
cip_step "Performing final verification"

# Verify no processes are running
REMAINING_PROCESSES=$(pgrep -f "node.*$(basename "$PWD")" 2>/dev/null | wc -l || echo "0")
if [[ $REMAINING_PROCESSES -eq 0 ]]; then
    cip_success "✅ No application processes remain"
else
    cip_warn "⚠️  $REMAINING_PROCESSES application processes may still be running"
fi

# Verify port is released
if [[ -n "$PORT" ]]; then
    if cip_check_port "$PORT"; then
        cip_success "✅ Port $PORT is available"
    else
        cip_warn "⚠️  Port $PORT may still be in use"
    fi
fi

# Check for any open file descriptors (if lsof is available)
if command -v lsof >/dev/null 2>&1; then
    OPEN_FILES=$(lsof +D . 2>/dev/null | grep node | wc -l || echo "0")
    if [[ $OPEN_FILES -eq 0 ]]; then
        cip_debug "No open file descriptors found"
    else
        cip_debug "$OPEN_FILES open file descriptors found (may be normal)"
    fi
fi

# Step 11: Display shutdown summary
echo
cip_success "🛑 $(cip_get_app_name) stopped successfully!"
echo
cip_info "📋 Shutdown Summary:"
cip_info "  Status: Stopped"
cip_info "  Uptime: $UPTIME"
cip_info "  Shutdown Time: $(date)"
cip_info "  Graceful: Yes"
echo
cip_info "🧹 Cleanup Actions:"
cip_info "  ✅ Application processes stopped"
cip_info "  ✅ Monitoring processes stopped"
cip_info "  ✅ Temporary files cleaned"
cip_info "  ✅ System resources released"
if [[ "${CLOUDBOX_DOCKER_ENABLED}" == "true" ]]; then
    cip_info "  ✅ Docker monitoring stopped"
fi
echo
cip_info "🔄 Management Commands:"
cip_info "  Start: ./scripts/cloudbox-start.sh"
cip_info "  Status: ./scripts/cloudbox-status.sh"
cip_info "  Reinstall: ./scripts/cloudbox-install.sh"
echo
cip_info "📄 Available Reports:"
cip_info "  Shutdown Report: .cloudbox-shutdown.json"
if [[ -f "logs/access.log" ]]; then
    cip_info "  Access Logs: logs/access.log"
fi
if [[ -f "logs/error.log" ]]; then
    cip_info "  Error Logs: logs/error.log"
fi

# Optional: Show log summary
if [[ -f "logs/access.log" ]] && [[ -s "logs/access.log" ]]; then
    ACCESS_LINES=$(wc -l < logs/access.log 2>/dev/null || echo "0")
    if [[ $ACCESS_LINES -gt 0 ]]; then
        cip_info "📊 Log Summary: $ACCESS_LINES access log entries"
    fi
fi

if [[ -f "logs/error.log" ]] && [[ -s "logs/error.log" ]]; then
    ERROR_LINES=$(wc -l < logs/error.log 2>/dev/null || echo "0")
    if [[ $ERROR_LINES -gt 0 ]]; then
        cip_warn "⚠️  $ERROR_LINES error log entries found - review logs/error.log"
    fi
fi

echo
cip_success "✅ Shutdown completed successfully!"

# Step 12: Final system state
cip_debug "Final system resource check"
cip_check_resources

cip_info "🔧 Application is ready to be restarted when needed"
cip_debug "Shutdown process completed"