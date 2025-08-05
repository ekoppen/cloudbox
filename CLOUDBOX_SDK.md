# CloudBox SDK v1.0 🚀

**Complete Developer Guide voor CloudBox-Compatible Applicaties**

## 📖 Inhoudsopgave

1. [Quick Start](#-quick-start)
2. [CloudBox Install Protocol (CIP)](#-cloudbox-install-protocol-cip)
3. [App Manifest (cloudbox.json)](#-app-manifest-cloudboxjson)
4. [Install Scripts](#-install-scripts)
5. [Environment Variables](#-environment-variables)
6. [Port Management](#-port-management)
7. [Docker Integration](#-docker-integration)
8. [Templates & Examples](#-templates--examples)
9. [Testing & Validation](#-testing--validation)
10. [Deployment Guide](#-deployment-guide)

---

## 🚀 Quick Start

### Stap 1: Voeg CloudBox compatibiliteit toe aan je bestaande app

```bash
# Installeer CloudBox CLI (toekomstig)
npm install -g cloudbox-cli

# Initialiseer CloudBox in je project
cloudbox init --template portfolio

# Of handmatig:
# 1. Creëer cloudbox.json
# 2. Voeg scripts toe aan scripts/ directory
# 3. Test lokaal met cloudbox validate
```

### Stap 2: Basisbestanden

Elke CloudBox-compatible app heeft deze bestanden nodig:

```
my-app/
├── cloudbox.json                    # App manifest
├── scripts/
│   ├── cloudbox-common.sh          # Herbruikbare functies
│   ├── cloudbox-install.sh         # Installatie script
│   ├── cloudbox-start.sh           # Start script
│   ├── cloudbox-stop.sh            # Stop script
│   ├── cloudbox-status.sh          # Status check
│   └── cloudbox-health.sh          # Health check
├── docker-compose.cloudbox.yml     # Docker monitoring
├── package.json                    # App dependencies
└── [je app bestanden]
```

---

## 📋 CloudBox Install Protocol (CIP)

### Protocol Overzicht

Het CloudBox Install Protocol (CIP) is een gestandaardiseerde interface die zorgt voor consistente deployment van applicaties via CloudBox. Elke app die via CloudBox gedeployed wil worden moet dit protocol implementeren.

### Core Principes

1. **Standardization**: Elke app gebruikt dezelfde interface voor deployment
2. **Environment Injection**: CloudBox injecteert automatisch alle benodigde variabelen
3. **Remote Execution**: Scripts worden uitgevoerd op de doelserver via SSH
4. **Interactive Support**: Scripts kunnen interactieve prompts tonen
5. **Docker Integration**: Optionele Docker-based monitoring en health checks

---

## 📄 App Manifest (cloudbox.json)

### Complete Template

```json
{
  "name": "my-portfolio-app",
  "version": "1.0.0",
  "type": "frontend|backend|fullstack",
  "cloudbox_version": "1.0",
  
  "description": "Photography portfolio app",
  "author": "Developer Name",
  "license": "MIT",
  "repository": "https://github.com/user/repo",
  
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
    "docker": ">=20.0.0",
    "git": ">=2.0.0"
  },
  
  "ports": {
    "web": {
      "port": 3000,
      "required": true,
      "description": "Main web server port",
      "configurable": true,
      "variable": "WEB_PORT"
    },
    "api": {
      "port": 4000,
      "required": false,
      "description": "API server port",
      "configurable": true,
      "variable": "API_PORT"
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
    },
    "DATABASE_URL": {
      "default": "{{CLOUDBOX_DATABASE_URL}}",
      "required": false,
      "description": "Database connection string"
    }
  },
  
  "docker": {
    "enabled": true,
    "compose_file": "docker-compose.cloudbox.yml",
    "monitoring": {
      "enabled": true,
      "health_endpoint": "/health",
      "metrics_endpoint": "/metrics",
      "logs": true
    }
  },
  
  "cloudbox": {
    "data_api": true,
    "auth_api": true,
    "storage_api": true,
    "templates": ["photoportfolio", "blog", "ecommerce"],
    "categories": ["portfolio", "photography"]
  },
  
  "deployment": {
    "build_command": "npm run build",
    "start_command": "npm start",
    "install_dependencies": "npm ci --production",
    "pre_deploy": ["npm run test"],
    "post_deploy": ["npm run migrate"]
  }
}
```

### Veld Definities

| Veld | Type | Verplicht | Beschrijving |
|------|------|----------|--------------|
| `name` | string | ✅ | App naam (lowercase, kebab-case) |
| `version` | string | ✅ | Semantic versioning (1.0.0) |
| `type` | enum | ✅ | frontend, backend, fullstack |
| `cloudbox_version` | string | ✅ | Ondersteunde CloudBox versie |
| `scripts` | object | ✅ | Pad naar CIP scripts |
| `ports` | object | ❌ | Port requirements en configuratie |
| `environment` | object | ❌ | Environment variabelen |
| `docker` | object | ❌ | Docker configuratie |
| `cloudbox` | object | ❌ | CloudBox specifieke instellingen |

---

## 🔧 Install Scripts

### Script Structuur

Alle CIP scripts volgen dezelfde structuur:

```bash
#!/bin/bash
set -e

# Laad gemeenschappelijke functies
source "$(dirname "$0")/cloudbox-common.sh"

# Script header
cip_header "Installing $(cip_get_app_name)"

# Validatie
cip_require_env "CLOUDBOX_API_URL" "CLOUDBOX_WEB_PORT"

# Implementatie
cip_step "Installing dependencies"
npm ci --production

cip_step "Building application"
npm run build

# Configuratie
cip_step "Configuring CloudBox integration"
cat > .env.production << EOF
NODE_ENV=production
PORT=${CLOUDBOX_WEB_PORT}
API_URL=${CLOUDBOX_API_URL}
EOF

# Success
cip_success "Installation completed successfully"
```

### 1. cloudbox-install.sh

```bash
#!/bin/bash
set -e

source "$(dirname "$0")/cloudbox-common.sh"

cip_header "Installing $(cip_get_app_name) v$(cip_get_app_version)"

# Environment validation
cip_step "Validating CloudBox environment"
cip_require_env "CLOUDBOX_API_URL" "CLOUDBOX_WEB_PORT" "CLOUDBOX_PROJECT_ID"

# System requirements check
cip_step "Checking system requirements"
cip_check_command "node" "Node.js is required"
cip_check_command "npm" "NPM is required"

# Install dependencies
cip_step "Installing application dependencies"
if [[ -f "package-lock.json" ]]; then
    npm ci --production
else
    npm install --production
fi

# Build application
cip_step "Building application for production"
if npm run build; then
    cip_success "Application built successfully"
else
    cip_fatal "Build failed - check your build configuration"
fi

# Environment configuration
cip_step "Configuring production environment"
cat > .env.production << EOF
# CloudBox Auto-Generated Configuration
NODE_ENV=production
PORT=${CLOUDBOX_WEB_PORT}
API_URL=${CLOUDBOX_API_URL}
PROJECT_ID=${CLOUDBOX_PROJECT_ID}

# App-specific configuration
VITE_API_URL=${CLOUDBOX_API_URL}
VITE_PROJECT_ID=${CLOUDBOX_PROJECT_ID}
EOF

# Database setup (if needed)
if [[ "${CLOUDBOX_DATABASE_URL}" != "" ]]; then
    cip_step "Configuring database connection"
    echo "DATABASE_URL=${CLOUDBOX_DATABASE_URL}" >> .env.production
    
    if [[ -f "package.json" ]] && grep -q "migrate" package.json; then
        cip_step "Running database migrations"
        npm run migrate || cip_warn "Database migration failed - manual intervention may be required"
    fi
fi

# Docker setup
if [[ "${CLOUDBOX_DOCKER_ENABLED}" == "true" ]]; then
    cip_step "Setting up Docker monitoring"
    cip_docker_setup
fi

# File permissions
cip_step "Setting file permissions"
chmod +x scripts/cloudbox-*.sh 2>/dev/null || true

cip_success "✅ Installation completed successfully!"
cip_info "🌐 App will be available on port ${CLOUDBOX_WEB_PORT}"
cip_info "📡 API endpoint: ${CLOUDBOX_API_URL}"
```

### 2. cloudbox-start.sh

```bash
#!/bin/bash
set -e

source "$(dirname "$0")/cloudbox-common.sh"

cip_header "Starting $(cip_get_app_name)"

# Load environment
if [[ -f ".env.production" ]]; then
    source .env.production
    cip_success "Production environment loaded"
else
    cip_warn "No .env.production found - using defaults"
fi

# Pre-start checks
cip_step "Performing pre-start health checks"
if [[ ! -d "node_modules" ]]; then
    cip_fatal "Dependencies not installed - run install script first"
fi

if [[ ! -d "build" ]] && [[ ! -d "dist" ]] && [[ ! -f "index.js" ]]; then
    cip_fatal "Application not built - run install script first"
fi

# Stop existing process
cip_step "Stopping existing processes"
if pgrep -f "node.*$(basename $PWD)" > /dev/null; then
    pkill -f "node.*$(basename $PWD)" || true
    sleep 2
fi

# Port check
cip_step "Checking port availability"
if ! cip_check_port ${PORT:-3000}; then
    cip_fatal "Port ${PORT:-3000} is already in use"
fi

# Start application
cip_step "Starting application server"
if [[ -f "package.json" ]]; then
    # Use npm start if available
    if jq -e '.scripts.start' package.json > /dev/null; then
        npm start &
        APP_PID=$!
        echo $APP_PID > .cloudbox.pid
    else
        cip_fatal "No start script found in package.json"
    fi
else
    cip_fatal "No package.json found"
fi

# Verify startup
cip_step "Verifying application startup"
if cip_wait_for_port ${PORT:-3000} 30; then
    cip_success "✅ Application started successfully on port ${PORT:-3000}"
    cip_info "🔗 Access your app at: http://localhost:${PORT:-3000}"
else
    cip_fatal "Application failed to start within 30 seconds"
fi

# Docker monitoring start
if [[ "${CLOUDBOX_DOCKER_ENABLED}" == "true" ]] && [[ -f "docker-compose.cloudbox.yml" ]]; then
    cip_step "Starting Docker monitoring"
    docker-compose -f docker-compose.cloudbox.yml up -d monitoring
fi
```

### 3. cloudbox-stop.sh

```bash
#!/bin/bash
set -e

source "$(dirname "$0")/cloudbox-common.sh"

cip_header "Stopping $(cip_get_app_name)"

# Stop main application
cip_step "Stopping application server"
if [[ -f ".cloudbox.pid" ]]; then
    PID=$(cat .cloudbox.pid)
    if kill -0 $PID 2>/dev/null; then
        kill $PID
        sleep 3
        if kill -0 $PID 2>/dev/null; then
            cip_warn "Process still running, force killing..."
            kill -9 $PID
        fi
        rm -f .cloudbox.pid
        cip_success "Application stopped"
    else
        cip_info "Application was not running"
        rm -f .cloudbox.pid
    fi
else
    # Fallback: kill by process name
    if pgrep -f "node.*$(basename $PWD)" > /dev/null; then
        pkill -f "node.*$(basename $PWD)"
        cip_success "Application processes stopped"
    else
        cip_info "No running processes found"
    fi
fi

# Stop Docker monitoring
if [[ "${CLOUDBOX_DOCKER_ENABLED}" == "true" ]] && [[ -f "docker-compose.cloudbox.yml" ]]; then
    cip_step "Stopping Docker monitoring"
    docker-compose -f docker-compose.cloudbox.yml down
fi

cip_success "✅ Application stopped successfully"
```

### 4. cloudbox-status.sh

```bash
#!/bin/bash
set -e

source "$(dirname "$0")/cloudbox-common.sh"

cip_header "Status Check for $(cip_get_app_name)"

# Check if process is running
cip_step "Checking application process"
if [[ -f ".cloudbox.pid" ]]; then
    PID=$(cat .cloudbox.pid)
    if kill -0 $PID 2>/dev/null; then
        cip_success "✅ Application is running (PID: $PID)"
        STATUS="running"
    else
        cip_error "❌ Application is not running (stale PID file)"
        rm -f .cloudbox.pid
        STATUS="stopped"
    fi
else
    if pgrep -f "node.*$(basename $PWD)" > /dev/null; then
        cip_warn "⚠️  Application is running but no PID file found"
        STATUS="running"
    else
        cip_info "ℹ️  Application is not running"
        STATUS="stopped"
    fi
fi

# Port check
if [[ -f ".env.production" ]]; then
    source .env.production
    PORT=${PORT:-3000}
    
    cip_step "Checking port ${PORT}"
    if nc -z localhost $PORT 2>/dev/null; then
        cip_success "✅ Port ${PORT} is responding"
        PORT_STATUS="open"
    else
        cip_error "❌ Port ${PORT} is not responding"
        PORT_STATUS="closed"
    fi
else
    cip_warn "No .env.production found - cannot check port"
    PORT_STATUS="unknown"
fi

# Health check
cip_step "Performing health check"
if [[ -f "scripts/cloudbox-health.sh" ]]; then
    if ./scripts/cloudbox-health.sh; then
        HEALTH_STATUS="healthy"
    else
        HEALTH_STATUS="unhealthy"
    fi
else
    cip_info "No health check script available"
    HEALTH_STATUS="unknown"
fi

# Docker status
if [[ "${CLOUDBOX_DOCKER_ENABLED}" == "true" ]] && [[ -f "docker-compose.cloudbox.yml" ]]; then
    cip_step "Checking Docker monitoring"
    if docker-compose -f docker-compose.cloudbox.yml ps -q monitoring | grep -q .; then
        cip_success "✅ Docker monitoring is running"
        DOCKER_STATUS="running"
    else
        cip_warn "⚠️  Docker monitoring is not running"
        DOCKER_STATUS="stopped"
    fi
else
    DOCKER_STATUS="disabled"
fi

# Summary
echo
cip_info "📊 Status Summary:"
echo "   Application: $STATUS"
echo "   Port:        $PORT_STATUS"
echo "   Health:      $HEALTH_STATUS"
echo "   Docker:      $DOCKER_STATUS"

# Exit with appropriate code
if [[ "$STATUS" == "running" ]] && [[ "$PORT_STATUS" == "open" ]] && [[ "$HEALTH_STATUS" != "unhealthy" ]]; then
    exit 0
else
    exit 1
fi
```

### 5. cloudbox-health.sh

```bash
#!/bin/bash
set -e

source "$(dirname "$0")/cloudbox-common.sh"

cip_header "Health Check for $(cip_get_app_name)"

# Load environment
if [[ -f ".env.production" ]]; then
    source .env.production
    PORT=${PORT:-3000}
else
    PORT=3000
fi

# HTTP health check
cip_step "Checking HTTP health endpoint"
if cip_health_check "/health" $PORT; then
    cip_success "✅ Health endpoint responding"
    ENDPOINT_STATUS="healthy"
else
    # Fallback: check root endpoint
    if cip_health_check "/" $PORT; then
        cip_success "✅ Root endpoint responding"
        ENDPOINT_STATUS="healthy"
    else
        cip_error "❌ No endpoints responding"
        ENDPOINT_STATUS="unhealthy"
    fi
fi

# Resource checks
cip_step "Checking system resources"

# Memory usage
MEMORY_PERCENT=$(free | awk 'FNR==2{printf "%.0f", $3/$2*100}')
if [[ $MEMORY_PERCENT -lt 90 ]]; then
    cip_success "✅ Memory usage: ${MEMORY_PERCENT}%"
    MEMORY_STATUS="good"
else
    cip_warn "⚠️  High memory usage: ${MEMORY_PERCENT}%"
    MEMORY_STATUS="warning"
fi

# Disk usage
DISK_PERCENT=$(df . | awk 'FNR==2{print $5}' | sed 's/%//')
if [[ $DISK_PERCENT -lt 90 ]]; then
    cip_success "✅ Disk usage: ${DISK_PERCENT}%"
    DISK_STATUS="good"
else
    cip_warn "⚠️  High disk usage: ${DISK_PERCENT}%"
    DISK_STATUS="warning"
fi

# Application-specific health checks
if [[ -f "package.json" ]] && grep -q "healthcheck" package.json; then
    cip_step "Running application health checks"
    if npm run healthcheck; then
        cip_success "✅ Application health checks passed"
        APP_HEALTH="healthy"
    else
        cip_error "❌ Application health checks failed"
        APP_HEALTH="unhealthy"
    fi
else
    APP_HEALTH="unknown"
fi

# CloudBox API connectivity
if [[ -n "${CLOUDBOX_API_URL}" ]]; then
    cip_step "Testing CloudBox API connectivity"
    if curl -f -s "${CLOUDBOX_API_URL}/health" > /dev/null 2>&1; then
        cip_success "✅ CloudBox API is reachable"
        API_STATUS="connected"
    else
        cip_warn "⚠️  CloudBox API is not reachable"
        API_STATUS="disconnected"
    fi
else
    API_STATUS="not_configured"
fi

# Summary
echo
cip_info "🏥 Health Summary:"
echo "   HTTP:        $ENDPOINT_STATUS"
echo "   Memory:      $MEMORY_STATUS (${MEMORY_PERCENT}%)"
echo "   Disk:        $DISK_STATUS (${DISK_PERCENT}%)"
echo "   App Health:  $APP_HEALTH"
echo "   CloudBox:    $API_STATUS"

# Exit with appropriate code
if [[ "$ENDPOINT_STATUS" == "healthy" ]] && [[ "$APP_HEALTH" != "unhealthy" ]]; then
    exit 0
else
    exit 1
fi
```

---

## 🌐 Environment Variables

### CloudBox Provided Variables

CloudBox injecteert automatisch deze environment variabelen in je scripts:

| Variable | Description | Example |
|----------|-------------|---------|
| `CLOUDBOX_API_URL` | CloudBox API endpoint voor dit project | `https://cloudbox.domain/api/projects/123` |
| `CLOUDBOX_PROJECT_ID` | Unieke project identifier | `123` |
| `CLOUDBOX_PROJECT_SLUG` | Project slug/naam | `my-portfolio` |
| `CLOUDBOX_DEPLOYMENT_ID` | Deployment ID | `456` |
| `CLOUDBOX_DEPLOYMENT_PATH` | Target deployment directory | `/home/user/deploys/my-app` |
| `CLOUDBOX_WEB_PORT` | Toegewezen web server port | `3000` |
| `CLOUDBOX_ENVIRONMENT` | Environment type | `production` |
| `CLOUDBOX_VERSION` | CloudBox versie | `1.0` |
| `CLOUDBOX_DOCKER_ENABLED` | Docker monitoring enabled | `true` |

### Port Variables

Voor elke gedefinieerde port in `cloudbox.json`:

```bash
CLOUDBOX_WEB_PORT=3000      # Main web port
CLOUDBOX_API_PORT=4000      # API server port  
CLOUDBOX_DB_PORT=5432       # Database port
```

### Application Environment

Je app kan deze variabelen gebruiken in `.env.production`:

```bash
# Base configuration
NODE_ENV=production
PORT=${CLOUDBOX_WEB_PORT}
API_URL=${CLOUDBOX_API_URL}

# CloudBox integration
VITE_API_URL=${CLOUDBOX_API_URL}
REACT_APP_API_URL=${CLOUDBOX_API_URL}
NEXT_PUBLIC_API_URL=${CLOUDBOX_API_URL}

# Database (indien beschikbaar)
DATABASE_URL=${CLOUDBOX_DATABASE_URL}

# Custom app configuration
PHOTO_UPLOAD_PATH=${CLOUDBOX_DEPLOYMENT_PATH}/uploads
CACHE_DIR=${CLOUDBOX_DEPLOYMENT_PATH}/cache
```

---

## 🔌 Port Management

### Port Configuration

DefiniÃªer ports in `cloudbox.json`:

```json
{
  "ports": {
    "web": {
      "port": 3000,
      "required": true,
      "description": "Main web server",
      "configurable": true,
      "variable": "WEB_PORT"
    },
    "api": {
      "port": 4000,
      "required": false,
      "description": "API server",
      "configurable": true,
      "variable": "API_PORT"
    },
    "websocket": {
      "port": 8080,
      "required": false,
      "description": "WebSocket server",
      "configurable": true,
      "variable": "WS_PORT"
    }
  }
}
```

### Port Availability Checking

CloudBox checkt automatisch port beschikbaarheid:

```bash
# In je scripts kun je gebruiken:
if ! cip_check_port $PORT; then
    cip_fatal "Port $PORT is already in use"
fi

# Wacht tot port beschikbaar is:
cip_wait_for_port $PORT 30  # 30 seconden timeout
```

---

## 🐳 Docker Integration

### Docker Compose voor Monitoring

Creëer `docker-compose.cloudbox.yml`:

```yaml
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
      start_period: 60s
    labels:
      - "cloudbox.enable=true"
      - "cloudbox.project=${CLOUDBOX_PROJECT_SLUG}"
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

  logs:
    image: grafana/loki:latest
    ports:
      - "3100:3100"
    restart: unless-stopped
    volumes:
      - ./loki-config.yml:/etc/loki/local-config.yaml
    labels:
      - "cloudbox.enable=true"
      - "cloudbox.service=logs"
```

### Dockerfile Template

```dockerfile
# Multi-stage build voor efficiency
FROM node:18-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production && npm cache clean --force

COPY . .
RUN npm run build

# Production stage
FROM node:18-alpine AS production

# Install dependencies voor health checks
RUN apk add --no-cache curl

WORKDIR /app

# Copy built application
COPY --from=builder /app/build ./build
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package*.json ./
COPY --from=builder /app/scripts ./scripts

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD curl -f http://localhost:${PORT:-3000}/health || exit 1

# Non-root user
RUN addgroup -g 1001 -S nodejs
RUN adduser -S nodejs -u 1001
USER nodejs

EXPOSE ${PORT:-3000}

CMD ["npm", "start"]
```

---

## 📚 Templates & Examples

### 1. SvelteKit Portfolio Template

```json
{
  "name": "sveltekit-portfolio",
  "version": "1.0.0",
  "type": "frontend",
  "cloudbox_version": "1.0",
  
  "scripts": {
    "install": "./scripts/cloudbox-install.sh",
    "start": "./scripts/cloudbox-start.sh",
    "stop": "./scripts/cloudbox-stop.sh",
    "status": "./scripts/cloudbox-status.sh",
    "health": "./scripts/cloudbox-health.sh"
  },
  
  "ports": {
    "web": {
      "port": 3000,
      "required": true,
      "configurable": true,
      "variable": "PORT"
    }
  },
  
  "environment": {
    "NODE_ENV": {
      "default": "production",
      "required": true
    },
    "ORIGIN": {
      "default": "http://localhost:3000",
      "required": true
    }
  },
  
  "deployment": {
    "build_command": "npm run build",
    "start_command": "node build",
    "install_dependencies": "npm ci --production"
  }
}
```

### 2. Express.js API Template

```json
{
  "name": "express-api",
  "version": "1.0.0", 
  "type": "backend",
  "cloudbox_version": "1.0",
  
  "ports": {
    "api": {
      "port": 4000,
      "required": true,
      "configurable": true,
      "variable": "PORT"
    }
  },
  
  "environment": {
    "NODE_ENV": {
      "default": "production",
      "required": true
    },
    "DATABASE_URL": {
      "default": "{{CLOUDBOX_DATABASE_URL}}",
      "required": true
    },
    "JWT_SECRET": {
      "required": true,
      "description": "JWT signing secret"
    }
  }
}
```

### 3. Full-Stack Next.js Template

```json
{
  "name": "nextjs-fullstack",
  "version": "1.0.0",
  "type": "fullstack", 
  "cloudbox_version": "1.0",
  
  "ports": {
    "web": {
      "port": 3000,
      "required": true,
      "configurable": true
    }
  },
  
  "environment": {
    "NEXTAUTH_URL": {
      "default": "http://localhost:3000",
      "required": true
    },
    "NEXTAUTH_SECRET": {
      "required": true
    }
  },
  
  "deployment": {
    "build_command": "npm run build",
    "start_command": "npm start"
  }
}
```

---

## ✅ Testing & Validation

### Lokale Validatie

```bash
# Valideer cloudbox.json syntax
jq empty cloudbox.json

# Test alle scripts
chmod +x scripts/cloudbox-*.sh

# Test installatie
CLOUDBOX_API_URL="http://localhost:3001" \
CLOUDBOX_WEB_PORT="3000" \
CLOUDBOX_PROJECT_ID="1" \
./scripts/cloudbox-install.sh

# Test start
./scripts/cloudbox-start.sh

# Test status 
./scripts/cloudbox-status.sh

# Test health
./scripts/cloudbox-health.sh

# Test stop
./scripts/cloudbox-stop.sh
```

### CloudBox CLI Validation (toekomstig)

```bash
# Installeer CloudBox CLI
npm install -g cloudbox-cli

# Valideer project
cloudbox validate

# Test deployment lokaal
cloudbox test-deploy --local

# Preview deployment
cloudbox deploy --dry-run
```

### CI/CD Integration

```yaml
# .github/workflows/cloudbox-validate.yml
name: CloudBox Validation

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  validate:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Setup Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
        
    - name: Install CloudBox CLI
      run: npm install -g cloudbox-cli
      
    - name: Validate CloudBox Manifest
      run: cloudbox validate
      
    - name: Test CIP Scripts
      run: cloudbox test-scripts
```

---

## 🚀 Deployment Guide

### Deployment via CloudBox UI

1. **Repository Setup**
   - Voeg `cloudbox.json` toe aan je repository
   - Implementeer alle vereiste CIP scripts
   - Test lokaal met `cloudbox validate`

2. **CloudBox Configuration**
   - Voeg repository toe in CloudBox
   - Configureer deployment instellingen
   - Kies target server en poorten

3. **Deploy**
   - Klik "Deploy" in CloudBox UI
   - Monitor real-time logs 
   - Verify deployment success

### Manual Deployment

```bash
# Clone repository op doelserver
git clone https://github.com/user/repo.git
cd repo

# Set CloudBox environment variables
export CLOUDBOX_API_URL="https://cloudbox.domain/api/projects/123"
export CLOUDBOX_WEB_PORT="3000"
export CLOUDBOX_PROJECT_ID="123"
export CLOUDBOX_DEPLOYMENT_PATH="/home/user/deploys/my-app"

# Run CIP installation
./scripts/cloudbox-install.sh

# Start application  
./scripts/cloudbox-start.sh

# Verify deployment
./scripts/cloudbox-status.sh
./scripts/cloudbox-health.sh
```

### Rollback Strategy

```bash
# Stop current deployment
./scripts/cloudbox-stop.sh

# Backup current deployment
cp -r . ../backup-$(date +%Y%m%d_%H%M%S)

# Restore previous version
git checkout [previous-commit]
./scripts/cloudbox-install.sh
./scripts/cloudbox-start.sh

# Verify rollback
./scripts/cloudbox-health.sh
```

---

## 🔧 Advanced Configuration

### Custom Environment Variables

```json
{
  "environment": {
    "CUSTOM_CONFIG": {
      "default": "default-value",
      "required": false,
      "description": "Custom configuration",
      "validation": "^[a-zA-Z0-9_-]+$"
    }
  }
}
```

### Multi-Environment Support

```json
{
  "environments": {
    "development": {
      "ports": { "web": { "port": 3000 } },
      "environment": { "DEBUG": { "default": "true" } }
    },
    "staging": {
      "ports": { "web": { "port": 3001 } },
      "environment": { "DEBUG": { "default": "false" } }
    },
    "production": {
      "ports": { "web": { "port": 80 } },
      "environment": { "DEBUG": { "default": "false" } }
    }
  }
}
```

### Load Balancing Support

```json
{
  "scaling": {
    "min_instances": 1,
    "max_instances": 5,
    "cpu_threshold": 70,
    "memory_threshold": 80
  }
}
```

---

## 📖 Best Practices

### 1. Script Security

```bash
# Always validate input
cip_require_env "REQUIRED_VAR"

# Use proper escaping
SAFE_VALUE=$(printf '%q' "$USER_INPUT")

# Check file permissions
if [[ ! -r "$CONFIG_FILE" ]]; then
    cip_fatal "Cannot read config file: $CONFIG_FILE"
fi
```

### 2. Error Handling

```bash
# Use proper error codes
set -e  # Exit on error

# Provide meaningful error messages
cip_fatal "Database connection failed. Check DATABASE_URL configuration."

# Clean up on failure
trap cip_cleanup EXIT
```

### 3. Logging

```bash
# Use structured logging
cip_step "Processing configuration"
cip_info "Using port: $PORT"
cip_warn "Feature X is deprecated"
cip_error "Operation failed"
cip_success "Configuration complete"
```

### 4. Performance

```bash
# Check resource usage
if [[ $(free -m | awk 'NR==2{printf "%.0f", $3*100/$2}') -gt 90 ]]; then
    cip_warn "High memory usage detected"
fi

# Optimize for production
if [[ "$NODE_ENV" == "production" ]]; then
    npm ci --production --silent
fi
```

---

## 🆘 Troubleshooting

### Common Issues

1. **Permission Denied**
   ```bash
   chmod +x scripts/cloudbox-*.sh
   ```

2. **Port Already in Use**
   ```bash
   # Check what's using the port
   netstat -tlnp | grep :3000
   # Kill process if safe
   fuser -k 3000/tcp
   ```

3. **Environment Variables Missing**
   ```bash
   # Check if CloudBox provided all required vars
   env | grep CLOUDBOX_
   ```

4. **Build Failures**
   ```bash
   # Clear cache and rebuild
   rm -rf node_modules package-lock.json
   npm install
   npm run build
   ```

### Debug Mode

```bash
# Enable debug logging
export CLOUDBOX_DEBUG=true
./scripts/cloudbox-install.sh

# Or add to script:
if [[ "${CLOUDBOX_DEBUG}" == "true" ]]; then
    set -x  # Enable command tracing
fi
```

---

## 📞 Support & Resources

- 📚 **Documentation**: [CloudBox Docs](https://docs.cloudbox.domain)
- 🐛 **Issues**: [GitHub Issues](https://github.com/cloudbox/cloudbox/issues)
- 💬 **Community**: [Discord Server](https://discord.gg/cloudbox)
- 📧 **Email**: support@cloudbox.domain

---

**CloudBox SDK v1.0** - Gebouwd voor developers, door developers 🚀