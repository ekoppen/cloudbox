#!/bin/bash

# CloudBox Installation and Update Script
# Usage: ./install.sh [options]

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
FRONTEND_PORT=3000
BACKEND_PORT=8080
DB_PORT=5432
REDIS_PORT=6379
ALLOWED_HOST=""
API_HOST=""
INSTALL_MODE="install"
ENV_FILE=".env"

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    cat << EOF
CloudBox Installation Script

USAGE:
    ./install.sh [OPTIONS]

DESCRIPTION:
    Interactive installation script for CloudBox Backend-as-a-Service platform.
    If no hostname is specified via --host, you will be prompted to enter one
    during installation for optimal configuration.

OPTIONS:
    -h, --help              Show this help message
    -u, --update            Update existing installation
    -p, --frontend-port     Frontend port (default: 3000)
    -b, --backend-port      Backend port (default: 8080)
    -d, --db-port          Database port (default: 5432)
    -r, --redis-port       Redis port (default: 6379)
    -H, --host             Add allowed host (e.g., server hostname)
    --api-host             API host/domain (e.g., api.example.com)
    --env-file             Environment file path (default: .env)

EXAMPLES:
    # Interactive installation (recommended)
    ./install.sh

    # Silent installation with hostname
    ./install.sh --host myserver.com

    # Install with custom ports
    ./install.sh -p 8080 -b 9000 --host example.com

    # Install with subdomain for API
    ./install.sh --host cloudbox.example.com --api-host api.cloudbox.example.com

    # Local development installation
    ./install.sh -p 3001 -b 8081

    # Update existing installation
    ./install.sh --update

INTERACTIVE MODE:
    When run without --host, the script will prompt you to enter:
    - Hostname or domain name for remote access
    - Port configuration confirmation
    - Installation summary review

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -u|--update)
            INSTALL_MODE="update"
            shift
            ;;
        -p|--frontend-port)
            FRONTEND_PORT="$2"
            shift 2
            ;;
        -b|--backend-port)
            BACKEND_PORT="$2"
            shift 2
            ;;
        -d|--db-port)
            DB_PORT="$2"
            shift 2
            ;;
        -r|--redis-port)
            REDIS_PORT="$2"
            shift 2
            ;;
        -H|--host)
            ALLOWED_HOST="$2"
            shift 2
            ;;
        --api-host)
            API_HOST="$2"
            shift 2
            ;;
        --env-file)
            ENV_FILE="$2"
            shift 2
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Function to detect and set Docker Compose command
detect_docker_compose() {
    if command -v docker-compose &> /dev/null; then
        DOCKER_COMPOSE_CMD="docker-compose"
        print_info "Using docker-compose (standalone)"
    elif docker compose version &> /dev/null; then
        DOCKER_COMPOSE_CMD="docker compose"
        print_info "Using docker compose (Docker CLI plugin)"
    else
        print_error "Docker Compose not found. Please install Docker Compose."
        print_info "Install Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi
}

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        print_info "Install Docker: https://docs.docker.com/get-docker/"
        exit 1
    fi
    
    # Detect Docker Compose command
    detect_docker_compose
    
    # Check if curl is installed (needed for health checks)
    if ! command -v curl &> /dev/null; then
        print_error "curl is not installed. Please install curl first."
        print_info "Install curl: sudo apt-get install curl (Ubuntu/Debian) or brew install curl (macOS)"
        exit 1
    fi
    
    # Check if ports are available
    if ss -tlnp | grep -q ":${FRONTEND_PORT} "; then
        print_warning "Port ${FRONTEND_PORT} is already in use. Frontend might conflict."
    fi
    
    if ss -tlnp | grep -q ":${BACKEND_PORT} "; then
        print_warning "Port ${BACKEND_PORT} is already in use. Backend might conflict."
    fi
    
    print_success "Prerequisites check completed"
}

# Function to generate environment file
generate_env_file() {
    print_info "Generating environment file: ${ENV_FILE}"
    
    # Generate random JWT secret
    JWT_SECRET=$(openssl rand -hex 32 2>/dev/null || echo "your-super-secret-jwt-key-$(date +%s)")
    
    # Determine API URL based on configuration
    if [[ -n "$API_HOST" ]]; then
        PUBLIC_API_URL="https://${API_HOST}"
        print_info "Using API subdomain: $API_HOST"
    elif [[ -n "$ALLOWED_HOST" ]]; then
        PUBLIC_API_URL="https://${ALLOWED_HOST}:${BACKEND_PORT}"
        print_info "Using main domain with backend port: $ALLOWED_HOST:$BACKEND_PORT"
    else
        PUBLIC_API_URL="http://localhost:${BACKEND_PORT}"
        print_info "Using localhost configuration for development"
    fi
    
    cat > "${ENV_FILE}" << EOF
# CloudBox Environment Configuration
# Generated on $(date)

# Database Configuration
DB_HOST=postgres
DB_PORT=${DB_PORT}
DB_USER=cloudbox
DB_PASSWORD=cloudbox_secure_password_$(date +%s | tail -c 6)
DB_NAME=cloudbox
DATABASE_URL=postgresql://\${DB_USER}:\${DB_PASSWORD}@\${DB_HOST}:\${DB_PORT}/\${DB_NAME}?sslmode=disable

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=${REDIS_PORT}
REDIS_URL=redis://\${REDIS_HOST}:\${REDIS_PORT}

# JWT Configuration
JWT_SECRET=${JWT_SECRET}
JWT_EXPIRES_IN=24h
REFRESH_TOKEN_EXPIRES_IN=720h

# Security Configuration
MASTER_KEY=$(openssl rand -hex 32 2>/dev/null || echo "master-key-$(date +%s)")

# Server Configuration
SERVER_PORT=${BACKEND_PORT}
API_URL=http://localhost:${BACKEND_PORT}

# Docker Configuration
FRONTEND_PORT=${FRONTEND_PORT}
BACKEND_PORT=${BACKEND_PORT}
POSTGRES_PORT=${DB_PORT}
REDIS_EXTERNAL_PORT=${REDIS_PORT}

# Application Configuration
APP_ENV=production
LOG_LEVEL=info
EOF

    # Set CORS origins based on configuration
    if [[ -n "$ALLOWED_HOST" ]]; then
        echo "CORS_ORIGINS=http://${ALLOWED_HOST}:${FRONTEND_PORT},https://${ALLOWED_HOST}:${FRONTEND_PORT}" >> "${ENV_FILE}"
        echo "FRONTEND_URL=http://${ALLOWED_HOST}:${FRONTEND_PORT}" >> "${ENV_FILE}"
        print_info "CORS configured for hostname: $ALLOWED_HOST"
    else
        echo "CORS_ORIGINS=http://localhost:${FRONTEND_PORT}" >> "${ENV_FILE}"
        echo "FRONTEND_URL=http://localhost:${FRONTEND_PORT}" >> "${ENV_FILE}"
        print_info "CORS configured for localhost"
    fi

    cat >> "${ENV_FILE}" << EOF

# Upload Configuration
MAX_FILE_SIZE=10MB
UPLOAD_DIR=./uploads

# Frontend Configuration
PUBLIC_API_URL=${PUBLIC_API_URL}

# Email Configuration (Optional)
# SMTP_HOST=
# SMTP_PORT=587
# SMTP_USER=
# SMTP_PASS=
# FROM_EMAIL=noreply@cloudbox.local

EOF

    print_success "Environment file created: ${ENV_FILE}"
}

# Function to update Vite config with allowed hosts
update_vite_config() {
    local vite_config="frontend/vite.config.ts"
    
    if [[ ! -f "$vite_config" ]]; then
        print_warning "Vite config not found, skipping host configuration"
        return
    fi
    
    print_info "Updating Vite configuration..."
    
    # Create backup
    cp "$vite_config" "${vite_config}.backup"
    
    # Build allowed hosts array
    local hosts_array=""
    if [[ -n "$ALLOWED_HOST" ]]; then
        hosts_array="'${ALLOWED_HOST}', 'localhost', '127.0.0.1', '0.0.0.0'"
        print_info "Adding allowed host: $ALLOWED_HOST"
    else
        hosts_array="'localhost', '127.0.0.1', '0.0.0.0'"
        print_info "Configuring for localhost access only"
    fi
    
    # Update or add allowedHosts configuration
    if grep -q "allowedHosts:" "$vite_config"; then
        # Update existing allowedHosts
        sed -i.tmp "s/allowedHosts: \[.*\]/allowedHosts: [${hosts_array}]/" "$vite_config"
        print_success "Updated existing allowedHosts configuration"
    else
        # Add allowedHosts to server configuration
        if grep -q "host: '0.0.0.0'," "$vite_config"; then
            # Add after host line
            sed -i.tmp "/host: '0.0.0.0',/a\\
\\t\\tallowedHosts: [${hosts_array}]," "$vite_config"
            print_success "Added allowedHosts to server configuration"
        else
            print_warning "Could not automatically add allowedHosts, manual configuration may be needed"
        fi
    fi
    
    # Clean up temporary files
    rm -f "${vite_config}.tmp"
    
    # Show the relevant part of the config
    print_info "Vite server configuration:"
    grep -A 5 -B 2 "server:" "$vite_config" | head -10
}

# Function to update Docker Compose with custom ports
update_docker_compose() {
    print_info "Updating Docker Compose configuration..."
    
    local compose_file="docker-compose.yml"
    
    if [[ -f "$compose_file" ]]; then
        # Backup original
        cp "$compose_file" "${compose_file}.bak"
        
        # Update ports in docker-compose.yml using environment variables
        sed -i.tmp "s/- \"3000:3000\"/- \"\${FRONTEND_PORT:-3000}:3000\"/" "$compose_file"
        sed -i.tmp "s/- \"8080:8080\"/- \"\${BACKEND_PORT:-8080}:8080\"/" "$compose_file"
        sed -i.tmp "s/- \"5432:5432\"/- \"\${POSTGRES_PORT:-5432}:5432\"/" "$compose_file"
        sed -i.tmp "s/- \"6379:6379\"/- \"\${REDIS_EXTERNAL_PORT:-6379}:6379\"/" "$compose_file"
        
        rm -f "${compose_file}.tmp"
        print_success "Docker Compose configuration updated"
    else
        print_error "docker-compose.yml not found!"
        exit 1
    fi
}

# Function to initialize database
init_database() {
    print_info "Initializing database and services..."
    
    # Wait for all services to be healthy
    print_info "Waiting for services to become healthy..."
    local max_attempts=60
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml ps | grep -q "postgres.*healthy" && \
           $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml ps | grep -q "redis.*healthy"; then
            print_success "All services are healthy!"
            break
        fi
        
        if [ $attempt -eq 1 ] || [ $((attempt % 10)) -eq 0 ]; then
            print_info "Attempt $attempt/$max_attempts: Waiting for services to be healthy..."
            $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml ps
        fi
        
        sleep 5
        attempt=$((attempt + 1))
    done
    
    if [ $attempt -gt $max_attempts ]; then
        print_warning "Services took longer than expected to start"
        print_info "You can check service status with: $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml ps"
        print_info "View logs with: $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml logs -f"
    fi
    
    # Give backend a moment to complete its startup
    print_info "Allowing backend to complete startup..."
    sleep 10
}

# Function to create default admin user
create_admin_user() {
    print_info "Setting up default admin user..."
    
    # Wait for backend to be fully ready
    print_info "Waiting for backend to be ready..."
    local max_attempts=60  # Increased from 30 to 60 (2 minutes)
    local attempt=1
    
    # First wait for Docker containers to be healthy
    print_info "Checking Docker container health..."
    local container_wait=0
    while [ $container_wait -lt 30 ]; do
        if $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml ps | grep -q "cloudbox-backend.*healthy"; then
            print_info "Backend container is healthy!"
            break
        fi
        sleep 2
        container_wait=$((container_wait + 1))
    done
    
    # Now check the HTTP health endpoint
    print_info "Testing backend HTTP endpoint..."
    while [ $attempt -le $max_attempts ]; do
        # Check both localhost and container network
        if curl -f -s http://localhost:${BACKEND_PORT}/health >/dev/null 2>&1 || \
           curl -f -s http://127.0.0.1:${BACKEND_PORT}/health >/dev/null 2>&1; then
            print_success "Backend HTTP endpoint is ready!"
            # Extra wait to ensure database connections are stable
            sleep 3
            break
        fi
        
        if [ $((attempt % 10)) -eq 0 ]; then
            print_info "Attempt $attempt/$max_attempts: Backend HTTP endpoint not ready yet..."
        fi
        
        sleep 3  # Increased from 2 to 3 seconds
        attempt=$((attempt + 1))
    done
    
    if [ $attempt -gt $max_attempts ]; then
        print_warning "Backend took too long to start, admin user creation may fail"
        print_info "You can try running the installation again or create the admin user manually"
    fi
    
    # Create default admin user in database
    print_info "Creating default admin user..."
    
    # Password hash for 'admin123' using bcrypt
    local password_hash='\$2a\$10\$72Eg6eu/TToC/T5MzsnEuOwbmp8ITu0m1LfYiDY3KmGofxkwEZCD.'
    
    local create_user_sql="
    INSERT INTO users (email, password_hash, name, role, is_active, created_at, updated_at) 
    VALUES (
        'admin@cloudbox.local', 
        '${password_hash}', 
        'CloudBox Admin',
        'superadmin', 
        true, 
        NOW(), 
        NOW()
    ) ON CONFLICT (email) DO UPDATE SET 
        password_hash = EXCLUDED.password_hash,
        role = EXCLUDED.role,
        name = EXCLUDED.name,
        updated_at = NOW();
    "
    
    if $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml exec -T postgres psql -U cloudbox -d cloudbox -c "$create_user_sql" >/dev/null 2>&1; then
        print_success "Default admin user created successfully!"
        print_info "Login credentials:"
        print_info "  Email:    admin@cloudbox.local"
        print_info "  Password: admin123"
        print_warning "⚠️  Remember to change these credentials after first login!"
    else
        print_warning "Failed to create admin user in database"
        print_info "You can create the admin user manually after installation"
        print_info "Default credentials would be: admin@cloudbox.local / admin123"
    fi
}

# Function to prompt for configuration
prompt_for_configuration() {
    echo
    print_info "📋 CloudBox Installation Configuration"
    echo "======================================"
    echo
    
    # Hostname/Domain prompt
    if [[ -z "$ALLOWED_HOST" ]]; then
        echo "🌐 Domain/Hostname Configuration:"
        echo "   Enter the hostname or domain name for this CloudBox installation."
        echo "   This can be a server hostname, domain name, or IP address."
        echo "   Examples: myserver.com, cloudbox.example.com, 192.168.1.100, mgmt01"
        echo
        read -p "🔗 Hostname/Domain (leave empty for localhost only): " input_host
        
        if [[ -n "$input_host" ]]; then
            ALLOWED_HOST="$input_host"
            print_success "Will configure for hostname: $ALLOWED_HOST"
        else
            print_info "Configuration set for localhost access only"
        fi
        echo
    fi
    
    # API Host configuration prompt
    if [[ -n "$ALLOWED_HOST" ]] && [[ -z "$API_HOST" ]]; then
        echo "🔗 API Configuration:"
        echo "   Do you want to use a separate subdomain for the API?"
        echo "   Example: api.${ALLOWED_HOST} instead of ${ALLOWED_HOST}:8080"
        echo
        read -p "🚀 API subdomain (leave empty to use main domain with port): " input_api_host
        
        if [[ -n "$input_api_host" ]]; then
            API_HOST="$input_api_host"
            print_success "Will configure API subdomain: $API_HOST"
        else
            print_info "Will use main domain with backend port"
        fi
        echo
    fi
    
    # Port configuration confirmation
    if [[ "$FRONTEND_PORT" != "3000" ]] || [[ "$BACKEND_PORT" != "8080" ]]; then
        print_info "🔧 Custom port configuration detected:"
        echo "   Frontend: $FRONTEND_PORT"
        echo "   Backend:  $BACKEND_PORT"
        echo
    fi
    
    # Summary
    print_info "📝 Installation Summary:"
    if [[ -n "$ALLOWED_HOST" ]]; then
        echo "   Frontend URL: https://${ALLOWED_HOST}:$FRONTEND_PORT"
        if [[ -n "$API_HOST" ]]; then
            echo "   API URL:      https://${API_HOST}"
        else
            echo "   API URL:      https://${ALLOWED_HOST}:$BACKEND_PORT"
        fi
        echo "   Admin URL:    https://${ALLOWED_HOST}:$FRONTEND_PORT/admin"
    else
        echo "   Frontend URL: http://localhost:$FRONTEND_PORT"
        echo "   API URL:      http://localhost:$BACKEND_PORT"
        echo "   Admin URL:    http://localhost:$FRONTEND_PORT/admin"
    fi
    echo
    
    read -p "✅ Continue with this configuration? (Y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        print_info "Installation cancelled by user."
        exit 0
    fi
}

# Function to perform installation
perform_install() {
    print_info "Starting CloudBox installation..."
    
    # Check if already installed
    if [[ -f "$ENV_FILE" ]] && [[ "$INSTALL_MODE" == "install" ]]; then
        print_warning "CloudBox appears to be already installed."
        print_info "Use --update flag to update existing installation."
        read -p "Continue with installation anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Installation cancelled."
            exit 0
        fi
    fi
    
    # Prompt for configuration if not provided via CLI
    prompt_for_configuration
    
    # Generate environment file
    generate_env_file
    
    # Update configurations
    update_vite_config
    update_docker_compose
    
    # Pull and build images
    print_info "Pulling Docker images..."
    $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml pull
    
    print_info "Building application images..."
    $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml build
    
    # Start services
    print_info "Starting CloudBox services..."
    $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml up -d
    
    # Initialize database
    init_database
    
    # Create admin user
    create_admin_user
    
    print_success "CloudBox installation completed!"
    show_access_info
}

# Function to perform update
perform_update() {
    print_info "Starting CloudBox update..."
    
    # Stop services
    print_info "Stopping services..."
    $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml down
    
    # Backup database
    print_info "Creating database backup..."
    if $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml ps -q postgres > /dev/null 2>&1; then
        $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml up -d postgres
        sleep 5
        $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml exec -T postgres pg_dump -U cloudbox cloudbox > "backup_$(date +%Y%m%d_%H%M%S).sql"
        print_success "Database backup created"
    fi
    
    # Pull latest images
    print_info "Pulling latest Docker images..."
    $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml pull
    
    # Rebuild images
    print_info "Rebuilding application images..."
    $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml build --no-cache
    
    # Update configurations if host is specified
    if [[ -n "$ALLOWED_HOST" ]]; then
        update_vite_config
    fi
    
    # Start services
    print_info "Starting updated services..."
    $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml up -d
    
    print_success "CloudBox update completed!"
    show_access_info
}

# Function to show access information
show_access_info() {
    echo
    print_success "🚀 CloudBox is now running!"
    echo
    echo "📊 Access your CloudBox installation:"
    
    # Local access
    echo "   📍 Local Access:"
    echo "     Frontend:  http://localhost:${FRONTEND_PORT}"
    echo "     Backend:   http://localhost:${BACKEND_PORT}"
    echo "     Admin:     http://localhost:${FRONTEND_PORT}/admin"
    echo
    
    # Remote access if hostname was configured
    if [[ -n "$ALLOWED_HOST" ]]; then
        echo "   🌐 Remote Access:"
        echo "     Frontend:  http://${ALLOWED_HOST}:${FRONTEND_PORT}"
        echo "     Backend:   http://${ALLOWED_HOST}:${BACKEND_PORT}"
        echo "     Admin:     http://${ALLOWED_HOST}:${FRONTEND_PORT}/admin"
        echo
        print_info "✅ Configured for hostname: $ALLOWED_HOST"
        echo "   Make sure your firewall allows connections on ports $FRONTEND_PORT and $BACKEND_PORT"
        echo
    fi
    
    echo "🔐 Default admin credentials:"
    echo "   Email:     admin@cloudbox.local"
    echo "   Password:  admin123"
    echo "   📝 Remember to change these credentials after first login!"
    echo
    echo "📝 Useful commands:"
    echo "   View logs:    $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml logs -f"
    echo "   Stop:         $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml down"
    echo "   Restart:      $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml restart"
    echo "   Update:       ./install.sh --update"
    if [[ -n "$ALLOWED_HOST" ]]; then
        echo "   Reconfigure: ./install.sh --host $ALLOWED_HOST --update"
    fi
    echo
    echo "📚 Documentation: https://github.com/ekoppen/cloudbox"
    echo
}

# Main execution
main() {
    echo "🚀 CloudBox Installation Script"
    echo "================================"
    echo
    
    check_prerequisites
    
    if [[ "$INSTALL_MODE" == "update" ]]; then
        perform_update
    else
        perform_install
    fi
}

# Run main function
main "$@"