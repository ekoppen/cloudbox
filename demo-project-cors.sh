#!/bin/bash

# CloudBox Project-Specific CORS Configuration Demo
# This script demonstrates how CloudBox already supports project-specific CORS

echo "🔧 CloudBox Project-Specific CORS Demo"
echo "======================================"
echo

CLOUDBOX_URL="http://localhost:8080"
PROJECT_ID="1"  # Using default demo project

# First, let's get an admin JWT token to manage CORS settings
echo "📋 Step 1: Login as admin to get JWT token..."
LOGIN_RESPONSE=$(curl -s -X POST "${CLOUDBOX_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@cloudbox.dev", 
    "password": "admin123"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ Failed to get authentication token"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi

echo "✅ Successfully authenticated"
echo

# Step 2: Get current CORS configuration
echo "📋 Step 2: Get current CORS configuration for project $PROJECT_ID..."
CURRENT_CORS=$(curl -s -X GET "${CLOUDBOX_URL}/api/v1/projects/${PROJECT_ID}/cors" \
  -H "Authorization: Bearer $TOKEN")

echo "Current CORS config:"
echo $CURRENT_CORS | jq .
echo

# Step 3: Update CORS configuration with custom settings
echo "📋 Step 3: Update CORS configuration with custom settings..."
NEW_CORS_CONFIG='{
  "allowed_origins": ["http://localhost:3000", "http://localhost:5173", "https://myapp.example.com"],
  "allowed_methods": ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"],
  "allowed_headers": ["Content-Type", "Authorization", "X-API-Key", "X-Custom-Header"],
  "exposed_headers": ["X-Total-Count", "X-Page-Count"],
  "allow_credentials": true,
  "max_age": 7200
}'

UPDATE_RESPONSE=$(curl -s -X PUT "${CLOUDBOX_URL}/api/v1/projects/${PROJECT_ID}/cors" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$NEW_CORS_CONFIG")

echo "Update response:"
echo $UPDATE_RESPONSE | jq .
echo

# Step 4: Verify the CORS configuration was updated
echo "📋 Step 4: Verify updated CORS configuration..."
UPDATED_CORS=$(curl -s -X GET "${CLOUDBOX_URL}/api/v1/projects/${PROJECT_ID}/cors" \
  -H "Authorization: Bearer $TOKEN")

echo "Updated CORS config:"
echo $UPDATED_CORS | jq .
echo

# Step 5: Test CORS with different origins
echo "📋 Step 5: Test CORS with different origins..."

echo "Testing allowed origin (http://localhost:3000):"
curl -s -X OPTIONS -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET" \
  "${CLOUDBOX_URL}/p/${PROJECT_ID}/api/collections" \
  -I | grep -E "(HTTP|Access-Control)"
echo

echo "Testing allowed origin (https://myapp.example.com):"
curl -s -X OPTIONS -H "Origin: https://myapp.example.com" \
  -H "Access-Control-Request-Method: POST" \
  "${CLOUDBOX_URL}/p/${PROJECT_ID}/api/collections" \
  -I | grep -E "(HTTP|Access-Control)"
echo

echo "Testing blocked origin (https://malicious.example.com):"
curl -s -X OPTIONS -H "Origin: https://malicious.example.com" \
  -H "Access-Control-Request-Method: GET" \
  "${CLOUDBOX_URL}/p/${PROJECT_ID}/api/collections" \
  -I | grep -E "(HTTP|Access-Control)" || echo "No CORS headers (origin blocked)"
echo

# Step 6: Reset to default CORS configuration
echo "📋 Step 6: Reset to default CORS configuration..."
DEFAULT_CORS_CONFIG='{
  "allowed_origins": ["*"],
  "allowed_methods": ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"],
  "allowed_headers": ["*"],
  "exposed_headers": [],
  "allow_credentials": false,
  "max_age": 3600
}'

RESET_RESPONSE=$(curl -s -X PUT "${CLOUDBOX_URL}/api/v1/projects/${PROJECT_ID}/cors" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$DEFAULT_CORS_CONFIG")

echo "Reset response:"
echo $RESET_RESPONSE | jq .
echo

echo "🎉 Demo Complete!"
echo
echo "Summary:"
echo "- ✅ Project-specific CORS configuration is already fully implemented"
echo "- ✅ Each project can have its own allowed origins, methods, and headers"
echo "- ✅ CORS settings are stored in the database per project"
echo "- ✅ Project API routes use project-specific CORS middleware"
echo "- ✅ Admin API routes use global CORS configuration"
echo "- ✅ Fallback to global CORS if project doesn't have specific settings"
echo
echo "API Endpoints:"
echo "- GET /api/v1/projects/:id/cors    - Get project CORS config"
echo "- PUT /api/v1/projects/:id/cors    - Update project CORS config"
echo
echo "Architecture:"
echo "- Global CORS: /api/v1/* routes use global .env CORS_ORIGINS"
echo "- Project CORS: /p/:project_slug/api/* routes use project-specific CORS"
echo "- Database: cors_configs table stores per-project settings"
echo "- Middleware: ProjectCORS checks project settings first, falls back to global"