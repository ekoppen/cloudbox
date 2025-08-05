#!/bin/bash
set -e

# Load CloudBox Install Protocol common functions
source "$(dirname "$0")/cloudbox-common.sh"

# Display start header
cip_header "Starting $(cip_get_app_name) v$(cip_get_app_version)"

# Step 1: Pre-start validation
cip_step "Performing pre-start validation"

# Check if app was installed
if [[ ! -f ".cloudbox-info.json" ]]; then
    cip_fatal "Application not properly installed - run cloudbox-install.sh first"
fi

# Load environment configuration
if [[ -f ".env.production" ]]; then
    source .env.production
    cip_success "Production environment loaded"
    cip_debug "PORT: ${PORT}"
    cip_debug "API_URL: ${API_URL}"
else
    cip_warn "No .env.production found - using CloudBox defaults"
    PORT=${CLOUDBOX_WEB_PORT:-3000}
    API_URL=${CLOUDBOX_API_URL:-http://localhost:3001}
fi

# Validate dependencies
if [[ ! -d "node_modules" ]]; then
    cip_fatal "Dependencies not installed - run cloudbox-install.sh first"
fi

if [[ ! -d "build" ]] && [[ ! -f "build/index.js" ]]; then
    cip_fatal "Application not built - run cloudbox-install.sh first"
fi

cip_success "Pre-start validation passed"

# Step 2: Stop existing processes
cip_step "Stopping existing processes"
cip_stop_process

# Clean up stale PID files
rm -f .cloudbox.pid .cloudbox-startup.log

# Step 3: Port availability check
cip_step "Checking port availability"
if ! cip_check_port "${PORT}"; then
    cip_error "Port ${PORT} is already in use"
    cip_info "Checking what's using the port..."
    
    if command -v netstat >/dev/null 2>&1; then
        netstat -tlnp | grep ":${PORT} " | head -5
    elif command -v ss >/dev/null 2>&1; then
        ss -tlnp | grep ":${PORT} " | head -5
    fi
    
    cip_fatal "Cannot start - port ${PORT} is not available"
fi
cip_success "Port ${PORT} is available"

# Step 4: Resource checks
cip_step "Checking system resources"
cip_check_resources 90 95  # Higher thresholds for startup

# Check disk space for logs
AVAILABLE_SPACE=$(df . | awk 'FNR==2{print $4}')
if [[ $AVAILABLE_SPACE -lt 100000 ]]; then  # Less than ~100MB
    cip_warn "Low disk space available: ${AVAILABLE_SPACE}KB"
    cip_info "Consider cleaning up logs and temporary files"
fi

# Step 5: Backup previous logs
cip_step "Preparing log files"
if [[ -f "application.log" ]]; then
    mv application.log "application.log.$(date +%Y%m%d_%H%M%S)"
    cip_debug "Previous log file backed up"
fi

# Create log directory if needed
mkdir -p logs
touch logs/access.log logs/error.log

# Step 6: Environment setup
cip_step "Setting up runtime environment"

# Export all CloudBox variables for the application
export NODE_ENV=production
export PORT="${PORT}"
export HOST="${HOST:-0.0.0.0}"
export API_URL="${API_URL}"
export PROJECT_ID="${PROJECT_ID:-${CLOUDBOX_PROJECT_ID}}"

# SvelteKit specific environment
export ORIGIN="http://localhost:${PORT}"
export BODY_SIZE_LIMIT="${BODY_SIZE_LIMIT:-10485760}"

# CloudBox integration variables
export CLOUDBOX_STORAGE_API="${API_URL}/storage"
export CLOUDBOX_AUTH_API="${API_URL}/auth"
export CLOUDBOX_DATA_API="${API_URL}/data"

cip_success "Runtime environment configured"
cip_debug "Application will run on http://localhost:${PORT}"

# Step 7: Start the application
cip_step "Starting SvelteKit application server"

# Choose startup method based on what's available
if [[ -f "package.json" ]] && npm run start --if-present > /dev/null 2>&1; then
    cip_info "Using npm start script"
    
    # Start with output redirection and background processing
    nohup npm start > logs/access.log 2> logs/error.log &
    APP_PID=$!
    
elif [[ -f "build/index.js" ]]; then
    cip_info "Using direct Node.js execution"
    
    # Start with node directly
    nohup node build/index.js > logs/access.log 2> logs/error.log &
    APP_PID=$!
    
else
    cip_fatal "Cannot determine how to start the application"
fi

# Save PID for process management
echo $APP_PID > .cloudbox.pid
echo "$(date -u +"%Y-%m-%dT%H:%M:%SZ")" > .cloudbox-startup-time

cip_success "Application server started (PID: $APP_PID)"

# Step 8: Startup verification
cip_step "Verifying application startup"

# Wait for the application to be ready
if cip_wait_for_port "${PORT}" 60; then
    cip_success "✅ Application is responding on port ${PORT}"
else
    cip_error "Application failed to start within 60 seconds"
    
    # Show recent logs for debugging
    if [[ -f "logs/error.log" ]]; then
        cip_error "Recent error logs:"
        tail -10 logs/error.log | while read line; do
            cip_error "  $line"
        done
    fi
    
    # Clean up failed process
    if [[ -n "$APP_PID" ]] && kill -0 "$APP_PID" 2>/dev/null; then
        kill "$APP_PID"
    fi
    rm -f .cloudbox.pid
    
    cip_fatal "Application startup failed"
fi

# Step 9: Health check verification
cip_step "Performing initial health check"
sleep 2  # Give the app a moment to fully initialize

if cip_health_check "/health" "${PORT}"; then
    cip_success "✅ Health check endpoint is responding"
else
    cip_warn "⚠️  Health check endpoint not responding (this may be normal)"
    cip_info "The application may still be initializing"
fi

# Step 10: Docker monitoring (if enabled)
if [[ "${CLOUDBOX_DOCKER_ENABLED}" == "true" ]] && [[ -f "docker-compose.cloudbox.yml" ]]; then
    cip_step "Starting Docker monitoring"
    
    if docker-compose -f docker-compose.cloudbox.yml up -d monitoring 2>/dev/null; then
        cip_success "Docker monitoring started"
    else
        cip_warn "Failed to start Docker monitoring - continuing without it"
    fi
else
    cip_debug "Docker monitoring disabled or not configured"
fi

# Step 11: Create runtime status file
cat > .cloudbox-status.json << EOF
{
  "status": "running",
  "pid": $APP_PID,
  "started_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "port": ${PORT},
  "health_endpoint": "http://localhost:${PORT}/health",
  "log_files": {
    "access": "logs/access.log",
    "error": "logs/error.log"
  },
  "configuration": {
    "node_env": "${NODE_ENV}",
    "host": "${HOST}",
    "api_url": "${API_URL}",
    "project_id": "${PROJECT_ID}"
  }
}
EOF

# Step 12: Display startup summary
echo
cip_success "🎉 $(cip_get_app_name) started successfully!"
echo
cip_info "📋 Application Details:"
cip_info "  Status: Running (PID: $APP_PID)"
cip_info "  Port: ${PORT}"
cip_info "  Host: ${HOST:-0.0.0.0}"
cip_info "  Environment: ${NODE_ENV}"
echo
cip_info "🔗 Access URLs:"
cip_info "  Portfolio: http://localhost:${PORT}"
cip_info "  Health Check: http://localhost:${PORT}/health"
cip_info "  CloudBox API: ${API_URL}"
echo
cip_info "📊 Monitoring:"
cip_info "  Access Logs: tail -f logs/access.log"
cip_info "  Error Logs: tail -f logs/error.log"
cip_info "  Status Check: ./scripts/cloudbox-status.sh"
cip_info "  Health Check: ./scripts/cloudbox-health.sh"
echo
cip_info "🛑 Management:"
cip_info "  Stop Application: ./scripts/cloudbox-stop.sh"
cip_info "  Restart: ./scripts/cloudbox-stop.sh && ./scripts/cloudbox-start.sh"

# Step 13: Final resource monitoring
cip_debug "Setting up basic monitoring"

# Create a simple monitoring script that runs in background
cat > .cloudbox-monitor.sh << 'EOF'
#!/bin/bash
while true; do
    if [[ -f ".cloudbox.pid" ]]; then
        PID=$(cat .cloudbox.pid)
        if ! kill -0 "$PID" 2>/dev/null; then
            echo "$(date): Application process $PID died unexpectedly" >> logs/error.log
            rm -f .cloudbox.pid .cloudbox-status.json
            break
        fi
    else
        break
    fi
    sleep 30
done
EOF
chmod +x .cloudbox-monitor.sh

# Start monitoring in background
nohup ./.cloudbox-monitor.sh > /dev/null 2>&1 &
MONITOR_PID=$!
echo $MONITOR_PID > .cloudbox-monitor.pid

cip_success "✅ Application monitoring started"

# Step 14: CloudBox integration verification
if [[ -n "${API_URL}" ]] && [[ "${API_URL}" != "http://localhost:3001" ]]; then
    cip_step "Verifying CloudBox integration"
    
    # Test API connectivity
    if cip_health_check "/health" "$(echo ${API_URL} | sed 's|.*://||' | cut -d'/' -f1 | cut -d':' -f2)"; then
        cip_success "✅ CloudBox API is reachable"
    else
        cip_warn "⚠️  CloudBox API connectivity test failed"
        cip_info "This may be normal if API is not yet running"
    fi
fi

echo
cip_success "🚀 SvelteKit Portfolio is now running and ready to serve requests!"
cip_info "⏱️  Startup completed in $(date +%s) seconds"

# Display resource usage
if command -v ps >/dev/null 2>&1; then
    MEMORY_USAGE=$(ps -o pid,pmem,vsz,rss,comm -p $APP_PID --no-headers 2>/dev/null || echo "N/A")
    if [[ "$MEMORY_USAGE" != "N/A" ]]; then
        cip_info "💾 Memory Usage: $MEMORY_USAGE"
    fi
fi

# Optional: Show first few log lines
if [[ -f "logs/access.log" ]]; then
    sleep 1
    if [[ -s "logs/access.log" ]]; then
        cip_info "📄 Recent logs:"
        tail -3 logs/access.log | while read line; do
            cip_info "  $line"
        done
    fi
fi

cip_debug "Startup process completed successfully"