#!/bin/bash
set -e

# Load CloudBox Install Protocol common functions
source "$(dirname "$0")/cloudbox-common.sh"

# Display health check header
cip_header "Health Check for $(cip_get_app_name) v$(cip_get_app_version)"

# Initialize health status variables
ENDPOINT_STATUS="unknown"
RESOURCES_STATUS="unknown"
APP_HEALTH_STATUS="unknown"
CLOUDBOX_STATUS="unknown"
OVERALL_HEALTH="unknown"

# Load environment variables
if [[ -f ".env.production" ]]; then
    source .env.production
    PORT=${PORT:-3000}
    API_URL=${API_URL:-http://localhost:3001}
else
    PORT=${CLOUDBOX_WEB_PORT:-3000}
    API_URL=${CLOUDBOX_API_URL:-http://localhost:3001}
fi

# Step 1: Basic connectivity check
cip_step "Checking basic connectivity"

if ! cip_check_port "$PORT"; then
    cip_error "❌ Port $PORT is not accessible"
    ENDPOINT_STATUS="unreachable"
else
    cip_success "✅ Port $PORT is accessible"
    
    # Step 2: HTTP health endpoint check
    cip_step "Checking HTTP health endpoint"
    
    if cip_health_check "/health" "$PORT"; then
        cip_success "✅ Health endpoint responding"
        ENDPOINT_STATUS="healthy"
        
        # Get detailed health response
        if command -v curl >/dev/null 2>&1; then
            HEALTH_RESPONSE=$(curl -s --connect-timeout 5 --max-time 10 "http://localhost:$PORT/health" 2>/dev/null || echo "")
            if [[ -n "$HEALTH_RESPONSE" ]]; then
                cip_info "Health Response:"
                
                # Pretty print JSON if possible
                if command -v jq >/dev/null 2>&1; then
                    echo "$HEALTH_RESPONSE" | jq . 2>/dev/null || echo "$HEALTH_RESPONSE"
                else
                    echo "$HEALTH_RESPONSE"
                fi
                
                # Extract specific health metrics
                if echo "$HEALTH_RESPONSE" | grep -q '"status"'; then
                    HEALTH_API_STATUS=$(echo "$HEALTH_RESPONSE" | grep -o '"status":"[^"]*"' | cut -d'"' -f4 2>/dev/null || echo "unknown")
                    cip_info "API Health Status: $HEALTH_API_STATUS"
                fi
            fi
        fi
    else
        # Fallback to root endpoint
        if cip_health_check "/" "$PORT"; then
            cip_warn "⚠️  Root endpoint responding but no dedicated health endpoint"
            ENDPOINT_STATUS="partial"
        else
            cip_error "❌ No HTTP endpoints responding"
            ENDPOINT_STATUS="unhealthy"
        fi
    fi
fi

# Step 3: Advanced endpoint testing
if [[ "$ENDPOINT_STATUS" != "unreachable" ]]; then
    cip_step "Performing advanced endpoint tests"
    
    # Test response time
    if command -v curl >/dev/null 2>&1; then
        RESPONSE_TIME=$(curl -o /dev/null -s -w "%{time_total}\n" --connect-timeout 5 "http://localhost:$PORT/" 2>/dev/null || echo "0")
        RESPONSE_TIME_MS=$(echo "$RESPONSE_TIME * 1000" | bc -l 2>/dev/null | cut -d. -f1 || echo "0")
        
        if [[ $RESPONSE_TIME_MS -lt 500 ]]; then
            cip_success "✅ Response time: ${RESPONSE_TIME_MS}ms (excellent)"
        elif [[ $RESPONSE_TIME_MS -lt 2000 ]]; then
            cip_warn "⚠️  Response time: ${RESPONSE_TIME_MS}ms (acceptable)"
        else
            cip_error "❌ Response time: ${RESPONSE_TIME_MS}ms (slow)"
        fi
    fi
    
    # Test different endpoints
    ENDPOINTS_TO_TEST=("/" "/health")
    
    # Add app-specific endpoints if they exist
    if [[ -f "src/routes/api/photos/+server.js" ]]; then
        ENDPOINTS_TO_TEST+=("/api/photos")
    fi
    
    for endpoint in "${ENDPOINTS_TO_TEST[@]}"; do
        if command -v curl >/dev/null 2>&1; then
            HTTP_STATUS=$(curl -o /dev/null -s -w "%{http_code}" --connect-timeout 5 "http://localhost:$PORT$endpoint" 2>/dev/null || echo "000")
            
            case $HTTP_STATUS in
                200|201|204)
                    cip_debug "✅ $endpoint: HTTP $HTTP_STATUS (OK)"
                    ;;
                404)
                    cip_debug "ℹ️  $endpoint: HTTP $HTTP_STATUS (Not Found - may be normal)"
                    ;;
                500|502|503|504)
                    cip_warn "⚠️  $endpoint: HTTP $HTTP_STATUS (Server Error)"
                    ;;
                000)
                    cip_debug "❌ $endpoint: Connection failed"
                    ;;
                *)
                    cip_debug "ℹ️  $endpoint: HTTP $HTTP_STATUS"
                    ;;
            esac
        fi
    done
fi

# Step 4: System resource health
cip_step "Checking system resource health"

RESOURCE_ISSUES=()

# Memory check
if command -v free >/dev/null 2>&1; then
    MEMORY_TOTAL=$(free -m | awk 'NR==2{print $2}')
    MEMORY_USED=$(free -m | awk 'NR==2{print $3}')
    MEMORY_PERCENT=$(free | awk 'FNR==2{printf "%.0f", $3/$2*100}')
    
    if [[ $MEMORY_PERCENT -lt 70 ]]; then
        cip_success "✅ Memory usage: ${MEMORY_PERCENT}% (${MEMORY_USED}MB/${MEMORY_TOTAL}MB)"
        MEMORY_STATUS="healthy"
    elif [[ $MEMORY_PERCENT -lt 85 ]]; then
        cip_warn "⚠️  Memory usage: ${MEMORY_PERCENT}% (${MEMORY_USED}MB/${MEMORY_TOTAL}MB)"
        MEMORY_STATUS="warning"
        RESOURCE_ISSUES+=("high_memory")
    else
        cip_error "❌ Critical memory usage: ${MEMORY_PERCENT}% (${MEMORY_USED}MB/${MEMORY_TOTAL}MB)"
        MEMORY_STATUS="critical"
        RESOURCE_ISSUES+=("critical_memory")
    fi
else
    MEMORY_STATUS="unknown"
fi

# Disk space check
if command -v df >/dev/null 2>&1; then
    DISK_TOTAL=$(df -h . | awk 'NR==2{print $2}')
    DISK_USED=$(df -h . | awk 'NR==2{print $3}')
    DISK_PERCENT=$(df . | awk 'FNR==2{print $5}' | sed 's/%//')
    
    if [[ $DISK_PERCENT -lt 80 ]]; then
        cip_success "✅ Disk usage: ${DISK_PERCENT}% (${DISK_USED}/${DISK_TOTAL})"
        DISK_STATUS="healthy"
    elif [[ $DISK_PERCENT -lt 90 ]]; then
        cip_warn "⚠️  Disk usage: ${DISK_PERCENT}% (${DISK_USED}/${DISK_TOTAL})"
        DISK_STATUS="warning"
        RESOURCE_ISSUES+=("high_disk")
    else
        cip_error "❌ Critical disk usage: ${DISK_PERCENT}% (${DISK_USED}/${DISK_TOTAL})"
        DISK_STATUS="critical"
        RESOURCE_ISSUES+=("critical_disk")
    fi
else
    DISK_STATUS="unknown"
fi

# Load average check (Linux/macOS)
if command -v uptime >/dev/null 2>&1; then
    LOAD_AVERAGE=$(uptime | awk -F'load average:' '{print $2}' | awk '{print $1}' | sed 's/,//')
    CPU_CORES=$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo "1")
    
    if command -v bc >/dev/null 2>&1; then
        LOAD_RATIO=$(echo "scale=2; $LOAD_AVERAGE / $CPU_CORES" | bc 2>/dev/null || echo "0")
        
        if (( $(echo "$LOAD_RATIO < 0.7" | bc -l 2>/dev/null || echo "1") )); then
            cip_success "✅ Load average: $LOAD_AVERAGE (${LOAD_RATIO} per core)"
            LOAD_STATUS="healthy"
        elif (( $(echo "$LOAD_RATIO < 1.0" | bc -l 2>/dev/null || echo "0") )); then
            cip_warn "⚠️  Load average: $LOAD_AVERAGE (${LOAD_RATIO} per core)"
            LOAD_STATUS="warning"
        else
            cip_error "❌ High load average: $LOAD_AVERAGE (${LOAD_RATIO} per core)"
            LOAD_STATUS="critical"
            RESOURCE_ISSUES+=("high_load")
        fi
    else
        cip_info "ℹ️  Load average: $LOAD_AVERAGE ($CPU_CORES cores available)"
        LOAD_STATUS="unknown"
    fi
fi

# Determine overall resource status
if [[ ${#RESOURCE_ISSUES[@]} -eq 0 ]]; then
    RESOURCES_STATUS="healthy"
    cip_success "✅ System resources are healthy"
elif [[ " ${RESOURCE_ISSUES[@]} " =~ " critical_" ]]; then
    RESOURCES_STATUS="critical"
    cip_error "❌ Critical resource issues detected"
else
    RESOURCES_STATUS="warning"
    cip_warn "⚠️  Resource warnings detected"
fi

# Step 5: Application-specific health checks
cip_step "Running application-specific health checks"

# Check if npm health script exists
if [[ -f "package.json" ]] && grep -q "healthcheck" package.json 2>/dev/null; then
    cip_info "Running custom health checks..."
    
    if npm run healthcheck --silent 2>/dev/null; then
        cip_success "✅ Application health checks passed"
        APP_HEALTH_STATUS="healthy"
    else
        cip_error "❌ Application health checks failed"
        APP_HEALTH_STATUS="unhealthy"
        
        # Show npm healthcheck output for debugging
        cip_info "Health check output:"
        npm run healthcheck 2>&1 | head -5 | while read line; do
            cip_info "  $line"
        done
    fi
else
    # Perform basic application health checks
    cip_info "Performing basic application health checks..."
    
    # Check critical files
    CRITICAL_FILES=("package.json" "build/index.js" ".env.production")
    MISSING_FILES=()
    
    for file in "${CRITICAL_FILES[@]}"; do
        if [[ ! -f "$file" ]]; then
            MISSING_FILES+=("$file")
        fi
    done
    
    if [[ ${#MISSING_FILES[@]} -eq 0 ]]; then
        cip_success "✅ All critical files present"
        
        # Check if build is recent
        if [[ -f "build/index.js" ]]; then
            BUILD_AGE=$(find build/index.js -mtime +7 2>/dev/null | wc -l)
            if [[ $BUILD_AGE -gt 0 ]]; then
                cip_warn "⚠️  Build is older than 7 days - consider rebuilding"
            else
                cip_debug "Build is recent"
            fi
        fi
        
        APP_HEALTH_STATUS="healthy"
    else
        cip_error "❌ Missing critical files: ${MISSING_FILES[*]}"
        APP_HEALTH_STATUS="unhealthy"
    fi
fi

# Step 6: Database/external service connectivity
if [[ -n "${DATABASE_URL}" ]]; then
    cip_step "Checking database connectivity"
    
    # This is a basic check - in a real app, you'd want to test actual DB connection
    cip_info "Database URL configured: ${DATABASE_URL}"
    
    # Extract database type and host
    DB_TYPE=$(echo "$DATABASE_URL" | cut -d: -f1)
    DB_HOST=$(echo "$DATABASE_URL" | sed 's|.*@\([^:/]*\).*|\1|')
    DB_PORT=$(echo "$DATABASE_URL" | sed 's|.*:\([0-9]*\)/.*|\1|')
    
    if [[ "$DB_PORT" == "$DATABASE_URL" ]]; then
        # Default ports for common databases
        case $DB_TYPE in
            postgres|postgresql) DB_PORT=5432 ;;
            mysql) DB_PORT=3306 ;;
            mongodb) DB_PORT=27017 ;;
            *) DB_PORT=5432 ;;
        esac
    fi
    
    if command -v nc >/dev/null 2>&1; then
        if nc -z "$DB_HOST" "$DB_PORT" 2>/dev/null; then
            cip_success "✅ Database server is reachable ($DB_TYPE://$DB_HOST:$DB_PORT)"
            DB_STATUS="reachable"
        else
            cip_error "❌ Cannot reach database server ($DB_TYPE://$DB_HOST:$DB_PORT)"
            DB_STATUS="unreachable"
        fi
    else
        cip_info "Cannot test database connectivity (nc not available)"
        DB_STATUS="unknown"
    fi
else
    DB_STATUS="not_configured"
fi

# Step 7: CloudBox API integration health
cip_step "Checking CloudBox API integration"

if [[ -n "$API_URL" ]] && [[ "$API_URL" != "http://localhost:3001" ]]; then
    # Extract API host and port
    API_HOST=$(echo "$API_URL" | sed 's|.*://||' | cut -d'/' -f1 | cut -d':' -f1)
    API_PORT=$(echo "$API_URL" | sed 's|.*://||' | cut -d'/' -f1 | cut -d':' -f2)
    
    if [[ "$API_PORT" == "$API_HOST" ]]; then
        if [[ "$API_URL" == https* ]]; then
            API_PORT="443"
        else
            API_PORT="80"
        fi
    fi
    
    cip_info "Testing CloudBox API: $API_URL"
    
    if command -v curl >/dev/null 2>&1; then
        # Test basic connectivity
        API_HTTP_STATUS=$(curl -o /dev/null -s -w "%{http_code}" --connect-timeout 10 "${API_URL}/health" 2>/dev/null || echo "000")
        
        case $API_HTTP_STATUS in
            200|201|204)
                cip_success "✅ CloudBox API is healthy (HTTP $API_HTTP_STATUS)"
                CLOUDBOX_STATUS="healthy"
                ;;
            404)
                cip_warn "⚠️  CloudBox API accessible but no health endpoint (HTTP $API_HTTP_STATUS)"
                CLOUDBOX_STATUS="partial"
                ;;
            500|502|503|504)
                cip_error "❌ CloudBox API has server errors (HTTP $API_HTTP_STATUS)"
                CLOUDBOX_STATUS="unhealthy"
                ;;
            000)
                cip_error "❌ Cannot connect to CloudBox API"
                CLOUDBOX_STATUS="unreachable"
                ;;
            *)
                cip_warn "⚠️  CloudBox API returned HTTP $API_HTTP_STATUS"
                CLOUDBOX_STATUS="issues"
                ;;
        esac
        
        # Test API response time
        if [[ "$CLOUDBOX_STATUS" != "unreachable" ]]; then
            API_RESPONSE_TIME=$(curl -o /dev/null -s -w "%{time_total}\n" --connect-timeout 10 "$API_URL" 2>/dev/null || echo "0")
            API_RESPONSE_TIME_MS=$(echo "$API_RESPONSE_TIME * 1000" | bc -l 2>/dev/null | cut -d. -f1 || echo "0")
            
            if [[ $API_RESPONSE_TIME_MS -lt 1000 ]]; then
                cip_success "✅ CloudBox API response time: ${API_RESPONSE_TIME_MS}ms"
            elif [[ $API_RESPONSE_TIME_MS -lt 5000 ]]; then
                cip_warn "⚠️  CloudBox API response time: ${API_RESPONSE_TIME_MS}ms (slow)"
            else
                cip_error "❌ CloudBox API response time: ${API_RESPONSE_TIME_MS}ms (very slow)"
            fi
        fi
    elif command -v nc >/dev/null 2>&1; then
        if nc -z "$API_HOST" "$API_PORT" 2>/dev/null; then
            cip_success "✅ CloudBox API server is reachable"
            CLOUDBOX_STATUS="reachable"
        else
            cip_error "❌ Cannot reach CloudBox API server"
            CLOUDBOX_STATUS="unreachable"
        fi
    else
        cip_warn "Cannot test CloudBox API connectivity (no suitable tools)"
        CLOUDBOX_STATUS="unknown"
    fi
else
    cip_info "CloudBox API not configured or using local development server"
    CLOUDBOX_STATUS="local"
fi

# Step 8: Log file health analysis
cip_step "Analyzing log file health"

LOG_HEALTH_ISSUES=()

# Check error log
if [[ -f "logs/error.log" ]]; then
    ERROR_COUNT=$(wc -l < logs/error.log 2>/dev/null || echo "0")
    ERROR_LOG_SIZE=$(stat -f%z "logs/error.log" 2>/dev/null || stat -c%s "logs/error.log" 2>/dev/null || echo "0")
    
    if [[ $ERROR_COUNT -eq 0 ]]; then
        cip_success "✅ No errors in error log"
    elif [[ $ERROR_COUNT -lt 10 ]]; then
        cip_info "ℹ️  $ERROR_COUNT errors in error log (acceptable)"
    elif [[ $ERROR_COUNT -lt 50 ]]; then
        cip_warn "⚠️  $ERROR_COUNT errors in error log (review recommended)"
        LOG_HEALTH_ISSUES+=("moderate_errors")
    else
        cip_error "❌ $ERROR_COUNT errors in error log (immediate attention required)"
        LOG_HEALTH_ISSUES+=("high_errors")
        
        # Show recent critical errors
        CRITICAL_ERRORS=$(grep -i "critical\|fatal\|panic" logs/error.log 2>/dev/null | tail -3 || echo "")
        if [[ -n "$CRITICAL_ERRORS" ]]; then
            cip_error "Recent critical errors:"
            echo "$CRITICAL_ERRORS" | while read line; do
                cip_error "  $(echo "$line" | cut -c1-100)"
            done
        fi
    fi
    
    # Check log file size
    if [[ $ERROR_LOG_SIZE -gt 10485760 ]]; then  # 10MB
        cip_warn "⚠️  Error log file is large ($(( ERROR_LOG_SIZE / 1024 / 1024 ))MB) - consider rotation"
        LOG_HEALTH_ISSUES+=("large_error_log")
    fi
else
    cip_debug "No error log file found"
fi

# Check access log
if [[ -f "logs/access.log" ]]; then
    ACCESS_COUNT=$(wc -l < logs/access.log 2>/dev/null || echo "0")
    ACCESS_LOG_SIZE=$(stat -f%z "logs/access.log" 2>/dev/null || stat -c%s "logs/access.log" 2>/dev/null || echo "0")
    
    cip_info "📊 $ACCESS_COUNT requests in access log"
    
    # Check for recent activity
    if [[ $ACCESS_COUNT -gt 0 ]]; then
        RECENT_ACTIVITY=$(tail -1 logs/access.log 2>/dev/null | head -c 50 || echo "")
        if [[ -n "$RECENT_ACTIVITY" ]]; then
            cip_info "Recent activity: $RECENT_ACTIVITY..."
        fi
    fi
    
    # Check log file size
    if [[ $ACCESS_LOG_SIZE -gt 52428800 ]]; then  # 50MB
        cip_warn "⚠️  Access log file is large ($(( ACCESS_LOG_SIZE / 1024 / 1024 ))MB) - consider rotation"
        LOG_HEALTH_ISSUES+=("large_access_log")
    fi
fi

# Determine log health status
if [[ ${#LOG_HEALTH_ISSUES[@]} -eq 0 ]]; then
    LOG_HEALTH_STATUS="healthy"
elif [[ " ${LOG_HEALTH_ISSUES[@]} " =~ " high_errors " ]]; then
    LOG_HEALTH_STATUS="critical"
else
    LOG_HEALTH_STATUS="warning"
fi

# Step 9: Security health checks
cip_step "Performing security health checks"

SECURITY_ISSUES=()

# Check file permissions
if [[ -f ".env.production" ]]; then
    ENV_PERMS=$(stat -c%a .env.production 2>/dev/null || stat -f%A .env.production 2>/dev/null || echo "644")
    if [[ "$ENV_PERMS" == "644" ]] || [[ "$ENV_PERMS" == "600" ]]; then
        cip_success "✅ Environment file permissions are secure"
    else
        cip_warn "⚠️  Environment file permissions may be too open ($ENV_PERMS)"
        SECURITY_ISSUES+=("env_permissions")
    fi
fi

# Check for sensitive information in logs
if [[ -f "logs/access.log" ]]; then
    SENSITIVE_PATTERNS=("password" "secret" "token" "api_key")
    for pattern in "${SENSITIVE_PATTERNS[@]}"; do
        if grep -qi "$pattern" logs/access.log 2>/dev/null; then
            cip_warn "⚠️  Potential sensitive information in access logs: $pattern"
            SECURITY_ISSUES+=("sensitive_logs")
            break
        fi
    done
fi

# Check Node.js version for known vulnerabilities (basic check)
if command -v node >/dev/null 2>&1; then
    NODE_VERSION=$(node --version | sed 's/v//')
    NODE_MAJOR=$(echo "$NODE_VERSION" | cut -d. -f1)
    
    if [[ $NODE_MAJOR -lt 18 ]]; then
        cip_warn "⚠️  Node.js version $NODE_VERSION may have security vulnerabilities"
        SECURITY_ISSUES+=("old_nodejs")
    else
        cip_success "✅ Node.js version $NODE_VERSION is reasonably current"
    fi
fi

# Determine security status
if [[ ${#SECURITY_ISSUES[@]} -eq 0 ]]; then
    SECURITY_STATUS="healthy"
else
    SECURITY_STATUS="issues"
fi

# Step 10: Determine overall health status
cip_step "Determining overall health status"

HEALTH_SCORES=()

# Assign numeric scores for each component (0-100)
case $ENDPOINT_STATUS in
    "healthy") HEALTH_SCORES+=(100) ;;
    "partial") HEALTH_SCORES+=(70) ;;
    "unhealthy") HEALTH_SCORES+=(30) ;;
    "unreachable") HEALTH_SCORES+=(0) ;;
    *) HEALTH_SCORES+=(50) ;;
esac

case $RESOURCES_STATUS in
    "healthy") HEALTH_SCORES+=(100) ;;
    "warning") HEALTH_SCORES+=(60) ;;
    "critical") HEALTH_SCORES+=(20) ;;
    *) HEALTH_SCORES+=(50) ;;
esac

case $APP_HEALTH_STATUS in
    "healthy") HEALTH_SCORES+=(100) ;;
    "unhealthy") HEALTH_SCORES+=(0) ;;
    *) HEALTH_SCORES+=(50) ;;
esac

case $CLOUDBOX_STATUS in
    "healthy") HEALTH_SCORES+=(100) ;;
    "partial"|"reachable"|"local") HEALTH_SCORES+=(80) ;;
    "issues") HEALTH_SCORES+=(40) ;;
    "unhealthy"|"unreachable") HEALTH_SCORES+=(20) ;;
    *) HEALTH_SCORES+=(50) ;;
esac

# Calculate overall health score
TOTAL_SCORE=0
for score in "${HEALTH_SCORES[@]}"; do
    TOTAL_SCORE=$((TOTAL_SCORE + score))
done
OVERALL_SCORE=$((TOTAL_SCORE / ${#HEALTH_SCORES[@]}))

# Determine overall health status
if [[ $OVERALL_SCORE -ge 90 ]]; then
    OVERALL_HEALTH="excellent"
    cip_success "✅ Overall Health: EXCELLENT ($OVERALL_SCORE/100)"
elif [[ $OVERALL_SCORE -ge 75 ]]; then
    OVERALL_HEALTH="good"
    cip_success "✅ Overall Health: GOOD ($OVERALL_SCORE/100)"
elif [[ $OVERALL_SCORE -ge 60 ]]; then
    OVERALL_HEALTH="fair"
    cip_warn "⚠️  Overall Health: FAIR ($OVERALL_SCORE/100)"
elif [[ $OVERALL_SCORE -ge 40 ]]; then
    OVERALL_HEALTH="poor"
    cip_error "❌ Overall Health: POOR ($OVERALL_SCORE/100)"
else
    OVERALL_HEALTH="critical"
    cip_error "❌ Overall Health: CRITICAL ($OVERALL_SCORE/100)"
fi

# Step 11: Create comprehensive health report
cat > .cloudbox-health-report.json << EOF
{
  "timestamp": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "app_name": "$(cip_get_app_name)",
  "app_version": "$(cip_get_app_version)",
  "overall_health": "$OVERALL_HEALTH",
  "health_score": $OVERALL_SCORE,
  "endpoints": {
    "status": "$ENDPOINT_STATUS",
    "port": $PORT,
    "response_time_ms": ${RESPONSE_TIME_MS:-0}
  },
  "resources": {
    "status": "$RESOURCES_STATUS",
    "memory": {
      "status": "${MEMORY_STATUS:-unknown}",
      "percent": ${MEMORY_PERCENT:-0},
      "used_mb": ${MEMORY_USED:-0},
      "total_mb": ${MEMORY_TOTAL:-0}
    },
    "disk": {
      "status": "${DISK_STATUS:-unknown}",
      "percent": ${DISK_PERCENT:-0},
      "used": "${DISK_USED:-unknown}",
      "total": "${DISK_TOTAL:-unknown}"
    },
    "load": {
      "status": "${LOAD_STATUS:-unknown}",
      "average": "${LOAD_AVERAGE:-unknown}",
      "cores": ${CPU_CORES:-1}
    }
  },
  "application": {
    "status": "$APP_HEALTH_STATUS"
  },
  "external_services": {
    "cloudbox_api": {
      "status": "$CLOUDBOX_STATUS",
      "url": "$API_URL",
      "response_time_ms": ${API_RESPONSE_TIME_MS:-0}
    },
    "database": {
      "status": "${DB_STATUS:-not_configured}",
      "url": "${DATABASE_URL:-not_configured}"
    }
  },
  "logs": {
    "status": "${LOG_HEALTH_STATUS:-healthy}",
    "error_count": ${ERROR_COUNT:-0},
    "access_count": ${ACCESS_COUNT:-0}
  },
  "security": {
    "status": "${SECURITY_STATUS:-unknown}",
    "issues": $(printf '%s\n' "${SECURITY_ISSUES[@]}" | jq -R . | jq -s . 2>/dev/null || echo '[]')
  },
  "recommendations": []
}
EOF

# Step 12: Generate recommendations
RECOMMENDATIONS=()

if [[ "$ENDPOINT_STATUS" == "unreachable" ]]; then
    RECOMMENDATIONS+=("Application is not accessible - check if it's running and bound to correct port")
fi

if [[ "$RESOURCES_STATUS" == "critical" ]]; then
    RECOMMENDATIONS+=("Critical resource issues detected - restart application or upgrade hardware")
fi

if [[ "$APP_HEALTH_STATUS" == "unhealthy" ]]; then
    RECOMMENDATIONS+=("Application health checks failing - review error logs and configuration")
fi

if [[ "$CLOUDBOX_STATUS" == "unreachable" ]]; then
    RECOMMENDATIONS+=("CloudBox API is unreachable - check network connectivity and API server status")
fi

if [[ ${#LOG_HEALTH_ISSUES[@]} -gt 0 ]]; then
    RECOMMENDATIONS+=("Log file issues detected - consider log rotation and error investigation")
fi

if [[ ${#SECURITY_ISSUES[@]} -gt 0 ]]; then
    RECOMMENDATIONS+=("Security issues detected - review file permissions and update dependencies")
fi

# Step 13: Display comprehensive health summary
echo
cip_info "🏥 Comprehensive Health Report"
echo "=================================================="
echo
cip_info "📊 Overall Health Score: $OVERALL_SCORE/100 ($OVERALL_HEALTH)"
echo
cip_info "🌐 Endpoint Health:"
echo "   Status: $ENDPOINT_STATUS"
echo "   Port: $PORT"
if [[ -n "${RESPONSE_TIME_MS}" ]] && [[ "${RESPONSE_TIME_MS}" != "0" ]]; then
    echo "   Response Time: ${RESPONSE_TIME_MS}ms"
fi
echo
cip_info "💾 Resource Health:"
echo "   Overall: $RESOURCES_STATUS"
if [[ -n "${MEMORY_STATUS}" ]]; then
    echo "   Memory: ${MEMORY_STATUS} (${MEMORY_PERCENT}%)"
fi
if [[ -n "${DISK_STATUS}" ]]; then
    echo "   Disk: ${DISK_STATUS} (${DISK_PERCENT}%)"
fi
if [[ -n "${LOAD_STATUS}" ]]; then
    echo "   Load: ${LOAD_STATUS} (${LOAD_AVERAGE})"
fi
echo
cip_info "🔗 External Services:"
echo "   CloudBox API: $CLOUDBOX_STATUS"
if [[ -n "${DB_STATUS}" ]] && [[ "${DB_STATUS}" != "not_configured" ]]; then
    echo "   Database: $DB_STATUS"
fi
echo
cip_info "📄 Log Health:"
echo "   Status: ${LOG_HEALTH_STATUS}"
echo "   Errors: ${ERROR_COUNT:-0}"
echo "   Access: ${ACCESS_COUNT:-0}"
echo
cip_info "🔒 Security:"
echo "   Status: ${SECURITY_STATUS}"
if [[ ${#SECURITY_ISSUES[@]} -gt 0 ]]; then
    echo "   Issues: ${#SECURITY_ISSUES[@]} detected"
fi

# Step 14: Display recommendations
if [[ ${#RECOMMENDATIONS[@]} -gt 0 ]]; then
    echo
    cip_info "💡 Health Recommendations:"
    for rec in "${RECOMMENDATIONS[@]}"; do
        echo "   • $rec"
    done
fi

echo
cip_info "📋 Health report saved to: .cloudbox-health-report.json"

# Step 15: Exit with appropriate health code
case $OVERALL_HEALTH in
    "excellent"|"good")
        echo
        cip_success "🎉 Application is healthy and ready to serve requests!"
        exit 0
        ;;
    "fair")
        echo
        cip_warn "⚠️  Application health is fair - monitor and address issues"
        exit 1
        ;;
    "poor")
        echo
        cip_error "❌ Application health is poor - immediate attention recommended"
        exit 2
        ;;
    "critical")
        echo
        cip_error "💀 Application health is critical - urgent intervention required"
        exit 3
        ;;
    *)
        echo
        cip_warn "❓ Unable to determine application health status"
        exit 4
        ;;
esac