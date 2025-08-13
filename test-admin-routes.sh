#!/bin/bash

# Test script to verify the new admin routes are working

CLOUDBOX_URL="http://localhost:8080"
PROJECT_ID="2"  
JWT_TOKEN="YOUR_JWT_TOKEN_HERE"

echo "🧪 Testing CloudBox Admin Routes"
echo "================================"
echo ""

# Test 1: Admin Collections endpoint - the one that was failing
echo "1️⃣  Testing admin collections endpoint (NEW ROUTE)..."
echo "URL: $CLOUDBOX_URL/api/v1/projects/$PROJECT_ID/collections"
curl -s -H "Authorization: Bearer $JWT_TOKEN" \
     "$CLOUDBOX_URL/api/v1/projects/$PROJECT_ID/collections" | head -c 200
echo ""
echo ""

# Test 2: Admin Storage Buckets endpoint - the one that was failing  
echo "2️⃣  Testing admin storage buckets endpoint (NEW ROUTE)..."
echo "URL: $CLOUDBOX_URL/api/v1/projects/$PROJECT_ID/storage/buckets"
curl -s -H "Authorization: Bearer $JWT_TOKEN" \
     "$CLOUDBOX_URL/api/v1/projects/$PROJECT_ID/storage/buckets" | head -c 200
echo ""
echo ""

# Test 3: Admin Collections Documents endpoint
echo "3️⃣  Testing admin collections documents endpoint (NEW ROUTE)..."
echo "URL: $CLOUDBOX_URL/api/v1/projects/$PROJECT_ID/collections/test_collection/documents"
curl -s -H "Authorization: Bearer $JWT_TOKEN" \
     "$CLOUDBOX_URL/api/v1/projects/$PROJECT_ID/collections/test_collection/documents" | head -c 200
echo ""
echo ""

# Test 4: Admin Storage Files endpoint
echo "4️⃣  Testing admin storage files endpoint (NEW ROUTE)..."
echo "URL: $CLOUDBOX_URL/api/v1/projects/$PROJECT_ID/storage/buckets/test_bucket/files"
curl -s -H "Authorization: Bearer $JWT_TOKEN" \
     "$CLOUDBOX_URL/api/v1/projects/$PROJECT_ID/storage/buckets/test_bucket/files" | head -c 200
echo ""
echo ""

echo "🏁 Admin route testing complete!"
echo ""
echo "📝 Notes:"
echo "- Replace JWT_TOKEN with your actual JWT token"
echo "- These routes should now return data instead of 404s"
echo "- 404 errors indicate missing routes (BAD)"
echo "- 401 errors indicate auth problems (EXPECTED if no valid JWT)"
echo "- 200 with data indicates success (GOOD)"
echo ""
echo "Expected behavior:"
echo "- WITHOUT valid JWT: 401 Unauthorized"
echo "- WITH valid JWT: 200 OK with JSON data"