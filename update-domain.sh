#!/bin/bash

# CloudBox Domain Configuration Update Script
# Usage: ./update-domain.sh <your-domain.com>

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Check if domain is provided
if [ $# -eq 0 ]; then
    print_error "Domain name is required"
    echo "Usage: $0 <your-domain.com>"
    echo "Example: $0 myserver.com"
    exit 1
fi

DOMAIN="$1"
ENV_FILE=".env"
DOCKER_COMPOSE_CMD=""

# Detect Docker Compose command
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
elif docker compose version &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker compose"
else
    print_error "Docker Compose not found"
    exit 1
fi

print_info "🔧 Updating CloudBox configuration for domain: $DOMAIN"

# Check if .env file exists
if [ ! -f "$ENV_FILE" ]; then
    print_error ".env file not found. Please run ./install.sh first"
    exit 1
fi

# Create backup of current .env
cp "$ENV_FILE" "${ENV_FILE}.backup.$(date +%s)"
print_info "Created backup of current .env file"

# Update .env file with new domain
print_info "Updating environment variables..."

# Update PUBLIC_API_URL to use the domain with HTTPS
sed -i.tmp "s|PUBLIC_API_URL=.*|PUBLIC_API_URL=https://${DOMAIN}:8080|" "$ENV_FILE"

# Update FRONTEND_URL if needed
sed -i.tmp "s|FRONTEND_URL=.*|FRONTEND_URL=https://${DOMAIN}:3000|" "$ENV_FILE"

# Update CORS_ORIGINS
sed -i.tmp "s|CORS_ORIGINS=.*|CORS_ORIGINS=https://${DOMAIN}:3000|" "$ENV_FILE"

# Update DATABASE_URL if it contains variable references
if grep -q "DATABASE_URL=.*\${" "$ENV_FILE"; then
    # Extract current database credentials from .env
    DB_USER=$(grep "^DB_USER=" "$ENV_FILE" | cut -d'=' -f2)
    DB_PASSWORD=$(grep "^DB_PASSWORD=" "$ENV_FILE" | cut -d'=' -f2)
    DB_HOST=$(grep "^DB_HOST=" "$ENV_FILE" | cut -d'=' -f2)
    DB_PORT=$(grep "^DB_PORT=" "$ENV_FILE" | cut -d'=' -f2)
    DB_NAME=$(grep "^DB_NAME=" "$ENV_FILE" | cut -d'=' -f2)
    
    # Update DATABASE_URL with actual values
    sed -i.tmp "s|DATABASE_URL=.*|DATABASE_URL=postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable|" "$ENV_FILE"
fi

# Clean up temp file
rm -f "${ENV_FILE}.tmp"

print_success "Updated .env file with domain: $DOMAIN"

# Show the updated configuration
print_info "Updated configuration:"
echo "  PUBLIC_API_URL=$(grep PUBLIC_API_URL "$ENV_FILE" | cut -d'=' -f2)"
echo "  FRONTEND_URL=$(grep FRONTEND_URL "$ENV_FILE" | cut -d'=' -f2)"
echo "  CORS_ORIGINS=$(grep CORS_ORIGINS "$ENV_FILE" | cut -d'=' -f2)"

# Ask if user wants to rebuild and restart
echo
read -p "🚀 Rebuild and restart CloudBox with new configuration? (Y/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Nn]$ ]]; then
    print_info "Stopping current services..."
    $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml down
    
    print_info "Rebuilding frontend with new environment variables..."
    $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml build frontend --no-cache
    
    print_info "Starting services with new configuration..."
    $DOCKER_COMPOSE_CMD -f docker-compose.prod.yml up -d
    
    print_success "✅ CloudBox has been updated with new domain configuration!"
    
    echo
    print_info "📊 Your CloudBox is now accessible at:"
    echo "   Frontend: https://${DOMAIN}:3000"
    echo "   Backend:  https://${DOMAIN}:8080"
    echo "   Admin:    https://${DOMAIN}:3000/admin"
    echo
    print_info "🔐 Default admin credentials:"
    echo "   Email:    admin@cloudbox.local"
    echo "   Password: admin123"
    echo
    print_warning "⚠️  Important notes:"
    echo "   - Make sure your domain points to this server"
    echo "   - Ensure ports 3000 and 8080 are open in your firewall"
    echo "   - Consider setting up SSL certificates for production use"
    echo "   - Clear your browser cache if you experience login issues"
else
    print_info "Configuration updated but services not restarted"
    print_info "Run 'docker-compose -f docker-compose.prod.yml down && docker-compose -f docker-compose.prod.yml build frontend --no-cache && docker-compose -f docker-compose.prod.yml up -d' to apply changes"
fi

echo
print_success "🎉 Domain configuration update completed!"