# CloudBox Install Protocol (CIP) v1.0

**Gestandaardiseerde deployment interface voor CloudBox-compatible applicaties**

## 🎯 Protocol Overview

Het CloudBox Install Protocol (CIP) definieert een gestandaardiseerde manier waarop applicaties kunnen worden gedeployed en beheerd door CloudBox. Elke app die compatible wil zijn met CloudBox moet dit protocol implementeren.

## 📋 Required Files

Elke CloudBox-compatible app moet de volgende bestanden bevatten:

### 1. `cloudbox.json` - App Manifest
```json
{
  "name": "my-portfolio-app",
  "version": "1.0.0",
  "type": "frontend|backend|fullstack",
  "cloudbox_version": "1.0",
  "
  
  "description": "Photography portfolio app",
  "author": "Developer Name",
  "license": "MIT",
  
  "scripts": {
    "install": "./scripts/cloudbox-install.sh",
    "start": "./scripts/cloudbox-start.sh", 
    "stop": "./scripts/cloudbox-stop.sh",
    "status": "./scripts/cloudbox-status.sh",
    "health": "./scripts/cloudbox-health.sh"
  },
  
  "requirements": {
    "node": ">=18.0.0",
    "npm": ">=9.0.0",
    "docker": ">=20.0.0"
  },
  
  "ports": {
    "web": {
      "port": 3000,
      "required": true,
      "description": "Main web server port",
      "configurable": true
    }
  },
  
  "environment": {
    "NODE_ENV": {
      "default": "production",
      "required": true,
      "description": "Node.js environment"
    },
    "API_URL": {
      "default": "{{CLOUDBOX_API_URL}}",
      "required": true,
      "description": "CloudBox API endpoint"
    }
  },
  
  "docker": {
    "enabled": true,
    "compose_file": "docker-compose.cloudbox.yml",
    "monitoring": {
      "enabled": true,
      "health_endpoint": "/health",
      "metrics_endpoint": "/metrics"
    }
  },
  
  "cloudbox": {
    "data_api": true,
    "auth_api": true,
    "storage_api": true,
    "templates": ["photoportfolio"]
  }
}
```

### 2. `scripts/cloudbox-install.sh` - Installation Script
```bash
#!/bin/bash
set -e

source "$(dirname "$0")/cloudbox-common.sh"

cip_header "Installing $(cip_get_app_name)"

# CloudBox will provide these environment variables:
# - CLOUDBOX_API_URL: API endpoint for this project
# - CLOUDBOX_PROJECT_SLUG: Project identifier  
# - CLOUDBOX_DEPLOYMENT_PATH: Target deployment directory
# - CLOUDBOX_WEB_PORT: Assigned web server port
# - CLOUDBOX_ENVIRONMENT: production|staging|development

cip_step "Validating environment"
cip_require_env "CLOUDBOX_API_URL" "CLOUDBOX_WEB_PORT"

cip_step "Installing dependencies"
npm ci --production

cip_step "Building application"
npm run build

cip_step "Configuring CloudBox integration"
cat > .env.production << EOF
NODE_ENV=production
PORT=\${CLOUDBOX_WEB_PORT}
API_URL=\${CLOUDBOX_API_URL}
VITE_API_URL=\${CLOUDBOX_API_URL}
EOF

cip_step "Setting up Docker monitoring"
if [[ "\${CLOUDBOX_DOCKER_ENABLED}" == "true" ]]; then
    cip_docker_setup
fi

cip_success "Installation completed successfully"
cip_info "App will be available on port \${CLOUDBOX_WEB_PORT}"
```

### 3. `scripts/cloudbox-common.sh` - Common Functions
```bash
#!/bin/bash

# CloudBox Install Protocol - Common Functions v1.0

# Colors for output
readonly CIP_GREEN='\033[0;32m'
readonly CIP_BLUE='\033[0;34m'
readonly CIP_YELLOW='\033[1;33m'
readonly CIP_RED='\033[0;31m'
readonly CIP_NC='\033[0m' # No Color

# Logging functions
cip_header() {
    echo -e "\n${CIP_BLUE}================================================${CIP_NC}"
    echo -e "${CIP_BLUE}CloudBox Install Protocol v1.0${CIP_NC}"
    echo -e "${CIP_BLUE}$1${CIP_NC}"
    echo -e "${CIP_BLUE}================================================${CIP_NC}\n"
}

cip_step() {
    echo -e "${CIP_BLUE}[CIP]${CIP_NC} $1..."
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
    exit 1
}

# Environment validation
cip_require_env() {
    for var in "$@"; do
        if [[ -z "${!var}" ]]; then
            cip_fatal "Required environment variable $var is not set"
        fi
    done
}

# App manifest functions
cip_get_app_name() {
    if [[ -f "cloudbox.json" ]]; then
        jq -r '.name // "Unknown App"' cloudbox.json
    else
        echo "Unknown App"
    fi
}

cip_get_app_version() {
    if [[ -f "cloudbox.json" ]]; then
        jq -r '.version // "1.0.0"' cloudbox.json
    else
        echo "1.0.0"
    fi
}

# Port management
cip_check_port() {
    local port=$1
    if netstat -tuln 2>/dev/null | grep ":$port " >/dev/null; then
        return 1  # Port in use
    else
        return 0  # Port available
    fi
}

cip_wait_for_port() {
    local port=$1
    local timeout=${2:-30}
    local count=0
    
    cip_step "Waiting for service on port $port"
    while [[ $count -lt $timeout ]]; do
        if nc -z localhost $port 2>/dev/null; then
            cip_success "Service is responding on port $port"
            return 0
        fi
        sleep 1
        ((count++))
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
        cip_generate_docker_compose
    fi
    
    # Start monitoring containers
    docker-compose -f docker-compose.cloudbox.yml up -d
    
    cip_success "Docker monitoring enabled"
}

cip_generate_docker_compose() {
    cat > docker-compose.cloudbox.yml << 'EOF'
version: '3.8'

services:
  app:
    build: .
    ports:
      - "${CLOUDBOX_WEB_PORT}:${CLOUDBOX_WEB_PORT}"
    environment:
      - NODE_ENV=production
      - PORT=${CLOUDBOX_WEB_PORT}
      - API_URL=${CLOUDBOX_API_URL}
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:${CLOUDBOX_WEB_PORT}/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    labels:
      - "cloudbox.enable=true"
      - "cloudbox.project=${CLOUDBOX_PROJECT_SLUG}"
      - "cloudbox.type=frontend"

  monitoring:
    image: prom/node-exporter:latest
    ports:
      - "9100:9100"
    restart: unless-stopped
    labels:
      - "cloudbox.enable=true"
      - "cloudbox.service=monitoring"
EOF
}

# Health check functions
cip_health_check() {
    local endpoint=${1:-"/health"}
    local port=${2:-$CLOUDBOX_WEB_PORT}
    
    if curl -f -s "http://localhost:$port$endpoint" >/dev/null; then
        return 0
    else
        return 1
    fi
}

# Cleanup functions
cip_cleanup() {
    cip_step "Cleaning up temporary files"
    rm -rf /tmp/cloudbox-* 2>/dev/null || true
    cip_success "Cleanup completed"
}

# Install protocol validation
cip_validate_protocol() {
    if [[ ! -f "cloudbox.json" ]]; then
        cip_fatal "cloudbox.json not found - this app is not CloudBox compatible"
    fi
    
    local required_scripts=("install" "start" "stop" "status")
    for script in "${required_scripts[@]}"; do
        local script_path=$(jq -r ".scripts.$script // empty" cloudbox.json)
        if [[ -z "$script_path" ]]; then
            cip_fatal "Required script '$script' not defined in cloudbox.json"
        fi
        if [[ ! -f "$script_path" ]]; then
            cip_fatal "Required script file '$script_path' not found"
        fi
    done
    
    cip_success "CloudBox Install Protocol validation passed"
}
```

## 🚀 Implementation Benefits

### **For App Developers:**
- ✅ **Standardized Interface**: Clear contract voor CloudBox compatibility
- ✅ **Environment Injection**: CloudBox provides alle nodige variabelen  
- ✅ **Port Management**: Automatische port assignment en conflict resolution
- ✅ **Docker Integration**: Built-in monitoring en health checks
- ✅ **Interactive Support**: Scripts kunnen prompts gebruiken

### **For CloudBox:**
- ✅ **Clean Architecture**: Geen .env hacking meer
- ✅ **Remote Execution**: SSH-based interactive deployment
- ✅ **Monitoring**: Docker-based app monitoring
- ✅ **Standardization**: Consistente deployment experience

## 📖 SDK Documentation

Dit protocol wordt onderdeel van de CloudBox SDK met:
- ✅ **Template Generator**: `cloudbox create-app --template portfolio`
- ✅ **Validation Tool**: `cloudbox validate`  
- ✅ **Testing Suite**: `cloudbox test-deployment`
- ✅ **Migration Guide**: Voor bestaande apps

Zal ik nu de remote terminal execution implementeren? 🚀