#!/bin/bash
# CloudBox Project Setup Template
# 
# Usage: ./setup-template.sh <project_id> <api_key> [endpoint]
#
# This script sets up CloudBox collections and storage buckets
# for your application. Customize the collections and buckets
# sections below based on your specific needs.

set -e

# Configuration
CLOUDBOX_PROJECT_ID="${1:-}"
CLOUDBOX_API_KEY="${2:-}"
CLOUDBOX_ENDPOINT="${3:-http://localhost:8080}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Validate inputs
if [[ -z "$CLOUDBOX_PROJECT_ID" ]] || [[ -z "$CLOUDBOX_API_KEY" ]]; then
    echo -e "${RED}Usage: $0 <project_id> <api_key> [endpoint]${NC}"
    echo "Example: $0 9 cb_abc123xyz http://localhost:8080"
    exit 1
fi

echo -e "${GREEN}CloudBox Setup Starting...${NC}"
echo "Project ID: $CLOUDBOX_PROJECT_ID"
echo "Endpoint: $CLOUDBOX_ENDPOINT"
echo ""

# Helper function to create collections
create_collection() {
    local name="$1"
    shift
    local schema='['
    for field in "$@"; do
        schema+='"'$field'",'
    done
    schema="${schema%,}]"
    
    echo -e "${YELLOW}Creating collection: $name${NC}"
    
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST \
        "${CLOUDBOX_ENDPOINT}/p/${CLOUDBOX_PROJECT_ID}/api/collections" \
        -H "X-API-Key: ${CLOUDBOX_API_KEY}" \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"$name\",\"schema\":$schema}")
    
    http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    body=$(echo $response | sed -e 's/HTTPSTATUS\:.*//g')
    
    if [[ "$http_code" -eq 200 ]] || [[ "$http_code" -eq 201 ]]; then
        echo -e "${GREEN}✓ Collection '$name' created successfully${NC}"
    elif [[ "$body" == *"already exists"* ]]; then
        echo -e "${YELLOW}→ Collection '$name' already exists${NC}"
    else
        echo -e "${RED}✗ Failed to create collection '$name': HTTP $http_code${NC}"
        echo -e "${RED}  Response: $body${NC}"
    fi
}

# Helper function to create storage buckets
create_bucket() {
    local name="$1"
    local description="$2"
    local is_public="${3:-true}"
    local max_size="${4:-52428800}"  # 50MB default
    
    echo -e "${YELLOW}Creating bucket: $name${NC}"
    
    response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X POST \
        "${CLOUDBOX_ENDPOINT}/p/${CLOUDBOX_PROJECT_ID}/api/storage/buckets" \
        -H "X-API-Key: ${CLOUDBOX_API_KEY}" \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"$name\",\"description\":\"$description\",\"is_public\":$is_public,\"max_file_size\":$max_size}")
    
    http_code=$(echo $response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')
    body=$(echo $response | sed -e 's/HTTPSTATUS\:.*//g')
    
    if [[ "$http_code" -eq 200 ]] || [[ "$http_code" -eq 201 ]]; then
        echo -e "${GREEN}✓ Bucket '$name' created successfully${NC}"
    elif [[ "$body" == *"already exists"* ]]; then
        echo -e "${YELLOW}→ Bucket '$name' already exists${NC}"
    else
        echo -e "${RED}✗ Failed to create bucket '$name': HTTP $http_code${NC}"
        echo -e "${RED}  Response: $body${NC}"
    fi
}

# Test connection first
echo -e "${YELLOW}Testing CloudBox connection...${NC}"
health_response=$(curl -s -w "HTTPSTATUS:%{http_code}" "${CLOUDBOX_ENDPOINT}/health")
health_code=$(echo $health_response | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

if [[ "$health_code" -ne 200 ]]; then
    echo -e "${RED}✗ Cannot connect to CloudBox at $CLOUDBOX_ENDPOINT${NC}"
    echo -e "${RED}  Make sure CloudBox is running and accessible${NC}"
    exit 1
fi

# Test API key
api_test=$(curl -s -w "HTTPSTATUS:%{http_code}" \
    "${CLOUDBOX_ENDPOINT}/p/${CLOUDBOX_PROJECT_ID}/api/collections" \
    -H "X-API-Key: ${CLOUDBOX_API_KEY}")
api_code=$(echo $api_test | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

if [[ "$api_code" -eq 401 ]] || [[ "$api_code" -eq 403 ]]; then
    echo -e "${RED}✗ Invalid API key or project ID${NC}"
    echo -e "${RED}  Please check your credentials${NC}"
    exit 1
elif [[ "$api_code" -ne 200 ]]; then
    echo -e "${RED}✗ API test failed: HTTP $api_code${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Connection and API key verified${NC}"
echo ""

# =============================================================================
# CUSTOMIZE THIS SECTION FOR YOUR PROJECT
# =============================================================================

echo -e "${GREEN}Creating collections...${NC}"

# Example: Photo Portfolio Collections
# Uncomment and modify based on your needs:

# create_collection "pages" \
#     "title:string" \
#     "content:text" \
#     "path:string" \
#     "language:string" \
#     "published:boolean" \
#     "created_at:datetime" \
#     "updated_at:datetime"

# create_collection "albums" \
#     "name:string" \
#     "description:text" \
#     "cover_image_id:string" \
#     "images:array" \
#     "published:boolean" \
#     "sort_order:integer" \
#     "created_at:datetime"

# create_collection "images" \
#     "original_filename:string" \
#     "storage_path:string" \
#     "file_size:integer" \
#     "mime_type:string" \
#     "width:integer" \
#     "height:integer" \
#     "alt_text:string" \
#     "caption:text" \
#     "tags:array" \
#     "created_at:datetime"

# Example: Basic App Collections
# create_collection "users" \
#     "email:string" \
#     "name:string" \
#     "role:string" \
#     "created_at:datetime"

# create_collection "posts" \
#     "title:string" \
#     "content:text" \
#     "author_id:string" \
#     "published:boolean" \
#     "created_at:datetime"

# Example: E-commerce Collections
# create_collection "products" \
#     "name:string" \
#     "description:text" \
#     "price:float" \
#     "currency:string" \
#     "in_stock:boolean" \
#     "category:string" \
#     "images:array" \
#     "created_at:datetime"

# create_collection "orders" \
#     "user_id:string" \
#     "products:array" \
#     "total:float" \
#     "status:string" \
#     "shipping_address:json" \
#     "created_at:datetime"

echo ""
echo -e "${GREEN}Creating storage buckets...${NC}"

# Example: Photo Portfolio Buckets
# create_bucket "images" "Portfolio images and photos" true 10485760
# create_bucket "thumbnails" "Generated thumbnail images" true 2097152
# create_bucket "branding" "Site branding assets" true 5242880

# Example: General App Buckets
# create_bucket "uploads" "User uploaded files" true 52428800
# create_bucket "avatars" "User profile pictures" true 5242880

# Example: Private Buckets
# create_bucket "documents" "Private documents" false 104857600

# =============================================================================
# END CUSTOMIZATION SECTION
# =============================================================================

echo ""
echo -e "${GREEN}CloudBox setup completed!${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Update your application's environment variables:"
echo "   CLOUDBOX_ENDPOINT=$CLOUDBOX_ENDPOINT"
echo "   CLOUDBOX_PROJECT_ID=$CLOUDBOX_PROJECT_ID"
echo "   CLOUDBOX_API_KEY=$CLOUDBOX_API_KEY"
echo ""
echo "2. Install the CloudBox SDK in your project:"
echo "   npm install @ekoppen/cloudbox-sdk"
echo ""
echo "3. Use the SDK in your application:"
echo "   import { CloudBoxClient } from '@ekoppen/cloudbox-sdk';"
echo "   const cloudbox = new CloudBoxClient({...});"
echo ""
echo -e "${GREEN}Happy coding! 🚀${NC}"