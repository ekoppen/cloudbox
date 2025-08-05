#!/bin/bash
set -e

# Load CloudBox Install Protocol common functions
source "$(dirname "$0")/cloudbox-common.sh"

# Display installation header
cip_header "Installing $(cip_get_app_name) v$(cip_get_app_version)"
cip_info "App Type: $(cip_get_app_type)"
cip_info "Installation Path: $(pwd)"

# Step 1: Validate CloudBox Install Protocol compliance
cip_validate_protocol

# Step 2: Environment validation
cip_step "Validating CloudBox environment"
cip_require_env "CLOUDBOX_API_URL" "CLOUDBOX_WEB_PORT" "CLOUDBOX_PROJECT_ID"

cip_info "CloudBox Configuration:"
cip_info "  API URL: ${CLOUDBOX_API_URL}"
cip_info "  Web Port: ${CLOUDBOX_WEB_PORT}"
cip_info "  Project ID: ${CLOUDBOX_PROJECT_ID}"
cip_info "  Environment: ${CLOUDBOX_ENVIRONMENT:-production}"

# Step 3: System requirements check
cip_step "Checking system requirements"
cip_check_command "node" "Node.js is required for SvelteKit applications"
cip_check_command "npm" "NPM is required for package management"

# Verify Node.js version
NODE_VERSION=$(node --version | sed 's/v//')
REQUIRED_NODE="18.0.0"
if [[ "$(printf '%s\n' "$REQUIRED_NODE" "$NODE_VERSION" | sort -V | head -n1)" != "$REQUIRED_NODE" ]]; then
    cip_fatal "Node.js version $NODE_VERSION is too old. Required: $REQUIRED_NODE or higher"
fi
cip_success "Node.js version $NODE_VERSION meets requirements"

# Step 4: Resource checks
cip_step "Checking system resources"
cip_check_resources 85 90  # Warn at 85% memory, 90% disk

# Step 5: Clean previous installation
if [[ -d "node_modules" ]]; then
    cip_step "Cleaning previous installation"
    rm -rf node_modules package-lock.json
    cip_success "Previous installation cleaned"
fi

# Step 6: Install dependencies
cip_step "Installing application dependencies"
if [[ -f "package-lock.json" ]]; then
    cip_info "Using package-lock.json for deterministic install"
    npm ci --production
else
    cip_info "No package-lock.json found, using package.json"
    npm install --production
fi

# Verify critical dependencies
if [[ ! -d "node_modules/@sveltejs" ]]; then
    cip_fatal "SvelteKit framework not found in dependencies"
fi
cip_success "Dependencies installed successfully"

# Step 7: Build application
cip_step "Building SvelteKit application for production"

# Check if build script exists
if ! npm run build --if-present; then
    cip_fatal "Build failed - check your SvelteKit configuration and dependencies"
fi

# Verify build output
if [[ ! -d "build" ]] && [[ ! -f "build/index.js" ]]; then
    cip_fatal "Build output not found - SvelteKit build may have failed"
fi
cip_success "Application built successfully"

# Step 8: Environment configuration
cip_step "Configuring production environment"

# Create SvelteKit-specific environment file
cip_write_env_file ".env.production" "sveltekit"

# Add SvelteKit adapter-specific configuration
cat >> .env.production << EOF
# SvelteKit Node adapter configuration
HOST=0.0.0.0
PORT=${CLOUDBOX_WEB_PORT}
ORIGIN=http://localhost:${CLOUDBOX_WEB_PORT}
BODY_SIZE_LIMIT=10485760

# CloudBox portfolio integration
CLOUDBOX_STORAGE_API=${CLOUDBOX_API_URL}/storage
CLOUDBOX_AUTH_API=${CLOUDBOX_API_URL}/auth
CLOUDBOX_DATA_API=${CLOUDBOX_API_URL}/data

EOF

cip_success "SvelteKit environment configured"

# Step 9: Portfolio-specific setup
cip_step "Configuring portfolio features"

# Create upload directories
mkdir -p uploads/photos uploads/thumbnails uploads/temp
chmod 755 uploads/photos uploads/thumbnails uploads/temp
cip_success "Upload directories created"

# Create CloudBox API integration file
cat > src/lib/cloudbox.js << 'EOF'
// CloudBox API Integration
const API_URL = process.env.API_URL || 'http://localhost:3001';
const PROJECT_ID = process.env.PROJECT_ID || '1';

export class CloudBoxAPI {
    constructor() {
        this.baseURL = API_URL;
        this.projectId = PROJECT_ID;
    }

    async uploadPhoto(file, metadata = {}) {
        const formData = new FormData();
        formData.append('photo', file);
        formData.append('metadata', JSON.stringify(metadata));
        formData.append('project_id', this.projectId);

        const response = await fetch(`${this.baseURL}/storage/upload`, {
            method: 'POST',
            body: formData
        });

        if (!response.ok) {
            throw new Error(`Upload failed: ${response.statusText}`);
        }

        return response.json();
    }

    async getPhotos(page = 1, limit = 20) {
        const response = await fetch(
            `${this.baseURL}/data/photos?project_id=${this.projectId}&page=${page}&limit=${limit}`
        );

        if (!response.ok) {
            throw new Error(`Failed to fetch photos: ${response.statusText}`);
        }

        return response.json();
    }

    async deletePhoto(photoId) {
        const response = await fetch(`${this.baseURL}/data/photos/${photoId}`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ project_id: this.projectId })
        });

        if (!response.ok) {
            throw new Error(`Delete failed: ${response.statusText}`);
        }

        return response.json();
    }
}

// Export singleton instance
export const cloudbox = new CloudBoxAPI();
EOF

cip_success "CloudBox API integration configured"

# Step 10: Health check endpoint setup
cip_step "Setting up health check endpoint"

# Create health check route
mkdir -p src/routes/health
cat > src/routes/health/+server.js << 'EOF'
import { json } from '@sveltejs/kit/json';

export async function GET() {
    const health = {
        status: 'healthy',
        timestamp: new Date().toISOString(),
        version: process.env.npm_package_version || '1.0.0',
        environment: process.env.NODE_ENV || 'development',
        cloudbox: {
            api_url: process.env.API_URL,
            project_id: process.env.PROJECT_ID,
            port: process.env.PORT
        }
    };

    return json(health);
}
EOF

cip_success "Health check endpoint created"

# Step 11: Docker setup (if enabled)
if [[ "${CLOUDBOX_DOCKER_ENABLED}" == "true" ]]; then
    cip_step "Setting up Docker monitoring"
    cip_docker_setup
else
    cip_info "Docker monitoring disabled - skipping Docker setup"
fi

# Step 12: File permissions and security
cip_step "Setting file permissions and security"
chmod +x scripts/cloudbox-*.sh 2>/dev/null || true
chmod 644 .env.production
chmod -R 755 build/ 2>/dev/null || true
cip_success "File permissions configured"

# Step 13: Pre-start validation
cip_step "Performing pre-start validation"

# Check if port is available
if ! cip_check_port "${CLOUDBOX_WEB_PORT}"; then
    cip_warn "Port ${CLOUDBOX_WEB_PORT} appears to be in use"
    cip_info "The start script will handle port conflicts"
fi

# Verify build integrity
if [[ -f "build/index.js" ]]; then
    if node -c build/index.js; then
        cip_success "Build integrity verified"
    else
        cip_fatal "Build output contains syntax errors"
    fi
else
    cip_fatal "Build output missing - build may have failed"
fi

# Step 14: Generate startup information
cat > .cloudbox-info.json << EOF
{
  "app_name": "$(cip_get_app_name)",
  "app_version": "$(cip_get_app_version)",
  "installed_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "cloudbox_version": "1.0",
  "node_version": "$NODE_VERSION",
  "installation_path": "$(pwd)",
  "configuration": {
    "port": ${CLOUDBOX_WEB_PORT},
    "api_url": "${CLOUDBOX_API_URL}",
    "project_id": "${CLOUDBOX_PROJECT_ID}",
    "environment": "${CLOUDBOX_ENVIRONMENT:-production}"
  },
  "features": {
    "docker_enabled": ${CLOUDBOX_DOCKER_ENABLED:-false},
    "health_check": true,
    "cloudbox_integration": true,
    "photo_upload": true
  }
}
EOF

# Installation complete
echo
cip_success "🎉 Installation completed successfully!"
echo
cip_info "📋 Installation Summary:"
cip_info "  App: $(cip_get_app_name) v$(cip_get_app_version)"
cip_info "  Type: SvelteKit Photography Portfolio"
cip_info "  Port: ${CLOUDBOX_WEB_PORT}"
cip_info "  API: ${CLOUDBOX_API_URL}"
cip_info "  Path: $(pwd)"
echo
cip_info "🚀 Next Steps:"
cip_info "  1. Run './scripts/cloudbox-start.sh' to start the application"
cip_info "  2. Access your portfolio at http://localhost:${CLOUDBOX_WEB_PORT}"
cip_info "  3. Check status with './scripts/cloudbox-status.sh'"
cip_info "  4. View health with './scripts/cloudbox-health.sh'"
echo
cip_info "🔗 Useful URLs:"
cip_info "  Portfolio: http://localhost:${CLOUDBOX_WEB_PORT}"
cip_info "  Health Check: http://localhost:${CLOUDBOX_WEB_PORT}/health"
cip_info "  CloudBox API: ${CLOUDBOX_API_URL}"

# Final resource check
cip_debug "Final resource check"
cip_check_resources

cip_success "✅ SvelteKit Portfolio is ready for deployment!"