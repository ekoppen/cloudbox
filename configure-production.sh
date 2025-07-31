#!/bin/bash

# CloudBox Production Configuration Script
# This script configures the production environment variables for a deployed CloudBox instance

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
    echo "Usage: $0 <your-domain.com> [backend-port] [frontend-port]"
    echo "Example: $0 cloudbox.doorkoppen.nl 8080 3000"
    exit 1
fi

DOMAIN="$1"
BACKEND_PORT="${2:-8080}"
FRONTEND_PORT="${3:-3000}"
ENV_FILE=".env"

print_info "🔧 Configuring CloudBox for production domain: $DOMAIN"

# Check if .env file exists
if [ ! -f "$ENV_FILE" ]; then
    print_error ".env file not found. Please run ./install.sh first"
    exit 1
fi

# Create backup of current .env
cp "$ENV_FILE" "${ENV_FILE}.backup.$(date +%s)"
print_info "Created backup of current .env file"

# Update .env file with production settings
print_info "Updating environment variables for production..."

# Update PUBLIC_API_URL to use HTTPS with the domain
sed -i.tmp "s|PUBLIC_API_URL=.*|PUBLIC_API_URL=https://${DOMAIN}:${BACKEND_PORT}|" "$ENV_FILE"

# Update FRONTEND_URL if needed
sed -i.tmp "s|FRONTEND_URL=.*|FRONTEND_URL=https://${DOMAIN}:${FRONTEND_PORT}|" "$ENV_FILE"

# Update CORS_ORIGINS to match frontend URL
sed -i.tmp "s|CORS_ORIGINS=.*|CORS_ORIGINS=https://${DOMAIN}:${FRONTEND_PORT}|" "$ENV_FILE"

# Update BASE_URL for backend
sed -i.tmp "s|BASE_URL=.*|BASE_URL=https://${DOMAIN}:${BACKEND_PORT}|" "$ENV_FILE" 2>/dev/null || echo "BASE_URL not found, skipping"

# Update API_URL for backend
sed -i.tmp "s|API_URL=.*|API_URL=https://${DOMAIN}:${BACKEND_PORT}|" "$ENV_FILE" 2>/dev/null || echo "API_URL not found, skipping"

# Set production environment
sed -i.tmp "s|APP_ENV=.*|APP_ENV=production|" "$ENV_FILE"

# Clean up temp file
rm -f "${ENV_FILE}.tmp"

print_success "Updated .env file for production deployment on: $DOMAIN"

# Show the updated configuration
print_info "Updated production configuration:"
echo "  PUBLIC_API_URL=$(grep PUBLIC_API_URL "$ENV_FILE" | cut -d'=' -f2)"
echo "  FRONTEND_URL=$(grep FRONTEND_URL "$ENV_FILE" | cut -d'=' -f2)"
echo "  CORS_ORIGINS=$(grep CORS_ORIGINS "$ENV_FILE" | cut -d'=' -f2)"
echo "  APP_ENV=$(grep APP_ENV "$ENV_FILE" | cut -d'=' -f2)"

print_warning "⚠️  Important: After running this script, you need to:"
print_warning "   1. Restart your containers: docker-compose down && docker-compose up --build -d"
print_warning "   2. Ensure your reverse proxy/load balancer is configured correctly"
print_warning "   3. Verify SSL certificates are valid for $DOMAIN"

print_info "🚀 Production configuration complete!"