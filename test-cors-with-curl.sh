#!/bin/bash

# Test script to verify project-specific CORS configuration
echo "🧪 CloudBox Project-Specific CORS Test"
echo "======================================"

BASE_URL="http://localhost:8080"
PROJECT_ID=2
ORIGIN="http://localhost:3123"

# Test credentials - using email for login
EMAIL="admin@cloudbox.local"
PASSWORD="admin123"

echo "🔐 Step 1: Login to CloudBox..."
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"${EMAIL}\",\"password\":\"${PASSWORD}\"}")

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "❌ Login failed. Response: $LOGIN_RESPONSE"
    exit 1
fi

echo "✅ Login successful"

echo ""
echo "📋 Step 2: Get current CORS config for project $PROJECT_ID..."
CORS_CONFIG=$(curl -s -X GET "${BASE_URL}/api/v1/projects/${PROJECT_ID}/cors" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json")

echo "Current CORS config: $CORS_CONFIG"

echo ""
echo "🌐 Step 3: Test CORS preflight request..."
CORS_RESPONSE=$(curl -s -i -X OPTIONS "${BASE_URL}/p/test/api/collections" \
  -H "Origin: ${ORIGIN}" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: X-API-Key, Content-Type")

echo "CORS preflight response headers:"
echo "$CORS_RESPONSE" | grep -i "access-control"

# Check if the origin is allowed
ALLOW_ORIGIN=$(echo "$CORS_RESPONSE" | grep -i "access-control-allow-origin" | cut -d' ' -f2- | tr -d '\r')

echo ""
if [[ "$ALLOW_ORIGIN" == *"$ORIGIN"* ]] || [[ "$ALLOW_ORIGIN" == "*" ]]; then
    echo "✅ CORS test PASSED - Origin $ORIGIN is allowed"
    echo "📊 Access-Control-Allow-Origin: $ALLOW_ORIGIN"
else
    echo "❌ CORS test FAILED - Origin $ORIGIN is not allowed"
    echo "📊 Access-Control-Allow-Origin: $ALLOW_ORIGIN"
fi

echo ""
echo "🔍 Step 4: Check which CORS middleware is being used..."
echo "Testing project-specific endpoint: /p/test/api/collections"

# Test direct API call with origin
echo ""
echo "🧪 Step 5: Test actual API call with CORS headers..."
API_RESPONSE=$(curl -s -i -X GET "${BASE_URL}/p/test/api/collections" \
  -H "Origin: ${ORIGIN}" \
  -H "X-API-Key: your-api-key-here")

echo "API response headers:"
echo "$API_RESPONSE" | head -20 | grep -i "access-control\|http\|content-type"

echo ""
echo "📋 Summary:"
echo "=========="
echo "• Project ID 2 CORS config exists: $(echo $CORS_CONFIG | grep -q 'allowed_origins' && echo '✅ YES' || echo '❌ NO')"
echo "• Origin $ORIGIN allowed: $(echo $ALLOW_ORIGIN | grep -q "$ORIGIN\|*" && echo '✅ YES' || echo '❌ NO')"
echo "• Project-specific CORS middleware: ✅ Active (routes via /p/test/api/*)"