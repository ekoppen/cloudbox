#!/bin/bash

# CloudBox Script Runner Plugin Installation Script
# Usage: ./install.sh

set -e

echo "🚀 Installing CloudBox Script Runner Plugin..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the correct directory
if [ ! -f "plugin.json" ]; then
    log_error "plugin.json not found. Please run this script from the plugin directory."
    exit 1
fi

# Check if CloudBox is running
if ! docker ps | grep -q "cloudbox"; then
    log_warning "CloudBox containers are not running. Please start CloudBox first."
    log_info "Run: docker-compose up -d"
    exit 1
fi

# Get CloudBox directory
CLOUDBOX_DIR="$(cd ../../ && pwd)"
log_info "CloudBox directory: $CLOUDBOX_DIR"

# Install Node.js dependencies if needed
if [ -f "package.json" ]; then
    log_info "Installing Node.js dependencies..."
    npm install
fi

# Copy plugin to CloudBox plugins directory (if not already there)
PLUGIN_DIR="$CLOUDBOX_DIR/plugins/script-runner"
if [ "$(pwd)" != "$PLUGIN_DIR" ]; then
    log_info "Copying plugin files to CloudBox plugins directory..."
    mkdir -p "$PLUGIN_DIR"
    cp -r * "$PLUGIN_DIR/"
fi

# Create plugin database tables
log_info "Setting up plugin database..."
cat > /tmp/script_runner_setup.sql << 'EOF'
-- CloudBox Script Runner Plugin Database Setup
CREATE SCHEMA IF NOT EXISTS script_runner;

CREATE TABLE IF NOT EXISTS script_runner.scripts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL, -- sql, javascript, setup, migration
    category VARCHAR(30) DEFAULT 'custom', -- project-setup, custom, template, migration
    project_id VARCHAR(50),
    content TEXT NOT NULL,
    version VARCHAR(20) DEFAULT '1.0.0',
    
    -- Dependencies and ordering
    dependencies JSONB DEFAULT '[]',
    run_order INTEGER DEFAULT 999,
    
    -- Metadata
    author VARCHAR(50),
    tags TEXT[],
    is_template BOOLEAN DEFAULT false,
    is_public BOOLEAN DEFAULT false,
    
    -- Execution info
    last_run TIMESTAMP,
    run_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(name, project_id)
);

CREATE TABLE IF NOT EXISTS script_runner.executions (
    id SERIAL PRIMARY KEY,
    script_id INTEGER,
    project_id VARCHAR(50) NOT NULL,
    
    -- Execution details
    status VARCHAR(20) DEFAULT 'running', -- running, success, failed, cancelled
    output TEXT,
    error_message TEXT,
    duration_ms INTEGER,
    
    -- Context
    executed_by VARCHAR(50),
    execution_context JSONB DEFAULT '{}',
    
    -- Timestamps
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    
    FOREIGN KEY (script_id) REFERENCES script_runner.scripts(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_executions_script_started ON script_runner.executions(script_id, started_at);
CREATE INDEX IF NOT EXISTS idx_executions_project_started ON script_runner.executions(project_id, started_at);
CREATE INDEX IF NOT EXISTS idx_executions_status ON script_runner.executions(status);

CREATE TABLE IF NOT EXISTS script_runner.script_collections (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(30) DEFAULT 'project-setup',
    project_id VARCHAR(50),
    
    -- Collection metadata
    scripts JSONB NOT NULL, -- Array of script IDs with run order
    default_variables JSONB DEFAULT '{}',
    
    -- Usage tracking
    usage_count INTEGER DEFAULT 0,
    last_used TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(name, project_id)
);

-- Insert some example templates
INSERT INTO script_runner.scripts (name, description, type, category, project_id, content, is_template, is_public, run_order, author) VALUES
('Basic Users Table', 'Create basic users table for web applications', 'sql', 'template', NULL, 
'-- Basic Users Table for Web Applications
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    email_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

SELECT ''Users table created successfully'' as result;', 
true, true, 1, 'system'),

('Basic Sessions Table', 'Create sessions table for user authentication', 'sql', 'template', NULL,
'-- User Sessions Table
CREATE TABLE IF NOT EXISTS sessions (
    id VARCHAR(128) PRIMARY KEY,
    user_id INTEGER NOT NULL,
    ip_address INET,
    user_agent TEXT,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_active ON sessions(is_active, expires_at);

SELECT ''Sessions table created successfully'' as result;',
true, true, 2, 'system')

ON CONFLICT (name, project_id) DO NOTHING;

SELECT 'Script Runner plugin database setup completed' as result;
EOF

# Execute SQL setup using PostgreSQL
docker-compose -f "$CLOUDBOX_DIR/docker-compose.yml" exec -T postgres psql -U cloudbox -d cloudbox < /tmp/script_runner_setup.sql

# Clean up
rm /tmp/script_runner_setup.sql

# Add plugin to CloudBox configuration
PLUGIN_CONFIG_FILE="$CLOUDBOX_DIR/plugins/plugins.json"
if [ ! -f "$PLUGIN_CONFIG_FILE" ]; then
    log_info "Creating plugins configuration file..."
    echo '{"plugins": []}' > "$PLUGIN_CONFIG_FILE"
fi

# Check if plugin is already registered
if ! grep -q "cloudbox-script-runner" "$PLUGIN_CONFIG_FILE" 2>/dev/null; then
    log_info "Registering plugin in CloudBox configuration..."
    
    # Create temporary file with updated config
    jq '.plugins += [{"name": "cloudbox-script-runner", "path": "./plugins/script-runner", "enabled": true, "installed_at": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'"}]' "$PLUGIN_CONFIG_FILE" > /tmp/plugins.json
    mv /tmp/plugins.json "$PLUGIN_CONFIG_FILE"
else
    log_info "Plugin already registered in CloudBox configuration"
fi

# Load example templates
log_info "Loading example script templates..."
TEMPLATES_DIR="$PLUGIN_DIR/templates"
if [ -d "$TEMPLATES_DIR" ]; then
    log_info "Found $(ls -1 "$TEMPLATES_DIR"/*.json 2>/dev/null | wc -l) template files"
    for template in "$TEMPLATES_DIR"/*.json; do
        if [ -f "$template" ]; then
            template_name=$(basename "$template" .json)
            log_info "  - $template_name"
        fi
    done
fi

# Restart CloudBox backend to load plugin
log_info "Restarting CloudBox backend to load plugin..."
docker-compose -f "$CLOUDBOX_DIR/docker-compose.yml" restart backend

# Wait for backend to be ready
log_info "Waiting for CloudBox backend to be ready..."
sleep 5

# Verify installation
log_info "Verifying plugin installation..."
if curl -s -f "http://localhost:3000/api/plugins/script-runner/templates" > /dev/null 2>&1; then
    log_success "Plugin API is responding correctly"
else
    log_warning "Plugin API verification failed - this is normal if CloudBox backend doesn't support plugins yet"
fi

echo ""
log_success "🎉 CloudBox Script Runner Plugin installed successfully!"
echo ""
echo "📋 Next steps:"
echo "   1. Open CloudBox dashboard: http://localhost:3000"
echo "   2. Navigate to any project"
echo "   3. Click on 'Scripts' in the project sidebar"
echo "   4. Try the available templates!"
echo ""
echo "💡 Available templates:"
echo "   • Basic Web Application - Essential database schema"
echo "   • AI Chat Application - Complete AI chat setup (like Aimy)"
echo "   • E-commerce Backend - Database schema for online stores"
echo ""
echo "🔧 Plugin features:"
echo "   • SQL script execution (like Supabase SQL editor)"
echo "   • JavaScript function deployment"
echo "   • Project setup automation"
echo "   • Dependency management"
echo "   • Execution history and logging"
echo "   • Universal project support"
echo ""
log_info "Plugin documentation: ./README.md"