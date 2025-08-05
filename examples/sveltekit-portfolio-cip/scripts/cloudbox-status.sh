#!/bin/bash
set -e

# Load CloudBox Install Protocol common functions
source "$(dirname "$0")/cloudbox-common.sh"

# Display status header
cip_header "Status Check for $(cip_get_app_name) v$(cip_get_app_version)"

# Initialize status variables
APP_STATUS="unknown"
PORT_STATUS="unknown"
HEALTH_STATUS="unknown"
DOCKER_STATUS="unknown"
OVERALL_STATUS="unknown"

# Step 1: Check application process
cip_step "Checking application process"

APP_PID=""
if [[ -f ".cloudbox.pid" ]]; then
    APP_PID=$(cat .cloudbox.pid)
    if [[ -n "$APP_PID" ]] && kill -0 "$APP_PID" 2>/dev/null; then
        cip_success "✅ Application is running (PID: $APP_PID)"
        APP_STATUS="running"
        
        # Get process details
        if command -v ps >/dev/null 2>&1; then
            PROCESS_INFO=$(ps -o pid,ppid,pmem,pcpu,etime,comm -p "$APP_PID" --no-headers 2>/dev/null | tr -s ' ')
            if [[ -n "$PROCESS_INFO" ]]; then
                cip_info "Process Details: $PROCESS_INFO"
            fi
        fi
    else
        cip_error "❌ Application is not running (stale PID file)"
        rm -f .cloudbox.pid
        APP_STATUS="stopped"
        APP_PID=""
    fi
else
    # Check for processes by pattern
    APP_PID=$(cip_get_pid)
    if [[ -n "$APP_PID" ]]; then
        cip_warn "⚠️  Application is running but no PID file found (PID: $APP_PID)"
        echo "$APP_PID" > .cloudbox.pid  # Recreate PID file
        APP_STATUS="running"
    else
        cip_info "ℹ️  Application is not running"
        APP_STATUS="stopped"
    fi
fi

# Step 2: Check port status
cip_step "Checking port status"

PORT=""
if [[ -f ".env.production" ]]; then
    source .env.production
    PORT=${PORT:-3000}
elif [[ -n "${CLOUDBOX_WEB_PORT}" ]]; then
    PORT=${CLOUDBOX_WEB_PORT}
else
    PORT=3000
fi

cip_info "Checking port: $PORT"

if command -v nc >/dev/null 2>&1; then
    if nc -z localhost "$PORT" 2>/dev/null; then
        cip_success "✅ Port $PORT is responding"
        PORT_STATUS="open"
    else
        cip_error "❌ Port $PORT is not responding"
        PORT_STATUS="closed"
    fi
elif command -v curl >/dev/null 2>&1; then
    if curl -s --connect-timeout 3 "http://localhost:$PORT" >/dev/null 2>&1; then
        cip_success "✅ Port $PORT is responding"
        PORT_STATUS="open"
    else
        cip_error "❌ Port $PORT is not responding"
        PORT_STATUS="closed"
    fi
else
    cip_warn "Cannot check port status - no suitable tools available"
    PORT_STATUS="unknown"
fi

# Step 3: Advanced port diagnostics
if [[ "$PORT_STATUS" == "closed" ]] && [[ -n "$APP_PID" ]]; then
    cip_step "Performing port diagnostics"
    
    # Check if process is listening on any port
    if command -v netstat >/dev/null 2>&1; then
        LISTENING_PORTS=$(netstat -tlnp 2>/dev/null | grep "$APP_PID" | awk '{print $4}' | cut -d: -f2 | sort -u)
        if [[ -n "$LISTENING_PORTS" ]]; then
            cip_info "Process is listening on ports: $(echo $LISTENING_PORTS | tr '\n' ' ')"
        else
            cip_warn "Process is not listening on any ports"
        fi
    elif command -v ss >/dev/null 2>&1; then
        LISTENING_PORTS=$(ss -tlnp 2>/dev/null | grep "pid=$APP_PID" | awk '{print $4}' | cut -d: -f2 | sort -u)
        if [[ -n "$LISTENING_PORTS" ]]; then
            cip_info "Process is listening on ports: $(echo $LISTENING_PORTS | tr '\n' ' ')"
        fi
    fi
    
    # Check what's using the expected port
    if command -v lsof >/dev/null 2>&1; then
        PORT_USERS=$(lsof -i ":$PORT" -t 2>/dev/null | head -5)
        if [[ -n "$PORT_USERS" ]]; then
            cip_warn "Port $PORT is being used by processes: $(echo $PORT_USERS | tr '\n' ' ')"
        fi
    fi
fi

# Step 4: Health check
cip_step "Performing health check"

if [[ "$PORT_STATUS" == "open" ]]; then
    # Try health endpoint first
    if cip_health_check "/health" "$PORT"; then
        cip_success "✅ Health endpoint is responding"
        HEALTH_STATUS="healthy"
        
        # Get detailed health information
        if command -v curl >/dev/null 2>&1; then
            HEALTH_RESPONSE=$(curl -s --connect-timeout 5 "http://localhost:$PORT/health" 2>/dev/null || echo "")
            if [[ -n "$HEALTH_RESPONSE" ]]; then
                cip_info "Health Response: $HEALTH_RESPONSE"
            fi
        fi
    else
        # Fallback to root endpoint
        if cip_health_check "/" "$PORT"; then
            cip_warn "⚠️  Root endpoint responding but no health endpoint"
            HEALTH_STATUS="partial"
        else
            cip_error "❌ No endpoints are responding"
            HEALTH_STATUS="unhealthy"
        fi
    fi
else
    cip_info "Skipping health check - port not accessible"
    HEALTH_STATUS="unknown"
fi

# Step 5: Application-specific health checks
if [[ -f "package.json" ]] && grep -q "healthcheck" package.json 2>/dev/null; then
    cip_step "Running application health checks"
    
    if npm run healthcheck --silent 2>/dev/null; then
        cip_success "✅ Application health checks passed"
        if [[ "$HEALTH_STATUS" == "unknown" ]]; then
            HEALTH_STATUS="healthy"
        fi
    else
        cip_error "❌ Application health checks failed"
        HEALTH_STATUS="unhealthy"
    fi
fi

# Step 6: System resource status
cip_step "Checking system resources"

# Memory usage
if command -v free >/dev/null 2>&1; then
    MEMORY_PERCENT=$(free | awk 'FNR==2{printf "%.0f", $3/$2*100}' 2>/dev/null || echo "0")
    if [[ $MEMORY_PERCENT -lt 80 ]]; then
        cip_success "✅ Memory usage: ${MEMORY_PERCENT}%"
        MEMORY_STATUS="good"
    elif [[ $MEMORY_PERCENT -lt 90 ]]; then
        cip_warn "⚠️  Memory usage: ${MEMORY_PERCENT}%"
        MEMORY_STATUS="warning"
    else
        cip_error "❌ High memory usage: ${MEMORY_PERCENT}%"
        MEMORY_STATUS="critical"
    fi
else
    MEMORY_STATUS="unknown"
    MEMORY_PERCENT="N/A"
fi

# Disk usage
if command -v df >/dev/null 2>&1; then
    DISK_PERCENT=$(df . | awk 'FNR==2{print $5}' | sed 's/%//' 2>/dev/null || echo "0")
    if [[ $DISK_PERCENT -lt 85 ]]; then
        cip_success "✅ Disk usage: ${DISK_PERCENT}%"
        DISK_STATUS="good"
    elif [[ $DISK_PERCENT -lt 95 ]]; then
        cip_warn "⚠️  Disk usage: ${DISK_PERCENT}%"
        DISK_STATUS="warning"
    else
        cip_error "❌ High disk usage: ${DISK_PERCENT}%"
        DISK_STATUS="critical"
    fi
else
    DISK_STATUS="unknown"
    DISK_PERCENT="N/A"
fi

# CPU usage (if available)
CPU_PERCENT="N/A"
if [[ -n "$APP_PID" ]] && command -v ps >/dev/null 2>&1; then
    CPU_PERCENT=$(ps -o pcpu= -p "$APP_PID" 2>/dev/null | tr -d ' ' || echo "N/A")
    if [[ "$CPU_PERCENT" != "N/A" ]] && [[ "$CPU_PERCENT" != "" ]]; then
        if (( $(echo "$CPU_PERCENT < 50" | bc -l 2>/dev/null || echo "1") )); then
            cip_success "✅ CPU usage: ${CPU_PERCENT}%"
        else
            cip_warn "⚠️  CPU usage: ${CPU_PERCENT}%"
        fi
    fi
fi

# Step 7: Docker monitoring status
if [[ "${CLOUDBOX_DOCKER_ENABLED}" == "true" ]] && [[ -f "docker-compose.cloudbox.yml" ]]; then
    cip_step "Checking Docker monitoring"
    
    if command -v docker-compose >/dev/null 2>&1; then
        if docker-compose -f docker-compose.cloudbox.yml ps -q monitoring 2>/dev/null | grep -q .; then
            DOCKER_CONTAINER_STATUS=$(docker-compose -f docker-compose.cloudbox.yml ps monitoring 2>/dev/null | tail -n +3 | awk '{print $4}' || echo "unknown")
            if [[ "$DOCKER_CONTAINER_STATUS" == "Up" ]]; then
                cip_success "✅ Docker monitoring is running"
                DOCKER_STATUS="running"
            else
                cip_warn "⚠️  Docker monitoring status: $DOCKER_CONTAINER_STATUS"
                DOCKER_STATUS="issues"
            fi
        else
            cip_warn "⚠️  Docker monitoring is not running"
            DOCKER_STATUS="stopped"
        fi
    else
        cip_warn "Docker Compose not available"
        DOCKER_STATUS="unavailable"
    fi
else
    DOCKER_STATUS="disabled"
fi

# Step 8: Log file analysis
cip_step "Analyzing log files"

ERROR_COUNT=0
ACCESS_COUNT=0
RECENT_ERRORS=""

if [[ -f "logs/error.log" ]]; then
    ERROR_COUNT=$(wc -l < logs/error.log 2>/dev/null || echo "0")
    if [[ $ERROR_COUNT -gt 0 ]]; then
        cip_warn "⚠️  $ERROR_COUNT error log entries found"
        
        # Show recent errors (last 5 lines)
        RECENT_ERRORS=$(tail -5 logs/error.log 2>/dev/null | head -3 || echo "")
        if [[ -n "$RECENT_ERRORS" ]]; then
            cip_info "Recent errors:"
            echo "$RECENT_ERRORS" | while read line; do
                if [[ -n "$line" ]]; then
                    cip_info "  $(echo "$line" | cut -c1-100)"
                fi
            done
        fi
    else
        cip_success "✅ No errors in error log"
    fi
fi

if [[ -f "logs/access.log" ]]; then
    ACCESS_COUNT=$(wc -l < logs/access.log 2>/dev/null || echo "0")
    if [[ $ACCESS_COUNT -gt 0 ]]; then
        cip_info "📊 $ACCESS_COUNT access log entries"
        
        # Show recent access (last 2 lines)
        RECENT_ACCESS=$(tail -2 logs/access.log 2>/dev/null || echo "")
        if [[ -n "$RECENT_ACCESS" ]]; then
            cip_info "Recent activity:"
            echo "$RECENT_ACCESS" | while read line; do
                if [[ -n "$line" ]]; then
                    cip_info "  $(echo "$line" | cut -c1-80)"
                fi
            done
        fi
    fi
fi

# Step 9: Uptime calculation
cip_step "Calculating uptime"

UPTIME="unknown"
STARTUP_TIME=""
if [[ -f ".cloudbox-startup-time" ]]; then
    STARTUP_TIME=$(cat .cloudbox-startup-time)
    if command -v date >/dev/null 2>&1; then
        STARTUP_EPOCH=$(date -d "$STARTUP_TIME" +%s 2>/dev/null || date -j -f "%Y-%m-%dT%H:%M:%SZ" "$STARTUP_TIME" +%s 2>/dev/null || echo "0")
        CURRENT_EPOCH=$(date +%s)
        UPTIME_SECONDS=$((CURRENT_EPOCH - STARTUP_EPOCH))
        
        if [[ $UPTIME_SECONDS -gt 0 ]]; then
            if [[ $UPTIME_SECONDS -lt 60 ]]; then
                UPTIME="${UPTIME_SECONDS} seconds"
            elif [[ $UPTIME_SECONDS -lt 3600 ]]; then
                UPTIME="$((UPTIME_SECONDS / 60)) minutes, $((UPTIME_SECONDS % 60)) seconds"
            else
                HOURS=$((UPTIME_SECONDS / 3600))
                MINUTES=$(((UPTIME_SECONDS % 3600) / 60))
                UPTIME="${HOURS} hours, ${MINUTES} minutes"
            fi
            cip_success "✅ Uptime: $UPTIME"
        fi
    fi
fi

# Step 10: CloudBox integration status
if [[ -n "${CLOUDBOX_API_URL:-${API_URL}}" ]]; then
    cip_step "Checking CloudBox integration"
    
    API_URL_TO_CHECK="${CLOUDBOX_API_URL:-${API_URL}}"
    if [[ "$API_URL_TO_CHECK" != "http://localhost:3001" ]]; then
        # Extract host and port from API URL
        API_HOST=$(echo "$API_URL_TO_CHECK" | sed 's|.*://||' | cut -d'/' -f1 | cut -d':' -f1)
        API_PORT=$(echo "$API_URL_TO_CHECK" | sed 's|.*://||' | cut -d'/' -f1 | cut -d':' -f2)
        
        if [[ "$API_PORT" == "$API_HOST" ]]; then
            API_PORT="80"  # Default HTTP port
        fi
        
        if command -v nc >/dev/null 2>&1; then
            if nc -z "$API_HOST" "$API_PORT" 2>/dev/null; then
                cip_success "✅ CloudBox API is reachable at $API_URL_TO_CHECK"
                CLOUDBOX_STATUS="connected"
            else
                cip_warn "⚠️  CloudBox API is not reachable at $API_URL_TO_CHECK"
                CLOUDBOX_STATUS="disconnected"
            fi
        else
            CLOUDBOX_STATUS="unknown"
        fi
    else
        CLOUDBOX_STATUS="local"
    fi
else
    CLOUDBOX_STATUS="not_configured"
fi

# Step 11: Determine overall status
cip_step "Determining overall status"

if [[ "$APP_STATUS" == "running" ]] && [[ "$PORT_STATUS" == "open" ]] && [[ "$HEALTH_STATUS" =~ ^(healthy|partial)$ ]]; then
    OVERALL_STATUS="healthy"
    cip_success "✅ Overall status: HEALTHY"
elif [[ "$APP_STATUS" == "running" ]] && [[ "$PORT_STATUS" == "open" ]]; then
    OVERALL_STATUS="degraded"
    cip_warn "⚠️  Overall status: DEGRADED"
elif [[ "$APP_STATUS" == "running" ]]; then
    OVERALL_STATUS="unhealthy"
    cip_error "❌ Overall status: UNHEALTHY"
else
    OVERALL_STATUS="stopped"
    cip_info "ℹ️  Overall status: STOPPED"
fi

# Step 12: Create status report
cat > .cloudbox-status-report.json << EOF
{
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "app_name": "$(cip_get_app_name)",
  "app_version": "$(cip_get_app_version)",
  "overall_status": "$OVERALL_STATUS",
  "application": {
    "status": "$APP_STATUS",
    "pid": "${APP_PID:-null}",
    "uptime": "$UPTIME",
    "started_at": "${STARTUP_TIME:-null}"
  },
  "network": {
    "port": $PORT,
    "status": "$PORT_STATUS",
    "health_endpoint": "http://localhost:$PORT/health"
  },
  "health": {
    "status": "$HEALTH_STATUS",
    "checks_available": $(grep -q "healthcheck" package.json 2>/dev/null && echo "true" || echo "false")
  },
  "resources": {
    "memory_percent": "${MEMORY_PERCENT}",
    "memory_status": "${MEMORY_STATUS}",
    "disk_percent": "${DISK_PERCENT}",
    "disk_status": "${DISK_STATUS}",
    "cpu_percent": "${CPU_PERCENT}"
  },
  "docker": {
    "enabled": ${CLOUDBOX_DOCKER_ENABLED:-false},
    "status": "$DOCKER_STATUS"
  },
  "logs": {
    "error_count": $ERROR_COUNT,
    "access_count": $ACCESS_COUNT,
    "error_log": "logs/error.log",
    "access_log": "logs/access.log"
  },
  "cloudbox": {
    "integration_status": "${CLOUDBOX_STATUS}",
    "api_url": "${CLOUDBOX_API_URL:-${API_URL}}"
  }
}
EOF

# Step 13: Display comprehensive status summary
echo
cip_info "📊 Comprehensive Status Report"
echo "=================================================="
echo
cip_info "🏷️  Application Information:"
echo "   Name: $(cip_get_app_name)"
echo "   Version: $(cip_get_app_version)"
echo "   Type: $(cip_get_app_type)"
echo "   Status: $OVERALL_STATUS"
echo
cip_info "🔄 Process Information:"
echo "   Status: $APP_STATUS"
if [[ -n "$APP_PID" ]]; then
    echo "   PID: $APP_PID"
fi
echo "   Uptime: $UPTIME"
if [[ -n "$STARTUP_TIME" ]]; then
    echo "   Started: $STARTUP_TIME"
fi
echo
cip_info "🌐 Network Status:"
echo "   Port: $PORT"
echo "   Port Status: $PORT_STATUS"
echo "   Health Status: $HEALTH_STATUS"
if [[ "$PORT_STATUS" == "open" ]]; then
    echo "   URL: http://localhost:$PORT"
    echo "   Health: http://localhost:$PORT/health"
fi
echo
cip_info "💾 Resource Usage:"
echo "   Memory: ${MEMORY_PERCENT}% ($MEMORY_STATUS)"
echo "   Disk: ${DISK_PERCENT}% ($DISK_STATUS)"
echo "   CPU: ${CPU_PERCENT}%"
echo
cip_info "📄 Log Summary:"
echo "   Access Entries: $ACCESS_COUNT"
echo "   Error Entries: $ERROR_COUNT"
if [[ $ERROR_COUNT -gt 0 ]]; then
    echo "   Recent Errors: Available in logs/error.log"
fi
echo
cip_info "🐳 Docker Status:"
echo "   Docker Enabled: ${CLOUDBOX_DOCKER_ENABLED:-false}"
echo "   Monitoring: $DOCKER_STATUS"
echo
cip_info "🔗 CloudBox Integration:"
echo "   Status: ${CLOUDBOX_STATUS}"
if [[ -n "${CLOUDBOX_API_URL:-${API_URL}}" ]]; then
    echo "   API URL: ${CLOUDBOX_API_URL:-${API_URL}}"
fi
if [[ -n "${CLOUDBOX_PROJECT_ID:-${PROJECT_ID}}" ]]; then
    echo "   Project ID: ${CLOUDBOX_PROJECT_ID:-${PROJECT_ID}}"
fi

# Step 14: Management commands
echo
cip_info "🔧 Management Commands:"
if [[ "$APP_STATUS" == "running" ]]; then
    echo "   Stop: ./scripts/cloudbox-stop.sh"
    echo "   Restart: ./scripts/cloudbox-stop.sh && ./scripts/cloudbox-start.sh"
    echo "   Health: ./scripts/cloudbox-health.sh"
else
    echo "   Start: ./scripts/cloudbox-start.sh"
    echo "   Install: ./scripts/cloudbox-install.sh"
fi
echo "   Logs: tail -f logs/access.log"
if [[ $ERROR_COUNT -gt 0 ]]; then
    echo "   Errors: tail -f logs/error.log"
fi

# Step 15: Recommendations
echo
if [[ "$OVERALL_STATUS" != "healthy" ]]; then
    cip_info "💡 Recommendations:"
    
    if [[ "$APP_STATUS" == "stopped" ]]; then
        echo "   • Start the application with ./scripts/cloudbox-start.sh"
    fi
    
    if [[ "$PORT_STATUS" == "closed" ]] && [[ "$APP_STATUS" == "running" ]]; then
        echo "   • Check if application is binding to correct port ($PORT)"
        echo "   • Review logs for startup errors"
    fi
    
    if [[ "$HEALTH_STATUS" == "unhealthy" ]]; then
        echo "   • Check application health endpoint configuration"
        echo "   • Review error logs for health check failures"
    fi
    
    if [[ "$MEMORY_STATUS" == "critical" ]]; then
        echo "   • Consider restarting to free memory"
        echo "   • Monitor for memory leaks"
    fi
    
    if [[ "$DISK_STATUS" == "critical" ]]; then
        echo "   • Clean up log files and temporary data"
        echo "   • Consider log rotation"
    fi
    
    if [[ $ERROR_COUNT -gt 10 ]]; then
        echo "   • Review error logs for recurring issues"
        echo "   • Consider log rotation to manage disk space"
    fi
fi

echo
cip_info "📋 Status report saved to: .cloudbox-status-report.json"

# Step 16: Exit with appropriate code
if [[ "$OVERALL_STATUS" == "healthy" ]]; then
    exit 0
elif [[ "$OVERALL_STATUS" == "degraded" ]]; then
    exit 1
elif [[ "$OVERALL_STATUS" == "stopped" ]]; then
    exit 3
else
    exit 2  # unhealthy
fi