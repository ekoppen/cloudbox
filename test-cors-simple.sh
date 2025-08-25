#!/bin/bash

# Simple CORS test without authentication
echo "🧪 CloudBox Project-Specific CORS Test (Simple)"
echo "=============================================="

BASE_URL="http://localhost:8080"
ORIGIN="http://localhost:3123"

echo "🌐 Testing CORS preflight request to project-specific endpoint..."
echo "URL: ${BASE_URL}/p/test/api/collections"
echo "Origin: ${ORIGIN}"
echo ""

CORS_RESPONSE=$(curl -s -i -X OPTIONS "${BASE_URL}/p/test/api/collections" \
  -H "Origin: ${ORIGIN}" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: X-API-Key, Content-Type")

echo "Response:"
echo "========="
echo "$CORS_RESPONSE"

echo ""
echo "🔍 Extracting CORS headers:"
echo "=========================="

# Extract the relevant headers
ALLOW_ORIGIN=$(echo "$CORS_RESPONSE" | grep -i "access-control-allow-origin" | cut -d' ' -f2- | tr -d '\r')
ALLOW_METHODS=$(echo "$CORS_RESPONSE" | grep -i "access-control-allow-methods" | cut -d' ' -f2- | tr -d '\r')
STATUS_LINE=$(echo "$CORS_RESPONSE" | head -1 | tr -d '\r')

echo "📊 Status: $STATUS_LINE"
echo "📊 Access-Control-Allow-Origin: $ALLOW_ORIGIN"
echo "📊 Access-Control-Allow-Methods: $ALLOW_METHODS"

echo ""
echo "📋 Analysis:"
echo "==========="

if [[ "$ALLOW_ORIGIN" == *"$ORIGIN"* ]]; then
    echo "✅ Origin $ORIGIN is explicitly allowed"
elif [[ "$ALLOW_ORIGIN" == "*" ]]; then
    echo "✅ All origins allowed (wildcard)"
else
    echo "❌ Origin $ORIGIN is NOT allowed"
    echo "   Allowed origin: $ALLOW_ORIGIN"
fi

# Test if the endpoint is using project-specific CORS
echo ""
echo "🔍 Testing endpoint routing:"
echo "=========================="
echo "Project endpoint /p/test/api/* should use ProjectCORS middleware"

# Check if we get any response (even 401/403 is ok, we just want to see CORS headers)
if [[ "$STATUS_LINE" == *"200"* ]] || [[ "$STATUS_LINE" == *"401"* ]] || [[ "$STATUS_LINE" == *"403"* ]] || [[ "$STATUS_LINE" == *"204"* ]]; then
    echo "✅ Endpoint is accessible (project-specific routing works)"
else
    echo "❌ Endpoint returned unexpected status: $STATUS_LINE"
fi

echo ""
echo "📋 Conclusion:"
echo "============="

# Check database for project 2 CORS config
echo "🔍 Checking database for project 2 CORS configuration..."
DB_CORS=$(docker-compose exec postgres psql -U cloudbox -d cloudbox -t -c "SELECT allowed_origins FROM cors_configs WHERE project_id = 2;" 2>/dev/null | tr -d ' ' | tr -d '\n' | sed 's/^.*{\(.*\)}.*$/\1/')

if [[ ! -z "$DB_CORS" ]]; then
    echo "✅ Project 2 has CORS configuration in database: $DB_CORS"
    
    # Check if our origin is in the database config
    if [[ "$DB_CORS" == *"$ORIGIN"* ]] || [[ "$DB_CORS" == *"*"* ]]; then
        echo "✅ Origin $ORIGIN is configured in project 2 CORS settings"
        
        if [[ "$ALLOW_ORIGIN" == *"$ORIGIN"* ]] || [[ "$ALLOW_ORIGIN" == "*" ]]; then
            echo "🎯 SUCCESS: Project-specific CORS is working correctly!"
            echo "   • Database has correct configuration"
            echo "   • API returns correct CORS headers"
            echo "   • PhotoPortfolio (localhost:3123) can access CloudBox"
        else
            echo "⚠️  ISSUE: Database has correct config but API doesn't return it"
        fi
    else
        echo "❌ Origin $ORIGIN is NOT in database configuration"
    fi
else
    echo "❌ No CORS configuration found for project 2"
fi