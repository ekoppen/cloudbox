#!/bin/bash

# Test script to verify all routes are working after fixes

CLOUDBOX_URL="http://localhost:8080"
PROJECT_ID="2"  # Project ID (numeric) - this is what frontend uses
PROJECT_SLUG="dsqewdq"  # Project slug (text) - alternative
API_KEY="YOUR_API_KEY_HERE"  # Replace with actual API key

# Test both ID and slug patterns
echo "Frontend uses Project ID: /p/$PROJECT_ID/api/..."
echo "Alternative slug format: /p/$PROJECT_SLUG/api/..."

echo "🧪 Testing CloudBox Routes After Fixes"
echo "======================================"
echo ""

# Test 1: Health check
echo "1️⃣  Testing health endpoint..."
curl -s "$CLOUDBOX_URL/health" | head -c 100
echo ""
echo ""

# Test 2: Collections endpoint (using Project ID like frontend)
echo "2️⃣  Testing collections endpoint..."
echo "URL: $CLOUDBOX_URL/p/$PROJECT_ID/api/collections"
curl -s -H "X-API-Key: $API_KEY" \
     "$CLOUDBOX_URL/p/$PROJECT_ID/api/collections" | head -c 200
echo ""
echo ""

# Test 3: Auth users endpoint (the problematic one) - using Project ID
echo "3️⃣  Testing auth users endpoint (FRONTEND ROUTE)..."
echo "URL: $CLOUDBOX_URL/p/$PROJECT_ID/api/auth/users"
curl -s -H "X-API-Key: $API_KEY" \
     "$CLOUDBOX_URL/p/$PROJECT_ID/api/auth/users" | head -c 200
echo ""
echo ""

# Test 4: Auth settings endpoint - using Project ID
echo "4️⃣  Testing auth settings endpoint (FRONTEND ROUTE)..."
echo "URL: $CLOUDBOX_URL/p/$PROJECT_ID/api/auth/settings"
curl -s -H "X-API-Key: $API_KEY" \
     "$CLOUDBOX_URL/p/$PROJECT_ID/api/auth/settings" | head -c 200
echo ""
echo ""

# Test 5: Storage buckets - using Project ID
echo "5️⃣  Testing storage buckets..."
echo "URL: $CLOUDBOX_URL/p/$PROJECT_ID/api/storage/buckets"
curl -s -H "X-API-Key: $API_KEY" \
     "$CLOUDBOX_URL/p/$PROJECT_ID/api/storage/buckets" | head -c 200
echo ""
echo ""

# Test 6: Users endpoint - using Project ID  
echo "6️⃣  Testing users endpoint..."
echo "URL: $CLOUDBOX_URL/p/$PROJECT_ID/api/users"
curl -s -H "X-API-Key: $API_KEY" \
     "$CLOUDBOX_URL/p/$PROJECT_ID/api/users" | head -c 200
echo ""
echo ""

echo "🏁 Route testing complete!"
echo ""
echo "📝 Notes:"
echo "- Replace API_KEY with your actual key"
echo "- Replace PROJECT_SLUG if needed"
echo "- 404 errors indicate missing routes"
echo "- 401 errors indicate auth problems"
echo "- 200 with data indicates success"