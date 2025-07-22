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

OPTIONS:
    -h, --help              Show this help message
    -u, --update            Update existing installation
    -p, --frontend-port     Frontend port (default: 3000)
    -b, --backend-port      Backend port (default: 8080)
    -d, --db-port          Database port (default: 5432)
    -r, --redis-port       Redis port (default: 6379)
    -H, --host             Add allowed host (e.g., server hostname)
    --env-file             Environment file path (default: .env)

EXAMPLES:
    # Basic installation
    ./install.sh

    # Install with custom ports
    ./install.sh -p 8080 -b 9000

    # Install for remote server
    ./install.sh --host myserver.com

    # Update existing installation
    ./install.sh --update

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

# Server Configuration
SERVER_PORT=${BACKEND_PORT}
FRONTEND_URL=http://localhost:${FRONTEND_PORT}
API_URL=http://localhost:${BACKEND_PORT}

# Docker Configuration
FRONTEND_PORT=${FRONTEND_PORT}
BACKEND_PORT=${BACKEND_PORT}
POSTGRES_PORT=${DB_PORT}
REDIS_EXTERNAL_PORT=${REDIS_PORT}

# Application Configuration
APP_ENV=production
LOG_LEVEL=info
CORS_ORIGINS=http://localhost:${FRONTEND_PORT}

# Upload Configuration
MAX_FILE_SIZE=10MB
UPLOAD_DIR=./uploads

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
    if [[ -n "$ALLOWED_HOST" ]]; then
        print_info "Updating Vite configuration with allowed host: ${ALLOWED_HOST}"
        
        local vite_config="frontend/vite.config.ts"
        
        if [[ -f "$vite_config" ]]; then
            # Create allowed hosts array
            local hosts="['${ALLOWED_HOST}', 'localhost', '127.0.0.1', '0.0.0.0']"
            
            # Update vite config
            sed -i.bak "s/allowedHosts: \[.*\]/allowedHosts: ${hosts}/" "$vite_config" || {
                # If allowedHosts doesn't exist, add it
                sed -i.bak "/port: ${FRONTEND_PORT}/a\\
\\t\\tallowedHosts: ${hosts}" "$vite_config"
            }
            
            print_success "Vite configuration updated"
        else
            print_warning "Vite config not found, skipping host configuration"
        fi
    fi
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
        if $DOCKER_COMPOSE_CMD ps --format json | grep -q '"Health":"healthy".*postgres' && \
           $DOCKER_COMPOSE_CMD ps --format json | grep -q '"Health":"healthy".*redis'; then
            print_success "All services are healthy!"
            break
        fi
        
        if [ $attempt -eq 1 ] || [ $((attempt % 10)) -eq 0 ]; then
            print_info "Attempt $attempt/$max_attempts: Waiting for services to be healthy..."
            $DOCKER_COMPOSE_CMD ps
        fi
        
        sleep 5
        attempt=$((attempt + 1))
    done
    
    if [ $attempt -gt $max_attempts ]; then
        print_warning "Services took longer than expected to start"
        print_info "You can check service status with: $DOCKER_COMPOSE_CMD ps"
        print_info "View logs with: $DOCKER_COMPOSE_CMD logs -f"
    fi
    
    # Give backend a moment to complete its startup
    print_info "Allowing backend to complete startup..."
    sleep 10
}

# Function to create default admin user
create_admin_user() {
    print_info "Setting up default admin user..."
    
    # This would typically be done by the backend on first run
    # or through a separate admin setup script
    print_info "Default admin user will be created on first backend startup"
    print_info "Default credentials: admin@cloudbox.local / admin123"
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
    
    # Generate environment file
    generate_env_file
    
    # Update configurations
    update_vite_config
    update_docker_compose
    
    # Pull and build images
    print_info "Pulling Docker images..."
    $DOCKER_COMPOSE_CMD pull
    
    print_info "Building application images..."
    $DOCKER_COMPOSE_CMD build
    
    # Start services
    print_info "Starting CloudBox services..."
    $DOCKER_COMPOSE_CMD up -d
    
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
    $DOCKER_COMPOSE_CMD down
    
    # Backup database
    print_info "Creating database backup..."
    if $DOCKER_COMPOSE_CMD ps -q postgres > /dev/null 2>&1; then
        $DOCKER_COMPOSE_CMD up -d postgres
        sleep 5
        $DOCKER_COMPOSE_CMD exec -T postgres pg_dump -U cloudbox cloudbox > "backup_$(date +%Y%m%d_%H%M%S).sql"
        print_success "Database backup created"
    fi
    
    # Pull latest images
    print_info "Pulling latest Docker images..."
    $DOCKER_COMPOSE_CMD pull
    
    # Rebuild images
    print_info "Rebuilding application images..."
    $DOCKER_COMPOSE_CMD build --no-cache
    
    # Update configurations if host is specified
    if [[ -n "$ALLOWED_HOST" ]]; then
        update_vite_config
    fi
    
    # Start services
    print_info "Starting updated services..."
    $DOCKER_COMPOSE_CMD up -d
    
    print_success "CloudBox update completed!"
    show_access_info
}

# Function to show access information
show_access_info() {
    echo
    print_success "🚀 CloudBox is now running!"
    echo
    echo "📊 Access your CloudBox installation:"
    echo "   Frontend:  http://localhost:${FRONTEND_PORT}"
    echo "   Backend:   http://localhost:${BACKEND_PORT}"
    echo "   Admin:     http://localhost:${FRONTEND_PORT}/admin"
    echo
    if [[ -n "$ALLOWED_HOST" ]]; then
        echo "   Remote:    http://${ALLOWED_HOST}:${FRONTEND_PORT}"
        echo "   Admin:     http://${ALLOWED_HOST}:${FRONTEND_PORT}/admin"
        echo
    fi
    echo "🔐 Default admin credentials:"
    echo "   Email:     admin@cloudbox.local"
    echo "   Password:  admin123"
    echo
    echo "📝 Useful commands:"
    echo "   View logs:    $DOCKER_COMPOSE_CMD logs -f"
    echo "   Stop:         $DOCKER_COMPOSE_CMD down"
    echo "   Restart:      $DOCKER_COMPOSE_CMD restart"
    echo "   Update:       ./install.sh --update"
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