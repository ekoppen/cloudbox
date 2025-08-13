#!/bin/bash
# CloudBox Interactive Project Setup
# 
# This script provides an interactive setup for CloudBox projects.
# It handles both new project creation and existing project configuration.
#
# Features:
# - Interactive prompts for all configuration
# - Validates CloudBox connection and credentials
# - Creates project structure in CloudBox if needed
# - Generates Docker configuration
# - Creates environment files
# - Sets up collections and storage buckets

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration variables
CLOUDBOX_URL=""
CLOUDBOX_PROJECT_CODE=""
CLOUDBOX_API_KEY=""
APP_USER=""
APP_PORT=""
IS_EXISTING_PROJECT=""
PROJECT_TYPE=""

# Project templates
declare -A PROJECT_TEMPLATES
PROJECT_TEMPLATES[photo-portfolio]="Photo Portfolio (albums, images, pages)"
PROJECT_TEMPLATES[blog]="Blog/CMS (posts, pages, categories)"
PROJECT_TEMPLATES[ecommerce]="E-commerce (products, orders, customers)"
PROJECT_TEMPLATES[saas]="SaaS Application (users, subscriptions, features)"
PROJECT_TEMPLATES[portfolio]="Developer Portfolio (projects, skills, contact)"
PROJECT_TEMPLATES[custom]="Custom (I'll define my own structure)"

# Helper functions
print_header() {
    echo ""
    echo -e "${CYAN}================================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}================================================${NC}"
    echo ""
}

print_step() {
    echo -e "${YELLOW}► $1${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Input validation functions
validate_url() {
    local url="$1"
    if [[ ! "$url" =~ ^https?:// ]]; then
        return 1
    fi
    return 0
}

validate_port() {
    local port="$1"
    if [[ ! "$port" =~ ^[0-9]+$ ]] || [ "$port" -lt 1 ] || [ "$port" -gt 65535 ]; then
        return 1
    fi
    return 0
}

validate_user() {
    local user="$1"
    if [[ ! "$user" =~ ^[a-zA-Z][a-zA-Z0-9_-]*$ ]]; then
        return 1
    fi
    return 0
}

# CloudBox API functions
test_cloudbox_connection() {
    local url="$1"
    print_step "Testing CloudBox connection..."
    
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" --connect-timeout 10 "$url/health" 2>/dev/null || echo "HTTPSTATUS:000")
    http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    
    if [[ "$http_code" -eq 200 ]]; then
        print_success "CloudBox is accessible"
        return 0
    else
        print_error "Cannot connect to CloudBox at $url"
        return 1
    fi
}

test_api_credentials() {
    local url="$1"
    local project_code="$2"
    local api_key="$3"
    
    print_step "Testing API credentials..."
    
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" \
        "$url/p/$project_code/api/collections" \
        -H "X-API-Key: $api_key" 2>/dev/null || echo "HTTPSTATUS:000")
    http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    
    if [[ "$http_code" -eq 200 ]]; then
        print_success "API credentials are valid"
        return 0
    elif [[ "$http_code" -eq 401 ]] || [[ "$http_code" -eq 403 ]]; then
        print_error "Invalid API key or project code"
        return 1
    else
        print_error "API test failed (HTTP $http_code)"
        return 1
    fi
}

# Collection and bucket creation functions
create_collection() {
    local name="$1"
    shift
    local schema='['
    for field in "$@"; do
        schema+='"'$field'",'
    done
    schema="${schema%,}]"
    
    print_step "Creating collection: $name"
    
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST \
        "$CLOUDBOX_URL/p/$CLOUDBOX_PROJECT_CODE/api/collections" \
        -H "X-API-Key: $CLOUDBOX_API_KEY" \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"$name\",\"schema\":$schema}")
    
    http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    body=$(echo $response | sed -e 's/HTTPSTATUS\:.*//g')
    
    if [[ "$http_code" -eq 200 ]] || [[ "$http_code" -eq 201 ]]; then
        print_success "Collection '$name' created"
    elif [[ "$body" == *"already exists"* ]]; then
        print_info "Collection '$name' already exists"
    else
        print_error "Failed to create collection '$name': $body"
        return 1
    fi
}

create_bucket() {
    local name="$1"
    local description="$2"
    local is_public="${3:-true}"
    local max_size="${4:-52428800}"
    
    print_step "Creating storage bucket: $name"
    
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST \
        "$CLOUDBOX_URL/p/$CLOUDBOX_PROJECT_CODE/api/storage/buckets" \
        -H "X-API-Key: $CLOUDBOX_API_KEY" \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"$name\",\"description\":\"$description\",\"is_public\":$is_public,\"max_file_size\":$max_size}")
    
    http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    body=$(echo $response | sed -e 's/HTTPSTATUS\:.*//g')
    
    if [[ "$http_code" -eq 200 ]] || [[ "$http_code" -eq 201 ]]; then
        print_success "Bucket '$name' created"
    elif [[ "$body" == *"already exists"* ]]; then
        print_info "Bucket '$name' already exists"
    else
        print_error "Failed to create bucket '$name': $body"
        return 1
    fi
}

# Project template functions
setup_photo_portfolio() {
    print_header "Setting up Photo Portfolio structure"
    
    # Collections
    create_collection "pages" \
        "title:string" \
        "content:text" \
        "path:string" \
        "language:string" \
        "published:boolean" \
        "page_type:string" \
        "seo_title:string" \
        "seo_description:text" \
        "created_at:datetime" \
        "updated_at:datetime"
    
    create_collection "albums" \
        "name:string" \
        "description:text" \
        "cover_image_id:string" \
        "images:array" \
        "published:boolean" \
        "sort_order:integer" \
        "created_at:datetime" \
        "updated_at:datetime"
    
    create_collection "images" \
        "original_filename:string" \
        "storage_path:string" \
        "file_size:integer" \
        "mime_type:string" \
        "width:integer" \
        "height:integer" \
        "thumbnails:json" \
        "alt_text:string" \
        "caption:text" \
        "tags:array" \
        "created_at:datetime"
    
    create_collection "settings" \
        "key:string" \
        "value:text" \
        "type:string" \
        "category:string" \
        "updated_at:datetime"
    
    # Storage buckets
    create_bucket "images" "Portfolio images and photos" true 10485760
    create_bucket "thumbnails" "Generated thumbnail images" true 2097152
    create_bucket "branding" "Site branding assets" true 5242880
}

setup_blog() {
    print_header "Setting up Blog/CMS structure"
    
    create_collection "posts" \
        "title:string" \
        "slug:string" \
        "content:text" \
        "excerpt:text" \
        "author_id:string" \
        "category_id:string" \
        "tags:array" \
        "featured_image:string" \
        "published:boolean" \
        "published_at:datetime" \
        "created_at:datetime" \
        "updated_at:datetime"
    
    create_collection "categories" \
        "name:string" \
        "slug:string" \
        "description:text" \
        "color:string" \
        "sort_order:integer"
    
    create_collection "pages" \
        "title:string" \
        "slug:string" \
        "content:text" \
        "template:string" \
        "published:boolean" \
        "created_at:datetime" \
        "updated_at:datetime"
    
    create_collection "authors" \
        "name:string" \
        "email:string" \
        "bio:text" \
        "avatar:string" \
        "social_links:json"
    
    create_bucket "uploads" "Blog media uploads" true 52428800
    create_bucket "avatars" "Author profile pictures" true 5242880
}

setup_ecommerce() {
    print_header "Setting up E-commerce structure"
    
    create_collection "products" \
        "name:string" \
        "slug:string" \
        "description:text" \
        "short_description:text" \
        "price:float" \
        "sale_price:float" \
        "currency:string" \
        "sku:string" \
        "stock_quantity:integer" \
        "manage_stock:boolean" \
        "in_stock:boolean" \
        "category_ids:array" \
        "images:array" \
        "attributes:json" \
        "published:boolean" \
        "created_at:datetime" \
        "updated_at:datetime"
    
    create_collection "categories" \
        "name:string" \
        "slug:string" \
        "description:text" \
        "parent_id:string" \
        "image:string" \
        "sort_order:integer"
    
    create_collection "orders" \
        "order_number:string" \
        "customer_id:string" \
        "status:string" \
        "items:array" \
        "subtotal:float" \
        "tax_amount:float" \
        "shipping_amount:float" \
        "total:float" \
        "currency:string" \
        "billing_address:json" \
        "shipping_address:json" \
        "payment_method:string" \
        "payment_status:string" \
        "created_at:datetime" \
        "updated_at:datetime"
    
    create_collection "customers" \
        "email:string" \
        "first_name:string" \
        "last_name:string" \
        "phone:string" \
        "addresses:array" \
        "created_at:datetime"
    
    create_bucket "products" "Product images" true 10485760
    create_bucket "categories" "Category images" true 5242880
}

setup_saas() {
    print_header "Setting up SaaS Application structure"
    
    create_collection "users" \
        "email:string" \
        "name:string" \
        "avatar:string" \
        "role:string" \
        "subscription_id:string" \
        "last_login:datetime" \
        "email_verified:boolean" \
        "created_at:datetime" \
        "updated_at:datetime"
    
    create_collection "subscriptions" \
        "user_id:string" \
        "plan_id:string" \
        "status:string" \
        "current_period_start:datetime" \
        "current_period_end:datetime" \
        "cancel_at_period_end:boolean" \
        "stripe_subscription_id:string" \
        "created_at:datetime" \
        "updated_at:datetime"
    
    create_collection "plans" \
        "name:string" \
        "description:text" \
        "price:float" \
        "currency:string" \
        "interval:string" \
        "features:array" \
        "limits:json" \
        "active:boolean"
    
    create_collection "usage_logs" \
        "user_id:string" \
        "feature:string" \
        "count:integer" \
        "date:datetime"
    
    create_bucket "avatars" "User profile pictures" true 5242880
    create_bucket "uploads" "User file uploads" false 104857600
}

setup_portfolio() {
    print_header "Setting up Developer Portfolio structure"
    
    create_collection "projects" \
        "title:string" \
        "slug:string" \
        "description:text" \
        "long_description:text" \
        "image:string" \
        "gallery:array" \
        "technologies:array" \
        "github_url:string" \
        "demo_url:string" \
        "status:string" \
        "featured:boolean" \
        "sort_order:integer" \
        "created_at:datetime"
    
    create_collection "skills" \
        "name:string" \
        "category:string" \
        "level:integer" \
        "icon:string" \
        "sort_order:integer"
    
    create_collection "experience" \
        "company:string" \
        "position:string" \
        "description:text" \
        "start_date:datetime" \
        "end_date:datetime" \
        "current:boolean" \
        "technologies:array"
    
    create_collection "contact_messages" \
        "name:string" \
        "email:string" \
        "subject:string" \
        "message:text" \
        "read:boolean" \
        "created_at:datetime"
    
    create_bucket "projects" "Project screenshots and media" true 10485760
    create_bucket "resume" "Resume and documents" false 5242880
}

# Docker configuration generation
generate_docker_config() {
    print_header "Generating Docker configuration"
    
    # Create .env file
    cat > .env << EOF
# CloudBox Configuration
CLOUDBOX_ENDPOINT=$CLOUDBOX_URL
CLOUDBOX_PROJECT_CODE=$CLOUDBOX_PROJECT_CODE
CLOUDBOX_API_KEY=$CLOUDBOX_API_KEY

# Application Configuration
APP_PORT=$APP_PORT
NODE_ENV=production

# Docker Configuration
USER_ID=$(id -u)
GROUP_ID=$(id -g)
APP_USER=$APP_USER
EOF
    
    # Create docker-compose.yml
    cat > docker-compose.yml << EOF
version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
      args:
        - USER_ID=\${USER_ID}
        - GROUP_ID=\${GROUP_ID}
        - APP_USER=\${APP_USER}
    container_name: ${CLOUDBOX_PROJECT_CODE}-app
    restart: unless-stopped
    ports:
      - "\${APP_PORT}:3000"
    environment:
      - CLOUDBOX_ENDPOINT=\${CLOUDBOX_ENDPOINT}
      - CLOUDBOX_PROJECT_CODE=\${CLOUDBOX_PROJECT_CODE}
      - CLOUDBOX_API_KEY=\${CLOUDBOX_API_KEY}
      - NODE_ENV=\${NODE_ENV}
    env_file:
      - .env
    volumes:
      - ./logs:/app/logs
    networks:
      - app-network

networks:
  app-network:
    driver: bridge

volumes:
  app-data:
EOF
    
    # Create sample Dockerfile
    cat > Dockerfile.example << EOF
FROM node:18-alpine

# Create app user
ARG USER_ID=1000
ARG GROUP_ID=1000
ARG APP_USER=appuser

RUN addgroup -g \$GROUP_ID \$APP_USER && \\
    adduser -D -u \$USER_ID -G \$APP_USER \$APP_USER

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy application code
COPY . .

# Change ownership
RUN chown -R \$APP_USER:\$APP_USER /app

# Switch to app user
USER \$APP_USER

# Expose port
EXPOSE 3000

# Start application
CMD ["npm", "start"]
EOF
    
    print_success ".env file created"
    print_success "docker-compose.yml created"
    print_success "Dockerfile.example created"
}

# Interactive prompts
prompt_cloudbox_config() {
    print_header "CloudBox Configuration"
    
    # CloudBox URL
    while true; do
        read -p "CloudBox URL (e.g., http://localhost:8080): " CLOUDBOX_URL
        if validate_url "$CLOUDBOX_URL"; then
            # Remove trailing slash
            CLOUDBOX_URL=${CLOUDBOX_URL%/}
            if test_cloudbox_connection "$CLOUDBOX_URL"; then
                break
            fi
        else
            print_error "Please enter a valid URL (http:// or https://)"
        fi
    done
    
    # Project setup type
    echo ""
    print_step "Is this an existing CloudBox project or should we create a new setup?"
    echo "1) Existing project (I have project code and API key)"
    echo "2) New project setup (create collections and buckets)"
    read -p "Choose option (1-2): " setup_choice
    
    if [[ "$setup_choice" == "1" ]]; then
        IS_EXISTING_PROJECT="yes"
        
        # Project code
        read -p "Project code/slug: " CLOUDBOX_PROJECT_CODE
        
        # API key
        read -s -p "API Key: " CLOUDBOX_API_KEY
        echo ""
        
        # Test credentials
        if ! test_api_credentials "$CLOUDBOX_URL" "$CLOUDBOX_PROJECT_CODE" "$CLOUDBOX_API_KEY"; then
            print_error "Invalid credentials. Please check your project code and API key."
            exit 1
        fi
    else
        IS_EXISTING_PROJECT="no"
        
        # Project code
        read -p "Project code/slug (lowercase, no spaces): " CLOUDBOX_PROJECT_CODE
        
        # API key
        read -s -p "API Key (from CloudBox admin): " CLOUDBOX_API_KEY
        echo ""
        
        # Project type selection
        echo ""
        print_step "Select project template:"
        local i=1
        local template_keys=()
        for key in "${!PROJECT_TEMPLATES[@]}"; do
            echo "$i) ${PROJECT_TEMPLATES[$key]}"
            template_keys+=("$key")
            ((i++))
        done
        
        read -p "Choose template (1-${#PROJECT_TEMPLATES[@]}): " template_choice
        if [[ "$template_choice" -ge 1 ]] && [[ "$template_choice" -le "${#PROJECT_TEMPLATES[@]}" ]]; then
            PROJECT_TYPE="${template_keys[$((template_choice-1))]}"
        else
            print_error "Invalid choice"
            exit 1
        fi
    fi
}

prompt_docker_config() {
    print_header "Docker Configuration"
    
    # App user
    while true; do
        read -p "Application user name (default: appuser): " APP_USER
        APP_USER=${APP_USER:-appuser}
        if validate_user "$APP_USER"; then
            break
        else
            print_error "Invalid username. Use letters, numbers, underscore, hyphen only."
        fi
    done
    
    # App port
    while true; do
        read -p "Application port (default: 3000): " APP_PORT
        APP_PORT=${APP_PORT:-3000}
        if validate_port "$APP_PORT"; then
            break
        else
            print_error "Invalid port. Use a number between 1-65535."
        fi
    done
}

# Main setup process
main() {
    print_header "CloudBox Interactive Project Setup"
    print_info "This script will help you set up a new CloudBox project with Docker configuration."
    echo ""
    
    # Collect configuration
    prompt_cloudbox_config
    prompt_docker_config
    
    # Generate Docker configuration
    generate_docker_config
    
    # Set up project structure if needed
    if [[ "$IS_EXISTING_PROJECT" == "no" ]]; then
        case "$PROJECT_TYPE" in
            photo-portfolio) setup_photo_portfolio ;;
            blog) setup_blog ;;
            ecommerce) setup_ecommerce ;;
            saas) setup_saas ;;
            portfolio) setup_portfolio ;;
            custom) 
                print_info "Custom template selected. You'll need to create collections and buckets manually."
                ;;
        esac
    fi
    
    # Final instructions
    print_header "Setup Complete!"
    print_success "CloudBox project configured successfully"
    echo ""
    print_info "Next steps:"
    echo "1. Review the generated .env file"
    echo "2. Copy Dockerfile.example to Dockerfile and customize if needed"
    echo "3. Install CloudBox SDK: npm install @ekoppen/cloudbox-sdk"
    echo "4. Build and run your Docker container:"
    echo "   docker-compose up --build"
    echo ""
    print_info "Your application will be available at: http://localhost:$APP_PORT"
    echo ""
    print_success "Happy coding! 🚀"
}

# Run main function
main "$@"