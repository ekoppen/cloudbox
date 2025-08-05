#!/bin/bash

# CloudBox Install Protocol - Common Functions v1.0
# This file provides reusable functions for all CIP scripts

# Colors for output
readonly CIP_GREEN='\033[0;32m'
readonly CIP_BLUE='\033[0;34m'
readonly CIP_YELLOW='\033[1;33m'
readonly CIP_RED='\033[0;31m'
readonly CIP_CYAN='\033[0;36m'
readonly CIP_BOLD='\033[1m'
readonly CIP_NC='\033[0m' # No Color

# Logging functions
cip_header() {
    echo -e "\n${CIP_BLUE}================================================${CIP_NC}"
    echo -e "${CIP_BLUE}CloudBox Install Protocol v1.0${CIP_NC}"
    echo -e "${CIP_BLUE}$1${CIP_NC}"
    echo -e "${CIP_BLUE}================================================${CIP_NC}\n"
}

cip_step() {
    echo -e "${CIP_CYAN}▶ [CIP]${CIP_NC} $1..."
}

cip_success() {
    echo -e "${CIP_GREEN}✅ [CIP]${CIP_NC} $1"
}

cip_info() {
    echo -e "${CIP_BLUE}ℹ️  [CIP]${CIP_NC} $1"
}

cip_warn() {
    echo -e "${CIP_YELLOW}⚠️  [CIP]${CIP_NC} $1"
}

cip_error() {
    echo -e "${CIP_RED}❌ [CIP]${CIP_NC} $1" >&2
}

cip_fatal() {
    cip_error "$1"
    echo -e "${CIP_RED}💀 [CIP] Fatal error - aborting deployment${CIP_NC}" >&2
    exit 1
}

cip_debug() {
    if [[ "${CLOUDBOX_DEBUG}" == "true" ]]; then
        echo -e "${CIP_CYAN}🐛 [DEBUG]${CIP_NC} $1" >&2
    fi
}

# Environment validation
cip_require_env() {
    local missing_vars=()
    
    for var in "$@"; do
        if [[ -z "${!var}" ]]; then
            missing_vars+=("$var")
        else
            cip_debug "Environment variable $var is set: ${!var}"
        fi
    done
    
    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        cip_error "Missing required environment variables:"
        for var in "${missing_vars[@]}"; do
            echo "   - $var"
        done
        cip_fatal "Please ensure CloudBox provides all required environment variables"
    fi
}

# Command availability check
cip_check_command() {
    local cmd=$1
    local error_msg=${2:-"Command '$cmd' is required but not found"}
    
    if ! command -v "$cmd" >/dev/null 2>&1; then
        cip_fatal "$error_msg"
    else
        cip_debug "Command '$cmd' is available"
    fi
}

# App manifest functions
cip_get_app_name() {
    if [[ -f "cloudbox.json" ]]; then
        if command -v jq >/dev/null 2>&1; then
            jq -r '.name // "Unknown App"' cloudbox.json 2>/dev/null || echo "Unknown App"
        else
            # Fallback parsing without jq
            grep -o '"name"[[:space:]]*:[[:space:]]*"[^"]*"' cloudbox.json | cut -d'"' -f4 || echo "Unknown App"
        fi
    else
        echo "Unknown App"
    fi
}

cip_get_app_version() {
    if [[ -f "cloudbox.json" ]]; then
        if command -v jq >/dev/null 2>&1; then
            jq -r '.version // "1.0.0"' cloudbox.json 2>/dev/null || echo "1.0.0"
        else
            # Fallback parsing without jq
            grep -o '"version"[[:space:]]*:[[:space:]]*"[^"]*"' cloudbox.json | cut -d'"' -f4 || echo "1.0.0"
        fi
    else
        echo "1.0.0"
    fi
}

cip_get_app_type() {
    if [[ -f "cloudbox.json" ]]; then
        if command -v jq >/dev/null 2>&1; then
            jq -r '.type // "frontend"' cloudbox.json 2>/dev/null || echo "frontend"
        else
            # Fallback parsing without jq
            grep -o '"type"[[:space:]]*:[[:space:]]*"[^"]*"' cloudbox.json | cut -d'"' -f4 || echo "frontend"
        fi
    else
        echo "frontend"
    fi
}

# Port management
cip_check_port() {
    local port=$1
    
    # Check multiple ways to ensure port is available
    if command -v netstat >/dev/null 2>&1; then
        if netstat -tuln 2>/dev/null | grep ":$port " >/dev/null; then
            cip_debug "Port $port is in use (netstat)"
            return 1  # Port in use
        fi
    fi
    
    if command -v ss >/dev/null 2>&1; then
        if ss -tuln 2>/dev/null | grep ":$port " >/dev/null; then
            cip_debug "Port $port is in use (ss)"
            return 1  # Port in use
        fi
    fi
    
    # Try to bind to the port as final check
    if command -v nc >/dev/null 2>&1; then
        if ! nc -z localhost $port 2>/dev/null; then
            cip_debug "Port $port is available (nc test)"
            return 0  # Port available
        else
            cip_debug "Port $port is in use (nc test)"
            return 1  # Port in use
        fi
    fi
    
    cip_debug "Port $port availability check completed"
    return 0  # Assume available if no tools work
}

cip_wait_for_port() {
    local port=$1
    local timeout=${2:-30}
    local count=0
    
    cip_step "Waiting for service on port $port (timeout: ${timeout}s)"
    
    while [[ $count -lt $timeout ]]; do
        if command -v nc >/dev/null 2>&1; then
            if nc -z localhost $port 2>/dev/null; then
                cip_success "Service is responding on port $port"
                return 0
            fi
        elif command -v curl >/dev/null 2>&1; then
            if curl -s --connect-timeout 1 "http://localhost:$port" >/dev/null 2>&1; then
                cip_success "Service is responding on port $port"
                return 0
            fi
        fi
        
        sleep 1
        ((count++))
        
        # Progress indicator
        if [[ $((count % 5)) -eq 0 ]]; then
            cip_debug "Still waiting for port $port... (${count}/${timeout}s)"
        fi
    done
    
    cip_error "Service not responding on port $port after ${timeout}s"
    return 1
}

# Docker functions
cip_docker_setup() {
    if ! command -v docker >/dev/null 2>&1; then
        cip_warn "Docker not found - skipping container setup"
        return 0
    fi
    
    cip_step "Setting up Docker monitoring"
    
    # Create monitoring compose file if it doesn't exist
    if [[ ! -f "docker-compose.cloudbox.yml" ]]; then
        cip_info "Generating Docker monitoring configuration"
        cip_generate_docker_compose
    fi
    
    # Start monitoring containers
    if docker-compose -f docker-compose.cloudbox.yml up -d monitoring 2>/dev/null; then
        cip_success "Docker monitoring started"
    else
        cip_warn "Failed to start Docker monitoring - continuing without it"
    fi
}

cip_generate_docker_compose() {
    cat > docker-compose.cloudbox.yml << 'EOF'
version: '3.8'

services:
  app:
    build: .
    ports:
      - "${CLOUDBOX_WEB_PORT:-3000}:${CLOUDBOX_WEB_PORT:-3000}"
    environment:
      - NODE_ENV=production
      - PORT=${CLOUDBOX_WEB_PORT:-3000}
      - API_URL=${CLOUDBOX_API_URL}
      - ORIGIN=http://localhost:${CLOUDBOX_WEB_PORT:-3000}
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:${CLOUDBOX_WEB_PORT:-3000}/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s
    labels:
      - "cloudbox.enable=true"
      - "cloudbox.project=${CLOUDBOX_PROJECT_SLUG:-unknown}"
      - "cloudbox.type=frontend"

  monitoring:
    image: prom/node-exporter:latest
    ports:
      - "9100:9100"
    restart: unless-stopped
    command:
      - '--path.rootfs=/host'
    volumes:
      - '/:/host:ro,rslave'
    labels:
      - "cloudbox.enable=true"
      - "cloudbox.service=monitoring"
      - "cloudbox.project=${CLOUDBOX_PROJECT_SLUG:-unknown}"
EOF
    
    cip_success "Generated docker-compose.cloudbox.yml"
}

# Health check functions
cip_health_check() {
    local endpoint=${1:-"/health"}
    local port=${2:-${CLOUDBOX_WEB_PORT:-3000}}
    local url="http://localhost:$port$endpoint"
    
    cip_debug "Health checking endpoint: $url"
    
    if command -v curl >/dev/null 2>&1; then
        if curl -f -s --connect-timeout 5 --max-time 10 "$url" >/dev/null 2>&1; then
            cip_debug "Health check passed via curl"
            return 0
        fi
    elif command -v wget >/dev/null 2>&1; then
        if wget -q --timeout=10 --tries=1 --spider "$url" 2>/dev/null; then
            cip_debug "Health check passed via wget"
            return 0
        fi
    fi
    
    cip_debug "Health check failed for $url"
    return 1
}

# Process management functions
cip_get_pid() {
    local app_name=${1:-$(basename "$PWD")}
    
    if [[ -f ".cloudbox.pid" ]]; then
        local pid=$(cat .cloudbox.pid 2>/dev/null)
        if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
            echo "$pid"
            return 0
        else
            rm -f .cloudbox.pid
        fi
    fi
    
    # Fallback: find by process pattern
    pgrep -f "node.*$app_name" | head -1
}

cip_stop_process() {
    local pid=$(cip_get_pid)
    
    if [[ -n "$pid" ]]; then
        cip_step "Stopping process $pid"
        
        # Graceful shutdown
        if kill "$pid" 2>/dev/null; then
            # Wait up to 10 seconds for graceful shutdown
            local count=0
            while [[ $count -lt 10 ]] && kill -0 "$pid" 2>/dev/null; do
                sleep 1
                ((count++))
            done
            
            # Force kill if still running
            if kill -0 "$pid" 2>/dev/null; then
                cip_warn "Process didn't shut down gracefully, force killing..."
                kill -9 "$pid" 2>/dev/null
                sleep 1
            fi
            
            if ! kill -0 "$pid" 2>/dev/null; then
                cip_success "Process stopped successfully"
                rm -f .cloudbox.pid
                return 0
            fi
        fi
    fi
    
    # Fallback: kill by name pattern
    local app_name=$(basename "$PWD")
    if pgrep -f "node.*$app_name" >/dev/null; then
        cip_step "Stopping processes by pattern"
        pkill -f "node.*$app_name" 2>/dev/null
        sleep 2
        
        if ! pgrep -f "node.*$app_name" >/dev/null; then
            cip_success "Processes stopped by pattern"
            return 0
        fi
    fi
    
    cip_info "No running processes found to stop"
    return 0
}

# Resource monitoring
cip_check_resources() {
    local warn_memory=${1:-80}
    local warn_disk=${2:-85}
    
    # Memory check
    if command -v free >/dev/null 2>&1; then
        local memory_percent=$(free | awk 'FNR==2{printf "%.0f", $3/$2*100}' 2>/dev/null || echo "0")
        if [[ $memory_percent -gt $warn_memory ]]; then
            cip_warn "High memory usage: ${memory_percent}%"
        else
            cip_debug "Memory usage: ${memory_percent}%"
        fi
    fi
    
    # Disk check
    if command -v df >/dev/null 2>&1; then
        local disk_percent=$(df . | awk 'FNR==2{print $5}' | sed 's/%//' 2>/dev/null || echo "0")
        if [[ $disk_percent -gt $warn_disk ]]; then
            cip_warn "High disk usage: ${disk_percent}%"
        else
            cip_debug "Disk usage: ${disk_percent}%"
        fi
    fi
}

# Cleanup functions
cip_cleanup() {
    cip_debug "Running cleanup tasks"
    
    # Remove temporary files
    rm -rf /tmp/cloudbox-* 2>/dev/null || true
    rm -rf .tmp-* 2>/dev/null || true
    
    # Clean up npm cache if space is low
    if command -v npm >/dev/null 2>&1; then
        local disk_percent=$(df . | awk 'FNR==2{print $5}' | sed 's/%//' 2>/dev/null || echo "0")
        if [[ $disk_percent -gt 90 ]]; then
            cip_step "Cleaning npm cache due to low disk space"
            npm cache clean --force 2>/dev/null || true
        fi
    fi
    
    cip_debug "Cleanup completed"
}

# Install protocol validation
cip_validate_protocol() {
    cip_step "Validating CloudBox Install Protocol compliance"
    
    if [[ ! -f "cloudbox.json" ]]; then
        cip_fatal "cloudbox.json not found - this app is not CloudBox compatible"
    fi
    
    # Check if jq is available for JSON parsing
    if ! command -v jq >/dev/null 2>&1; then
        cip_warn "jq not found - installing for JSON validation"
        if command -v apt-get >/dev/null 2>&1; then
            sudo apt-get update && sudo apt-get install -y jq
        elif command -v yum >/dev/null 2>&1; then
            sudo yum install -y jq
        elif command -v apk >/dev/null 2>&1; then
            sudo apk add jq
        else
            cip_warn "Could not install jq - skipping JSON validation"
            return 0
        fi
    fi
    
    # Validate JSON syntax
    if ! jq empty cloudbox.json 2>/dev/null; then
        cip_fatal "cloudbox.json contains invalid JSON syntax"
    fi
    
    # Check required scripts
    local required_scripts=("install" "start" "stop" "status")
    for script in "${required_scripts[@]}"; do
        local script_path=$(jq -r ".scripts.$script // empty" cloudbox.json 2>/dev/null)
        if [[ -z "$script_path" ]]; then
            cip_fatal "Required script '$script' not defined in cloudbox.json"
        fi
        if [[ ! -f "$script_path" ]]; then
            cip_fatal "Required script file '$script_path' not found"
        fi
        if [[ ! -x "$script_path" ]]; then
            cip_warn "Script '$script_path' is not executable - fixing permissions"
            chmod +x "$script_path" 2>/dev/null || cip_fatal "Could not make script executable: $script_path"
        fi
        cip_debug "✅ Script '$script' validated: $script_path"
    done
    
    cip_success "CloudBox Install Protocol validation passed"
}

# Utility functions
cip_backup_file() {
    local file=$1
    local backup_suffix=${2:-".cloudbox.backup"}
    
    if [[ -f "$file" ]] && [[ ! -f "$file$backup_suffix" ]]; then
        cp "$file" "$file$backup_suffix"
        cip_debug "Created backup: $file$backup_suffix"
    fi
}

cip_restore_file() {
    local file=$1
    local backup_suffix=${2:-".cloudbox.backup"}
    
    if [[ -f "$file$backup_suffix" ]]; then
        cp "$file$backup_suffix" "$file"
        cip_success "Restored from backup: $file"
    else
        cip_warn "No backup found for: $file"
    fi
}

# Environment variable formatting for different frameworks
cip_write_env_file() {
    local env_file=${1:-".env.production"}
    local framework=${2:-"generic"}
    
    cip_step "Writing environment configuration to $env_file"
    
    # Backup existing file
    cip_backup_file "$env_file"
    
    # Write CloudBox environment
    cat > "$env_file" << EOF
# CloudBox Auto-Generated Configuration
# Generated: $(date)
# Framework: $framework

# Core configuration
NODE_ENV=production
PORT=${CLOUDBOX_WEB_PORT:-3000}

# CloudBox integration
API_URL=${CLOUDBOX_API_URL:-http://localhost:3001}
PROJECT_ID=${CLOUDBOX_PROJECT_ID:-1}
DEPLOYMENT_ID=${CLOUDBOX_DEPLOYMENT_ID:-1}

EOF

    # Framework-specific variables
    case "$framework" in
        "svelte"|"sveltekit")
            cat >> "$env_file" << EOF
# SvelteKit specific
ORIGIN=http://localhost:${CLOUDBOX_WEB_PORT:-3000}
VITE_API_URL=${CLOUDBOX_API_URL:-http://localhost:3001}
VITE_PROJECT_ID=${CLOUDBOX_PROJECT_ID:-1}

EOF
            ;;
        "react"|"nextjs")
            cat >> "$env_file" << EOF
# React/Next.js specific
REACT_APP_API_URL=${CLOUDBOX_API_URL:-http://localhost:3001}
NEXT_PUBLIC_API_URL=${CLOUDBOX_API_URL:-http://localhost:3001}

EOF
            ;;
        "vue"|"nuxt")
            cat >> "$env_file" << EOF
# Vue/Nuxt specific
VUE_APP_API_URL=${CLOUDBOX_API_URL:-http://localhost:3001}
NUXT_PUBLIC_API_URL=${CLOUDBOX_API_URL:-http://localhost:3001}

EOF
            ;;
    esac
    
    cip_success "Environment configuration written to $env_file"
}

# Initialize debug mode if requested
if [[ "${CLOUDBOX_DEBUG}" == "true" ]]; then
    cip_debug "Debug mode enabled"
    set -x  # Enable command tracing
fi

# Set up cleanup trap
trap cip_cleanup EXIT